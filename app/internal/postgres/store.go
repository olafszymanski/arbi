package postgres

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/olafszymanski/arbi/app/config"
)

type RecordPair struct {
	Symbol string
	Price  float64
}

type Record struct {
	Low       RecordPair
	High      RecordPair
	Value     float64
	Timestamp time.Time
}

type IStore interface {
	CreateRecord(r Record) error
}

type Store struct {
	Binance *BinanceStore
}

func NewStore(cfg *config.Config) *Store {
	url := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Database.Host, cfg.Database.Username, cfg.Database.Password, cfg.Database.Name)
	db, err := sqlx.Connect(cfg.Database.Driver, url)
	if err != nil {
		panic(err)
	}
	return &Store{
		Binance: NewBinanceStore(db),
	}
}
