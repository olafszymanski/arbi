package binance

import "sync"

type DataManager struct {
	data sync.Map
}

func NewDataManager(data sync.Map) *DataManager {
	return &DataManager{data}
}

func (m *DataManager) SymbolExists(key string) (Symbol, bool) {
	v, ok := m.data.Load(key)
	return v.(Symbol), ok
}

func (m *DataManager) FloatExists(key string) (float64, bool) {
	v, ok := m.data.Load(key)
	return v.(float64), ok
}

func (m *DataManager) LoadFloat(key string) float64 {
	v, _ := m.data.Load(key)
	return v.(float64)
}

func (m *DataManager) LoadSymbol(key string) Symbol {
	v, _ := m.data.Load(key)
	return v.(Symbol)
}

func (m *DataManager) StoreFloat(key string, value float64) {
	m.data.Store(key, value)
}

func (m *DataManager) StoreSymbol(key string, symbol Symbol) {
	m.data.Store(key, symbol)
}
