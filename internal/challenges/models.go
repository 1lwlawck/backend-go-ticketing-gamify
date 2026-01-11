package challenges

import "time"

// Challenge represents a weekly challenge.
type Challenge struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Type        string    `json:"type"` // "tickets", "xp", "streak", "comments"
	Target      int       `json:"target"`
	XPReward    int       `json:"xpReward"`
	StartDate   time.Time `json:"startDate"`
	EndDate     time.Time `json:"endDate"`
	Active      bool      `json:"active"`
}

// UserChallenge represents a user's progress on a challenge.
type UserChallenge struct {
	ChallengeID string     `json:"challengeId"`
	Challenge   Challenge  `json:"challenge"`
	UserID      string     `json:"userId"`
	Current     int        `json:"current"`
	Completed   bool       `json:"completed"`
	Percentage  int        `json:"percentage"`
	CompletedAt *time.Time `json:"completedAt,omitempty"`
}

// GetWeeklyChallenges returns the current week's challenges.
func GetWeeklyChallenges() []Challenge {
	now := time.Now()
	// Calculate start of current week (Monday)
	weekday := int(now.Weekday())
	if weekday == 0 {
		weekday = 7
	}
	startOfWeek := now.AddDate(0, 0, -weekday+1)
	startOfWeek = time.Date(startOfWeek.Year(), startOfWeek.Month(), startOfWeek.Day(), 0, 0, 0, 0, now.Location())
	endOfWeek := startOfWeek.AddDate(0, 0, 6)
	endOfWeek = time.Date(endOfWeek.Year(), endOfWeek.Month(), endOfWeek.Day(), 23, 59, 59, 0, now.Location())

	return []Challenge{
		{
			ID:          "weekly_tickets_5",
			Title:       "Ticket Sprint",
			Description: "Close 5 tickets this week",
			Type:        "tickets",
			Target:      5,
			XPReward:    100,
			StartDate:   startOfWeek,
			EndDate:     endOfWeek,
			Active:      true,
		},
		{
			ID:          "weekly_xp_200",
			Title:       "XP Rush",
			Description: "Earn 200 XP this week",
			Type:        "xp",
			Target:      200,
			XPReward:    50,
			StartDate:   startOfWeek,
			EndDate:     endOfWeek,
			Active:      true,
		},
		{
			ID:          "weekly_streak_3",
			Title:       "Streak Builder",
			Description: "Maintain a 3-day streak this week",
			Type:        "streak",
			Target:      3,
			XPReward:    75,
			StartDate:   startOfWeek,
			EndDate:     endOfWeek,
			Active:      true,
		},
		{
			ID:          "weekly_comments_10",
			Title:       "Team Player",
			Description: "Post 10 comments this week",
			Type:        "comments",
			Target:      10,
			XPReward:    60,
			StartDate:   startOfWeek,
			EndDate:     endOfWeek,
			Active:      true,
		},
	}
}
