-- allow_steward is superseded by access_level; drop the column
ALTER TABLE audio_messages DROP COLUMN IF EXISTS allow_steward;

---- create above / drop below ----

ALTER TABLE audio_messages ADD COLUMN IF NOT EXISTS allow_steward BOOLEAN NOT NULL DEFAULT FALSE;
