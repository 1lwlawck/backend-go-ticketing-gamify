package audit

import (
	"context"
	"fmt"
	"strings"
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

func (r *Repository) List(ctx context.Context, limit int, cursor *time.Time) ([]Entry, *string, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var (
		args []any
		idx  = 1
		sb   strings.Builder
	)
	sb.WriteString(`SELECT id, action, description, actor_id, entity_type, entity_id, created_at
FROM audit_log
WHERE 1=1`)
	if cursor != nil {
		sb.WriteString(fmt.Sprintf(" AND created_at < $%d", idx))
		args = append(args, *cursor)
		idx++
	}
	sb.WriteString(fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", idx))
	args = append(args, limit+1)

	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.Action, &e.Description, &e.ActorID, &e.EntityType, &e.EntityID, &e.CreatedAt); err != nil {
			return nil, nil, err
		}
		entries = append(entries, e)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	if len(entries) > limit {
		last := entries[limit-1]
		cursorStr := last.CreatedAt.Format(time.RFC3339Nano)
		entries = entries[:limit]
		return entries, &cursorStr, nil
	}
	return entries, nil, nil
}

func (r *Repository) Insert(ctx context.Context, entry Entry) error {
	const query = `
INSERT INTO audit_log (id, action, description, actor_id, entity_type, entity_id)
VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(ctx, query, entry.ID, entry.Action, entry.Description, entry.ActorID, entry.EntityType, entry.EntityID)
	return err
}
