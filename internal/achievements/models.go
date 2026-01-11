package achievements

import "time"

// Achievement represents a badge or achievement.
type Achievement struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Icon        string `json:"icon"`
	Category    string `json:"category"` // "tickets", "streaks", "xp", "community"
	Threshold   int    `json:"threshold"`
	XPReward    int    `json:"xpReward"`
}

// UserAchievement represents an achievement unlocked by a user.
type UserAchievement struct {
	ID            string      `json:"id"`
	UserID        string      `json:"userId"`
	AchievementID string      `json:"achievementId"`
	Achievement   Achievement `json:"achievement"`
	UnlockedAt    time.Time   `json:"unlockedAt"`
}

// Progress represents user's progress toward an achievement.
type Progress struct {
	AchievementID string `json:"achievementId"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Icon          string `json:"icon"`
	Current       int    `json:"current"`
	Target        int    `json:"target"`
	Percentage    int    `json:"percentage"`
	Unlocked      bool   `json:"unlocked"`
}

// DefaultAchievements returns the list of predefined achievements.
func DefaultAchievements() []Achievement {
	return []Achievement{
		{ID: "first_ticket", Name: "First Blood", Description: "Close your first ticket", Icon: "🎯", Category: "tickets", Threshold: 1, XPReward: 50},
		{ID: "tickets_10", Name: "Getting Started", Description: "Close 10 tickets", Icon: "🚀", Category: "tickets", Threshold: 10, XPReward: 100},
		{ID: "tickets_50", Name: "Ticket Master", Description: "Close 50 tickets", Icon: "🏆", Category: "tickets", Threshold: 50, XPReward: 300},
		{ID: "tickets_100", Name: "Ticket Legend", Description: "Close 100 tickets", Icon: "👑", Category: "tickets", Threshold: 100, XPReward: 500},
		{ID: "streak_7", Name: "Week Warrior", Description: "Maintain a 7-day streak", Icon: "🔥", Category: "streaks", Threshold: 7, XPReward: 150},
		{ID: "streak_30", Name: "Monthly Master", Description: "Maintain a 30-day streak", Icon: "⚡", Category: "streaks", Threshold: 30, XPReward: 500},
		{ID: "xp_1000", Name: "XP Hunter", Description: "Earn 1000 XP", Icon: "💎", Category: "xp", Threshold: 1000, XPReward: 100},
		{ID: "xp_5000", Name: "XP Champion", Description: "Earn 5000 XP", Icon: "🌟", Category: "xp", Threshold: 5000, XPReward: 300},
		{ID: "level_5", Name: "Rising Star", Description: "Reach level 5", Icon: "⭐", Category: "xp", Threshold: 5, XPReward: 200},
		{ID: "level_10", Name: "Elite Operator", Description: "Reach level 10", Icon: "🌠", Category: "xp", Threshold: 10, XPReward: 400},
	}
}
