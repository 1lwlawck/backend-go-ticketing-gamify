package reports

import "time"

// Summary contains overall dashboard metrics.
type Summary struct {
	TotalTickets   int `json:"totalTickets"`
	OpenTickets    int `json:"openTickets"`
	ClosedTickets  int `json:"closedTickets"`
	TotalProjects  int `json:"totalProjects"`
	ActiveProjects int `json:"activeProjects"`
	TotalUsers     int `json:"totalUsers"`
	TotalEpics     int `json:"totalEpics"`
}

// StatusBreakdown shows ticket counts per status.
type StatusBreakdown struct {
	Status string `json:"status"`
	Count  int    `json:"count"`
}

// AssigneeBreakdown shows ticket counts per assignee.
type AssigneeBreakdown struct {
	UserID      string `json:"userId"`
	UserName    string `json:"userName"`
	TicketCount int    `json:"ticketCount"`
	ClosedCount int    `json:"closedCount"`
	OpenCount   int    `json:"openCount"`
}

// TeamPerformance shows performance metrics.
type TeamPerformance struct {
	UserID        string     `json:"userId"`
	UserName      string     `json:"userName"`
	TotalXP       int        `json:"totalXp"`
	Level         int        `json:"level"`
	TicketsClosed int        `json:"ticketsClosed"`
	CurrentStreak int        `json:"currentStreak"`
	LastActiveAt  *time.Time `json:"lastActiveAt,omitempty"`
}

// PriorityBreakdown shows ticket counts per priority.
type PriorityBreakdown struct {
	Priority string `json:"priority"`
	Count    int    `json:"count"`
}

// TicketTrend shows tickets created/closed over time.
type TicketTrend struct {
	Date    string `json:"date"`
	Created int    `json:"created"`
	Closed  int    `json:"closed"`
}
