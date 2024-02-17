package publisher

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"github.com/IBM/sarama"
	"github.com/broswen/webhookrelay/internal/repository"
	"github.com/broswen/webhookrelay/internal/retry"
	"github.com/rs/zerolog/log"
	"time"
)

// producer/publisher should
// start a transaction
// lock the top N earliest webhooks that haven't been published
// publish each webhook to kafka
// mark the webhook as published
// complete transaction
// repeat

// on publish failure, retry X times with expo backoff
// if backoff is hit, complete transaction and unlock rest of the rows

func NewWebhookPublisher(db *sql.DB, brokers []string, topic string) (*Publisher, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &Publisher{
		db:       db,
		producer: producer,
		topic:    topic,
	}, nil
}

type Publisher struct {
	db       *sql.DB
	producer sarama.SyncProducer
	topic    string
}

func (p *Publisher) Close() error {
	return p.producer.Close()
}

func (p *Publisher) tryProduce(key string, body []byte) error {
	_, err := retry.NewRetry(time.Millisecond*50, 3, func() (any, error, bool) {
		err := p.produce(key, body)
		if err != nil {
			log.Error().Err(err).Msgf("failed to produce message: %v", err)
		}
		return nil, err, true
	})()
	return err
}

func (p *Publisher) produce(key string, body []byte) error {
	_, _, err := p.producer.SendMessage(&sarama.ProducerMessage{
		Topic: p.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(body),
	})
	return err
}

func (p *Publisher) Run(ctx context.Context) error {
	ticker := time.NewTicker(time.Second)
	log.Debug().Msg("starting publish loop")
	for {
		select {
		case <-ticker.C:
			txn, err := p.db.BeginTx(ctx, &sql.TxOptions{})
			if err != nil {
				log.Error().Err(err).Msg("failed to open transaction")
				continue
			}

			whs, err := repository.NewSqlWebhookRepository(txn).LockForPublishing(ctx, 100)
			if err != nil {
				if !errors.Is(err, repository.ErrWebhookNotFound{}) {
					log.Error().Err(err).Msg("failed to lock for publishing")
				}
			}
			log.Debug().Msgf("locked %d rows for publishing", len(whs))

			for _, wh := range whs {
				b, err := json.Marshal(&wh)
				if err != nil {
					log.Error().Err(err).Str("id", wh.Id).Msg("failed to marshall webhook while publishing")
					continue
				}
				start := time.Now()
				err = p.tryProduce(wh.Id, b)
				if err != nil {
					log.Error().Err(err).Str("id", wh.Id).Msg("failed to publish webhook")
					PublishLatency.WithLabelValues("failure").Observe(float64(time.Since(start).Milliseconds()))
					PublishAttempts.WithLabelValues("failure").Inc()
					continue
				} else {
					PublishLatency.WithLabelValues("success").Observe(float64(time.Since(start).Milliseconds()))
					PublishAttempts.WithLabelValues("success").Inc()
				}

				if err := repository.NewSqlWebhookRepository(txn).MarkPublished(ctx, wh.Id); err != nil {
					log.Error().Err(err).Str("id", wh.Id).Msg("failed to mark webhook as published")
					continue
				}
			}

			err = txn.Commit()
			if err != nil {
				log.Error().Err(err).Msg("failed to commit transaction")
				continue
			}
		case <-ctx.Done():
			log.Info().Msg("publisher context done")
			return nil
		}
	}
}
