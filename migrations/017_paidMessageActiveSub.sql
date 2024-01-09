ALTER TABLE audio_messages
ADD COLUMN is_free BOOLEAN default true;

create table if not exists global_variables(
    "id" uuid not null default gen_random_uuid() primary key,
    "activate_subscription" BOOLEAN default false,
    "date_created" timestamp default null,
    "last_updated" timestamp default null
);

-- Insert values into app_version table
INSERT INTO global_variables (activate_subscription, date_created)
VALUES (false, '2024-01-09 00:00:00');