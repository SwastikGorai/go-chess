package chess

type Board struct {
	squares   [64]*Piece
	turn      Color
	castling  CastlingRights
	enPassent Square
	halfMove  int
	fullMove  int

	// positionCounts: make(map[string]int), #TODO: implement threefold repition draw
}

type CastlingRights struct {
	WhiteKingside  bool
	WhiteQueenside bool
	BlackKingside  bool
	BlackQueenside bool
}

func newEmptyBoard(turn Color) *Board {
	b := &Board{
		squares:   [64]*Piece{},
		turn:      turn,
		enPassent: NoSquare,
	}
	return b
}

func NewBoard() *Board {
	b := &Board{
		turn: White,
		castling: CastlingRights{
			WhiteKingside:  true,
			WhiteQueenside: true,
			BlackKingside:  true,
			BlackQueenside: true,
		},
		enPassent: NoSquare,
		halfMove:  0,
		fullMove:  1,
	}
	b.setupStartingPosition()
	return b
}

func (b *Board) setupStartingPosition() {

	b.setPiece(A1, NewPiece(Rook, White))
	b.setPiece(B1, NewPiece(Knight, White))
	b.setPiece(C1, NewPiece(Bishop, White))
	b.setPiece(D1, NewPiece(Queen, White))
	b.setPiece(E1, NewPiece(King, White))
	b.setPiece(F1, NewPiece(Bishop, White))
	b.setPiece(G1, NewPiece(Knight, White))
	b.setPiece(H1, NewPiece(Rook, White))

	whitePawns := []Square{A2, B2, C2, D2, E2, F2, G2, H2}
	for _, sq := range whitePawns {
		b.setPiece(sq, NewPiece(Pawn, White))
	}

	b.setPiece(A8, NewPiece(Rook, Black))
	b.setPiece(B8, NewPiece(Knight, Black))
	b.setPiece(C8, NewPiece(Bishop, Black))
	b.setPiece(D8, NewPiece(Queen, Black))
	b.setPiece(E8, NewPiece(King, Black))
	b.setPiece(F8, NewPiece(Bishop, Black))
	b.setPiece(G8, NewPiece(Knight, Black))
	b.setPiece(H8, NewPiece(Rook, Black))

	blackPawns := []Square{A7, B7, C7, D7, E7, F7, G7, H7}
	for _, sq := range blackPawns {
		b.setPiece(sq, NewPiece(Pawn, Black))
	}
}

func (b *Board) setPiece(sq Square, p Piece) {
	b.squares[sq] = &p
}

func (b *Board) ClearSquare(sq Square) {
	if sq.isValid() {
		b.squares[sq] = nil
	}
}

func (b *Board) PieceAt(sq Square) *Piece {
	if !sq.isValid() {
		return nil
	}
	return b.squares[sq]
}

func (b *Board) Turn() Color {
	return b.turn
}

func (b *Board) IsEmpty(sq Square) bool {
	return b.squares[sq] == nil
}

func (b *Board) MakeMove(move Move) error {

	if err := ValidateMove(b, move); err != nil {
		return err
	}

	piece := b.PieceAt(move.From)
	if piece == nil {
		return ErrNoMoveablePiece
	}

	capturedPiece := b.PieceAt(move.To)
	prevEnPassent := b.enPassent

	if piece.Type == Pawn && move.To == prevEnPassent && capturedPiece == nil {
		capturedSq := Square(move.To.File()*8 + move.From.Rank())
		capturedPiece = b.PieceAt(capturedSq)
		b.ClearSquare(capturedSq)
	}

	if capturedPiece != nil && capturedPiece.Type == Rook {
		switch move.To {
		case A1:
			b.castling.WhiteQueenside = false
		case H1:
			b.castling.WhiteKingside = false
		case A8:
			b.castling.BlackQueenside = false
		case H8:
			b.castling.BlackKingside = false
		}
	}

	isCastle := piece.Type == King && (move.To.File()-move.From.File()) == 2 && move.From.Rank() == move.To.Rank()

	b.squares[move.To] = b.squares[move.From]
	b.ClearSquare(move.From)

	if isCastle {
		rank := move.To.Rank()
		if move.To.File() == 6 { // kingside
			rookFrom := Square(7*8 + rank) // H1/H8
			rookTo := Square(5*8 + rank)   // F1/F8
			rook := b.PieceAt(rookFrom)
			if rook == nil || rook.Type != Rook || rook.Color != piece.Color {
				return ErrIllegalMove
			}
			b.squares[rookTo] = b.squares[rookFrom]
			b.ClearSquare(rookFrom)
		}

		if move.To.File() == 2 { // queenside
			rookFrom := Square(0*8 + rank) // A1/A8
			rookTo := Square(3*8 + rank)   // D1/D8
			rook := b.PieceAt(rookFrom)
			if rook == nil || rook.Type != Rook || rook.Color != piece.Color {
				return ErrIllegalMove
			}
			b.squares[rookTo] = b.squares[rookFrom]
			b.ClearSquare(rookFrom)
		}
	}

	if move.isPromotion() {
		promotedPiece := NewPiece(move.Promotion, piece.Color)
		b.setPiece(move.To, promotedPiece)
	}

	b.UpdateCastlingRights(move, piece)

	if piece.Type == Pawn || capturedPiece != nil {
		b.halfMove = 0
	} else {
		b.halfMove++
	}

	if piece.Type == Pawn && abs(move.From.Rank()-move.To.Rank()) == 2 {
		direction := 1
		if piece.Color == Black {
			direction = -1
		}
		b.enPassent = Square(move.From.File()*8 + move.From.Rank() + direction)
	} else {
		b.enPassent = NoSquare
	}

	if b.turn == Black {
		b.fullMove++
	}

	b.turn = b.turn.Opposite()

	return nil
}

func (b *Board) UpdateCastlingRights(move Move, piece *Piece) {

	if piece.Type == King {
		if piece.Color == White {
			b.castling.WhiteKingside = false
			b.castling.WhiteQueenside = false
		} else {
			b.castling.BlackKingside = false
			b.castling.BlackQueenside = false
		}

		return
	}

	switch move.From {
	case A1:
		b.castling.WhiteQueenside = false
	case H1:
		b.castling.WhiteKingside = false
	case A8:
		b.castling.BlackQueenside = false
	case H8:
		b.castling.BlackKingside = false
	}
}

func (b *Board) Clone() *Board {
	nb := &Board{
		turn:      b.turn,
		castling:  b.castling,
		enPassent: b.enPassent,
		halfMove:  b.halfMove,
		fullMove:  b.fullMove,
	}

	for i := range b.squares {
		if b.squares[i] != nil {
			p := *b.squares[i]
			nb.squares[i] = &p
		}
	}
	return nb
}

func (b *Board) findKingSquare(color Color) Square {

	for sq := A1; sq <= H8; sq++ {
		p := b.PieceAt(sq)
		if p != nil && p.Type == King && p.Color == color {
			return sq
		}
	}
	return NoSquare
}

func (b *Board) InCheck(color Color) bool {
	kingSq := b.findKingSquare(color)
	if kingSq == NoSquare {
		return true // if king missing by any chance, treat as "Check"
	}
	return b.IsSquareAttacked(kingSq, color.Opposite())
}

func (b *Board) IsSquareAttacked(target Square, byColor Color) bool {
	if !target.isValid() {
		return false
	}

	tf := target.File()
	tr := target.Rank()

	// pawn
	{
		dir := 1
		if byColor == Black {
			dir = -1
		}
		attackersRank := tr - dir
		if attackersRank >= 0 && attackersRank <= 7 {
			for _, df := range []int{-1, 1} { //df delta file
				af := tf + df //af attackers file
				if af < 0 || af > 7 {
					continue
				}
				sq := Square(af*8 + attackersRank)
				p := b.PieceAt(sq)
				if p != nil && p.Color == byColor && p.Type == Pawn {
					return true
				}
			}
		}
	}

	//knight
	{
		offsets := [8][2]int{
			{2, 1}, {2, -1}, {-2, 1}, {-2, -1},
			{1, 2}, {1, -2}, {-1, 2}, {-1, -2},
		}
		for _, offset := range offsets {
			af := tf + offset[0]
			ar := tr + offset[1]
			if af < 0 || af > 7 || ar < 0 || ar > 7 {
				continue
			}
			sq := Square(af*8 + ar)
			p := b.PieceAt(sq)
			if p != nil && p.Color == byColor && p.Type == Knight {
				return true
			}
		}
	}

	//king
	{
		for df := -1; df <= 1; df++ {
			for dr := -1; dr <= 1; dr++ {
				if df == 0 && dr == 0 {
					continue
				}
				af := tf + df
				ar := tr + dr
				if af < 0 || af > 7 || ar < 0 || ar > 7 {
					continue
				}
				sq := Square(af*8 + ar)
				p := b.PieceAt(sq)
				if p != nil && p.Color == byColor && p.Type == King {
					return true
				}
			}
		}
	}

	// sliding attacks
	if b.isRayAttackedBySlider(target, byColor, StraightDirections, map[PieceType]bool{Rook: true, Queen: true}) {
		return true
	}
	if b.isRayAttackedBySlider(target, byColor, DiagonalDirections, map[PieceType]bool{Bishop: true, Queen: true}) {
		return true
	}

	return false
}

func (b *Board) isRayAttackedBySlider(target Square, byColor Color, dirs []Direction, allowed map[PieceType]bool) bool {

	tf := target.File()
	tr := target.Rank()
	for _, dir := range dirs {
		f := tf + dir.FileStep
		r := tr + dir.RankStep
		for f >= 0 && f <= 7 && r >= 0 && r <= 7 {
			sq := Square(f*8 + r)
			p := b.PieceAt(sq)
			if p != nil {
				if p.Color == byColor && allowed[p.Type] {
					return true
				}
				break // blocked by first piece encountered
			}
			f += dir.FileStep
			r += dir.RankStep
		}
	}

	return false
}

func (b *Board) applyMoveNoValidate(move Move) {

	piece := b.PieceAt(move.From)
	if piece == nil {
		return
	}
	captured := b.PieceAt(move.To)
	if piece.Type == Pawn && move.To == b.enPassent && captured == nil {
		capturedSq := Square(move.To.File()*8 + move.From.Rank())
		b.ClearSquare(capturedSq)
	}

	if piece.Type == King && abs(move.To.File()-move.From.File()) == 2 && move.From.Rank() == move.To.Rank() {
		rank := move.From.Rank()

		// kingside
		if move.To.File() == 6 {
			rookFrom := Square(7*8 + rank)
			rookTo := Square(5*8 + rank)
			b.squares[rookTo] = b.squares[rookFrom]
			b.ClearSquare(rookFrom)
		}

		// queenside
		if move.To.File() == 2 {
			rookFrom := Square(0*8 + rank)
			rookTo := Square(3*8 + rank)
			b.squares[rookTo] = b.squares[rookFrom]
			b.ClearSquare(rookFrom)
		}
	}

	b.squares[move.To] = b.squares[move.From]
	b.ClearSquare(move.From)

	if move.isPromotion() {
		promoted := NewPiece(move.Promotion, piece.Color)
		b.setPiece(move.To, promoted)
	}

	// king-safety simulation; not needed to update:
	// castling rights, halfmove/fullmove, enPassent, etc...
	// those will be for full engine state. Here, checking only whether would king be in check after 'this' move
}
