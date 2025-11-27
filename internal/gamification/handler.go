package gamification

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/response"
)

// Handler wires HTTP requests to the gamification service.
type Handler struct {
	service *Service
}

// NewHandler creates a Handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches gamification endpoints.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/stats/:userID", h.getStats)
	router.GET("/events", h.listEvents)
	router.GET("/leaderboard", h.leaderboard)
}

func (h *Handler) getStats(c *gin.Context) {
	userID := c.Param("userID")
	if userID == "" {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "userID is required")
		return
	}
	stats, err := h.service.GetStats(c.Request.Context(), userID)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if stats == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "stats not found")
		return
	}
	response.OK(c, stats)
}

func (h *Handler) listEvents(c *gin.Context) {
	userID := c.Query("userId")
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
	events, nextCursor, err := h.service.ListEvents(c.Request.Context(), userID, limit, cursorPtr)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	meta := gin.H{"limit": limit}
	if nextCursor != nil {
		meta["nextCursor"] = *nextCursor
	}
	response.WithMeta(c, http.StatusOK, events, meta)
}

func (h *Handler) leaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	cursor, _ := strconv.Atoi(c.DefaultQuery("cursor", "0"))
	if cursor < 0 {
		cursor = 0
	}
	rows, err := h.service.Leaderboard(c.Request.Context(), limit, cursor)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	meta := gin.H{"limit": limit}
	if len(rows) == limit {
		meta["nextCursor"] = cursor + limit
	}
	response.WithMeta(c, http.StatusOK, rows, meta)
}
