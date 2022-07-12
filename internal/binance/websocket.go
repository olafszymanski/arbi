package binance

import (
	"errors"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	log "github.com/sirupsen/logrus"
)

type jsonOrderBookTicker struct {
	Symbol      string `json:"s"`
	Bid         string `json:"b"`
	BidQuantity string `json:"B"`
	Ask         string `json:"a"`
	AskQuantity string `json:"A"`
}

type jsonBalance struct {
	Asset  string `json:"a"`
	Amount string `json:"f"`
}

type jsonAccountUpdate struct {
	Balances []jsonBalance `json:"B"`
}

type OrderBookWebsocket struct {
	connection *websocket.Conn
	factory    *URLFactory
}

func NewOrderBookWebsocket(factory *URLFactory) (*OrderBookWebsocket, error) {
	c, _, err := websocket.DefaultDialer.Dial(factory.OrderBookTickers(), nil)
	if err != nil {
		return nil, err
	}
	c.SetPingHandler(func(appData string) error {
		return c.WriteControl(websocket.PongMessage, []byte(nil), time.Now().Add(5*time.Second))
	})
	return &OrderBookWebsocket{c, factory}, nil
}

func (w *OrderBookWebsocket) Read() (*jsonOrderBookTicker, error) {
	var o jsonOrderBookTicker
	if err := w.connection.ReadJSON(&o); err != nil {
		if errors.Is(err, syscall.ECONNRESET) || err.Error() == "websocket: close 1006 (abnormal closure): unexpected EOF" {
			log.Warn("Order book websocket disconnected, trying to reconnect...")
			if err := w.reconnect(); err != nil {
				return nil, err
			}
			return w.Read()
		} else {
			return nil, err
		}
	}
	return &o, nil
}

func (w *OrderBookWebsocket) Close() {
	w.connection.Close()
}

func (w *OrderBookWebsocket) reconnect() error {
	c, _, err := websocket.DefaultDialer.Dial(w.factory.OrderBookTickers(), nil)
	if err != nil {
		return err
	}
	w.connection = c
	return nil
}

type WalletWebsocket struct {
	connection *websocket.Conn
	url        string
}

func NewWalletWebsocket(factory *URLFactory, listenKey string) (*WalletWebsocket, error) {
	u := factory.AccountUpdate(listenKey)
	c, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return nil, err
	}
	c.SetPingHandler(func(appData string) error {
		return c.WriteControl(websocket.PongMessage, []byte(nil), time.Now().Add(5*time.Second))
	})
	return &WalletWebsocket{c, u}, nil
}

func (w *WalletWebsocket) Read() ([]jsonBalance, error) {
	var o jsonAccountUpdate
	if err := w.connection.ReadJSON(&o); err != nil {
		if errors.Is(err, syscall.ECONNRESET) || err.Error() == "websocket: close 1006 (abnormal closure): unexpected EOF" {
			if err := w.reconnect(); err != nil {
				return nil, err
			}
			return w.Read()
		} else {
			return nil, err
		}
	}
	return o.Balances, nil
}

func (w *WalletWebsocket) Close() {
	w.connection.Close()
}

func (w *WalletWebsocket) reconnect() error {
	c, _, err := websocket.DefaultDialer.Dial(w.url, nil)
	if err != nil {
		return err
	}
	w.connection = c
	return nil
}
