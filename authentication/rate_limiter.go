package authentication

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/swarajkumarsingh/turbo-deploy/constants"
	"golang.org/x/time/rate"
)

// RateLimiter - rate limiting middleware
func RateLimit() gin.HandlerFunc {
	// Create a limiter that allows 100 requests per minute
	// This translates to approximately 1.67 requests per second
	var mu sync.Mutex
	limiter := rate.NewLimiter(rate.Every(time.Minute/constants.DefaultRateLimiterPerMinute), 10)
	
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
