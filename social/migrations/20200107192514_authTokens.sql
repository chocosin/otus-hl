-- +goose Up
-- SQL in this section is executed when the migration is applied.
create table if not exists auth_tokens
(
    token  char(36),
    userID char(36),
    PRIMARY KEY (token, userID)
);

-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
drop table auth_tokens;