package chess

func (b *Board) IsCheck(color Color) bool {
	return b.InCheck(color)
}

func (b *Board) IsCheckmate(color Color) bool {
	origTurn := b.turn
	b.turn = color

	defer func() { b.turn = origTurn }()

	if !b.InCheck(color) {
		return false
	}
	return len(b.LegalMoves()) == 0
}

func (b *Board) IsStalemate(color Color) bool {
	origTurn := b.turn
	b.turn = color

	defer func() { b.turn = origTurn }()

	if b.InCheck(color) {
		return false
	}
	return len(b.LegalMoves()) == 0
}
