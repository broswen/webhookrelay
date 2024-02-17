package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/broswen/webhookrelay/internal/model"
	"github.com/broswen/webhookrelay/internal/retry"
	"io"
	"net/http"
	"time"
)

type Edge interface {
	Get(ctx context.Context, id string) (model.EdgeWebhook, error)
	Create(ctx context.Context, webhook model.Webhook) error
}

func NewEdgeRepository(address, id, secret string) (Edge, error) {
	return &EdgeRepository{
		address:      address,
		accessId:     id,
		accessSecret: secret,
		client:       &http.Client{},
	}, nil
}

type EdgeRepository struct {
	address      string
	accessId     string
	accessSecret string
	client       *http.Client
}

func (r *EdgeRepository) Get(ctx context.Context, id string) (model.EdgeWebhook, error) {
	res, err := retry.NewRetry(time.Millisecond*50, 3, func() (*http.Response, error, bool) {
		res, err := r.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/api/webhooks/%s", id), nil)
		return res, err, true
	})()
	if err != nil {
		return model.EdgeWebhook{}, err
	}
	wh := model.EdgeWebhook{}
	err = json.NewDecoder(res.Body).Decode(&wh)
	if err != nil {
		if res.StatusCode == http.StatusNotFound {
			return model.EdgeWebhook{}, ErrWebhookNotFound{id: id}
		}
		return model.EdgeWebhook{}, err
	}
	return wh, nil
}

func (r *EdgeRepository) Create(ctx context.Context, webhook model.Webhook) error {
	edgeWebhook := model.EdgeWebhook{
		Id:      webhook.Id,
		Target:  webhook.Target,
		Payload: webhook.Payload,
	}
	b, err := json.Marshal(&edgeWebhook)
	if err != nil {
		return err
	}
	res, err := retry.NewRetry(time.Millisecond*50, 3, func() (*http.Response, error, bool) {
		res, err := r.makeRequest(ctx, http.MethodPost, "/api/webhooks", bytes.NewReader(b))
		return res, err, true
	})()

	if err != nil {
		return err
	}
	defer res.Body.Close()
	return nil
}

func (r *EdgeRepository) makeRequest(ctx context.Context, method string, path string, body io.Reader) (*http.Response, error) {
	url := fmt.Sprintf("%s%s", r.address, path)
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}
	if r.accessId != "" {
		req.Header.Set("CF-Access-Client-Id", r.accessId)
		req.Header.Set("CF-Access-Client-Secret", r.accessSecret)
	}
	return r.client.Do(req)
}
