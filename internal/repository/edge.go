package repository

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/broswen/webhookrelay/internal/model"
	"io"
	"net/http"
)

type Edge interface {
	Get(ctx context.Context, id string) (model.Webhook, error)
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

//TODO decide on model for webhook metadata and dispatcher store

func (r *EdgeRepository) Get(ctx context.Context, id string) (model.Webhook, error) {
	res, err := r.makeRequest(ctx, http.MethodGet, fmt.Sprintf("/webhook/%s", id), nil)
	if err != nil {
		return model.Webhook{}, err
	}
	wh := model.Webhook{}
	err = json.NewDecoder(res.Body).Decode(&wh)
	if err != nil {
		return model.Webhook{}, err
	}
	return wh, nil
}

//TODO decide on model for webhook metadata and dispatcher store

func (r *EdgeRepository) Create(ctx context.Context, webhook model.Webhook) error {
	b, err := json.Marshal(&webhook)
	if err != nil {
		return err
	}
	res, err := r.makeRequest(ctx, http.MethodPost, "/webhook", bytes.NewReader(b))
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
