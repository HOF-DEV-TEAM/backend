---- Create global_parameters table ----

CREATE TABLE IF NOT EXISTS global_parameters (
    id                     UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    activate_subscription  BOOLEAN NOT NULL DEFAULT TRUE
);

-- Seed a single row of defaults if none exists.
INSERT INTO global_parameters (activate_subscription)
SELECT TRUE
WHERE NOT EXISTS (SELECT 1 FROM global_parameters);

---- Create global_parameters table (down) ----

DROP TABLE IF EXISTS global_parameters;
