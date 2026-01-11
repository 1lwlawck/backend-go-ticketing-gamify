package achievements

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles achievements queries.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new achievements repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetUserProgress returns user's progress toward all achievements.
func (r *Repository) GetUserProgress(ctx context.Context, userID string) ([]Progress, error) {
	// Get user's gamification stats
	var ticketsClosed, currentStreak, totalXP, level int
	const statsQuery = `
		SELECT COALESCE(tickets_closed, 0), COALESCE(current_streak, 0), COALESCE(total_xp, 0), COALESCE(level, 1)
		FROM gamification_stats
		WHERE user_id = $1`
	err := r.db.QueryRow(ctx, statsQuery, userID).Scan(&ticketsClosed, &currentStreak, &totalXP, &level)
	if err != nil {
		// User may not have stats yet, use defaults
		ticketsClosed, currentStreak, totalXP, level = 0, 0, 0, 1
	}

	achievements := DefaultAchievements()
	var progress []Progress

	for _, a := range achievements {
		var current int
		switch a.Category {
		case "tickets":
			current = ticketsClosed
		case "streaks":
			current = currentStreak
		case "xp":
			if a.ID == "level_5" || a.ID == "level_10" {
				current = level
			} else {
				current = totalXP
			}
		}

		percentage := 0
		if a.Threshold > 0 {
			percentage = (current * 100) / a.Threshold
			if percentage > 100 {
				percentage = 100
			}
		}

		progress = append(progress, Progress{
			AchievementID: a.ID,
			Name:          a.Name,
			Description:   a.Description,
			Icon:          a.Icon,
			Current:       current,
			Target:        a.Threshold,
			Percentage:    percentage,
			Unlocked:      current >= a.Threshold,
		})
	}

	return progress, nil
}

// GetUnlockedAchievements returns achievements the user has unlocked.
func (r *Repository) GetUnlockedAchievements(ctx context.Context, userID string) ([]Progress, error) {
	all, err := r.GetUserProgress(ctx, userID)
	if err != nil {
		return nil, err
	}

	var unlocked []Progress
	for _, p := range all {
		if p.Unlocked {
			unlocked = append(unlocked, p)
		}
	}
	return unlocked, nil
}
