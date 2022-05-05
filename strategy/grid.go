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
	"sync"
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
		goBuyOrder(&symbolConfigs[i])
	}
}

func goBuyOrder(symbolConfig *domain.GridSymbolConfig) {
	// 更新价格
	err := exchange.GetPrice()
	if err != nil {
		log.Printf("更新价格失败, %s", err)
		return
	}
	// 更新余额
	err = exchange.UpdateBalance()
	if err != nil {
		log.Printf("获取用户余额失败, %s", err)
		return
	}
	gridTrades, err := ProcessGridTrades(symbolConfig)
	if err != nil {
		return
	}
	// fmt.Printf("gridTrades: %+v\n", gridTrades)
	// 获取当前价格
	price, err := exchange.LoadPrice(symbolConfig.Symbol)
	if err != nil {
		log.Println("获取当前价格时出现错误: ", err)
		return
	}

	log.Printf("币: %s 当前价格: %f", symbolConfig.Symbol, price)

	// 如果 当前价格 > 压力位 , 则风控告警, 看看是否需要人为调整压力位
	if price > symbolConfig.Stress {
		log.Println("当前价格 > 压力位")
		return
	}
	// 如果 当前价格 < 阻力位 , 则风控告警, 看看是否需要人为调整阻力位
	if price < symbolConfig.Resistance {
		log.Println("当前价格 < 阻力位")
		return
	}
	// 寻找当前价的靠近的网格下界, 是否当前网格有未完成的订单, 如果有 则跳过, 没有 则挂单
	closeGrid, err := findCloseGrid(gridTrades, price)
	if err != nil {
		log.Println("找不到靠近的网格: ", err)
		return
	}
	balance, err := exchange.GetBalance("BUSD")
	if err != nil {
		log.Println("获取余额时出现错误: ", err)
		return
	}
	// 余额<价格*数量(需要金额) , 退出
	if balance < closeGrid.LowPrice*closeGrid.Quantity {
		log.Println("余额 < 需要金额 本次无效退出")
		return
	}
	// 事务控制
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
		clientId := time.Now().UnixNano() / 1e6
		buyOrder.TradeType = string(binance.SideTypeBuy)
		buyOrder.Status = string(binance.OrderStatusTypeNew)
		buyOrder.CreateTime = time.Now().UnixNano() / 1e6
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

// 找到最靠近价格的网格  刚好比网格下端高一点
func findCloseGrid(trades []domain.GridTrade, price float64) (domain.GridTrade, error) {
	res := 0
	for i := range trades {
		if trades[i].LowPrice < price {
			// 这一步不用检查吧  毕竟是从小往大按照顺序排列的
			if trades[res].LowPrice < trades[i].LowPrice {
				res = i
			}
		}
	}
	return trades[res], nil
}

var once sync.Once

// ProcessGridTrades 计算网格  处理网格交易
// 从最低价格开始 每次加一个利润率百分比 直到大于上限才跳出
func ProcessGridTrades(symbolConfig *domain.GridSymbolConfig) (trades []domain.GridTrade, err error) {
	once.Do(func() {
		n := 0
		lp := symbolConfig.Resistance // 阻力位  从最低价格开始
		for {
			hP := lp * (1 + symbolConfig.ROI) // 阻力位 * (1+收益率) ? == 最低价格 * 1.2 == 当前价格
			if hP > symbolConfig.Stress {     // 上限 > 压力位, 全部卖出的价格 则取消操作
				break
			}
			// 数量 = 每单金额 / 上一格的价格 ??
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
		_, err = store.SaveTrades(trades) // 将网格保存在数据库
		if err != nil {
			fmt.Println(`将网格保存在数据库时出现错误`)
		}
	})

	return trades, nil
}

// newGrid 像是一个中间产物  已经删除了对其的调用  日后若不使用可删除
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
