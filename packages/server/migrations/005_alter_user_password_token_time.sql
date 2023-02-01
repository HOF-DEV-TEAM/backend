ALTER TABLE user_password_token
ALTER COLUMN password_reset_at TYPE INT
USING password_reset_at::integer;