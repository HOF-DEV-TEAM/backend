---- Create roles table ----

CREATE TABLE IF NOT EXISTS roles (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name        VARCHAR(50) NOT NULL UNIQUE,
    description VARCHAR(200),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Seed the five canonical roles.
INSERT INTO roles (name, description) VALUES
    ('steward',       'Church steward with elevated content access'),
    ('member',        'Regular church member'),
    ('church_friend', 'Friend or visitor of the church'),
    ('team_lead',     'Internal team leader'),
    ('church_admin',  'Church administrator with full access')
ON CONFLICT (name) DO NOTHING;

---- Create roles table (down) ----

DROP TABLE IF EXISTS roles;
