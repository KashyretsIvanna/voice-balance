package noteRoutes

import (
	authHandler "github.com/KashyretsIvanna/voice-balance/internals/handlers/auth"
	userHandler "github.com/KashyretsIvanna/voice-balance/internals/handlers/user"
	"github.com/gofiber/fiber/v2"
)

func SetupUserRoutes(router fiber.Router) {
	user := router.Group("/user")

	// Read all Users
	user.Get("/", authHandler.AuthMiddleware, userHandler.GetUsers)
	user.Get("/me", authHandler.AuthMiddleware, userHandler.GetMe)

	// // Read one User
	user.Get("/:userId", authHandler.AuthMiddleware, userHandler.GetUser)

	// // Delete one User
	user.Delete("/:userId", authHandler.AuthMiddleware, userHandler.DeleteUser)
}
