package chess

import (
	"fmt"
	"strconv"
	"strings"
)

func (b *Board) ToFEN() string {
	var sb strings.Builder

	// piece placement: ranks 8..1, files a..h
	for rank := 7; rank >= 0; rank-- {
		empty := 0
		for file := 0; file < 8; file++ {
			sq := Square(file*8 + rank)
			p := b.PieceAt(sq)
			if p == nil {
				empty++
				continue
			}
			if empty > 0 {
				sb.WriteByte(byte('0' + empty))
				empty = 0
			}
			sb.WriteByte(fenPieceChar(*p))
		}
		if empty > 0 {
			sb.WriteByte(byte('0' + empty))
		}
		if rank != 0 {
			sb.WriteByte('/')
		}
	}

	// side to move
	sb.WriteByte(' ')
	if b.turn == White {
		sb.WriteByte('w')
	} else {
		sb.WriteByte('b')
	}

	// castling rights
	sb.WriteByte(' ')
	castle := fenCastlingString(b.castling)
	sb.WriteString(castle)

	// en passant
	sb.WriteByte(' ')
	if b.enPassent == NoSquare {
		sb.WriteByte('-')
	} else {
		sb.WriteString(b.enPassent.String())
	}

	// halfmove / fullmove
	sb.WriteByte(' ')
	sb.WriteString(intToString(b.halfMove))
	sb.WriteByte(' ')
	sb.WriteString(intToString(b.fullMove))

	return sb.String()
}

func fenPieceChar(p Piece) byte {
	var c byte
	switch p.Type {
	case Pawn:
		c = 'p'
	case Knight:
		c = 'n'
	case Bishop:
		c = 'b'
	case Rook:
		c = 'r'
	case Queen:
		c = 'q'
	case King:
		c = 'k'
	default:
		c = '?'
	}
	if p.Color == White {
		return c - 32
	}
	return c
}

func fenCastlingString(cr CastlingRights) string {
	out := make([]byte, 0, 4)
	if cr.WhiteKingside {
		out = append(out, 'K')
	}
	if cr.WhiteQueenside {
		out = append(out, 'Q')
	}
	if cr.BlackKingside {
		out = append(out, 'k')
	}
	if cr.BlackQueenside {
		out = append(out, 'q')
	}
	if len(out) == 0 {
		return "-"
	}
	return string(out)
}

func intToString(n int) string {
	if n == 0 {
		return "0"
	}
	neg := false
	if n < 0 {
		neg = true
		n = -n
	}
	var buf [32]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + (n % 10))
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}


func LoadFEN(fen string) (*Board, error) {

	fields := strings.Fields(fen)
	if len(fields) != 6 {
		return nil, fmt.Errorf("invalid FEN: expected 6 fields, got %d", len(fields))
	}

	placement := fields[0]
	active := fields[1]
	castling := fields[2]
	enPassant := fields[3]
	halfmoveStr := fields[4]
	fullmoveStr := fields[5]

	b := &Board{
		turn:      White,
		castling:  CastlingRights{},
		enPassent: NoSquare,
		halfMove:  0,
		fullMove:  1,
	}

	// piece placement
	ranks := strings.Split(placement, "/")
	if len(ranks) != 8 {
		return nil, fmt.Errorf("invalid FEN: expected 8 ranks, got %d", len(ranks))
	}

	for rankIdx := range 8 {
		rankStr := ranks[rankIdx]
		boardRank := 7 - rankIdx

		file := 0
		for i := 0; i < len(rankStr); i++ {
			ch := rankStr[i]

			// digit = empty squares
			if ch >= '1' && ch <= '8' {
				file += int(ch - '0')
				if file > 8 {
					return nil, fmt.Errorf("invalid FEN: rank %d overflows file count", 8-rankIdx)
				}
				continue
			}

			// piece letter
			if file >= 8 {
				return nil, fmt.Errorf("invalid FEN: too many files in rank %d", 8-rankIdx)
			}
			p, err := pieceFromFENChar(ch)
			if err != nil {
				return nil, fmt.Errorf("invalid FEN: %v", err)
			}
			sq := Square(file*8 + boardRank)
			b.setPiece(sq, p)
			file++
		}

		if file != 8 {
			return nil, fmt.Errorf("invalid FEN: rank %d has %d files (expected 8)", 8-rankIdx, file)
		}
	}

	// active color
	switch active {
	case "w":
		b.turn = White
	case "b":
		b.turn = Black
	default:
		return nil, fmt.Errorf("invalid FEN: active color must be w or b, got %q", active)
	}

	// castling rights
	if castling != "-" {
		for i := 0; i < len(castling); i++ {
			switch castling[i] {
			case 'K':
				b.castling.WhiteKingside = true
			case 'Q':
				b.castling.WhiteQueenside = true
			case 'k':
				b.castling.BlackKingside = true
			case 'q':
				b.castling.BlackQueenside = true
			default:
				return nil, fmt.Errorf("invalid FEN: unknown castling char %q", castling[i])
			}
		}
	}

	// en passant
	if enPassant == "-" {
		b.enPassent = NoSquare
	} else {
		sq, err := GetSquare(enPassant)
		if err != nil {
			return nil, fmt.Errorf("invalid FEN: bad en passant square %q", enPassant)
		}
		b.enPassent = sq
	}

	// halfmove/fullmove
	half, err := strconv.Atoi(halfmoveStr)
	if err != nil || half < 0 {
		return nil, fmt.Errorf("invalid FEN: bad halfmove clock %q", halfmoveStr)
	}
	full, err := strconv.Atoi(fullmoveStr)
	if err != nil || full <= 0 {
		return nil, fmt.Errorf("invalid FEN: bad fullmove number %q", fullmoveStr)
	}
	b.halfMove = half
	b.fullMove = full

	return b, nil
}

func pieceFromFENChar(ch byte) (Piece, error) {
	color := Black
	if ch >= 'A' && ch <= 'Z' {
		color = White
		ch = ch + 32
	}

	var pt PieceType
	switch ch {
	case 'p':
		pt = Pawn
	case 'n':
		pt = Knight
	case 'b':
		pt = Bishop
	case 'r':
		pt = Rook
	case 'q':
		pt = Queen
	case 'k':
		pt = King
	default:
		return Piece{}, fmt.Errorf("unknown piece char %q", ch)
	}

	return NewPiece(pt, color), nil
}
