ALTER TABLE audio_messages ADD COLUMN IF NOT EXISTS is_private BOOLEAN NOT NULL DEFAULT FALSE;

---- create above / drop below ----

ALTER TABLE audio_messages DROP COLUMN IF EXISTS is_private;
