package event

import (
	"time"

	"gorm.io/gorm"
)

type EventRegistration struct {
	gorm.Model

	EventID   uint `gorm:"not null;uniqueIndex:idx_event_student" json:"event_id"`
	StudentID uint `gorm:"not null;uniqueIndex:idx_event_student" json:"student_id"`
}

type EventAttendance struct {
	gorm.Model

	EventID     uint      `gorm:"not null;uniqueIndex:idx_attendance_event_student" json:"event_id"`
	StudentID   uint      `gorm:"not null;uniqueIndex:idx_attendance_event_student" json:"student_id"`
	CheckedInAt time.Time `gorm:"not null" json:"checked_in_at"`
}
