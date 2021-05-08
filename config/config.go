package config

var (
	DataPath  = "./data/data.json"
	ApiKey    = ""
	SecretKey = ""
	Symbol = "XRPUSDT"
	Fee = 0.1
	Simulate = true
	NetRa = map[int]float64{
		1:0.2,
		2:0.6,
		3:1.0,
		4:1.4,
		5:2.0,
	}
)
