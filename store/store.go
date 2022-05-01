package store

import (
	"BandBinance/domain"
	"github.com/adshao/go-binance/v2"
	"github.com/pkg/errors"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"log"
	"strconv"
)

var DB = Open()

func Open() *gorm.DB {
	// open, err := gorm.Open(sqlite.Open("trade.db"), &gorm.Config{})
	dsn := "root:mCvw1SDpdccwSzlJ@tcp(127.0.0.1:30306)/tr?charset=utf8mb4&parseTime=True&loc=Local"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(`数据库连接错误`)
	}
	return db
}

func LoadData() (domain.History, error) {
	var data domain.History
	// 目前无历史数据
	return data, nil
}

func ExistNotBuyInGridTradeOrder(db *gorm.DB, symbol string, version int, index int) (bool, error) {
	var tradeOrder []domain.TradeOrder
	result := db.Where("symbol = ? and version = ? and `index` = ? and status = ?",
		symbol, version, index, string(binance.OrderStatusTypeNew)).
		Find(&tradeOrder)
	if result.Error != nil {
		return true, result.Error
	}

	if result.RowsAffected > 0 {
		return true, nil
	}

	return false, nil
}

func FindTradeOrder(symbol string, tradeType string, status string, version int) ([]domain.TradeOrder, error) {
	var tradeOrder []domain.TradeOrder
	result := DB.Where("symbol = ? and status = ? and trade_type = ? and version = ?", symbol, status, tradeType, version).Find(&tradeOrder)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected < 1 {
		return nil, errors.Wrapf(nil, "查找不到数据")
	}
	return tradeOrder, nil
}

func UpdateTradeOrderByOrderId(tx *gorm.DB, order domain.TradeOrder) (int, error) {
	save := tx.Model(&domain.TradeOrder{}).Where("order_id = ?", order.OrderId).Updates(order)
	if save.Error != nil {
		return 0, save.Error
	}

	return int(save.RowsAffected), nil
}

func FindNewTradeOrder(sideType binance.SideType) ([]domain.TradeOrder, error) {
	var tradeOrder []domain.TradeOrder
	result := DB.Where("status = ? and trade_type = ?", binance.OrderStatusTypeNew, sideType).Find(&tradeOrder)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected < 1 {
		return nil, errors.New("没有找到数据")
	}

	return tradeOrder, nil
}

func FindByClientId(clientId int64) ([]domain.TradeOrder, error) {
	var orders []domain.TradeOrder
	result := DB.Where("client_id = ?", clientId).Find(&orders)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected < 1 {
		return nil, errors.New("找不到数据")
	}
	return orders, nil
}

func SaveTrades(trades []domain.GridTrade) (int, error) {
	save := DB.Save(trades)
	if save.Error != nil {
		return 0, save.Error
	}
	return int(save.RowsAffected), nil
}

func FindGridById(id string) (domain.GridTrade, error) {
	var row domain.GridTrade
	result := DB.Where("id = ?", id).First(&row)
	if result.Error != nil {
		return row, result.Error
	}
	if result.RowsAffected < 1 {
		return row, errors.New("找不到数据")
	}
	return row, nil
}

func GenGridTradeId(symbol string, version int, index int) string {
	return symbol + strconv.Itoa(version) + strconv.Itoa(index)
}

func Begin() *gorm.DB {
	return DB.Begin()
}

func Tx(f func(tx *gorm.DB) error) error {
	return DB.Transaction(f)
}
