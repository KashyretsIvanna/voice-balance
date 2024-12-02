package main

import (
	"github.com/KashyretsIvanna/voice-balance/database"
	"github.com/KashyretsIvanna/voice-balance/router"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/swagger"
)

// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
func main() {
	// Start a new fiber app
	app := fiber.New()
	app.Get("/swagger/*", swagger.HandlerDefault) // Route to Swagger UI

	// Connect to the Database
	database.ConnectDB()
	app.Use(cors.New())

	// Setup the router
	router.SetupRoutes(app)

	// Listen on PORT 3000
	app.Listen(":8000")
}
