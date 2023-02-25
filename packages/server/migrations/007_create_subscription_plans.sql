-- Write your migrate up statements here
create table if not exists subscription_plans(
    "id" uuid not null default gen_random_uuid() primary key, 
    "name" varchar(200) not null,
    "type" int,
    "freq" int,
    "fee"  decimal(12,2),
    "status" int default 0,
    "date_added" timestamp default null,
    "last_updated" timestamp default null,
    "deleted_at" timestamp default null,
    "subscription_provider_id" uuid default null,
    constraint "Fk_subscription_providers" foreign key ("subscription_provider_id") references "subscription_providers" ("id")
)

---- create above / drop below ----

drop table subscription_plans;
