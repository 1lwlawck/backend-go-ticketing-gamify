-- enums & extension
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'user_role') THEN
    CREATE TYPE user_role AS ENUM ('developer','project_manager','admin');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'project_status') THEN
    CREATE TYPE project_status AS ENUM ('Active','Archived');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ticket_status') THEN
    CREATE TYPE ticket_status AS ENUM ('todo','in_progress','review','done','backlog');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ticket_priority') THEN
    CREATE TYPE ticket_priority AS ENUM ('low','medium','high','urgent');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'ticket_type') THEN
    CREATE TYPE ticket_type AS ENUM ('feature','bug','chore');
  END IF;
  IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'project_member_role') THEN
    CREATE TYPE project_member_role AS ENUM ('member','lead','viewer');
  END IF;
END$$;

-- Tables in dependency order

CREATE TABLE IF NOT EXISTS public.users (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name character varying NOT NULL,
  username character varying NOT NULL UNIQUE,
  password_hash character varying NOT NULL,
  role user_role NOT NULL DEFAULT 'developer',
  avatar_url character varying,
  badges text[] NOT NULL DEFAULT ARRAY[]::text[],
  bio text,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.projects (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  name character varying NOT NULL,
  description text,
  status project_status NOT NULL DEFAULT 'Active',
  created_by uuid NOT NULL REFERENCES public.users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.project_members (
  project_id uuid NOT NULL REFERENCES public.projects(id),
  user_id uuid NOT NULL REFERENCES public.users(id),
  member_role project_member_role NOT NULL DEFAULT 'member',
  joined_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (project_id, user_id)
);

CREATE TABLE IF NOT EXISTS public.project_invites (
  code character varying PRIMARY KEY,
  project_id uuid NOT NULL REFERENCES public.projects(id),
  max_uses integer NOT NULL DEFAULT 1,
  uses integer NOT NULL DEFAULT 0,
  expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.project_activity (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES public.projects(id),
  actor_id uuid REFERENCES public.users(id),
  message text NOT NULL,
  meta_json jsonb,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.epics (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES public.projects(id),
  title character varying NOT NULL,
  description text,
  status ticket_status NOT NULL DEFAULT 'backlog',
  start_date timestamptz,
  due_date timestamptz,
  owner_id uuid REFERENCES public.users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.tickets (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES public.projects(id),
  title character varying NOT NULL,
  description text,
  status ticket_status NOT NULL DEFAULT 'todo',
  priority ticket_priority NOT NULL DEFAULT 'medium',
  type ticket_type NOT NULL DEFAULT 'feature',
  reporter_id uuid NOT NULL REFERENCES public.users(id),
  assignee_id uuid REFERENCES public.users(id),
  due_date date,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  start_date timestamptz,
  epic_id uuid REFERENCES public.epics(id)
);

CREATE TABLE IF NOT EXISTS public.ticket_history (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id uuid NOT NULL REFERENCES public.tickets(id),
  text text NOT NULL,
  actor_id uuid REFERENCES public.users(id),
  timestamp timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.ticket_comments (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  ticket_id uuid NOT NULL REFERENCES public.tickets(id),
  author_id uuid NOT NULL REFERENCES public.users(id),
  text text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.gamification_user_stats (
  user_id uuid PRIMARY KEY REFERENCES public.users(id),
  xp_total integer NOT NULL DEFAULT 0,
  level integer NOT NULL DEFAULT 1,
  next_level_threshold integer NOT NULL DEFAULT 100,
  tickets_closed_count integer NOT NULL DEFAULT 0,
  streak_days integer NOT NULL DEFAULT 0,
  last_ticket_closed_at timestamptz
);

CREATE TABLE IF NOT EXISTS public.xp_events (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id uuid NOT NULL REFERENCES public.users(id),
  ticket_id uuid REFERENCES public.tickets(id),
  priority ticket_priority,
  xp_value integer NOT NULL,
  note text,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.refresh_tokens (
  id uuid PRIMARY KEY,
  user_id uuid NOT NULL REFERENCES public.users(id),
  token_hash text NOT NULL,
  expires_at timestamptz NOT NULL,
  revoked_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS public.audit_log (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  action character varying NOT NULL,
  description text,
  actor_id uuid REFERENCES public.users(id),
  entity_type character varying,
  entity_id uuid,
  created_at timestamptz NOT NULL DEFAULT now()
);
