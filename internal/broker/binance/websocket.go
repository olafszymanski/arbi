package binance

import (
	"encoding/json"
	"time"

	"github.com/gorilla/websocket"
	"github.com/olafszymanski/arbi/config"
	log "github.com/sirupsen/logrus"
)

type Websocket struct {
	cfg  *config.Config
	conn *websocket.Conn
}

func NewWebsocket(cfg *config.Config, symbol string) *Websocket {
	conn, _, err := websocket.DefaultDialer.Dial(makeWebsocketUrl(symbol), nil)
	i := 0
	for err != nil {
		if i > cfg.App.MaxTimeouts {
			log.WithError(err).Panic()
		}
		time.Sleep(time.Duration(cfg.App.TimeoutInterval) * time.Second)
		conn, _, err = websocket.DefaultDialer.Dial(makeWebsocketUrl(symbol), nil)
		i++
	}
	return &Websocket{cfg, conn}
}

func (w *Websocket) Read() Result {
	_, data, err := w.conn.ReadMessage()
	i := 0
	for err != nil {
		if i > w.cfg.App.MaxTimeouts {
			log.WithError(err).Panic()
		}
		time.Sleep(time.Duration(w.cfg.App.TimeoutInterval) * time.Second)
		_, data, err = w.conn.ReadMessage()
		i++
	}

	var res Result
	if err := json.Unmarshal(data, &res); err != nil {
		log.WithError(err).Panic()
	}
	return res
}

func (w *Websocket) Close() {
	w.conn.Close()
}
