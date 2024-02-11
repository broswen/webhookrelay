package main

import (
	"context"
	"errors"
	"flag"
	"github.com/broswen/webhookrelay/internal/api/rest"
	"github.com/broswen/webhookrelay/internal/db"
	"github.com/broswen/webhookrelay/internal/repository"
	"github.com/broswen/webhookrelay/internal/service"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
)

var restApiAddress = ":8080"
var redisAddress = "redis:6379"
var postgresDSN = ""
var edgeAddress = ""
var edgeAccessId = ""
var edgeAccessSecret = ""

func main() {

	flag.StringVar(&restApiAddress, "apiAddr", os.Getenv("API_ADDR"), "rest api address")
	if restApiAddress == "" {
		log.Fatal().Msg("rest api address must be specified")
	}

	flag.StringVar(&redisAddress, "redisAddr", os.Getenv("REDIS_ADDR"), "redis address")
	if redisAddress == "" {
		log.Fatal().Msg("redis address must be specified")
	}

	flag.StringVar(&postgresDSN, "postgresDSN", os.Getenv("DSN"), "postgres connection DSN")
	if postgresDSN == "" {
		log.Fatal().Msg("postgres DSN must be specified")
	}

	flag.StringVar(&edgeAddress, "webhookdispatcherAddress", os.Getenv("WEBHOOKDISPATCHER_ADDRESS"), "address to the webhook dispatcher api")
	if edgeAddress == "" {
		log.Fatal().Msg("webhook dispatcher address must be specified")
	}
	flag.StringVar(&edgeAccessId, "webhookdispatcherAccessId", os.Getenv("ACCESS_ID"), "access id for the webhook dispatcher api")
	flag.StringVar(&edgeAccessSecret, "webhookdispatcherAccessSecret", os.Getenv("ACCESS_SECRET"), "access secret for the webhook dispatcher api")

	pool, err := db.InitDB(postgresDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating postgres pool")
	}

	idem, err := repository.NewRedisIdempotencyRepository(redisAddress)
	if err != nil {
		log.Error().Err(err).Msg("failed to create redis idempotency repository")
	}

	edge, err := repository.NewEdgeRepository(edgeAddress, edgeAccessId, edgeAccessSecret)
	if err != nil {
		log.Error().Err(err).Msg("failed to create edge repository")
	}

	webhooks, err := service.NewWebhookService(idem, pool, edge)
	if err != nil {
		log.Error().Err(err).Msg("failed to create webhook service")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	eg, gCtx := errgroup.WithContext(ctx)

	restServer := rest.Server{
		Webhooks: webhooks,
	}
	server := http.Server{
		Addr:    restApiAddress,
		Handler: restServer.Router(),
	}

	eg.Go(func() error {
		log.Debug().Msgf("rest api listening on %s", restApiAddress)
		if err := server.ListenAndServe(); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				return err
			}
		}
		return nil
	})

	eg.Go(func() error {
		select {
		case <-gCtx.Done():
			return server.Shutdown(context.Background())
		}
	})

	if err := eg.Wait(); err != nil {
		log.Error().Err(err).Msg("received unexpected errorgroup error")
	}
}
