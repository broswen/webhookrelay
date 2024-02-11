package rest

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var HttpRequestLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "webhookrelay",
	Name:      "http_requests_ms",
	Help:      "http request latency histogram",
	Buckets:   []float64{10, 50, 100, 250, 500, 1000, 3000, 5000, 10_000},
}, []string{"method", "path", "status"})
