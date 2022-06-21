package database

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go"
	"github.com/olafszymanski/arbi/config"
	"github.com/olafszymanski/arbi/internal/broker"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type Store struct {
	*firestore.Client
	collection string
}

func NewStore(ctx context.Context, cfg *config.Config) *Store {
	var (
		col string
		app *firebase.App
		err error
	)
	fbCfg := &firebase.Config{ProjectID: cfg.App.GcpID}
	if cfg.App.Development > 0 {
		cred := option.WithCredentialsFile("credentials.json")
		col = "records-dev"
		app, err = firebase.NewApp(ctx, fbCfg, cred)
	} else {
		col = "records"
		app, err = firebase.NewApp(ctx, fbCfg)
	}
	if err != nil {
		log.WithError(err).Panic()
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.WithError(err).Panic()
	}
	return &Store{client, col}
}

func (s *Store) Disconnect() {
	if err := s.Close(); err != nil {
		log.WithError(err).Panic()
	}
}

func (s *Store) AddRecord(ctx context.Context, high, low *broker.Pair, value float64) error {
	_, _, err := s.Collection(s.collection).Add(ctx, map[string]interface{}{
		"high":      *high,
		"low":       *low,
		"value":     value,
		"timestamp": time.Now(),
	})
	return err
}
