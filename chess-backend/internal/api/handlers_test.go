package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestCreateGameAndIllegalMove(t *testing.T) {
	gin.SetMode(gin.TestMode)

	store := NewStore()
	handlers := NewHandlers(store)
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

	var created GameResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected id in response")
	}
	if created.Result != resultOngoing {
		t.Fatalf("expected ongoing result, got %q", created.Result)
	}

	moveBody := []byte(`{"uci":"e2e5"}`)
	rec = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPost, "/api/v1/games/"+created.ID+"/moves", bytes.NewBuffer(moveBody))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for illegal move, got %d", rec.Code)
	}
}
