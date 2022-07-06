package binance

import (
	"errors"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
)

type jsonOrderBookTicker struct {
	Symbol      string `json:"s"`
	Bid         string `json:"b"`
	BidQuantity string `json:"B"`
	Ask         string `json:"a"`
	AskQuantity string `json:"A"`
}

type Websocket struct {
	connection *websocket.Conn
	factory    *URLFactory
}

func NewWebsocket(factory *URLFactory) (*Websocket, error) {
	c, _, err := websocket.DefaultDialer.Dial(factory.OrderBookTickers(), nil)
	if err != nil {
		return nil, err
	}
	c.SetPingHandler(func(appData string) error {
		return c.WriteControl(websocket.PongMessage, []byte(nil), time.Now().Add(5*time.Second))
	})
	return &Websocket{c, factory}, nil
}

func (w *Websocket) Read() (*jsonOrderBookTicker, error) {
	var o jsonOrderBookTicker
	if err := w.connection.ReadJSON(&o); err != nil {
		if errors.Is(err, syscall.ECONNRESET) || websocket.IsUnexpectedCloseError(err) {
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

func (w *Websocket) Close() {
	w.connection.Close()
}

func (w *Websocket) reconnect() error {
	c, _, err := websocket.DefaultDialer.Dial(w.factory.OrderBookTickers(), nil)
	if err != nil {
		return err
	}
	w.connection = c
	return nil
}
