package service

import (
	"context"
	"database/sql"
	"errors"
	"github.com/broswen/webhookrelay/internal/model"
	"github.com/broswen/webhookrelay/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"net/url"
	"time"
)

type Webhook interface {
	Get(ctx context.Context, id string) (model.Webhook, error)
	List(ctx context.Context, deleted bool, offset int64, limit int64) ([]model.Webhook, error)
	Create(ctx context.Context, req CreateWebhookRequest) (model.Webhook, error)
}

func NewWebhookService(idempotency repository.Idempotency, db *sql.DB, edge repository.Edge) (Webhook, error) {
	return &WebhookService{
		idem: idempotency,
		edge: edge,
		db:   db,
	}, nil
}

type WebhookService struct {
	idem repository.Idempotency
	edge repository.Edge
	db   *sql.DB
}

func (s *WebhookService) Get(ctx context.Context, id string) (model.Webhook, error) {
	wh, err := repository.NewSqlWebhookRepository(s.db).Get(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("failed to get webhook")
		return model.Webhook{}, err
	}
	//TODO merge metadata with webhook model
	_, err = s.edge.Get(ctx, id)
	if err != nil {
		log.Error().Err(err).Str("id", id).Msg("failed to get webhook from edge")
	}
	return wh, nil
}

func (s *WebhookService) List(ctx context.Context, deleted bool, offset int64, limit int64) ([]model.Webhook, error) {

	whs, err := repository.NewSqlWebhookRepository(s.db).List(ctx, deleted, offset, limit)
	if err != nil {
		log.Error().Err(err).Int64("offset", offset).Int64("limit", limit).Bool("deleted", deleted).Msg("failed to list webhooks")
	}
	for _, wh := range whs {
		//TODO merge metadata with webhook model
		_, err = s.edge.Get(ctx, wh.Id)
		if err != nil {
			log.Error().Err(err).Str("id", wh.Id).Msg("failed to get webhook from edge")
		}
	}

	return whs, nil
}

type CreateWebhookRequest struct {
	IdempotencyToken string `json:"idempotencyToken"`
	Target           string `json:"target"`
	Payload          []byte `json:"payload"`
}

func (r CreateWebhookRequest) Validate() error {
	if _, err := url.ParseRequestURI(r.Target); err != nil {
		return err
	}
	if r.IdempotencyToken == "" {
		return errors.New("idempotency token must not be empty")
	}
	return nil
}

func (s *WebhookService) Create(ctx context.Context, req CreateWebhookRequest) (model.Webhook, error) {
	if err := req.Validate(); err != nil {
		return model.Webhook{}, err
	}
	//check if a request for this token already exists and return the results if it does
	id, err := s.idem.Get(ctx, req.IdempotencyToken)
	if err != nil {
		if !errors.Is(err, redis.Nil) {
			log.Error().Err(err).Msg("failed to check idempotency token")
		}
	} else {
		// if token exists in cache, check it's not associated to an in-progress request
		if id == "__inprogress" {
			return model.Webhook{}, errors.New("request for existing token in progress")
		}

		// else return the previously created webhook
		if id != "" {
			log.Debug().Str("id", id).Str("token", req.IdempotencyToken).Msg("previous request completed")
			return repository.NewSqlWebhookRepository(s.db).Get(ctx, id)
		}
	}

	err = s.idem.Set(ctx, req.IdempotencyToken, "__inprogress", time.Second*30)
	if err != nil {
		log.Error().Err(err).Msg("failed to bookmark idempotency token")
	}

	wh, err := repository.NewSqlWebhookRepository(s.db).Create(ctx, req.Target, req.Payload)
	if err != nil {
		log.Error().Err(err).Msg("failed to create webhook")
		return model.Webhook{}, err
	}

	err = s.idem.Set(ctx, req.IdempotencyToken, wh.Id, time.Minute*10)
	if err != nil {
		log.Error().Err(err).Msg("failed to finalize idempotency token")
	}

	return wh, nil
}
