package chess

type BishopValidator struct{}

func (v *BishopValidator) IsLegalMove(board *Board, move Move) bool {
	return IsValidSlidingMove(board, move, DiagonalDirections)
}
