package postgres

import "github.com/jmoiron/sqlx"

type BinanceStore struct {
	*sqlx.DB
}

func NewBinanceStore(db *sqlx.DB) *BinanceStore {
	return &BinanceStore{db}
}

func (b *BinanceStore) CreateRecord(lowSymbol, highSymbol string, lowPrice, highPrice float64, timestamp int64) error {
	return nil
}
