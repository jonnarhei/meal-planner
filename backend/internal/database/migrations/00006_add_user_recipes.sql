-- +goose Up
CREATE TABLE IF NOT EXISTS user_recipes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    image TEXT NOT NULL DEFAULT '',
    source_url TEXT NOT NULL DEFAULT '',
    instructions TEXT NOT NULL DEFAULT '',
    servings INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_recipe_ingredients (
    id BIGSERIAL PRIMARY KEY,
    user_recipe_id INT NOT NULL REFERENCES user_recipes(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    amount NUMERIC NOT NULL DEFAULT 0,
    unit TEXT NOT NULL DEFAULT '',
    position INT NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_user_recipes_user_id ON user_recipes(user_id);

CREATE INDEX IF NOT EXISTS idx_user_recipe_ingredients_recipe_id ON user_recipe_ingredients(user_recipe_id);

-- +goose Down
DROP TABLE IF EXISTS user_recipe_ingredients;
DROP TABLE IF EXISTS user_recipes;