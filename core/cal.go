package core

import (
	"BandBinance/config"
	"context"
	"encoding/json"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

type Bet struct {
	BuyPrice   float64 `json:"buy_price"`
	SellPrice  float64 `json:"sell_price"`
	AverageBuy float64 `json:"average_buy"`
	Step       int     `json:"step"`
}

type Profit struct {
	SellRa   float64 `json:"sell_ratio"`
	BuyRa    float64 `json:"buy_ratio"`
	Quantity float64 `json:"quantity"`
}

type SimulateBalance struct {
	Money float64 `json:"money"`
	Coin  float64 `json:"coin"`
}

type PriceData struct {
	E   Bet             `json:"bet"`
	SiL SimulateBalance `json:"simulate_balance"`
	P   Profit          `json:"profit"`
}

var client *binance.Client

func init() {
	client = binance.NewClient(config.ApiKey, config.SecretKey)
	client.BaseURL = "https://api.binance.cc"
}

func (p *PriceData) ToTrade() {
	curSymbol, err := client.NewListPricesService().Symbol(config.Symbol).Do(context.Background())
	price, _ := strconv.ParseFloat(curSymbol[0].Price, 64)
	if err != nil {
		fmt.Println(err)
		return
	}
	quantity := strconv.FormatFloat(p.P.Quantity, 'E', -1, 64)
	if p.E.BuyPrice >= price {
		buyPrice := strconv.FormatFloat(p.E.BuyPrice, 'E', -1, 64)
		if config.Simulate != true {
			_, err := client.NewCreateOrderService().Symbol(config.Symbol).Side(binance.SideTypeBuy).
				Type(binance.OrderTypeLimit).Quantity(quantity).Price(buyPrice).Do(context.Background())
			if err != nil {
				fmt.Println(err)
				return
			}
		}
		p.ModifyPrice(price, 1, "Buy", config.Simulate)
	}else if p.E.SellPrice < price{
		if p.E.Step == 0{
			p.ModifyPrice(price, 0, "Sell", config.Simulate)
		}else{
			if config.Simulate != true{
				sellPrice := strconv.FormatFloat(p.E.SellPrice, 'E', -1, 64)
				_, err := client.NewCreateOrderService().Symbol(config.Symbol).Side(binance.SideTypeSell).
					Type(binance.OrderTypeLimit).Quantity(quantity).Price(sellPrice).Do(context.Background())
				if err != nil {
					fmt.Println(err)
					return
				}
			}
			p.ModifyPrice(price, 1, "Sell", config.Simulate)
		}
	}

}

func (p *PriceData) ModifyPrice(dealPrice float64, step int, tradeType string, sli bool) {
	rightSize := len(strings.Split(strconv.FormatFloat(dealPrice, 'E', -1, 64), ".")[1])
	p.E.BuyPrice = round(dealPrice*(1-p.P.BuyRa), rightSize)
	p.E.SellPrice = round(dealPrice*(1+p.P.SellRa), rightSize)
	p.E.Step += step
	if sli == true && step != 0 {
		if tradeType == "Buy" {
			realDealPrice := dealPrice * (1 + config.Fee)
			p.SiL.Money -= realDealPrice
			p.SiL.Coin += p.P.Quantity

		} else {
			realDealPrice := dealPrice * (1 - config.Fee)
			p.SiL.Money += realDealPrice
			p.SiL.Coin -= p.P.Quantity
		}
	}
	if tradeType == "Buy" {
		p.E.AverageBuy = (dealPrice*(1+config.Fee) + p.E.AverageBuy) / 2
	}
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
