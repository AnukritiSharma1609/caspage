package core

import (
	"context"
)

// Options holds configuration for the paginator.
type Options struct {
	PageSize int
	Filters  map[string]interface{}
	Columns  []string
	Context  context.Context
	Logger   func(event string, data map[string]interface{})
	Metrics  MetricsCollector // optional metrics hook
}
