package chess

import (
	"fmt"
)

type Move struct {
	From      Square
	To        Square
	Promotion PieceType
}

func NewMove(from, to Square) Move {
	return Move{
		From:      from,
		To:        to,
		Promotion: 0,
	}
}

func NewMoveWithPromotion(from, to Square, promotion_piece PieceType) Move {
	return Move{
		From:      from,
		To:        to,
		Promotion: promotion_piece,
	}
}

func (m Move) isPromotion() bool {
	return m.Promotion != 0
}

func (m Move) String() string {
	if m.Promotion != 0 {
		return fmt.Sprintf("%s%s=%s", m.From, m.To, m.Promotion)
	}
	return fmt.Sprintf("%s%s", m.From, m.To)
}
