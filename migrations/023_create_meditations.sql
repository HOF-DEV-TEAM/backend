CREATE TABLE IF NOT EXISTS meditations (
    id         UUID        NOT NULL DEFAULT gen_random_uuid() PRIMARY KEY,
    name       VARCHAR(200) NOT NULL,
    image      VARCHAR(500),
    status     VARCHAR(50)  NOT NULL DEFAULT 'active',
    date_added TIMESTAMP   NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP
);

---- create above / drop below ----

DROP TABLE IF EXISTS meditations;
