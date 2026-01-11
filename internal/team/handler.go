package team

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/middleware"
	"backend-go-ticketing-gamify/internal/response"
)

// Handler handles HTTP requests for team.
type Handler struct {
	service *Service
}

// NewHandler creates a new team handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches team endpoints.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/members", h.getMembers)
	router.GET("/projects/:projectId/members", h.getProjectMembers)
}

func (h *Handler) getMembers(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "login required")
		return
	}

	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	members, err := h.service.GetMembers(c.Request.Context(), user.ID, user.Role, limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if members == nil {
		members = []Member{}
	}
	response.OK(c, members)
}

func (h *Handler) getProjectMembers(c *gin.Context) {
	projectID := c.Param("projectId")
	if projectID == "" {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "projectId is required")
		return
	}
	members, err := h.service.GetProjectMembers(c.Request.Context(), projectID)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if members == nil {
		members = []ProjectMember{}
	}
	response.OK(c, members)
}
