package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"chess-backend/internal/chess"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	pool *pgxpool.Pool
}

func NewPostgresStore(ctx context.Context, connString string) (*PostgresStore, error) {
	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &PostgresStore{pool: pool}, nil
}

func (s *PostgresStore) Close() {
	s.pool.Close()
}

func (s *PostgresStore) CreateGame(ctx context.Context, game *Game) error {
	query := `
		INSERT INTO games (
			id, start_fen, current_fen, result, winner, ended_by,
			pending_draw_offer_by, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.pool.Exec(
		ctx,
		query,
		game.ID,
		game.StartFEN,
		game.Board.ToFEN(),
		normalizeResult(game.Result),
		nullIfEmpty(game.Winner),
		nullIfEmpty(game.EndedBy),
		colorToNullableString(game.PendingDrawOfferBy),
		game.CreatedAt,
		game.UpdatedAt,
	)
	return err
}

func (s *PostgresStore) GetGame(ctx context.Context, id string) (*Game, error) {
	query := `
		SELECT id, start_fen, current_fen, result, winner, ended_by,
		       pending_draw_offer_by, created_at, updated_at
		FROM games
		WHERE id = $1
	`
	var (
		gameID   string
		startFEN string
		fen      string
		result   string
		winner   sql.NullString
		endedBy  sql.NullString
		pending  sql.NullString
		created  time.Time
		updated  time.Time
	)

	err := s.pool.QueryRow(ctx, query, id).Scan(
		&gameID,
		&startFEN,
		&fen,
		&result,
		&winner,
		&endedBy,
		&pending,
		&created,
		&updated,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, ErrNotFound
		}
		return nil, err
	}

	board, err := chess.LoadFEN(fen)
	if err != nil {
		return nil, fmt.Errorf("invalid FEN in store: %w", err)
	}

	game := &Game{
		ID:        gameID,
		Board:     board,
		StartFEN:  startFEN,
		Result:    result,
		CreatedAt: created,
		UpdatedAt: updated,
	}
	if winner.Valid {
		game.Winner = winner.String
	}
	if endedBy.Valid {
		game.EndedBy = endedBy.String
	}
	if pending.Valid {
		color, err := parseColor(pending.String)
		if err == nil {
			game.PendingDrawOfferBy = &color
		}
	}

	return game, nil
}

func (s *PostgresStore) UpdateGame(ctx context.Context, game *Game) error {
	query := `
		UPDATE games
		SET current_fen = $2,
		    result = $3,
		    winner = $4,
		    ended_by = $5,
		    pending_draw_offer_by = $6,
		    updated_at = $7
		WHERE id = $1
	`
	ct, err := s.pool.Exec(
		ctx,
		query,
		game.ID,
		game.Board.ToFEN(),
		normalizeResult(game.Result),
		nullIfEmpty(game.Winner),
		nullIfEmpty(game.EndedBy),
		colorToNullableString(game.PendingDrawOfferBy),
		game.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *PostgresStore) UpdateGameWithMove(ctx context.Context, game *Game, move string) error {
	query := `
		WITH updated AS (
			UPDATE games
			SET current_fen = $2,
			    result = $3,
			    winner = $4,
			    ended_by = $5,
			    pending_draw_offer_by = $6,
			    updated_at = $7
			WHERE id = $1
			RETURNING id
		),
		next_ply AS (
			SELECT COALESCE(MAX(ply), 0) + 1 AS ply
			FROM moves
			WHERE game_id = $1
		),
		inserted AS (
			INSERT INTO moves (game_id, ply, move_number, color, uci, created_at)
			SELECT
				$1,
				next_ply.ply,
				(next_ply.ply + 1) / 2,
				CASE WHEN next_ply.ply % 2 = 1 THEN 'w' ELSE 'b' END,
				$8,
				$7
			FROM next_ply
			RETURNING 1
		)
		SELECT 1 FROM updated, inserted
	`
	err := s.pool.QueryRow(
		ctx,
		query,
		game.ID,
		game.Board.ToFEN(),
		normalizeResult(game.Result),
		nullIfEmpty(game.Winner),
		nullIfEmpty(game.EndedBy),
		colorToNullableString(game.PendingDrawOfferBy),
		game.UpdatedAt,
		move,
	).Scan(new(int))
	if err != nil {
		if err == pgx.ErrNoRows {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *PostgresStore) ListMoves(ctx context.Context, id string) ([]string, error) {
	query := `
		SELECT uci
		FROM moves
		WHERE game_id = $1
		ORDER BY ply ASC
	`
	rows, err := s.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	moves := []string{}
	for rows.Next() {
		var uci string
		if err := rows.Scan(&uci); err != nil {
			return nil, err
		}
		moves = append(moves, uci)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}

	if len(moves) == 0 {
		_, err := s.GetGame(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	return moves, nil
}

func nullIfEmpty(value string) interface{} {
	if value == "" {
		return nil
	}
	return value
}

func normalizeResult(result string) string {
	if result == "" {
		return "ongoing"
	}
	return result
}

func colorToNullableString(color *chess.Color) interface{} {
	if color == nil {
		return nil
	}
	return color.String()
}

func parseColor(value string) (chess.Color, error) {
	switch value {
	case "white":
		return chess.White, nil
	case "black":
		return chess.Black, nil
	default:
		return chess.White, fmt.Errorf("invalid color %q", value)
	}
}
