package core

import (
	"BandBinance/config"
	"context"
	"strconv"
)


func Hours24Tickers() float64 {
	res,err:=client.NewListPriceChangeStatsService().Symbol(config.Symbol).Do(context.Background())
	if err != nil{
		return 0
	}
	c,_:=strconv.ParseFloat( res[0].PriceChangePercent,64)
	return c
}


func InitSaveData()  {
	res,_:=client.NewListBookTickersService().Symbol(config.Symbol).Do(context.Background())
	re := res[0]
	bp,_ := strconv.ParseFloat(re.AskPrice,64)
	sp,_ := strconv.ParseFloat(re.BidPrice,64)

	p := PriceData{
		E: Bet{
			BuyPrice:  bp,
			SellPrice:sp,
			AverageBuy: bp,
			Step: 0,
		},
		SiL: SimulateBalance{
			Money: 100,
			Coin: 21,
		},
		P: Profit{
			SellRa: 0.7,
			BuyRa: 0.7,
			Quantity: 3,
			StartPrice: sp,
		},
	}
	p.save()

}