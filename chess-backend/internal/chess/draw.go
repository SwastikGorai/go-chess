package chess

func (b *Board) CanClaimFiftyMoveDraw() bool {
	return b.halfMove >= 100
}

func (b *Board) IsInsufficientMaterial() bool {
	type counts struct {
		pawns   int
		knights int
		bishops int
		rooks   int
		queens  int

		bishopsLight int
		bishopsDark  int
	}

	var w, bl counts

	for sq := A1; sq <= H8; sq++ {
		p := b.PieceAt(sq)
		if p == nil {
			continue
		}
		var c *counts
		if p.Color == White {
			c = &w
		} else {
			c = &bl
		}

		switch p.Type {
		case Pawn:
			c.pawns++
		case Knight:
			c.knights++
		case Bishop:
			c.bishops++
			if isLightSquare(sq) {
				c.bishopsLight++
			} else {
				c.bishopsDark++
			}
		case Rook:
			c.rooks++
		case Queen:
			c.queens++
		}
	}

	if w.pawns+w.rooks+w.queens > 0 || bl.pawns+bl.rooks+bl.queens > 0 {
		return false
	}

	// If either side has 2+ minor pieces (N/B) they *can* mate in some configurations
	// (like: K+BB vs K, K+BN vs K). Treat that as sufficient. Can refine more later
	if w.knights+w.bishops >= 2 {
		return false
	}
	if bl.knights+bl.bishops >= 2 {
		return false
	}

	// - K vs K
	// - K+N vs K
	// - K+B vs K
	// - K+B vs K+B (possibly same-colored bishops)
	// now each side has at most 1 minor. since rejected 2+ minors earlier (BB, BN)

	// If no minor -> K vs K
	if w.knights+w.bishops == 0 && bl.knights+bl.bishops == 0 {
		return true
	}

	// K+N vs K, K+B vs K
	if w.knights+w.bishops == 1 && bl.knights+bl.bishops == 0 {
		return true
	}
	if bl.knights+bl.bishops == 1 && w.knights+w.bishops == 0 {
		return true
	}

	// K+B vs K+B where both bishops are on same color squares
	if w.bishops == 1 && bl.bishops == 1 && w.knights == 0 && bl.knights == 0 {
		whiteLight := w.bishopsLight == 1
		blackLight := bl.bishopsLight == 1
		// same color bishops -> insufficient
		return whiteLight == blackLight
	}

	return false
}

func isLightSquare(sq Square) bool {
	// with 0-based file/rank: (file+rank)%2 == 0 is one color, == 1 the other r vice versa.
	// doesnt matter as long as consistent
	return (sq.File()+sq.Rank())%2 == 0
}
