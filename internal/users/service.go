package users

import (
	"context"
	"fmt"

	"backend-go-ticketing-gamify/internal/audit"
)

// Service exposes user use-cases.
type Service struct {
	repo  *Repository
	audit *audit.Service
}

func NewService(repo *Repository, auditSvc *audit.Service) *Service {
	return &Service{repo: repo, audit: auditSvc}
}

func (s *Service) List(ctx context.Context) ([]User, error) {
	return s.repo.List(ctx)
}

func (s *Service) Get(ctx context.Context, id string) (*User, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) UpdateProfile(ctx context.Context, id string, input UpdateProfileInput) (*User, error) {
	user, err := s.repo.UpdateProfile(ctx, id, input)
	if err != nil || user == nil {
		return user, err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s updated profile", user.Name)
		actorID := id
		entityType := "user"
		entityID := id
		_ = s.audit.Log(ctx, "profile_updated", desc, &actorID, &entityType, &entityID)
	}
	return user, nil
}

func (s *Service) UpdateRole(ctx context.Context, id, role string) (*User, error) {
	user, err := s.repo.UpdateRole(ctx, id, role)
	if err != nil || user == nil {
		return user, err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s role changed to %s", user.Name, role)
		actorID := id
		entityType := "user"
		entityID := id
		_ = s.audit.Log(ctx, "role_changed", desc, &actorID, &entityType, &entityID)
	}
	return user, nil
}
