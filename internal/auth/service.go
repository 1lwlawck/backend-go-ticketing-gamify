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
var ErrEmailNotVerified = errors.New("email not verified")

// EmailSender is the interface for sending emails
type EmailSender interface {
	SendVerificationEmail(to, name, token, frontendURL string) error
	IsConfigured() bool
}

// Service coordinates authentication flows.
type Service struct {
	repo         *Repository
	gamification *gamification.Service
	audit        *audit.Service
	email        EmailSender
	jwtSecret    string
	frontendURL  string
	refreshTTL   time.Duration
}

func NewService(repo *Repository, gamificationSvc *gamification.Service, auditSvc *audit.Service, emailSvc EmailSender, jwtSecret, frontendURL string) *Service {
	return &Service{
		repo:         repo,
		gamification: gamificationSvc,
		audit:        auditSvc,
		email:        emailSvc,
		jwtSecret:    jwtSecret,
		frontendURL:  frontendURL,
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

// Register creates a new user and sends verification email.
func (s *Service) Register(ctx context.Context, input RegisterInput) (*RegisterResponse, error) {
	// Basic required field validation
	if input.Name == "" || input.Username == "" || input.Password == "" || input.Email == "" {
		return nil, fmt.Errorf("name, email, username, and password are required")
	}

	// Email validation
	if !isValidEmail(input.Email) {
		return nil, fmt.Errorf("invalid email format")
	}

	// Name validation
	if len(input.Name) < 2 {
		return nil, fmt.Errorf("name must be at least 2 characters")
	}
	if len(input.Name) > 50 {
		return nil, fmt.Errorf("name must be less than 50 characters")
	}

	// Username validation
	if len(input.Username) < 3 {
		return nil, fmt.Errorf("username must be at least 3 characters")
	}
	if len(input.Username) > 30 {
		return nil, fmt.Errorf("username must be less than 30 characters")
	}
	if !isValidUsername(input.Username) {
		return nil, fmt.Errorf("username must start with a letter and contain only letters, numbers, underscores, and dots")
	}

	// Password validation
	if len(input.Password) < 8 {
		return nil, fmt.Errorf("password must be at least 8 characters")
	}
	if !hasUppercase(input.Password) {
		return nil, fmt.Errorf("password must contain at least one uppercase letter")
	}
	if !hasLowercase(input.Password) {
		return nil, fmt.Errorf("password must contain at least one lowercase letter")
	}
	if !hasNumber(input.Password) {
		return nil, fmt.Errorf("password must contain at least one number")
	}
	if !hasSpecialChar(input.Password) {
		return nil, fmt.Errorf("password must contain at least one special character")
	}

	// Check username uniqueness
	exists, err := s.repo.UsernameExists(ctx, input.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("username already exists")
	}

	// Check email uniqueness
	emailExists, err := s.repo.EmailExists(ctx, input.Email)
	if err != nil {
		return nil, err
	}
	if emailExists {
		return nil, fmt.Errorf("email already registered")
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Generate verification token
	verificationToken := uuid.NewString()
	tokenExpires := time.Now().Add(24 * time.Hour)

	params := CreateUserParams{
		ID:                       uuid.NewString(),
		Name:                     input.Name,
		Username:                 input.Username,
		Email:                    input.Email,
		VerificationToken:        verificationToken,
		VerificationTokenExpires: tokenExpires,
		PasswordHash:             string(hashed),
		Role:                     input.Role,
		AvatarURL:                input.AvatarURL,
		Badges:                   input.Badges,
		Bio:                      input.Bio,
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

	// Send verification email
	if s.email != nil {
		_ = s.email.SendVerificationEmail(input.Email, input.Name, verificationToken, s.frontendURL)
	}

	return &RegisterResponse{
		Message:           "Registration successful. Please check your email to verify your account.",
		NeedsVerification: true,
		UserID:            user.ID,
	}, nil
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
	var email string
	if user.Email != nil {
		email = *user.Email
	}
	return &LoginResponse{
		Token:         token,
		RefreshToken:  refreshToken,
		EmailVerified: user.EmailVerified,
		User: UserPublic{
			ID:            user.ID,
			Name:          user.Name,
			Username:      user.Username,
			Email:         email,
			EmailVerified: user.EmailVerified,
			Role:          user.Role,
			AvatarURL:     user.AvatarURL,
			Badges:        user.Badges,
			Bio:           bio,
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
	Token         string     `json:"token"`
	RefreshToken  string     `json:"refreshToken"`
	User          UserPublic `json:"user"`
	EmailVerified bool       `json:"emailVerified"`
}

// RegisterResponse returned after registration.
type RegisterResponse struct {
	Message           string `json:"message"`
	NeedsVerification bool   `json:"needsVerification"`
	UserID            string `json:"userId"`
}

// UserPublic is safe subset of user profile.
type UserPublic struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Username      string   `json:"username"`
	Email         string   `json:"email,omitempty"`
	EmailVerified bool     `json:"emailVerified"`
	Role          string   `json:"role"`
	AvatarURL     string   `json:"avatarUrl"`
	Badges        []string `json:"badges"`
	Bio           string   `json:"bio,omitempty"`
}

// RegisterInput captures registration payload.
type RegisterInput struct {
	Name      string   `json:"name"`
	Email     string   `json:"email"`
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

// Validation helper functions

func isValidUsername(username string) bool {
	if len(username) == 0 {
		return false
	}
	// Must start with a letter
	first := username[0]
	if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')) {
		return false
	}
	// Must contain only letters, numbers, underscores, and dots
	for _, c := range username {
		if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '.') {
			return false
		}
	}
	return true
}

func hasUppercase(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

func hasLowercase(s string) bool {
	for _, c := range s {
		if c >= 'a' && c <= 'z' {
			return true
		}
	}
	return false
}

func hasNumber(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

func hasSpecialChar(s string) bool {
	specialChars := "!@#$%^&*(),.?\":{}|<>"
	for _, c := range s {
		if strings.ContainsRune(specialChars, c) {
			return true
		}
	}
	return false
}

func isValidEmail(email string) bool {
	// Basic email validation
	if len(email) < 5 || len(email) > 254 {
		return false
	}
	atIndex := strings.LastIndex(email, "@")
	if atIndex < 1 || atIndex == len(email)-1 {
		return false
	}
	domain := email[atIndex+1:]
	if !strings.Contains(domain, ".") {
		return false
	}
	return true
}

// VerifyEmail verifies user's email with the provided token
func (s *Service) VerifyEmail(ctx context.Context, token string) error {
	fmt.Printf("[AUTH] Verifying email with token: %s\n", token)
	if token == "" {
		return fmt.Errorf("verification token is required")
	}

	user, err := s.repo.FindByVerificationToken(ctx, token)
	if err != nil {
		fmt.Printf("[AUTH] Error finding token: %v\n", err)
		return err
	}
	if user == nil {
		fmt.Printf("[AUTH] No user found for token: %s\n", token)
		return fmt.Errorf("invalid or expired verification token")
	}

	// Check if token is expired
	if user.VerificationTokenExpires != nil && user.VerificationTokenExpires.Before(time.Now()) {
		fmt.Printf("[AUTH] Token expired for user: %s\n", user.ID)
		return fmt.Errorf("invalid or expired verification token")
	}

	// Mark email as verified
	if err := s.repo.SetEmailVerified(ctx, user.ID); err != nil {
		fmt.Printf("[AUTH] Failed to set verified: %v\n", err)
		return err
	}

	// Log the verification
	if s.audit != nil {
		action := "email_verified"
		description := fmt.Sprintf("%s verified their email", user.Name)
		actorID := user.ID
		entityType := "user"
		entityID := user.ID
		_ = s.audit.Log(ctx, action, description, &actorID, &entityType, &entityID)
	}

	return nil
}

// ResendVerification generates a new verification token and sends email
func (s *Service) ResendVerification(ctx context.Context, userID string) error {
	if userID == "" {
		return fmt.Errorf("user ID is required")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	if user.Email == nil || *user.Email == "" {
		return fmt.Errorf("user has no email address")
	}

	// Generate new token
	newToken := uuid.NewString()
	tokenExpires := time.Now().Add(24 * time.Hour)

	if err := s.repo.UpdateVerificationToken(ctx, userID, newToken, tokenExpires); err != nil {
		return err
	}

	// Send verification email
	if s.email != nil {
		return s.email.SendVerificationEmail(*user.Email, user.Name, newToken, s.frontendURL)
	}

	return nil
}

func (s *Service) UpdateUnverifiedEmail(ctx context.Context, userID, password, newEmail string) error {
	if !isValidEmail(newEmail) {
		return fmt.Errorf("invalid email format")
	}

	user, err := s.repo.FindByID(ctx, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return fmt.Errorf("user not found")
	}

	if user.EmailVerified {
		return fmt.Errorf("email already verified")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return ErrInvalidCredentials
	}

	// Check if new email exists
	exists, err := s.repo.EmailExists(ctx, newEmail)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("email already in use")
	}

	// Update email
	if err := s.repo.UpdateEmail(ctx, userID, newEmail); err != nil {
		return err
	}

	// New token and resend
	newToken := uuid.NewString()
	tokenExpires := time.Now().Add(24 * time.Hour)
	fmt.Printf("[AUTH] Updating email for user %s to %s with new token: %s\n", userID, newEmail, newToken)

	if err := s.repo.UpdateVerificationToken(ctx, userID, newToken, tokenExpires); err != nil {
		return err
	}

	if s.email != nil {
		return s.email.SendVerificationEmail(newEmail, user.Name, newToken, s.frontendURL)
	}

	return nil
}
