---- create above / drop below ----

-- Add a unique partial index on audio_url ignoring soft-deleted rows
CREATE UNIQUE INDEX IF NOT EXISTS idx_audio_messages_audio_url_unique_not_deleted
  ON audio_messages (audio_url)
  WHERE deleted_at IS NULL;

---- create above / drop below ----

-- Drop the partial unique index
DROP INDEX IF EXISTS idx_audio_messages_audio_url_unique_not_deleted;

