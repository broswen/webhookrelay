package provisioner

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/IBM/sarama"
	"github.com/broswen/webhookrelay/internal/model"
	"github.com/broswen/webhookrelay/internal/repository"
	"github.com/rs/zerolog/log"
	"time"
)

func NewProvisionerHandler(edge repository.Edge) *Handler {
	return &Handler{edge: edge}
}

type Handler struct {
	edge repository.Edge
}

func (h *Handler) Setup(session sarama.ConsumerGroupSession) error {
	log.Info().Any("claims", session.Claims()).Msg("acquired claims")
	Rebalances.Inc()
	for topic, parts := range session.Claims() {
		AcquiredPartitionCount.WithLabelValues(topic).Set(float64(len(parts)))
	}
	return nil
}

func (h *Handler) Cleanup(session sarama.ConsumerGroupSession) error {
	log.Info().Any("claims", session.Claims()).Msg("released claims")
	for topic := range session.Claims() {
		AcquiredPartitionCount.WithLabelValues(topic).Set(0)
	}
	return nil
}

func (h *Handler) HandleMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	wh := model.Webhook{}
	if err := json.NewDecoder(bytes.NewReader(message.Value)).Decode(&wh); err != nil {
		return err
	}
	return h.edge.Create(ctx, wh)
}

func (h *Handler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			if message == nil {
				continue
			}

			start := time.Now()
			err := h.HandleMessage(session.Context(), message)

			if err != nil {
				log.Error().Err(err).Msg("failed to provision")
				ProvisionLatency.WithLabelValues("failure").Observe(float64(time.Since(start).Milliseconds()))
				ProvisionAttempts.WithLabelValues("failure").Inc()
				continue
			} else {
				ProvisionLatency.WithLabelValues("success").Observe(float64(time.Since(start).Milliseconds()))
				ProvisionAttempts.WithLabelValues("success").Inc()
			}

			session.MarkMessage(message, "")

		case <-session.Context().Done():
			return nil
		}
	}
}
