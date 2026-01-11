package reports

import "context"

// Service provides business logic for reports.
type Service struct {
	repo *Repository
}

// NewService creates a new reports service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetSummary returns overall dashboard metrics.
func (s *Service) GetSummary(ctx context.Context) (*Summary, error) {
	return s.repo.GetSummary(ctx)
}

// GetStatusBreakdown returns ticket count per status.
func (s *Service) GetStatusBreakdown(ctx context.Context) ([]StatusBreakdown, error) {
	return s.repo.GetStatusBreakdown(ctx)
}

// GetPriorityBreakdown returns ticket count per priority.
func (s *Service) GetPriorityBreakdown(ctx context.Context) ([]PriorityBreakdown, error) {
	return s.repo.GetPriorityBreakdown(ctx)
}

// GetAssigneeBreakdown returns ticket count per assignee.
func (s *Service) GetAssigneeBreakdown(ctx context.Context, limit int) ([]AssigneeBreakdown, error) {
	return s.repo.GetAssigneeBreakdown(ctx, limit)
}

// GetTeamPerformance returns team performance metrics.
func (s *Service) GetTeamPerformance(ctx context.Context, limit int) ([]TeamPerformance, error) {
	return s.repo.GetTeamPerformance(ctx, limit)
}

// GetTicketTrend returns ticket creation/closure trend.
func (s *Service) GetTicketTrend(ctx context.Context, days int) ([]TicketTrend, error) {
	return s.repo.GetTicketTrend(ctx, days)
}
