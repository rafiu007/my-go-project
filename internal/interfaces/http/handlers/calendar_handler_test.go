package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"my_go_project/internal/domain/entity"
	"gorm.io/gorm"
)

type MockDatabase struct {
	entries []entity.CalendarEntry
}

func (m *MockDatabase) Create(entry *entity.CalendarEntry) *gorm.DB {
	m.entries = append(m.entries, *entry)
	return &gorm.DB{}
}

func (m *MockDatabase) Where(query string, args ...interface{}) *gorm.DB {
	return &gorm.DB{}
}

func (m *MockDatabase) Find(out interface{}) *gorm.DB {
	if len(m.entries) > 0 {
		outValue := out.(*[]entity.CalendarEntry)
		*outValue = m.entries
	}
	return &gorm.DB{}
}

func TestCreateEntry(t *testing.T) {
	mockDB := &MockDatabase{}
	handler := NewCalendarHandler(mockDB)

	entry := entity.CalendarEntry{
		StartDate: time.Now(),
		StopDate:  time.Now().Add(1 * time.Hour),
	}

	// Convert entry to JSON
	entryJSON, _ := json.Marshal(entry)

	req := httptest.NewRequest(http.MethodPost, "/calendar", httptest.NewBody(entryJSON))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	handler.CreateEntry(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	var createdEntry entity.CalendarEntry
	json.Unmarshal(rec.Body.Bytes(), &createdEntry)

	assert.NotEmpty(t, createdEntry.ID)
	assert.Equal(t, entry.StartDate, createdEntry.StartDate)
	assert.Equal(t, entry.StopDate, createdEntry.StopDate)
}

func TestGetActiveEntries(t *testing.T) {
	mockDB := &MockDatabase{
		entries: []entity.CalendarEntry{
			{
				ID:        1,
				StartDate: time.Now(),
				StopDate:  time.Now().Add(1 * time.Hour),
			},
		},
	}
	handler := NewCalendarHandler(mockDB)

	req := httptest.NewRequest(http.MethodGet, "/calendar", nil)
	rec := httptest.NewRecorder()

	handler.GetActiveEntries(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var entries []entity.CalendarEntry
	json.Unmarshal(rec.Body.Bytes(), &entries)

	assert.Equal(t, 1, len(entries))
	assert.Equal(t, mockDB.entries[0].ID, entries[0].ID)
}

