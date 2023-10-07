-- Write your migrate up statements here
ALTER TABLE subscriptions
    ADD COLUMN email_token VARCHAR(200) DEFAULT NULL;

---- create above / drop below ----

ALTER TABLE subscriptions
DROP COLUMN email_token;
-- Write your migrate down statements here. If this migration is irreversible
-- Then delete the separator line above.
