package audit

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Entry captures the audit log payload.
type Entry struct {
	ID          string    `json:"id"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	ActorID     *string   `json:"actorId,omitempty"`
	EntityType  *string   `json:"entityType,omitempty"`
	EntityID    *string   `json:"entityId,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// Repository manages persistence.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, limit int) ([]Entry, error) {
	const query = `
SELECT id, action, description, actor_id, entity_type, entity_id, created_at
FROM audit_log
ORDER BY created_at DESC
LIMIT $1`
	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Action, &e.Description, &e.ActorID, &e.EntityType, &e.EntityID, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (r *Repository) Insert(ctx context.Context, entry Entry) error {
	const query = `
INSERT INTO audit_log (id, action, description, actor_id, entity_type, entity_id)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, entry.ID, entry.Action, entry.Description, entry.ActorID, entry.EntityType, entry.EntityID)
	return err
}
