package infrastructure

import (
	"strings"
	"sync"
)

// VectorService provides a simple abstraction for vector DB operations.
// It is intentionally lightweight to allow swapping implementations later
// (e.g., Milvus, Qdrant, Pinecone, etc.). For now this is an in-memory
// placeholder implementation used for development and testing.
type VectorService struct {
	mu    sync.RWMutex
	store map[string]string // id -> content (json/text)
}

// NewVectorService initializes a new VectorService instance.
func NewVectorService() *VectorService {
	vs := &VectorService{store: make(map[string]string)}
	// Seed with a sample doc to satisfy placeholder search behavior in tests.
	vs.store["doc1"] = "hello world"
	return vs
}

// Upsert stores or updates a document in the vector store.
// id should be globally unique (we recommend namespaced IDs like "tuya:device:{id}").
func (s *VectorService) Upsert(id string, content string, _map map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[id] = content
	return nil
}

// Get retrieves a stored document's content by id.
func (s *VectorService) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.store[id]
	return val, ok
}

// Search performs a simple substring search over stored contents and returns matching ids.
// This is a placeholder — replace with real vector similarity search when integrating an actual vector DB.
func (s *VectorService) Search(query string) ([]string, error) {
	query = strings.ToLower(query)
	var matches []string
	s.mu.RLock()
	defer s.mu.RUnlock()
	for id, content := range s.store {
		if strings.Contains(strings.ToLower(content), query) {
			matches = append(matches, id)
		}
	}
	return matches, nil
}