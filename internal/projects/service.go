package projects

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"backend-go-ticketing-gamify/internal/audit"
	"backend-go-ticketing-gamify/internal/middleware"
)

// Service wraps project operations.
type Service struct {
	repo  *Repository
	audit *audit.Service
}

func NewService(repo *Repository, audit *audit.Service) *Service {
	return &Service{repo: repo, audit: audit}
}

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotMember = errors.New("not_member")
)

func (s *Service) List(ctx context.Context, actor *middleware.UserContext, filter ListFilter) ([]Project, error) {
	if actor == nil {
		return nil, ErrForbidden
	}
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if isElevated(actor.Role) {
		return s.repo.List(ctx, filter)
	}
	return s.repo.ListForMember(ctx, actor.ID, filter)
}

func (s *Service) Get(ctx context.Context, actor *middleware.UserContext, id string) (*Detail, error) {
	if actor == nil {
		return nil, ErrForbidden
	}
	project, err := s.repo.Get(ctx, id)
	if err != nil || project == nil {
		return project, err
	}
	if isElevated(actor.Role) {
		return project, nil
	}
	isMember, err := s.repo.IsMember(ctx, id, actor.ID)
	if err != nil {
		return nil, err
	}
	if !isMember {
		return nil, ErrForbidden
	}
	return project, nil
}

func (s *Service) Create(ctx context.Context, actor *middleware.UserContext, input CreateInput) (*Detail, error) {
	project, err := s.repo.Create(ctx, actor.ID, input)
	if err != nil {
		return nil, err
	}
	desc := fmt.Sprintf("%s created project %s", actor.Name, project.Name)
	actorID := actor.ID
	entityType := "project"
	entityID := project.ID
	_ = s.audit.Log(ctx, "project_created", desc, &actorID, &entityType, &entityID)
	_ = s.repo.AddActivity(ctx, project.ID, &actorID, desc)
	return project, nil
}

func (s *Service) AddMember(ctx context.Context, actor *middleware.UserContext, projectID string, input AddMemberInput) error {
	role := normalizeMemberRole(input.Role)
	if err := s.repo.AddMember(ctx, projectID, input.UserID, role); err != nil {
		return err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s added member %s to project %s", actor.Name, input.UserID, projectID)
		actorID := actor.ID
		entityType := "project"
		entityID := projectID
		_ = s.audit.Log(ctx, "project_member_added", desc, &actorID, &entityType, &entityID)
	}
	_ = s.repo.AddActivity(ctx, projectID, &actor.ID, fmt.Sprintf("%s joined as %s", input.UserID, role))
	return nil
}

func (s *Service) CreateInvite(ctx context.Context, actor *middleware.UserContext, projectID string, input InviteInput) (*Invite, error) {
	if input.MaxUses <= 0 {
		return nil, fmt.Errorf("maxUses must be positive")
	}
	if input.ExpiryDays <= 0 {
		return nil, fmt.Errorf("expiryDays must be positive")
	}
	code := generateInviteCode()
	expires := time.Now().Add(time.Duration(input.ExpiryDays) * 24 * time.Hour)
	invite, err := s.repo.CreateInvite(ctx, projectID, code, input.MaxUses, expires)
	if err != nil {
		return nil, err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s generated invite for project %s", actor.Name, projectID)
		actorID := actor.ID
		entityType := "project"
		entityID := projectID
		_ = s.audit.Log(ctx, "project_invite_created", desc, &actorID, &entityType, &entityID)
	}
	_ = s.repo.AddActivity(ctx, projectID, &actor.ID, fmt.Sprintf("%s created an invite (%s)", actor.Name, code))
	return invite, nil
}

func (s *Service) JoinByCode(ctx context.Context, actor *middleware.UserContext, code string) (*Detail, error) {
	project, err := s.repo.JoinByCode(ctx, code, actor.ID, "member")
	if err != nil {
		return nil, err
	}
	desc := fmt.Sprintf("%s joined project %s with invite", actor.Name, project.ID)
	if s.audit != nil {
		actorID := actor.ID
		entityType := "project"
		entityID := project.ID
		_ = s.audit.Log(ctx, "project_joined", desc, &actorID, &entityType, &entityID)
	}
	_ = s.repo.AddActivity(ctx, project.ID, &actor.ID, desc)
	return project, nil
}

func (s *Service) Leave(ctx context.Context, actor *middleware.UserContext, projectID string) error {
	if actor == nil {
		return ErrForbidden
	}
	isMember, err := s.repo.IsMember(ctx, projectID, actor.ID)
	if err != nil {
		return err
	}
	if !isMember {
		return ErrNotMember
	}
	if err := s.repo.RemoveMember(ctx, projectID, actor.ID); err != nil {
		return err
	}
	if s.audit != nil {
		desc := fmt.Sprintf("%s left project %s", actor.Name, projectID)
		actorID := actor.ID
		entityType := "project"
		entityID := projectID
		_ = s.audit.Log(ctx, "project_left", desc, &actorID, &entityType, &entityID)
	}
	return nil
}

func generateInviteCode() string {
	code := strings.ReplaceAll(uuid.NewString(), "-", "")
	if len(code) > 8 {
		code = code[:8]
	}
	return strings.ToUpper(code)
}

func normalizeMemberRole(role string) string {
	switch role {
	case "admin", "project_manager", "lead":
		return "lead"
	case "viewer":
		return "viewer"
	default:
		return "member"
	}
}

func isElevated(role string) bool {
	switch role {
	case "admin", "project_manager":
		return true
	default:
		return false
	}
}
