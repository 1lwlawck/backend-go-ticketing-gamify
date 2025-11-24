-- Epics schema and ticket linkage (requires backlog value already added in 0008)
CREATE TABLE IF NOT EXISTS epics (
  id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
  project_id uuid NOT NULL REFERENCES projects(id),
  title varchar NOT NULL,
  description text,
  status ticket_status NOT NULL DEFAULT 'backlog',
  start_date timestamptz,
  due_date timestamptz,
  owner_id uuid REFERENCES users(id),
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

ALTER TABLE tickets
  ADD COLUMN IF NOT EXISTS epic_id uuid REFERENCES epics(id);

CREATE INDEX IF NOT EXISTS idx_epics_project ON epics(project_id);
CREATE INDEX IF NOT EXISTS idx_tickets_epic ON tickets(epic_id);
