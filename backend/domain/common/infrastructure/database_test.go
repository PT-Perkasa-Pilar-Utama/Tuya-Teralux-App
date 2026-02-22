package infrastructure

import (
	"os"
	"teralux_app/domain/common/utils"
	"testing"
)

func TestInitDB(t *testing.T) {
	t.Run("Invalid Config", func(t *testing.T) {
		// Set invalid environment variables to force failure
		_ = os.Setenv("MYSQL_HOST", "invalid_host")
		_ = os.Setenv("MYSQL_PORT", "9999")
		_ = os.Setenv("MYSQL_USER", "invalid")
		_ = os.Setenv("MYSQL_PASSWORD", "invalid")
		_ = os.Setenv("MYSQL_DATABASE", "invalid")
		defer func() {
			_ = os.Unsetenv("MYSQL_HOST")
			_ = os.Unsetenv("MYSQL_PORT")
			_ = os.Unsetenv("MYSQL_USER")
			_ = os.Unsetenv("MYSQL_PASSWORD")
			_ = os.Unsetenv("MYSQL_DATABASE")
		}()

		utils.AppConfig = nil
		_, err := InitDB()
		if err == nil {
			t.Fatal("Expected error when initializing with invalid MySQL config, got nil")
		}
	})
	
	t.Run("Valid Config (if available)", func(t *testing.T) {
		// This test depends on a real MySQL instance being available.
		// We'll try to initialize with env variables if they exist.
		utils.AppConfig = nil
		db, err := InitDB()
		if err != nil {
			t.Skipf("Skipping success path test: MySQL connection failed (expected if DB not reachable): %v", err)
		}
		
		if db == nil {
			t.Fatal("Expected database instance, got nil")
		}
		
		_ = CloseDB()
	})
}

func TestPingDB(t *testing.T) {
	t.Run("Database not initialized", func(t *testing.T) {
		// Reset global DB
		DB = nil

		err := PingDB()
		if err == nil {
			t.Fatal("Expected error when DB is nil, got nil")
		}
	})

	t.Run("With DB pointer", func(t *testing.T) {
		// We can't easily mock Ping without a real connection or a mock driver.
		// Since we're using MySQL directly, we'll try PingDB but expect failure if not connected.
		if DB == nil {
			t.Skip("Skipping since DB is nil")
		}
		_ = PingDB()
	})
}

func TestCloseDB(t *testing.T) {
	t.Run("Close nil database", func(t *testing.T) {
		DB = nil
		err := CloseDB()
		if err != nil {
			t.Errorf("CloseDB should not error on nil DB, got: %v", err)
		}
	})

	t.Run("Close initialized database", func(t *testing.T) {
		// Test behavior when DB is already initialized
		utils.AppConfig = nil
		_, err := InitDB()
		if err != nil {
			t.Skip("Skipping CloseDB test: InitDB failed")
		}

		err = CloseDB()
		if err != nil {
			t.Errorf("CloseDB failed: %v", err)
		}
	})
}
