package authentication

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter - rate limiting middleware
func RateLimit() gin.HandlerFunc {
	// Create a limiter that allows 100 requests per minute
	// This translates to approximately 1.67 requests per second
	limiter := rate.NewLimiter(rate.Every(time.Minute/100000), 10)

	var mu sync.Mutex

	return func(c *gin.Context) {
		mu.Lock()
		defer mu.Unlock()

		// Check if the request can be allowed
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   true,
				"message": "Rate limit exceeded. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
