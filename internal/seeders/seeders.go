package seeders

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Options controls how many fake records are created.
type Options struct {
	Users    int
	Projects int
	Tickets  int
	Comments int
	// Preset switches seeding strategy; use "demo"/"realistic" to load curated data.
	Preset string
}

type seedUser struct {
	ID       string
	Name     string
	Username string
}

type seedProject struct {
	ID     string
	Name   string
	Owner  string
	Member []string
}

// SeedAll populates users, projects, tickets, and comments with fake data.
// It is idempotent enough for dev: conflicts are ignored.
func SeedAll(ctx context.Context, db *pgxpool.Pool, opt Options) error {
	switch strings.ToLower(opt.Preset) {
	case "demo", "realistic", "sample", "":
		return SeedSampleData(ctx, db)
	case "random":
		// fallthrough to random generation
	default:
		// also default to random if unknown? No, default to sample.
		return SeedSampleData(ctx, db)
	}

	if opt.Users == 0 {
		opt.Users = 10
	}
	if opt.Projects == 0 {
		opt.Projects = 3
	}
	if opt.Tickets == 0 {
		opt.Tickets = 25
	}
	if opt.Comments == 0 {
		opt.Comments = 40
	}

	users, err := seedUsers(ctx, db, opt.Users)
	if err != nil {
		return err
	}
	projects, err := seedProjects(ctx, db, opt.Projects, users)
	if err != nil {
		return err
	}
	if err := seedTickets(ctx, db, opt.Tickets, projects); err != nil {
		return err
	}
	if err := seedComments(ctx, db, opt.Comments); err != nil {
		return err
	}
	if err := seedGamificationStats(ctx, db, users); err != nil {
		return err
	}
	log.Println("seeding complete")
	return nil
}

// CleanAll truncates all tables to ensure a fresh start.
func CleanAll(ctx context.Context, db *pgxpool.Pool) error {
	tables := []string{
		"ticket_comments", "ticket_history", "tickets", "epics",
		"project_activity", "project_members", "projects",
		"xp_events", "gamification_user_stats", "users",
	}
	for _, table := range tables {
		if _, err := db.Exec(ctx, fmt.Sprintf("TRUNCATE TABLE %s CASCADE", table)); err != nil {
			// Ignore error if table doesn't exist (e.g. fresh db), but usually we want to know.
			// For now let's log and continue or return error?
			// Better to return error to be safe.
			// But if table might not exist we should check.
			// Assuming schema exists.
			return fmt.Errorf("truncate %s: %w", table, err)
		}
	}
	log.Println("database cleaned")
	return nil
}

func seedUsers(ctx context.Context, db *pgxpool.Pool, n int) ([]seedUser, error) {
	const passwordHash = "$2a$10$uPyE3.vi.GBwMRC4b9N8.OMICBTEGbwwhe05UaZBo8zeMtruchomZi" // "password"ypt for "password"
	var out []seedUser
	for i := 0; i < n; i++ {
		id := uuid.NewString()
		name := gofakeit.Name()
		username := fmt.Sprintf("%s_%s", gofakeit.Username(), gofakeit.LetterN(4))
		role := randomChoice([]string{"developer", "developer", "project_manager", "admin"})
		_, err := db.Exec(ctx, `
INSERT INTO users (id, name, username, password_hash, role, badges, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, ARRAY[]::text[], NOW(), NOW())
ON CONFLICT (username) DO NOTHING`, id, name, username, passwordHash, role)
		if err != nil {
			return nil, err
		}
		out = append(out, seedUser{ID: id, Name: name, Username: username})
	}
	return out, nil
}

func seedProjects(ctx context.Context, db *pgxpool.Pool, n int, users []seedUser) ([]seedProject, error) {
	var projects []seedProject
	for i := 0; i < n; i++ {
		id := uuid.NewString()
		owner := users[gofakeit.Number(0, len(users)-1)].ID
		name := gofakeit.AppName()
		desc := gofakeit.Sentence(10)
		status := randomChoice([]string{"Active", "Active", "Active", "Active"})

		_, err := db.Exec(ctx, `
INSERT INTO projects (id, name, description, status, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
ON CONFLICT (id) DO NOTHING`, id, name, desc, status, owner)
		if err != nil {
			return nil, err
		}

		// members: owner as lead, plus a few random members
		memberCount := gofakeit.Number(1, min(4, len(users)))
		members := map[string]string{owner: "lead"}
		for len(members) < memberCount {
			u := users[gofakeit.Number(0, len(users)-1)].ID
			if _, exists := members[u]; !exists {
				members[u] = randomChoice([]string{"member", "member", "viewer"})
			}
		}
		for uid, role := range members {
			_, err := db.Exec(ctx, `
INSERT INTO project_members (project_id, user_id, member_role, joined_at)
VALUES ($1, $2, $3::project_member_role, NOW())
ON CONFLICT (project_id, user_id) DO NOTHING`, id, uid, role)
			if err != nil {
				return nil, err
			}
		}
		projects = append(projects, seedProject{ID: id, Name: name, Owner: owner})
	}
	return projects, nil
}

func seedTickets(ctx context.Context, db *pgxpool.Pool, n int, projects []seedProject) error {
	statuses := []string{"todo", "in_progress", "review", "done"}
	priorities := []string{"low", "medium", "high", "urgent"}
	types := []string{"feature", "bug"}

	for i := 0; i < n; i++ {
		project := projects[gofakeit.Number(0, len(projects)-1)]
		ticketID := uuid.NewString()
		title := gofakeit.Sentence(4)
		desc := gofakeit.Paragraph(1, 3, 12, " ")
		status := randomChoice(statuses)
		priority := randomChoice(priorities)
		typ := randomChoice(types)
		reporter := project.Owner
		var assignee *string
		if gofakeit.Bool() {
			id := project.Owner
			assignee = &id
		}
		due := time.Now().AddDate(0, 0, gofakeit.Number(2, 30))

		// random created_at in last 30 days
		daysAgo := gofakeit.Number(0, 30)
		createdAt := time.Now().AddDate(0, 0, -daysAgo)
		// updated_at slightly after created
		updatedAt := createdAt.Add(time.Duration(gofakeit.Number(1, 24)) * time.Hour)

		_, err := db.Exec(ctx, `
INSERT INTO tickets (id, project_id, title, description, status, priority, type, reporter_id, assignee_id, due_date, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5::ticket_status, $6::ticket_priority, $7::ticket_type, $8, $9, $10, $11, $12)
ON CONFLICT (id) DO NOTHING`, ticketID, project.ID, title, desc, status, priority, typ, reporter, assignee, due, createdAt, updatedAt)
		if err != nil {
			return err
		}
		// add history entry
		_, _ = db.Exec(ctx, `
INSERT INTO ticket_history (id, ticket_id, text, actor_id, timestamp)
VALUES ($1, $2, $3, $4, NOW())`, uuid.NewString(), ticketID, fmt.Sprintf("Ticket created (%s)", status), reporter)
	}
	return nil
}

func seedComments(ctx context.Context, db *pgxpool.Pool, n int) error {
	// collect some ticket/user IDs
	ticketRows, err := db.Query(ctx, `SELECT id FROM tickets ORDER BY random() LIMIT 50`)
	if err != nil {
		return err
	}
	defer ticketRows.Close()
	var ticketIDs []string
	for ticketRows.Next() {
		var id string
		if err := ticketRows.Scan(&id); err != nil {
			return err
		}
		ticketIDs = append(ticketIDs, id)
	}
	userRows, err := db.Query(ctx, `SELECT id FROM users ORDER BY random() LIMIT 50`)
	if err != nil {
		return err
	}
	defer userRows.Close()
	var userIDs []string
	for userRows.Next() {
		var id string
		if err := userRows.Scan(&id); err != nil {
			return err
		}
		userIDs = append(userIDs, id)
	}
	if len(ticketIDs) == 0 || len(userIDs) == 0 {
		return nil
	}

	for i := 0; i < n; i++ {
		ticketID := ticketIDs[gofakeit.Number(0, len(ticketIDs)-1)]
		author := userIDs[gofakeit.Number(0, len(userIDs)-1)]
		body := gofakeit.Sentence(8)
		_, err := db.Exec(ctx, `
INSERT INTO ticket_comments (id, ticket_id, author_id, text, created_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (id) DO NOTHING`, uuid.NewString(), ticketID, author, body)
		if err != nil {
			return err
		}
	}
	return nil
}

func seedGamificationStats(ctx context.Context, db *pgxpool.Pool, users []seedUser) error {
	for _, u := range users {
		xp := gofakeit.Number(0, 5000)
		level := xp / 1000
		if level < 1 {
			level = 1
		}
		closed := gofakeit.Number(0, 50)
		streak := gofakeit.Number(0, 10)
		lastActive := time.Now().AddDate(0, 0, -gofakeit.Number(0, 5))

		_, err := db.Exec(ctx, `
INSERT INTO gamification_user_stats (user_id, xp_total, level, next_level_threshold, tickets_closed_count, streak_days, last_ticket_closed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id) DO UPDATE SET 
	xp_total = EXCLUDED.xp_total,
	level = EXCLUDED.level,
	tickets_closed_count = EXCLUDED.tickets_closed_count,
	streak_days = EXCLUDED.streak_days
`, u.ID, xp, level, (level+1)*1000, closed, streak, lastActive)
		if err != nil {
			return err
		}
	}
	return nil
}

func randomChoice[T any](items []T) T {
	return items[gofakeit.Number(0, len(items)-1)]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
