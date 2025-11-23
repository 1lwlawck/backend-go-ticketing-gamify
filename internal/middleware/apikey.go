package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// APIKeyGuard enforces an API key if one is configured. If apiKey is empty, it is a no-op.
func APIKeyGuard(apiKey, header string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if apiKey == "" {
			c.Next()
			return
		}
		if c.GetHeader(header) != apiKey {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}
