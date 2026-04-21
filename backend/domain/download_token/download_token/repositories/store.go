package repositories

import (
	"fmt"
	"sync"
	"time"

	"sensio/domain/download_token/download_token/entities"
)

type Store struct {
	mu     sync.RWMutex
	tokens map[string]*entities.Token
}

func NewStore() *Store {
	return &Store{tokens: make(map[string]*entities.Token)}
}

func (s *Store) Save(token *entities.Token) {
	if token == nil || token.TokenID == "" {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.tokens[token.TokenID] = cloneToken(token)
}

func (s *Store) Get(tokenID string) (*entities.Token, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	token, ok := s.tokens[tokenID]
	if !ok {
		return nil, false
	}

	return cloneToken(token), true
}

func (s *Store) MarkConsumed(tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, ok := s.tokens[tokenID]
	if !ok {
		return fmt.Errorf("token not found")
	}
	if token.RevokedAt != nil {
		return fmt.Errorf("token revoked")
	}
	if token.ConsumedAt != nil {
		return fmt.Errorf("token already consumed")
	}

	now := time.Now().UTC()
	token.ConsumedAt = &now

	return nil
}

func (s *Store) Revoke(tokenID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	token, ok := s.tokens[tokenID]
	if !ok {
		return fmt.Errorf("token not found")
	}
	if token.RevokedAt != nil {
		return fmt.Errorf("token already revoked")
	}

	now := time.Now().UTC()
	token.RevokedAt = &now

	return nil
}

func cloneToken(token *entities.Token) *entities.Token {
	if token == nil {
		return nil
	}

	clone := *token
	if token.ConsumedAt != nil {
		consumed := *token.ConsumedAt
		clone.ConsumedAt = &consumed
	}
	if token.RevokedAt != nil {
		revoked := *token.RevokedAt
		clone.RevokedAt = &revoked
	}

	return &clone
}
