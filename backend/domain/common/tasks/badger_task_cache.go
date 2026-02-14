package tasks

import (
	"encoding/json"
	"time"

	"teralux_app/domain/common/infrastructure"
)

// BadgerTaskCache provides namespaced cache helpers for async task statuses.
type BadgerTaskCache struct {
	badger    *infrastructure.BadgerService
	keyPrefix string
}

func NewBadgerTaskCache(badger *infrastructure.BadgerService, keyPrefix string) *BadgerTaskCache {
	return &BadgerTaskCache{
		badger:    badger,
		keyPrefix: keyPrefix,
	}
}

func (c *BadgerTaskCache) key(taskID string) string {
	return c.keyPrefix + taskID
}

func (c *BadgerTaskCache) Set(taskID string, status any) error {
	if c == nil || c.badger == nil {
		return nil
	}
	data, err := json.Marshal(status)
	if err != nil {
		return err
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
