package team

import "time"

// Member represents a team member with project assignments.
type Member struct {
	ID           string    `json:"id"`
	Name         string    `json:"name"`
	Username     string    `json:"username"`
	Role         string    `json:"role"`
	AvatarURL    *string   `json:"avatarUrl,omitempty"`
	ProjectCount int       `json:"projectCount"`
	TicketCount  int       `json:"ticketCount"`
	TotalXP      int       `json:"totalXp"`
	Level        int       `json:"level"`
	JoinedAt     time.Time `json:"joinedAt"`
}

// ProjectMember represents a member's role in a project.
type ProjectMember struct {
	UserID      string `json:"userId"`
	UserName    string `json:"userName"`
	ProjectID   string `json:"projectId"`
	ProjectName string `json:"projectName"`
	Role        string `json:"role"`
}
