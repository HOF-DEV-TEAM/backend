create table if not exists audio_series(
    "id" uuid not null default gen_random_uuid() primary key,
    "title" varchar(200) not null,
    "author" varchar(200),
    "description" varchar(200),
    "image_url" varchar(200) not null,
    "date_added" timestamp default null,
    "last_updated" timestamp default null,
    "deleted_at" timestamp default null
    );
---- create above / drop below ----

drop table audio_series;