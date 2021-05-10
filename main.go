package main

import (
	"BandBinance/core/simulate"
	"BandBinance/data"
)

func main() {
	simulate.InitSaveData()
	data.DelAll()
	p := simulate.InitPriceData()
	for{
		p.ToLimitTrade()
	}
}