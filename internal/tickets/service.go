package tickets

import (
	"context"
	"errors"
	"fmt"

	"backend-go-ticketing-gamify/internal/audit"
	"backend-go-ticketing-gamify/internal/gamification"
	"backend-go-ticketing-gamify/internal/middleware"
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

func (s *Service) List(ctx context.Context, filter Filter) ([]Ticket, error) {
	return s.repo.List(ctx, filter)
}

func (s *Service) Get(ctx context.Context, id string) (*Ticket, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, actor *middleware.UserContext, input CreateInput) (*Ticket, error) {
	ticket, err := s.repo.Create(ctx, input)
	if err != nil {
		return nil, err
	}
	desc := fmt.Sprintf("%s created ticket %s", actor.Name, ticket.Title)
	actorID := actor.ID
	entityType := "ticket"
	entityID := ticket.ID
	_ = s.audit.Log(ctx, "ticket_created", desc, &actorID, &entityType, &entityID)
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
	desc := fmt.Sprintf("%s moved ticket %s to %s", actor.Name, ticket.Title, status)
	_ = s.audit.Log(ctx, "ticket_status", desc, &actorID, &entityType, &entityID)

	if status == "done" {
		userID := actor.ID
		if ticket.AssigneeID != nil && *ticket.AssigneeID != "" {
			userID = *ticket.AssigneeID
		}
		xp := priorityXP[ticket.Priority]
		if xp == 0 {
			xp = priorityXP["medium"]
		}
		_ = s.gamification.AwardXP(ctx, gamification.AwardInput{
			UserID:   userID,
			TicketID: ticket.ID,
			Priority: ticket.Priority,
			XP:       xp,
			Note:     fmt.Sprintf("ticket %s completed", ticket.Title),
		})
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
	ticket, err := s.repo.UpdateFields(ctx, ticketID, input)
	if err != nil || ticket == nil {
		return ticket, err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s updated ticket %s", actor.Name, ticket.Title)
		actorID := actor.ID
		entityType := "ticket"
		entityID := ticket.ID
		_ = s.audit.Log(ctx, "ticket_updated", desc, &actorID, &entityType, &entityID)
	}
	return ticket, nil
}

func (s *Service) AddComment(ctx context.Context, actor *middleware.UserContext, ticketID, text string) (*Comment, error) {
	if text == "" {
		return nil, fmt.Errorf("comment text required")
	}
	comment, err := s.repo.AddComment(ctx, ticketID, actor.ID, text)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s commented on ticket %s", actor.Name, ticketID)
		actorID := actor.ID
		entityType := "ticket"
		entityID := ticketID
		_ = s.audit.Log(ctx, "ticket_commented", desc, &actorID, &entityType, &entityID)
	}
	return comment, nil
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
		desc := fmt.Sprintf("%s deleted ticket %s", actor.Name, ticketID)
		actorID := actor.ID
		entityType := "ticket"
		entityID := ticketID
		_ = s.audit.Log(ctx, "ticket_deleted", desc, &actorID, &entityType, &entityID)
	}
	return nil
}
