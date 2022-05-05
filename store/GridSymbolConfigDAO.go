package store

import (
	"BandBinance/domain"
	"errors"
)

var GridSymbolConfig GridSymbolConfigDAO = &GridSymbolConfigDAOImpl{}

func init() {
	err := DB.AutoMigrate(&domain.GridSymbolConfig{})
	if err != nil {
		panic(err)

	}
	err = DB.AutoMigrate(&domain.GridTrade{})
	if err != nil {
		panic(err)
	}
	err = DB.AutoMigrate(&domain.TradeOrder{})
	if err != nil {
		panic(err)
	}
}

type GridSymbolConfigDAO interface {
	FindEnable() ([]domain.GridSymbolConfig, error)
}

type GridSymbolConfigDAOImpl struct {
}

func (g *GridSymbolConfigDAOImpl) FindEnable() ([]domain.GridSymbolConfig, error) {
	var data []domain.GridSymbolConfig

	result := DB.Where("enable = ?", true).Find(&data)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected < 1 {
		return nil, errors.New("can not find grid symbol config")
	}
	return data, nil
}
