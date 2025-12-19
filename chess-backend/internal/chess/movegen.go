package chess

func (b *Board) LegalMoves() []Move {

	pseudo := b.PseudoLegalMoves()

	legal := make([]Move, 0, len(pseudo))
	for _, m := range pseudo {
		if err := ValidateBasicMove(b, m); err != nil {
			continue
		}

		p := b.PieceAt(m.From)
		if p == nil {
			continue
		}
		v := getValidator(p.Type)
		if !v.IsLegalMove(b, m) {
			continue
		}

		sim := b.Clone()
		sim.applyMoveNoValidate(m)
		if sim.InCheck(p.Color) {
			continue
		}

		legal = append(legal, m)
	}

	return legal
}

func (b *Board) PseudoLegalMoves() []Move {
	moves := make([]Move, 0, 64)

	for from := A1; from <= H8; from++ {
		p := b.PieceAt(from)
		if p == nil || p.Color != b.turn {
			continue
		}

		switch p.Type {
		case Pawn:
			moves = append(moves, b.pawnPseudoMoves(from)...)
		case Knight:
			moves = append(moves, b.knightPseudoMoves(from)...)
		case Bishop:
			moves = append(moves, b.sliderPseudoMoves(from, DiagonalDirections)...)
		case Rook:
			moves = append(moves, b.sliderPseudoMoves(from, StraightDirections)...)
		case Queen:
			all := append(StraightDirections, DiagonalDirections...)
			moves = append(moves, b.sliderPseudoMoves(from, all)...)
		case King:
			moves = append(moves, b.kingPseudoMoves(from)...)
		}
	}

	return moves
}

func (b *Board) knightPseudoMoves(from Square) []Move {
	moves := []Move{}
	p := b.PieceAt(from)
	if p == nil {
		return moves
	}

	offsets := [][2]int{
		{2, 1}, {2, -1}, {-2, 1}, {-2, -1},
		{1, 2}, {1, -2}, {-1, 2}, {-1, -2},
	}

	f := from.File()
	r := from.Rank()

	for _, o := range offsets {
		nf := f + o[0] //newfile
		nr := r + o[1] //newrank
		if nf < 0 || nf > 7 || nr < 0 || nr > 7 {
			continue
		}
		to := Square(nf*8 + nr)
		dest := b.PieceAt(to)
		if dest == nil || dest.Color != p.Color {
			moves = append(moves, NewMove(from, to))
		}
	}

	return moves
}

func (b *Board) kingPseudoMoves(from Square) []Move {
	moves := []Move{}
	p := b.PieceAt(from)
	if p == nil {
		return moves
	}

	f := from.File()
	r := from.Rank()

	for df := -1; df <= 1; df++ {
		for dr := -1; dr <= 1; dr++ {
			if df == 0 && dr == 0 {
				continue
			}
			nf := f + df
			nr := r + dr
			if nf < 0 || nf > 7 || nr < 0 || nr > 7 {
				continue
			}
			to := Square(nf*8 + nr)
			dest := b.PieceAt(to)
			if dest == nil || dest.Color != p.Color {
				moves = append(moves, NewMove(from, to))
			}
		}
	}

	homeRank := 0
	if p.Color == Black {
		homeRank = 7
	}
	if from.File() == 4 && from.Rank() == homeRank { // E1/E8
		// E -> G (kingside), E -> C (queenside)
		moves = append(moves, NewMove(from, Square(6*8+homeRank)))
		moves = append(moves, NewMove(from, Square(2*8+homeRank)))
	}

	return moves
}

func (b *Board) sliderPseudoMoves(from Square, dirs []Direction) []Move {
	moves := []Move{}
	p := b.PieceAt(from)
	if p == nil {
		return moves
	}

	ff := from.File()
	fr := from.Rank()

	for _, d := range dirs {
		f := ff + d.FileStep
		r := fr + d.RankStep

		for f >= 0 && f <= 7 && r >= 0 && r <= 7 {
			to := Square(f*8 + r)
			dest := b.PieceAt(to)

			if dest == nil {
				moves = append(moves, NewMove(from, to))
			} else {
				// capture enemy then stop
				if dest.Color != p.Color {
					moves = append(moves, NewMove(from, to))
				}
				break
			}

			f += d.FileStep
			r += d.RankStep
		}
	}

	return moves
}

func (b *Board) pawnPseudoMoves(from Square) []Move {
	moves := []Move{}
	p := b.PieceAt(from)
	if p == nil {
		return moves
	}

	dir := 1
	startRank := 1
	promoRank := 7
	if p.Color == Black {
		dir = -1
		startRank = 6
		promoRank = 0
	}

	f := from.File()
	r := from.Rank()

	oneRank := r + dir
	if oneRank >= 0 && oneRank <= 7 {
		one := Square(f*8 + oneRank)

		// forward 1 if empty
		if b.IsEmpty(one) {
			if oneRank == promoRank {
				moves = append(moves,
					NewMoveWithPromotion(from, one, Queen),
					NewMoveWithPromotion(from, one, Rook),
					NewMoveWithPromotion(from, one, Bishop),
					NewMoveWithPromotion(from, one, Knight),
				)
			} else {
				moves = append(moves, NewMove(from, one))
			}

			// forward 2 from start if both empty
			twoRank := r + 2*dir
			if r == startRank && twoRank >= 0 && twoRank <= 7 {
				two := Square(f*8 + twoRank)
				if b.IsEmpty(two) {
					moves = append(moves, NewMove(from, two))
				}
			}
		}

		// captures (diagonals) + en passant
		for _, df := range []int{-1, 1} {
			nf := f + df
			if nf < 0 || nf > 7 {
				continue
			}
			to := Square(nf*8 + oneRank)
			dest := b.PieceAt(to)

			// normal capture
			if dest != nil && dest.Color != p.Color {
				if oneRank == promoRank {
					moves = append(moves,
						NewMoveWithPromotion(from, to, Queen),
						NewMoveWithPromotion(from, to, Rook),
						NewMoveWithPromotion(from, to, Bishop),
						NewMoveWithPromotion(from, to, Knight),
					)
				} else {
					moves = append(moves, NewMove(from, to))
				}
			}

			// en passant capture candidate
			if dest == nil && b.enPassent == to {
				moves = append(moves, NewMove(from, to)) // promotion not allowed with capture
			}
		}
	}

	return moves
}
