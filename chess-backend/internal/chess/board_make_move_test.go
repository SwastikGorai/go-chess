package chess

import "testing"

func TestMakeMove_HalfMoveClock_ResetsAndIncrements(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.castling = CastlingRights{WhiteKingside: true, WhiteQueenside: true, BlackKingside: true, BlackQueenside: true}

	b.setPiece(B1, NewPiece(Knight, White))
	b.setPiece(G8, NewPiece(Knight, Black))
	b.setPiece(E2, NewPiece(Pawn, White))

	if err := b.MakeMove(NewMove(B1, A3)); err != nil {
		t.Fatalf("expected B1->A3 to be legal, got %v", err)
	}
	if b.halfMove != 1 {
		t.Fatalf("expected halfMove=1 after quiet move, got %d", b.halfMove)
	}

	if err := b.MakeMove(NewMove(G8, H6)); err != nil {
		t.Fatalf("expected G8->H6 to be legal, got %v", err)
	}
	if b.halfMove != 2 {
		t.Fatalf("expected halfMove=2 after second quiet move, got %d", b.halfMove)
	}

	if err := b.MakeMove(NewMove(E2, E4)); err != nil {
		t.Fatalf("expected E2->E4 to be legal, got %v", err)
	}
	if b.halfMove != 0 {
		t.Fatalf("expected halfMove reset to 0 after pawn move, got %d", b.halfMove)
	}
}

func TestMakeMove_EnPassent_ClearedAfterNonDoubleMove(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(E2, NewPiece(Pawn, White))
	b.setPiece(G8, NewPiece(Knight, Black))

	if err := b.MakeMove(NewMove(E2, E4)); err != nil {
		t.Fatalf("expected E2->E4 to be legal, got %v", err)
	}
	if b.enPassent != E3 {
		t.Fatalf("expected enPassent=%s after double pawn move, got %s", E3, b.enPassent)
	}

	if err := b.MakeMove(NewMove(G8, H6)); err != nil {
		t.Fatalf("expected G8->H6 to be legal, got %v", err)
	}
	if b.enPassent != NoSquare {
		t.Fatalf("expected enPassent cleared after non-double move, got %s", b.enPassent)
	}
}

func TestMakeMove_CaptureRook_DisablesCastlingRights(t *testing.T) {
	b := newEmptyBoard(Black)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.castling = CastlingRights{WhiteKingside: true, WhiteQueenside: true, BlackKingside: true, BlackQueenside: true}
	b.setPiece(A1, NewPiece(Rook, White))
	b.setPiece(A3, NewPiece(Queen, Black))

	if err := b.MakeMove(NewMove(A3, A1)); err != nil {
		t.Fatalf("expected A3xA1 to be legal, got %v", err)
	}

	if b.castling.WhiteQueenside {
		t.Fatalf("expected white queenside castling right to be cleared after rook capture")
	}
	if !b.castling.WhiteKingside {
		t.Fatalf("expected white kingside castling right to remain true")
	}
}
