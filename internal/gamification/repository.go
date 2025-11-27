package gamification

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles DB operations.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetStats(ctx context.Context, userID string) (*UserStats, error) {
	const query = `
SELECT user_id, xp_total, level, next_level_threshold, tickets_closed_count, streak_days, COALESCE(last_ticket_closed_at, NOW()) 
FROM gamification_user_stats
WHERE user_id = $1`
	var stats UserStats
	if err := r.db.QueryRow(ctx, query, userID).Scan(
		&stats.UserID,
		&stats.XPTotal,
		&stats.Level,
		&stats.NextLevelThreshold,
		&stats.TicketsClosed,
		&stats.StreakDays,
		&stats.LastTicketClosedAt,
	); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &stats, nil
}

func (r *Repository) ListEvents(ctx context.Context, userID string, limit int, cursor *time.Time) ([]XPEvent, *string, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var (
		args []any
		idx  = 1
		sb   strings.Builder
	)
	sb.WriteString(`SELECT id, user_id, ticket_id, priority, xp_value, note, created_at
FROM xp_events
WHERE 1=1`)
	if userID != "" {
		sb.WriteString(fmt.Sprintf(" AND user_id = $%d::uuid", idx))
		args = append(args, userID)
		idx++
	}
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

	var events []XPEvent
	for rows.Next() {
		var e XPEvent
		if err := rows.Scan(&e.ID, &e.UserID, &e.TicketID, &e.Priority, &e.XP, &e.Note, &e.CreatedAt); err != nil {
			return nil, nil, err
		}
		events = append(events, e)
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	if len(events) > limit {
		last := events[limit-1]
		cursorStr := last.CreatedAt.Format(time.RFC3339Nano)
		events = events[:limit]
		return events, &cursorStr, nil
	}
	return events, nil, nil
}

func (r *Repository) Award(ctx context.Context, input AwardInput) error {
	return r.Adjust(ctx, AdjustInput{
		UserID:      input.UserID,
		TicketID:    input.TicketID,
		Priority:    input.Priority,
		XP:          input.XP,
		Note:        input.Note,
		ClosedDelta: 1,
	})
}

// Adjust can add or subtract XP and adjust closed ticket count.
func (r *Repository) Adjust(ctx context.Context, input AdjustInput) error {
	if input.UserID == "" || input.XP == 0 {
		return nil
	}
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	const insertEvent = `
INSERT INTO xp_events (id, user_id, ticket_id, priority, xp_value, note)
VALUES ($1, $2, $3, $4, $5, $6)`
	if _, err := tx.Exec(ctx, insertEvent, uuid.NewString(), input.UserID, input.TicketID, input.Priority, input.XP, input.Note); err != nil {
		return err
	}

	const upsertStats = `
INSERT INTO gamification_user_stats (user_id, xp_total, level, next_level_threshold, tickets_closed_count, streak_days, last_ticket_closed_at)
VALUES ($1, $2, 1, 100, $3, 1, NOW())
ON CONFLICT (user_id) DO UPDATE
SET xp_total = GREATEST(gamification_user_stats.xp_total + EXCLUDED.xp_total, 0),
    tickets_closed_count = GREATEST(gamification_user_stats.tickets_closed_count + EXCLUDED.tickets_closed_count, 0),
    level = FLOOR(GREATEST(gamification_user_stats.xp_total + EXCLUDED.xp_total, 0)/100)::int + 1,
    next_level_threshold = (FLOOR(GREATEST(gamification_user_stats.xp_total + EXCLUDED.xp_total, 0)/100)::int + 1) * 100,
    streak_days = CASE
        WHEN EXCLUDED.tickets_closed_count > 0 THEN
            CASE
                WHEN gamification_user_stats.last_ticket_closed_at >= CURRENT_DATE THEN gamification_user_stats.streak_days
                WHEN gamification_user_stats.last_ticket_closed_at = CURRENT_DATE - INTERVAL '1 day' THEN gamification_user_stats.streak_days + 1
                ELSE 1
            END
        ELSE gamification_user_stats.streak_days
    END,
    last_ticket_closed_at = CASE WHEN EXCLUDED.tickets_closed_count > 0 THEN NOW() ELSE gamification_user_stats.last_ticket_closed_at END`
	if _, err := tx.Exec(ctx, upsertStats, input.UserID, input.XP, input.ClosedDelta); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) EnsureUser(ctx context.Context, userID string) error {
	const query = `
INSERT INTO gamification_user_stats (user_id, xp_total, level, next_level_threshold, tickets_closed_count, streak_days, last_ticket_closed_at)
VALUES ($1, 0, 1, 100, 0, 0, NULL)
ON CONFLICT (user_id) DO NOTHING`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *Repository) Leaderboard(ctx context.Context, limit int, cursor int) ([]LeaderboardRow, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	if cursor < 0 {
		cursor = 0
	}
	const query = `
SELECT u.id,
       u.name,
       u.username,
       u.role,
       COALESCE(g.xp_total, 0)   AS xp,
       COALESCE(g.level, 1)      AS level,
       COALESCE(g.tickets_closed_count, 0) AS tickets_closed_count
FROM users u
LEFT JOIN gamification_user_stats g ON g.user_id = u.id
ORDER BY COALESCE(g.xp_total, 0) DESC, COALESCE(g.level, 1) DESC
LIMIT $1 OFFSET $2`
	rows, err := r.db.Query(ctx, query, limit, cursor)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rowsOut []LeaderboardRow
	position := 1
	var leaderXP int

	for rows.Next() {
		var row LeaderboardRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Username, &row.Role, &row.XP, &row.Level, &row.TicketsClosedCount); err != nil {
			return nil, err
		}
		row.Rank = position
		if position == 1 {
			leaderXP = row.XP
			row.XPGap = 0
		} else {
			row.XPGap = leaderXP - row.XP
			if row.XPGap < 0 {
				row.XPGap = 0
			}
		}
		rowsOut = append(rowsOut, row)
		position++
	}
	return rowsOut, rows.Err()
}

// RefreshClosedCount recalculates tickets_closed_count from tickets table for a user.
func (r *Repository) RefreshClosedCount(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	const query = `
WITH stats AS (
  SELECT 
    COUNT(*)::int AS closed_count,
    MAX(updated_at) AS last_closed_at
  FROM tickets
  WHERE assignee_id = $1::uuid AND status = 'done'
)
INSERT INTO gamification_user_stats (user_id, xp_total, level, next_level_threshold, tickets_closed_count, streak_days, last_ticket_closed_at)
VALUES ($1, 0, 1, 100, (SELECT closed_count FROM stats), 0, (SELECT last_closed_at FROM stats))
ON CONFLICT (user_id) DO UPDATE
SET tickets_closed_count = GREATEST((SELECT closed_count FROM stats), 0),
    last_ticket_closed_at = COALESCE((SELECT last_closed_at FROM stats), gamification_user_stats.last_ticket_closed_at)`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// RefreshAllClosedCounts recalculates closed ticket counts for all assignees.
func (r *Repository) RefreshAllClosedCounts(ctx context.Context) error {
	const query = `
WITH counts AS (
  SELECT assignee_id AS user_id,
         COUNT(*)::int AS closed_count,
         MAX(updated_at) AS last_closed_at
  FROM tickets
  WHERE assignee_id IS NOT NULL AND status = 'done'
  GROUP BY assignee_id
)
INSERT INTO gamification_user_stats (user_id, xp_total, level, next_level_threshold, tickets_closed_count, streak_days, last_ticket_closed_at)
SELECT user_id, 0, 1, 100, closed_count, 0, last_closed_at
FROM counts
ON CONFLICT (user_id) DO UPDATE
SET tickets_closed_count = EXCLUDED.tickets_closed_count,
    last_ticket_closed_at = COALESCE(EXCLUDED.last_ticket_closed_at, gamification_user_stats.last_ticket_closed_at);`
	_, err := r.db.Exec(ctx, query)
	return err
}
