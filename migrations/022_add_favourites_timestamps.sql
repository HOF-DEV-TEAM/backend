ALTER TABLE favourites
    ADD COLUMN IF NOT EXISTS date_added  TIMESTAMP NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS last_updated TIMESTAMP NOT NULL DEFAULT NOW(),
    ADD COLUMN IF NOT EXISTS deleted_at  TIMESTAMP;

---- add above / drop below ----

ALTER TABLE favourites
    DROP COLUMN IF EXISTS date_added,
    DROP COLUMN IF EXISTS last_updated,
    DROP COLUMN IF EXISTS deleted_at;
