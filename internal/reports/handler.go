package reports

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/response"
)

// Handler handles HTTP requests for reports.
type Handler struct {
	service *Service
}

// NewHandler creates a new reports handler.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes attaches report endpoints.
func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/summary", h.getSummary)
	router.GET("/tickets/by-status", h.getByStatus)
	router.GET("/tickets/by-priority", h.getByPriority)
	router.GET("/tickets/by-assignee", h.getByAssignee)
	router.GET("/team-performance", h.getTeamPerformance)
	router.GET("/tickets/trend", h.getTicketTrend)
}

func (h *Handler) getSummary(c *gin.Context) {
	summary, err := h.service.GetSummary(c.Request.Context())
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, summary)
}

func (h *Handler) getByStatus(c *gin.Context) {
	breakdown, err := h.service.GetStatusBreakdown(c.Request.Context())
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, breakdown)
}

func (h *Handler) getByPriority(c *gin.Context) {
	breakdown, err := h.service.GetPriorityBreakdown(c.Request.Context())
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, breakdown)
}

func (h *Handler) getByAssignee(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	breakdown, err := h.service.GetAssigneeBreakdown(c.Request.Context(), limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, breakdown)
}

func (h *Handler) getTeamPerformance(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	performance, err := h.service.GetTeamPerformance(c.Request.Context(), limit)
	if err != nil {
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, performance)
}

func (h *Handler) getTicketTrend(c *gin.Context) {
	days, _ := strconv.Atoi(c.DefaultQuery("days", "30"))
	trend, err := h.service.GetTicketTrend(c.Request.Context(), days)
	if err != nil {
		fmt.Printf("GetTicketTrend error: %v\n", err)
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, trend)
}
