-- Write your migrate up statements here
CREATE TABLE if NOT EXISTS subscription_plans(
    "id" UUID NOT NULL DEFAULT GEN_RANDOM_UUID() PRIMARY KEY, 
    "name" VARCHAR(200) NOT NULL,
    "type" INT,
    "freq" INT,
    "fee"  DECIMAL(12,2),
    "currency" VARCHAR(10) NOT NULL,
    "status" INT DEFAULT 0,
    "code" VARCHAR(200) DEFAULT NULL,
    "date_added" TIMESTAMP DEFAULT NULL,
    "last_updated" TIMESTAMP DEFAULT NULL,
    "deleted_at" TIMESTAMP DEFAULT NULL,
    "plan_id" VARCHAR(200) DEFAULT NULL,
    "subscription_provider_id" uuid DEFAULT NULL,
    CONSTRAINT "Fk_subscription_providers" FOREIGN KEY ("subscription_provider_id") REFERENCES "subscription_providers" ("id")
)

---- create above / drop below ----

DROP TABLE subscription_plans;
