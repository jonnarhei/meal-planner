-- +goose Up
CREATE TABLE shopping_list_items (
    id         bigserial PRIMARY KEY,
    user_id    bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name       text NOT NULL,
    amount     numeric NOT NULL DEFAULT 0,
    unit       text NOT NULL DEFAULT '',
    checked    boolean NOT NULL DEFAULT false,
    source     text NOT NULL DEFAULT 'manual',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE (user_id, name, unit)
);


-- +goose Down
DROP TABLE IF EXISTS shopping_list_items;
