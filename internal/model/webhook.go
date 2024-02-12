package model

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"time"
)

type NullableTime struct {
	sql.NullTime
}

func (t NullableTime) MarshalJSON() ([]byte, error) {
	if !t.NullTime.Valid {
		return []byte("null"), nil
	}

	return json.Marshal(t.Time.UTC())
}

func (t *NullableTime) UnmarshalJSON(data []byte) error {
	if bytes.Equal(data, []byte("null")) {
		t.NullTime.Valid = false
		return nil
	}
	parsedTime := time.Time{}
	err := json.Unmarshal(data, &parsedTime)
	if err != nil {
		t.NullTime.Valid = false
		return err
	}
	t.NullTime.Time = parsedTime
	t.NullTime.Valid = true
	return nil
}

type Webhook struct {
	Id            string           `json:"id"`
	Target        string           `json:"target"`
	Payload       []byte           `json:"payload"`
	CreatedAt     time.Time        `json:"createdAt"`
	DeletedAt     NullableTime     `json:"deletedAt"`
	PublishedAt   NullableTime     `json:"publishedAt"`
	ProvisionedAt NullableTime     `json:"provisionedAt"`
	Attempts      []WebhookAttempt `json:"attempts"`
}

type EdgeWebhook struct {
	Id            string           `json:"id"`
	Target        string           `json:"target"`
	Payload       []byte           `json:"payload"`
	ProvisionedAt NullableTime     `json:"provisionedAt"`
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
