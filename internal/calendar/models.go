package calendar

import "time"

// CalendarEvent represents a deadline or scheduled event.
type CalendarEvent struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Type        string    `json:"type"` // "ticket_deadline", "epic_start", "epic_end"
	Date        time.Time `json:"date"`
	ProjectID   string    `json:"projectId,omitempty"`
	ProjectName string    `json:"projectName,omitempty"`
	Status      string    `json:"status,omitempty"`
	Priority    string    `json:"priority,omitempty"`
}

// Filter for calendar events.
type Filter struct {
	StartDate *time.Time
	EndDate   *time.Time
	ProjectID string
	Type      string // "ticket", "epic", "all"
}
