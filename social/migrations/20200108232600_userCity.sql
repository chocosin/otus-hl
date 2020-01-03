-- +goose Up
-- SQL in this section is executed when the migration is applied.
ALTER TABLE users
    ADD COLUMN city VARCHAR(50) NOT NULL DEFAULT '';
-- +goose Down
-- SQL in this section is executed when the migration is rolled back.
ALTER TABLE users
    DROP COLUMN city;