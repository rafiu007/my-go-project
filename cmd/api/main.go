// cmd/api/main.go
package main

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"

    "my_go_project/config"
    "my_go_project/internal/infrastructure/db"
    "my_go_project/internal/interfaces/http/handlers"
)

func main() {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }

    // Initialize database
    database, err := db.NewDatabase(cfg.Database.DSN)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    // Run auto-migration
    if err := database.AutoMigrate(); err != nil {
        log.Fatalf("Failed to run auto-migration: %v", err)
    }

    // Initialize handlers
    calendarHandler := handlers.NewCalendarHandler(database)

    // Set up routes
    mux := http.NewServeMux()
    mux.HandleFunc("/calendar", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case http.MethodPost:
            calendarHandler.CreateEntry(w, r)
        case http.MethodGet:
            calendarHandler.GetActiveEntries(w, r)
        default:
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        }
    })

    // Create server
    srv := &http.Server{
        Addr:         ":" + cfg.Server.Port,
        Handler:      mux,  // Set the mux as the handler
        ReadTimeout:  cfg.Server.ReadTimeout,
        WriteTimeout: cfg.Server.WriteTimeout,
    }

    // Start server
    go func() {
        log.Printf("Starting server on port %s", cfg.Server.Port)
        if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Failed to start server: %v", err)
        }
    }()

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    log.Println("Shutting down server...")

    // Create shutdown context with timeout
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    // Shutdown server gracefully
    if err := srv.Shutdown(ctx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exiting")
}