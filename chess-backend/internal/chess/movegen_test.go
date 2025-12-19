package chess

import "testing"

func TestLegalMoves_StartPosition_Count(t *testing.T) {
	b := NewBoard()
	m := b.LegalMoves()
	if len(m) != 20 {
		t.Fatalf("expected 20 legal moves from starting position, got %d", len(m))
	}
}

func TestLegalMoves_FilterKingExposure(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E2, NewPiece(Rook, White))
	b.setPiece(E8, NewPiece(Rook, Black))

	// rook moves that leave king exposed should not appear
	legal := b.LegalMoves()
	for _, mv := range legal {
		if mv.From == E2 && mv.To.File() != mv.From.File() {
			t.Fatalf("did not expect any rook move from E2; all expose check. Found: %s", mv)
		}
	}
}

func TestLegalMoves_PromotionOptions(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(A7, NewPiece(Pawn, White))

	legal := b.LegalMoves()

	promoCount := 0
	for _, mv := range legal {
		if mv.From == A7 && mv.To == A8 {
			promoCount++
		}
	}

	if promoCount != 4 {
		t.Fatalf("expected 4 promotion moves A7->A8, got %d", promoCount)
	}
}
