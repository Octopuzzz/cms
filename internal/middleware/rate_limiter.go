package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter returns a per-IP rate limiting middleware using token bucket
func RateLimiter(rps float64, burst int) gin.HandlerFunc {
	limiters := newIPLimiter(rps, burst)
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := limiters.get(ip)
		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// ipLimiter holds per-IP limiters
type ipLimiter struct {
	limiters map[string]*rate.Limiter
	rps      float64
	burst    int
}

func newIPLimiter(rps float64, burst int) *ipLimiter {
	return &ipLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rps,
		burst:    burst,
	}
}

func (il *ipLimiter) get(ip string) *rate.Limiter {
	if l, ok := il.limiters[ip]; ok {
		return l
	}
	l := rate.NewLimiter(rate.Limit(il.rps), il.burst)
	il.limiters[ip] = l
	return l
}

// SecurityHeaders adds security response headers
func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Next()
	}
}
