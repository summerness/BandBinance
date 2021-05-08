package data

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"time"
)

var db *sql.DB

func init() {
	db, _ = sql.Open("sqlite3", "data/trade.db")
}

func InsertOne(tradeType string, price, coin, real_price, real_coin float64, btype int) {
	stmt, _ := db.Prepare("INSERT INTO trade(trade_type, price, coin, real_price,real_coin,type) values(?,?,?,?,?,?)")
	stmt.Exec(tradeType, price, coin, real_price, real_coin, btype)
}

type Trade struct {
	Id        int
	TradeType string
	Price     float64
	Coin      float64
	IsDeal    int
	RealPrice float64
	RealCoin  float64
}

func UpdateDeal(dealprice float64) []Trade {
	tds := []Trade{}
	rows, err := db.Query(fmt.Sprintf("SELECT * FROM trade where is_deal=0 and ((trade_type = 'Buy' and price >= %f) or (trade_type = 'Sell' and price <= %f))", dealprice, dealprice))
	if err != nil {
		return tds
	}
	for rows.Next() {
		var id int
		var trade_type string
		var price float64
		var coin float64
		var real_price float64
		var real_coin float64
		var is_deal int
		var create_time time.Time
		var btype int
		rows.Scan(&id, &trade_type, &is_deal, &create_time, &price, &coin, &real_price, &real_coin,&btype)
		tds = append(tds, Trade{
			id,
			trade_type,
			price,
			coin,
			is_deal,
			real_price,
			real_coin,
		})
	}
	stmt, _ := db.Prepare("UPDATE trade set is_deal = 1 where is_deal=0 and ((trade_type = 'Buy' and price >= ?) or (trade_type = 'Sell' and price <= ?))")
	stmt.Exec(dealprice, dealprice)
	return tds
}
