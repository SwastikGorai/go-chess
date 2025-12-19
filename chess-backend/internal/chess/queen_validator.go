package chess

type QueenValidator struct{}

func (v *QueenValidator) IsLegalMove(board *Board, move Move) bool {
	allDirections := append(StraightDirections, DiagonalDirections...)
	return IsValidSlidingMove(board, move, allDirections)
}
