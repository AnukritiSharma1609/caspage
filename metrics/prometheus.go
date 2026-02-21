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
	}

	prometheus.MustRegister(
		c.pageFetchCount,
		c.pageFetchDuration,
		c.pageErrorCount,
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

var _ core.MetricsCollector = (*PrometheusCollector)(nil) // compile-time check
