package exchange

import (
	"github.com/adshao/go-binance/v2"
	"log"
	"strconv"
	"testing"
	"time"
)

func TestGetOrder(t *testing.T) {
	////order, err := GetOrder(293474365,)
	//if err != nil {
	//	return
	//}
	//log.Printf("%d", order.OrderId)
}

func TestCreateOrder(t *testing.T) {

	formatInt := strconv.FormatInt(time.Now().UnixMilli(), 10)
	order, err := CreateOrder(0.18, 60, binance.SideTypeBuy, formatInt, "AXSBUSD", 100, 100)
	if err != nil {
		return
	}
	log.Printf("%d", order)

	order, err = CreateOrder(0.18, 60, binance.SideTypeBuy, formatInt, "AXSBUSD", 100, 100)
	if err != nil {
		log.Printf("正常失败")
	}
	log.Printf("%d", order)

	order, err = CreateOrder(0.18, 80, binance.SideTypeSell, formatInt, "AXSBUSD", 100, 100)
	if err != nil {
		panic(err)
	}
	order, err = CreateOrder(0.18, 80, binance.SideTypeSell, formatInt, "AXSBUSD", 100, 100)
	if err != nil {
		log.Printf("正常失败")
	}
}

func Test_processDuplicateOrderSentError(t *testing.T) {

}
