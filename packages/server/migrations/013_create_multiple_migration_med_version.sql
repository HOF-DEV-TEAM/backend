ALTER TABLE audio_series
ADD COLUMN of_the_month BOOLEAN default false;

create table if not exists meditation(
    "id" uuid not null default gen_random_uuid() primary key,
    "name" varchar(200) not null,
    "image_url" varchar(200) not null,
    "status" varchar(200) not null,
    "date_added" timestamp default null,
    "deleted_at" timestamp default null
);

---- create above / drop below ----

drop table meditation;