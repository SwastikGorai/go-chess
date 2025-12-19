package chess

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

type MoveValidator interface {
	IsLegalMove(board *Board, move Move) bool
}

func ValidateMove(board *Board, move Move) error {
	if err := ValidateBasicMove(board, move); err != nil {
		return err
	}

	piece := board.PieceAt(move.From)
	if piece == nil {
		return ErrNoMoveablePiece
	}

	validator := getValidator(piece.Type)

	if !validator.IsLegalMove(board, move) {
		return ErrIllegalMove
	}

	sim := board.Clone()
	sim.applyMoveNoValidate(move)
	if sim.InCheck(piece.Color) {
		return ErrIllegalMove
	}

	return nil

}

func ValidateBasicMove(board *Board, move Move) error {
	if !move.From.isValid() || !move.To.isValid() {
		return ErrInvalidSquare
	}
	if move.From == move.To {
		return ErrSameSquare
	}

	piece := board.PieceAt(move.From)
	if piece == nil {
		return ErrNoMoveablePiece
	}
	if piece.Color != board.Turn() {
		return ErrWrongTurn
	}

	destPiece := board.PieceAt(move.To)
	if destPiece != nil && destPiece.Color == piece.Color {
		return ErrCaptureOwnPiece
	}

	return nil
}

func getValidator(piecetype PieceType) MoveValidator {
	switch piecetype {
	case Pawn:
		return &PawnValidator{}
	case Knight:
		return &KnightValidator{}
	case Bishop:
		return &BishopValidator{}
	case Rook:
		return &RookValidator{}
	case Queen:
		return &QueenValidator{}
	case King:
		return &KingValidator{}
	default:
		return &NullValidator{}
	}
}

type NullValidator struct{}

func (n *NullValidator) IsLegalMove(board *Board, move Move) bool {
	return false
}
