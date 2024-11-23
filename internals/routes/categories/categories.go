package noteRoutes

import (
	handlers "github.com/KashyretsIvanna/voice-balance/internals/handlers/categories"
	authHandler "github.com/KashyretsIvanna/voice-balance/internals/handlers/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupCategoriesRoutes(router fiber.Router) {
	categories := router.Group("/categories")
	categories.Post("", authHandler.AuthMiddleware, handlers.AddCategoryHandler)
	categories.Get("", authHandler.AuthMiddleware, handlers.GetCategoriesByUserID)

}
