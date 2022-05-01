package domain

import (
	"gorm.io/gorm"
)

// Scene 场景枚举
type Scene int

const (
	GRID Scene = 1 // GRID 网格场景
)

type History struct {
}

// Grid 网格  像是一个中间产物  已经删除了对其的调用  日后若不使用可删除
type Grid struct {
	Symbol     string  // 币种
	ROI        float64 // 收益率
	Amount     float64 // 网格数量, 表示下一次交易的数量, 比如币就是买入多少金额的币
	Stress     float64 // 压力位, 全部卖出的价格
	Resistance float64 // 阻力位, 禁止买入的价格
	Price      float64 // 当前价格
}

// TradeOrder 交易订单
type TradeOrder struct {
	gorm.Model
	Symbol          string  // 币种
	OrderId         int64   // 订单id
	TradeType       string  // 交易类型, 买入, 卖出
	Status          string  // 订单状态
	Index           int     // 交易网格index
	Version         int     // 交易网格版本
	BuyPrice        float64 // 挂单价格
	BuySuccessPrice float64 // 买入价格
	Quantity        float64 // 买入数量
	CreateTime      int64   // 创建时间
	DealTime        int64   // 成交时间
	CancelTime      int64   // 撤销时间
	ClientId        int64   // clientId
}

func (t TradeOrder) GetSpend() float64 {
	return t.BuyPrice * t.Quantity
}

type GridTrade struct {
	gorm.Model
	Id        string
	Symbol    string  // symbol
	HighPrice float64 // 价格上限
	LowPrice  float64 // 价格下界
	Quantity  float64 // 数量
	Index     int     // index, 网格区间位置, 每个Grid都有固定位置
	Version   int     // 版本
}

type GridSymbolConfig struct {
	gorm.Model
	Symbol     string  // 币种
	Stress     float64 // 压力位
	Resistance float64 // 阻力位
	ROI        float64 // 收益率
	Amount     float64 // 每单金额
	Version    int     // 策略版本
	Sleep      int     // 休眠 单位:s
	RetryTimes int     // 网络请求重试次数
	RetryGap   int     // 网络请求重试时间间隔,单位:ms
	Enable     bool    // 启用状态
}
