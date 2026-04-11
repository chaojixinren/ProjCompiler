package cache

import (
	"sync"

	"projcompiler/internal/understanding/parsers"
	"projcompiler/internal/understanding/types"
)

type Store interface {
	Get(key string) (parsers.ParseResult, bool)
	Put(key string, value parsers.ParseResult)
}

type InMemoryStore struct {
	mu    sync.RWMutex
	items map[string]parsers.ParseResult
}

func NewInMemoryStore() *InMemoryStore {
	return &InMemoryStore{items: make(map[string]parsers.ParseResult)}
}

func (s *InMemoryStore) Get(key string) (parsers.ParseResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, ok := s.items[key]
	return value, ok
}

func (s *InMemoryStore) Put(key string, value parsers.ParseResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key] = value
}

func Key(file types.File, parserVersion string) string {
	return file.ContentHash + ":" + parserVersion + ":" + file.Language
}
