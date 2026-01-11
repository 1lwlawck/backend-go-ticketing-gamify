package achievements

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/response"
)

// Handler handles HTTP requests for achievements.
type Handler struct {
	service *Service
}

// NewHandler creates a new achievements handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches achievements endpoints.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/", h.getAllAchievements)
	router.GET("/user/:userId", h.getUserProgress)
	router.GET("/user/:userId/unlocked", h.getUnlockedAchievements)
}

func (h *Handler) getAllAchievements(c *gin.Context) {
	achievements := h.service.GetAllAchievements()
	response.OK(c, achievements)
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
		progress = []Progress{}
	}
	response.OK(c, progress)
}

func (h *Handler) getUnlockedAchievements(c *gin.Context) {
	userID := c.Param("userId")
	if userID == "" {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "userId is required")
		return
	}
	unlocked, err := h.service.GetUnlockedAchievements(c.Request.Context(), userID)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if unlocked == nil {
		unlocked = []Progress{}
	}
	response.OK(c, unlocked)
}
