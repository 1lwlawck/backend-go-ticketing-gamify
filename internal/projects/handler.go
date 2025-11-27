package projects

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/middleware"
	"backend-go-ticketing-gamify/internal/response"
)

// Handler exposes HTTP routes.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.list)
	router.POST("", middleware.RequireRoles("admin", "project_manager"), h.create)
	router.GET("/:id", h.get)
	router.POST("/:id/members", middleware.RequireRoles("admin", "project_manager"), h.addMember)
	router.POST("/:id/invites", middleware.RequireRoles("admin", "project_manager"), h.createInvite)
	router.DELETE("/:id/members/me", h.leaveSelf)
	router.POST("/join", h.joinByCode)
}

func (h *Handler) list(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
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
	filter := ListFilter{
		Limit:  limit,
		Search: c.Query("q"),
		Status: c.Query("status"),
		Cursor: cursorPtr,
	}
	projects, err := h.service.List(c.Request.Context(), user, filter)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
		} else {
			response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		}
		return
	}
	meta := gin.H{"limit": limit}
	if len(projects) == limit {
		last := projects[len(projects)-1]
		meta["nextCursor"] = last.CreatedAt.Format(time.RFC3339Nano)
	}
	response.WithMeta(c, http.StatusOK, projects, meta)
}

func (h *Handler) get(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("activityLimit", "25"))
	if limit <= 0 || limit > 200 {
		limit = 25
	}
	var cursorPtr *time.Time
	if cursorStr := c.Query("activityCursor"); cursorStr != "" {
		if ts, err := time.Parse(time.RFC3339, cursorStr); err == nil {
			cursorPtr = &ts
		}
	}
	activityFilter := ActivityFilter{
		Limit:  limit,
		Search: c.Query("activitySearch"),
		Cursor: cursorPtr,
	}

	project, err := h.service.Get(c.Request.Context(), user, c.Param("id"), &activityFilter)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if project == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "project not found")
		return
	}
	response.OK(c, project)
}

func (h *Handler) create(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}

	var payload CreateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if payload.Name == "" || payload.Description == "" {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "name and description are required")
		return
	}

	project, err := h.service.Create(c.Request.Context(), user, payload)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.Created(c, project)
}

func (h *Handler) addMember(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload AddMemberInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.service.AddMember(c.Request.Context(), user, c.Param("id"), payload); err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) createInvite(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload InviteInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	invite, err := h.service.CreateInvite(c.Request.Context(), user, c.Param("id"), payload)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.Created(c, invite)
}

func (h *Handler) joinByCode(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload struct {
		Code string `json:"code" binding:"required"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	project, err := h.service.JoinByCode(c.Request.Context(), user, payload.Code)
	if err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	response.OK(c, project)
}

func (h *Handler) leaveSelf(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	if err := h.service.Leave(c.Request.Context(), user, c.Param("id")); err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
		case errors.Is(err, ErrNotMember):
			response.ErrorCode(c, http.StatusBadRequest, "validation_error", "you are not a member of this project")
		default:
			response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		}
		return
	}
	c.Status(http.StatusNoContent)
}
