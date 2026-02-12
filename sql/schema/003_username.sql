-- +goose Up
ALTER TABLE users
ADD COLUMN username TEXT NOT NULL
DEFAULT 'unset';

-- +goose Down
ALTER TABLE users
DROP COLUMN username;