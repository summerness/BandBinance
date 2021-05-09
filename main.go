package main

import (
	"BandBinance/core"
	"BandBinance/data"
)

func main() {
	core.InitSaveData()
	data.DelAll()
	p := core.InitPriceData()
	for{
		p.ToLimitTrade()
	}
}