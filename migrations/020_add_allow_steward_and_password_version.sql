---- Add allow_steward to audio_messages and password_version to users ----

-- allow_steward: when true, steward-role users may access this message
-- regardless of subscription status.
ALTER TABLE audio_messages
    ADD COLUMN IF NOT EXISTS allow_steward BOOLEAN NOT NULL DEFAULT FALSE;

-- password_version: tracks the hashing algorithm used for the stored password.
-- 'bcrypt' is the default for all new accounts.
-- 'md5' marks legacy accounts that have not yet been upgraded.
ALTER TABLE users
    ADD COLUMN IF NOT EXISTS password_version VARCHAR(10) NOT NULL DEFAULT 'bcrypt';

-- Mark all existing passwords as md5 so the login layer can upgrade them
-- to bcrypt transparently on the user's next successful login.
UPDATE users
SET password_version = 'md5'
WHERE password_version = 'bcrypt'
  AND length(password) = 32;  -- MD5 hex digests are exactly 32 characters.

---- Add allow_steward to audio_messages and password_version to users (down) ----

ALTER TABLE audio_messages DROP COLUMN IF EXISTS allow_steward;
ALTER TABLE users DROP COLUMN IF EXISTS password_version;
