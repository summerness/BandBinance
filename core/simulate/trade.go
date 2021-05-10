package simulate

import (
	"BandBinance/config"
	"BandBinance/data"
	"BandBinance/logger"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Bet struct {
	BuyPrice    float64 `json:"buy_price"`
	SellPrice   float64 `json:"sell_price"`
	BuyAverage  float64 `json:"buy_average"`
	SellAverage float64 `json:"sell_average"`
	Step        int     `json:"step"`
	Type        int     `json:"type"`
}

type SimBalance struct {
	Money        float64 `json:"money"`
	Coin         float64 `json:"coin"`
	HoldingMoney float64 `json:"holding_money"`
	HoldingCoin  float64 `json:"holding_coin"`
}

type Ori struct {
	OriMoney float64 `json:"ori_money"`
	OriCoin  float64 `json:"ori_coin"`
}

type TradeData struct {
	Bs         []Bet      `json:"bets"`
	SiL        SimBalance `json:"sim_balance"`
	Spend      float64    `json:"spend"`
	SetupPrice float64    `json:"setup_price"`
	LimitQ     int        `json:"limitq"` //最少
	O          Ori        `json:"ori"`
	SaveCoin   float64    `json:"save_coin"`
}



func (p *TradeData) quantityBySellBuy(buyPrice, sellPrice float64) (float64,float64) {
	maxQuantity := p.roundQuantity(p.Spend / buyPrice * (1 - config.Fee/100))      //实际买到的币
	realSellSpend := p.Spend / (1 - config.Fee/100)                                //实际应该卖到不亏的价格trade
	minQuantity := p.roundQuantity(realSellSpend / sellPrice)                      //最少卖出币
	savedQuantity := maxQuantity - minQuantity                                     //实际可留下的利润
	realSavedQuantity := p.roundQuantity((1 - config.SaveRa) / 10 * savedQuantity) //按照设置留下利润
	return minQuantity + realSavedQuantity,realSavedQuantity
}

func (p *TradeData) simLimitBuy(price float64, bType int) {
	quantity := p.roundQuantity(p.Spend / price)
	realQuantity := p.roundQuantity(quantity * (1 - config.Fee/100)) //真实买到的
	data.InsertOne("Buy", price, quantity, p.Spend, realQuantity, bType)
	p.SiL.Money -= p.Spend
	p.SiL.HoldingMoney += p.Spend
}

func (p *TradeData) simLimitSell(buyPrice, sellPrice float64, bType int) {
	quantity,reQ := p.quantityBySellBuy(buyPrice, sellPrice)
	bs := strconv.FormatFloat(buyPrice, 'E', -1, 64)
	rightSize := len(strings.Split(bs, ".")[1])
	realPrice := round(quantity*sellPrice*(1-config.Fee/100), rightSize) //真实收到的钱
	data.InsertOne("Sell", sellPrice, quantity, realPrice, quantity, bType)
	p.SiL.Coin -= quantity
	p.SiL.HoldingCoin += quantity
	p.SaveCoin += reQ
}

func (p *TradeData) countProfit(price float64) {
	totalCoin := p.SiL.Coin + p.SiL.HoldingCoin
	totalMoneyNow := p.SiL.Money + p.SiL.HoldingMoney + totalCoin*price
	totalMoneyOri := p.SiL.Money + p.SiL.HoldingMoney + totalCoin*p.SetupPrice
	originMoney := p.O.OriMoney + p.O.OriCoin*p.SetupPrice
	profitNow := totalMoneyNow - originMoney
	profitRaNow := profitNow / originMoney
	profitOri := totalMoneyOri - originMoney
	profitRaOri := profitOri / originMoney
	coinReal := price - p.SetupPrice
	coinRa := coinReal / p.SetupPrice
	logger.Info(fmt.Sprintf("Limit交易方式 剩余钱：%f, 剩余币：%f, 订单中的钱：%f, "+
		"订单中的币：%f, 按照现价盈利：%f (%f) ,按照策略开始时 %f (%f), 币价：%f (%f)",
		p.SiL.Money, p.SiL.Coin, p.SiL.HoldingMoney, p.SiL.HoldingCoin, profitNow, profitRaNow,

		profitOri, profitRaOri, coinReal, coinRa))
}

func (p *TradeData) sellSaveCoin(price float64) {
	if price > p.SetupPrice*(1+config.Earn/100) {
		if p.SiL.Coin > p.SaveCoin && p.SaveCoin > 0{
			realPrice := price * p.SaveCoin * (1 - config.Fee/100) //真实收到的钱
			go data.InsertOne("Sell", price, p.SaveCoin, realPrice, p.SaveCoin, 100)
			p.SiL.Coin -= p.SaveCoin
			p.SiL.HoldingCoin += p.SaveCoin
		}
	}
	// 重置网格
}

func (p *TradeData) ToLimitTrade() {
	curSymbol, err :=  client.NewListPricesService().Symbol(config.Symbol).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	price, _ := strconv.ParseFloat(curSymbol[0].Price, 64)
	openOrders := data.GetOrder(price)
	for _, each := range openOrders {
		if each.TradeType == "Buy" {
			p.SiL.HoldingMoney -= each.RealPrice
			p.SiL.Coin += each.RealCoin
		} else if each.TradeType == "Sell" {
			p.SiL.HoldingCoin -= each.RealCoin
			p.SiL.Money += each.RealPrice
		}
		go data.UpdateOrder(each.Id)
	}
	if len(openOrders) > 0 {
		p.countProfit(price)
	}
	p.sellSaveCoin(price)
	for _, each := range p.Bs {
		if each.BuyPrice >= price {
			if each.Step == config.MaxStep {
				p.ModifyPrice(price, 0, "Buy", each.Type)
				continue
			}
			if p.Spend > p.SiL.Money { //没钱了
				p.ModifyPrice(price, 0, "ba", each.Type)
			} else {
				p.simLimitBuy(price, each.Type)
				p.simLimitSell(price, each.SellPrice, each.Type)
				p.ModifyPrice(each.BuyAverage, 1, "Sell", each.Type)
			}
		} else if each.SellPrice < price {
			quantity,_ := p.quantityBySellBuy(each.BuyPrice, price)
			if each.Step == 0 {
				p.ModifyPrice(price, 0, "Sell", each.Type)
				continue
			}
			if quantity > p.SiL.Coin { //没币了
				p.ModifyPrice(price, 0, "ba", each.Type)
			} else {
				p.simLimitBuy(each.BuyPrice, each.Type)
				p.simLimitSell(each.BuyPrice, price, each.Type)
				p.ModifyPrice(price, -1, "Sell", each.Type)
			}
		}
	}

}

func (p *TradeData) ModifyPrice(dealPrice float64, step int, tradeType string, bType int) {
	rightSize := len(strings.Split(strconv.FormatFloat(dealPrice, 'E', -1, 64), ".")[1])
	for index, each := range p.Bs {
		if each.Type == bType {
			if tradeType == "Buy" {
				p.Bs[index].BuyAverage = (p.Bs[index].BuyAverage + p.Bs[index].BuyPrice) / 2
			} else if tradeType == "Sell" {
				p.Bs[index].SellAverage = (p.Bs[index].SellAverage + p.Bs[index].SellPrice) / 2
			}
			p.Bs[index].BuyPrice = round(dealPrice*(1-config.NetRa[each.Type]/100), rightSize)
			p.Bs[index].SellPrice = round(dealPrice*(1+config.NetRa[each.Type]/100), rightSize)
			p.Bs[index].Step += step
			break
		}
	}
	//p.changeRatio()
	p.save()
}

func (p *TradeData) save() (bool, error) {
	saveData, _ := json.Marshal(p)
	dataPath := config.DataPath
	err := ioutil.WriteFile(dataPath, saveData, os.ModeAppend)
	if err != nil {
		return false, err
	}
	return true, nil
}


func InitPriceData() (p *TradeData) {
	dataPath := config.DataPath
	jsonFile, err := os.Open(dataPath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &p)
	return
}

func (p *TradeData) roundQuantity(f float64) float64 {
	return round(f, p.LimitQ)
}
