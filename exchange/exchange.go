package exchange

import (
	"BandBinance/config"
	"BandBinance/domain"
	"BandBinance/notify"
	"BandBinance/util"
	"context"
	"errors"
	"fmt"
	"github.com/adshao/go-binance/v2"
	"github.com/adshao/go-binance/v2/delivery"
	"github.com/adshao/go-binance/v2/futures"
	"log"
	"strconv"
	"sync"
	"time"
)

var b = NewBiance()
var priceMap = make(map[string]float64)
var balanceMap = make(map[string]float64)
var lockPriceMap sync.RWMutex
var lockBalanceMap sync.RWMutex
var Order iOrder = &orderImpl{}

type iOrder interface {
	Cancel(symbol string, orderId int64) error
}

type orderImpl struct {
}

func (o *orderImpl) Cancel(symbol string, orderId int64) error {
	_, err := b.bc.NewCancelOrderService().Symbol(symbol).
		OrderID(orderId).Do(context.Background())

	return err
}

// Biance  https://github.com/adshao/go-binance
type Biance struct {
	bc *binance.Client
	fc *futures.Client
	dc *delivery.Client
}

func NewBiance() *Biance {
	client := binance.NewClient(config.AK, config.SK)
	futuresClient := binance.NewFuturesClient(config.AK, config.SK)   // USDT-M Futures
	deliveryClient := binance.NewDeliveryClient(config.AK, config.SK) // Coin-M Futures
	biance := Biance{bc: client, fc: futuresClient, dc: deliveryClient}
	return &biance
}

// LoadPrice 获取最新价格
func LoadPrice(symbol string) (float64, error) {
	lockPriceMap.Lock()
	defer lockPriceMap.Unlock()
	var res float64
	res = priceMap[symbol]
	if res == 0 {
		sprintf := fmt.Sprintf("没有此价格, 建议换币种, %s", symbol)
		notify.FeiShu.DoNotify(sprintf)
		return 0, errors.New(sprintf)
	}
	return res, nil
}

// GetPrice 获取最新价格
func GetPrice() error {
	lockPriceMap.Lock()
	defer lockPriceMap.Unlock()
	err := util.Retry(10000, 1000, func() error {
		prices, err := b.bc.NewListPricesService().Do(context.Background())
		if err != nil {
			return err
		}
		for _, p := range prices {
			price, err := strconv.ParseFloat(p.Price, 64)
			if err != nil {
				return err
			}
			priceMap[p.Symbol] = price
		}
		return nil
	})
	return err
}

// GetOrder 获取订单
func GetOrder(orderId int64, symbol string) (domain.TradeOrder, error) {
	var order domain.TradeOrder
	err := util.Retry(100, 100, func() error {
		response, err := b.bc.NewGetOrderService().Symbol(symbol).
			OrderID(orderId).Do(context.Background())
		if err != nil {
			return err
		}
		order, err = convert(response)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return domain.TradeOrder{}, err
	}
	return order, nil
}

// CreateOrder 创建订单
func CreateOrder(quantity float64, price float64, sideType binance.SideType, clientId string, symbol string, retryTimes int, sleep float64) (int64, error) {
	var orderId int64
	err := util.Retry(retryTimes, sleep, func() error {
		q := strconv.FormatFloat(quantity, 'f', 2, 64)
		p := strconv.FormatFloat(price, 'f', 2, 64)
		response, err := b.bc.NewCreateOrderService().Symbol(symbol).
			Side(sideType).Type(binance.OrderTypeLimit).
			TimeInForce(binance.TimeInForceTypeGTC).Quantity(q).
			Price(p).NewClientOrderID(clientId).Do(context.Background())

		if err != nil {
			return err
		}
		log.Printf("创建订单成功, 订单id=%d", response.OrderID)
		orderId = response.OrderID
		return nil
	})

	if err != nil {
		return orderId, err
	}
	return orderId, nil
}

// 处理重复创建订单, 实际成功, 但是没入库, 需要从数据库查询出来,
func processDuplicateOrderSentError(price float64, sideType binance.SideType, clientId string, symbol string) (domain.TradeOrder, error) {
	//
	var res domain.TradeOrder
	parseInt, err := strconv.ParseInt(clientId, 10, 64)
	if err != nil {
		return res, err
	}
	orders, err := b.bc.NewListOrdersService().Symbol(symbol).StartTime(parseInt).EndTime(time.Now().UnixMilli()).
		Do(context.Background())
	if err != nil {
		return domain.TradeOrder{}, err
	}
	for _, order := range orders {
		if order.ClientOrderID == clientId {
			res = domain.TradeOrder{
				OrderId:   order.OrderID,
				TradeType: string(sideType),
				Status:    string(binance.OrderStatusTypeNew),
				BuyPrice:  price,
			}
			return res, nil
		}
	}

	return res, errors.New(fmt.Sprintf("找不到重复订单, 人工check, clientId = %s", clientId))
}

func convert(order *binance.Order) (domain.TradeOrder, error) {

	buyPrice, err := strconv.ParseFloat(order.Price, 64)
	quantity, err := strconv.ParseFloat(order.OrigQuantity, 64)

	if err != nil {
		return domain.TradeOrder{}, err
	}

	return domain.TradeOrder{
		Symbol:          order.Symbol,
		OrderId:         order.OrderID,
		Status:          string(order.Status),
		BuyPrice:        buyPrice,
		BuySuccessPrice: buyPrice,
		DealTime:        time.Unix(order.UpdateTime, 0).UnixMilli(),
		Quantity:        quantity,
	}, nil
}

func UpdateBalance() error {
	lockBalanceMap.Lock()
	defer lockBalanceMap.Unlock()
	response, err := b.bc.NewGetAccountService().Do(context.Background())
	if err != nil {
		return err
	}
	balances := response.Balances
	if len(balances) > 1 {
		for i := range balances {
			free, err := strconv.ParseFloat(balances[i].Free, 64)
			if err != nil {
				return err
			}
			balanceMap[balances[i].Asset] = free
		}
	}
	return nil
}

func GetBalance(symbol string) (float64, error) {
	lockBalanceMap.Lock()
	defer lockBalanceMap.Unlock()
	return balanceMap[symbol], nil
}
