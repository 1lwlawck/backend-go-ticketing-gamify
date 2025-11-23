package projects

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles persistence of projects.
type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) List(ctx context.Context) ([]Project, error) {
	const query = `
SELECT p.id, p.name, p.description, p.status, COALESCE(tc.cnt, 0), p.created_at
FROM projects p
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS cnt FROM tickets t WHERE t.project_id = p.id
) tc ON true
ORDER BY p.created_at DESC`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.TicketsCount, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *Repository) ListForMember(ctx context.Context, userID string) ([]Project, error) {
	const query = `
SELECT p.id, p.name, p.description, p.status, COALESCE(tc.cnt, 0), p.created_at
FROM project_members pm
JOIN projects p ON p.id = pm.project_id
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS cnt FROM tickets t WHERE t.project_id = p.id
) tc ON true
WHERE pm.user_id = $1
ORDER BY p.created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []Project
	for rows.Next() {
		var p Project
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.TicketsCount, &p.CreatedAt); err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (r *Repository) Get(ctx context.Context, id string) (*Detail, error) {
	const query = `
SELECT p.id, p.name, p.description, p.status, COALESCE(tc.cnt, 0), p.created_at
FROM projects p
LEFT JOIN LATERAL (
  SELECT COUNT(*)::int AS cnt FROM tickets t WHERE t.project_id = p.id
) tc ON true
WHERE p.id = $1`
	var detail Detail
	if err := r.db.QueryRow(ctx, query, id).Scan(&detail.ID, &detail.Name, &detail.Description, &detail.Status, &detail.TicketsCount, &detail.CreatedAt); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	const membersQuery = `
SELECT u.id, u.name, pm.member_role
FROM project_members pm
JOIN users u ON u.id = pm.user_id
WHERE pm.project_id = $1
ORDER BY u.name`
	rows, err := r.db.Query(ctx, membersQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Role); err != nil {
			return nil, err
		}
		detail.Members = append(detail.Members, m)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	const invitesQuery = `
SELECT code, max_uses, uses, expires_at
FROM project_invites
WHERE project_id = $1
ORDER BY created_at DESC`
	inviteRows, err := r.db.Query(ctx, invitesQuery, id)
	if err != nil {
		return nil, err
	}
	defer inviteRows.Close()
	for inviteRows.Next() {
		var inv Invite
		if err := inviteRows.Scan(&inv.Code, &inv.MaxUses, &inv.Uses, &inv.ExpiresAt); err != nil {
			return nil, err
		}
		detail.Invites = append(detail.Invites, inv)
	}
	if err := inviteRows.Err(); err != nil {
		return nil, err
	}

	const activityQuery = `
SELECT id, message, created_at
FROM project_activity
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT 25`
	activityRows, err := r.db.Query(ctx, activityQuery, id)
	if err != nil {
		return nil, err
	}
	defer activityRows.Close()
	for activityRows.Next() {
		var entry Activity
		if err := activityRows.Scan(&entry.ID, &entry.Text, &entry.Timestamp); err != nil {
			return nil, err
		}
		detail.Activity = append(detail.Activity, entry)
	}
	if err := activityRows.Err(); err != nil {
		return nil, err
	}

	return &detail, nil
}

// AddActivity stores a project activity entry.
func (r *Repository) AddActivity(ctx context.Context, projectID string, actorID *string, message string) error {
	const query = `
INSERT INTO project_activity (id, project_id, actor_id, message, created_at)
VALUES ($1, $2, $3, $4, NOW())`
	_, err := r.db.Exec(ctx, query, uuid.NewString(), projectID, actorID, message)
	return err
}

func (r *Repository) Create(ctx context.Context, creatorID string, input CreateInput) (*Detail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	projectID := uuid.NewString()
	now := time.Now()
	const insertProject = `
INSERT INTO projects (id, name, description, status, created_by, created_at, updated_at)
VALUES ($1, $2, $3, 'Active', $4, $5, $5)`
	if _, err := tx.Exec(ctx, insertProject, projectID, input.Name, input.Description, creatorID, now); err != nil {
		return nil, err
	}

	memberIDs := map[string]string{creatorID: "lead"}
	for _, memberID := range input.Members {
		if memberID == "" {
			continue
		}
		if _, exists := memberIDs[memberID]; !exists {
			memberIDs[memberID] = "member"
		}
	}

	const insertMember = `
INSERT INTO project_members (project_id, user_id, member_role, joined_at)
VALUES ($1, $2, $3::project_member_role, $4)
ON CONFLICT (project_id, user_id) DO NOTHING`
	for memberID, role := range memberIDs {
		if _, err := tx.Exec(ctx, insertMember, projectID, memberID, role, now); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.Get(ctx, projectID)
}

func (r *Repository) AddMember(ctx context.Context, projectID, userID, role string) error {
	const query = `
INSERT INTO project_members (project_id, user_id, member_role, joined_at)
VALUES ($1, $2, COALESCE(NULLIF($3, ''), 'member')::project_member_role, NOW())
ON CONFLICT (project_id, user_id) DO NOTHING`
	_, err := r.db.Exec(ctx, query, projectID, userID, role)
	return err
}

func (r *Repository) RemoveMember(ctx context.Context, projectID, userID string) error {
	const query = `
DELETE FROM project_members
WHERE project_id = $1 AND user_id = $2`
	_, err := r.db.Exec(ctx, query, projectID, userID)
	return err
}

func (r *Repository) IsMember(ctx context.Context, projectID, userID string) (bool, error) {
	const query = `
SELECT 1
FROM project_members
WHERE project_id = $1 AND user_id = $2
LIMIT 1`
	var dummy int
	if err := r.db.QueryRow(ctx, query, projectID, userID).Scan(&dummy); err != nil {
		if err == pgx.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) CreateInvite(ctx context.Context, projectID, code string, maxUses int, expires time.Time) (*Invite, error) {
	const query = `
INSERT INTO project_invites (project_id, code, max_uses, uses, expires_at, created_at)
VALUES ($1, $2, $3, 0, $4, NOW())
RETURNING code, max_uses, uses, expires_at`
	var invite Invite
	if err := r.db.QueryRow(ctx, query, projectID, code, maxUses, expires).Scan(&invite.Code, &invite.MaxUses, &invite.Uses, &invite.ExpiresAt); err != nil {
		return nil, err
	}
	return &invite, nil
}

func (r *Repository) JoinByCode(ctx context.Context, code, userID, role string) (*Detail, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	const inviteQuery = `
SELECT project_id, max_uses, uses, expires_at
FROM project_invites
WHERE code = $1
FOR UPDATE`
	var (
		projectID string
		maxUses   int
		uses      int
		expires   time.Time
	)
	if err := tx.QueryRow(ctx, inviteQuery, code).Scan(&projectID, &maxUses, &uses, &expires); err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("invite not found")
		}
		return nil, err
	}
	if time.Now().After(expires) {
		return nil, fmt.Errorf("invite expired")
	}
	if uses >= maxUses {
		return nil, fmt.Errorf("invite is fully used")
	}

	const updateInvite = `UPDATE project_invites SET uses = uses + 1 WHERE code = $1`
	if _, err := tx.Exec(ctx, updateInvite, code); err != nil {
		return nil, err
	}

	const addMember = `
INSERT INTO project_members (project_id, user_id, member_role, joined_at)
VALUES ($1, $2, COALESCE(NULLIF($3, ''), 'member')::project_member_role, NOW())
ON CONFLICT (project_id, user_id) DO NOTHING`
	if _, err := tx.Exec(ctx, addMember, projectID, userID, role); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return r.Get(ctx, projectID)
}
