package tasks

import (
	"encoding/json"
	"time"

	"sensio/domain/common/infrastructure"
)

// BadgerStore defines the interface for badger database operations
type BadgerStore interface {
	Set(key string, value []byte) error
	SetWithTTL(key string, value []byte, ttl time.Duration) error
	SetPreserveTTL(key string, value []byte) error
	GetWithTTL(key string) ([]byte, time.Duration, error)
	Delete(key string) error
	KeysWithPrefix(prefix string) ([]string, error)
}

// BadgerTaskCache provides namespaced cache helpers for async task statuses.
type BadgerTaskCache struct {
	badger    BadgerStore
	keyPrefix string
}

func NewBadgerTaskCache(badger BadgerStore, keyPrefix string) *BadgerTaskCache {
	return &BadgerTaskCache{
		badger:    badger,
		keyPrefix: keyPrefix,
	}
}

// NewBadgerTaskCacheFromService creates a BadgerTaskCache from a concrete BadgerService
func NewBadgerTaskCacheFromService(badger *infrastructure.BadgerService, keyPrefix string) *BadgerTaskCache {
	return &BadgerTaskCache{
		badger:    badger,
		keyPrefix: keyPrefix,
	}
}

func (c *BadgerTaskCache) key(taskID string) string {
	return c.keyPrefix + taskID
}

func (c *BadgerTaskCache) Set(taskID string, status any) error {
	return c.SetWithTTL(taskID, status, 0) // Uses store default
}

func (c *BadgerTaskCache) SetWithTTL(taskID string, status any, ttl time.Duration) error {
	if c == nil || c.badger == nil {
		return nil
	}
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	if ttl > 0 {
		return c.badger.SetWithTTL(c.key(taskID), data, ttl)
	}
	return c.badger.Set(c.key(taskID), data)
}

func (c *BadgerTaskCache) SetPreserveTTL(taskID string, status any) error {
	if c == nil || c.badger == nil {
		return nil
	}
	data, err := json.Marshal(status)
	if err != nil {
		return err
	}
	return c.badger.SetPreserveTTL(c.key(taskID), data)
}

func (c *BadgerTaskCache) GetWithTTL(taskID string, out any) (time.Duration, bool, error) {
	if c == nil || c.badger == nil {
		return 0, false, nil
	}
	data, ttl, err := c.badger.GetWithTTL(c.key(taskID))
	if err != nil {
		return 0, false, err
	}
	if len(data) == 0 {
		return 0, false, nil
	}
	if err := json.Unmarshal(data, out); err != nil {
		return 0, false, err
	}
	return ttl, true, nil
}
