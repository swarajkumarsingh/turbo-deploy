package prometheus

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)


func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

var (
	totalRequests = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "app_total_requests",
			Help: "Total number of requests to my app",
		},
	)
)

var cpuTemp = prometheus.NewGauge(prometheus.GaugeOpts{
	Name: "cpu_temperature_celsius_a",
	Help: "Current temperature of the CPU.",
})

func CustomMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cpuTemp.Set(float64(100))
		totalRequests.Inc()
		c.Next()
	}
}


func init() {
	prometheus.MustRegister(cpuTemp)
	prometheus.MustRegister(totalRequests)
}