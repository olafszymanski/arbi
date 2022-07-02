package binance

import (
	"errors"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/broker"
	log "github.com/sirupsen/logrus"
)

type PricesWebsocket struct {
	cfg    *config.Config
	conn   *websocket.Conn
	symbol string
}

func NewPricesWebsocket(cfg *config.Config, symbol string) *PricesWebsocket {
	conn, _, err := websocket.DefaultDialer.Dial(websocketPricesUrl(symbol), nil)
	if err != nil {
		log.WithError(err).Panic()
	}

	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte("pong"), time.Now().Add(5*time.Second))
	})

	return &PricesWebsocket{cfg, conn, symbol}
}

func (w *PricesWebsocket) Read() broker.Price {
	type tempPrice struct {
		Symbol string `json:"s"`
		Price  string `json:"c"`
	}

	var tmpPrice tempPrice
	if err := w.conn.ReadJSON(&tmpPrice); err != nil {
		if errors.Is(err, syscall.ECONNRESET) {
			log.Warn("Websocket '", w.symbol, "' disconnected, retrying...")
			w.Reconnect()
			return w.Read()
		} else {
			log.WithError(err).Panic()
		}
	}
	return broker.Price{
		Symbol: tmpPrice.Symbol,
		Price:  stf64(tmpPrice.Price),
	}
}

func (w *PricesWebsocket) Close() {
	w.conn.Close()
}

func (w *PricesWebsocket) Reconnect() {
	conn, _, err := websocket.DefaultDialer.Dial(websocketPricesUrl(w.symbol), nil)
	if err != nil {
		log.WithError(err).Panic()
	}

	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte("pong"), time.Now().Add(5*time.Second))
	})
	w.conn = conn
}

type UserDataWebsocket struct {
	cfg  *config.Config
	conn *websocket.Conn
	key  string
}

func NewUserDataWebsocket(cfg *config.Config, key string) *UserDataWebsocket {
	conn, _, err := websocket.DefaultDialer.Dial(websocketUserDataUrl(key), nil)
	if err != nil {
		log.WithError(err).Panic()
	}
	return &UserDataWebsocket{cfg, conn, key}
}

func (w *UserDataWebsocket) Read() []broker.Balance {
	type tempBalance struct {
		Asset  string `json:"a"`
		Amount string `json:"f"`
	}
	type tempUpdateInfo struct {
		Type     string        `json:"e"`
		Time     uint64        `json:"E"`
		Balances []tempBalance `json:"B"`
	}

	var tmpInfo tempUpdateInfo
	if err := w.conn.ReadJSON(&tmpInfo); err != nil {
		if errors.Is(err, syscall.ECONNRESET) {
			log.Warn("User Data Websocket disconnected, retrying...")
			w.Reconnect()
			return w.Read()
		} else {
			log.WithError(err).Panic()
		}
	}

	if tmpInfo.Type == "outboundAccountPosition" {
		var bal []broker.Balance
		for _, b := range tmpInfo.Balances {
			bal = append(bal, broker.Balance{
				Asset:  b.Asset,
				Amount: stf64(b.Amount),
			})
		}
		return bal
	}
	return nil
}

func (w *UserDataWebsocket) Close() {
	w.conn.Close()
}

func (w *UserDataWebsocket) Reconnect() {
	conn, _, err := websocket.DefaultDialer.Dial(websocketUserDataUrl(w.key), nil)
	if err != nil {
		log.WithError(err).Panic()
	}
	w.conn = conn
}
