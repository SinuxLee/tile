package ginx

import (
	"time"

	"github.com/gin-gonic/gin"
	limiter "github.com/julianshen/gin-limiter"
)

// NewRateLimiter ...
func NewRateLimiter(interval time.Duration, cap int64) gin.HandlerFunc {
	return limiter.NewRateLimiter(interval, cap, func(ctx *gin.Context) (string, error) {
		return "", nil
	}).Middleware()
}
