package main

import (
	"BandBinance/config"
	"BandBinance/risk"
	"time"
)

func main() {
	t := time.NewTicker(time.Duration(config.Run.CancelOrderSleep) * time.Second)
	for {
		risk.CancelOrder.ProcessCancel()
		<-t.C
	}
}
