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

	eventService := event.NewService(db)
	eventHandler := event.NewHandler(eventService)

	events := r.Group("/events")

	// Alunos e professores
	events.GET(
		"",
		middleware.AuthMiddleware(),
		eventHandler.List,
	)

	events.GET(
		"/:id",
		middleware.AuthMiddleware(),
		eventHandler.Get,
	)

	// Professores
	events.POST(
		"",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("teacher"),
		eventHandler.Create,
	)

	events.PUT(
		"/:id",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("teacher"),
		eventHandler.Update,
	)

	events.DELETE(
		"/:id",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("teacher"),
		eventHandler.Delete,
	)

	events.GET(
		"/:id/check-in",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("teacher"),
		eventHandler.CheckInScreen,
	)

	// Alunos
	events.POST(
		"/:id/register",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("student"),
		eventHandler.Register,
	)

	events.DELETE(
		"/:id/register",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("student"),
		eventHandler.CancelRegistration,
	)

	events.POST(
		"/:id/check-in",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("student"),
		eventHandler.CheckIn,
	)

	// Certificados
	r.GET(
		"/certificates/:id/pdf",
		middleware.AuthMiddleware(),
		middleware.RoleMiddleware("student"),
		eventHandler.CertificatePDF,
	)

	r.GET(
		"/certificates/verify/:code",
		eventHandler.VerifyCertificate,
	)

	return r
}
