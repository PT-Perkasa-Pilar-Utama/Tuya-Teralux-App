package infrastructure

import (
	"os"
	"path/filepath"
	"teralux_app/domain/common/utils"
	"testing"
)

func TestInitDB_SQLite(t *testing.T) {
	// Use a temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Set environment variables
	_ = os.Setenv("DB_TYPE", "sqlite")
	_ = os.Setenv("DB_SQLITE_PATH", dbPath)
	defer func() {
		_ = os.Unsetenv("DB_TYPE")
		_ = os.Unsetenv("DB_SQLITE_PATH")
	}()

	utils.AppConfig = nil
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if db == nil {
		t.Fatal("Expected database instance, got nil")
	}

	// Verify database file was created
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("Database file was not created at %s", dbPath)
	}

	// Clean up
	_ = CloseDB()
}

func TestInitDB_DefaultsToSQLite(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "default.db")

	// Don't set DB_TYPE, should default to SQLite
	_ = os.Unsetenv("DB_TYPE")
	_ = os.Setenv("DB_SQLITE_PATH", dbPath)
	defer func() { _ = os.Unsetenv("DB_SQLITE_PATH") }()

	utils.AppConfig = nil
	db, err := InitDB()
	if err != nil {
		t.Fatalf("InitDB failed: %v", err)
	}

	if db == nil {
		t.Fatal("Expected database instance, got nil")
	}

	_ = CloseDB()
}

func TestInitDB_UnsupportedType(t *testing.T) {
	_ = os.Setenv("DB_TYPE", "unsupported")
	defer func() { _ = os.Unsetenv("DB_TYPE") }()

	utils.AppConfig = nil
	_, err := InitDB()
	if err == nil {
		t.Fatal("Expected error for unsupported DB type, got nil")
	}
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

	t.Run("Database initialized", func(t *testing.T) {
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "ping_test.db")

		_ = os.Setenv("DB_TYPE", "sqlite")
		_ = os.Setenv("DB_SQLITE_PATH", dbPath)
		defer func() {
			_ = os.Unsetenv("DB_TYPE")
			_ = os.Unsetenv("DB_SQLITE_PATH")
		}()

		utils.AppConfig = nil
		_, err := InitDB()
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		err = PingDB()
		if err != nil {
			t.Errorf("PingDB failed: %v", err)
		}

		_ = CloseDB()
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
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "close_test.db")

		_ = os.Setenv("DB_TYPE", "sqlite")
		_ = os.Setenv("DB_SQLITE_PATH", dbPath)
		defer func() {
			_ = os.Unsetenv("DB_TYPE")
			_ = os.Unsetenv("DB_SQLITE_PATH")
		}()

		utils.AppConfig = nil
		_, err := InitDB()
		if err != nil {
			t.Fatalf("InitDB failed: %v", err)
		}

		err = CloseDB()
		if err != nil {
			t.Errorf("CloseDB failed: %v", err)
		}
	})
}
