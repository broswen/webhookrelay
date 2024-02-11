package rest

import (
	"fmt"
	"github.com/broswen/webhookrelay/internal/service"
	"github.com/rs/zerolog/log"
	"net/http"
)

var GetWebhookPath = fmt.Sprintf("/webhooks/{%s}", WEBHOOK_ID_KEY)

func HandleGetWebhook(webhookService service.Webhook) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := webhookId(r)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
		wh, err := webhookService.Get(r.Context(), id)
		if err != nil {
			log.Error().Err(err).Str("id", id).Msg("error getting webhook")
			writeErr(w, nil, err)
			return
		}

		err = writeOK(w, http.StatusOK, wh)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
	}
}

var ListWebhooksPath = "/webhooks"

func HandleListWebhooks(webhookService service.Webhook) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		page := pagination(r)
		deleted := r.URL.Query().Get("deleted")
		includeDeleted := deleted == "true"
		whs, err := webhookService.List(r.Context(), includeDeleted, page.Offset, page.Limit)
		if err != nil {
			log.Error().Err(err).Msg("error listing webhooks")
			writeErr(w, nil, err)
			return
		}
		err = writeOK(w, http.StatusOK, whs)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
	}
}

var CreateWebhookPath = "/webhooks"

func HandleCreateWebhook(webhookService service.Webhook) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := service.CreateWebhookRequest{}
		err := readJSON(w, r, &req)
		if err != nil {
			writeErr(w, nil, err)
			return
		}

		// TODO handle conflict error code with in progress idempotency token
		n, err := webhookService.Create(r.Context(), req)
		if err != nil {
			log.Error().Err(err).Msg("error creating webhook")
			writeErr(w, nil, err)
			return
		}

		err = writeOK(w, http.StatusOK, n)
		if err != nil {
			writeErr(w, nil, err)
			return
		}
	}
}
