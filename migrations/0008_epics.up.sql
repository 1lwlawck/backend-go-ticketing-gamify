-- Add backlog status to ticket_status enum (must be committed before usage)
DO $$
BEGIN
  ALTER TYPE ticket_status ADD VALUE IF NOT EXISTS 'backlog';
EXCEPTION
  WHEN duplicate_object THEN NULL;
END$$;
