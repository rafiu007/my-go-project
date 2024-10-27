// internal/infrastructure/db/mysql.go
package db

import (
    "fmt"
    "log"
    "my_go_project/internal/domain/entity"
    "gorm.io/driver/mysql"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

type Database struct {
    *gorm.DB
}

func NewDatabase(dsn string) (*Database, error) {
    config := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
    }

    db, err := gorm.Open(mysql.Open(dsn), config)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to database: %w", err)
    }

    // Enable foreign key constraints
    db = db.Set("gorm:table_options", "ENGINE=InnoDB")

    return &Database{db}, nil
}

func (db *Database) AutoMigrate() error {
    log.Println("Running auto-migration...")
    
    err := db.DB.AutoMigrate(
        &entity.CalendarEntry{},
        // Add other entities here as needed
    )
    if err != nil {
        return fmt.Errorf("failed to auto-migrate database: %w", err)
    }

    log.Println("Auto-migration completed successfully")
    return nil
}