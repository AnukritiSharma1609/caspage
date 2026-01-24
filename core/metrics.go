package core

import "time"

// MetricsCollector defines hooks for collecting pagination metrics.
// Implementations can use Prometheus, Datadog, custom logging, etc.
type MetricsCollector interface {
	ObservePageFetch(rows int, duration time.Duration)
	ObserveError(err error)
	ObserveActiveTokens(count int)
}
