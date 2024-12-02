package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/authentication"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"github.com/swarajkumarsingh/turbo-deploy/controller/prometheus"
	"github.com/swarajkumarsingh/turbo-deploy/functions/logger"
	deploymentRoutes "github.com/swarajkumarsingh/turbo-deploy/routes/deployment"
	deploymentLogRoutes "github.com/swarajkumarsingh/turbo-deploy/routes/deployment_log"
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
	if constants.STAGE == constants.ENV_PROD {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.Default()

	// Custom middleware
	r.Use(enableCORS())
	r.Use(prometheus.CustomMetricsMiddleware())
	r.Use(authentication.RateLimit())

	// Run migrations
	MigrateDB()

	// Health check route
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"error":   false,
			"message": "health ok",
		})
	})

	// Version check
	r.GET("/version", func(c *gin.Context) {
		c.JSON(200, gin.H{"version": version})
	})

	// Metrics route
	r.GET("/metrics", prometheus.PrometheusHandler())

	// Add routes
	userRoutes.AddRoutes(r)
	projectRoutes.AddRoutes(r)
	deploymentRoutes.AddRoutes(r)
	deploymentLogRoutes.AddRoutes(r)

	// Create server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Implementation for Graceful Shutdown
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("Server Started, version: %s", version)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panicf("Server failed to start: %v", err)
		}
	}()

	<-done
	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Panicf("Server shutdown failed: %v", err)
	}

	log.Println("Server gracefully stopped")
}
