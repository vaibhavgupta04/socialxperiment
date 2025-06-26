package db

import (
    "fmt"
    "log"
    "time"
    "os"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
	"github.com/gopro/internal/config"
)

func InitPostgres(cfg *config.Config) *gorm.DB {
    // Example: Read DB config from env or cfg
    host := os.Getenv("PG_HOST")
    port := os.Getenv("PG_PORT")
    user := os.Getenv("PG_USER")
    password := os.Getenv("PG_PASSWORD")
    dbname := os.Getenv("PG_DBNAME")

    dsn := fmt.Sprintf(
        "host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        host, port, user, password, dbname,
    )

    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        PrepareStmt: true, // Caches prepared statements for high concurrency
        Logger:      logger.Default.LogMode(logger.Warn),
    })
    if err != nil {
        log.Fatalf("failed to connect to postgres: %v", err)
    }

    // Set connection pool settings for high concurrency
    sqlDB, err := db.DB()
    if err != nil {
        log.Fatalf("failed to get db instance: %v", err)
    }
    sqlDB.SetMaxOpenConns(100)                  // Max open connections
    sqlDB.SetMaxIdleConns(20)                   // Max idle connections
    sqlDB.SetConnMaxLifetime(30 * time.Minute)  // Max connection lifetime

    return db
}