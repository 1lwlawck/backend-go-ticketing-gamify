package epics

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/middleware"
	"backend-go-ticketing-gamify/internal/response"
)

// Handler exposes epics endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/projects/:id/epics", h.listByProject)
	router.POST("/projects/:id/epics", middleware.RequireRoles("admin", "project_manager"), h.create)
	router.GET("/epics/:id", h.get)
	router.PATCH("/epics/:id", middleware.RequireRoles("admin", "project_manager"), h.update)
	router.DELETE("/epics/:id", middleware.RequireRoles("admin", "project_manager"), h.delete)
}

func (h *Handler) listByProject(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	projectID := c.Param("id")
	epics, err := h.service.List(c.Request.Context(), projectID)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, epics)
}

func (h *Handler) create(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload CreateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		// proceed to manual validation below
	}
	payload.ProjectID = c.Param("id")
	if payload.Title == "" {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "title is required")
		return
	}
	epic, err := h.service.Create(c.Request.Context(), payload)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.Created(c, epic)
}

func (h *Handler) get(c *gin.Context) {
	if middleware.CurrentUser(c) == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	epic, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if epic == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "epic not found")
		return
	}
	response.OK(c, epic)
}

func (h *Handler) update(c *gin.Context) {
	if middleware.CurrentUser(c) == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload UpdateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	epic, err := h.service.Update(c.Request.Context(), c.Param("id"), payload)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if epic == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "epic not found")
		return
	}
	response.OK(c, epic)
}

func (h *Handler) delete(c *gin.Context) {
	if middleware.CurrentUser(c) == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	deleted, err := h.service.Delete(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if !deleted {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "epic not found")
		return
	}
	response.NoContent(c)
}
