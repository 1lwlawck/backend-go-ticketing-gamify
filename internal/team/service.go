package team

import "context"

// Service provides business logic for team.
type Service struct {
	repo *Repository
}

// NewService creates a new team service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetMembers returns all team members, filtered by scope if necessary.
func (s *Service) GetMembers(ctx context.Context, userID, role string, limit int) ([]Member, error) {
	// Admin and Project Manager can see everyone
	if role == "admin" || role == "project_manager" {
		return s.repo.GetMembers(ctx, limit)
	}
	// Developers (and others) see only teammates
	return s.repo.GetTeammates(ctx, userID, limit)
}

// GetProjectMembers returns members of a specific project.
func (s *Service) GetProjectMembers(ctx context.Context, projectID string) ([]ProjectMember, error) {
	return s.repo.GetProjectMembers(ctx, projectID)
}
