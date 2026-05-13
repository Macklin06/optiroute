package services

import (
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	LocationUpdateDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "optiroute_location_update_duration_seconds",
			Help:    "How long each driver loaction update takes",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"status"},
	)

	LocationUpdateTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "optiroute_location_updates_total",
			Help: "Total number of drivers location update attempts",
		},
		[]string{"status"},
	)

	ActiveDriversGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "optiroute_active_drivers",
			Help: "Number of drivers currently active in Redis",
		},
	)
)

func init() {
	fmt.Println("METRICS INIT RUNNING")

	prometheus.MustRegister(LocationUpdateDuration)
	prometheus.MustRegister(LocationUpdateTotal)
	prometheus.MustRegister(ActiveDriversGauge)
}
