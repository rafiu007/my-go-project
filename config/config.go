// config/config.go
package config

import (
    "fmt"
    "os"
    "time"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Queue    QueueConfig
}

type ServerConfig struct {
    Port         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Database string
    DSN      string
}

type QueueConfig struct {
    Endpoint  string
    Region    string
    QueueURL  string
}

func Load() (*Config, error) {
    cfg := &Config{
        Server: ServerConfig{
            Port:         getEnvOrDefault("SERVER_PORT", "8080"),
            ReadTimeout:  time.Second * 15,
            WriteTimeout: time.Second * 15,
        },
        Database: DatabaseConfig{
            Host:     getEnvOrDefault("DB_HOST", "localhost"),
            Port:     getEnvOrDefault("DB_PORT", "3306"),
            User:     getEnvOrDefault("DB_USER", "calendar_user"),
            Password: getEnvOrDefault("DB_PASSWORD", "calendar_pass"),
            Database: getEnvOrDefault("DB_NAME", "calendar_db"),
        },
        Queue: QueueConfig{
            Endpoint: getEnvOrDefault("AWS_ENDPOINT", "http://localhost:4566"),
            Region:   getEnvOrDefault("AWS_REGION", "us-east-1"),
            QueueURL: getEnvOrDefault("SQS_QUEUE_URL", "http://localhost:4566/000000000000/calendar-entries"),
        },
    }

    // Construct DSN
    cfg.Database.DSN = fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true",
        cfg.Database.User,
        cfg.Database.Password,
        cfg.Database.Host,
        cfg.Database.Port,
        cfg.Database.Database,
    )

    return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}