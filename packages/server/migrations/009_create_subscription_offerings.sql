-- Write your migrate up statements here
CREATE TABLE IF NOT EXISTS subscription_offerings(
    "id" UUID NOT NULL DEFAULT GEN_RANDOM_UUID() PRIMARY KEY, 
    "name" VARCHAR(200) NOT NULL,    
    "status" INT DEFAULT 1,
    "date_added" TIMESTAMP DEFAULT NULL,
    "last_updated" TIMESTAMP DEFAULT NULL,
    "deleted_at" TIMESTAMP DEFAULT NULL,
    "subscription_provider_id" uuid DEFAULT NULL,
    CONSTRAINT "Fk_subscription_providers" FOREIGN KEY ("subscription_provider_id") REFERENCES "subscription_providers" ("id")
)

---- create above / drop below ----

DROP TABLE subscription_offerings;
