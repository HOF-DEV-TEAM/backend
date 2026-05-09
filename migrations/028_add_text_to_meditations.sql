ALTER TABLE meditations ADD COLUMN IF NOT EXISTS text TEXT;

---- create above / drop below ----

ALTER TABLE meditations DROP COLUMN IF EXISTS text;