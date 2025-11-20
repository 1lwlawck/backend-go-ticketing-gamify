package gamification

import "time"

// UserStats holds aggregated metrics.
type UserStats struct {
	UserID             string    `json:"userId"`
	XPTotal            int       `json:"xpTotal"`
	Level              int       `json:"level"`
	NextLevelThreshold int       `json:"nextLevelThreshold"`
	TicketsClosed      int       `json:"ticketsClosed"`
	StreakDays         int       `json:"streakDays"`
	LastTicketClosedAt time.Time `json:"lastTicketClosedAt"`
}

// XPEvent describes XP award events.
type XPEvent struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	TicketID  string    `json:"ticketId"`
	Priority  string    `json:"priority"`
	XP        int       `json:"xp"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"createdAt"`
}

// AwardInput parameters for awarding xp.
type AwardInput struct {
	UserID   string
	TicketID string
	Priority string
	XP       int
	Note     string
}
