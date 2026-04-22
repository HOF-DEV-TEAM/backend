-- Add pricing/type fields to subscription_plan_offerings that the entity now requires
ALTER TABLE subscription_plan_offerings
    ADD COLUMN IF NOT EXISTS name     VARCHAR(200) NOT NULL DEFAULT '',
    ADD COLUMN IF NOT EXISTS freq     INT          NOT NULL DEFAULT 3,
    ADD COLUMN IF NOT EXISTS type     INT          NOT NULL DEFAULT 0,
    ADD COLUMN IF NOT EXISTS fee      DECIMAL(10,2)         DEFAULT 0,
    ADD COLUMN IF NOT EXISTS code     VARCHAR(200)          DEFAULT '',
    ADD COLUMN IF NOT EXISTS currency VARCHAR(10)           DEFAULT 'NGN';

---- create above / drop below ----

ALTER TABLE subscription_plan_offerings
    DROP COLUMN IF EXISTS name,
    DROP COLUMN IF EXISTS freq,
    DROP COLUMN IF EXISTS type,
    DROP COLUMN IF EXISTS fee,
    DROP COLUMN IF EXISTS code,
    DROP COLUMN IF EXISTS currency;
