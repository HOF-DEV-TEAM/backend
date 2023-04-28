-- Write your migrate up statements here
create table if not exists subscription_providers(
    "id" uuid not null default gen_random_uuid() primary key, 
    "name" varchar(200) not null,
    "is_default" int default 0,
    "date_added" timestamp default null,
    "last_updated" timestamp default null,
    "deleted_at" timestamp default null
)

---- create above / drop below ----

drop table subscription_providers;
