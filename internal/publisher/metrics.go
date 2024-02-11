package publisher

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var PublishAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "webhookrelay",
	Name:      "publish_attempts",
	Help:      "publish attempts counter",
}, []string{"result"})

var PublishLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "webhookrelay",
	Name:      "publish_latency_ms",
	Help:      "publish latency histogram",
	Buckets:   []float64{10, 50, 100, 250, 500, 1000, 3000, 5000, 10_000},
}, []string{"result"})
