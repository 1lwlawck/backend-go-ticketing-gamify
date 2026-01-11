package activity

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/response"
)

// Handler handles HTTP requests for activity.
type Handler struct {
	service *Service
}

// NewHandler creates a new activity handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches activity endpoints.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/", h.getActivity)
	router.GET("/user/:userId", h.getUserActivity)
}

func (h *Handler) getActivity(c *gin.Context) {
	filter := Filter{
		EntityType: c.Query("entityType"),
	}
	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		if t, err := time.Parse(time.RFC3339, cursorStr); err == nil {
			filter.Cursor = &t
		}
	}

	activities, nextCursor, err := h.service.GetUserActivity(c.Request.Context(), filter)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if activities == nil {
		activities = []ActivityItem{}
	}

	meta := gin.H{"limit": filter.Limit}
	if nextCursor != nil {
		meta["nextCursor"] = nextCursor.Format(time.RFC3339)
	}
	response.WithMeta(c, http.StatusOK, activities, meta)
}

func (h *Handler) getUserActivity(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}

	filter := Filter{
		UserID: userID,
	}
	filter.Limit, _ = strconv.Atoi(c.DefaultQuery("limit", "50"))
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		if t, err := time.Parse(time.RFC3339, cursorStr); err == nil {
			filter.Cursor = &t
		}
	}

	activities, nextCursor, err := h.service.GetUserActivity(c.Request.Context(), filter)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if activities == nil {
		activities = []ActivityItem{}
	}

	meta := gin.H{"limit": filter.Limit}
	if nextCursor != nil {
		meta["nextCursor"] = nextCursor.Format(time.RFC3339)
	}
	response.WithMeta(c, http.StatusOK, activities, meta)
}
