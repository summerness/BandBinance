package domain

type Strategy interface {
	// Process  计算下一个网格的交易
	Process()
}
