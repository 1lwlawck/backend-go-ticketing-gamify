package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"backend-go-ticketing-gamify/internal/middleware"
	"backend-go-ticketing-gamify/internal/response"
)

// Handler exposes HTTP routes for auth.
type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) RegisterRoutes(router *gin.RouterGroup) {
	router.POST("/login", h.login)
	router.POST("/register", h.register)
	router.POST("/refresh", h.refresh)
}

// RegisterProtected mounts routes that need authentication.
func (h *Handler) RegisterProtected(router *gin.RouterGroup) {
	router.POST("/change-password", h.changePassword)
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *Handler) login(c *gin.Context) {
	var payload loginRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	result, err := h.service.Login(c.Request.Context(), payload.Username, payload.Password)
	if err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		if err == ErrInvalidCredentials {
			status = http.StatusUnauthorized
			code = "invalid_credentials"
		}
		response.ErrorCode(c, status, code, err.Error())
		return
	}
	response.OK(c, result)
}

type registerRequest struct {
	Name      string   `json:"name" binding:"required"`
	Username  string   `json:"username" binding:"required"`
	Password  string   `json:"password" binding:"required"`
	Role      string   `json:"role"`
	AvatarURL string   `json:"avatarUrl"`
	Badges    []string `json:"badges"`
	Bio       *string  `json:"bio"`
}

func (h *Handler) register(c *gin.Context) {
	var payload registerRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if payload.Role != "" && !isAllowedRole(payload.Role) {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "invalid role")
		return
	}
	result, err := h.service.Register(c.Request.Context(), RegisterInput{
		Name:      payload.Name,
		Username:  payload.Username,
		Password:  payload.Password,
		Role:      payload.Role,
		AvatarURL: payload.AvatarURL,
		Badges:    payload.Badges,
		Bio:       payload.Bio,
	})
	if err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	response.Created(c, result)
}

type refreshRequest struct {
	RefreshToken string `json:"refreshToken" binding:"required"`
}

func (h *Handler) refresh(c *gin.Context) {
	var payload refreshRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	result, err := h.service.Refresh(c.Request.Context(), payload.RefreshToken)
	if err != nil {
		if err == ErrInvalidCredentials {
			response.ErrorCode(c, http.StatusUnauthorized, "invalid_refresh_token", "invalid refresh token")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	response.OK(c, result)
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *Handler) changePassword(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		response.ErrorCode(c, http.StatusUnauthorized, "unauthenticated", "unauthenticated")
		return
	}
	var payload changePasswordRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if payload.NewPassword == payload.OldPassword {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", "new password must differ from old password")
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), user.ID, payload.OldPassword, payload.NewPassword); err != nil {
		if err == ErrInvalidCredentials {
			response.ErrorCode(c, http.StatusUnauthorized, "invalid_credentials", "invalid old password")
			return
		}
		response.ErrorCode(c, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}
	c.Status(http.StatusNoContent)
}

func isAllowedRole(role string) bool {
	switch role {
	case "admin", "project_manager", "developer", "viewer":
		return true
	default:
		return false
	}
}
