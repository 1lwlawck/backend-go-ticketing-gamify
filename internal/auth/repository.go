package auth

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles user lookups.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// User represents user row.
type User struct {
	ID                       string
	Name                     string
	Username                 string
	Email                    *string
	EmailVerified            bool
	VerificationToken        *string
	VerificationTokenExpires *time.Time
	PasswordHash             string
	Role                     string
	AvatarURL                string
	Badges                   []string
	Bio                      *string
}

type CreateUserParams struct {
	ID                       string
	Name                     string
	Username                 string
	Email                    string
	VerificationToken        string
	VerificationTokenExpires time.Time
	PasswordHash             string
	Role                     string
	AvatarURL                string
	Badges                   []string
	Bio                      *string
}

type RefreshToken struct {
	ID        string
	UserID    string
	TokenHash string
	ExpiresAt time.Time
	RevokedAt *time.Time
	CreatedAt time.Time
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	const query = `
SELECT id, name, username, email, email_verified, verification_token, verification_token_expires,
       password_hash, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), bio
FROM users
WHERE username = $1`
	var u User
	if err := r.db.QueryRow(ctx, query, username).Scan(
		&u.ID,
		&u.Name,
		&u.Username,
		&u.Email,
		&u.EmailVerified,
		&u.VerificationToken,
		&u.VerificationTokenExpires,
		&u.PasswordHash,
		&u.Role,
		&u.AvatarURL,
		&u.Badges,
		&u.Bio,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*User, error) {
	const query = `
SELECT id, name, username, email, email_verified, verification_token, verification_token_expires,
       password_hash, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), bio
FROM users
WHERE id = $1`
	var u User
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Name,
		&u.Username,
		&u.Email,
		&u.EmailVerified,
		&u.VerificationToken,
		&u.VerificationTokenExpires,
		&u.PasswordHash,
		&u.Role,
		&u.AvatarURL,
		&u.Badges,
		&u.Bio,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UsernameExists(ctx context.Context, username string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE username = $1`
	if err := r.db.QueryRow(ctx, query, username).Scan(new(int)); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) CreateUser(ctx context.Context, params CreateUserParams) (*User, error) {
	const query = `
INSERT INTO users (id, name, username, email, email_verified, verification_token, verification_token_expires, password_hash, role, avatar_url, badges, bio, created_at, updated_at)
VALUES ($1, $2, $3, $4, false, $5, $6, $7, $8, $9, $10, $11, NOW(), NOW())
RETURNING id, name, username, email, email_verified, verification_token, verification_token_expires, password_hash, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), bio`
	var u User
	if err := r.db.QueryRow(
		ctx,
		query,
		params.ID,
		params.Name,
		params.Username,
		params.Email,
		params.VerificationToken,
		params.VerificationTokenExpires,
		params.PasswordHash,
		params.Role,
		params.AvatarURL,
		params.Badges,
		params.Bio,
	).Scan(
		&u.ID,
		&u.Name,
		&u.Username,
		&u.Email,
		&u.EmailVerified,
		&u.VerificationToken,
		&u.VerificationTokenExpires,
		&u.PasswordHash,
		&u.Role,
		&u.AvatarURL,
		&u.Badges,
		&u.Bio,
	); err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdatePassword(ctx context.Context, userID, newHash string) error {
	const query = `
UPDATE users
SET password_hash = $2, updated_at = NOW()
WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, userID, newHash)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) CreateRefreshToken(ctx context.Context, token RefreshToken) error {
	const query = `
INSERT INTO refresh_tokens (id, user_id, token_hash, expires_at, revoked_at, created_at)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, token.ID, token.UserID, token.TokenHash, token.ExpiresAt, token.RevokedAt, token.CreatedAt)
	return err
}

func (r *Repository) GetRefreshToken(ctx context.Context, id string) (*RefreshToken, error) {
	const query = `
SELECT id, user_id, token_hash, expires_at, revoked_at, created_at
FROM refresh_tokens
WHERE id = $1`
	var t RefreshToken
	if err := r.db.QueryRow(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.TokenHash, &t.ExpiresAt, &t.RevokedAt, &t.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &t, nil
}

func (r *Repository) RevokeRefreshToken(ctx context.Context, id string) error {
	const query = `UPDATE refresh_tokens SET revoked_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// Email verification methods

func (r *Repository) EmailExists(ctx context.Context, email string) (bool, error) {
	const query = `SELECT 1 FROM users WHERE email = $1`
	if err := r.db.QueryRow(ctx, query, email).Scan(new(int)); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) FindByVerificationToken(ctx context.Context, token string) (*User, error) {
	const query = `
SELECT id, name, username, email, email_verified, verification_token, verification_token_expires,
       password_hash, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), bio
FROM users
WHERE verification_token = $1`
	var u User
	if err := r.db.QueryRow(ctx, query, token).Scan(
		&u.ID,
		&u.Name,
		&u.Username,
		&u.Email,
		&u.EmailVerified,
		&u.VerificationToken,
		&u.VerificationTokenExpires,
		&u.PasswordHash,
		&u.Role,
		&u.AvatarURL,
		&u.Badges,
		&u.Bio,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) SetEmailVerified(ctx context.Context, userID string) error {
	const query = `
UPDATE users
SET email_verified = true, verification_token = NULL, verification_token_expires = NULL, updated_at = NOW()
WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) UpdateVerificationToken(ctx context.Context, userID, token string, expires time.Time) error {
	const query = `
UPDATE users
SET verification_token = $2, verification_token_expires = $3, updated_at = NOW()
WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, userID, token, expires)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
func (r *Repository) UpdateEmail(ctx context.Context, userID, newEmail string) error {
	const query = `
UPDATE users
SET email = $2, updated_at = NOW()
WHERE id = $1`
	tag, err := r.db.Exec(ctx, query, userID, newEmail)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}
