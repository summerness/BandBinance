package analyse

import (
	"BandBinance/domain"
	"BandBinance/notify"
	"BandBinance/store"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"log"
)

func Scene(history domain.History) (domain.Scene, error) {
	return domain.GRID, nil
}

func Profit() error {
	// panic("")
	return nil
}

// 计算收益?
func CalculateBenefit() {
	// 获取总交易对
	tradePairTotal, err := store.TradeOrder.CountClientId()
	if err != nil {
		log.Printf("获取总交易对时出现错误: %s", err)
		return
	}

	// 已完成交易对的clientId
	clientIds, err := store.TradeOrder.FindCompletedGrid()
	if err != nil {
		log.Printf("获取已完成交易对的clientId时出现错误: %s", err)
		return
	}

	// 找到全部交易的订单
	tradeOrders, err := store.TradeOrder.FindTradeOrderByClientIds(clientIds)
	if err != nil {
		log.Printf("计算收益, %s", err)
		return
	}

	// 累计成本
	var cost float64
	// 累计收益
	var benefit float64

	for i := range tradeOrders {
		spend := tradeOrders[i].BuyPrice * tradeOrders[i].Quantity
		if string(binance.SideTypeSell) == tradeOrders[i].TradeType {
			benefit += spend
		} else if string(binance.SideTypeBuy) == tradeOrders[i].TradeType {
			cost += spend
		}
	}

	// 买入未卖出
	buyInNotSellTotal, buyInNotSellTotalLockCost := CalculateBuyInNotSell()

	// 挂单未买入
	createdNotBuyTotal, createdNotBuyTotalLockCost := CalculateCreatedNotBuy()

	notify.DefaultNotify.Do(fmt.Sprintf(
		"收益按照网格成功交易计算 \n累计收益 %.2f BUSD , 累计买入 %.2f BUSD, "+
			"累计卖出 %.2f BUSD \n买入未卖出单: %d,  %.2f BUSD \n挂单未买入: %d , %.2f BUSD \n总交易对: %d , 网格成功数: %d, 网格成功率: %.2f ",
		benefit-cost, cost, benefit,
		buyInNotSellTotal, buyInNotSellTotalLockCost,
		createdNotBuyTotal, createdNotBuyTotalLockCost,
		tradePairTotal, len(clientIds), float64(len(clientIds))/float64(tradePairTotal),
	))
}

// CalculateCreatedNotBuy 挂单未买入
func CalculateCreatedNotBuy() (int, float64) {
	orders, err := store.TradeOrder.FindCreatedNotBuy()
	if err != nil {
		return 0, 0
	}
	var money float64
	for i := range orders {
		money += orders[i].GetSpend()
	}

	return len(orders), money
}

// CalculateBuyInNotSell 计算锁仓成本
func CalculateBuyInNotSell() (int, float64) {
	var r float64
	// 买入未卖出
	orders, err := store.TradeOrder.FindBuyInNotSellTradeOrders()
	if err != nil {
		log.Printf("计算锁仓成本 %s", err)
		return 0, 0
	}
	var n int
	for i := range orders {
		if orders[i].TradeType == string(binance.SideTypeBuy) {
			spend := orders[i].GetSpend()
			r += spend
			n++
		}
	}
	return n, r
}
