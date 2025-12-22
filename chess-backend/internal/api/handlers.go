package api

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"chess-backend/internal/chess"
	"chess-backend/internal/store"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handlers struct {
	store store.GameStore
	hub   *StreamHub
}

func NewHandlers(store store.GameStore) *Handlers {
	return &Handlers{
		store: store,
		hub:   NewStreamHub(),
	}
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

	var creatorColor chess.Color
	if strings.TrimSpace(req.PreferredColor) == "" {
		creatorColor = chess.White
	} else {
		creatorColor, err = parseColor(req.PreferredColor)
		if err != nil {
			writeError(c, http.StatusBadRequest, err.Error())
			return
		}
	}

	id, err := store.NewGameID()
	if err != nil {
		writeError(c, http.StatusInternalServerError, "failed to create game id")
		return
	}

	now := time.Now().UTC()
	playerToken := newPlayerToken()
	game := &store.Game{
		ID:        id,
		Board:     board,
		StartFEN:  board.ToFEN(),
		Moves:     []string{},
		CreatedAt: now,
		UpdatedAt: now,
	}
	if creatorColor == chess.White {
		game.PlayerWhiteToken = playerToken
		game.PlayerWhiteJoinedAt = &now
	} else {
		game.PlayerBlackToken = playerToken
		game.PlayerBlackJoinedAt = &now
	}

	status := computeStatus(game)
	game.Result = status.Result
	game.Winner = status.Winner
	game.EndedBy = status.EndedBy

	if err := h.store.CreateGame(c.Request.Context(), game); err != nil {
		writeError(c, http.StatusInternalServerError, "failed to store game: "+err.Error())
		return
	}
	response := buildGameResponseForToken(game, playerToken)

	c.JSON(http.StatusOK, PlayerGameResponse{
		GameResponse:  response,
		PlayerToken:   playerToken,
		OpponentColor: creatorColor.Opposite().String(),
	})
}

func (h *Handlers) GetGame(c *gin.Context) {
	id := c.Param("id")
	token := playerTokenFromRequest(c)

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	response := buildGameResponseForToken(game, token)

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) JoinGame(c *gin.Context) {
	id := c.Param("id")

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}

	if game.PlayerWhiteToken != "" && game.PlayerBlackToken != "" {
		writeError(c, http.StatusConflict, "game full")
		return
	}

	now := time.Now().UTC()
	playerToken := newPlayerToken()
	var playerColor chess.Color
	if game.PlayerWhiteToken == "" {
		game.PlayerWhiteToken = playerToken
		game.PlayerWhiteJoinedAt = &now
		playerColor = chess.White
	} else {
		game.PlayerBlackToken = playerToken
		game.PlayerBlackJoinedAt = &now
		playerColor = chess.Black
	}
	game.UpdatedAt = now

	if err := h.store.UpdateGame(c.Request.Context(), game); err != nil {
		handleStoreError(c, err)
		return
	}

	response := buildGameResponseForToken(game, playerToken)
	h.broadcastGame(game)

	c.JSON(http.StatusOK, PlayerGameResponse{
		GameResponse:  response,
		PlayerToken:   playerToken,
		OpponentColor: playerColor.Opposite().String(),
	})
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

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
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

	c.JSON(http.StatusOK, LegalMovesResponse{Moves: moves})
}

func (h *Handlers) MakeMove(c *gin.Context) {
	id := c.Param("id")
	token := playerTokenFromRequest(c)

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

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	color, ok := requirePlayerToken(c, game, token)
	if !ok {
		return
	}
	if game.Board.Turn() != color {
		writeError(c, http.StatusConflict, "not your turn")
		return
	}

	status := computeStatus(game)
	if status.Result != resultOngoing {
		writeError(c, http.StatusConflict, "game already finished")
		return
	}

	if err := game.Board.MakeMove(move); err != nil {
		writeError(c, http.StatusUnprocessableEntity, err.Error())
		return
	}

	moveUCI := uciFromMove(move)
	game.PendingDrawOfferBy = nil
	game.UpdatedAt = time.Now().UTC()

	status = computeStatus(game)
	game.Result = status.Result
	game.Winner = status.Winner
	game.EndedBy = status.EndedBy

	if err := h.store.UpdateGameWithMove(c.Request.Context(), game, moveUCI); err != nil {
		handleStoreError(c, err)
		return
	}
	response := buildMoveResponseForToken(game, token)
	h.broadcastGame(game)

	c.JSON(http.StatusOK, response)
}

func (h *Handlers) Status(c *gin.Context) {
	id := c.Param("id")

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	status := computeStatus(game)

	c.JSON(http.StatusOK, StatusResponse{
		Result:  status.Result,
		Winner:  status.Winner,
		EndedBy: status.EndedBy,
		Flags:   status.Flags,
	})
}

func (h *Handlers) History(c *gin.Context) {
	id := c.Param("id")

	moves, err := h.store.ListMoves(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}

	c.JSON(http.StatusOK, HistoryResponse{
		ID:    id,
		Moves: moves,
	})
}

func (h *Handlers) Resign(c *gin.Context) {
	id := c.Param("id")
	token := playerTokenFromRequest(c)

	var req ResignRequest
	if err := c.ShouldBindJSON(&req); err != nil && !isEmptyBody(err) {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	color, ok := requirePlayerToken(c, game, token)
	if !ok {
		return
	}

	status := computeStatus(game)
	if status.Result != resultOngoing {
		writeError(c, http.StatusConflict, "game already finished")
		return
	}

	game.Result = resultResigned
	game.EndedBy = endedByResignation
	game.Winner = color.Opposite().String()
	game.PendingDrawOfferBy = nil
	game.UpdatedAt = time.Now().UTC()

	status = computeStatus(game)
	if err := h.store.UpdateGame(c.Request.Context(), game); err != nil {
		handleStoreError(c, err)
		return
	}
	h.broadcastGame(game)

	c.JSON(http.StatusOK, ResignResponse{
		Result:  status.Result,
		Winner:  status.Winner,
		EndedBy: status.EndedBy,
		Flags:   status.Flags,
	})
}

func (h *Handlers) OfferDraw(c *gin.Context) {
	id := c.Param("id")
	token := playerTokenFromRequest(c)

	var req OfferDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil && !isEmptyBody(err) {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	color, ok := requirePlayerToken(c, game, token)
	if !ok {
		return
	}
	if game.Board.Turn() != color {
		writeError(c, http.StatusConflict, "not your turn")
		return
	}

	status := computeStatus(game)
	if status.Result != resultOngoing {
		writeError(c, http.StatusConflict, "game already finished")
		return
	}

	game.Result = status.Result
	game.Winner = status.Winner
	game.EndedBy = status.EndedBy
	game.PendingDrawOfferBy = &color
	game.UpdatedAt = time.Now().UTC()
	if err := h.store.UpdateGame(c.Request.Context(), game); err != nil {
		handleStoreError(c, err)
		return
	}
	h.broadcastGame(game)

	c.JSON(http.StatusOK, OfferDrawResponse{Offer: "pending"})
}

func (h *Handlers) AcceptDraw(c *gin.Context) {
	id := c.Param("id")
	token := playerTokenFromRequest(c)

	var req AcceptDrawRequest
	if err := c.ShouldBindJSON(&req); err != nil && !isEmptyBody(err) {
		writeError(c, http.StatusBadRequest, "invalid request body")
		return
	}

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}
	color, ok := requirePlayerToken(c, game, token)
	if !ok {
		return
	}
	if game.Board.Turn() != color {
		writeError(c, http.StatusConflict, "not your turn")
		return
	}

	status := computeStatus(game)
	if status.Result != resultOngoing {
		writeError(c, http.StatusConflict, "game already finished")
		return
	}

	if game.PendingDrawOfferBy == nil {
		writeError(c, http.StatusConflict, "no pending draw offer")
		return
	}
	if *game.PendingDrawOfferBy == color {
		writeError(c, http.StatusConflict, "draw offer must be accepted by opponent")
		return
	}

	game.Result = resultDraw
	game.EndedBy = endedByDrawAgreement
	game.Winner = "none"
	game.PendingDrawOfferBy = nil
	game.UpdatedAt = time.Now().UTC()

	status = computeStatus(game)
	if err := h.store.UpdateGame(c.Request.Context(), game); err != nil {
		handleStoreError(c, err)
		return
	}
	h.broadcastGame(game)

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

func buildGameResponse(game *store.Game) GameResponse {
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

func buildMoveResponse(game *store.Game) MoveResponse {
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

func buildGameResponseForToken(game *store.Game, token string) GameResponse {
	response := buildGameResponse(game)
	if color, ok := playerColorForToken(game, token); ok {
		response.PlayerColor = color.String()
		response.BoardOrientation = color.String()
	}
	return response
}

func buildMoveResponseForToken(game *store.Game, token string) MoveResponse {
	response := buildMoveResponse(game)
	if color, ok := playerColorForToken(game, token); ok {
		response.PlayerColor = color.String()
		response.BoardOrientation = color.String()
	}
	return response
}

func playerColorForToken(game *store.Game, token string) (chess.Color, bool) {
	if token == "" {
		return 0, false
	}
	if token == game.PlayerWhiteToken {
		return chess.White, true
	}
	if token == game.PlayerBlackToken {
		return chess.Black, true
	}
	return 0, false
}

func requirePlayerToken(c *gin.Context, game *store.Game, token string) (chess.Color, bool) {
	color, ok := playerColorForToken(game, token)
	if !ok {
		writeError(c, http.StatusForbidden, "invalid player token")
		return 0, false
	}
	return color, true
}

func playerTokenFromRequest(c *gin.Context) string {
	token := strings.TrimSpace(c.GetHeader("X-Player-Token"))
	if token != "" {
		return token
	}
	return strings.TrimSpace(c.Query("token"))
}

func newPlayerToken() string {
	return uuid.NewString()
}

func writeError(c *gin.Context, code int, message string) {
	c.JSON(code, ErrorResponse{Error: message})
}

func isEmptyBody(err error) bool {
	return strings.Contains(err.Error(), "EOF")
}

func handleStoreError(c *gin.Context, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeError(c, http.StatusNotFound, "game not found")
		return
	}
	writeError(c, http.StatusInternalServerError, "storage error")
}

func (h *Handlers) broadcastGame(game *store.Game) {
	if h.hub == nil {
		return
	}
	h.hub.BroadcastGame(game, buildGameResponseForToken)
}
