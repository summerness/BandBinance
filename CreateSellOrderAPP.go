package main

import (
	"BandBinance/config"
	"BandBinance/domain"
	"BandBinance/exchange"
	"BandBinance/notify"
	"BandBinance/store"
	"BandBinance/strategy"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"strconv"
	"time"
)

func main() {
	prefix := "创建卖单"
	t := time.NewTicker(time.Duration(config.Run.SellSleep) * time.Second)
	for {
		// 找出币配置
		configs, err := store.GridSymbolConfig.FindEnable()
		if err != nil {
			log.Printf("%s 币种配置没有 %s", prefix, err)
			continue
		}

		for i := range configs {
			err = ProcessBuyOrderForSell(&configs[i])
			if err != nil {
				log.Printf("%s 处理币种配置时出现错误, id=%s, symbol=%s,err: %s", prefix, configs[i].ID, configs[i].Symbol, err)
			}
		}
		fmt.Println(prefix, `检查完成`)
		<-t.C
	}
}

// ProcessBuyOrderForSell 遍历买入单, 更新状态, 并且创建卖出单, 事务回滚,
func ProcessBuyOrderForSell(symbolConfig *domain.GridSymbolConfig) error {
	gridTrades, err := strategy.ProcessGridTrades(symbolConfig)
	if err != nil {
		return errors.Wrapf(err, fmt.Sprintf("获取网格时出现错误, %s", err))
	}
	orders, err := store.FindTradeOrder(symbolConfig.Symbol, string(binance.SideTypeBuy), string(binance.OrderStatusTypeNew), symbolConfig.Version)
	if err != nil {
		//	找不到订单,跳过
		return errors.Wrap(err, fmt.Sprintf("遍历买入单创建卖出单, 找不到订单, %s", err))
	}
	for i := range orders {
		// 查询订单状态
		binanceOrder, err := exchange.GetOrder(orders[i].OrderId, symbolConfig.Symbol)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("获取币安订单失败,id=%d, order_id=%d", orders[i].ID, orders[i].OrderId))
		}
		// 事务控制
		err = store.Tx(func(tx *gorm.DB) error {
			// 如果已经成交, 则创建卖出单
			if binanceOrder.Status == string(binance.OrderStatusTypeFilled) {
				// 更新订单状态
				orders[i].BuyPrice = binanceOrder.BuyPrice
				orders[i].DealTime = binanceOrder.DealTime
				orders[i].Quantity = binanceOrder.Quantity
				orders[i].Status = binanceOrder.Status
				err = store.TradeOrder.Update(tx, &orders[i])
				if err != nil {
					log.Printf("更新买入单失败, orderId = %d", binanceOrder.OrderId)
					return err
				}
				// 找到网格
				gridTrade := gridTrades[orders[i].Index]

				// 创建卖单
				var sellOrder domain.TradeOrder
				sellOrder.TradeType = string(binance.SideTypeSell)
				sellOrder.Index = orders[i].Index
				sellOrder.Version = orders[i].Version
				sellOrder.Quantity = orders[i].Quantity
				sellOrder.BuyPrice = gridTrade.HighPrice
				sellOrder.CreateTime = time.Now().UnixNano() / 1e6
				sellOrder.Status = string(binance.OrderStatusTypeNew)
				sellOrder.ClientId = orders[i].ClientId
				sellOrder.Symbol = orders[i].Symbol
				// 保存卖出单
				err = store.TradeOrder.CreateTradeOrder(tx, &sellOrder)
				if err != nil {
					log.Printf("保存卖出单失败, 卖出单id=%d", sellOrder.OrderId)
					return err
				}
				log.Printf("卖出单id=%d", sellOrder.ID)

				// 创建卖出单
				clientId := strconv.FormatInt(time.Now().UnixNano()/1e6, 10)
				binanceOrderId, err := exchange.CreateOrder(binanceOrder.Quantity, gridTrade.HighPrice, binance.SideTypeSell, clientId, symbolConfig.Symbol, symbolConfig.RetryTimes, float64(symbolConfig.RetryGap))
				if err != nil {
					log.Printf("创建卖出单失败, 买入单id=%d", binanceOrder.OrderId)
					return err
				}
				// 更新卖出单
				sellOrder.OrderId = binanceOrderId
				err = store.TradeOrder.Update(tx, &sellOrder)
				if err != nil {
					log.Printf("更新卖出单order_id失败, 买入单id=%d, order_id = %d", binanceOrder.ID, binanceOrder.OrderId)
					return err
				}
				notify.DefaultNotify.Trade(sellOrder)
			}
			return nil
		})
	}
	return err
}
