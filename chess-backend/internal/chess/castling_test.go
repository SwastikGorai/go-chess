package chess

import "testing"

func TestCastling_Kingside_MovesRook(t *testing.T) {
	b := newEmptyBoard(White)
	b.castling = CastlingRights{WhiteKingside: true, WhiteQueenside: true, BlackKingside: true, BlackQueenside: true}

	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(H1, NewPiece(Rook, White))

	if err := b.MakeMove(NewMove(E1, G1)); err != nil {
		t.Fatalf("expected O-O to be legal, got %v", err)
	}

	k := b.PieceAt(G1)
	if k == nil || k.Type != King || k.Color != White {
		t.Fatalf("expected white king on G1, got %v", k)
	}

	// Rook should be on F1, and H1 should be empty
	r := b.PieceAt(F1)
	if r == nil || r.Type != Rook || r.Color != White {
		t.Fatalf("expected white rook on F1, got %v", r)
	}
	if b.PieceAt(H1) != nil {
		t.Fatalf("expected H1 to be empty after castling")
	}
}

func TestCastling_Illegal_ThroughCheck(t *testing.T) {
	b := newEmptyBoard(White)
	b.castling = CastlingRights{WhiteKingside: true, WhiteQueenside: true, BlackKingside: true, BlackQueenside: true}

	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(H1, NewPiece(Rook, White))

	// Black rook attacks F1 (square king passes through) along F-file.
	b.setPiece(F8, NewPiece(Rook, Black))

	// Ensure path on F-file is open: clear squares F2..F7 by default since empty board.
	if err := ValidateMove(b, NewMove(E1, G1)); err == nil {
		t.Fatalf("expected castling through attacked square (F1) to be illegal")
	}
}

func TestCastling_Illegal_WhileInCheck(t *testing.T) {
	b := newEmptyBoard(White)
	b.castling = CastlingRights{WhiteKingside: true, WhiteQueenside: true, BlackKingside: true, BlackQueenside: true}

	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(H1, NewPiece(Rook, White))

	// Black rook gives check on E-file
	b.setPiece(E8, NewPiece(Rook, Black))

	if err := ValidateMove(b, NewMove(E1, G1)); err == nil {
		t.Fatalf("expected castling while in check to be illegal")
	}
}
