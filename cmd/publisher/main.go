package main

import (
	"context"
	"errors"
	"flag"
	"github.com/broswen/webhookrelay/internal/db"
	"github.com/broswen/webhookrelay/internal/publisher"
	"github.com/go-chi/chi/v5"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
	"net/http"
	"os"
	"os/signal"
	"strings"
)

var metricsAddress = ":8081"
var postgresDSN = ""
var brokers = ""
var topic = ""

func main() {

	flag.StringVar(&metricsAddress, "metricsAddr", os.Getenv("METRICS_ADDR"), "metrics server address")
	if metricsAddress == "" {
		log.Fatal().Msg("metrics address must be specified")
	}

	flag.StringVar(&postgresDSN, "postgresDSN", os.Getenv("DSN"), "postgres connection DSN")
	if postgresDSN == "" {
		log.Fatal().Msg("postgres DSN must be specified")
	}

	flag.StringVar(&brokers, "brokers", os.Getenv("BROKERS"), "kafka brokers")
	if brokers == "" {
		log.Fatal().Msg("kafka brokers must be specified")
	}

	flag.StringVar(&topic, "topic", os.Getenv("TOPIC"), "kafka topic")
	if topic == "" {
		log.Fatal().Msg("kafka topic must be specified")
	}

	pool, err := db.InitDB(postgresDSN)
	if err != nil {
		log.Fatal().Err(err).Msg("error creating postgres pool")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	eg, gCtx := errgroup.WithContext(ctx)

	producer, err := publisher.NewWebhookPublisher(pool, strings.Split(brokers, ","), topic)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create publisher")
	}
	defer producer.Close()

	eg.Go(func() error {
		return producer.Run(gCtx)
	})

	eg.Go(func() error {
		r := chi.NewRouter()
		r.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(metricsAddress, r); err != nil {
			if !errors.Is(err, http.ErrServerClosed) {
				return err
			}
		}
		return nil
	})

	if err := eg.Wait(); err != nil {
		log.Error().Err(err).Msg("received unexpected errorgroup error")
	}

}
