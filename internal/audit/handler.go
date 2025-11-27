package audit

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/response"
)

// Handler provides HTTP endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.list)
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var cursorPtr *time.Time
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		if ts, err := time.Parse(time.RFC3339, cursorStr); err == nil {
			cursorPtr = &ts
		}
	}
	entries, nextCursor, err := h.service.List(c.Request.Context(), limit, cursorPtr)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	meta := gin.H{"limit": limit}
	if nextCursor != nil {
		meta["nextCursor"] = *nextCursor
	}
	response.WithMeta(c, http.StatusOK, entries, meta)
}
