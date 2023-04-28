-- Write your migrate up statements here
create table if not exists subscriptions(
    "id" uuid not null default gen_random_uuid() primary key, 
    "status" int default 1,    
    "user_id" uuid not null,
    "subscription_plan_id" uuid not null,
    "date_added" timestamp default null,
    "last_updated" timestamp default null,
    "next_payment_date" timestamp default null,
    "deleted_at" timestamp default null,
    constraint "Fk_users" foreign key ("user_id") references "users" ("id"),
    constraint "Fk_subscription_plans" foreign key ("subscription_plan_id") references "subscription_plans" ("id")

)

---- create above / drop below ----

drop table subscriptions;
