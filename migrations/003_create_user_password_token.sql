create table if not exists user_password_token(
    "id" uuid not null default gen_random_uuid() primary key,
    "email" varchar(200) not null,
    "password_reset_token" varchar(200) not null,    
    "password_reset_at" int,
    "validated" BOOLEAN default false
);

---- create above / drop below ----

drop table user_password_token;