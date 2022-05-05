package main

import (
	"BandBinance/analyse"
	"BandBinance/config"
	"BandBinance/gather"
	"BandBinance/store"
	"BandBinance/strategy"
	"log"
	"time"
)

func main() {
	var num int64 = 0
	ticker := time.NewTicker(time.Duration(config.Run.Sleep) * time.Second)
	defer ticker.Stop()

	for {
		// 采集数据
		err := gather.GatherData()
		if err != nil {
			log.Printf("%s", err)
			return
		}

		data, err := store.LoadData()
		if err != nil {
			log.Printf("%s", err)
			return
		}
		// 数据回测趋势
		scene, err := analyse.Scene(data)
		if err != nil {
			log.Printf("%s", err)
			return
		}

		// 趋势选择策略
		routeStrategy, err := strategy.RouteStrategy(scene)
		if err != nil {
			log.Printf("%s", err)
			return
		}

		// 策略推算交易
		routeStrategy.Process()

		// 交易收益分析
		err = analyse.Profit()
		if err != nil {
			log.Printf("%s", err)
			return
		}
		num++
		log.Printf("第 %d 轮结束, 休眠 %d s", num, config.Run.Sleep)
		<-ticker.C
	}

}
