package users

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/middleware"
	"backend-go-ticketing-gamify/internal/response"
)

// Handler wires user routes.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", middleware.RequireRoles("admin", "project_manager"), h.list)
	router.GET("/me", h.me)
	router.PATCH("/me", h.updateMe)
	router.GET("/:id", middleware.RequireRoles("admin", "project_manager"), h.get)
	router.PATCH("/:id/role", middleware.RequireRoles("admin"), h.updateRole)
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	users, err := h.service.List(c.Request.Context(), limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.WithMeta(c, http.StatusOK, users, gin.H{"limit": limit})
}

func (h *Handler) me(c *gin.Context) {
	userCtx := middleware.CurrentUser(c)
	if userCtx == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	user, err := h.service.Get(c.Request.Context(), userCtx.ID)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if user == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "user not found")
		return
	}
	response.OK(c, user)
}

func (h *Handler) updateMe(c *gin.Context) {
	userCtx := middleware.CurrentUser(c)
	if userCtx == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload UpdateProfileInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	user, err := h.service.UpdateProfile(c.Request.Context(), userCtx.ID, payload)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if user == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "user not found")
		return
	}
	response.OK(c, user)
}

func (h *Handler) get(c *gin.Context) {
	user, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if user == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "user not found")
		return
	}
	response.OK(c, user)
}

func (h *Handler) updateRole(c *gin.Context) {
	var payload struct {
		Role string `json:"role" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if payload.Role != "" && !isAllowedRole(payload.Role) {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "invalid role")
		return
	}
	user, err := h.service.UpdateRole(c.Request.Context(), c.Param("id"), payload.Role)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if user == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "user not found")
		return
	}
	response.OK(c, user)
}

func isAllowedRole(role string) bool {
	switch role {
	case "admin", "project_manager", "developer", "viewer":
		return true
	default:
		return false
	}
}
