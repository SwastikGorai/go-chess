package chess

import "testing"

func TestZobrist_RoundTripFENStable(t *testing.T) {
	b := NewBoard()
	h1 := b.Zobrist()

	fen := b.ToFEN()
	b2, err := LoadFEN(fen)
	if err != nil {
		t.Fatalf("LoadFEN: %v", err)
	}
	h2 := b2.Zobrist()

	if h1 != h2 {
		t.Fatalf("zobrist mismatch after FEN roundtrip: %d vs %d", h1, h2)
	}
}

func TestZobrist_ChangesAfterMove(t *testing.T) {
	b := NewBoard()
	h1 := b.Zobrist()

	if err := b.MakeMove(NewMove(E2, E4)); err != nil {
		t.Fatalf("move: %v", err)
	}
	h2 := b.Zobrist()

	if h1 == h2 {
		t.Fatalf("expected zobrist to change after a move")
	}
}
