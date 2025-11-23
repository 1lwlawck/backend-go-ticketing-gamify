ALTER TABLE tickets
ADD COLUMN IF NOT EXISTS start_date timestamptz NULL;
