package challenges

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles challenges queries.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new challenges repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetUserChallengeProgress returns user's progress on current week's challenges.
func (r *Repository) GetUserChallengeProgress(ctx context.Context, userID string) ([]UserChallenge, error) {
	challenges := GetWeeklyChallenges()
	now := time.Now()

	// Calculate start of current week
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, -weekday+1)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, now.Location())

	var result []UserChallenge

	for _, c := range challenges {
		var current int

		switch c.Type {
		case "tickets":
			// Count tickets closed this week
			const query = `
				SELECT COUNT(*) FROM tickets 
				WHERE assignee_id = $1 
				  AND status = 'done' 
				  AND updated_at >= $2`
			_ = r.db.QueryRow(ctx, query, userID, startOfWeek).Scan(&current)

		case "xp":
			// Count XP earned this week from xp_events
			const query = `
				SELECT COALESCE(SUM(xp), 0) FROM xp_events 
				WHERE user_id = $1 
				  AND created_at >= $2`
			_ = r.db.QueryRow(ctx, query, userID, startOfWeek).Scan(&current)

		case "streak":
			// Get current streak
			const query = `
				SELECT COALESCE(current_streak, 0) FROM gamification_stats 
				WHERE user_id = $1`
			_ = r.db.QueryRow(ctx, query, userID).Scan(&current)

		case "comments":
			// Count comments posted this week
			const query = `
				SELECT COUNT(*) FROM ticket_comments 
				WHERE author_id = $1 
				  AND created_at >= $2`
			_ = r.db.QueryRow(ctx, query, userID, startOfWeek).Scan(&current)
		}

		percentage := 0
		if c.Target > 0 {
			percentage = (current * 100) / c.Target
			if percentage > 100 {
				percentage = 100
			}
		}

		result = append(result, UserChallenge{
			ChallengeID: c.ID,
			Challenge:   c,
			UserID:      userID,
			Current:     current,
			Completed:   current >= c.Target,
			Percentage:  percentage,
		})
	}

	return result, nil
}
