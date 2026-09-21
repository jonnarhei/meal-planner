-- +goose Up
ALTER TABLE meal_plan_recipes 
    ADD COLUMN source TEXT NOT NULL DEFAULT 'spoonacular',
    ADD COLUMN user_recipe_id BIGINT REFERENCES user_recipes(id) ON DELETE SET NULL;

ALTER TABLE meal_plan_recipes
    ADD CONSTRAINT meal_plan_recipes_source_check CHECK (source IN ('spoonacular', 'user'));

CREATE INDEX IF NOT EXISTS idx_meal_plan_recipes_user_recipe_id ON meal_plan_recipes(user_recipe_id);

-- +goose Down
DROP INDEX IF EXISTS idx_meal_plan_recipes_user_recipe_id;
ALTER TABLE meal_plan_recipes DROP CONSTRAINT IF EXISTS meal_plan_recipes_source_check;
ALTER TABLE meal_plan_recipes DROP COLUMN IF EXISTS user_recipe_id;
ALTER TABLE meal_plan_recipes DROP COLUMN IF EXISTS source;