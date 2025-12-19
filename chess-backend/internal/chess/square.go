package chess

import (
	"fmt"
)

type Square int

const (
	A1 Square = iota
	A2
	A3
	A4
	A5
	A6
	A7
	A8
	B1
	B2
	B3
	B4
	B5
	B6
	B7
	B8
	C1
	C2
	C3
	C4
	C5
	C6
	C7
	C8
	D1
	D2
	D3
	D4
	D5
	D6
	D7
	D8
	E1
	E2
	E3
	E4
	E5
	E6
	E7
	E8
	F1
	F2
	F3
	F4
	F5
	F6
	F7
	F8
	G1
	G2
	G3
	G4
	G5
	G6
	G7
	G8
	H1
	H2
	H3
	H4
	H5
	H6
	H7
	H8
	NoSquare Square = -1
)

func (s Square) File() int {
	return int(s) / 8 //a-h
}

func (s Square) Rank() int {
	return int(s) % 8 // 1-8
}

func (s Square) isValid() bool {
	return s >= A1 && s <= H8
}

func (s Square) String() string {
	if s == NoSquare {
		return "-"
	}
	file := 'a' + rune(s.File())
	rank := '1' + rune(s.Rank())
	return string([]rune{file, rank})
}

func GetSquare(s string) (Square, error) {
	if len(s) != 2 {
		return NoSquare, fmt.Errorf("invalid Square Notation %s", s)
	}

	file := int(s[0] - 'a')
	rank := int(s[1] - '1')

	if file < 0 || file > 7 || rank < 0 || rank > 7 {
		return NoSquare, fmt.Errorf("invalid Square Notation %s", s)
	}

	return Square(file*8 + rank), nil
}
