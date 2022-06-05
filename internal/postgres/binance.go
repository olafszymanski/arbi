package postgres

import "github.com/jmoiron/sqlx"

type BinanceStore struct {
	db *sqlx.DB
}

func NewBinanceStore(db *sqlx.DB) *BinanceStore {
	return &BinanceStore{db}
}

func (b *BinanceStore) CreateRecord(r Record) error {
	_, err := b.db.DB.Exec("INSERT INTO records (lowSymbol, lowPrice, highSymbol, highPrice, value, timestamp) VALUES ($1, $2, $3, $4, $5, $6)", r.Low.Symbol, r.Low.Price, r.High.Symbol, r.High.Price, r.Value, r.Timestamp)
	return err
}
