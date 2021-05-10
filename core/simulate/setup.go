package simulate

import (
	"BandBinance/config"
	"BandBinance/logger"
	"context"
	"github.com/adshao/go-binance/v2"
	"strconv"
	"strings"
)


var (
	client *binance.Client
)

func init() {
	client = binance.NewClient(config.ApiKey, config.SecretKey)
	logger.Setup()
}

func InitSaveData() {
	res, err := client.NewListPriceChangeStatsService().Symbol(config.Symbol).Do(context.Background())
	if err != nil {
		return
	}
	weightedAvgPrice, _ := strconv.ParseFloat(res[0].WeightedAvgPrice, 64)
	curSymbol, err := client.NewListPricesService().Symbol(config.Symbol).Do(context.Background())
	price, _ := strconv.ParseFloat(curSymbol[0].Price, 64)
	if price < weightedAvgPrice{
		weightedAvgPrice = price
	}

	bs := []Bet{}
	rightSize := len(strings.Split(res[0].WeightedAvgPrice, ".")[1])
	for k, v := range config.NetRa {
		bp := round(weightedAvgPrice*(1-v/100), rightSize)
		sp := round(weightedAvgPrice*(1+v/100), rightSize)
		te := Bet{
			BuyPrice:  bp,
			SellPrice: sp,
			BuyAverage:bp,
			SellAverage:sp,
			Step:      0,
			Type:      k,
		}
		bs = append(bs, te)
	}
	p := TradeData{
		Bs:bs,
		SiL:SimBalance{
			Money:100,
			Coin:20,
		},
		Spend:5,
		SetupPrice:weightedAvgPrice,
		O:Ori{
			OriMoney:100,
			OriCoin:20,
		},
		LimitQ:2,
	}
	p.save()
}

