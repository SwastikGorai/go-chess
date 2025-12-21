-- +goose Up
CREATE TABLE moves (
    id BIGSERIAL PRIMARY KEY,
    game_id TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    uci TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX moves_game_id_idx ON moves(game_id);

-- +goose Down
DROP INDEX IF EXISTS moves_game_id_idx;
DROP TABLE moves;
