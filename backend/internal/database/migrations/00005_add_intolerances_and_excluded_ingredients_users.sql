-- +goose Up
ALTER TABLE users ADD COLUMN intolerances text[] NOT NULL DEFAULT '{}';
ALTER TABLE users ADD COLUMN excluded_ingredients text[] NOT NULL DEFAULT '{}';

-- +goose Down
ALTER TABLE users DROP COLUMN intolerances;
ALTER TABLE users DROP COLUMN excluded_ingredients;