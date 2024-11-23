package noteRoutes

import (
	authHandler "github.com/KashyretsIvanna/voice-balance/internals/handlers/auth"
	voiceHandler "github.com/KashyretsIvanna/voice-balance/internals/handlers/voice"

	"github.com/gofiber/fiber/v2"
)

func SetupVoiceRoutes(router fiber.Router) {
	transaction := router.Group("/voice")

	// Create a Note
	transaction.Post("/", authHandler.AuthMiddleware, voiceHandler.TranscribeAudio)

}
