package chess

import "testing"

// Checkmate, Bblack to move:
// Black king: h8
// White queen: g7 (g7 attacks h8)
// White king:  f6 (defends g7 so KxQ is illegal)
// Simple mate.
func TestIsCheckmate_Simple(t *testing.T) {
	b := newEmptyBoard(Black)
	b.setPiece(H8, NewPiece(King, Black))
	b.setPiece(G7, NewPiece(Queen, White))
	b.setPiece(F6, NewPiece(King, White))

	if !b.IsCheck(Black) {
		t.Fatalf("expected black to be in check")
	}
	if !b.IsCheckmate(Black) {
		t.Fatalf("expected checkmate for black")
	}
	if b.IsStalemate(Black) {
		t.Fatalf("did not expect stalemate")
	}
}

// Stalemate, Black to move:
// Black king: h8
// White king: f7 (controls g8/g7)
// White queen: g6 (controls h7 and g7)
// Black is not in check, and has no legal moves.
func TestIsStalemate_Simple(t *testing.T) {
	b := newEmptyBoard(Black)
	b.setPiece(H8, NewPiece(King, Black))
	b.setPiece(F7, NewPiece(King, White))
	b.setPiece(G6, NewPiece(Queen, White))

	if b.IsCheck(Black) {
		t.Fatalf("expected black NOT to be in check")
	}
	if !b.IsStalemate(Black) {
		t.Fatalf("expected stalemate for black")
	}
	if b.IsCheckmate(Black) {
		t.Fatalf("did not expect checkmate")
	}
}
