package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"chess-backend/internal/store"

	"github.com/gin-gonic/gin"
)

func TestCreateGameAndIllegalMove(t *testing.T) {
	gin.SetMode(gin.TestMode)

	memStore := store.NewMemoryStore()
	handlers := NewHandlers(memStore)
	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.POST("/games", handlers.CreateGame)
	v1.POST("/games/:id/moves", handlers.MakeMove)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/games", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var created PlayerGameResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected id in response")
	}
	if created.PlayerToken == "" {
		t.Fatalf("expected player token in response")
	}
	if created.Result != resultOngoing {
		t.Fatalf("expected ongoing result, got %q", created.Result)
	}

	moveBody := []byte(`{"uci":"e2e5"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/games/"+created.ID+"/moves", bytes.NewBuffer(moveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Player-Token", created.PlayerToken)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for illegal move, got %d", rec.Code)
	}
}

func TestJoinGameRestrictions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	memStore := store.NewMemoryStore()
	handlers := NewHandlers(memStore)
	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.POST("/games", handlers.CreateGame)
	v1.POST("/games/:id/join", handlers.JoinGame)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/games", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var created PlayerGameResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/games/"+created.ID+"/join", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/games/"+created.ID+"/join", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for full game, got %d", rec.Code)
	}
}

func TestMoveAuthorization(t *testing.T) {
	gin.SetMode(gin.TestMode)

	memStore := store.NewMemoryStore()
	handlers := NewHandlers(memStore)
	router := gin.New()
	v1 := router.Group("/api/v1")
	v1.POST("/games", handlers.CreateGame)
	v1.POST("/games/:id/join", handlers.JoinGame)
	v1.POST("/games/:id/moves", handlers.MakeMove)

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/games", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var created PlayerGameResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/games/"+created.ID+"/join", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var joined PlayerGameResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &joined); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	moveBody := []byte(`{"uci":"e2e4"}`)

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/games/"+created.ID+"/moves", bytes.NewBuffer(moveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Player-Token", "invalid-token")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for invalid token, got %d", rec.Code)
	}

	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/games/"+created.ID+"/moves", bytes.NewBuffer(moveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Player-Token", joined.PlayerToken)
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for wrong turn, got %d", rec.Code)
	}
}
