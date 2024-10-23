package prometheus


import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)	


func PrometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()

	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}


func CustomMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cpuTemp.Set(float64(100))
		totalRequests.Inc()
		c.Next()
	}
}
