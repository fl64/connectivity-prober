package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

// Metrics содержит все метрики приложения
type Metrics struct {
	SuccessCount *prometheus.CounterVec
	LatencyHist  *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		SuccessCount: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "probe_success_total",
				Help: "Total number of successful probes",
			},
			[]string{"probe_type", "target"},
		),
		LatencyHist: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name: "probe_latency_seconds",
				Help: "Latency of probes in seconds",
				Buckets: []float64{
					0.001, 0.01, 0.1, 0.5, 1.0, 2.0, 5.0,
				},
			},
			[]string{"probe_type", "target"},
		),
	}
}
