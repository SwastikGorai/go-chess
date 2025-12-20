package chess

import "testing"

func TestToFEN_StartingPosition(t *testing.T) {
	b := NewBoard()
	got := b.ToFEN()

	// new board starting FEN
	want := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	if got != want {
		t.Fatalf("unexpected FEN\n got: %s\nwant: %s", got, want)
	}
}


func TestLoadFEN_StartingPosition_RoundTrip(t *testing.T) {
	fen := "rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1"
	b, err := LoadFEN(fen)
	if err != nil {
		t.Fatalf("LoadFEN error: %v", err)
	}

	got := b.ToFEN()
	if got != fen {
		t.Fatalf("round-trip mismatch\n got: %s\nwant: %s", got, fen)
	}
}

//  black to move
// only white kingside castling right
// en passant target e3
// halfmove=12 fullmove=7
func TestLoadFEN_CustomState(t *testing.T) {

	fen := "8/8/8/8/8/8/4P3/4K2R b K e3 12 7"

	b, err := LoadFEN(fen)
	if err != nil {
		t.Fatalf("LoadFEN error: %v", err)
	}

	if b.turn != Black {
		t.Fatalf("expected turn=Black, got %v", b.turn)
	}
	if !b.castling.WhiteKingside || b.castling.WhiteQueenside || b.castling.BlackKingside || b.castling.BlackQueenside {
		t.Fatalf("unexpected castling rights: %+v", b.castling)
	}
	if b.enPassent != E3 {
		t.Fatalf("expected enPassent=E3, got %s", b.enPassent)
	}
	if b.halfMove != 12 || b.fullMove != 7 {
		t.Fatalf("expected halfMove=12 fullMove=7, got halfMove=%d fullMove=%d", b.halfMove, b.fullMove)
	}

	// round-trip matches (exact same FEN should come out)
	got := b.ToFEN()
	if got != fen {
		t.Fatalf("round-trip mismatch\n got: %s\nwant: %s", got, fen)
	}
}
