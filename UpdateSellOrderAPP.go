package main

import (
	"BandBinance/analyse"
	"BandBinance/config"
	"BandBinance/exchange"
	"BandBinance/store"
	"github.com/adshao/go-binance/v2"
	"log"
	"time"
)

// 更新卖单状态, 并计算收益率, 通知飞书
func main() {
	analyse.CalculateBenefit()
	ticker := time.NewTicker(time.Duration(config.Run.BenefitSleep) * time.Second)
	defer ticker.Stop()

	for {
		log.Printf("开始 更新卖单状态, 并计算收益率, 通知飞书")
		ProcessSellOrder()
		log.Printf("结束 更新卖单状态, 并计算收益率, 通知飞书")
		<-ticker.C
	}
}

// ProcessSellOrder 处理卖单状态
func ProcessSellOrder() {
	// 获取NEW状态的卖单
	orders, err := store.FindNewTradeOrder(binance.SideTypeSell)
	if err != nil {
		log.Printf("获取NEW状态的卖单, %s", err)
		return
	}

	// 查看卖单状态
	for i := range orders {
		binanceOrder, err := exchange.GetOrder(orders[i].OrderId, orders[i].Symbol)
		if err != nil {
			log.Printf("获取币安订单失败, %s", err)
			continue
		}

		// 更新状态
		orders[i].Status = binanceOrder.Status
		orders[i].DealTime = binanceOrder.DealTime
		err = store.TradeOrder.Update(store.DB, &orders[i])
		if err != nil {
			log.Printf("更新卖出单失败, %s", err)
			continue
		}

		if string(binance.OrderStatusTypeFilled) == binanceOrder.Status {
			// 卖出单成交, 触发收益监控
			analyse.CalculateBenefit()
		}

	}

}
