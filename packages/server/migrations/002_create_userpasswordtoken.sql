create table userpasswordtoken(
                      "id" serial primary key,
                      "email" varchar(200) not null,
                      "password_reset_token" varchar(200) not null,
                      "password_reset_at" varchar(200) not null
);

---- create above / drop below ----

drop table users;