package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"crypto/sha256"
	"encoding/hex"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend-go-ticketing-gamify/internal/audit"
	"backend-go-ticketing-gamify/internal/gamification"
)

// ErrInvalidCredentials indicates login failure.
var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrForbidden = errors.New("forbidden")

// Service coordinates authentication flows.
type Service struct {
	repo         *Repository
	gamification *gamification.Service
	audit        *audit.Service
	jwtSecret    string
	refreshTTL   time.Duration
}

func NewService(repo *Repository, gamificationSvc *gamification.Service, auditSvc *audit.Service, jwtSecret string) *Service {
	return &Service{
		repo:         repo,
		gamification: gamificationSvc,
		audit:        auditSvc,
		jwtSecret:    jwtSecret,
		refreshTTL:   7 * 24 * time.Hour,
	}
}

// Login validates credentials and returns token.
func (s *Service) Login(ctx context.Context, username, password string) (*LoginResponse, error) {
	if username == "" || password == "" {
		return nil, ErrInvalidCredentials
	}

	user, err := s.repo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.buildLoginResponse(ctx, user, "")
}

// Register creates a new user and immediately issues a token.
func (s *Service) Register(ctx context.Context, input RegisterInput) (*LoginResponse, error) {
	if input.Name == "" || input.Username == "" || input.Password == "" {
		return nil, fmt.Errorf("name, username, and password are required")
	}
	exists, err := s.repo.UsernameExists(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	params := CreateUserParams{
		ID:           uuid.NewString(),
		Name:         input.Name,
		Username:     input.Username,
		PasswordHash: string(hashed),
		Role:         input.Role,
		AvatarURL:    input.AvatarURL,
		Badges:       input.Badges,
		Bio:          input.Bio,
	}
	if params.Role == "" {
		params.Role = "developer"
	}
	if params.Badges == nil {
		params.Badges = []string{"Initiate"}
	}

	user, err := s.repo.CreateUser(ctx, params)
	if err != nil {
		return nil, err
	}

	if s.gamification != nil {
		_ = s.gamification.EnsureUser(ctx, user.ID)
	}
	if s.audit != nil {
		action := "user_registered"
		description := fmt.Sprintf("%s joined the workspace", user.Name)
		actorID := user.ID
		entityType := "user"
		entityID := user.ID
		_ = s.audit.Log(ctx, action, description, &actorID, &entityType, &entityID)
	}
	return s.buildLoginResponse(ctx, user, "")
}

// ChangePassword lets an authenticated user rotate password.
func (s *Service) ChangePassword(ctx context.Context, userID, oldPassword, newPassword string) error {
	if userID == "" || oldPassword == "" || newPassword == "" {
		return fmt.Errorf("old and new password are required")
	}
	u, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if u == nil {
		return ErrInvalidCredentials
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(oldPassword)); err != nil {
		return ErrInvalidCredentials
	}
	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	if err := s.repo.UpdatePassword(ctx, userID, string(newHash)); err != nil {
		return err
	}
	if s.audit != nil {
		action := "password_changed"
		desc := fmt.Sprintf("%s changed password", u.Username)
		actorID := u.ID
		entityType := "user"
		entityID := u.ID
		_ = s.audit.Log(ctx, action, desc, &actorID, &entityType, &entityID)
	}
	return nil
}

func (s *Service) buildLoginResponse(ctx context.Context, user *User, existingRefreshID string) (*LoginResponse, error) {
	token, err := s.createToken(user)
	if err != nil {
		return nil, err
	}
	refreshToken, err := s.issueRefreshToken(ctx, user.ID, existingRefreshID)
	if err != nil {
		return nil, err
	}
	var bio string
	if user.Bio != nil {
		bio = *user.Bio
	}
	return &LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User: UserPublic{
			ID:        user.ID,
			Name:      user.Name,
			Username:  user.Username,
			Role:      user.Role,
			AvatarURL: user.AvatarURL,
			Badges:    user.Badges,
			Bio:       bio,
		},
	}, nil
}

func (s *Service) createToken(user *User) (string, error) {
	claims := jwt.MapClaims{
		"sub":  user.ID,
		"name": user.Name,
		"role": user.Role,
		"exp":  time.Now().Add(24 * time.Hour).Unix(),
		"iat":  time.Now().Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.jwtSecret))
}

// LoginResponse returned to clients.
type LoginResponse struct {
	Token        string     `json:"token"`
	RefreshToken string     `json:"refreshToken"`
	User         UserPublic `json:"user"`
}

// UserPublic is safe subset of user profile.
type UserPublic struct {
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Username  string   `json:"username"`
	Role      string   `json:"role"`
	AvatarURL string   `json:"avatarUrl"`
	Badges    []string `json:"badges"`
	Bio       string   `json:"bio,omitempty"`
}

// RegisterInput captures registration payload.
type RegisterInput struct {
	Name      string   `json:"name"`
	Username  string   `json:"username"`
	Password  string   `json:"password"`
	Role      string   `json:"role"`
	AvatarURL string   `json:"avatarUrl"`
	Badges    []string `json:"badges"`
	Bio       *string  `json:"bio"`
}

// Refresh validates a refresh token, rotates it, and issues new tokens.
func (s *Service) Refresh(ctx context.Context, refreshPlain string) (*LoginResponse, error) {
	refreshID, secret, err := splitRefresh(refreshPlain)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	stored, err := s.repo.GetRefreshToken(ctx, refreshID)
	if err != nil {
		return nil, err
	}
	if stored == nil || stored.RevokedAt != nil || time.Now().After(stored.ExpiresAt) {
		return nil, ErrInvalidCredentials
	}
	secretHash := hashRefresh(secret)
	if secretHash != stored.TokenHash {
		return nil, ErrInvalidCredentials
	}
	user, err := s.repo.FindByID(ctx, stored.UserID)
	if err != nil || user == nil {
		return nil, ErrInvalidCredentials
	}
	// rotate: revoke old and issue new
	_ = s.repo.RevokeRefreshToken(ctx, stored.ID)
	return s.buildLoginResponse(ctx, user, stored.ID)
}

func (s *Service) issueRefreshToken(ctx context.Context, userID string, previousID string) (string, error) {
	secret := uuid.NewString() + uuid.NewString()
	secret = strings.ReplaceAll(secret, "-", "")
	refreshID := uuid.NewString()
	now := time.Now()
	token := RefreshToken{
		ID:        refreshID,
		UserID:    userID,
		TokenHash: hashRefresh(secret),
		ExpiresAt: now.Add(s.refreshTTL),
		RevokedAt: nil,
		CreatedAt: now,
	}
	if err := s.repo.CreateRefreshToken(ctx, token); err != nil {
		return "", err
	}
	return refreshID + "." + secret, nil
}

func splitRefresh(raw string) (string, string, error) {
	parts := strings.Split(raw, ".")
	if len(parts) != 2 {
		return "", "", ErrInvalidCredentials
	}
	return parts[0], parts[1], nil
}

func hashRefresh(secret string) string {
	sum := sha256.Sum256([]byte(secret))
	return hex.EncodeToString(sum[:])
}
