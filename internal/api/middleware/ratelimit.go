package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimiter 基于 IP 的滑动窗口限流器。
type RateLimiter struct {
	mu      sync.Mutex
	windows map[string][]time.Time
	max     int
	window  time.Duration
}

// NewRateLimiter 创建限流器。
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		windows: make(map[string][]time.Time),
		max:     max,
		window:  window,
	}
}

// Middleware 返回限流中间件。
func (r *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		now := time.Now()

		r.mu.Lock()
		var kept []time.Time
		for _, t := range r.windows[ip] {
			if now.Sub(t) < r.window {
				kept = append(kept, t)
			}
		}
		if len(kept) >= r.max {
			r.windows[ip] = kept
			r.mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"code": 40300, "message": "请求过于频繁，请稍后再试"})
			return
		}
		kept = append(kept, now)
		r.windows[ip] = kept
		r.mu.Unlock()

		c.Next()
	}
}
