package main

import (
	"BandBinance/config"
	"BandBinance/risk"
	"time"
)

func main() {
	for true {
		risk.CancelOrder.ProcessCancel()
		time.Sleep(time.Duration(config.CancelOrderSleep) * time.Second)
	}
}
