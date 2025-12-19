package chess

import "testing"

func TestCanClaimFiftyMoveDraw(t *testing.T) {
	b := newEmptyBoard(White)
	b.halfMove = 99
	if b.CanClaimFiftyMoveDraw() {
		t.Fatalf("expected not claimable at 99 plies")
	}

	b.halfMove = 100
	if !b.CanClaimFiftyMoveDraw() {
		t.Fatalf("expected claimable at 100 plies")
	}
}

func TestIsInsufficientMaterial_KvK(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))

	if !b.IsInsufficientMaterial() {
		t.Fatalf("expected K vs K to be insufficient material")
	}
}

func TestIsInsufficientMaterial_KNvK(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(G1, NewPiece(Knight, White))

	if !b.IsInsufficientMaterial() {
		t.Fatalf("expected K+N vs K to be insufficient material")
	}
}

func TestIsInsufficientMaterial_KBvK(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(C1, NewPiece(Bishop, White))

	if !b.IsInsufficientMaterial() {
		t.Fatalf("expected K+B vs K to be insufficient material")
	}
}

func TestIsInsufficientMaterial_KBvKB_SameColorBishops(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))

	// C1: file=2, rank=0 => even => "light" by our function
	b.setPiece(C1, NewPiece(Bishop, White))

	// F4: file=5, rank=3 => 8 even => also "light"
	b.setPiece(F4, NewPiece(Bishop, Black))

	if !b.IsInsufficientMaterial() {
		t.Fatalf("expected same-colored bishops to be insufficient material")
	}
}

func TestIsInsufficientMaterial_NotInsufficientWithTwoMinors(t *testing.T) {
	b := newEmptyBoard(White)
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(E8, NewPiece(King, Black))

	b.setPiece(C1, NewPiece(Bishop, White))
	b.setPiece(G1, NewPiece(Knight, White))

	if b.IsInsufficientMaterial() {
		t.Fatalf("expected K+BN vs K to NOT be insufficient material")
	}
}
