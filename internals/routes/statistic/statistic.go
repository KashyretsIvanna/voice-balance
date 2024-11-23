package noteRoutes

import (
	handlers "github.com/KashyretsIvanna/voice-balance/internals/handlers/statistics"
	authHandler "github.com/KashyretsIvanna/voice-balance/internals/handlers/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupStatisticsRoutes(router fiber.Router) {
	statistics := router.Group("/statistics")
	statistics.Get("/category",authHandler.AuthMiddleware, handlers.GetStatisticsByCategory)

}
