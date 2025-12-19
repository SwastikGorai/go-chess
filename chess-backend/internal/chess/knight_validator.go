package chess

type KnightValidator struct{}

func (v *KnightValidator) IsLegalMove(board *Board, move Move) bool {
	fileDiff := abs(move.To.File() - move.From.File())
	rankDiff := abs(move.To.Rank() - move.From.Rank())

	return (fileDiff == 2 && rankDiff == 1) || (fileDiff == 1 && rankDiff == 2)
}
