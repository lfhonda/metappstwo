package event

import (
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	Title           string    `gorm:"not null" json:"title"`
	Description     string    `gorm:"not null" json:"description"`
	Place           string    `gorm:"not null" json:"place"`
	Category        string    `gorm:"not null" json:"category"`
	Date            time.Time `gorm:"not null" json:"date"`
	EndDate         time.Time `gorm:"not null" json:"end_date"`
	MaxParticipants int       `gorm:"not null" json:"max_participants"`
	CheckInToken    string    `gorm:"uniqueIndex;not null" json:"-"`
}
