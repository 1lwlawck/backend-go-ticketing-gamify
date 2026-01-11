package team

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository handles team queries.
type Repository struct {
	db *pgxpool.Pool
}

// NewRepository creates a new team repository.
func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// GetMembers returns all team members with their stats.
func (r *Repository) GetMembers(ctx context.Context, limit int) ([]Member, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	const query = `
		SELECT 
			u.id,
			u.name,
			u.username,
			u.role,
			u.avatar_url,
			(SELECT COUNT(DISTINCT pm.project_id) FROM project_members pm WHERE pm.user_id = u.id) as project_count,
			(SELECT COUNT(*) FROM tickets t WHERE t.assignee_id = u.id) as ticket_count,
			COALESCE(gs.xp_total, 0) as total_xp,
			COALESCE(gs.level, 1) as level,
			u.created_at
		FROM users u
		LEFT JOIN gamification_user_stats gs ON gs.user_id = u.id
		ORDER BY u.name ASC
		LIMIT $1`

	rows, err := r.db.Query(ctx, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Username, &m.Role, &m.AvatarURL, &m.ProjectCount, &m.TicketCount, &m.TotalXP, &m.Level, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// GetTeammates returns team members who share at least one project with the given user.
func (r *Repository) GetTeammates(ctx context.Context, userID string, limit int) ([]Member, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	const query = `
		SELECT 
			u.id, u.name, u.username, u.role, u.avatar_url,
			(SELECT COUNT(DISTINCT pm.project_id) FROM project_members pm WHERE pm.user_id = u.id) as project_count,
			(SELECT COUNT(*) FROM tickets t WHERE t.assignee_id = u.id) as ticket_count,
			COALESCE(gs.xp_total, 0) as total_xp,
			COALESCE(gs.level, 1) as level,
			u.created_at
		FROM users u
		LEFT JOIN gamification_user_stats gs ON gs.user_id = u.id
		WHERE EXISTS (
			SELECT 1 
			FROM project_members pm_me
			JOIN project_members pm_them ON pm_them.project_id = pm_me.project_id
			WHERE pm_me.user_id = $1 
			  AND pm_them.user_id = u.id
		)
		ORDER BY u.name ASC
		LIMIT $2`

	rows, err := r.db.Query(ctx, query, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Username, &m.Role, &m.AvatarURL, &m.ProjectCount, &m.TicketCount, &m.TotalXP, &m.Level, &m.JoinedAt); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}

// GetProjectMembers returns members of a specific project.
func (r *Repository) GetProjectMembers(ctx context.Context, projectID string) ([]ProjectMember, error) {
	const query = `
		SELECT pm.user_id, u.name, pm.project_id, p.name, pm.role
		FROM project_members pm
		JOIN users u ON u.id = pm.user_id
		JOIN projects p ON p.id = pm.project_id
		WHERE pm.project_id = $1
		ORDER BY u.name ASC`

	rows, err := r.db.Query(ctx, query, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []ProjectMember
	for rows.Next() {
		var m ProjectMember
		if err := rows.Scan(&m.UserID, &m.UserName, &m.ProjectID, &m.ProjectName, &m.Role); err != nil {
			return nil, err
		}
		members = append(members, m)
	}
	return members, rows.Err()
}
