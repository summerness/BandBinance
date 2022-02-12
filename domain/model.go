package domain

import (
	"gorm.io/gorm"
)

// Scene 场景枚举
type Scene int

const (
	// GRID 网格场景
	GRID Scene = 1
)

type History struct {
}

// Grid 网格
type Grid struct {
	// 币种
	Symbol string
	// 收益率
	ROI float64
	// 网格数量, 表示下一次交易的数量, 比如币就是买入多少金额的币
	Amount float64
	// 压力位, 全部卖出的价格
	Stress float64
	// 阻力位, 禁止买入的价格
	Resistance float64
	// 当前价格
	Price float64
}

// TradeOrder 交易订单
type TradeOrder struct {
	gorm.Model
	// 币种
	Symbol string
	// 订单id
	OrderId int64
	// 交易类型, 买入, 卖出
	TradeType string
	// 订单状态
	Status string
	// 交易网格index
	Index int
	// 交易网格版本
	Version int
	// 挂单价格
	BuyPrice float64
	// 买入价格
	BuySuccessPrice float64
	// 买入数量
	Quantity float64
	// 创建时间
	CreateTime int64
	// 成交时间
	DealTime int64
	// 撤销时间
	CancelTime int64
	// clientId
	ClientId int64
}

type GridTrade struct {
	gorm.Model
	Id string
	// symbol
	Symbol string
	// 价格下界
	HighPrice float64
	// 价格下界
	LowPrice float64
	//
	Quantity float64
	// index, 网格区间位置, 每个Grid都有固定位置
	Index int
	// 版本
	Version int
}

type GridSymbolConfig struct {
	gorm.Model
	// 币种
	Symbol string
	// 压力位
	Stress float64
	// 阻力位
	Resistance float64
	// 收益率
	ROI float64
	// 每单金额
	Amount float64
	// 策略版本
	Version int
	// 休眠 单位:s
	Sleep int
	// 网络请求重试次数
	RetryTimes int
	// 网络请求重试时间间隔,单位:ms
	RetryGap int
	// 启用状态
	Enable bool
}

func (t TradeOrder) GetSpend() float64 {
	return t.BuyPrice * t.Quantity
}
