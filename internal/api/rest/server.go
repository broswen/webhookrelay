package rest

import (
	"github.com/broswen/webhookrelay/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"net/http"
)

const WEBHOOK_ID_KEY = "webhookId"

type Server struct {
	Webhooks service.Webhook
}

func (s *Server) Router() http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.StripSlashes)
	r.Use(middleware.Logger)
	r.Use(middleware.RealIP)
	r.Use(middleware.SetHeader("Content-Type", "application/json"))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://webhookrelay.broswen.com", "http://localhost:3000", "http://localhost:8080"},
		AllowedMethods:   []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:   []string{"Origin", "Accept", "Content-Type", "Authorization"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	r.Use(Metrics)
	r.NotFound(func(w http.ResponseWriter, r *http.Request) {
		writeErr(w, nil, ErrNotFound)
	})
	r.Get("/healthcheck", func(w http.ResponseWriter, r *http.Request) {
		writeOK(w, http.StatusOK, "OK")
	})

	r.Route("/api", func(r chi.Router) {
		r.Get(ListWebhooksPath, HandleListWebhooks(s.Webhooks))
		r.Get(GetWebhookPath, HandleGetWebhook(s.Webhooks))
		r.Post(CreateWebhookPath, HandleCreateWebhook(s.Webhooks))
	})

	return r
}
