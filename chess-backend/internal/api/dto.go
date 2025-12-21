package api

import "time"

type CreateGameRequest struct {
	Fen string `json:"fen"`
}

type MoveRequest struct {
	UCI string `json:"uci"`
}

type ResignRequest struct {
	Color string `json:"color"`
}

type OfferDrawRequest struct {
	Color string `json:"color"`
}

type AcceptDrawRequest struct {
	Color string `json:"color"`
}

type Flags struct {
	InCheck       bool   `json:"inCheck"`
	Checkmate     bool   `json:"checkmate"`
	Stalemate     bool   `json:"stalemate"`
	Draw          bool   `json:"draw"`
	DrawReason    string `json:"drawReason,omitempty"`
	DrawClaimable bool   `json:"drawClaimable,omitempty"`
}

type Meta struct {
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
	StartFEN  string    `json:"startFEN"`
}

type GameResponse struct {
	ID       string `json:"id"`
	FEN      string `json:"fen"`
	Turn     string `json:"turn"`
	Result   string `json:"result"`
	Winner   string `json:"winner,omitempty"`
	EndedBy  string `json:"endedBy,omitempty"`
	Flags    Flags  `json:"flags"`
	Halfmove int    `json:"halfmove"`
	Fullmove int    `json:"fullmove"`
	Meta     Meta   `json:"meta"`
}

type MoveResponse struct {
	FEN      string `json:"fen"`
	Turn     string `json:"turn"`
	Result   string `json:"result"`
	Winner   string `json:"winner,omitempty"`
	EndedBy  string `json:"endedBy,omitempty"`
	Flags    Flags  `json:"flags"`
	Halfmove int    `json:"halfmove"`
	Fullmove int    `json:"fullmove"`
}

type StatusResponse struct {
	Result string `json:"result"`
	Flags  Flags  `json:"flags"`
}

type HistoryResponse struct {
	ID    string   `json:"id"`
	Moves []string `json:"moves"`
}

type LegalMovesResponse struct {
	Moves []string `json:"moves"`
}

type ResignResponse struct {
	Result  string `json:"result"`
	Winner  string `json:"winner"`
	EndedBy string `json:"endedBy"`
	Flags   Flags  `json:"flags"`
}

type OfferDrawResponse struct {
	Offer string `json:"offer"`
}

type AcceptDrawResponse struct {
	Result  string `json:"result"`
	Winner  string `json:"winner"`
	EndedBy string `json:"endedBy"`
	Flags   Flags  `json:"flags"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}
