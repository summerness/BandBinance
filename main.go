package main

import (
	"BandBinance/core"
)

func main() {
	core.InitSaveData()
	p := core.InitPriceData()
	for{
		p.ToTrade()
	}
}