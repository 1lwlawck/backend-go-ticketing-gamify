package achievements

import "context"

// Service provides business logic for achievements.
type Service struct {
	repo *Repository
}

// NewService creates a new achievements service.
func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

// GetAllAchievements returns all available achievements.
func (s *Service) GetAllAchievements() []Achievement {
	return DefaultAchievements()
}

// GetUserProgress returns user's progress toward all achievements.
func (s *Service) GetUserProgress(ctx context.Context, userID string) ([]Progress, error) {
	return s.repo.GetUserProgress(ctx, userID)
}

// GetUnlockedAchievements returns achievements the user has unlocked.
func (s *Service) GetUnlockedAchievements(ctx context.Context, userID string) ([]Progress, error) {
	return s.repo.GetUnlockedAchievements(ctx, userID)
}
