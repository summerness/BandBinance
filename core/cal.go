package core

import (
	"BandBinance/config"
	"BandBinance/data"
	"BandBinance/logger"
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/futures"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"
)

type Bet struct {
	BuyPrice   float64 `json:"buy_price"`
	SellPrice  float64 `json:"sell_price"`
	AverageBuy float64 `json:"average_buy"`
	Step       int     `json:"step"`
}

type Profit struct {
	SellRa     float64 `json:"sell_ratio"`
	BuyRa      float64 `json:"buy_ratio"`
	Quantity   float64 `json:"quantity"`
	StartPrice float64 `json:"start_price"`
}

type SimulateBalance struct {
	Money        float64 `json:"money"`
	Coin         float64 `json:"coin"`
	HoldingMoney float64 `json:"holding_money"`
	HoldingCoin  float64 `json:"holding_coin"`
}

type PriceData struct {
	E   Bet             `json:"bet"`
	SiL SimulateBalance `json:"simulate_balance"`
	P   Profit          `json:"profit"`
}

var (
	client  *binance.Client
	fclient *futures.Client
)

func init() {
	client = binance.NewClient(config.ApiKey, config.SecretKey)
	fclient = binance.NewFuturesClient(config.ApiKey, config.SecretKey)
	fclient.BaseURL = "https://www.binance.com/api/v1"
	logger.Setup()
	//client.BaseURL = "https://api.binance.cc"
}

func (p *PriceData) changeRatio() {
	ht := Hours24Tickers()
	if ht > 10 {
		p.P.BuyRa = p.P.BuyRa * (1 - float64(p.E.Step)*0.01)
	}
	if ht < - 10 {
		p.P.SellRa = p.P.SellRa * (1 + float64(p.E.Step)*0.01)
	}
}

func (p *PriceData) ToTrade() {
	curSymbol, err := client.NewListPricesService().Symbol(config.Symbol).Do(context.Background())
	if err != nil {
		fmt.Println(err)
		return
	}
	price, _ := strconv.ParseFloat(curSymbol[0].Price, 64)
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
			origin_money := 100 + 21*p.P.StartPrice
			profit := total_money / origin_money
			coin_w := p.P.StartPrice/price
			logger.Info(fmt.Sprintf("剩余钱：%f, 剩余币：%f, 订单中的钱：%f, 订单中的币：%f, 如果现在平仓盈利：%f, 币价格涨幅：%f",
				p.SiL.Money, p.SiL.Coin, p.SiL.HoldingMoney, p.SiL.HoldingCoin, profit,coin_w))
		}
	}
	quantity := strconv.FormatFloat(p.P.Quantity, 'E', -1, 64)
	if p.E.BuyPrice >= price {
		buyPrice := strconv.FormatFloat(price, 'E', -1, 64)
		if config.Simulate != true {
			_, err := client.NewCreateOrderService().Symbol(config.Symbol).Side(binance.SideTypeBuy).
				Type(binance.OrderTypeLimit).Price(buyPrice).Quantity(quantity).Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}
		} else {
			realPrice := price * p.P.Quantity
			realQuantity := p.P.Quantity * (1 - config.Fee/100)
			if realPrice > p.SiL.Money{
				p.ModifyPrice(price,0,"Buy")
				return
			}


			go data.InsertOne("Buy", price, p.P.Quantity, realPrice, realQuantity)
			p.SiL.Money -= realPrice
			p.SiL.HoldingMoney += realPrice
		}
		p.ModifyPrice(price, 1, "Buy")
	} else if p.E.SellPrice < price {
		if p.E.Step == 0 {
			p.ModifyPrice(price, 0, "Sell")
		} else {
			if config.Simulate != true {
				sellPrice := strconv.FormatFloat(p.E.SellPrice, 'E', -1, 64)
				_, err := client.NewCreateOrderService().Symbol(config.Symbol).Side(binance.SideTypeSell).
					Type(binance.OrderTypeLimit).Quantity(quantity).Price(sellPrice).Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
			} else {
				{
					realPrice := price * p.P.Quantity * (1 - config.Fee/100)
					realQuantity := p.P.Quantity
					if realQuantity > p.SiL.Coin{
						p.ModifyPrice(price,-1,"Sell")
						return
					}


					p.SiL.Coin -= p.P.Quantity
					p.SiL.HoldingCoin += p.P.Quantity
					go data.InsertOne("Sell", price, realQuantity, realPrice, realQuantity)

				}
			}
			p.ModifyPrice(price, -1, "Sell")
		}
	}
}

func (p *PriceData) ModifyPrice(dealPrice float64, step int, tradeType string) {
	rightSize := len(strings.Split(strconv.FormatFloat(dealPrice, 'E', -1, 64), ".")[1])
	p.E.BuyPrice = round(dealPrice*(1-p.P.BuyRa/100), rightSize)
	p.E.SellPrice = round(dealPrice*(1+p.P.SellRa/100), rightSize)
	p.E.Step += step
	if tradeType == "Buy" {
		p.E.AverageBuy = (dealPrice*(1+config.Fee/100) + p.E.AverageBuy) / 2
	}
	//p.changeRatio()
	p.save()
	time.Sleep(time.Minute)
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
