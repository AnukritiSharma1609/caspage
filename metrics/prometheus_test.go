package metrics_test

import (
	"testing"

	"github.com/AnukritiSharma1609/caspage/metrics"
)

func TestNewPrometheusCollector(t *testing.T) {
	_ = metrics.NewPrometheusCollector() // Should not panic
}
