package infrastructure

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// VectorService provides a simple abstraction for vector DB operations.
// It is intentionally lightweight to allow swapping implementations later
// (e.g., Milvus, Qdrant, Pinecone, etc.). For now this is an in-memory
// placeholder implementation used for development and testing.
type VectorService struct {
	mu       sync.RWMutex
	store    map[string]string // id -> content (json/text)
	filePath string
}

// NewVectorService initializes a new VectorService instance with persistence.
func NewVectorService(filePath string) *VectorService {
	vs := &VectorService{
		store:    make(map[string]string),
		filePath: filePath,
	}

	// Load existing data if file exists
	if filePath != "" {
		if data, err := os.ReadFile(filePath); err == nil {
			_ = json.Unmarshal(data, &vs.store)
		}
	}

	return vs
}

// save persists the current store to the file.
func (s *VectorService) save() error {
	if s.filePath == "" {
		return nil
	}

	// Ensure directory exists
	dir := filepath.Dir(s.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(s.store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.filePath, data, 0644)
}

// Upsert stores or updates a document in the vector store.
// id should be globally unique (we recommend namespaced IDs like "tuya:device:{id}").
func (s *VectorService) Upsert(id string, content string, _map map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store[id] = content
	return s.save()
}

// Get retrieves a stored document's content by id.
func (s *VectorService) Get(id string) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.store[id]
	return val, ok
}

// Search performs a simple keyword-based matching over stored contents.
func (s *VectorService) Search(query string) ([]string, error) {
	query = strings.ToLower(query)
	words := strings.Fields(query) // Split query into words (e.g., "turn on the lamp" -> ["turn", "on", "the", "lamp"])

	var matches []string
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, content := range s.store {
		contentLower := strings.ToLower(content)

		// If any keyword matches the content, consider it a candidate
		for _, word := range words {
			// Ignore very short words like "on", "the", "a" to avoid noise
			if len(word) <= 2 {
				continue
			}
			if strings.Contains(contentLower, word) {
				matches = append(matches, id)
				break
			}
		}
	}
	return matches, nil
}

// Count returns the number of documents in the store.
func (s *VectorService) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.store)
}

// FlushAll clears all stored documents from the vector store and persists the change.
func (s *VectorService) FlushAll() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.store = make(map[string]string)
	return s.save()
}
