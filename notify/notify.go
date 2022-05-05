package notify

import (
	"BandBinance/domain"
)

var DefaultNotify iNotify = &DDNotify{}

type iNotify interface {
	Do(sprintf string)
	Trade(trade domain.TradeOrder)
}
