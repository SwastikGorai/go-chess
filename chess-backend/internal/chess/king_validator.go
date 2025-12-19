package chess

type KingValidator struct{}

func (v *KingValidator) IsLegalMove(board *Board, move Move) bool {
	fileDiff := abs(move.From.File() - move.To.File())
	rankDiff := abs(move.From.Rank() - move.To.Rank())

	if fileDiff <= 1 && rankDiff <= 1 {
		return true
	}

	// castling: king moves 2 files, stays in same rank
	if rankDiff == 0 && fileDiff == 2 {
		return v.isLegalCastle(board, move)
	}

	return false
}

func (v *KingValidator) isLegalCastle(board *Board, move Move) bool {
	king := board.PieceAt(move.From)
	if king == nil || king.Type != King {
		return false
	}

	if !board.IsEmpty(move.To) {
		return false
	}

	homeRank := 0
	if king.Color == Black {
		homeRank = 7
	}
	if move.From.File() != 4 || move.From.Rank() != homeRank { // must be E1/E8
		return false
	}

	if board.InCheck(king.Color) {
		return false
	}

	if move.To.File() == 6 && move.To.Rank() == homeRank {
		if king.Color == White && !board.castling.WhiteKingside {
			return false
		}
		if king.Color == Black && !board.castling.BlackKingside {
			return false
		}

		rookSq := Square(7*8 + homeRank)
		rook := board.PieceAt(rookSq)
		if rook == nil || rook.Type != Rook || rook.Color != king.Color {
			return false
		}

		fSq := Square(5*8 + homeRank)
		gSq := Square(6*8 + homeRank)
		if !board.IsEmpty(fSq) || !board.IsEmpty(gSq) {
			return false
		}

		if board.IsSquareAttacked(fSq, king.Color.Opposite()) {
			return false
		}
		if board.IsSquareAttacked(gSq, king.Color.Opposite()) {
			return false
		}

		return true
	}

	if move.To.File() == 2 && move.To.Rank() == homeRank {
		if king.Color == White && !board.castling.WhiteQueenside {
			return false
		}
		if king.Color == Black && !board.castling.BlackQueenside {
			return false
		}

		rookSq := Square(0*8 + homeRank)
		rook := board.PieceAt(rookSq)
		if rook == nil || rook.Type != Rook || rook.Color != king.Color {
			return false
		}

		dSq := Square(3*8 + homeRank)
		cSq := Square(2*8 + homeRank)
		bSq := Square(1*8 + homeRank)
		if !board.IsEmpty(dSq) || !board.IsEmpty(cSq) || !board.IsEmpty(bSq) {
			return false
		}

		if board.IsSquareAttacked(dSq, king.Color.Opposite()) {
			return false
		}
		if board.IsSquareAttacked(cSq, king.Color.Opposite()) {
			return false
		}

		return true
	}

	return false
}
