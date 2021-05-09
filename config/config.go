package config

var (
	DataPath  = "./data/data.json"
	ApiKey    = ""
	SecretKey = ""
	Symbol = "XRPUSDT"
	Fee = 0.1
	Simulate = true
	NetRa = map[int]float64{
		1:0.3,
		2:0.4,
		3:0.5,
		4:0.6,
		5:0.7,
	}
	Earn = 3.00
	SaveRa = 4.0
	MaxStep = 7
)
