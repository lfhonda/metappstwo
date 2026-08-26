package main

import (
	"github.com/lfhonda/metappstwo.git/backend/internal/platform/database"
	"github.com/lfhonda/metappstwo.git/backend/internal/router"
)

func main() {
	db := database.Connect()

	r := router.SetupRouter(db)

	r.Run(":8080")
}
