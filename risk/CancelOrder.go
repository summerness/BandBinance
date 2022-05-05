package risk

import (
	"BandBinance/domain"
	"BandBinance/exchange"
	"BandBinance/notify"
	"BandBinance/store"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/pkg/errors"
	"gorm.io/gorm"
	"log"
	"time"
)

var CancelOrder iCancelOrder = &cancelOrderProcessor{}

type iCancelOrder interface {
	ProcessCancel()
}

type cancelOrderProcessor struct {
}

func (c *cancelOrderProcessor) ProcessCancel() {
	// 找出3小时前的NEW买入单, 取消他
	orders, err := store.TradeOrder.FindNewBuyOrderByCreatedTime(time.Now().Add(-3 * time.Hour))
	if err != nil {
		log.Printf("获取订单, %+v", err)
		return
	}
	for i := range orders {
		doCancelOrder(&orders[i])
	}
}

func doCancelOrder(order *domain.TradeOrder) {
	// 开启事务
	err := store.Tx(func(tx *gorm.DB) error {
		// 更新数据库
		order.Status = string(binance.OrderStatusTypeCanceled)
		order.CancelTime = time.Now().UnixNano() / 1e6

		err := store.TradeOrder.Update(tx, order)
		if err != nil {
			return errors.Wrapf(err, "更新订单取消失败")
		}

		// 取消币安订单
		err = exchange.Order.Cancel(order.Symbol, order.OrderId)
		if err != nil {
			return errors.Wrap(err, "exchange 取消订单失败")
		}

		return nil
	})
	if err != nil {
		log.Printf("取消订单失败, %+v", err)
		return
	}
	msg := fmt.Sprintf("取消订单成功, id=%d, order_id = %d, symbol=%s", order.ID, order.OrderId, order.Symbol)
	log.Printf(msg)
	notify.DefaultNotify.Do(msg)
}
