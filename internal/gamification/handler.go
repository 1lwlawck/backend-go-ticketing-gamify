package gamification

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "userID is required"})
		return
	}
	stats, err := h.service.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if stats == nil {
		stats = &UserStats{
			UserID:             userID,
			XPTotal:            0,
			Level:              1,
			NextLevelThreshold: 100,
			TicketsClosed:      0,
			StreakDays:         0,
		}
	}
	c.JSON(http.StatusOK, stats)
}

func (h *Handler) listEvents(c *gin.Context) {
	userID := c.Query("userId")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	events, err := h.service.ListEvents(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": events})
}

func (h *Handler) leaderboard(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	rows, err := h.service.Leaderboard(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": rows})
}
