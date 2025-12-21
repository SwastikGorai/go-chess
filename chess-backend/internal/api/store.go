package api

import (
	"sync"
	"time"

	"github.com/google/uuid"

	"chess-backend/internal/chess"
)

type Game struct {
	ID                 string
	Board              *chess.Board
	StartFEN           string
	Moves              []string
	PendingDrawOfferBy *chess.Color
	CreatedAt          time.Time
	UpdatedAt          time.Time
	Result             string
	Winner             string
	EndedBy            string
}

type Store struct {
	mu    sync.RWMutex
	games map[string]*Game
}

func NewStore() *Store {
	return &Store{
		// TODO: replace with persistent storage boundary once user/game records exist.
		games: make(map[string]*Game),
	}
}

func newGameID() (string, error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	return id.String(), nil
}
