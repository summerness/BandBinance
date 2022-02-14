package notify

import (
	"BandBinance/config"
	"BandBinance/domain"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

var FeiShu iNotify = &feiShuNotify{}

type iNotify interface {
	DoNotify(sprintf string)
	NotifyTrade(trade domain.TradeOrder)
}

type feiShuNotify struct {
}

func (f *feiShuNotify) NotifyTrade(trade domain.TradeOrder) {
	sprintf := fmt.Sprintf("订单创建成功,id= %d, %s %f %f %s", trade.ID, trade.TradeType, trade.Quantity, trade.BuyPrice, trade.Symbol)
	FeiShu.DoNotify(sprintf)
}

func (f *feiShuNotify) DoNotify(msg string) {
	data := make(map[string]interface{})
	data["msg_type"] = "text"
	data["content"] = map[string]string{"text": msg}
	dataJson, err := json.Marshal(data)
	if err != nil {
		// 构造参数失败
		log.Printf("构造参数失败 %+v", err)
		return
	}

	// 发起请求
	resp, err := http.Post(config.FeiShuAlarmUrl, "application/json;UTF-8", bytes.NewReader(dataJson))
	if err != nil {
		log.Printf("请求飞书失败, %+v", err)
		return
	}
	// 关闭http流
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Printf("关闭http流失败 %+v", err)
			return
		}
	}(resp.Body)

	// 读取返回
	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取失败 %+v", err)
		return
	}

	return
}
