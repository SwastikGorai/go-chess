package chess

type RookValidator struct{}

func (v *RookValidator) IsLegalMove(board *Board, move Move) bool {
	return IsValidSlidingMove(board, move, StraightDirections)
}
