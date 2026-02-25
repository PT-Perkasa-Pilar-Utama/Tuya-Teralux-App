package infrastructure

import (
	"os"
	"teralux_app/domain/common/utils"
	"testing"
)

func TestInitDB(t *testing.T) {
	t.Run("Invalid Config", func(t *testing.T) {
		// Set invalid environment variables to force failure
		_ = os.Setenv("GO_TEST", "true")
		_ = os.Setenv("MYSQL_HOST", "invalid_host")
		_ = os.Setenv("MYSQL_PORT", "9999")
		_ = os.Setenv("MYSQL_USER", "invalid")
		_ = os.Setenv("MYSQL_PASSWORD", "invalid")
		_ = os.Setenv("MYSQL_DATABASE", "invalid")
		defer func() {
			_ = os.Unsetenv("GO_TEST")
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
}

func TestCloseDB(t *testing.T) {
	t.Run("Close nil database", func(t *testing.T) {
		DB = nil
		err := CloseDB()
		if err != nil {
			t.Errorf("CloseDB should not error on nil DB, got: %v", err)
		}
	})
}
