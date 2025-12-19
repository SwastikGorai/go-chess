package chess

type Color int

const (
	White Color = iota
	Black
)

func (c Color) String() string {
	if c == White {
		return "white"
	}
	return "black"
}

func (c Color) Opposite() Color {
	if c == White {
		return Black
	}
	return White
}

type PieceType int

const (
	Pawn PieceType = iota
	Knight
	Bishop
	Rook
	King
	Queen
)

func (piecetype PieceType) String() string {
	switch piecetype {
	case Pawn:
		return "pawn"
	case Knight:
		return "knight"
	case Bishop:
		return "bishop"
	case Rook:
		return "rook"
	case King:
		return "king"
	case Queen:
		return "queen"
	default:
		return "unknwon"
	}
}

type Piece struct {
	Type  PieceType
	Color Color
}

func NewPiece(piece_type PieceType, color Color) Piece {
	return Piece{
		Type:  piece_type,
		Color: color,
	}
}

func (p Piece) String() string {
	return p.Color.String() + " " + p.Type.String()
}
