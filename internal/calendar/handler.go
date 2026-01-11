package calendar

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/response"
)

// Handler handles HTTP requests for calendar.
type Handler struct {
	service *Service
}

// NewHandler creates a new calendar handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches calendar endpoints.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/events", h.getEvents)
}

func (h *Handler) getEvents(c *gin.Context) {
	var filter Filter

	if startStr := c.Query("start"); startStr != "" {
		if t, err := time.Parse("2006-01-02", startStr); err == nil {
			filter.StartDate = &t
		}
	}
	if endStr := c.Query("end"); endStr != "" {
		if t, err := time.Parse("2006-01-02", endStr); err == nil {
			filter.EndDate = &t
		}
	}
	filter.ProjectID = c.Query("projectId")
	filter.Type = c.DefaultQuery("type", "all")

	events, err := h.service.GetEvents(c.Request.Context(), filter)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if events == nil {
		events = []CalendarEvent{}
	}
	response.OK(c, events)
}
