package grpc

import (
	"testing"
	"time"

	"github.com/broswen/webhookrelay/internal/model"
	apiv1 "github.com/broswen/webhookrelay/pkg/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestModelToProtoWebhook(t *testing.T) {
	now := time.Now()
	webhook := model.Webhook{
		Id:        "test-id-123",
		Target:    "https://example.com/webhook",
		Payload:   []byte("test payload"),
		CreatedAt: now,
		Status:    "PENDING",
		Attempts: []model.WebhookAttempt{
			{
				Timestamp:    now,
				TargetStatus: 200,
				Message:      "Success",
			},
		},
	}

	result := modelToProtoWebhook(webhook)

	assert.NotNil(t, result)
	assert.Equal(t, webhook.Id, result.Id)
	assert.Equal(t, webhook.Target, result.Target)
	assert.Equal(t, webhook.Payload, result.Payload)
	assert.Equal(t, timestamppb.New(now), result.CreatedAt)
	assert.Equal(t, apiv1.WebhookStatus_WEBHOOK_STATUS_PENDING, result.Status)
	assert.Len(t, result.Attempts, 1)
	assert.Equal(t, int32(200), result.Attempts[0].Status)
	assert.Equal(t, "Success", result.Attempts[0].Message)
}

func TestStringToProtoStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected apiv1.WebhookStatus
	}{
		{"UNKNOWN", "UNKNOWN", apiv1.WebhookStatus_WEBHOOK_STATUS_UNKNOWN},
		{"PENDING", "PENDING", apiv1.WebhookStatus_WEBHOOK_STATUS_PENDING},
		{"FAILED", "FAILED", apiv1.WebhookStatus_WEBHOOK_STATUS_FAILED},
		{"SUCCEEDED", "SUCCEEDED", apiv1.WebhookStatus_WEBHOOK_STATUS_SUCCEEDED},
		{"Empty", "", apiv1.WebhookStatus_WEBHOOK_STATUS_UNSPECIFIED},
		{"Invalid", "INVALID", apiv1.WebhookStatus_WEBHOOK_STATUS_UNSPECIFIED},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringToProtoStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
