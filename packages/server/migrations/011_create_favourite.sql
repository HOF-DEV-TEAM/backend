create table if not exists favourites(
    "id" uuid not null default gen_random_uuid() primary key,
    "user_id" uuid not null,
    "message_id" uuid default null,
    "series_id" uuid default null,
    "fav" BOOLEAN default null,
    "date_added" timestamp default null,
    "deleted_at" timestamp default null,
    constraint "Fk_users" foreign key ("user_id") references "users" ("id"),
    constraint "Fk_audio_messages" foreign key ("message_id") references "audio_messages" ("id"),
    constraint "Fk_audio_series" foreign key ("series_id") references "audio_series" ("id")
    );

---- create above / drop below ----

drop table favourites;