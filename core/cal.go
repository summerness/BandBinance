package core

import (
	"BandBinance/config"
	"BandBinance/data"
	"BandBinance/logger"
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type Bet struct {
	BuyPrice     float64 `json:"buy_price"`
	SellPrice    float64 `json:"sell_price"`
	BuyAverage  float64 `json:"buy_average"`
	SellAverage float64 `json:"sell_average"`
	Step         int     `json:"step"`
	Type         int     `json:"type"`
}

type SimulateBalance struct {
	Money        float64 `json:"money"`
	Coin         float64 `json:"coin"`
	HoldingMoney float64 `json:"holding_money"`
	HoldingCoin  float64 `json:"holding_coin"`
}

type Ori struct {
	OriMoney float64 `json:"ori_money"`
	OriCoin  float64 `json:"ori_coin"`
}

type PriceData struct {
	Bs         []Bet           `json:"bets"`
	SiL        SimulateBalance `json:"simulate_balance"`
	Spend      float64         `json:"spend"`
	SetupPrice float64         `json:"setup_price"`
	LimitQ     int             `json:"limitq"`
	O          Ori             `json:"ori"`
}

var (
	client *binance.Client
)

func init() {
	client = binance.NewClient(config.ApiKey, config.SecretKey)
	logger.Setup()
	//client.BaseURL = "https://api.binance.cc"
}

func (p *PriceData) ToTrade() {
	curSymbol, err := client.NewListPricesService().Symbol(config.Symbol).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	price, _ := strconv.ParseFloat(curSymbol[0].Price, 64)
	// 模拟收益
	if config.Simulate == true {
		tras := data.UpdateDeal(price)
		for _, each := range tras {
			if each.TradeType == "Buy" {
				p.SiL.HoldingMoney -= each.RealPrice
				p.SiL.Coin += each.RealCoin
			} else {
				p.SiL.HoldingCoin -= each.RealCoin
				p.SiL.Money += each.RealPrice
			}
		}
		if len(tras) != 0 {
			total_coin := p.SiL.Coin + p.SiL.HoldingCoin
			total_money := p.SiL.Money + p.SiL.HoldingMoney + total_coin*price
			origin_money := p.O.OriMoney + p.O.OriCoin*p.SetupPrice
			profit := total_money / origin_money
			coin_w := p.SetupPrice / price
			logger.Info(fmt.Sprintf("剩余钱：%f, 剩余币：%f, 订单中的钱：%f, 订单中的币：%f, 如果现在平仓盈利：%f, 币价格涨幅：%f",
				p.SiL.Money, p.SiL.Coin, p.SiL.HoldingMoney, p.SiL.HoldingCoin, profit, coin_w))
		}
	}
	quantity := round(p.Spend/price, p.LimitQ)

	for _, each := range p.Bs {
		if price >= p.SetupPrice*1.1 {
			saved_coin := p.SiL.Coin - p.O.OriCoin
			reals := price * saved_coin * (1 - config.Fee/100)
			if config.Simulate == true {
				go data.InsertOne("Sell", price, quantity, reals, saved_coin, 100)
			} else {
				sellQuantity := strconv.FormatFloat(saved_coin, 'E', -1, 64)
				sellPrice := strconv.FormatFloat(price, 'E', -1, 64)
				_, err := client.NewCreateOrderService().Symbol(config.Symbol).Side(binance.SideTypeSell).
					Type(binance.OrderTypeLimit).Quantity(sellQuantity).Price(sellPrice).Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			p.ModifyPrice(price, 0, "Sell", each.Type)
		}
		if each.BuyPrice >= price {
			p.ModifyPrice(price, 1, "Buy", each.Type)
			//模拟Buy
			if config.Simulate == true {
				realQuantity := quantity * (1 - config.Fee/100) //真实买到的
				if p.Spend > p.SiL.Money { //没钱了
					p.ModifyPrice(each.BuyAverage, 0, "Sell", each.Type)
					return
				}
				go data.InsertOne("Buy", price, quantity, p.Spend, realQuantity, each.Type)
				p.SiL.Money -= p.Spend
				p.SiL.HoldingMoney += p.Spend
			} else {
				buyQuantity := strconv.FormatFloat(quantity, 'E', -1, 64)
				buyPrice := strconv.FormatFloat(price, 'E', -1, 64)
				_, err := client.NewCreateOrderService().Symbol(config.Symbol).Side(binance.SideTypeBuy).
					Type(binance.OrderTypeLimit).Price(buyPrice).Quantity(buyQuantity).Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
			}

		} else if each.SellPrice <= price {
			if each.Step == 0 {
				p.ModifyPrice(price, 0, "Sell", each.Type)
				return
			}
			p.ModifyPrice(price, -1, "Sell", each.Type)
			//模拟Sell
			if config.Simulate == true {
				if quantity > p.SiL.Coin { //没币了
					p.ModifyPrice(each.SellAverage, 0, "Buy", each.Type)
					return
				}
				realPrice := p.Spend * (1 - config.Fee/100)
				go data.InsertOne("Sell", price, quantity, realPrice, quantity, each.Type)
				p.SiL.Coin -= quantity
				p.SiL.HoldingCoin += quantity
			} else {
				sellQuantity := strconv.FormatFloat(quantity, 'E', -1, 64)
				sellPrice := strconv.FormatFloat(price, 'E', -1, 64)
				_, err := client.NewCreateOrderService().Symbol(config.Symbol).Side(binance.SideTypeSell).
					Type(binance.OrderTypeLimit).Quantity(sellQuantity).Price(sellPrice).Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
			}
		}
	}
	time.Sleep(time.Minute)
}




func (p *PriceData) ModifyPrice(dealPrice float64, step int, tradeType string, bType int) {
	rightSize := len(strings.Split(strconv.FormatFloat(dealPrice, 'E', -1, 64), ".")[1])
	for index, each := range p.Bs {
		if each.Type == bType {
			if tradeType == "Buy" {
				p.Bs[index].BuyAverage = (p.Bs[index].BuyAverage + p.Bs[index].BuyPrice)/2
				p.Bs[index].BuyPrice = round(dealPrice*(1-config.NetRa[each.Type]/100), rightSize)
			} else {
				p.Bs[index].SellAverage = (p.Bs[index].SellAverage + p.Bs[index].SellPrice)/2
				p.Bs[index].SellPrice = round(dealPrice*(1+config.NetRa[each.Type]/100), rightSize)
			}
			p.Bs[index].Step += step
			break
		}
	}
	//p.changeRatio()
	p.save()
}

func InitPriceData() (p *PriceData) {
	jsonFile, err := os.Open(config.DataPath)
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()
	byteValue, _ := ioutil.ReadAll(jsonFile)
	json.Unmarshal(byteValue, &p)
	return
}

func (p *PriceData) save() (bool, error) {
	saveData, _ := json.Marshal(p)
	err := ioutil.WriteFile(config.DataPath, saveData, os.ModeAppend)
	if err != nil {
		return false, err
	}
	return true, nil
}

func round(f float64, n int) float64 {
	floatStr := fmt.Sprintf("%."+strconv.Itoa(n)+"f", f)
	inst, _ := strconv.ParseFloat(floatStr, 64)
	return inst
}
