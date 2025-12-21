-- +goose Up
CREATE TABLE moves (
    id BIGSERIAL PRIMARY KEY,
    game_id TEXT NOT NULL REFERENCES games(id) ON DELETE CASCADE,
    ply INTEGER NOT NULL,
    move_number INTEGER NOT NULL,
    color CHAR(1) NOT NULL CHECK (color IN ('w', 'b')),
    uci VARCHAR(5) NOT NULL,
    san TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (game_id, ply)
);

CREATE INDEX moves_game_ply_idx ON moves(game_id, ply);

-- +goose Down
DROP INDEX IF EXISTS moves_game_ply_idx;
DROP TABLE moves;
