package binance

import (
	"encoding/json"
	"errors"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olafszymanski/arbi/config"
	log "github.com/sirupsen/logrus"
)

type Websocket struct {
	cfg    *config.Config
	conn   *websocket.Conn
	symbol string
}

func NewWebsocket(cfg *config.Config, symbol string) *Websocket {
	conn, _, err := websocket.DefaultDialer.Dial(websocketUrl(symbol), nil)
	if err != nil {
		log.WithError(err).Panic()
	}

	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte("pong"), time.Now().Add(5*time.Second))
	})

	return &Websocket{cfg, conn, symbol}
}

func (w *Websocket) ReadPrice() Price {
	type tempPrice struct {
		Symbol string `json:"s"`
		Price  string `json:"c"`
	}

	_, data, err := w.conn.ReadMessage()
	if err != nil {
		if errors.Is(err, syscall.ECONNRESET) {
			log.Warn("Websocket '", w.symbol, "' disconnected, retrying...")
			w.Reconnect()
			return w.ReadPrice()
		} else {
			log.WithError(err).Panic()
		}
	}

	var tmpRes tempPrice
	if err := json.Unmarshal(data, &tmpRes); err != nil {
		log.WithError(err).Panic()
	}
	return Price{tmpRes.Symbol, stf64(tmpRes.Price)}
}

func (w *Websocket) Close() {
	w.conn.Close()
}

func (w *Websocket) Reconnect() {
	conn, _, err := websocket.DefaultDialer.Dial(websocketUrl(w.symbol), nil)
	if err != nil {
		log.WithError(err).Panic()
	}

	conn.SetPingHandler(func(appData string) error {
		return conn.WriteControl(websocket.PongMessage, []byte("pong"), time.Now().Add(5*time.Second))
	})

	w.conn = conn
}
