package activity

import "time"

// ActivityItem represents a user activity entry.
type ActivityItem struct {
	ID         string    `json:"id"`
	UserID     string    `json:"userId,omitempty"`
	UserName   string    `json:"userName,omitempty"`
	Action     string    `json:"action"`
	Details    string    `json:"details,omitempty"`
	EntityType string    `json:"entityType,omitempty"` // "ticket", "project", "epic", "comment"
	EntityID   string    `json:"entityId,omitempty"`
	CreatedAt  time.Time `json:"createdAt"`
}

// Filter for activity queries.
type Filter struct {
	UserID     string
	EntityType string
	Limit      int
	Cursor     *time.Time
}
