package tickets

import (
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/middleware"
	"backend-go-ticketing-gamify/internal/response"
)

// Handler exposes ticket routes.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.list)
	router.POST("", h.create)
	router.GET("/:id", h.get)
	router.PATCH("/:id/status", h.updateStatus)
	router.PATCH("/:id/details", h.updateDetails)
	router.PATCH("/:id/epic", h.updateEpic)
	router.POST("/:id/comments", h.addComment)
	router.PATCH("/comments/:commentId", h.updateComment)
	router.DELETE("/comments/:commentId", h.deleteComment)
	router.DELETE("/:id", h.delete)
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	projectID := c.Query("projectId")
	if strings.EqualFold(projectID, "all") {
		projectID = ""
	}
	var cursorPtr *time.Time
	if cursorStr := c.Query("cursor"); cursorStr != "" {
		if ts, err := time.Parse(time.RFC3339, cursorStr); err == nil {
			cursorPtr = &ts
		}
	}
	filter := Filter{
		ProjectID:  projectID,
		AssigneeID: c.Query("assigneeId"),
		Status:     c.Query("status"),
		EpicID:     c.Query("epicId"),
		Search:     c.Query("q"),
		Cursor:     cursorPtr,
		Limit:      limit,
	}
	tickets, err := h.service.List(c.Request.Context(), filter)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	meta := gin.H{"limit": limit}
	if len(tickets) == limit {
		// keyset pagination using last item's createdAt
		last := tickets[len(tickets)-1]
		meta["nextCursor"] = last.CreatedAt.Format(time.RFC3339Nano)
	}
	response.WithMeta(c, http.StatusOK, tickets, meta)
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
	if err := validateTicketEnums("", payload.Priority, payload.Type); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	payload.ReporterID = user.ID
	ticket, err := h.service.Create(c.Request.Context(), user, payload)
	if err != nil {
		if errors.Is(err, ErrEpicProjectMismatch) {
			response.ErrorCode(c, http.StatusBadRequest, "validation_error", "epic must belong to the same project")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.Created(c, ticket)
}

func (h *Handler) get(c *gin.Context) {
	ticket, err := h.service.Get(c.Request.Context(), c.Param("id"))
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if ticket == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "ticket not found")
		return
	}
	response.OK(c, ticket)
}

func (h *Handler) updateStatus(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload UpdateStatusInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := validateTicketEnums(payload.Status, "", ""); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	ticket, err := h.service.UpdateStatus(c.Request.Context(), user, c.Param("id"), payload.Status)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if ticket == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "ticket not found")
		return
	}
	response.OK(c, ticket)
}

func (h *Handler) updateDetails(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload UpdateInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := validateTicketEnums("", deref(payload.Priority), deref(payload.Type)); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if payload.EpicID != nil && *payload.EpicID == "" {
		payload.EpicID = nil
	}
	ticket, err := h.service.UpdateDetails(c.Request.Context(), user, c.Param("id"), payload)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		if errors.Is(err, ErrEpicProjectMismatch) {
			response.ErrorCode(c, http.StatusBadRequest, "validation_error", "epic must belong to the same project")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if ticket == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "ticket not found")
		return
	}
	response.OK(c, ticket)
}

func (h *Handler) updateEpic(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload struct {
		EpicID *string `json:"epicId"`
	}
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	ticket, err := h.service.UpdateDetails(c.Request.Context(), user, c.Param("id"), UpdateInput{EpicID: payload.EpicID})
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		if errors.Is(err, ErrEpicProjectMismatch) {
			response.ErrorCode(c, http.StatusBadRequest, "validation_error", "epic must belong to the same project")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if ticket == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "ticket not found")
		return
	}
	response.OK(c, ticket)
}

func (h *Handler) addComment(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload CommentInput
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	comment, err := h.service.AddComment(c.Request.Context(), user, c.Param("id"), payload.Text)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.Created(c, comment)
}

func (h *Handler) updateComment(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload CommentUpdate
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	comment, err := h.service.UpdateComment(c.Request.Context(), user, c.Param("commentId"), payload.Text)
	if err != nil {
		if errors.Is(err, ErrForbidden) {
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	if comment == nil {
		response.ErrorCode(c, http.StatusNotFound, "not_found", "comment not found")
		return
	}
	response.OK(c, comment)
}

func (h *Handler) deleteComment(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	if err := h.service.DeleteComment(c.Request.Context(), user, c.Param("commentId")); err != nil {
		if errors.Is(err, ErrForbidden) {
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		if errors.Is(err, ErrNotFound) {
			response.ErrorCode(c, http.StatusNotFound, "not_found", "comment not found")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) delete(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	if err := h.service.Delete(c.Request.Context(), user, c.Param("id")); err != nil {
		if errors.Is(err, ErrForbidden) {
			response.ErrorCode(c, http.StatusForbidden, "forbidden", "forbidden")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func validateTicketEnums(status, priority, typ string) error {
	if status != "" {
		switch status {
		case "todo", "backlog", "in_progress", "review", "done":
		default:
			return errors.New("invalid status")
		}
	}
	if priority != "" {
		switch priority {
		case "low", "medium", "high", "urgent":
		default:
			return errors.New("invalid priority")
		}
	}
	if typ != "" {
		switch typ {
		case "feature", "bug", "chore":
		default:
			return errors.New("invalid type")
		}
	}
	return nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
