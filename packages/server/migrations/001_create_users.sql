
create table if not exists users(
  "id" serial primary key,  
  "first_name" varchar(200) not null,
  "last_name" varchar(200) not null,
  "password" varchar(200) not null,
  "email" varchar(200) not null,
  "mobile" varchar(15) default null,
  "address" varchar(100) default null,
  "username" varchar(200) default null,
  "gender" varchar(10) default null,
  "password_hash" varchar(255) default null,
  "is_verified" int default null
);

---- create above / drop below ----

drop table users;