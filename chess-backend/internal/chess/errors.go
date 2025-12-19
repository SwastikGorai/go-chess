package chess

import "errors"

var (
	ErrInvalidSquare    = errors.New("invalid square")
	ErrSameSquare       = errors.New("cannot move to same square")
	ErrNoMoveablePiece  = errors.New("no piece to move")
	ErrWrongTurn        = errors.New("not your turn")
	ErrCaptureOwnPiece  = errors.New("cannot capture own piece")
	ErrIllegalMove      = errors.New("illegal move for this piece")
	ErrPathBlocked      = errors.New("path is blocked")
	ErrInvalidPromotion = errors.New("invalid promotion")
)
