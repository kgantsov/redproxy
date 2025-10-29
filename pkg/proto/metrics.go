package proto

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
	CommandsProxiedTotal *prometheus.CounterVec
	Connections          *prometheus.GaugeVec
	Latency              *prometheus.HistogramVec
	Registry             *prometheus.Registry
}

func NewPrometheusMetrics(registry prometheus.Registerer, namespace, subsystem string) *PrometheusMetrics {
	m := &PrometheusMetrics{}

	m.CommandsProxiedTotal = promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "redproxy_commands_proxied_total",
			Help:      "Number of commands proxied",
		},
		[]string{},
	)

	m.Connections = promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "redproxy_connections",
			Help:      "Number of connections to Redis",
		},
		[]string{},
	)

	m.Latency = promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: namespace,
			Subsystem: subsystem,
			Name:      "redproxy_latency",
			Help:      "Latency of Redis commands",
			Buckets: []float64{
				0.000001, // 1µs
				0.000002,
				0.000005,
				0.00001, // 10µs
				0.00002,
				0.00005,
				0.0001, // 100µs
				0.0002,
				0.0005,
				0.001, // 1ms
				0.002,
				0.005,
				0.01, // 10ms
				0.02,
				0.05,
				0.1, // 100 ms
				0.2,
				0.5,
				1.0, // 1s
				2.0,
				5.0,
			},
		},
		[]string{},
	)

	return m
}
