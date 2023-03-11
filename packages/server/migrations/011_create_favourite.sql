create table if not exists favourites(
    "id" uuid not null default gen_random_uuid() primary key,
    "user_id" uuid not null,
    "fav" JSONB,
    constraint "Fk_users" foreign key ("user_id") references "users" ("id")
    );

---- create above / drop below ----

drop table favourites;