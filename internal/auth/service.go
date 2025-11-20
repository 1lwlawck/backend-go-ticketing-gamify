package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"backend-go-ticketing-gamify/internal/audit"
	"backend-go-ticketing-gamify/internal/gamification"
)

// ErrInvalidCredentials indicates login failure.
var ErrInvalidCredentials = errors.New("invalid credentials")

// Service coordinates authentication flows.
type Service struct {
	repo         *Repository
	gamification *gamification.Service
	audit        *audit.Service
	jwtSecret    string
}

func NewService(repo *Repository, gamificationSvc *gamification.Service, auditSvc *audit.Service, jwtSecret string) *Service {
	return &Service{
		repo:         repo,
		gamification: gamificationSvc,
		audit:        auditSvc,
		jwtSecret:    jwtSecret,
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

	return s.buildLoginResponse(user)
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
	return s.buildLoginResponse(user)
}

func (s *Service) buildLoginResponse(user *User) (*LoginResponse, error) {
	token, err := s.createToken(user)
	if err != nil {
		return nil, err
	}
	var bio string
	if user.Bio != nil {
		bio = *user.Bio
	}
	return &LoginResponse{
		Token: token,
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
	Token string     `json:"token"`
	User  UserPublic `json:"user"`
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
