create table if not exists favourites(
    "id" uuid not null default gen_random_uuid() primary key,
    "user_id" uuid default null,
    "message_id" uuid default null,
    "series_id" uuid default null,
    "fav" BOOLEAN default null,
    "date_added" timestamp default null,
    "deleted_at" timestamp default null
    );

---- create above / drop below ----

drop table favourites;