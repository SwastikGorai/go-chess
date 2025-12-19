package chess

type Direction struct {
	FileStep int
	RankStep int
}

var (
	DiagonalDirections = []Direction{
		{1, 1},
		{1, -1},
		{-1, 1},
		{-1, -1},
	}

	StraightDirections = []Direction{
		{0, 1},
		{0, -1},
		{1, 0},
		{-1, 0},
	}
)

func isPathClear(board *Board, from, to Square, dir Direction) bool {
	currentFile := from.File() + dir.FileStep
	currentRank := from.Rank() + dir.RankStep

	for {
		if currentFile < 0 || currentFile > 7 || currentRank < 0 || currentRank > 7 {
			return false
		}

		current := Square(currentFile*8 + currentRank)

		if current == to {
			return true
		}

		if !board.IsEmpty(current) {
			return false
		}

		currentFile += dir.FileStep
		currentRank += dir.RankStep

	}
}

func isMovingInDirection(move Move, dir Direction) bool {
	fileDiff := move.To.File() - move.From.File()
	rankDiff := move.To.Rank() - move.From.Rank()

	if dir.FileStep == 0 {
		return fileDiff == 0 && rankDiff*dir.RankStep > 0
	}

	if dir.RankStep == 0 {
		return rankDiff == 0 && fileDiff*dir.FileStep > 0
	}

	if abs(fileDiff) != abs(rankDiff) {
		return false
	}

	return fileDiff*dir.FileStep > 0 && rankDiff*dir.RankStep > 0

}

func IsValidSlidingMove(board *Board, move Move, dirs []Direction) bool {
	for _, dir := range dirs {
		if isMovingInDirection(move, dir) {
			return isPathClear(board, move.From, move.To, dir)
		}
	}
	return false
}
