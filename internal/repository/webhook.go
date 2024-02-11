package repository

import (
	"context"
	"database/sql"
	"errors"
	"github.com/broswen/webhookrelay/internal/db"
	"github.com/broswen/webhookrelay/internal/model"
)

type Webhook interface {
	Get(ctx context.Context, id string) (model.Webhook, error)
	List(ctx context.Context, deleted bool, offset int64, limit int64) ([]model.Webhook, error)
	Create(ctx context.Context, target string, payload []byte) (model.Webhook, error)
	LockForPublishing(ctx context.Context, limit int64) ([]model.Webhook, error)
	MarkPublished(ctx context.Context, id string) error
}

func NewSqlWebhookRepository(conn db.Conn) Webhook {
	return &SqlWebhookRepository{
		conn: conn,
	}
}

type SqlWebhookRepository struct {
	conn db.Conn
}

func (r *SqlWebhookRepository) Get(ctx context.Context, id string) (model.Webhook, error) {
	wh := model.Webhook{}
	err := r.conn.QueryRowContext(ctx, `select id, target, payload, created_at, published_at, deleted_at from webhooks where id = $1;`, id).
		Scan(&wh.Id, &wh.Target, &wh.Payload, &wh.CreatedAt, &wh.PublishedAt, &wh.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.Webhook{}, ErrWebhookNotFound{err.Error()}
		}
		return model.Webhook{}, db.PgError(err)
	}

	return wh, nil
}

func (r *SqlWebhookRepository) List(ctx context.Context, deleted bool, offset int64, limit int64) ([]model.Webhook, error) {
	var rows *sql.Rows
	var err error
	if deleted {
		rows, err = r.conn.QueryContext(ctx, `select id, target, payload, created_at, published_at, deleted_at from webhooks where deleted_at is not null offset $1 limit $2;`, offset, limit)
	} else {
		rows, err = r.conn.QueryContext(ctx, `select id, target, payload, created_at, published_at, deleted_at from webhooks where deleted_at is null offset $1 limit $2;`, offset, limit)
	}
	err = db.PgError(err)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	whs := make([]model.Webhook, 0)
	for rows.Next() {
		wh := model.Webhook{}
		err = rows.Scan(&wh.Id, &wh.Target, &wh.Payload, &wh.CreatedAt, &wh.PublishedAt, &wh.DeletedAt)
		if err != nil {
			return whs, err
		}
		whs = append(whs, wh)
	}
	if rows.Err() != nil {
		return whs, db.PgError(err)
	}
	return whs, nil

}

func (r *SqlWebhookRepository) Create(ctx context.Context, target string, payload []byte) (model.Webhook, error) {
	wh := model.Webhook{}
	err := r.conn.QueryRowContext(ctx, `insert into webhooks (target, payload) values ($1, $2) returning id, target, payload, created_at, published_at, deleted_at;`, target, payload).
		Scan(&wh.Id, &wh.Target, &wh.Payload, &wh.CreatedAt, &wh.PublishedAt, &wh.DeletedAt)
	if err != nil {
		return model.Webhook{}, db.PgError(err)
	}
	return wh, nil
}

func (r *SqlWebhookRepository) LockForPublishing(ctx context.Context, limit int64) ([]model.Webhook, error) {
	rows, err := r.conn.QueryContext(ctx, `select id, target, payload, created_at, published_at, deleted_at from webhooks where published_at is null order by created_at asc limit $1 for update skip locked;`, limit)
	err = db.PgError(err)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	whs := make([]model.Webhook, 0)
	for rows.Next() {
		wh := model.Webhook{}
		err = rows.Scan(&wh.Id, &wh.Target, &wh.Payload, &wh.CreatedAt, &wh.PublishedAt, &wh.DeletedAt)
		if err != nil {
			return whs, err
		}
		whs = append(whs, wh)
	}
	if rows.Err() != nil {
		return whs, db.PgError(err)
	}
	return whs, nil
}

func (r *SqlWebhookRepository) MarkPublished(ctx context.Context, id string) error {
	res, err := r.conn.ExecContext(ctx, `update webhooks set published_at = now() where id = $1;`, id)
	err = db.PgError(err)
	if err != nil {
		return err
	}
	if count, _ := res.RowsAffected(); count == 0 {
		return ErrWebhookNotFound{}
	}
	return nil
}
