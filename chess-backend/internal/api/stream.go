package api

import (
	"net/http"
	"sync"
	"time"

	"chess-backend/internal/store"

	"github.com/gin-gonic/gin"
	"go.jetify.com/sse"
)

type streamSubscriber struct {
	token string
	ch    chan GameResponse
}

type StreamHub struct {
	mu   sync.RWMutex
	subs map[string]map[chan GameResponse]streamSubscriber
}

func NewStreamHub() *StreamHub {
	return &StreamHub{subs: make(map[string]map[chan GameResponse]streamSubscriber)}
}

func (h *StreamHub) Subscribe(gameID, token string) chan GameResponse {
	ch := make(chan GameResponse, 1)
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[gameID] == nil {
		h.subs[gameID] = make(map[chan GameResponse]streamSubscriber)
	}
	h.subs[gameID][ch] = streamSubscriber{token: token, ch: ch}
	return ch
}

func (h *StreamHub) Unsubscribe(gameID string, ch chan GameResponse) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.subs[gameID] == nil {
		return
	}
	delete(h.subs[gameID], ch)
	if len(h.subs[gameID]) == 0 {
		delete(h.subs, gameID)
	}
	close(ch)
}

func (h *StreamHub) BroadcastGame(game *store.Game, builder func(*store.Game, string) GameResponse) {
	h.mu.RLock()
	subscribers := h.subs[game.ID]
	if len(subscribers) == 0 {
		h.mu.RUnlock()
		return
	}
	copySubs := make([]streamSubscriber, 0, len(subscribers))
	for _, sub := range subscribers {
		copySubs = append(copySubs, sub)
	}
	h.mu.RUnlock()

	for _, sub := range copySubs {
		response := builder(game, sub.token)
		select {
		case sub.ch <- response:
		default:
		}
	}
}

func (h *Handlers) StreamGame(c *gin.Context) {
	id := c.Param("id")
	token := playerTokenFromRequest(c)

	game, err := h.store.GetGame(c.Request.Context(), id)
	if err != nil {
		handleStoreError(c, err)
		return
	}

	conn, err := sse.Upgrade(c.Request.Context(), c.Writer, sse.WithHeartbeatInterval(30*time.Second))
	if err != nil {
		writeError(c, http.StatusInternalServerError, err.Error())
		return
	}
	defer conn.Close()

	ch := h.hub.Subscribe(id, token)
	defer h.hub.Unsubscribe(id, ch)

	if err := conn.SendEvent(c.Request.Context(), &sse.Event{Data: buildGameResponseForToken(game, token)}); err != nil {
		return
	}

	for {
		select {
		case <-c.Request.Context().Done():
			return
		case resp, ok := <-ch:
			if !ok {
				return
			}
			if err := conn.SendEvent(c.Request.Context(), &sse.Event{Data: resp}); err != nil {
				return
			}
		}
	}
}
