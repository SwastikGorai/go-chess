package store

import (
	"context"
	"errors"
)

var ErrNotFound = errors.New("game not found")

type GameStore interface {
	CreateGame(ctx context.Context, game *Game) error
	GetGame(ctx context.Context, id string) (*Game, error)
	UpdateGame(ctx context.Context, game *Game) error
	UpdateGameWithMove(ctx context.Context, game *Game, move string) error
	ListMoves(ctx context.Context, id string) ([]string, error)
}
