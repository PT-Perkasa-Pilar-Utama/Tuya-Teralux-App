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
	words := strings.Fields(query) // Split query into words

	type match struct {
		id    string
		score int
	}

	matchScores := make(map[string]int) // Track match scores
	s.mu.RLock()
	defer s.mu.RUnlock()

	for id, content := range s.store {
		contentLower := strings.ToLower(content)
		score := 0

		// Check each word in the query
		for _, word := range words {
			// Ignore noise and common smart home terms
			if len(word) <= 1 || word == "on" || word == "off" || word == "to" || word == "in" || word == "is" ||
				word == "dan" || word == "set" || word == "ke" || word == "nyalakan" || word == "matikan" ||
				word == "turn" || word == "the" {
				continue
			}

			// Synonym Expansion
			targets := []string{word}
			switch word {
			case "ac":
				targets = append(targets, "air conditioner", "conditioner")
			case "lamp", "light":
				targets = append(targets, "lamp", "light", "switch")
			case "tv":
				targets = append(targets, "television")
			}

			// Check if any target matches
			for _, t := range targets {
				if strings.Contains(contentLower, t) {
					// Higher score for exact word match in device name
					if strings.Contains(contentLower, "device: "+t) || strings.Contains(contentLower, t+" ") {
						score += 10 // Exact match in name
					} else {
						score += 1 // Generic match
					}
					break
				}
			}
		}

		// Only include devices with positive scores
		if score > 0 {
			matchScores[id] = score
		}
	}

	// Sort matches by score (highest first)
	matches := make([]match, 0, len(matchScores))
	for id, score := range matchScores {
		matches = append(matches, match{id: id, score: score})
	}

	// Sort by score descending
	for i := 0; i < len(matches); i++ {
		for j := i + 1; j < len(matches); j++ {
			if matches[j].score > matches[i].score {
				matches[i], matches[j] = matches[j], matches[i]
			}
		}
	}

	// Return only IDs, prioritizing higher scores
	// If top score is high (>= 10), only return those with the MAXIMUM score
	// This avoids "generic" matches (like 'temperature' matching a sensor)
	// from cluttering results when a specific match (like 'Sharp AC') is found.
	var result []string
	if len(matches) > 0 {
		topScore := matches[0].score
		if topScore >= 10 {
			// Return only matches with the TOP score
			for _, m := range matches {
				if m.score == topScore {
					result = append(result, m.id)
				}
			}
		} else {
			// Return matches with score >= 1 (at least one word matched via synonym or exact)
			// Changed from >= 2 to allow single-word synonym matches like "light" -> "lamp"
			for _, m := range matches {
				if m.score >= 1 {
					result = append(result, m.id)
				}
			}
		}
	}

	return result, nil
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
