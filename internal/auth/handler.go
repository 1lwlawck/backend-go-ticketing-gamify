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
	router.POST("/verify-email", h.verifyEmail)
	router.POST("/resend-verification", h.resendVerification)
	router.POST("/update-unverified-email", h.updateUnverifiedEmail)
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
	Email     string   `json:"email" binding:"required"`
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
		Email:     payload.Email,
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

// Email verification handlers

type verifyEmailRequest struct {
	Token string `json:"token" binding:"required"`
}

func (h *Handler) verifyEmail(c *gin.Context) {
	var payload verifyEmailRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.service.VerifyEmail(c.Request.Context(), payload.Token); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "verification_failed", err.Error())
		return
	}
	response.OK(c, gin.H{"message": "Email verified successfully"})
}

type resendVerificationRequest struct {
	UserID string `json:"userId" binding:"required"`
}

func (h *Handler) resendVerification(c *gin.Context) {
	var payload resendVerificationRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.service.ResendVerification(c.Request.Context(), payload.UserID); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "resend_failed", err.Error())
		return
	}

	response.OK(c, gin.H{"message": "Verification email sent"})
}

type updateUnverifiedEmailRequest struct {
	UserID   string `json:"userId" binding:"required"`
	Password string `json:"password" binding:"required"`
	NewEmail string `json:"newEmail" binding:"required"`
}

func (h *Handler) updateUnverifiedEmail(c *gin.Context) {
	var payload updateUnverifiedEmailRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		response.ErrorCode(c, http.StatusBadRequest, "validation_error", err.Error())
		return
	}
	if err := h.service.UpdateUnverifiedEmail(c.Request.Context(), payload.UserID, payload.Password, payload.NewEmail); err != nil {
		status := http.StatusInternalServerError
		code := "internal_error"
		if err == ErrInvalidCredentials {
			status = http.StatusUnauthorized
			code = "invalid_credentials"
		} else if err.Error() == "email already verified" {
			status = http.StatusBadRequest
			code = "already_verified"
		} else if err.Error() == "email already in use" {
			status = http.StatusConflict
			code = "email_in_use"
		}
		response.ErrorCode(c, status, code, err.Error())
		return
	}
	response.OK(c, nil)
}

func isAllowedRole(role string) bool {
	switch role {
	case "admin", "project_manager", "developer", "viewer":
		return true
	default:
		return false
	}
}
