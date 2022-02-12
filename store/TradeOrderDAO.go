package store

import (
	"BandBinance/domain"
	"errors"
	"github.com/adshao/go-binance/v2"
	"gorm.io/gorm"
	"time"
)

var TradeOrder TradeOrderDAO = &TradeOrderDAOImpl{}

func init() {
	err := DB.AutoMigrate(&domain.GridSymbolConfig{})
	if err != nil {
		panic(err)

	}
}

type TradeOrderDAO interface {
	FindCompletedGrid() ([]int64, error)
	CalculateFutureBenefitByClientId(int64) (float64, error)
	CountClientId() (int, error)
	FindTradeOrderByClientIds(ids []int64) ([]domain.TradeOrder, error)
	FindBuyInNotSellTradeOrders() ([]domain.TradeOrder, error)
	FindCreatedNotBuy() ([]domain.TradeOrder, error)
	FindNewBuyOrderByCreatedTime(time time.Time) ([]domain.TradeOrder, error)
	Update(tx *gorm.DB, order *domain.TradeOrder) error
	CreateTradeOrder(tx *gorm.DB, d *domain.TradeOrder) error
}

type TradeOrderDAOImpl struct {
}

func (t *TradeOrderDAOImpl) CreateTradeOrder(db *gorm.DB, order *domain.TradeOrder) error {
	result := db.Create(order)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (t *TradeOrderDAOImpl) Update(tx *gorm.DB, order *domain.TradeOrder) error {
	result := tx.Save(order)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected < 1 {
		return errors.New("更新失败")
	}
	return nil
}

func (t *TradeOrderDAOImpl) FindNewBuyOrderByCreatedTime(time time.Time) ([]domain.TradeOrder, error) {
	var r []domain.TradeOrder
	result := DB.Where("created_at < ? and trade_type = ? and status = ?", time, binance.SideTypeBuy, binance.OrderStatusTypeNew).Find(&r).Limit(1)
	if result.Error != nil {
		return r, result.Error
	}
	if result.RowsAffected < 1 {
		return r, errors.New("找不到交易对")
	}
	return r, nil
}

func (t *TradeOrderDAOImpl) FindCreatedNotBuy() ([]domain.TradeOrder, error) {
	var r []domain.TradeOrder
	result := DB.Where("status = ? and trade_type =? ", binance.OrderStatusTypeNew, binance.SideTypeBuy).Find(&r)
	if result.Error != nil {
		return r, result.Error
	}
	if result.RowsAffected < 1 {
		return r, errors.New("找不到交易对")
	}
	return r, nil
}

// FindBuyInNotSellTradeOrders 买入但未卖出
func (t *TradeOrderDAOImpl) FindBuyInNotSellTradeOrders() ([]domain.TradeOrder, error) {
	var r []domain.TradeOrder

	var clientIds []string
	result := DB.Raw("select a.client_id from  ( select client_id from trade_orders where status != 'CANCELED'  group by client_id having count(1) > 1 ) a " +
		"left join " +
		"( select client_id from trade_orders where status = 'FILLED' group by client_id having count(1) > 1 ) b " +
		"on a.client_id = b.client_id " +
		"where b.client_id is null ").Scan(&clientIds)
	if result.Error != nil {
		return r, result.Error
	}
	if result.RowsAffected < 1 {
		return r, errors.New("找不到交易对")
	}
	result = DB.Where("client_id in ?", clientIds).Find(&r)
	if result.Error != nil {
		return r, result.Error
	}
	if result.RowsAffected < 1 {
		return r, errors.New("找不到交易对")
	}
	return r, nil
}

func (t *TradeOrderDAOImpl) FindTradeOrderByClientIds(ids []int64) ([]domain.TradeOrder, error) {
	var r []domain.TradeOrder
	result := DB.Where("client_id in ?", ids).Find(&r)
	if result.Error != nil {
		return r, result.Error
	}
	if result.RowsAffected < 1 {
		return r, errors.New("找不到数据")
	}
	return r, nil

}

func (t *TradeOrderDAOImpl) CountClientId() (int, error) {
	var r int
	result := DB.Raw("select count(1) from (select distinct(client_id) from trade_orders) a").Scan(&r)
	if result.Error != nil {
		return r, result.Error
	}
	if result.RowsAffected < 1 {
		return r, errors.New("找不到数据")
	}
	return r, nil
}

func (t *TradeOrderDAOImpl) FindCompletedGrid() ([]int64, error) {
	var r []int64
	result := DB.Raw("select client_id from trade_orders where status = ? group by client_id having count(1) > 1", binance.OrderStatusTypeFilled).Scan(&r)
	if result.Error != nil {
		return r, result.Error
	}
	if result.RowsAffected < 1 {
		return r, errors.New("找不到数据")
	}
	return r, nil
}

func (t *TradeOrderDAOImpl) CalculateFutureBenefitByClientId(clientId int64) (float64, error) {
	var r float64
	result := DB.Raw("select sum(spend) as r from (select (case when trade_type = 'BUY' then -buy_price else buy_price end) * quantity  as spend from trade_orders where client_id = ? ) a", clientId).
		Scan(&r)
	if result.Error != nil {
		return r, result.Error
	}
	if result.RowsAffected < 1 {
		return r, errors.New("找不到数据")
	}
	return r, nil

}
