package gamification

import (
	"net/http"
	"strconv"

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
	events, err := h.service.ListEvents(c.Request.Context(), userID, limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.WithMeta(c, http.StatusOK, events, gin.H{"limit": limit})
}

func (h *Handler) leaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	rows, err := h.service.Leaderboard(c.Request.Context(), limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, rows)
}
