package model

import (
	"database/sql"
	"time"
)

type Webhook struct {
	Id            string           `json:"id"`
	Target        string           `json:"target"`
	Payload       []byte           `json:"payload"`
	CreatedAt     time.Time        `json:"createdAt"`
	DeletedAt     sql.NullTime     `json:"deletedAt"`
	PublishedAt   sql.NullTime     `json:"publishedAt"`
	ProvisionedAt sql.NullTime     `json:"provisionedAt"`
	Attempts      []WebhookAttempt `json:"attempts"`
}

type EdgeWebhook struct {
	Id            string           `json:"id"`
	Target        string           `json:"target"`
	Payload       []byte           `json:"payload"`
	ProvisionedAt sql.NullTime     `json:"provisionedAt"`
	Attempts      []WebhookAttempt `json:"attempts"`
}

type WebhookAttempt struct {
	// The timestamp the attempt was made
	Timestamp time.Time `json:"timestamp"`
	// The response http status code from the target
	TargetStatus int `json:"status"`
	// Additional attempt info
	Message string `json:"message"`
}
