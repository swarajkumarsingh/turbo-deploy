package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/controller/prometheus"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	projectRoutes "github.com/swarajkumarsingh/turbo-deploy/routes/project"
	userRoutes "github.com/swarajkumarsingh/turbo-deploy/routes/user"
)

var log = logger.Log
var version string = "1.0"

func enableCORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers",
			"Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-Api-Key, token, User-Agent, Referer")
		c.Writer.Header().Set("AllowCredentials", "true")
		c.Writer.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		if c.Request.Method == "OPTIONS" {
			return
		}
		c.Next()
	}
}

func main() {
	r := gin.Default()

	// custom middleware
	r.Use(enableCORS())
	r.Use(prometheus.CustomMetricsMiddleware())

	// run migrations
	MigrateDB()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"error":   false,
			"message": "health ok",
		})
	})
	r.GET("/metrics", prometheus.PrometheusHandler())
	
	userRoutes.AddRoutes(r)
	projectRoutes.AddRoutes(r)

	log.Printf("Server Started, version: %s", version)
	http.ListenAndServe(":8080", r)
}
