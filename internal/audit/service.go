package audit

import (
	"context"

	"github.com/google/uuid"
)

// Service wraps repository operations.
type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, limit int) ([]Entry, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	return s.repo.List(ctx, limit)
}

func (s *Service) Log(ctx context.Context, action, description string, actorID, entityType, entityID *string) error {
	entry := Entry{
		ID:          uuid.NewString(),
		Action:      action,
		Description: description,
		ActorID:     actorID,
		EntityType:  entityType,
		EntityID:    entityID,
	}
	return s.repo.Insert(ctx, entry)
}
