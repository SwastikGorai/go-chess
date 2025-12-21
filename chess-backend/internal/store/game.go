package store

import (
	"time"

	"chess-backend/internal/chess"
	"github.com/google/uuid"
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

func NewGameID() (string, error) {
	return uuid.NewString(), nil
}
