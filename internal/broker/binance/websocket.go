package binance

import (
	"encoding/json"

	"github.com/gorilla/websocket"
	"github.com/olafszymanski/arbi/config"
	log "github.com/sirupsen/logrus"
)

type Websocket struct {
	conn *websocket.Conn
}

func NewWebsocket(cfg *config.Config, symbol string) *Websocket {
	conn, _, err := websocket.DefaultDialer.Dial(makeWebsocketUrl(symbol), nil)
	if err != nil {
		log.WithError(err).Panic()
	}
	return &Websocket{conn}
}

func (b *Websocket) Read() Result {
	_, data, err := b.conn.ReadMessage()
	if err != nil {
		log.WithError(err).Panic()
	}

	var res Result
	if err := json.Unmarshal(data, &res); err != nil {
		log.WithError(err).Panic()
	}
	return res
}

func (b *Websocket) Close() {
	b.conn.Close()
}
