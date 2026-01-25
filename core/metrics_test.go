package core_test

import (
	"testing"
	"time"

	"github.com/AnukritiSharma1609/caspage/core"
	"github.com/AnukritiSharma1609/caspage/metrics"
)

func TestPrometheusCollector(t *testing.T) {
	m := metrics.NewPrometheusCollector()

	m.ObservePageFetch(10, 50*time.Millisecond)
	m.ObserveError(core.ErrQueryFailed)
	m.ObserveActiveTokens(3)

	// No direct output — just ensures it runs without panic
}
