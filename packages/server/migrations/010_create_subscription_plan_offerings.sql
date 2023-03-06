-- Write your migrate up statements here
CREATE TABLE IF NOT EXISTS subscription_plan_offerings(
    "id" UUID NOT NULL DEFAULT GEN_RANDOM_UUID() PRIMARY KEY,     
    "status" INT DEFAULT 1,
    "date_added" TIMESTAMP DEFAULT NULL,
    "last_updated" TIMESTAMP DEFAULT NULL,
    "deleted_at" TIMESTAMP DEFAULT NULL,
    "subscription_plan_id" UUID NOT NULL,
    "subscription_offering_id" UUID NOT NULL,
    CONSTRAINT "Fk_subscription_offerings" FOREIGN KEY ("subscription_offering_id") REFERENCES "subscription_offerings" ("id"),
    CONSTRAINT "Fk_subscription_plans" FOREIGN KEY ("subscription_plan_id") REFERENCES "subscription_plans" ("id")
)

---- create above / drop below ----

DROP TABLE subscription_plan_offerings;
