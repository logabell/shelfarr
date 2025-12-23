package db

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Initialize creates and configures the database connection
func Initialize(dbPath string) (*gorm.DB, error) {
	// SQLite with WAL mode for concurrent reads/writes
	db, err := gorm.Open(sqlite.Open(dbPath+"?_journal_mode=WAL&_busy_timeout=5000"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

// Migrate runs all database migrations
func Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Author{},
		&Series{},
		&Book{},
		&MediaFile{},
		&User{},
		&ReadProgress{},
		&Indexer{},
		&DownloadClient{},
		&QualityProfile{},
		&Notification{},
		&HardcoverList{},
		&Download{},
		&Setting{},
		&RootFolder{},
	)
}
