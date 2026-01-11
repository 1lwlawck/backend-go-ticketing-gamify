package calendar

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles calendar queries.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new calendar repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetEvents returns calendar events (deadlines, epic dates).
func (r *Repository) GetEvents(ctx context.Context, filter Filter) ([]CalendarEvent, error) {
	var events []CalendarEvent

	// Default date range: current month
	now := time.Now()
	startDate := filter.StartDate
	endDate := filter.EndDate
	if startDate == nil {
		firstOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
		startDate = &firstOfMonth
	}
	if endDate == nil {
		lastOfMonth := startDate.AddDate(0, 1, -1)
		endDate = &lastOfMonth
	}

	// Get ticket deadlines
	if filter.Type == "" || filter.Type == "all" || filter.Type == "ticket" {
		ticketQuery := `
			SELECT t.id, t.title, 'ticket_deadline', t.due_date, t.project_id, p.name, t.status, t.priority
			FROM tickets t
			JOIN projects p ON p.id = t.project_id
			WHERE t.due_date IS NOT NULL
			  AND t.due_date >= $1
			  AND t.due_date <= $2`
		args := []any{startDate, endDate}

		if filter.ProjectID != "" {
			ticketQuery += ` AND t.project_id = $3`
			args = append(args, filter.ProjectID)
		}
		ticketQuery += ` ORDER BY t.due_date ASC`

		rows, err := r.db.Query(ctx, ticketQuery, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var e CalendarEvent
			if err := rows.Scan(&e.ID, &e.Title, &e.Type, &e.Date, &e.ProjectID, &e.ProjectName, &e.Status, &e.Priority); err != nil {
				return nil, err
			}
			events = append(events, e)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}
	}

	// Get epic start dates
	if filter.Type == "" || filter.Type == "all" || filter.Type == "epic" {
		epicStartQuery := `
			SELECT e.id, e.title, 'epic_start', e.start_date, e.project_id, p.name, e.status, ''
			FROM epics e
			JOIN projects p ON p.id = e.project_id
			WHERE e.start_date IS NOT NULL
			  AND e.start_date >= $1
			  AND e.start_date <= $2`
		args := []any{startDate, endDate}

		if filter.ProjectID != "" {
			epicStartQuery += ` AND e.project_id = $3`
			args = append(args, filter.ProjectID)
		}

		rows, err := r.db.Query(ctx, epicStartQuery, args...)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		for rows.Next() {
			var e CalendarEvent
			if err := rows.Scan(&e.ID, &e.Title, &e.Type, &e.Date, &e.ProjectID, &e.ProjectName, &e.Status, &e.Priority); err != nil {
				return nil, err
			}
			events = append(events, e)
		}
		if err := rows.Err(); err != nil {
			return nil, err
		}

		// Get epic end dates
		epicEndQuery := `
			SELECT e.id, e.title, 'epic_end', e.due_date, e.project_id, p.name, e.status, ''
			FROM epics e
			JOIN projects p ON p.id = e.project_id
			WHERE e.due_date IS NOT NULL
			  AND e.due_date >= $1
			  AND e.due_date <= $2`
		args = []any{startDate, endDate}

		if filter.ProjectID != "" {
			epicEndQuery += ` AND e.project_id = $3`
			args = append(args, filter.ProjectID)
		}

		rows2, err := r.db.Query(ctx, epicEndQuery, args...)
		if err != nil {
			return nil, err
		}
		defer rows2.Close()

		for rows2.Next() {
			var e CalendarEvent
			if err := rows2.Scan(&e.ID, &e.Title, &e.Type, &e.Date, &e.ProjectID, &e.ProjectName, &e.Status, &e.Priority); err != nil {
				return nil, err
			}
			events = append(events, e)
		}
		if err := rows2.Err(); err != nil {
			return nil, err
		}
	}

	return events, nil
}
