package challenges

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/response"
)

// Handler handles HTTP requests for challenges.
type Handler struct {
	service *Service
}

// NewHandler creates a new challenges handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches challenges endpoints.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/active", h.getActiveChallenges)
	router.GET("/user/:userId", h.getUserProgress)
}

func (h *Handler) getActiveChallenges(c *gin.Context) {
	challenges := h.service.GetActiveChallenges()
	response.OK(c, challenges)
}

func (h *Handler) getUserProgress(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	progress, err := h.service.GetUserProgress(c.Request.Context(), userID)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if progress == nil {
		progress = []UserChallenge{}
	}
	response.OK(c, progress)
}
