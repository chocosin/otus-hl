-- +goose Up
-- SQL in this section is executed when the migration is applied.
create table if not exists users
(
    id        char(36) primary key,

    username  varchar(50)  not null,
    password  char(32)  not null,

    firstName varchar(50)  not null,
    lastName  varchar(50)  not null,
    age       int          not null,
    gender    char(10)     not null,
    interests varchar(255) not null
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
drop table users;