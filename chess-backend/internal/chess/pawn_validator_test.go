package chess

import "testing"

func TestPawnMove_DoubleStep_SetsEnPassent(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(E2, NewPiece(Pawn, White))

	if err := b.MakeMove(NewMove(E2, E4)); err != nil {
		t.Fatalf("expected E2->E4 to be legal, got %v", err)
	}

	if b.enPassent != E3 {
		t.Fatalf("expected enPassent to be %s, got %s", E3, b.enPassent)
	}
}

func TestPawnMove_EnPassantCapture_RemovesCapturedPawn(t *testing.T) {
	b := newEmptyBoard(Black)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(D7, NewPiece(Pawn, Black))
	b.setPiece(E5, NewPiece(Pawn, White))

	if err := b.MakeMove(NewMove(D7, D5)); err != nil {
		t.Fatalf("expected D7->D5 to be legal, got %v", err)
	}
	if b.enPassent != D6 {
		t.Fatalf("expected enPassent to be %s, got %s", D6, b.enPassent)
	}

	if err := b.MakeMove(NewMove(E5, D6)); err != nil {
		t.Fatalf("expected E5xD6 en passant to be legal, got %v", err)
	}

	if b.PieceAt(D5) != nil {
		t.Fatalf("expected captured pawn on %s to be removed", D5)
	}
	p := b.PieceAt(D6)
	if p == nil || p.Type != Pawn || p.Color != White {
		t.Fatalf("expected white pawn to land on %s, got %v", D6, p)
	}
}

func TestPawnMove_EnPassantCapture_RequiresCapturablePawn(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(E5, NewPiece(Pawn, White))
	b.enPassent = D6

	if err := ValidateMove(b, NewMove(E5, D6)); err == nil {
		t.Fatalf("expected en passant without capturable pawn to be illegal")
	}
}

func TestPawnMove_PromotionRules(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(A7, NewPiece(Pawn, White))

	if err := ValidateMove(b, NewMove(A7, A8)); err == nil {
		t.Fatalf("expected A7->A8 without promotion to be illegal")
	}

	if err := ValidateMove(b, NewMoveWithPromotion(A7, A8, Queen)); err != nil {
		t.Fatalf("expected A7->A8=Q to be legal, got %v", err)
	}

	b2 := newEmptyBoard(White)
	b2.setPiece(A2, NewPiece(Pawn, White))
	if err := ValidateMove(b2, NewMoveWithPromotion(A2, A3, Queen)); err == nil {
		t.Fatalf("expected non-promotion move with promotion piece to be illegal")
	}
}
