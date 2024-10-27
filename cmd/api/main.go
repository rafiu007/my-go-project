// cmd/api/main.go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "my-go-project/config"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize logger (we'll implement this later)
    logger := log.New(os.Stdout, "", log.LstdFlags)

	// Initialize database
    database, err := db.NewDatabase(cfg.Database.DSN)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

	// Run auto-migration
    if err := database.AutoMigrate(); err != nil {
        log.Fatalf("Failed to run auto-migration: %v", err)
    }

    // Create server
    srv := &http.Server{
        Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
        // We'll add the handler later
    }

    // Start server
    go func() {
        logger.Printf("Starting server on port %s", cfg.Server.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            logger.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    logger.Println("Shutting down server...")

    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Shutdown server gracefully
    if err := srv.Shutdown(ctx); err != nil {
        logger.Fatalf("Server forced to shutdown: %v", err)
    }

    logger.Println("Server exiting")
}