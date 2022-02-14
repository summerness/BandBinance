package strategy

import (
	"BandBinance/domain"
)

// RouteStrategy 路由策略
func RouteStrategy(scene domain.Scene) (domain.Strategy, error) {
	var strategy domain.Strategy
	switch scene {
	case domain.GRID:
		gridStrategy, err := newGridStrategy()
		if err != nil {
			return nil, err
		}
		strategy = gridStrategy
		return strategy, nil
	default:
		panic("没有策略")
	}
}
