package infrastructure

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// InitDB initializes the database connection pool using SQLite.
// Returns the database instance and any error encountered.
func InitDB() (*gorm.DB, error) {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType != "" && dbType != "sqlite" {
		return nil, fmt.Errorf("unsupported DB_TYPE: %s (only 'sqlite' is supported)", dbType)
	}

	dbPath := os.Getenv("DB_SQLITE_PATH")
	if dbPath == "" {
		dbPath = "./tmp/teralux.db"
	}

	if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
		return nil, fmt.Errorf("failed to create sqlite directory: %w", err)
	}

	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open sqlite database: %w", err)
	}

	DB = db
	log.Printf("✅ Database initialized using SQLite at %s", dbPath)
	return db, nil
}

// CloseDB closes the database connection gracefully
func CloseDB() error {
	if DB == nil {
		return nil
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Close(); err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}

	log.Println("✅ Database connection closed successfully")
	return nil
}

// PingDB checks if the database connection is alive
func PingDB() error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get database instance: %w", err)
	}

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("database ping failed: %w", err)
	}

	return nil
}
