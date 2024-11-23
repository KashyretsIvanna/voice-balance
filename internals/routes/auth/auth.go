package noteRoutes

import (
	handlers "github.com/KashyretsIvanna/voice-balance/internals/handlers/auth"
	"github.com/gofiber/fiber/v2"
)

func SetupAuthRoutes(router fiber.Router) {
	auth := router.Group("/auth")

	// Google authentication
	auth.Get("/google", handlers.GoogleLogin)      // Redirect to Google login
	auth.Get("/callback", handlers.GoogleCallback) // Google callback for OAuth

	// Email/password authentication
	auth.Post("/email-login", handlers.EmailPasswordLogin) // Login with email and password

	// Token handling
	auth.Post("/refresh", handlers.RefreshToken) // Refresh access token

	// Register
	auth.Post("/register", handlers.Register) // Logout user and clear tokens

	// Logout
	auth.Get("/logout", handlers.Logout) // Logout user and clear tokens

}
