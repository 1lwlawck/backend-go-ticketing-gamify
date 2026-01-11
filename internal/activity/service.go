package activity

import (
	"context"
	"time"
)

// Service provides business logic for activity.
type Service struct {
	repo *Repository
}

// NewService creates a new activity service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetUserActivity returns activity log for a user.
func (s *Service) GetUserActivity(ctx context.Context, filter Filter) ([]ActivityItem, *time.Time, error) {
	return s.repo.GetUserActivity(ctx, filter)
}
