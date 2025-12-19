package chess

type PawnValidator struct{}

func (v *PawnValidator) IsLegalMove(board *Board, move Move) bool {
	piece := board.PieceAt(move.From)
	if piece == nil {
		return false
	}

	direction := 1
	startRank := 1
	if piece.Color == Black {
		direction = -1
		startRank = 6
	}

	fileDiff := move.To.File() - move.From.File()
	rankDiff := move.To.Rank() - move.From.Rank()

	if rankDiff*direction <= 0 {
		return false
	}

	destPiece := board.PieceAt(move.To)

	// moving forward (not capturing)
	if fileDiff == 0 {
		if destPiece != nil {
			return false
		}

		// single square forward
		if rankDiff == direction {
			return v.checkPromotion(move, piece.Color)
		}

		// Two squares forward from starting position
		if rankDiff == 2*direction && move.From.Rank() == startRank {
			middleSquare := Square(move.From.File()*8 + move.From.Rank() + direction)
			return board.IsEmpty(middleSquare) && v.checkPromotion(move, piece.Color)
		}

		return false
	}

	if abs(fileDiff) == 1 && rankDiff == direction {
		// Normal capture
		if destPiece != nil && destPiece.Color != piece.Color {
			return v.checkPromotion(move, piece.Color)
		}

		// en passant capture (destination must be empty, captured pawn must exist).
		if destPiece == nil && board.enPassent == move.To && move.Promotion == 0 {
			capturedSq := Square(move.To.File()*8 + move.From.Rank())
			capturedPiece := board.PieceAt(capturedSq)
			return capturedPiece != nil &&
				capturedPiece.Type == Pawn &&
				capturedPiece.Color != piece.Color
		}

		return false
	}
	return false
}

func (v *PawnValidator) checkPromotion(move Move, color Color) bool {

	promotionRank := 7
	if color == Black {
		promotionRank = 0
	}

	if move.To.Rank() == promotionRank {
		return move.Promotion == Queen ||
			move.Promotion == Rook ||
			move.Promotion == Bishop ||
			move.Promotion == Knight
	}

	return move.Promotion == 0
}
