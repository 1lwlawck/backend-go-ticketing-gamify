CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TYPE user_role AS ENUM ('admin','project_manager','developer','viewer');
CREATE TYPE project_status AS ENUM ('Active','Archived');
CREATE TYPE ticket_status AS ENUM ('todo','in_progress','review','done');
CREATE TYPE ticket_priority AS ENUM ('low','medium','high','urgent');
CREATE TYPE ticket_type AS ENUM ('feature','bug','chore');
CREATE TYPE project_member_role AS ENUM ('member','lead','viewer');

CREATE TABLE users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar NOT NULL,
  username varchar UNIQUE NOT NULL,
  password_hash varchar NOT NULL,
  role user_role NOT NULL DEFAULT 'developer',
  avatar_url varchar,
  badges text[] NOT NULL DEFAULT ARRAY[]::text[],
  bio text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name varchar NOT NULL,
  description text,
  status project_status NOT NULL DEFAULT 'Active',
  created_by uuid NOT NULL REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE project_members (
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  member_role project_member_role NOT NULL DEFAULT 'member',
  joined_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, user_id)
);

CREATE TABLE project_invites (
  code varchar PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  max_uses int NOT NULL DEFAULT 1,
  uses int NOT NULL DEFAULT 0,
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE tickets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
  title varchar NOT NULL,
  description text,
  status ticket_status NOT NULL DEFAULT 'todo',
  priority ticket_priority NOT NULL DEFAULT 'medium',
  type ticket_type NOT NULL DEFAULT 'feature',
  reporter_id uuid NOT NULL REFERENCES users(id),
  assignee_id uuid REFERENCES users(id),
  due_date date,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ticket_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  text text NOT NULL,
  actor_id uuid REFERENCES users(id),
  timestamp timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE ticket_comments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id uuid NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
  author_id uuid NOT NULL REFERENCES users(id),
  text text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE gamification_user_stats (
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  xp_total int NOT NULL DEFAULT 0,
  level int NOT NULL DEFAULT 1,
  next_level_threshold int NOT NULL DEFAULT 100,
  tickets_closed_count int NOT NULL DEFAULT 0,
  streak_days int NOT NULL DEFAULT 0,
  last_ticket_closed_at timestamptz
);

CREATE TABLE xp_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  ticket_id uuid REFERENCES tickets(id) ON DELETE SET NULL,
  priority ticket_priority,
  xp_value int NOT NULL,
  note text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE audit_log (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  action varchar NOT NULL,
  description text,
  actor_id uuid REFERENCES users(id),
  entity_type varchar,
  entity_id uuid,
  created_at timestamptz NOT NULL DEFAULT now()
);
