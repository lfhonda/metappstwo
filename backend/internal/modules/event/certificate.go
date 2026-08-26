package event

import (
	"time"

	"gorm.io/gorm"
)

type Certificate struct {
	gorm.Model

	EventID   uint      `gorm:"not null;index" json:"event_id"`
	StudentID uint      `gorm:"not null;index" json:"student_id"`
	Code      string    `gorm:"uniqueIndex;not null" json:"code"`
	IssuedAt  time.Time `gorm:"not null" json:"issued_at"`
}
