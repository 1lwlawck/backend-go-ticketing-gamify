package reports

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles report queries.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new reports repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetSummary returns overall dashboard metrics.
func (r *Repository) GetSummary(ctx context.Context) (*Summary, error) {
	var s Summary

	// Total tickets
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tickets`).Scan(&s.TotalTickets)
	if err != nil {
		return nil, err
	}

	// Open vs Closed tickets
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM tickets WHERE status != 'done'`).Scan(&s.OpenTickets)
	if err != nil {
		return nil, err
	}
	s.ClosedTickets = s.TotalTickets - s.OpenTickets

	// Projects
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM projects`).Scan(&s.TotalProjects)
	if err != nil {
		return nil, err
	}
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM projects WHERE status = 'Active'`).Scan(&s.ActiveProjects)
	if err != nil {
		return nil, err
	}

	// Users
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&s.TotalUsers)
	if err != nil {
		return nil, err
	}

	// Epics
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM epics`).Scan(&s.TotalEpics)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// GetStatusBreakdown returns ticket count per status.
func (r *Repository) GetStatusBreakdown(ctx context.Context) ([]StatusBreakdown, error) {
	const query = `
		SELECT status, COUNT(*) as count
		FROM tickets
		GROUP BY status
		ORDER BY count DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []StatusBreakdown
	for rows.Next() {
		var sb StatusBreakdown
		if err := rows.Scan(&sb.Status, &sb.Count); err != nil {
			return nil, err
		}
		result = append(result, sb)
	}
	return result, rows.Err()
}

// GetPriorityBreakdown returns ticket count per priority.
func (r *Repository) GetPriorityBreakdown(ctx context.Context) ([]PriorityBreakdown, error) {
	const query = `
		SELECT priority, COUNT(*) as count
		FROM tickets
		GROUP BY priority
		ORDER BY count DESC`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PriorityBreakdown
	for rows.Next() {
		var pb PriorityBreakdown
		if err := rows.Scan(&pb.Priority, &pb.Count); err != nil {
			return nil, err
		}
		result = append(result, pb)
	}
	return result, rows.Err()
}

// GetAssigneeBreakdown returns ticket count per assignee.
func (r *Repository) GetAssigneeBreakdown(ctx context.Context, limit int) ([]AssigneeBreakdown, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	const query = `
		SELECT 
			u.id, 
			u.name,
			COUNT(t.id) as ticket_count,
			COUNT(t.id) FILTER (WHERE t.status = 'done') as closed_count,
			COUNT(t.id) FILTER (WHERE t.status != 'done') as open_count
		FROM users u
		LEFT JOIN tickets t ON t.assignee_id = u.id
		GROUP BY u.id, u.name
		HAVING COUNT(t.id) > 0
		ORDER BY ticket_count DESC
		LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AssigneeBreakdown
	for rows.Next() {
		var ab AssigneeBreakdown
		if err := rows.Scan(&ab.UserID, &ab.UserName, &ab.TicketCount, &ab.ClosedCount, &ab.OpenCount); err != nil {
			return nil, err
		}
		result = append(result, ab)
	}
	return result, rows.Err()
}

// GetTeamPerformance returns team performance metrics.
func (r *Repository) GetTeamPerformance(ctx context.Context, limit int) ([]TeamPerformance, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	const query = `
		SELECT 
			u.id,
			u.name,
			COALESCE(gs.xp_total, 0) as total_xp,
			COALESCE(gs.level, 1) as level,
			COALESCE(gs.tickets_closed_count, 0) as tickets_closed,
			COALESCE(gs.streak_days, 0) as current_streak,
			gs.last_ticket_closed_at
		FROM users u
		LEFT JOIN gamification_user_stats gs ON gs.user_id = u.id
		ORDER BY total_xp DESC
		LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TeamPerformance
	for rows.Next() {
		var tp TeamPerformance
		if err := rows.Scan(&tp.UserID, &tp.UserName, &tp.TotalXP, &tp.Level, &tp.TicketsClosed, &tp.CurrentStreak, &tp.LastActiveAt); err != nil {
			return nil, err
		}
		result = append(result, tp)
	}
	return result, rows.Err()
}

// GetTicketTrend returns ticket creation/closure trend for last N days.
func (r *Repository) GetTicketTrend(ctx context.Context, days int) ([]TicketTrend, error) {
	if days <= 0 || days > 90 {
		days = 30
	}

	const query = `
		WITH date_series AS (
			SELECT (CURRENT_DATE - i) as date
			FROM generate_series($1::integer - 1, 0, -1) as i
		)
		SELECT 
			ds.date::text,
			COALESCE(COUNT(t.id) FILTER (WHERE t.created_at::date = ds.date), 0) as created,
			COALESCE(COUNT(t.id) FILTER (WHERE t.status = 'done' AND t.updated_at::date = ds.date), 0) as closed
		FROM date_series ds
		LEFT JOIN tickets t ON t.created_at::date = ds.date OR (t.status = 'done' AND t.updated_at::date = ds.date)
		GROUP BY ds.date
		ORDER BY ds.date`

	rows, err := r.db.Query(ctx, query, days)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []TicketTrend
	for rows.Next() {
		var tt TicketTrend
		if err := rows.Scan(&tt.Date, &tt.Created, &tt.Closed); err != nil {
			return nil, err
		}
		result = append(result, tt)
	}
	return result, rows.Err()
}
