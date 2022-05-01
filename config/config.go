package config

import (
	"github.com/spf13/viper"
)

var Run *RunConfig
var Auth *AuthConfig
var Proxy *ProxyConfig
var Notify *NotifyConfig

type RunConfig struct {
	BenefitSleep     int // 更新卖单状态间隔(s)
	Sleep            int // 购买间隔
	SellSleep        int // 卖出检查间隔
	CancelOrderSleep int
}

type AuthConfig struct {
	ApiKey    string
	SecretKey string
}

type ProxyConfig struct {
	Switch bool
	Path   string
}

type NotifyConfig struct {
	FeiShu struct {
		Url string
	}
	DD struct {
		Url    string
		Secret string
	}
}

func init() {
	vp := viper.New()
	vp.AddConfigPath("config")
	vp.SetConfigType("yml")
	vp.SetConfigName("config") // 可以不指定文件名

	err := vp.ReadInConfig()
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}

	err = vp.UnmarshalKey("Run", &Run)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}
	err = vp.UnmarshalKey("Auth", &Auth)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}
	err = vp.UnmarshalKey("Proxy", &Proxy)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}
	err = vp.UnmarshalKey("Notify", &Notify)
	if err != nil {
		panic(`读取配置文件时出现错误:` + err.Error())
	}

}
