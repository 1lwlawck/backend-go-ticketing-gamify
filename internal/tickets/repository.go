package tickets

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository interacts with tickets table.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context, filter Filter) ([]Ticket, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}

	var (
		args []any
		idx  = 1
		sb   strings.Builder
	)
	sb.WriteString(`SELECT t.id, t.project_id, t.title, t.description, t.status, t.priority, t.type, t.reporter_id, t.assignee_id, assignee.name, t.due_date, t.created_at, t.updated_at FROM tickets t LEFT JOIN users assignee ON assignee.id = t.assignee_id WHERE 1=1`)
	if filter.ProjectID != "" {
		sb.WriteString(fmt.Sprintf(" AND project_id = $%d", idx))
		args = append(args, filter.ProjectID)
		idx++
	}
	if filter.AssigneeID != "" {
		sb.WriteString(fmt.Sprintf(" AND assignee_id = $%d", idx))
		args = append(args, filter.AssigneeID)
		idx++
	}
	if filter.Status != "" {
		sb.WriteString(fmt.Sprintf(" AND status = $%d", idx))
		args = append(args, filter.Status)
		idx++
	}
	sb.WriteString(fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d", idx))
	args = append(args, filter.Limit)

	rows, err := r.db.Query(ctx, sb.String(), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tickets []Ticket
	for rows.Next() {
		var t Ticket
		if err := rows.Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Type, &t.ReporterID, &t.AssigneeID, &t.AssigneeName, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tickets = append(tickets, t)
	}
	return tickets, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id string) (*Ticket, error) {
	const query = `
SELECT t.id, t.project_id, t.title, t.description, t.status, t.priority, t.type, t.reporter_id, t.assignee_id, assignee.name, t.due_date, t.created_at, t.updated_at
FROM tickets t
LEFT JOIN users assignee ON assignee.id = t.assignee_id
WHERE t.id = $1`
	var t Ticket
	if err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Type, &t.ReporterID, &t.AssigneeID, &t.AssigneeName, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if err := r.attachDetails(ctx, &t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) Create(ctx context.Context, input CreateInput) (*Ticket, error) {
	const query = `
INSERT INTO tickets (id, project_id, title, description, status, priority, type, reporter_id, assignee_id, due_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, 'todo', $5, $6, $7, $8, $9, $10, $10)
RETURNING id, project_id, title, description, status, priority, type, reporter_id, assignee_id, due_date, created_at, updated_at`
	now := time.Now()
	var t Ticket
	ticketID := uuid.NewString()
	if err := r.db.QueryRow(ctx, query, ticketID, input.ProjectID, input.Title, input.Description, input.Priority, input.Type, input.ReporterID, input.AssigneeID, input.DueDate, now).
		Scan(&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Type, &t.ReporterID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt); err != nil {
		return nil, err
	}
	_ = r.addHistory(ctx, t.ID, "Ticket created", nil)
	return r.Get(ctx, t.ID)
}

func (r *Repository) UpdateStatus(ctx context.Context, ticketID string, status string) (*Ticket, error) {
	const query = `
UPDATE tickets
SET status = $2, updated_at = $3
WHERE id = $1
RETURNING id, project_id, title, description, status, priority, type, reporter_id, assignee_id, due_date, created_at, updated_at`
	now := time.Now()
	var t Ticket
	if err := r.db.QueryRow(ctx, query, ticketID, status, now).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Type, &t.ReporterID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	_ = r.addHistory(ctx, t.ID, fmt.Sprintf("Status changed to %s", status), nil)
	return r.Get(ctx, t.ID)
}

func (r *Repository) UpdateFields(ctx context.Context, ticketID string, input UpdateInput) (*Ticket, error) {
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
	if input.Priority != nil {
		setParts = append(setParts, fmt.Sprintf("priority = $%d", idx))
		args = append(args, *input.Priority)
		idx++
	}
	if input.Type != nil {
		setParts = append(setParts, fmt.Sprintf("type = $%d", idx))
		args = append(args, *input.Type)
		idx++
	}
	if input.AssigneeID != nil {
		setParts = append(setParts, fmt.Sprintf("assignee_id = $%d", idx))
		args = append(args, input.AssigneeID)
		idx++
	}
	if input.DueDate != nil {
		setParts = append(setParts, fmt.Sprintf("due_date = $%d", idx))
		args = append(args, input.DueDate)
		idx++
	}

	if len(setParts) == 0 {
		return r.Get(ctx, ticketID)
	}

	setParts = append(setParts, fmt.Sprintf("updated_at = $%d", idx))
	args = append(args, time.Now())
	idx++

	query := fmt.Sprintf(`
UPDATE tickets
SET %s
WHERE id = $%d
RETURNING id, project_id, title, description, status, priority, type, reporter_id, assignee_id, due_date, created_at, updated_at`, strings.Join(setParts, ", "), idx)

	args = append(args, ticketID)

	var t Ticket
	if err := r.db.QueryRow(ctx, query, args...).Scan(
		&t.ID, &t.ProjectID, &t.Title, &t.Description, &t.Status, &t.Priority, &t.Type, &t.ReporterID, &t.AssigneeID, &t.DueDate, &t.CreatedAt, &t.UpdatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return r.Get(ctx, t.ID)
}

func (r *Repository) AddComment(ctx context.Context, ticketID, authorID, text string) (*Comment, error) {
	const query = `
WITH ins AS (
  INSERT INTO ticket_comments (id, ticket_id, author_id, text, created_at)
  VALUES ($1, $2, $3, $4, NOW())
  RETURNING id, ticket_id, author_id, text, created_at
)
SELECT ins.id, ins.ticket_id, ins.author_id, COALESCE(u.name, ''), ins.text, ins.created_at
FROM ins
LEFT JOIN users u ON u.id = ins.author_id`
	var cmt Comment
	if err := r.db.QueryRow(ctx, query, uuid.NewString(), ticketID, authorID, text).Scan(
		&cmt.ID, &cmt.TicketID, &cmt.AuthorID, &cmt.Author, &cmt.Body, &cmt.CreatedAt,
	); err != nil {
		return nil, err
	}
	return &cmt, nil
}

func (r *Repository) UpdateComment(ctx context.Context, commentID, authorID, text string) (*Comment, error) {
	const query = `
UPDATE ticket_comments
SET text = $3, created_at = created_at
WHERE id = $1 AND author_id = $2
RETURNING id, ticket_id, author_id, text, created_at`
	var cmt Comment
	if err := r.db.QueryRow(ctx, query, commentID, authorID, text).Scan(
		&cmt.ID, &cmt.TicketID, &cmt.AuthorID, &cmt.Body, &cmt.CreatedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	// fetch author name
	const nameQuery = `SELECT COALESCE(name, '') FROM users WHERE id = $1`
	_ = r.db.QueryRow(ctx, nameQuery, cmt.AuthorID).Scan(&cmt.Author)
	return &cmt, nil
}

func (r *Repository) DeleteComment(ctx context.Context, commentID, authorID string) error {
	const query = `DELETE FROM ticket_comments WHERE id = $1 AND author_id = $2`
	tag, err := r.db.Exec(ctx, query, commentID, authorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return pgx.ErrNoRows
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, ticketID string) error {
	const query = `DELETE FROM tickets WHERE id = $1`
	_, err := r.db.Exec(ctx, query, ticketID)
	return err
}

// AddProjectActivity logs a project-level activity entry (best effort).
func (r *Repository) AddProjectActivity(ctx context.Context, projectID string, actorID *string, message string) {
	const query = `
INSERT INTO project_activity (id, project_id, actor_id, message, created_at)
VALUES ($1, $2, $3, $4, NOW())`
	_, _ = r.db.Exec(ctx, query, uuid.NewString(), projectID, actorID, message)
}

func (r *Repository) addHistory(ctx context.Context, ticketID, text string, actorID *string) error {
	const query = `
INSERT INTO ticket_history (id, ticket_id, text, actor_id, timestamp)
VALUES ($1, $2, $3, $4, NOW())`
	_, err := r.db.Exec(ctx, query, uuid.NewString(), ticketID, text, actorID)
	return err
}

func (r *Repository) attachDetails(ctx context.Context, ticket *Ticket) error {
	const historyQuery = `
SELECT id, ticket_id, text, actor_id, timestamp
FROM ticket_history
WHERE ticket_id = $1
ORDER BY timestamp DESC`
	hRows, err := r.db.Query(ctx, historyQuery, ticket.ID)
	if err != nil {
		return err
	}
	defer hRows.Close()
	for hRows.Next() {
		var entry HistoryEntry
		if err := hRows.Scan(&entry.ID, &entry.TicketID, &entry.Text, &entry.ActorID, &entry.Timestamp); err != nil {
			return err
		}
		ticket.History = append(ticket.History, entry)
	}
	if err := hRows.Err(); err != nil {
		return err
	}

	const commentQuery = `
SELECT tc.id, tc.ticket_id, tc.author_id, COALESCE(u.name, ''), tc.text, tc.created_at
FROM ticket_comments tc
LEFT JOIN users u ON u.id = tc.author_id
WHERE tc.ticket_id = $1
ORDER BY tc.created_at ASC`
	cRows, err := r.db.Query(ctx, commentQuery, ticket.ID)
	if err != nil {
		return err
	}
	defer cRows.Close()
	for cRows.Next() {
		var comment Comment
		if err := cRows.Scan(&comment.ID, &comment.TicketID, &comment.AuthorID, &comment.Author, &comment.Body, &comment.CreatedAt); err != nil {
			return err
		}
		ticket.Comments = append(ticket.Comments, comment)
	}
	return cRows.Err()
}
