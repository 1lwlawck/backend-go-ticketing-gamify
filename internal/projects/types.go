package projects

import "time"

// Project model returned in listing.
type Project struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description"`
	Status       string    `json:"status"`
	TicketsCount int       `json:"ticketsCount"`
	CreatedAt    time.Time `json:"createdAt"`
}

// Member describes a project member.
type Member struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// Detail extends project with members, invites, and activity feed.
type Detail struct {
	Project
	Members  []Member   `json:"members"`
	Invites  []Invite   `json:"invites"`
	Activity []Activity `json:"activity"`
}

// ListFilter supports searching and pagination.
type ListFilter struct {
	Limit   int
	Search  string
	Status  string
	Cursor  *time.Time // created_at cursor (created_at < cursor)
}

// CreateInput payload for project creation.
type CreateInput struct {
	Name        string   `json:"name" binding:"required"`
	Description string   `json:"description" binding:"required"`
	Members     []string `json:"members"`
}

// Invite represents a project invite.
type Invite struct {
	Code      string    `json:"code"`
	MaxUses   int       `json:"maxUses"`
	Uses      int       `json:"uses"`
	ExpiresAt time.Time `json:"expiresAt"`
}

// Activity represents a project activity entry.
type Activity struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

// AddMemberInput request body.
type AddMemberInput struct {
	UserID string `json:"userId" binding:"required"`
	Role   string `json:"role"`
}

// InviteInput configuration for generating invite.
type InviteInput struct {
	MaxUses    int `json:"maxUses" binding:"required"`
	ExpiryDays int `json:"expiryDays" binding:"required"`
}
