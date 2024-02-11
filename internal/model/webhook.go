package model

import (
	"database/sql"
	"time"
)

type Webhook struct {
	Id          string       `json:"id"`
	Target      string       `json:"target"`
	Payload     []byte       `json:"payload"`
	CreatedAt   time.Time    `json:"createdAt"`
	DeletedAt   sql.NullTime `json:"deletedAt"`
	PublishedAt sql.NullTime `json:"publishedAt"`
}
