package tickets

import (
	"context"
	"errors"
	"fmt"

	"backend-go-ticketing-gamify/internal/audit"
	"backend-go-ticketing-gamify/internal/gamification"
	"backend-go-ticketing-gamify/internal/middleware"
	"github.com/jackc/pgx/v5"
)

var (
	priorityXP = map[string]int{
		"low":    5,
		"medium": 10,
		"high":   20,
		"urgent": 30,
	}
	// ErrForbidden is returned when the user has no permission to mutate ticket.
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("not_found")
	// ErrEpicProjectMismatch when epic does not belong to the ticket's project.
	ErrEpicProjectMismatch = errors.New("epic_project_mismatch")
)

func canModify(actor *middleware.UserContext, ticket *Ticket) bool {
	if actor == nil || ticket == nil {
		return false
	}
	switch actor.Role {
	case "admin", "project_manager":
		return true
	}
	if ticket.AssigneeID != nil && *ticket.AssigneeID == actor.ID {
		return true
	}
	if ticket.ReporterID == actor.ID {
		return true
	}
	return false
}

// Service coordinates workflows.
type Service struct {
	repo         *Repository
	audit        *audit.Service
	gamification *gamification.Service
}

func NewService(repo *Repository, audit *audit.Service, gamification *gamification.Service) *Service {
	return &Service{repo: repo, audit: audit, gamification: gamification}
}

func formatStatusLabel(status string) string {
	switch status {
	case "backlog":
		return "Backlog"
	case "todo":
		return "Todo"
	case "in_progress":
		return "In progress"
	case "review":
		return "Review"
	case "done":
		return "Selesai"
	default:
		return status
	}
}

func (s *Service) List(ctx context.Context, filter Filter) ([]Ticket, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id string) (*Ticket, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, actor *middleware.UserContext, input CreateInput) (*Ticket, error) {
	if input.EpicID != nil && *input.EpicID != "" {
		ok, err := s.repo.EpicBelongsToProject(ctx, *input.EpicID, input.ProjectID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrEpicProjectMismatch
		}
	}
	ticket, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	desc := fmt.Sprintf("%s created ticket %s", actor.Name, ticket.Title)
	actorID := actor.ID
	entityType := "ticket"
	entityID := ticket.ID
	_ = s.audit.Log(ctx, "ticket_created", desc, &actorID, &entityType, &entityID)
	activity := fmt.Sprintf("%s membuat tiket %s", actor.Name, ticket.Title)
	s.repo.AddProjectActivity(ctx, ticket.ProjectID, &actorID, activity)
	return ticket, nil
}

func (s *Service) UpdateStatus(ctx context.Context, actor *middleware.UserContext, ticketID, status string) (*Ticket, error) {
	current, err := s.repo.Get(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}
	wasDone := current.Status == "done"
	if !canModify(actor, current) {
		return nil, ErrForbidden
	}
	ticket, err := s.repo.UpdateStatus(ctx, ticketID, status)
	if err != nil || ticket == nil {
		return ticket, err
	}
	actorID := actor.ID
	entityType := "ticket"
	entityID := ticket.ID
	desc := fmt.Sprintf("%s memindahkan tiket %s ke %s", actor.Name, ticket.Title, formatStatusLabel(status))
	_ = s.audit.Log(ctx, "ticket_status", desc, &actorID, &entityType, &entityID)
	s.repo.AddProjectActivity(ctx, ticket.ProjectID, &actorID, desc)

	userID := actor.ID
	if ticket.AssigneeID != nil && *ticket.AssigneeID != "" {
		userID = *ticket.AssigneeID
	}
	xp := priorityXP[ticket.Priority]
	if xp == 0 {
		xp = priorityXP["medium"]
	}
	if status == "done" && !wasDone {
		_ = s.gamification.AdjustXP(ctx, gamification.AdjustInput{
			UserID:      userID,
			TicketID:    ticket.ID,
			Priority:    ticket.Priority,
			XP:          xp,
			Note:        fmt.Sprintf("ticket %s completed", ticket.Title),
			ClosedDelta: 1,
		})
		_ = s.gamification.RefreshClosedCount(ctx, userID)
	}
	if wasDone && status != "done" {
		_ = s.gamification.AdjustXP(ctx, gamification.AdjustInput{
			UserID:      userID,
			TicketID:    ticket.ID,
			Priority:    ticket.Priority,
			XP:          -xp,
			Note:        fmt.Sprintf("ticket %s reopened", ticket.Title),
			ClosedDelta: -1,
		})
		_ = s.gamification.RefreshClosedCount(ctx, userID)
	}
	return ticket, nil
}

func (s *Service) UpdateDetails(ctx context.Context, actor *middleware.UserContext, ticketID string, input UpdateInput) (*Ticket, error) {
	current, err := s.repo.Get(ctx, ticketID)
	if err != nil {
		return nil, err
	}
	if current == nil {
		return nil, nil
	}
	if !canModify(actor, current) {
		return nil, ErrForbidden
	}

	if input.EpicID != nil && *input.EpicID != "" {
		ok, err := s.repo.EpicBelongsToProject(ctx, *input.EpicID, current.ProjectID)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, ErrEpicProjectMismatch
		}
	}

	ticket, err := s.repo.UpdateFields(ctx, ticketID, input)
	if err != nil || ticket == nil {
		return ticket, err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s memperbarui tiket %s", actor.Name, ticket.Title)
		actorID := actor.ID
		entityType := "ticket"
		entityID := ticket.ID
		_ = s.audit.Log(ctx, "ticket_updated", desc, &actorID, &entityType, &entityID)
	}
	// keep activity concise; log only when something actually changed
	if ticket != nil {
		desc := fmt.Sprintf("%s memperbarui tiket %s", actor.Name, ticket.Title)
		s.repo.AddProjectActivity(ctx, ticket.ProjectID, &actor.ID, desc)
	}
	return ticket, nil
}

func (s *Service) AddComment(ctx context.Context, actor *middleware.UserContext, ticketID, text string) (*Comment, error) {
	if text == "" {
		return nil, fmt.Errorf("comment text required")
	}
	if actor == nil {
		return nil, ErrForbidden
	}
	comment, err := s.repo.AddComment(ctx, ticketID, actor.ID, text)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s menambahkan komentar pada tiket %s", actor.Name, ticketID)
		actorID := actor.ID
		entityType := "ticket"
		entityID := ticketID
		_ = s.audit.Log(ctx, "ticket_commented", desc, &actorID, &entityType, &entityID)
	}
	// add comment activity
	tk, _ := s.repo.Get(ctx, ticketID)
	if tk != nil {
		desc := fmt.Sprintf("%s menambahkan komentar pada tiket %s", actor.Name, tk.Title)
		s.repo.AddProjectActivity(ctx, tk.ProjectID, &actor.ID, desc)
	}
	return comment, nil
}

func (s *Service) UpdateComment(ctx context.Context, actor *middleware.UserContext, commentID, text string) (*Comment, error) {
	if actor == nil {
		return nil, ErrForbidden
	}
	if text == "" {
		return nil, fmt.Errorf("comment text required")
	}
	comment, err := s.repo.UpdateComment(ctx, commentID, actor.ID, text)
	if err != nil {
		return nil, err
	}
	if comment == nil {
		return nil, ErrForbidden
	}
	return comment, nil
}

func (s *Service) DeleteComment(ctx context.Context, actor *middleware.UserContext, commentID string) error {
	if actor == nil {
		return ErrForbidden
	}
	if err := s.repo.DeleteComment(ctx, commentID, actor.ID); err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, actor *middleware.UserContext, ticketID string) error {
	current, err := s.repo.Get(ctx, ticketID)
	if err != nil {
		return err
	}
	if current == nil {
		return nil
	}
	if !canModify(actor, current) {
		return ErrForbidden
	}
	if err := s.repo.Delete(ctx, ticketID); err != nil {
		return err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s menghapus tiket %s", actor.Name, ticketID)
		actorID := actor.ID
		entityType := "ticket"
		entityID := ticketID
		_ = s.audit.Log(ctx, "ticket_deleted", desc, &actorID, &entityType, &entityID)
	}
	if current != nil {
		desc := fmt.Sprintf("%s menghapus tiket %s", actor.Name, current.Title)
		s.repo.AddProjectActivity(ctx, current.ProjectID, &actor.ID, desc)
	}
	return nil
}
