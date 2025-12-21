package store

import (
	"context"
	"errors"
	"sync"
)

type MemoryStore struct {
	mu    sync.RWMutex
	games map[string]*Game
	moves map[string][]string
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		games: make(map[string]*Game),
		moves: make(map[string][]string),
	}
}

func (s *MemoryStore) CreateGame(_ context.Context, game *Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.games[game.ID]; exists {
		return errors.New("game already exists")
	}
	s.games[game.ID] = cloneGame(game)
	s.moves[game.ID] = append([]string(nil), game.Moves...)
	return nil
}

func (s *MemoryStore) GetGame(_ context.Context, id string) (*Game, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	game, ok := s.games[id]
	if !ok {
		return nil, ErrNotFound
	}
	return cloneGame(game), nil
}

func (s *MemoryStore) UpdateGame(_ context.Context, game *Game) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.games[game.ID]; !ok {
		return ErrNotFound
	}
	s.games[game.ID] = cloneGame(game)
	return nil
}

func (s *MemoryStore) UpdateGameWithMove(_ context.Context, game *Game, move string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.games[game.ID]; !ok {
		return ErrNotFound
	}
	s.games[game.ID] = cloneGame(game)
	s.moves[game.ID] = append(s.moves[game.ID], move)
	return nil
}

func (s *MemoryStore) ListMoves(_ context.Context, id string) ([]string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	moves, ok := s.moves[id]
	if !ok {
		if _, exists := s.games[id]; !exists {
			return nil, ErrNotFound
		}
		return []string{}, nil
	}
	out := make([]string, len(moves))
	copy(out, moves)
	return out, nil
}

func cloneGame(game *Game) *Game {
	if game == nil {
		return nil
	}

	clone := *game
	if game.Board != nil {
		clone.Board = game.Board.Clone()
	}
	if game.PendingDrawOfferBy != nil {
		color := *game.PendingDrawOfferBy
		clone.PendingDrawOfferBy = &color
	}
	if game.Moves != nil {
		clone.Moves = append([]string(nil), game.Moves...)
	}
	return &clone
}
