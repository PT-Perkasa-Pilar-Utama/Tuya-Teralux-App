package infrastructure

import (
	"fmt"
	"time"

	"sensio/domain/common/utils"

	"github.com/dgraph-io/badger/v3"
)

// BadgerService handles BadgerDB operations for caching and data persistence.
// It wraps the raw BadgerDB client to provide simplified methods for common operations.
type BadgerService struct {
	db         *badger.DB
	defaultTTL time.Duration
}

// NewBadgerService initializes a new BadgerService instance.
//
// param dbPath rule="required" The file system path where the database directory will be created or opened.
// return *BadgerService A pointer to the initialized service instance ready for use.
// return error An error if the database cannot be opened (e.g., permissions, locked).
// @throws error If BadgerDB fails to open the database file.
func NewBadgerService(dbPath string) (*BadgerService, error) {
	opts := badger.DefaultOptions(dbPath)
	opts.Logger = nil

	db, err := badger.Open(opts)
	if err != nil {
		return nil, fmt.Errorf("failed to open badger db: %w", err)
	}

	ttlStr := utils.AppConfig.CacheTTL
	ttl, err := time.ParseDuration(ttlStr)
	if err != nil {
		ttl = 1 * time.Hour // Default to 1 hour if invalid or not set
	}

	return &BadgerService{db: db, defaultTTL: ttl}, nil
}

// Close terminates the database connection and ensures all data is flushed to disk.
// This method should be called ensuring graceful shutdown of the application.
//
// return error An error if the closing process encounters any issue.
func (s *BadgerService) Close() error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Close()
}

// Set stores a key-value pair in the database using the configured default Time-To-Live (TTL).
//
// param key The unique identifier for the data.
// param value The byte array data to store.
// return error An error if the write operation fails.
// @throws error If the transaction fails to commit.
func (s *BadgerService) Set(key string, value []byte) error {
	return s.SetWithTTL(key, value, s.defaultTTL)
}

// SetWithTTL stores a key-value pair in the database with a custom Time-To-Live (TTL).
func (s *BadgerService) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	if s == nil || s.db == nil {
		return nil
	}
	start := time.Now()
	err := s.db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte(key), value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
	duration := time.Since(start)
	if err != nil {
		utils.LogError("BadgerService: Set failed | key=%s | ttl=%v | duration_ms=%d | error=%v", key, ttl, duration.Milliseconds(), err)
		return err
	}
	utils.LogDebug("BadgerService: Set completed | key=%s | ttl=%v | duration_ms=%d", key, ttl, duration.Milliseconds())
	return nil
}

// Get retrieves a value associated with the given key.
// It handles the transaction view automatically.
//
// param key The unique identifier to search for.
// return []byte The value stored under the key, or nil if the key does not exist.
// return error An error if the read operation fails (excluding KeyNotFound).
// @throws error if an internal database error occurs during the view transaction.
func (s *BadgerService) Get(key string) ([]byte, error) {
	val, _, err := s.GetWithTTL(key)
	return val, err
}

// GetWithTTL retrieves a value and also returns the remaining TTL for the key.
// If the key has no TTL (persistent), the duration returned will be 0.
func (s *BadgerService) GetWithTTL(key string) ([]byte, time.Duration, error) {
	if s == nil || s.db == nil {
		return nil, 0, nil
	}
	start := time.Now()
	var valCopy []byte
	var ttlRemaining time.Duration
	err := s.db.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			return err
		}

		expiresAt := item.ExpiresAt()
		if expiresAt > 0 {
			ttlRemaining = time.Until(time.Unix(int64(expiresAt), 0))
			utils.LogDebug("Cache Hit for '%s' | Expires in: %v", key, ttlRemaining)
		} else {
			utils.LogDebug("Cache Hit for '%s' | Expires in: Never (Persistent)", key)
		}

		valCopy, err = item.ValueCopy(nil)
		return err
	})
	duration := time.Since(start)

	if err != nil {
		if err == badger.ErrKeyNotFound {
			utils.LogDebug("Cache Miss for '%s' | duration_ms=%d", key, duration.Milliseconds())
			return nil, 0, nil // Return nil if not found, distinct from error
		}
		utils.LogError("BadgerService: Get failed | key=%s | duration_ms=%d | error=%v", key, duration.Milliseconds(), err)
		return nil, 0, err
	}

	utils.LogDebug("BadgerService: Get completed | key=%s | duration_ms=%d | value_size=%d", key, duration.Milliseconds(), len(valCopy))
	return valCopy, ttlRemaining, nil
}

// Delete removes a key and its associated value from the database.
//
// param key The unique identifier to remove.
// return error An error if the delete operation fails.
// @throws error If the transaction fails to commit.
func (s *BadgerService) Delete(key string) error {
	if s == nil || s.db == nil {
		return nil
	}
	start := time.Now()
	err := s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(key))
	})
	duration := time.Since(start)
	if err != nil {
		utils.LogError("BadgerService: Delete failed | key=%s | duration_ms=%d | error=%v", key, duration.Milliseconds(), err)
		return err
	}
	utils.LogDebug("BadgerService: Delete completed | key=%s | duration_ms=%d", key, duration.Milliseconds())
	return nil
}

// ClearWithPrefix removes all keys that start with the specified prefix.
// This is useful for clearing a group of related cache items.
//
// param prefix The string pattern to match at the beginning of keys.
// return error An error if the bulk drop operation fails.
func (s *BadgerService) ClearWithPrefix(prefix string) error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.DropPrefix([]byte(prefix))
}

// SetPersistent stores a key-value pair in the database WITHOUT a Time-To-Live (TTL).
// This is used for persistent data that should survive cache flushes, such as device states.
//
// param key The unique identifier for the data.
// param value The byte array data to store.
// return error An error if the write operation fails.
// @throws error If the transaction fails to commit.
func (s *BadgerService) SetPersistent(key string, value []byte) error {
	if s == nil || s.db == nil {
		return nil
	}
	start := time.Now()
	err := s.db.Update(func(txn *badger.Txn) error {
		// No TTL - data persists indefinitely
		return txn.Set([]byte(key), value)
	})
	duration := time.Since(start)
	if err != nil {
		utils.LogError("BadgerService: SetPersistent failed | key=%s | duration_ms=%d | error=%v", key, duration.Milliseconds(), err)
		return err
	}
	utils.LogDebug("BadgerService: SetPersistent completed | key=%s | duration_ms=%d | value_size=%d", key, duration.Milliseconds(), len(value))
	return nil
}

// SetPreserveTTL updates an existing key's value while preserving its original TTL (time-to-live).
// If the key does not exist, this behaves like Set (stores with the default TTL).
// If the existing key has no TTL (persistent), the new value will be stored persistently as well.
//
// param key The key to update.
// param value The value to write.
// return error An error if the operation fails.
func (s *BadgerService) SetPreserveTTL(key string, value []byte) error {
	if s == nil || s.db == nil {
		return nil
	}
	return s.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte(key))
		if err != nil {
			if err == badger.ErrKeyNotFound {
				// key not found; behave like Set (with default TTL)
				entry := badger.NewEntry([]byte(key), value).WithTTL(s.defaultTTL)
				return txn.SetEntry(entry)
			}
			return err
		}

		expiresAt := item.ExpiresAt()
		if expiresAt == 0 {
			// Persistent entry -> set persistently
			return txn.Set([]byte(key), value)
		}

		// Compute remaining TTL
		ttl := time.Until(time.Unix(int64(expiresAt), 0))
		if ttl <= 0 {
			// TTL already expired or zero; write using default TTL
			entry := badger.NewEntry([]byte(key), value).WithTTL(s.defaultTTL)
			return txn.SetEntry(entry)
		}

		entry := badger.NewEntry([]byte(key), value).WithTTL(ttl)
		return txn.SetEntry(entry)
	})
}

// KeysWithPrefix is an alias for GetAllKeysWithPrefix to satisfy BadgerStore interface.
func (s *BadgerService) KeysWithPrefix(prefix string) ([]string, error) {
	return s.GetAllKeysWithPrefix(prefix)
}

// GetAllKeysWithPrefix retrieves all keys that start with the specified prefix.
// This is useful for cleanup operations or listing related items.
//
// param prefix The string pattern to match at the beginning of keys.
// return []string A slice of all matching keys.
// return error An error if the iteration fails.
func (s *BadgerService) GetAllKeysWithPrefix(prefix string) ([]string, error) {
	if s == nil || s.db == nil {
		return nil, nil
	}
	var keys []string
	err := s.db.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false // We only need keys, not values
		it := txn.NewIterator(opts)
		defer it.Close()

		prefixBytes := []byte(prefix)
		for it.Seek(prefixBytes); it.ValidForPrefix(prefixBytes); it.Next() {
			item := it.Item()
			key := string(item.Key())
			keys = append(keys, key)
		}
		return nil
	})

	if err != nil {
		utils.LogError("BadgerService: failed to get keys with prefix %s: %v", prefix, err)
		return nil, err
	}

	utils.LogDebug("BadgerService: Found %d keys with prefix '%s'", len(keys), prefix)
	return keys, nil
}

// FlushAll removes all CACHE data from the database (keys with "cache:" prefix).
// Device state and other persistent data (without "cache:" prefix) are preserved.
// This is a selective flush operation, not a complete database wipe.
//
// return error An error if the drop operation fails.
func (s *BadgerService) FlushAll() error {
	if s == nil || s.db == nil {
		return nil
	}
	// Only clear keys with "cache:" prefix
	cachePrefix := "cache:"
	err := s.db.DropPrefix([]byte(cachePrefix))
	if err != nil {
		utils.LogError("BadgerService: failed to flush cache: %v", err)
		return err
	}
	utils.LogInfo("BadgerService: Flushed all cache data (preserved persistent data)")
	return nil
}
