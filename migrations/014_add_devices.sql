-- Create devices table if not exists
CREATE TABLE IF NOT EXISTS devices (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id UUID NOT NULL,
    devices JSONB NOT NULL,
    CONSTRAINT Fk_users FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
    );

-- Create app_version table if not exists
CREATE TABLE IF NOT EXISTS app_version (
    id UUID NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    version VARCHAR(200) NOT NULL,
    force BOOLEAN DEFAULT false,
    date_added TIMESTAMP DEFAULT NULL,
    last_updated TIMESTAMP DEFAULT NULL
    );

-- Insert values into app_version table
INSERT INTO app_version (version, force, date_added)
VALUES ('1.0', true, '2023-04-14 23:00:00');
-- Add more INSERT statements for additional rows if needed