create table if not exists audio_messages(
    "id" uuid not null default gen_random_uuid() primary key,
    "title" varchar(200) not null,
    "author" varchar(200),
    "image_url" varchar(200),
    "audio_url" varchar(200) not null,
    "description" varchar(200),
    "date_added" timestamp default null,
    "last_updated" timestamp default null,
    "series_id" uuid default null,
    "deleted_at" timestamp default null
);

---- create above / drop below ----

drop table audio_messages;