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
	cfg       *config.Config
	client    *firestore.Client
	batch     *firestore.WriteBatch
	queueSize uint8
}

func NewStore(ctx context.Context, cfg *config.Config) *Store {
	var (
		cred option.ClientOption = nil
		app  *firebase.App
		err  error
	)
	fbCfg := &firebase.Config{ProjectID: cfg.App.GcpID}
	if cfg.App.Development > 0 {
		cred = option.WithCredentialsFile("credentials.json")
	}
	app, err = firebase.NewApp(ctx, fbCfg, cred)
	if err != nil {
		log.WithError(err).Panic()
	}

	client, err := app.Firestore(ctx)
	if err != nil {
		log.WithError(err).Panic()
	}
	batch := client.Batch()
	return &Store{cfg, client, batch, 0}
}

func (s *Store) Disconnect() {
	if err := s.client.Close(); err != nil {
		log.WithError(err).Panic()
	}
}

func (s *Store) PushRecord(high, low *broker.Pair, value float64) {
	ref := s.client.Collection(s.cfg.Database.Collection).NewDoc()
	s.batch.Set(ref, map[string]interface{}{
		"high":      *high,
		"low":       *low,
		"value":     value,
		"timestamp": time.Now(),
	})
	s.queueSize++
}

func (s *Store) Commit(ctx context.Context) error {
	var err error = nil
	if s.queueSize > 10 {
		_, err = s.batch.Commit(ctx)
		s.queueSize = 0
		s.batch = s.client.Batch()
	}
	return err
}
