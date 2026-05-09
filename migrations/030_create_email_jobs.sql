CREATE TABLE IF NOT EXISTS email_jobs (
    id              VARCHAR(36)  PRIMARY KEY,
    "to"            VARCHAR(255) NOT NULL,
    subject         VARCHAR(255) NOT NULL,
    template        VARCHAR(255) NOT NULL,
    data_json       TEXT         NOT NULL,
    status          VARCHAR(20)  NOT NULL DEFAULT 'pending',
    attempts        INT          NOT NULL DEFAULT 0,
    last_error      TEXT,
    scheduled_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    last_attempted_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_email_jobs_status_scheduled
    ON email_jobs (status, scheduled_at)
    WHERE status = 'pending';

---- create above / drop below ----

DROP INDEX IF EXISTS idx_email_jobs_status_scheduled;
DROP TABLE IF EXISTS email_jobs;
