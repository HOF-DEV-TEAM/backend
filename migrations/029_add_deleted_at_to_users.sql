ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMPTZ;

---- create above / drop below ----

ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
