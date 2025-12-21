package api

import (
	"testing"

	"chess-backend/internal/chess"
)

func TestParseUCI(t *testing.T) {
	move, err := parseUCI("e2e4")
	if err != nil {
		t.Fatalf("parseUCI error: %v", err)
	}
	if move.From != chess.E2 || move.To != chess.E4 || move.Promotion != 0 {
		t.Fatalf("unexpected move: %+v", move)
	}

	promo, err := parseUCI("e7e8q")
	if err != nil {
		t.Fatalf("parseUCI promo error: %v", err)
	}
	if promo.Promotion != chess.Queen {
		t.Fatalf("expected queen promotion, got %+v", promo)
	}
}

func TestUCIFromMove(t *testing.T) {
	move := chess.NewMoveWithPromotion(chess.E7, chess.E8, chess.Queen)
	if got := uciFromMove(move); got != "e7e8q" {
		t.Fatalf("expected promotion uci, got %q", got)
	}
}
