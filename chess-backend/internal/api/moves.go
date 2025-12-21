package api

import (
	"fmt"
	"strings"

	"chess-backend/internal/chess"
)

func parseUCI(input string) (chess.Move, error) {
	input = strings.ToLower(strings.TrimSpace(input))
	if len(input) != 4 && len(input) != 5 {
		return chess.Move{}, fmt.Errorf("uci must be 4 or 5 chars")
	}

	from, err := chess.GetSquare(input[0:2])
	if err != nil {
		return chess.Move{}, err
	}
	to, err := chess.GetSquare(input[2:4])
	if err != nil {
		return chess.Move{}, err
	}

	if len(input) == 4 {
		return chess.NewMove(from, to), nil
	}

	promo, err := parsePromotion(input[4])
	if err != nil {
		return chess.Move{}, err
	}
	return chess.NewMoveWithPromotion(from, to, promo), nil
}

func parsePromotion(b byte) (chess.PieceType, error) {
	switch strings.ToLower(string([]byte{b})) {
	case "q":
		return chess.Queen, nil
	case "r":
		return chess.Rook, nil
	case "b":
		return chess.Bishop, nil
	case "n":
		return chess.Knight, nil
	default:
		return 0, fmt.Errorf("invalid promotion piece %q", b)
	}
}

func uciFromMove(move chess.Move) string {
	if move.Promotion == 0 {
		return move.From.String() + move.To.String()
	}
	return move.From.String() + move.To.String() + promotionSuffix(move.Promotion)
}

func promotionSuffix(p chess.PieceType) string {
	switch p {
	case chess.Queen:
		return "q"
	case chess.Rook:
		return "r"
	case chess.Bishop:
		return "b"
	case chess.Knight:
		return "n"
	default:
		return ""
	}
}
