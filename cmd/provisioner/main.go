package main

import (
	"context"
	"errors"
	"flag"
	"github.com/IBM/sarama"
	"github.com/broswen/webhookrelay/internal/provisioner"
	"github.com/broswen/webhookrelay/internal/repository"
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
var edgeAddress = ""
var edgeAccessId = ""
var edgeAccessSecret = ""
var brokers = ""
var topic = ""
var group = ""

func main() {

	flag.StringVar(&metricsAddress, "metricsAddr", os.Getenv("METRICS_ADDR"), "metrics server address")
	if metricsAddress == "" {
		log.Fatal().Msg("metrics address must be specified")
	}

	flag.StringVar(&brokers, "brokers", os.Getenv("BROKERS"), "kafka brokers")
	if brokers == "" {
		log.Fatal().Msg("kafka brokers must be specified")
	}

	flag.StringVar(&topic, "topic", os.Getenv("TOPIC"), "kafka topic")
	if topic == "" {
		log.Fatal().Msg("kafka topic must be specified")
	}

	flag.StringVar(&group, "group", os.Getenv("GROUP"), "kafka consumer group id")
	if group == "" {
		log.Fatal().Msg("kafka consumer group id must be specified")
	}

	flag.StringVar(&edgeAddress, "webhookdispatcherAddress", os.Getenv("WEBHOOKDISPATCHER_ADDRESS"), "address to the webhook dispatcher api")
	if edgeAddress == "" {
		log.Fatal().Msg("webhook dispatcher address must be specified")
	}
	flag.StringVar(&edgeAccessId, "webhookdispatcherAccessId", os.Getenv("ACCESS_ID"), "access id for the webhook dispatcher api")
	flag.StringVar(&edgeAccessSecret, "webhookdispatcherAccessSecret", os.Getenv("ACCESS_SECRET"), "access secret for the webhook dispatcher api")

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()
	eg, gCtx := errgroup.WithContext(ctx)

	edge, err := repository.NewEdgeRepository(edgeAddress, edgeAccessId, edgeAccessSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create edge repository")
	}

	config := sarama.NewConfig()
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategySticky()}
	consumer, err := sarama.NewConsumerGroup(strings.Split(brokers, ","), group, config)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create consumer group")
	}
	handler := provisioner.NewProvisionerHandler(edge)

	eg.Go(func() error {
		for {
			if err := consumer.Consume(gCtx, []string{topic}, handler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return nil
				}
				return err
			}
		}
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
