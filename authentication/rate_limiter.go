package authentication

import (
	"net/http"

	"github.com/didip/tollbooth/v7"
	"github.com/didip/tollbooth/v7/limiter"
	"github.com/gin-gonic/gin"
)

func RateLimit() gin.HandlerFunc {
	lmt := limiter.New(&limiter.ExpirableOptions{
		DefaultExpirationTTL: 0,
		ExpireJobInterval:    1,
	})

	lmt.SetMax(10)

	return func(c *gin.Context) {
		httpError := tollbooth.LimitByRequest(lmt, c.Writer, c.Request)

		if httpError != nil {
			c.JSON(http.StatusTooManyRequests, gin.H{"error": false, "message": "rate limit exceeded"})
			c.Abort()
			return
		}

		c.Next()
	}
}
