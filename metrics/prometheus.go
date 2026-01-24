package metrics

import (
	"time"

	"github.com/AnukritiSharma1609/caspage/core"
	"github.com/prometheus/client_golang/prometheus"
)

// PrometheusCollector implements core.MetricsCollector
type PrometheusCollector struct {
	pageFetchCount    prometheus.Counter
	pageFetchDuration prometheus.Histogram
	pageErrorCount    prometheus.Counter
	activeTokens      prometheus.Gauge
}

// NewPrometheusCollector creates and registers Prometheus metrics.
func NewPrometheusCollector() *PrometheusCollector {
	c := &PrometheusCollector{
		pageFetchCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "caspage_page_fetch_total",
			Help: "Total number of pages fetched",
		}),
		pageFetchDuration: prometheus.NewHistogram(prometheus.HistogramOpts{
			Name:    "caspage_page_duration_seconds",
			Help:    "Time taken per page fetch",
			Buckets: prometheus.DefBuckets,
		}),
		pageErrorCount: prometheus.NewCounter(prometheus.CounterOpts{
			Name: "caspage_errors_total",
			Help: "Total number of pagination errors",
		}),
		activeTokens: prometheus.NewGauge(prometheus.GaugeOpts{
			Name: "caspage_active_tokens_total",
			Help: "Number of currently cached pagination tokens",
		}),
	}

	prometheus.MustRegister(
		c.pageFetchCount,
		c.pageFetchDuration,
		c.pageErrorCount,
		c.activeTokens,
	)

	return c
}

func (c *PrometheusCollector) ObservePageFetch(rows int, duration time.Duration) {
	c.pageFetchCount.Inc()
	c.pageFetchDuration.Observe(duration.Seconds())
}

func (c *PrometheusCollector) ObserveError(err error) {
	c.pageErrorCount.Inc()
}

func (c *PrometheusCollector) ObserveActiveTokens(count int) {
	c.activeTokens.Set(float64(count))
}

var _ core.MetricsCollector = (*PrometheusCollector)(nil) // compile-time check
