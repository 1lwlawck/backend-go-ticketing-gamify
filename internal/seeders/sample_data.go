package seeders

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sampleUser struct {
	ID       uuid.UUID
	Name     string
	Username string
	Role     string
	Bio      string
	Avatar   string
}

type sampleProject struct {
	ID          uuid.UUID
	Name        string
	Description string
	Status      string
	Owner       uuid.UUID
	Members     []projectMember
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type projectMember struct {
	User uuid.UUID
	Role string
}

type sampleEpic struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	Title       string
	Description string
	Status      string
	StartDate   time.Time
	DueDate     time.Time
	Owner       uuid.UUID
}

type sampleTicket struct {
	ID          uuid.UUID
	ProjectID   uuid.UUID
	EpicID      uuid.UUID
	Title       string
	Description string
	Status      string
	Priority    string
	Type        string
	Reporter    uuid.UUID
	Assignee    *uuid.UUID
	DueDate     time.Time
	StartDate   time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type sampleHistory struct {
	ID       uuid.UUID
	TicketID uuid.UUID
	Text     string
	Actor    uuid.UUID
	Time     time.Time
}

type sampleComment struct {
	ID       uuid.UUID
	TicketID uuid.UUID
	Author   uuid.UUID
	Text     string
	Time     time.Time
}

type sampleActivity struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	ActorID   uuid.UUID
	Message   string
	Meta      string
	Time      time.Time
}

type sampleXPEvent struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	TicketID uuid.UUID
	Priority string
	Value    int
	Note     string
	Time     time.Time
}

// SeedSampleData loads a curated dataset that feels closer to a real sprint board.
func SeedSampleData(ctx context.Context, db *pgxpool.Pool) error {
	const passwordHash = "$2a$10$uPyE3.vi.GBwMRC4b9N8.OMICBTEGbwwhe05UaZBo8zeMtruchomZi" // "password"
	base := time.Now()

	users := []sampleUser{
		{
			ID:       uuid.MustParse("8bb6e0b5-4f91-4e6b-b955-45e954d5b5c1"),
			Name:     "Alexandra Reyes",
			Username: "alex.pm",
			Role:     "project_manager",
			Bio:      "Product manager untuk portal customer dan billing, fokus di reliability.",
			Avatar:   "https://i.pravatar.cc/150?img=47",
		},
		{
			ID:       uuid.MustParse("b59f82bc-9cba-4958-9bcb-9a1754bdc298"),
			Name:     "Maya Chen",
			Username: "maya.dev",
			Role:     "developer",
			Bio:      "Fullstack engineer, nyaman dengan Go, Vue, dan sistem auth.",
			Avatar:   "https://i.pravatar.cc/150?img=12",
		},
		{
			ID:       uuid.MustParse("6c4b2b42-6fe7-4e85-926d-8578ef6e7c80"),
			Name:     "Rafi Pratama",
			Username: "rafi.dev",
			Role:     "developer",
			Bio:      "Backend engineer, banyak menangani observability & infra.",
			Avatar:   "https://i.pravatar.cc/150?img=32",
		},
		{
			ID:       uuid.MustParse("f2f77b52-2977-4b7a-9f43-1f3a38192002"),
			Name:     "Nina Wibowo",
			Username: "nina.qa",
			Role:     "developer",
			Bio:      "QA/automation, merilis regression pack setiap sprint.",
			Avatar:   "https://i.pravatar.cc/150?img=58",
		},
		{
			ID:       uuid.MustParse("8120df96-26c3-4712-bdc9-98355ea26ec9"),
			Name:     "Satria Putra",
			Username: "satria.dev",
			Role:     "developer",
			Bio:      "Platform engineer, payment reliability & message queueing.",
			Avatar:   "https://i.pravatar.cc/150?img=19",
		},
		{
			ID:       uuid.MustParse("6b062d3e-1c3f-4a6f-936e-73f8c4e78a60"),
			Name:     "Adit Santoso",
			Username: "adit.admin",
			Role:     "admin",
			Bio:      "Engineering manager, memastikan SLA billing 99.95%.",
			Avatar:   "https://i.pravatar.cc/150?img=5",
		},
	}

	for _, u := range users {
		_, err := db.Exec(ctx, `
INSERT INTO users (id, name, username, password_hash, role, avatar_url, badges, bio, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5::user_role, $6, ARRAY[]::text[], $7, $8, $8)
ON CONFLICT (username) DO UPDATE
SET name = EXCLUDED.name, role = EXCLUDED.role, avatar_url = EXCLUDED.avatar_url, bio = EXCLUDED.bio, updated_at = NOW()`,
			u.ID, u.Name, u.Username, passwordHash, u.Role, u.Avatar, u.Bio, base)
		if err != nil {
			return err
		}
	}

	projects := []sampleProject{
		{
			ID:          uuid.MustParse("5b0a7e2e-4989-4d69-bc65-0c92ddc26428"),
			Name:        "Customer Portal Revamp",
			Description: "Redesign portal dengan MFA, profil mandiri, dan audit login.",
			Status:      "Active",
			Owner:       users[0].ID,
			CreatedAt:   base.AddDate(0, 0, -18),
			UpdatedAt:   base.AddDate(0, 0, -1),
			Members: []projectMember{
				{User: users[0].ID, Role: "lead"},
				{User: users[1].ID, Role: "member"},
				{User: users[2].ID, Role: "member"},
				{User: users[3].ID, Role: "viewer"},
				{User: users[5].ID, Role: "viewer"},
			},
		},
		{
			ID:          uuid.MustParse("4e7863d6-0d0c-47bb-b61d-7e3c7d2581ff"),
			Name:        "Billing Automation",
			Description: "Automasi invoice pipeline, retry payment, dan audit pengiriman.",
			Status:      "Active",
			Owner:       users[5].ID,
			CreatedAt:   base.AddDate(0, 0, -25),
			UpdatedAt:   base.AddDate(0, 0, -2),
			Members: []projectMember{
				{User: users[5].ID, Role: "lead"},
				{User: users[1].ID, Role: "member"},
				{User: users[3].ID, Role: "member"},
				{User: users[4].ID, Role: "member"},
			},
		},
	}

	for _, p := range projects {
		_, err := db.Exec(ctx, `
INSERT INTO projects (id, name, description, status, created_by, created_at, updated_at)
VALUES ($1, $2, $3, $4::project_status, $5, $6, $7)
ON CONFLICT (id) DO UPDATE
SET name = EXCLUDED.name, description = EXCLUDED.description, status = EXCLUDED.status, updated_at = NOW()`,
			p.ID, p.Name, p.Description, p.Status, p.Owner, p.CreatedAt, p.UpdatedAt)
		if err != nil {
			return err
		}
		for _, m := range p.Members {
			_, err := db.Exec(ctx, `
INSERT INTO project_members (project_id, user_id, member_role, joined_at)
VALUES ($1, $2, $3::project_member_role, $4)
ON CONFLICT (project_id, user_id) DO NOTHING`,
				p.ID, m.User, m.Role, p.CreatedAt.AddDate(0, 0, 1))
			if err != nil {
				return err
			}
		}
	}

	activities := []sampleActivity{
		{ID: uuid.MustParse("c5d3c519-5fd1-47c5-942d-3d377a0062d8"), ProjectID: projects[0].ID, ActorID: users[0].ID, Message: "Kickoff sprint fokus MFA & profil", Meta: `{"sprint":"2025-12-01"}`, Time: base.AddDate(0, 0, -14)},
		{ID: uuid.MustParse("e1a147ad-6dad-4a2b-a6af-d55f1f08cbae"), ProjectID: projects[0].ID, ActorID: users[1].ID, Message: "Migrate login page ke komponen baru", Meta: `{"branch":"feature/login-shell"}`, Time: base.AddDate(0, 0, -9)},
		{ID: uuid.MustParse("8c7a7c91-5a5b-4a6f-9fa8-c1c1d1b2c7c4"), ProjectID: projects[1].ID, ActorID: users[5].ID, Message: "Align SLA retry 3x selama 24 jam", Meta: `{"slo":"99.95"}`, Time: base.AddDate(0, 0, -12)},
		{ID: uuid.MustParse("7b0f0d85-5e41-4c4c-9bea-4a0f09d6678a"), ProjectID: projects[1].ID, ActorID: users[4].ID, Message: "Consumer DLQ sudah metrics ke Grafana", Meta: `{"dashboard":"billing-retry"}`, Time: base.AddDate(0, 0, -5)},
	}

	for _, a := range activities {
		_, err := db.Exec(ctx, `
INSERT INTO project_activity (id, project_id, actor_id, message, meta_json, created_at)
VALUES ($1, $2, $3, $4, $5::jsonb, $6)
ON CONFLICT (id) DO NOTHING`,
			a.ID, a.ProjectID, a.ActorID, a.Message, a.Meta, a.Time)
		if err != nil {
			return err
		}
	}

	epics := []sampleEpic{
		{
			ID:          uuid.MustParse("b89d8bdb-d0c7-4ef3-9c69-9ca9e446b410"),
			ProjectID:   projects[0].ID,
			Title:       "Authentication Upgrade",
			Description: "Tambah MFA, session yang aman, dan audit login.",
			Status:      "in_progress",
			StartDate:   base.AddDate(0, 0, -17),
			DueDate:     base.AddDate(0, 0, 10),
			Owner:       users[0].ID,
		},
		{
			ID:          uuid.MustParse("0f4c45a7-6d9f-4c1c-9127-5f3e2d2143f8"),
			ProjectID:   projects[0].ID,
			Title:       "Self-service Profile",
			Description: "Profil customer bisa update data sendiri dengan guard rails audit.",
			Status:      "review",
			StartDate:   base.AddDate(0, 0, -12),
			DueDate:     base.AddDate(0, 0, 6),
			Owner:       users[1].ID,
		},
		{
			ID:          uuid.MustParse("a38a61c0-2c5b-4c5b-9d4f-8f5a4f2940f7"),
			ProjectID:   projects[1].ID,
			Title:       "Invoice Pipeline",
			Description: "Dokumentasi invoice & PDF generator resilient.",
			Status:      "in_progress",
			StartDate:   base.AddDate(0, 0, -20),
			DueDate:     base.AddDate(0, 0, 8),
			Owner:       users[5].ID,
		},
		{
			ID:          uuid.MustParse("6215b6b7-6d7b-4a1e-bb68-8e5148a2fb95"),
			ProjectID:   projects[1].ID,
			Title:       "Payment Retries",
			Description: "Observability dan backoff untuk retry pembayaran.",
			Status:      "todo",
			StartDate:   base.AddDate(0, 0, -7),
			DueDate:     base.AddDate(0, 0, 14),
			Owner:       users[4].ID,
		},
	}

	for _, e := range epics {
		_, err := db.Exec(ctx, `
INSERT INTO epics (id, project_id, title, description, status, start_date, due_date, owner_id, created_at, updated_at)
VALUES ($1, $2, $3, $4, $5::ticket_status, $6, $7, $8, $9, $10)
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title, description = EXCLUDED.description, status = EXCLUDED.status, due_date = EXCLUDED.due_date, updated_at = NOW()`,
			e.ID, e.ProjectID, e.Title, e.Description, e.Status, e.StartDate, e.DueDate, e.Owner, e.StartDate, e.StartDate)
		if err != nil {
			return err
		}
	}

	tickets := []sampleTicket{
		{
			ID:          uuid.MustParse("7b77b95c-f6b7-476a-9134-45e26dfec66f"),
			ProjectID:   projects[0].ID,
			EpicID:      epics[0].ID,
			Title:       "MFA enrollment UX dan fallback OTP",
			Description: "Perbaiki alur MFA untuk user lama; fallback OTP via email jika push gagal.",
			Status:      "in_progress",
			Priority:    "high",
			Type:        "feature",
			Reporter:    users[0].ID,
			Assignee:    &users[1].ID,
			DueDate:     base.AddDate(0, 0, 9),
			StartDate:   base.AddDate(0, 0, -6),
			CreatedAt:   base.AddDate(0, 0, -8),
			UpdatedAt:   base.AddDate(0, 0, -1),
		},
		{
			ID:          uuid.MustParse("7889e7ab-2c3f-4a6d-b9f7-0e2b5a162e36"),
			ProjectID:   projects[0].ID,
			EpicID:      epics[0].ID,
			Title:       "Perbaiki session timeout flakey",
			Description: "Session logout mendadak saat idle 5m; cek redis TTL & load balancer.",
			Status:      "review",
			Priority:    "urgent",
			Type:        "bug",
			Reporter:    users[0].ID,
			Assignee:    &users[2].ID,
			DueDate:     base.AddDate(0, 0, 4),
			StartDate:   base.AddDate(0, 0, -10),
			CreatedAt:   base.AddDate(0, 0, -12),
			UpdatedAt:   base.AddDate(0, 0, -2),
		},
		{
			ID:          uuid.MustParse("f7137ff8-1c25-4e0b-9f47-7c4e4e034d4e"),
			ProjectID:   projects[0].ID,
			EpicID:      epics[0].ID,
			Title:       "Audit log login & device fingerprint",
			Description: "Simpan event login dengan fingerprint device + IP untuk audit SOC2.",
			Status:      "todo",
			Priority:    "medium",
			Type:        "feature",
			Reporter:    users[0].ID,
			Assignee:    nil,
			DueDate:     base.AddDate(0, 0, 13),
			StartDate:   base.AddDate(0, 0, -1),
			CreatedAt:   base.AddDate(0, 0, -1),
			UpdatedAt:   base.AddDate(0, 0, -1),
		},
		{
			ID:          uuid.MustParse("0ef7e4c0-54bb-4fa3-8fdb-ea88e7e423c8"),
			ProjectID:   projects[0].ID,
			EpicID:      epics[1].ID,
			Title:       "Profile editor dengan audit trail",
			Description: "Customer bisa update kontak + preferensi komunikasi; setiap perubahan dicatat.",
			Status:      "done",
			Priority:    "medium",
			Type:        "feature",
			Reporter:    users[0].ID,
			Assignee:    &users[1].ID,
			DueDate:     base.AddDate(0, 0, -1),
			StartDate:   base.AddDate(0, 0, -9),
			CreatedAt:   base.AddDate(0, 0, -10),
			UpdatedAt:   base.AddDate(0, 0, -1),
		},
		{
			ID:          uuid.MustParse("c41f7ab8-3f5b-4cb9-8d12-1a27b6a8cbbc"),
			ProjectID:   projects[1].ID,
			EpicID:      epics[2].ID,
			Title:       "Generate invoice PDF async",
			Description: "Pisah generator PDF ke worker; queue via RabbitMQ + monitoring metrics.",
			Status:      "in_progress",
			Priority:    "medium",
			Type:        "feature",
			Reporter:    users[5].ID,
			Assignee:    &users[4].ID,
			DueDate:     base.AddDate(0, 0, 11),
			StartDate:   base.AddDate(0, 0, -7),
			CreatedAt:   base.AddDate(0, 0, -9),
			UpdatedAt:   base.AddDate(0, 0, -3),
		},
		{
			ID:          uuid.MustParse("d5df1bb3-5a7e-4f8f-9f65-7c3d7ee2f9b3"),
			ProjectID:   projects[1].ID,
			EpicID:      epics[2].ID,
			Title:       "DLQ visibility dashboard",
			Description: "Bikin dashboard Grafana untuk retry queue + alert ke Slack.",
			Status:      "done",
			Priority:    "high",
			Type:        "chore",
			Reporter:    users[5].ID,
			Assignee:    &users[4].ID,
			DueDate:     base.AddDate(0, 0, -2),
			StartDate:   base.AddDate(0, 0, -11),
			CreatedAt:   base.AddDate(0, 0, -13),
			UpdatedAt:   base.AddDate(0, 0, -2),
		},
		{
			ID:          uuid.MustParse("d9d82a23-51bc-4fc6-8a8a-705b5fe7d5c0"),
			ProjectID:   projects[1].ID,
			EpicID:      epics[3].ID,
			Title:       "Validasi signature webhook",
			Description: "Verifikasi HMAC untuk partner webhook + rotate secret otomatis.",
			Status:      "done",
			Priority:    "medium",
			Type:        "feature",
			Reporter:    users[5].ID,
			Assignee:    &users[3].ID,
			DueDate:     base.AddDate(0, 0, 1),
			StartDate:   base.AddDate(0, 0, -5),
			CreatedAt:   base.AddDate(0, 0, -7),
			UpdatedAt:   base.AddDate(0, 0, 0),
		},
		{
			ID:          uuid.MustParse("a4c4b1d4-25c1-4d67-8b78-616ccc8ef7fa"),
			ProjectID:   projects[1].ID,
			EpicID:      epics[3].ID,
			Title:       "Regression pack pembayaran",
			Description: "Tambah 15 skenario critical path ke suite nightly.",
			Status:      "todo",
			Priority:    "low",
			Type:        "chore",
			Reporter:    users[5].ID,
			Assignee:    &users[3].ID,
			DueDate:     base.AddDate(0, 0, 15),
			StartDate:   base.AddDate(0, 0, 0),
			CreatedAt:   base.AddDate(0, 0, -1),
			UpdatedAt:   base.AddDate(0, 0, 0),
		},
	}

	for _, t := range tickets {
		_, err := db.Exec(ctx, `
INSERT INTO tickets (id, project_id, title, description, status, priority, type, reporter_id, assignee_id, due_date, created_at, updated_at, start_date, epic_id)
VALUES ($1, $2, $3, $4, $5::ticket_status, $6::ticket_priority, $7::ticket_type, $8, $9, $10, $11, $12, $13, $14)
ON CONFLICT (id) DO UPDATE
SET title = EXCLUDED.title, description = EXCLUDED.description, status = EXCLUDED.status, priority = EXCLUDED.priority, type = EXCLUDED.type, assignee_id = EXCLUDED.assignee_id, due_date = EXCLUDED.due_date, updated_at = NOW()`,
			t.ID, t.ProjectID, t.Title, t.Description, t.Status, t.Priority, t.Type, t.Reporter, t.Assignee, t.DueDate, t.CreatedAt, t.UpdatedAt, t.StartDate, t.EpicID)
		if err != nil {
			return err
		}
	}

	history := []sampleHistory{
		{ID: uuid.MustParse("c5b37b23-5f8d-4c14-9c7c-6bb5b05bd97c"), TicketID: tickets[0].ID, Text: "Ticket dibuat dan diambil oleh Maya", Actor: users[1].ID, Time: tickets[0].CreatedAt},
		{ID: uuid.MustParse("55dd9f5f-6d89-4054-b2b3-5b3f39eabf4d"), TicketID: tickets[0].ID, Text: "Status diubah ke in_progress", Actor: users[1].ID, Time: tickets[0].StartDate},
		{ID: uuid.MustParse("6fdab5b9-4413-4e8c-9138-6e7a3a0655e9"), TicketID: tickets[1].ID, Text: "Investigasi redis TTL selesai", Actor: users[2].ID, Time: tickets[1].StartDate},
		{ID: uuid.MustParse("1c87c5c7-384a-4d7a-8c54-d4d27a130c0a"), TicketID: tickets[1].ID, Text: "Status diubah ke review", Actor: users[2].ID, Time: tickets[1].UpdatedAt},
		{ID: uuid.MustParse("d9ad7c21-8c3f-4bd7-95a8-516eaf38e4a2"), TicketID: tickets[3].ID, Text: "Status diubah ke done setelah UAT", Actor: users[0].ID, Time: tickets[3].UpdatedAt},
		{ID: uuid.MustParse("c7e2c3e6-5f3b-4f57-a93f-58289f6a1d8c"), TicketID: tickets[5].ID, Text: "Grafana dashboard live", Actor: users[4].ID, Time: tickets[5].UpdatedAt},
		{ID: uuid.MustParse("c56e1ce2-0c3e-4d48-aa45-1d3163a5ea8e"), TicketID: tickets[6].ID, Text: "Webhook signature diverifikasi", Actor: users[3].ID, Time: tickets[6].UpdatedAt},
	}

	for _, h := range history {
		_, err := db.Exec(ctx, `
INSERT INTO ticket_history (id, ticket_id, text, actor_id, timestamp)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO NOTHING`,
			h.ID, h.TicketID, h.Text, h.Actor, h.Time)
		if err != nil {
			return err
		}
	}

	comments := []sampleComment{
		{ID: uuid.MustParse("1dbb880d-3f39-4bea-9d43-d205c3bd6c52"), TicketID: tickets[0].ID, Author: users[1].ID, Text: "Copy MFA sudah diupdate, butuh QA untuk fallback OTP.", Time: base.AddDate(0, 0, -2)},
		{ID: uuid.MustParse("5a3cfbb0-39e6-425f-9a5b-237c54b4cf2c"), TicketID: tickets[0].ID, Author: users[3].ID, Text: "Siap QA besok, mock OTP di staging ready.", Time: base.AddDate(0, 0, -1)},
		{ID: uuid.MustParse("d2a89bfb-0ccf-410f-9466-07bcb35c50c8"), TicketID: tickets[1].ID, Author: users[2].ID, Text: "Root cause: LB cookie tidak sticky saat refresh token.", Time: base.AddDate(0, 0, -3)},
		{ID: uuid.MustParse("1a1d9d44-780b-4aaf-9450-89a1a3e3c749"), TicketID: tickets[3].ID, Author: users[0].ID, Text: "UAT selesai, enable gradual rollout 20%.", Time: base.AddDate(0, 0, -1)},
		{ID: uuid.MustParse("ed8a47d3-20f0-40fb-8d10-7becbd2f30d5"), TicketID: tickets[5].ID, Author: users[4].ID, Text: "Alert Slack #oncall sudah dihubungkan.", Time: base.AddDate(0, 0, -2)},
		{ID: uuid.MustParse("b9201f4c-9a74-4f5e-99f0-1dfbea06979c"), TicketID: tickets[6].ID, Author: users[3].ID, Text: "Tambahkan example payload ke docs partner.", Time: base.AddDate(0, 0, 0)},
	}

	for _, c := range comments {
		_, err := db.Exec(ctx, `
INSERT INTO ticket_comments (id, ticket_id, author_id, text, created_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (id) DO NOTHING`,
			c.ID, c.TicketID, c.Author, c.Text, c.Time)
		if err != nil {
			return err
		}
	}

	xpEvents := []sampleXPEvent{
		{ID: uuid.MustParse("0c8778c4-5eaf-4a7d-8771-a35d7a91e8a0"), UserID: users[1].ID, TicketID: tickets[3].ID, Priority: "medium", Value: 120, Note: "Profil editor selesai", Time: tickets[3].UpdatedAt},
		{ID: uuid.MustParse("a0b9c7b2-7e61-4d31-8a08-1e9c621f4f0a"), UserID: users[4].ID, TicketID: tickets[5].ID, Priority: "high", Value: 150, Note: "DLQ observability selesai", Time: tickets[5].UpdatedAt},
		{ID: uuid.MustParse("4f17e308-15c3-42ba-b1f2-bbd1f23dbe15"), UserID: users[3].ID, TicketID: tickets[6].ID, Priority: "medium", Value: 90, Note: "Webhook validation dirilis", Time: tickets[6].UpdatedAt},
	}

	for _, xp := range xpEvents {
		_, err := db.Exec(ctx, `
INSERT INTO xp_events (id, user_id, ticket_id, priority, xp_value, note, created_at)
VALUES ($1, $2, $3, $4::ticket_priority, $5, $6, $7)
ON CONFLICT (id) DO NOTHING`,
			xp.ID, xp.UserID, xp.TicketID, xp.Priority, xp.Value, xp.Note, xp.Time)
		if err != nil {
			return err
		}
	}

	stats := []struct {
		UserID             uuid.UUID
		Total              int
		Level              int
		NextThreshold      int
		TicketsClosedCount int
		Streak             int
		LastClosedAt       time.Time
	}{
		{UserID: users[1].ID, Total: 240, Level: 3, NextThreshold: 300, TicketsClosedCount: 2, Streak: 3, LastClosedAt: tickets[3].UpdatedAt},
		{UserID: users[4].ID, Total: 180, Level: 2, NextThreshold: 250, TicketsClosedCount: 1, Streak: 2, LastClosedAt: tickets[5].UpdatedAt},
		{UserID: users[3].ID, Total: 120, Level: 2, NextThreshold: 200, TicketsClosedCount: 1, Streak: 1, LastClosedAt: tickets[6].UpdatedAt},
	}

	for _, s := range stats {
		_, err := db.Exec(ctx, `
INSERT INTO gamification_user_stats (user_id, xp_total, level, next_level_threshold, tickets_closed_count, streak_days, last_ticket_closed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id) DO UPDATE
SET xp_total = EXCLUDED.xp_total, level = EXCLUDED.level, next_level_threshold = EXCLUDED.next_level_threshold, tickets_closed_count = EXCLUDED.tickets_closed_count, streak_days = EXCLUDED.streak_days, last_ticket_closed_at = EXCLUDED.last_ticket_closed_at`,
			s.UserID, s.Total, s.Level, s.NextThreshold, s.TicketsClosedCount, s.Streak, s.LastClosedAt)
		if err != nil {
			return err
		}
	}

	log.Println("seeded curated demo dataset")
	return nil
}
