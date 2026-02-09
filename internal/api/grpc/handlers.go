package grpc

import (
	"context"
	"errors"
	"github.com/broswen/webhookrelay/internal/model"
	"github.com/broswen/webhookrelay/internal/service"
	apiv1 "github.com/broswen/webhookrelay/pkg/api/v1"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Server implements the gRPC WebhookRelayService
type Server struct {
	apiv1.UnimplementedWebhookRelayServiceServer
	Webhooks service.Webhook
}

// GetWebhook retrieves a single webhook by ID
func (s *Server) GetWebhook(ctx context.Context, req *apiv1.GetWebhookRequest) (*apiv1.GetWebhookResponse, error) {
	if req.WebhookId == "" {
		return nil, status.Error(codes.InvalidArgument, "webhook_id is required")
	}

	wh, err := s.Webhooks.Get(ctx, req.WebhookId)
	if err != nil {
		log.Error().Err(err).Str("id", req.WebhookId).Msg("error getting webhook")
		return nil, toGRPCError(err)
	}

	return &apiv1.GetWebhookResponse{
		Webhook: modelToProtoWebhook(wh),
	}, nil
}

// ListWebhooks retrieves multiple webhooks with pagination
func (s *Server) ListWebhooks(ctx context.Context, req *apiv1.ListWebhooksRequest) (*apiv1.ListWebhooksResponse, error) {
	offset := req.Offset
	limit := req.Limit
	if limit == 0 {
		limit = 100 // default limit
	}

	whs, err := s.Webhooks.List(ctx, req.Deleted, offset, limit)
	if err != nil {
		log.Error().Err(err).Msg("error listing webhooks")
		return nil, toGRPCError(err)
	}

	protoWebhooks := make([]*apiv1.Webhook, len(whs))
	for i, wh := range whs {
		protoWebhooks[i] = modelToProtoWebhook(wh)
	}

	return &apiv1.ListWebhooksResponse{
		Webhooks: protoWebhooks,
	}, nil
}

// CreateWebhook creates a new webhook
func (s *Server) CreateWebhook(ctx context.Context, req *apiv1.CreateWebhookRequest) (*apiv1.CreateWebhookResponse, error) {
	if req.Target == "" {
		return nil, status.Error(codes.InvalidArgument, "target is required")
	}
	if req.IdempotencyToken == "" {
		return nil, status.Error(codes.InvalidArgument, "idempotency_token is required")
	}

	createReq := service.CreateWebhookRequest{
		IdempotencyToken: req.IdempotencyToken,
		Target:           req.Target,
		Payload:          req.Payload,
	}

	wh, err := s.Webhooks.Create(ctx, createReq)
	if err != nil {
		log.Error().Err(err).Msg("error creating webhook")
		return nil, toGRPCError(err)
	}

	return &apiv1.CreateWebhookResponse{
		Webhook: modelToProtoWebhook(wh),
	}, nil
}

// toGRPCError converts service errors to gRPC status errors
func toGRPCError(err error) error {
	if err == nil {
		return nil
	}

	// Check for specific error types
	var invalidReq service.ErrInvalidRequest
	if errors.As(err, &invalidReq) {
		return status.Error(codes.InvalidArgument, err.Error())
	}

	var tokenInProgress service.ErrTokenInProgress
	if errors.As(err, &tokenInProgress) {
		return status.Error(codes.AlreadyExists, err.Error())
	}

	// Default to internal error
	return status.Error(codes.Internal, "internal server error")
}

// modelToProtoWebhook converts internal model.Webhook to proto message
func modelToProtoWebhook(wh model.Webhook) *apiv1.Webhook {
	protoWh := &apiv1.Webhook{
		Id:        wh.Id,
		Target:    wh.Target,
		Payload:   wh.Payload,
		CreatedAt: timestamppb.New(wh.CreatedAt),
		Status:    stringToProtoStatus(wh.Status),
	}

	// Handle optional timestamps
	if wh.DeletedAt.Valid {
		protoWh.DeletedAt = timestamppb.New(wh.DeletedAt.Time)
	}
	if wh.PublishedAt.Valid {
		protoWh.PublishedAt = timestamppb.New(wh.PublishedAt.Time)
	}
	if wh.ProvisionedAt.Valid {
		protoWh.ProvisionedAt = timestamppb.New(wh.ProvisionedAt.Time)
	}

	// Convert attempts
	if len(wh.Attempts) > 0 {
		protoWh.Attempts = make([]*apiv1.WebhookAttempt, len(wh.Attempts))
		for i, attempt := range wh.Attempts {
			protoWh.Attempts[i] = &apiv1.WebhookAttempt{
				Timestamp: timestamppb.New(attempt.Timestamp),
				Status:    int32(attempt.TargetStatus),
				Message:   attempt.Message,
			}
		}
	}

	return protoWh
}

// stringToProtoStatus converts string status to proto enum
func stringToProtoStatus(status string) apiv1.WebhookStatus {
	switch status {
	case "UNKNOWN":
		return apiv1.WebhookStatus_WEBHOOK_STATUS_UNKNOWN
	case "PENDING":
		return apiv1.WebhookStatus_WEBHOOK_STATUS_PENDING
	case "FAILED":
		return apiv1.WebhookStatus_WEBHOOK_STATUS_FAILED
	case "SUCCEEDED":
		return apiv1.WebhookStatus_WEBHOOK_STATUS_SUCCEEDED
	default:
		return apiv1.WebhookStatus_WEBHOOK_STATUS_UNSPECIFIED
	}
}
