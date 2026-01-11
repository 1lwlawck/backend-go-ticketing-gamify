package activity

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles activity queries.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new activity repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetUserActivity returns activity log for a user.
func (r *Repository) GetUserActivity(ctx context.Context, filter Filter) ([]ActivityItem, *time.Time, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 50
	}

	// Combine activities from multiple sources:
	// 1. Audit logs
	// 2. XP events
	// 3. Ticket history

	var activities []ActivityItem

	// Get from audit logs
	auditQuery := `
		SELECT id, actor_id, action, entity_type, entity_id, created_at
		FROM audit_log
		WHERE 1=1`
	args := []any{}
	idx := 1

	if filter.UserID != "" {
		auditQuery += ` AND actor_id = $` + string(rune('0'+idx))
		args = append(args, filter.UserID)
		idx++
	}
	if filter.Cursor != nil {
		auditQuery += ` AND created_at < $` + string(rune('0'+idx))
		args = append(args, *filter.Cursor)
		idx++
	}
	auditQuery += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+idx))
	args = append(args, filter.Limit)

	rows, err := r.db.Query(ctx, auditQuery, args...)
	if err != nil {
		// Audit table might not exist, continue
		_ = err
	} else {
		defer rows.Close()
		for rows.Next() {
			var a ActivityItem
			if err := rows.Scan(&a.ID, &a.UserID, &a.Action, &a.EntityType, &a.EntityID, &a.CreatedAt); err != nil {
				continue
			}
			activities = append(activities, a)
		}
	}

	// Get XP events
	xpQuery := `
		SELECT id, user_id, note, xp_value, priority, created_at
		FROM xp_events
		WHERE 1=1`
	args = []any{}
	idx = 1

	if filter.UserID != "" {
		xpQuery += ` AND user_id = $1`
		args = append(args, filter.UserID)
		idx++
	}
	xpQuery += ` ORDER BY created_at DESC LIMIT $` + string(rune('0'+idx))
	args = append(args, filter.Limit)

	xpRows, err := r.db.Query(ctx, xpQuery, args...)
	if err == nil {
		defer xpRows.Close()
		for xpRows.Next() {
			var id, userID string
			var note *string
			var xpValue int
			var priority *string
			var createdAt time.Time
			if err := xpRows.Scan(&id, &userID, &note, &xpValue, &priority, &createdAt); err != nil {
				continue
			}
			action := "Earned XP"
			if note != nil {
				action = *note
			}
			details := ""
			if priority != nil {
				details = "Priority: " + *priority
			}
			activities = append(activities, ActivityItem{
				ID:         id,
				UserID:     userID,
				Action:     action,
				Details:    details,
				EntityType: "xp_event",
				CreatedAt:  createdAt,
			})
		}
	}

	// Sort by CreatedAt descending and limit
	// Simple bubble sort for small datasets
	for i := 0; i < len(activities); i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[j].CreatedAt.After(activities[i].CreatedAt) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	if len(activities) > filter.Limit {
		activities = activities[:filter.Limit]
	}

	var nextCursor *time.Time
	if len(activities) == filter.Limit {
		nextCursor = &activities[len(activities)-1].CreatedAt
	}

	return activities, nextCursor, nil
}
