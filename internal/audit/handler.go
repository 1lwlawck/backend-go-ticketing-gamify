package audit

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/response"
)

// Handler provides HTTP endpoints.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("", h.list)
}

func (h *Handler) list(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	entries, err := h.service.List(c.Request.Context(), limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.WithMeta(c, http.StatusOK, entries, gin.H{"limit": limit})
}
