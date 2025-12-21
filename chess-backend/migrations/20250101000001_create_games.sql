-- +goose Up
CREATE TABLE games (
    id TEXT PRIMARY KEY,
    start_fen TEXT NOT NULL,
    current_fen TEXT NOT NULL,
    result TEXT NOT NULL DEFAULT 'ongoing',
    winner TEXT,
    ended_by TEXT,
    pending_draw_offer_by TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

-- +goose Down
DROP TABLE games;
