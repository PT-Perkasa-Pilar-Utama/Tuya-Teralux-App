package infrastructure

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB is the global database instance
var DB *gorm.DB

// InitDB initializes the database connection pool
// Returns the database instance and any error encountered
func InitDB() (*gorm.DB, error) {
	dbType := strings.ToLower(os.Getenv("DB_TYPE"))
	if dbType == "" {
		dbType = "sqlite"
	}

	switch dbType {
	case "mysql":
		host := os.Getenv("DB_HOST")
		port := os.Getenv("DB_PORT")
		user := os.Getenv("DB_USER")
		password := os.Getenv("DB_PASSWORD")
		dbname := os.Getenv("DB_NAME")

		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			user, password, host, port, dbname,
		)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to connect to database: %w", err)
		}

		sqlDB, err := db.DB()
		if err != nil {
			return nil, fmt.Errorf("failed to get database instance: %w", err)
		}

		// Set connection pool settings
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(5)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)

		// Test the connection
		if err := sqlDB.Ping(); err != nil {
			return nil, fmt.Errorf("failed to ping database: %w", err)
		}

		DB = db
		log.Println("✅ Database connection established successfully (MySQL)")
		return db, nil

	case "sqlite":
		dbPath := os.Getenv("DB_SQLITE_PATH")
		if dbPath == "" {
			dbPath = "./tmp/teralux.db"
		}
		if err := os.MkdirAll(filepath.Dir(dbPath), 0o755); err != nil {
			return nil, fmt.Errorf("failed to create sqlite directory: %w", err)
		}

		db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Error),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to open sqlite database: %w", err)
		}

		DB = db
		log.Printf("✅ Database initialized using SQLite at %s", dbPath)
		return db, nil

	default:
		return nil, fmt.Errorf("unsupported DB_TYPE: %s", dbType)
	}
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
