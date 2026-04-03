-- +goose Up
ALTER TABLE users ADD COLUMN dietary_preferences text[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE users DROP COLUMN dietary_preferences;
