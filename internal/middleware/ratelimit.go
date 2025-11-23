package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimit limits requests per key within a window.
// If an API key header is present, it is used as the bucket key and apiLimit is applied.
// Otherwise the client IP is used with ipLimit.
func RateLimit(ipLimit, apiLimit int, window time.Duration, apiKeyHeader string) gin.HandlerFunc {
	type counter struct {
		count     int
		windowEnd time.Time
	}
	var mu sync.Mutex
	buckets := make(map[string]*counter)

	return func(c *gin.Context) {
		now := time.Now()
		key := c.ClientIP()
		limit := ipLimit
		if apiKey := c.GetHeader(apiKeyHeader); apiKey != "" {
			key = "api:" + apiKey
			if apiLimit > 0 {
				limit = apiLimit
			}
		}

		mu.Lock()
		b, ok := buckets[key]
		if !ok || now.After(b.windowEnd) {
			buckets[key] = &counter{count: 1, windowEnd: now.Add(window)}
			mu.Unlock()
			c.Next()
			return
		}

		if b.count >= limit {
			mu.Unlock()
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		b.count++
		mu.Unlock()
		c.Next()
	}
}
