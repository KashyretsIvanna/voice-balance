package router

import (
	swagger "github.com/arsmn/fiber-swagger/v2"
	"github.com/gofiber/fiber/v2"

	_ "github.com/KashyretsIvanna/voice-balance/docs"
	authRoutes "github.com/KashyretsIvanna/voice-balance/internals/routes/auth"
	categoryRoutes "github.com/KashyretsIvanna/voice-balance/internals/routes/categories"
	statisticRoutes "github.com/KashyretsIvanna/voice-balance/internals/routes/statistic"
	transactionRoutes "github.com/KashyretsIvanna/voice-balance/internals/routes/transaction"
	userRoutes "github.com/KashyretsIvanna/voice-balance/internals/routes/user"
	voiceRoutes "github.com/KashyretsIvanna/voice-balance/internals/routes/voice"

	"github.com/gofiber/fiber/v2/middleware/logger"
)

func SetupRoutes(app *fiber.App) {

	app.Get("/swagger/*", swagger.Handler)
	// Group api calls with param '/api'
	api := app.Group("/api", logger.New())

	// Setup note routes, can use same syntax to add routes for more models
	statisticRoutes.SetupStatisticsRoutes(api)
	transactionRoutes.SetupTransactionRoutes(api)
	authRoutes.SetupAuthRoutes(api)
	categoryRoutes.SetupCategoriesRoutes(api)
	userRoutes.SetupUserRoutes(api)
	voiceRoutes.SetupVoiceRoutes(api)

}
