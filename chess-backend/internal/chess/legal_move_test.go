package chess

import "testing"

func TestValidateMove_RejectsMoveThatLeavesKingInCheck(t *testing.T) {
	b := newEmptyBoard(White)

	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E2, NewPiece(Rook, White))
	b.setPiece(E8, NewPiece(Rook, Black))

	// move the blocking rook away -> exposes check on E-file -> should be illegal
	if err := ValidateMove(b, NewMove(E2, D2)); err == nil {
		t.Fatalf("expected move exposing king to check to be illegal")
	}
}

func TestValidateMove_RejectsKingMoveIntoCheck(t *testing.T) {
	b := newEmptyBoard(White)

	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(Rook, Black))

	// E2 is attacked by rook on E8 (clear file)
	if err := ValidateMove(b, NewMove(E1, E2)); err == nil {
		t.Fatalf("expected king move into check to be illegal")
	}
}
