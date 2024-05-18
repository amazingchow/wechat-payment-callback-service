package metrics

import (
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	NAMESPACE = "amazingchow"
	SUBSYSTEM = "wechat-payment-callback-service"
)

var (
	_MetricsList = []prometheus.Collector{}
)

var _RegisterMetricsOnce sync.Once

// Register all metrics.
func Register() {
	_RegisterMetricsOnce.Do(func() {
		for _, metrics := range _MetricsList {
			prometheus.MustRegister(metrics)
		}
	})
}

// SinceInSeconds gets the time since the specified start in seconds.
func SinceInSeconds(st time.Time) float64 {
	return time.Since(st).Seconds()
}
