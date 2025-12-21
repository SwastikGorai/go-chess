package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"chess-backend/internal/chess"

	"github.com/gin-gonic/gin"
)

type Handlers struct {
	store *Store
}

func NewHandlers(store *Store) *Handlers {
	return &Handlers{store: store}
}

func (h *Handlers) CreateGame(c *gin.Context) {
	var req CreateGameRequest
	if err := c.ShouldBindJSON(&req); err != nil && !isEmptyBody(err) {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	var board *chess.Board
	var err error
	if strings.TrimSpace(req.Fen) == "" {
		board = chess.NewBoard()
	} else {
		board, err = chess.LoadFEN(req.Fen)
		if err != nil {
			writeError(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	id, err := newGameID()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to create game id")
		return
	}

	now := time.Now().UTC()
	game := &Game{
		ID:        id,
		Board:     board,
		StartFEN:  board.ToFEN(),
		Moves:     []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}

	h.store.mu.Lock()
	h.store.games[id] = game
	response := buildGameResponse(game)
	h.store.mu.Unlock()

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) GetGame(c *gin.Context) {
	id := c.Param("id")

	h.store.mu.RLock()
	game, ok := h.store.games[id]
	if !ok {
		h.store.mu.RUnlock()
		writeError(c, http.StatusNotFound, "game not found")
		return
	}
	response := buildGameResponse(game)
	h.store.mu.RUnlock()

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) LegalMoves(c *gin.Context) {
	id := c.Param("id")
	fromFilter := strings.ToLower(strings.TrimSpace(c.Query("from")))

	var fromSquare *chess.Square
	if fromFilter != "" {
		sq, err := chess.GetSquare(fromFilter)
		if err != nil {
			writeError(c, http.StatusBadRequest, "invalid from square")
			return
		}
		fromSquare = &sq
	}

	h.store.mu.RLock()
	game, ok := h.store.games[id]
	if !ok {
		h.store.mu.RUnlock()
		writeError(c, http.StatusNotFound, "game not found")
		return
	}

	legal := game.Board.LegalMoves()
	moves := make([]string, 0, len(legal))
	for _, m := range legal {
		if fromSquare != nil && m.From != *fromSquare {
			continue
		}
		moves = append(moves, uciFromMove(m))
	}
	h.store.mu.RUnlock()

	c.JSON(http.StatusOK, LegalMovesResponse{Moves: moves})
}

func (h *Handlers) MakeMove(c *gin.Context) {
	id := c.Param("id")

	var req MoveRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	move, err := parseUCI(strings.TrimSpace(req.UCI))
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	h.store.mu.Lock()
	game, ok := h.store.games[id]
	if !ok {
		h.store.mu.Unlock()
		writeError(c, http.StatusNotFound, "game not found")
		return
	}

	status := computeStatus(game)
	if status.Result != resultOngoing {
		h.store.mu.Unlock()
		writeError(c, http.StatusConflict, "game already finished")
		return
	}

	if err := game.Board.MakeMove(move); err != nil {
		h.store.mu.Unlock()
		writeError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	game.Moves = append(game.Moves, uciFromMove(move))
	game.PendingDrawOfferBy = nil
	game.UpdatedAt = time.Now().UTC()
	response := buildMoveResponse(game)
	h.store.mu.Unlock()

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) Status(c *gin.Context) {
	id := c.Param("id")

	h.store.mu.RLock()
	game, ok := h.store.games[id]
	if !ok {
		h.store.mu.RUnlock()
		writeError(c, http.StatusNotFound, "game not found")
		return
	}
	status := computeStatus(game)
	h.store.mu.RUnlock()

	c.JSON(http.StatusOK, StatusResponse{
		Result: status.Result,
		Flags:  status.Flags,
	})
}

func (h *Handlers) History(c *gin.Context) {
	id := c.Param("id")

	h.store.mu.RLock()
	game, ok := h.store.games[id]
	if !ok {
		h.store.mu.RUnlock()
		writeError(c, http.StatusNotFound, "game not found")
		return
	}
	moves := make([]string, len(game.Moves))
	copy(moves, game.Moves)
	h.store.mu.RUnlock()

	c.JSON(http.StatusOK, HistoryResponse{
		ID:    id,
		Moves: moves,
	})
}

func (h *Handlers) Resign(c *gin.Context) {
	id := c.Param("id")

	var req ResignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	color, err := parseColor(req.Color)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	h.store.mu.Lock()
	game, ok := h.store.games[id]
	if !ok {
		h.store.mu.Unlock()
		writeError(c, http.StatusNotFound, "game not found")
		return
	}

	status := computeStatus(game)
	if status.Result != resultOngoing {
		h.store.mu.Unlock()
		writeError(c, http.StatusConflict, "game already finished")
		return
	}

	game.Result = resultResigned
	game.EndedBy = endedByResignation
	game.Winner = color.Opposite().String()
	game.PendingDrawOfferBy = nil
	game.UpdatedAt = time.Now().UTC()

	status = computeStatus(game)
	h.store.mu.Unlock()

	c.JSON(http.StatusOK, ResignResponse{
		Result:  status.Result,
		Winner:  status.Winner,
		EndedBy: status.EndedBy,
		Flags:   status.Flags,
	})
}

func (h *Handlers) OfferDraw(c *gin.Context) {
	id := c.Param("id")

	var req OfferDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	color, err := parseColor(req.Color)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	h.store.mu.Lock()
	game, ok := h.store.games[id]
	if !ok {
		h.store.mu.Unlock()
		writeError(c, http.StatusNotFound, "game not found")
		return
	}

	status := computeStatus(game)
	if status.Result != resultOngoing {
		h.store.mu.Unlock()
		writeError(c, http.StatusConflict, "game already finished")
		return
	}

	game.PendingDrawOfferBy = &color
	game.UpdatedAt = time.Now().UTC()
	h.store.mu.Unlock()

	c.JSON(http.StatusOK, OfferDrawResponse{Offer: "pending"})
}

func (h *Handlers) AcceptDraw(c *gin.Context) {
	id := c.Param("id")

	var req AcceptDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	color, err := parseColor(req.Color)
	if err != nil {
		writeError(c, http.StatusBadRequest, err.Error())
		return
	}

	h.store.mu.Lock()
	game, ok := h.store.games[id]
	if !ok {
		h.store.mu.Unlock()
		writeError(c, http.StatusNotFound, "game not found")
		return
	}

	status := computeStatus(game)
	if status.Result != resultOngoing {
		h.store.mu.Unlock()
		writeError(c, http.StatusConflict, "game already finished")
		return
	}

	if game.PendingDrawOfferBy == nil {
		h.store.mu.Unlock()
		writeError(c, http.StatusConflict, "no pending draw offer")
		return
	}
	if *game.PendingDrawOfferBy == color {
		h.store.mu.Unlock()
		writeError(c, http.StatusConflict, "draw offer must be accepted by opponent")
		return
	}

	game.Result = resultDraw
	game.EndedBy = endedByDrawAgreement
	game.Winner = "none"
	game.PendingDrawOfferBy = nil
	game.UpdatedAt = time.Now().UTC()

	status = computeStatus(game)
	h.store.mu.Unlock()

	c.JSON(http.StatusOK, AcceptDrawResponse{
		Result:  status.Result,
		Winner:  status.Winner,
		EndedBy: status.EndedBy,
		Flags:   status.Flags,
	})
}

func parseColor(value string) (chess.Color, error) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "white":
		return chess.White, nil
	case "black":
		return chess.Black, nil
	default:
		return 0, fmt.Errorf("invalid color %q", value)
	}
}

func buildGameResponse(game *Game) GameResponse {
	status := computeStatus(game)
	return GameResponse{
		ID:       game.ID,
		FEN:      game.Board.ToFEN(),
		Turn:     game.Board.Turn().String(),
		Result:   status.Result,
		Winner:   status.Winner,
		EndedBy:  status.EndedBy,
		Flags:    status.Flags,
		Halfmove: game.Board.HalfMove(),
		Fullmove: game.Board.FullMove(),
		Meta: Meta{
			CreatedAt: game.CreatedAt,
			UpdatedAt: game.UpdatedAt,
			StartFEN:  game.StartFEN,
		},
	}
}

func buildMoveResponse(game *Game) MoveResponse {
	status := computeStatus(game)
	return MoveResponse{
		FEN:      game.Board.ToFEN(),
		Turn:     game.Board.Turn().String(),
		Result:   status.Result,
		Winner:   status.Winner,
		EndedBy:  status.EndedBy,
		Flags:    status.Flags,
		Halfmove: game.Board.HalfMove(),
		Fullmove: game.Board.FullMove(),
	}
}

func writeError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: message})
}

func isEmptyBody(err error) bool {
	return strings.Contains(err.Error(), "EOF")
}
