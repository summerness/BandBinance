module BandBinance

go 1.16

replace (
	xiaofei-tool => ../xiaofei-tool
)
require (
	github.com/adshao/go-binance/v2 v2.3.5
	github.com/pkg/errors v0.9.1
	github.com/spf13/viper v1.11.0
	gorm.io/driver/mysql v1.2.3
	gorm.io/gorm v1.22.4
	xiaofei-tool v0.0.0
)
