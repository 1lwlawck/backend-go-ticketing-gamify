package auth

import (
	"context"

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
	ID           string
	Name         string
	Username     string
	PasswordHash string
	Role         string
	AvatarURL    string
	Badges       []string
	Bio          *string
}

type CreateUserParams struct {
	ID           string
	Name         string
	Username     string
	PasswordHash string
	Role         string
	AvatarURL    string
	Badges       []string
	Bio          *string
}

func (r *Repository) FindByUsername(ctx context.Context, username string) (*User, error) {
	const query = `
SELECT id, name, username, password_hash, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), bio
FROM users
WHERE username = $1`
	var u User
	if err := r.db.QueryRow(ctx, query, username).Scan(
		&u.ID,
		&u.Name,
		&u.Username,
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
INSERT INTO users (id, name, username, password_hash, role, avatar_url, badges, bio, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
RETURNING id, name, username, password_hash, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), bio`
	var u User
	if err := r.db.QueryRow(
		ctx,
		query,
		params.ID,
		params.Name,
		params.Username,
		params.PasswordHash,
		params.Role,
		params.AvatarURL,
		params.Badges,
		params.Bio,
	).Scan(
		&u.ID,
		&u.Name,
		&u.Username,
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
