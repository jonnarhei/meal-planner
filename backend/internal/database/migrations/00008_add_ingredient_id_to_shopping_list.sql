-- +goose Up
ALTER TABLE shopping_list_items ADD COLUMN ingredient_id BIGINT;

ALTER TABLE shopping_list_items
    DROP CONSTRAINT shopping_list_items_user_id_name_unit_key;

ALTER TABLE shopping_list_items
    ADD COLUMN agg_key TEXT GENERATED ALWAYS AS (
        CASE WHEN ingredient_id IS NULL
            THEN 'name:' || name
            ELSE 'id:' || ingredient_id::text
        END
    ) STORED;

CREATE UNIQUE INDEX shopping_list_items_agg_key_idx
    ON shopping_list_items (user_id, agg_key, unit);

-- +goose Down
DROP INDEX IF EXISTS shopping_list_items_agg_key_idx;
ALTER TABLE shopping_list_items DROP COLUMN IF EXISTS agg_key;
ALTER TABLE shopping_list_items DROP COLUMN IF EXISTS ingredient_id;
ALTER TABLE shopping_list_items ADD CONSTRAINT shopping_list_items_user_id_name_unit_key UNIQUE (user_id, name, unit);