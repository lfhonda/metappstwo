package database

import (
	"fmt"

	"github.com/glebarez/sqlite"
	"github.com/lfhonda/metappstwo.git/backend/internal/modules/auth"
	"github.com/lfhonda/metappstwo.git/backend/internal/modules/event"
	"gorm.io/gorm"
)

func Connect() *gorm.DB {
	db, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{
		TranslateError: true,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to open database: %v", err))
	}

	db.AutoMigrate(
		&auth.User{},
		&event.Event{},
		&event.EventRegistration{},
	)

	return db
}
