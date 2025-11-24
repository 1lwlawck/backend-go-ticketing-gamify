package epics

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListByProject(ctx context.Context, projectID string) ([]Epic, error) {
	const query = `
SELECT e.id, e.project_id, e.title, e.description, e.status, e.start_date, e.due_date, e.owner_id, e.created_at, e.updated_at,
       COALESCE(done.count, 0) AS done_count,
	   COALESCE(total.count, 0) AS total_count
FROM epics e
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS count FROM tickets t WHERE t.epic_id = e.id AND t.status = 'done'
) done ON true
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS count FROM tickets t WHERE t.epic_id = e.id
) total ON true
WHERE e.project_id = $1
ORDER BY e.created_at DESC`
	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var epics []Epic
	for rows.Next() {
		var e Epic
		if err := rows.Scan(&e.ID, &e.ProjectID, &e.Title, &e.Description, &e.Status, &e.StartDate, &e.DueDate, &e.OwnerID, &e.CreatedAt, &e.UpdatedAt, &e.DoneCount, &e.TotalCount); err != nil {
			return nil, err
		}
		epics = append(epics, e)
	}
	return epics, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id string) (*Epic, error) {
	const query = `
SELECT e.id, e.project_id, e.title, e.description, e.status, e.start_date, e.due_date, e.owner_id, e.created_at, e.updated_at,
       COALESCE(done.count, 0) AS done_count,
	   COALESCE(total.count, 0) AS total_count
FROM epics e
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS count FROM tickets t WHERE t.epic_id = e.id AND t.status = 'done'
) done ON true
LEFT JOIN LATERAL (
    SELECT COUNT(*)::int AS count FROM tickets t WHERE t.epic_id = e.id
) total ON true
WHERE e.id = $1`
	var e Epic
	if err := r.db.QueryRow(ctx, query, id).Scan(&e.ID, &e.ProjectID, &e.Title, &e.Description, &e.Status, &e.StartDate, &e.DueDate, &e.OwnerID, &e.CreatedAt, &e.UpdatedAt, &e.DoneCount, &e.TotalCount); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &e, nil
}

func (r *Repository) Create(ctx context.Context, input CreateInput) (*Epic, error) {
	const query = `
INSERT INTO epics (project_id, title, description, status, start_date, due_date, owner_id)
VALUES ($1, $2, $3, COALESCE($4, 'backlog')::ticket_status, $5, $6, $7)
RETURNING id, project_id, title, description, status, start_date, due_date, owner_id, created_at, updated_at`
	var e Epic
	if err := r.db.QueryRow(ctx, query, input.ProjectID, input.Title, input.Description, input.Status, input.StartDate, input.DueDate, input.OwnerID).
		Scan(&e.ID, &e.ProjectID, &e.Title, &e.Description, &e.Status, &e.StartDate, &e.DueDate, &e.OwnerID, &e.CreatedAt, &e.UpdatedAt); err != nil {
		return nil, err
	}
	e.DoneCount = 0
	e.TotalCount = 0
	return &e, nil
}

func (r *Repository) Update(ctx context.Context, id string, input UpdateInput) (*Epic, error) {
	setParts := []string{}
	args := []any{}
	idx := 1

	if input.Title != nil {
		setParts = append(setParts, fmt.Sprintf("title = $%d", idx))
		args = append(args, *input.Title)
		idx++
	}
	if input.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", idx))
		args = append(args, *input.Description)
		idx++
	}
	if input.Status != nil {
		setParts = append(setParts, fmt.Sprintf("status = $%d::ticket_status", idx))
		args = append(args, *input.Status)
		idx++
	}
	if input.StartDate != nil {
		setParts = append(setParts, fmt.Sprintf("start_date = $%d", idx))
		args = append(args, input.StartDate)
		idx++
	} else if input.ClearStart {
		setParts = append(setParts, "start_date = NULL")
	}
	if input.DueDate != nil {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", idx))
		args = append(args, input.DueDate)
		idx++
	} else if input.ClearDue {
		setParts = append(setParts, "due_date = NULL")
	}
	if input.OwnerID != nil {
		setParts = append(setParts, fmt.Sprintf("owner_id = $%d", idx))
		args = append(args, input.OwnerID)
		idx++
	}

	if len(setParts) == 0 {
		return r.Get(ctx, id)
	}
	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now())
	idx++
	args = append(args, id)

	query := fmt.Sprintf(`UPDATE epics SET %s WHERE id = $%d`, strings.Join(setParts, ", "), idx)
	if _, err := r.db.Exec(ctx, query, args...); err != nil {
		return nil, err
	}
	return r.Get(ctx, id)
}

func (r *Repository) Delete(ctx context.Context, id string) (bool, error) {
	const clearTickets = `UPDATE tickets SET epic_id = NULL WHERE epic_id = $1`
	if _, err := r.db.Exec(ctx, clearTickets, id); err != nil {
		return false, err
	}
	const deleteEpic = `DELETE FROM epics WHERE id = $1`
	tag, err := r.db.Exec(ctx, deleteEpic, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}
