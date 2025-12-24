-- +goose Up
ALTER TABLE games
    ADD COLUMN player_white_token TEXT,
    ADD COLUMN player_black_token TEXT,
    ADD COLUMN player_white_joined_at TIMESTAMPTZ,
    ADD COLUMN player_black_joined_at TIMESTAMPTZ;

-- +goose Down
ALTER TABLE games
    DROP COLUMN player_white_token,
    DROP COLUMN player_black_token,
    DROP COLUMN player_white_joined_at,
    DROP COLUMN player_black_joined_at;
