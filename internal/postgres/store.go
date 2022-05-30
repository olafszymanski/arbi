package postgres

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/olafszymanski/arbi/cmd/config"
)

type IStore interface {
	CreateRecord(lowSymbol, highSymbol string, lowPrice, highPrice float64, timestamp int64) error
}

type Store struct {
	binance *BinanceStore
}

func NewStore(cfg *config.Config) *Store {
	url := fmt.Sprintf("host=%s user=%s password=%s dbname=%s sslmode=disable", cfg.Database.Host, cfg.Database.Username, cfg.Database.Password, cfg.Database.Name)
	db, err := sqlx.Connect(cfg.Database.Driver, url)
	if err != nil {
		panic(err)
	}
	return &Store{
		binance: NewBinanceStore(db),
	}
}
