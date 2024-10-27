// internal/scheduler/queue_scheduler.go
package scheduler

import (
    "context"
    "log"
    "time"
    "sync"

    "my_go_project/internal/infrastructure/db"
    "my_go_project/internal/infrastructure/queue"
)

type QueueScheduler struct {
    db      *db.Database
    sqs     *queue.SQSService
    ticker  *time.Ticker
    done    chan bool
    wg      sync.WaitGroup
}

func NewQueueScheduler(db *db.Database, sqs *queue.SQSService) *QueueScheduler {
    return &QueueScheduler{
        db:   db,
        sqs:  sqs,
        done: make(chan bool),
    }
}

func (s *QueueScheduler) Start() {
    s.ticker = time.NewTicker(5 * time.Minute)
    s.wg.Add(1)
    go s.run()
}

func (s *QueueScheduler) run() {
    defer s.wg.Done()

    // Run immediately on start
    s.processEntries()

    for {
        select {
        case <-s.ticker.C:
            s.processEntries()
        case <-s.done:
            return
        }
    }
}

func (s *QueueScheduler) processEntries() {
    ctx := context.Background()
    var entries []entity.CalendarEntry
    
    if err := s.db.Where("stop_date > ?", time.Now()).Find(&entries).Error; err != nil {
        log.Printf("Error fetching active entries: %v", err)
        return
    }

    if err := s.sqs.SendEntries(ctx, entries); err != nil {
        log.Printf("Error sending entries to queue: %v", err)
        return
    }

    log.Printf("Successfully processed %d entries", len(entries))
}

func (s *QueueScheduler) Stop() {
    s.ticker.Stop()
    s.done <- true
    s.wg.Wait()
}