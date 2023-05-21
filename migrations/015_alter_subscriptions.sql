-- Write your migrate up statements here
ALTER TABLE subscriptions
    ADD COLUMN sub_code VARCHAR(200) DEFAULT NULL;

---- create above / drop below ----

ALTER TABLE subscriptions
    DROP COLUMN sub_code;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
