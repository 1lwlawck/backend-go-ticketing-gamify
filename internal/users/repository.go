package users

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles persistence for users.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	const query = `
SELECT id, name, username, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), COALESCE(bio, ''), created_at
FROM users
ORDER BY name`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []User
	for rows.Next() {
		var u User
		if err := rows.Scan(&u.ID, &u.Name, &u.Username, &u.Role, &u.AvatarURL, &u.Badges, &u.Bio, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id string) (*User, error) {
	const query = `
SELECT id, name, username, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), COALESCE(bio, ''), created_at
FROM users
WHERE id = $1`
	var u User
	if err := r.db.QueryRow(ctx, query, id).Scan(&u.ID, &u.Name, &u.Username, &u.Role, &u.AvatarURL, &u.Badges, &u.Bio, &u.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdateProfile(ctx context.Context, id string, input UpdateProfileInput) (*User, error) {
	const query = `
UPDATE users
SET name = COALESCE(NULLIF($2, ''), name),
    bio = $3,
    avatar_url = COALESCE(NULLIF($4, ''), avatar_url),
    updated_at = NOW()
WHERE id = $1
RETURNING id, name, username, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), COALESCE(bio, ''), created_at`
	var u User
	if err := r.db.QueryRow(ctx, query, id, input.Name, input.Bio, input.AvatarURL).Scan(
		&u.ID, &u.Name, &u.Username, &u.Role, &u.AvatarURL, &u.Badges, &u.Bio, &u.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdateRole(ctx context.Context, id, role string) (*User, error) {
	const query = `
UPDATE users
SET role = $2, updated_at = NOW()
WHERE id = $1
RETURNING id, name, username, role, COALESCE(avatar_url, ''), COALESCE(badges, ARRAY[]::text[]), COALESCE(bio, ''), created_at`
	var u User
	if err := r.db.QueryRow(ctx, query, id, role).Scan(
		&u.ID, &u.Name, &u.Username, &u.Role, &u.AvatarURL, &u.Badges, &u.Bio, &u.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &u, nil
}
