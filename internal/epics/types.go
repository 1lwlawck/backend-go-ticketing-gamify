package epics

import "time"

type Epic struct {
	ID          string     `json:"id"`
	ProjectID   string     `json:"projectId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"startDate,omitempty"`
	DueDate     *time.Time `json:"dueDate,omitempty"`
	OwnerID     *string    `json:"ownerId,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
	DoneCount   int        `json:"doneCount"`
	TotalCount  int        `json:"totalCount"`
}

// Filter for listing epics by project.
type Filter struct {
	ProjectID string
	Status    string
	Search    string
	Cursor    *time.Time // created_at cursor (created_at < cursor)
	Limit     int
}

type CreateInput struct {
	ProjectID   string     `json:"projectId"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	StartDate   *time.Time `json:"startDate"`
	DueDate     *time.Time `json:"dueDate"`
	OwnerID     *string    `json:"ownerId"`
}

type UpdateInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	StartDate   *time.Time `json:"startDate"`
	DueDate     *time.Time `json:"dueDate"`
	OwnerID     *string    `json:"ownerId"`
	ClearStart  bool       `json:"clearStartDate"`
	ClearDue    bool       `json:"clearDueDate"`
}
