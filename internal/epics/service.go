package epics

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) List(ctx context.Context, projectID string) ([]Epic, error) {
	return s.repo.ListByProject(ctx, projectID)
}

func (s *Service) Get(ctx context.Context, id string) (*Epic, error) {
	return s.repo.Get(ctx, id)
}

func (s *Service) Create(ctx context.Context, input CreateInput) (*Epic, error) {
	return s.repo.Create(ctx, input)
}

func (s *Service) Update(ctx context.Context, id string, input UpdateInput) (*Epic, error) {
	return s.repo.Update(ctx, id, input)
}

func (s *Service) Delete(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}
