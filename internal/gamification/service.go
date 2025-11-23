package gamification

import "context"

type LeaderboardRow struct {
	ID                 string `json:"id"`
	Name               string `json:"name"`
	Username           string `json:"username"`
	Role               string `json:"role"`
	XP                 int    `json:"xp"`
	Level              int    `json:"level"`
	TicketsClosedCount int    `json:"ticketsClosedCount"`
	Rank               int    `json:"rank"`
	XPGap              int    `json:"xpGap"`
}

// Service exposes business logic for gamification.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetStats(ctx context.Context, userID string) (*UserStats, error) {
	return s.repo.GetStats(ctx, userID)
}

func (s *Service) ListEvents(ctx context.Context, userID string, limit int) ([]XPEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.ListEvents(ctx, userID, limit)
}

func (s *Service) AwardXP(ctx context.Context, input AwardInput) error {
	if input.UserID == "" || input.XP <= 0 {
		return nil
	}
	return s.repo.Adjust(ctx, AdjustInput{
		UserID:      input.UserID,
		TicketID:    input.TicketID,
		Priority:    input.Priority,
		XP:          input.XP,
		Note:        input.Note,
		ClosedDelta: 1,
	})
}

// AdjustXP allows applying negative XP (rollback) and adjusting closed ticket count.
func (s *Service) AdjustXP(ctx context.Context, input AdjustInput) error {
	if input.UserID == "" || input.XP == 0 {
		return nil
	}
	return s.repo.Adjust(ctx, input)
}

func (s *Service) EnsureUser(ctx context.Context, userID string) error {
	if userID == "" {
		return nil
	}
	return s.repo.EnsureUser(ctx, userID)
}

func (s *Service) Leaderboard(ctx context.Context, limit int) ([]LeaderboardRow, error) {
	return s.repo.Leaderboard(ctx, limit)
}
