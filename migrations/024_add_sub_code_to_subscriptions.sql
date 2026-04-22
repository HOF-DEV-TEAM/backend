ALTER TABLE subscriptions
    ADD COLUMN IF NOT EXISTS sub_code VARCHAR(200);

CREATE INDEX IF NOT EXISTS idx_subscriptions_sub_code ON subscriptions (sub_code) WHERE sub_code IS NOT NULL;

---- add above / drop below ----

DROP INDEX IF EXISTS idx_subscriptions_sub_code;
ALTER TABLE subscriptions DROP COLUMN IF EXISTS sub_code;
