package controllers

import (
	"testing"
)

func TestNewCacheController(t *testing.T) {
	// Just test that constructor doesn't panic with nil
	controller := NewCacheController(nil, nil)

	if controller == nil {
		t.Fatal("NewCacheController returned nil")
	}
}
