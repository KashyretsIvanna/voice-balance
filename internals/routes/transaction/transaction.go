package noteRoutes

import (
	handlers "github.com/KashyretsIvanna/voice-balance/internals/handlers/transaction"
	authHandler "github.com/KashyretsIvanna/voice-balance/internals/handlers/auth"

	"github.com/gofiber/fiber/v2"
)

func SetupTransactionRoutes(router fiber.Router) {
	transaction := router.Group("/transaction")

	// Create a Note
	transaction.Post("/",authHandler.AuthMiddleware, handlers.AddTransaction)
	transaction.Get("/",authHandler.AuthMiddleware,  handlers.GetTransactions)


}
