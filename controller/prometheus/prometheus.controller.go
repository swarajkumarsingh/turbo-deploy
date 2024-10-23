package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

var (
	totalRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_total_requests",
			Help: "Total number of requests to my app",
		},
	)
)

var cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "cpu_temperature_celsius",
	Help: "Current temperature of the CPU.",
})

func init() {
	prometheus.MustRegister(cpuTemp)
	prometheus.MustRegister(totalRequests)
}