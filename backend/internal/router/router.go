package router

import (
	"github.com/gin-gonic/gin"
	"github.com/lfhonda/metappstwo.git/backend/internal/middleware"
	"github.com/lfhonda/metappstwo.git/backend/internal/modules/auth"
	"github.com/lfhonda/metappstwo.git/backend/internal/modules/event"
	"gorm.io/gorm"
)

func SetupRouter(db *gorm.DB) *gin.Engine {
	r := gin.New()

	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	// Auth
	authService := auth.NewAuthService(db)
	authHandler := auth.NewAuthHandler(authService)

	r.POST("/auth/login", authHandler.Login)
	r.POST("/auth/register-teacher", authHandler.RegisterTeacher)

	r.POST(
		"/auth/register-student",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("teacher"),
		authHandler.RegisterStudent,
	)

	// Events
	eventService := event.NewEventService(db)
	eventHandler := event.NewEventHandler(eventService)

	// Eventos podem ser visualizados por usuários autenticados.
	r.GET(
		"/events",
		middleware.AuthMiddleware(),
		eventHandler.List,
	)

	r.GET(
		"/events/:id",
		middleware.AuthMiddleware(),
		eventHandler.GetByID,
	)

	// Professores gerenciam eventos.
	r.POST(
		"/events",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("teacher"),
		eventHandler.Create,
	)

	r.PUT(
		"/events/:id",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("teacher"),
		eventHandler.Update,
	)

	r.DELETE(
		"/events/:id",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("teacher"),
		eventHandler.Delete,
	)

	// Alunos gerenciam suas próprias inscrições.
	r.POST(
		"/events/:id/register",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("student"),
		eventHandler.Register,
	)

	r.DELETE(
		"/events/:id/register",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("student"),
		eventHandler.CancelRegistration,
	)

	return r
}
