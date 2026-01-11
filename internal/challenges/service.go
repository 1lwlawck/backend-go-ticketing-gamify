package challenges

import "context"

// Service provides business logic for challenges.
type Service struct {
	repo *Repository
}

// NewService creates a new challenges service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetActiveChallenges returns current week's active challenges.
func (s *Service) GetActiveChallenges() []Challenge {
	return GetWeeklyChallenges()
}

// GetUserProgress returns user's progress on current challenges.
func (s *Service) GetUserProgress(ctx context.Context, userID string) ([]UserChallenge, error) {
	return s.repo.GetUserChallengeProgress(ctx, userID)
}
