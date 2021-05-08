package config

var (
	DataPath  = "./data/data.json"
	ApiKey    = "Bu3dG7d8vsxr9EdCFNfJRZ6QN53Ujvw7g97KOqsQRgcqoueXhRSO3eFVvZD0k1Qy"
	SecretKey = "JtiLkP4wn83s68o9RHaKxIt5LuWYrKnNBZ42gXUbZPoaMsnVT5Duj6pmJvcBoPyl"
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
