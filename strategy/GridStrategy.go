package strategy

import (
	"BandBinance/domain"
	"BandBinance/exchange"
	"BandBinance/store"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"strconv"
	"time"
)

type GridStrategy struct {
}

func (g *GridStrategy) Process() {
	// 加载配置
	symbolConfigs, err := store.GridSymbolConfig.FindEnable()
	if err != nil {
		log.Printf("找不到币种配置, %s", err)
		return
	}
	for i := range symbolConfigs {
		err := exchange.GetPrice()
		if err != nil {
			log.Printf("更新价格失败, %s", err)
			return
		}
		goBuyOrder(symbolConfigs[i])
	}
}

func goBuyOrder(symbolConfig domain.GridSymbolConfig) {
	gridTrades, err := GetTrades(symbolConfig)
	if err != nil {
		return
	}

	// 获取当前价格
	price, err := exchange.LoadPrice(symbolConfig.Symbol)
	err = exchange.UpdateBalance()
	balance, err := exchange.GetBalance("BUSD")
	if err != nil {
		return
	}
	if err != nil {
		return
	}
	log.Printf("币: %s 当前价格: %f", symbolConfig.Symbol, price)

	// 如果 当前价格 > 压力位 , 则风控告警, 看看是否需要人为调整压力位
	if price > symbolConfig.Stress {
		return
	}
	// 如果 当前价格 < 阻力位 , 则风控告警, 看看是否需要人为调整阻力位
	if price < symbolConfig.Resistance {
		return
	}
	// 寻找当前价的靠近的网格下界, 是否当前网格有未完成的订单, 如果有 则跳过, 没有 则挂单
	closeGrid, err := findCloseGrid(gridTrades, price)
	if err != nil {
		log.Printf(fmt.Sprintf("找不到靠近的网格, %s", err))
		return
	}

	// 余额比价格小, 退出
	if balance < closeGrid.LowPrice*closeGrid.Quantity {
		return
	}
	//事务控制
	err = store.Tx(func(tx *gorm.DB) error {
		// 买入单优化, 买单, 挂单未买入, 暂停买入
		exist, err := store.ExistNotBuyInGridTradeOrder(tx, symbolConfig.Symbol, closeGrid.Version, closeGrid.Index)
		if err != nil {
			return err
		}
		if exist {
			return errors.New(fmt.Sprintf("symbol = %s, version= %d, index = %d, 网格单已创建, 挂单未买入, 暂停买入", symbolConfig.Symbol, closeGrid.Version, closeGrid.Index))
		}

		// 创建买单
		var buyOrder domain.TradeOrder
		clientId := time.Now().UnixMilli()
		buyOrder.TradeType = string(binance.SideTypeBuy)
		buyOrder.Status = string(binance.OrderStatusTypeNew)
		buyOrder.CreateTime = time.Now().UnixMilli()
		buyOrder.Version = closeGrid.Version
		buyOrder.Index = closeGrid.Index
		buyOrder.Quantity = closeGrid.Quantity
		buyOrder.BuyPrice = closeGrid.LowPrice
		buyOrder.Symbol = symbolConfig.Symbol
		buyOrder.ClientId = clientId
		// 插入交易表
		err = store.TradeOrder.CreateTradeOrder(tx, &buyOrder)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("订单入库失败, %s", err))
		}
		log.Printf("插入买单成功, id=%d", buyOrder.ID)

		// 交易所创建订单id
		binanceOrderId, err := exchange.CreateOrder(closeGrid.Quantity, closeGrid.LowPrice, binance.SideTypeBuy, strconv.FormatInt(clientId, 10),
			symbolConfig.Symbol, symbolConfig.RetryTimes, float64(symbolConfig.RetryGap))
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("创建订单失败, %s", err))
		}

		// 更新订单id
		buyOrder.OrderId = binanceOrderId
		err = store.TradeOrder.Update(tx, &buyOrder)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("更新订单id失败"))
		}

		log.Printf("创建买入单,id=%d, orderId = %d, price = %.2f , quantity = %.2f", buyOrder.ID, buyOrder.OrderId, buyOrder.BuyPrice, buyOrder.Quantity)
		return nil
	})

	if err != nil {
		log.Printf("币种: %s, 创建买入单结束, %s", symbolConfig.Symbol, err)
	}
}

func GetTrades(symbolConfig domain.GridSymbolConfig) ([]domain.GridTrade, error) {
	// 获取交易网格
	grid, err := newGrid(symbolConfig.ROI, symbolConfig.Amount, symbolConfig.Stress, symbolConfig.Resistance, symbolConfig.Symbol)
	if err != nil {
		return nil, errors.Wrap(err, "获取交易网格失败")
	}
	// 计算静态交易网格
	gridTrades, err := ProcessGridTrades(symbolConfig, grid)
	if err != nil {
		return nil, errors.Wrap(err, "计算静态交易网格失败")
	}
	return gridTrades, err
}

// 找到最靠近价格的网格
func findCloseGrid(trades []domain.GridTrade, price float64) (domain.GridTrade, error) {
	res := 0
	for i := range trades {
		if trades[i].LowPrice < price {
			if trades[res].LowPrice < trades[i].LowPrice {
				res = i
			}
		}
	}
	return trades[res], nil
}

// ProcessGridTrades 计算网格
func ProcessGridTrades(symbolConfig domain.GridSymbolConfig, grid *domain.Grid) ([]domain.GridTrade, error) {
	trades := make([]domain.GridTrade, 0)
	n := 0
	lp := grid.Resistance
	for true {
		hP := lp * (1 + grid.ROI)
		if hP > symbolConfig.Stress {
			break
		}
		// 每次买入 10 busd
		qu := symbolConfig.Amount / lp
		trades = append(trades, domain.GridTrade{
			Id:        store.GenGridTradeId(symbolConfig.Symbol, symbolConfig.Version, n),
			Symbol:    symbolConfig.Symbol,
			HighPrice: hP,
			LowPrice:  lp,
			Index:     n,
			Version:   symbolConfig.Version,
			Quantity:  qu,
		})
		lp = hP
		n++
	}
	_, err := store.SaveTrades(trades)
	if err != nil {
		return nil, err
	}
	return trades, nil
}

func newGrid(roi float64, amount float64, stress float64, resistance float64, symbol string) (*domain.Grid, error) {
	return &domain.Grid{
		Symbol:     symbol,
		ROI:        roi,
		Amount:     amount,
		Stress:     stress,
		Resistance: resistance,
	}, nil
}

func newGridStrategy() (domain.Strategy, error) {
	var strategy domain.Strategy = &GridStrategy{}
	return strategy, nil
}
