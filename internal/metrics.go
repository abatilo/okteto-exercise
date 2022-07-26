package internal

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type (
	Gauge     = prometheus.Gauge
	GaugeOpts = prometheus.GaugeOpts
)

// MetricsClient is a prometheus metrics client
type MetricsClient interface {
	NewGauge(opts prometheus.GaugeOpts) Gauge
}

// NoopMetrics is an empty metrics client that doesn't register to any metrics collector
type NoopMetrics struct{}

// NewGauge will create an empty prometheus counter but will not register it
func (n *NoopMetrics) NewGauge(opts prometheus.GaugeOpts) Gauge {
	return prometheus.NewGauge(opts)
}

// PrometheusMetrics represents a prometheus metrics client
type PrometheusMetrics struct{}

// NewGauge returns a new counter metric that's registered to the automatic prometheus collector
func (p *PrometheusMetrics) NewGauge(opts prometheus.GaugeOpts) Gauge {
	return promauto.NewGauge(opts)
}
