// internal/domain/entity/calendar_entry.go
package entity

import (
    "time"
    "gorm.io/gorm"
)

type CalendarEntry struct {
    ID        uint           `gorm:"primaryKey" json:"id"`
    StartDate time.Time      `gorm:"not null;index" json:"startDate"`
    StopDate  time.Time      `gorm:"not null;index" json:"stopDate"`
    CreatedAt time.Time      `gorm:"autoCreateTime" json:"createdAt"`
    UpdatedAt time.Time      `gorm:"autoUpdateTime" json:"updatedAt"`
}

// TableName specifies the table name for GORM
func (CalendarEntry) TableName() string {
    return "calendar_entries"
}