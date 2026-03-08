package chess

import "math/rand"

type zobristTables struct {
	pieceSq [2][6][64]uint64
	side    uint64
	castle  [4]uint64
	epFile  [8]uint64
}

var zb zobristTables

func init() {
	r := rand.New(rand.NewSource(0xC0FFEE))

	next := func() uint64 {
		hi := uint64(r.Uint32())
		lo := uint64(r.Uint32())
		return (hi << 32) | lo
	}

	for c := 0; c < 2; c++ {
		for pt := 0; pt < 6; pt++ {
			for sq := 0; sq < 64; sq++ {
				zb.pieceSq[c][pt][sq] = next()
			}
		}
	}

	zb.side = next()
	for i := 0; i < 4; i++ {
		zb.castle[i] = next()
	}
	for f := 0; f < 8; f++ {
		zb.epFile[f] = next()
	}
}

func zobristPieceIndex(pt PieceType) (int, bool) {
	switch pt {
	case Pawn:
		return 0, true
	case Knight:
		return 1, true
	case Bishop:
		return 2, true
	case Rook:
		return 3, true
	case Queen:
		return 4, true
	case King:
		return 5, true
	default:
		return 0, false
	}
}

func zobristColorIndex(c Color) int {
	if c == White {
		return 0
	}
	return 1
}

func (b *Board) computeZobrist() uint64 {
	var h uint64

	for sq := A1; sq <= H8; sq++ {
		p := b.PieceAt(sq)
		if p == nil {
			continue
		}
		pi, ok := zobristPieceIndex(p.Type)
		if !ok {
			continue
		}
		ci := zobristColorIndex(p.Color)
		h ^= zb.pieceSq[ci][pi][int(sq)]
	}

	if b.turn == Black {
		h ^= zb.side
	}

	if b.castling.WhiteKingside {
		h ^= zb.castle[0]
	}
	if b.castling.WhiteQueenside {
		h ^= zb.castle[1]
	}
	if b.castling.BlackKingside {
		h ^= zb.castle[2]
	}
	if b.castling.BlackQueenside {
		h ^= zb.castle[3]
	}

	if b.enPassent != NoSquare && b.enPassentCapturableForHash() {
		h ^= zb.epFile[b.enPassent.File()]
	}

	return h
}

func (b *Board) enPassentCapturableForHash() bool {
	ep := b.enPassent
	if ep == NoSquare {
		return false
	}

	if b.turn == White {
		r := ep.Rank() - 1
		if r < 0 {
			return false
		}
		for _, df := range []int{-1, 1} {
			f := ep.File() + df
			if f < 0 || f > 7 {
				continue
			}
			sq := Square(f*8 + r)
			p := b.PieceAt(sq)
			if p != nil && p.Color == White && p.Type == Pawn {
				return true
			}
		}
		return false
	}

	r := ep.Rank() + 1
	if r > 7 {
		return false
	}
	for _, df := range []int{-1, 1} {
		f := ep.File() + df
		if f < 0 || f > 7 {
			continue
		}
		sq := Square(f*8 + r)
		p := b.PieceAt(sq)
		if p != nil && p.Color == Black && p.Type == Pawn {
			return true
		}
	}
	return false
}

func (b *Board) Zobrist() uint64 {
	return b.zkey
}
