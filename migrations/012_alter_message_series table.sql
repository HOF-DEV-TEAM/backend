ALTER TABLE audio_messages
ADD COLUMN date_released timestamp default null;

ALTER TABLE audio_series
ADD COLUMN date_released timestamp default null;

ALTER TABLE users
ADD COLUMN date_added timestamp default null,
ADD COLUMN last_updated timestamp default null;
