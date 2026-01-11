package calendar

import "context"

// Service provides business logic for calendar.
type Service struct {
	repo *Repository
}

// NewService creates a new calendar service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetEvents returns calendar events.
func (s *Service) GetEvents(ctx context.Context, filter Filter) ([]CalendarEvent, error) {
	return s.repo.GetEvents(ctx, filter)
}
