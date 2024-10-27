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

    "my_go_project/config"
    "my_go_project/internal/domain/entity"
    "my_go_project/internal/infrastructure/db"
    "my_go_project/internal/infrastructure/queue"
    "my_go_project/internal/interfaces/http/handlers"
)

func main() {
    // Initialize configuration
    cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

    // Initialize database
    database, err := db.NewDatabase(cfg.Database.DSN)
    if err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }

    if err := database.AutoMigrate(); err != nil {
        log.Fatalf("Failed to run auto-migration: %v", err)
    }

    // Initialize SQS client
    sqsClient, err := queue.NewSQSClient(cfg.Queue.Endpoint, cfg.Queue.QueueURL)
    if err != nil {
        log.Fatalf("Failed to initialize SQS client: %v", err)
    }

    // Create context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()

    // Start SQS consumer and scheduler
    sqsClient.StartConsumer(ctx)
    sqsClient.StartScheduler(ctx, database)

    // Initialize handlers
    calendarHandler := handlers.NewCalendarHandler(database)

    // Set up routes
    mux := http.NewServeMux()

    // Calendar CRUD endpoints
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

    // Queue endpoint
    mux.HandleFunc("/calendar/queue", func(w http.ResponseWriter, r *http.Request) {
        if r.Method != http.MethodPost {
            http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
            return
        }

        var entries []entity.CalendarEntry
        if err := database.DB.Where("stop_date > ?", time.Now()).Find(&entries).Error; err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        if err := sqsClient.SendMessages(r.Context(), entries); err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }

        w.WriteHeader(http.StatusOK)
        fmt.Fprintf(w, "Successfully queued %d entries", len(entries))
    })

    // Create server
    srv := &http.Server{
        Addr:         ":" + cfg.Server.Port,
        Handler:      mux,
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
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer shutdownCancel()

    // Cancel the long-running operations
    cancel()

    // Shutdown server gracefully
    if err := srv.Shutdown(shutdownCtx); err != nil {
        log.Fatalf("Server forced to shutdown: %v", err)
    }

    log.Println("Server exiting")
}