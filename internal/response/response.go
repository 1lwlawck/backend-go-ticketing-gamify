package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// JSON helpers for consistent envelopes.
//
// ErrorBody adds a stable machine-friendly code plus human-readable message.
type ErrorBody struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func OK(c *gin.Context, data any) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

func Created(c *gin.Context, data any) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

func NoContent(c *gin.Context) {
	c.Status(http.StatusNoContent)
}

// ErrorCode returns an error envelope with code + message.
func ErrorCode(c *gin.Context, status int, code, message string) {
	c.JSON(status, gin.H{"error": ErrorBody{Code: code, Message: message}})
}

// ErrorCodeDetails returns an error envelope and optional details (e.g., validation errors).
func ErrorCodeDetails(c *gin.Context, status int, code, message string, details any) {
	c.JSON(status, gin.H{"error": ErrorBody{Code: code, Message: message, Details: details}})
}

// WithMeta allows returning meta (e.g., pagination) alongside data.
func WithMeta(c *gin.Context, status int, data any, meta gin.H) {
	if meta == nil {
		meta = gin.H{}
	}
	resp := gin.H{"data": data}
	for k, v := range meta {
		resp[k] = v
	}
	c.JSON(status, resp)
}
