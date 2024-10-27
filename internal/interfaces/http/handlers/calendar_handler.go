// internal/interfaces/http/handlers/calendar_handler.go
package handlers

import (
    "encoding/json"
    "net/http"
    "time"

    "my_go_project/internal/domain/entity"
    "my_go_project/internal/infrastructure/db"
)

type CalendarHandler struct {
    db *db.Database
}

func NewCalendarHandler(db *db.Database) *CalendarHandler {
    return &CalendarHandler{db: db}
}

func (h *CalendarHandler) CreateEntry(w http.ResponseWriter, r *http.Request) {
    var entry entity.CalendarEntry
    if err := json.NewDecoder(r.Body).Decode(&entry); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    if err := h.db.Create(&entry).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(entry)
}

func (h *CalendarHandler) GetActiveEntries(w http.ResponseWriter, r *http.Request) {
    var entries []entity.CalendarEntry
    if err := h.db.Where("stop_date > ?", time.Now()).Find(&entries).Error; err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(entries)
}