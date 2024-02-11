package provisioner

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var ProvisionAttempts = promauto.NewCounterVec(prometheus.CounterOpts{
	Namespace: "webhookrelay",
	Name:      "provision_attempts",
	Help:      "provision attempts counter",
}, []string{"result"})

var ProvisionLatency = promauto.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: "webhookrelay",
	Name:      "provision_latency_ms",
	Help:      "provision latency histogram",
	Buckets:   []float64{10, 50, 100, 250, 500, 1000, 3000, 5000, 10_000},
}, []string{"result"})

var Rebalances = promauto.NewCounter(prometheus.CounterOpts{
	Namespace: "webhookrelay",
	Name:      "rebalances",
	Help:      "rebalance counter",
})

var AcquiredPartitionCount = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: "webhookrelay",
	Name:      "acquired_partitions",
	Help:      "acquired partitions gauge",
}, []string{"topic"})
