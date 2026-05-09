---- create above / drop below ----

-- Add access_level column to audio_messages to control viewer access
ALTER TABLE audio_messages
  ADD COLUMN IF NOT EXISTS access_level varchar(50) DEFAULT 'members';

---- create above / drop below ----

-- Drop access_level column
ALTER TABLE audio_messages
  DROP COLUMN IF EXISTS access_level;

