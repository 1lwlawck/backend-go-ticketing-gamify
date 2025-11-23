package tickets

import "time"

// Ticket base model.
type Ticket struct {
	ID           string         `json:"id"`
	ProjectID    string         `json:"projectId"`
	Title        string         `json:"title"`
	Description  string         `json:"description"`
	Status       string         `json:"status"`
	Priority     string         `json:"priority"`
	Type         string         `json:"type"`
	ReporterID   string         `json:"reporterId"`
	AssigneeID   *string        `json:"assigneeId,omitempty"`
	AssigneeName *string        `json:"assigneeName,omitempty"`
	DueDate      *time.Time     `json:"dueDate,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
	UpdatedAt    time.Time      `json:"updatedAt"`
	History      []HistoryEntry `json:"history"`
	Comments     []Comment      `json:"comments"`
}

// Filter query params for listing.
type Filter struct {
	ProjectID  string
	AssigneeID string
	Status     string
	Limit      int
}

// CreateInput payload for new ticket.
type CreateInput struct {
	ProjectID   string     `json:"projectId" binding:"required"`
	Title       string     `json:"title" binding:"required"`
	Description string     `json:"description" binding:"required"`
	Priority    string     `json:"priority" binding:"required"`
	Type        string     `json:"type" binding:"required"`
	ReporterID  string     `json:"reporterId"`
	AssigneeID  *string    `json:"assigneeId"`
	DueDate     *time.Time `json:"dueDate"`
}

// UpdateStatusInput change status payload.
type UpdateStatusInput struct {
	Status string `json:"status" binding:"required"`
}

// UpdateInput captures editable ticket fields.
type UpdateInput struct {
	Title       *string    `json:"title"`
	Description *string    `json:"description"`
	Priority    *string    `json:"priority"`
	Type        *string    `json:"type"`
	AssigneeID  *string    `json:"assigneeId"`
	DueDate     *time.Time `json:"dueDate"`
}

// Comment represents ticket comment.
type Comment struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticketId"`
	AuthorID  string    `json:"authorId"`
	Author    string    `json:"author"`
	Body      string    `json:"text"`
	CreatedAt time.Time `json:"timestamp"`
}

// CommentUpdate represents update body.
type CommentUpdate struct {
	Text string `json:"text" binding:"required"`
}

// HistoryEntry represents audit trail.
type HistoryEntry struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticketId"`
	Text      string    `json:"text"`
	ActorID   *string   `json:"actorId,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// CommentInput request payload.
type CommentInput struct {
	Text string `json:"text" binding:"required"`
}
