package auth

import (
	"backend-go-ticketing-gamify/internal/middleware"
	"net/http"

	"github.com/gin-gonic/gin"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	result, err := h.service.Login(c.Request.Context(), payload.Username, payload.Password)
	if err != nil {
		status := http.StatusInternalServerError
		if err == ErrInvalidCredentials {
			status = http.StatusUnauthorized
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, result)
}

type changePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required"`
}

func (h *Handler) changePassword(c *gin.Context) {
	user := middleware.CurrentUser(c)
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthenticated"})
		return
	}
	var payload changePasswordRequest
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if payload.NewPassword == payload.OldPassword {
		c.JSON(http.StatusBadRequest, gin.H{"error": "new password must differ from old password"})
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), user.ID, payload.OldPassword, payload.NewPassword); err != nil {
		if err == ErrInvalidCredentials {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid old password"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
