package handlers

import (
	"fmt"
	"time"

	"github.com/KashyretsIvanna/voice-balance/database"
	models "github.com/KashyretsIvanna/voice-balance/internals/model"
	services "github.com/KashyretsIvanna/voice-balance/internals/services"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// AddTransaction godoc
// @Summary      Add a new transaction
// @Description  Adds an income or expense transaction by category
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        transaction  body      models.Transaction  true  "Transaction Data"
// @Success      200          {object}  models.Transaction
// @Failure      400          {object}  interface{} // This tells Swagger to expect any object as a response
// @Failure      500          {object}  interface{}
// @Router       /api/transaction [post]
func AddTransaction(c *fiber.Ctx) error {
	transaction := new(models.Transaction)
	userID, ok := c.Locals("ID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User ID not found in context",
		})
	}

	// userIDStr := "74c508d6-3b65-4583-a962-95a06ff2eb5b" // Example UUID as string
	// userID, err := uuid.Parse(userIDStr)
	// if err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "Invalid user ID",
	// 	})
	// }
	transaction.UserID = userID

	if err := c.BodyParser(transaction); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	if err := services.CreateTransaction(transaction); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(transaction)
}

// GetTransactions godoc
// @Summary      Get transactions grouped by category
// @Description  Retrieve transactions by category and date range
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Tags         transactions
// @Accept       json
// @Produce      json
// @Param        categoryId query      string  false "Category ID"
// @Param        startDate  query      string  false "Start Date in YYYY-MM-DD format"
// @Param        endDate    query      string  false "End Date in YYYY-MM-DD format"
// @Success      200        {array}   models.TransactionGroupedByCategory
// @Failure      400        {object}  interface{} // This tells Swagger to expect any object as a response
// @Failure      500        {object}  interface{}
// @Router       /api/transaction [get]
func GetTransactions(c *fiber.Ctx) error {
	db := database.DB
	userID, ok := c.Locals("ID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User ID not found in context",
		})
	}
	// userIDStr := "74c508d6-3b65-4583-a962-95a06ff2eb5b" // Example UUID as string
	// userID, err := uuid.Parse(userIDStr)
	// if err != nil {
	// 	return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
	// 		"error": "Invalid user ID",
	// 	})
	// }

	categoryID := c.Query("categoryId") // Get the category ID from the query params
	startDate := c.Query("startDate")   // Get the start date from the query params
	endDate := c.Query("endDate")       // Get the end date from the query params

	var transactions []models.Transaction

	// Build the query
	query := db.Where("user_id = ?", userID)

	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}

	if startDate != "" {
		start, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid start date format: %s", err.Error()),
			})
		}
		query = query.Where("date >= ?", start)
	}

	if endDate != "" {
		end, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid end date format: %s", err.Error()),
			})
		}
		query = query.Where("date <= ?", end)
	}

	// Execute the query
	if err := query.Preload("Category").Find(&transactions).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	// Group transactions by category
	groupedTransactions := make(map[uuid.UUID][]models.Transaction)
	for _, transaction := range transactions {
		groupedTransactions[transaction.CategoryID] = append(groupedTransactions[transaction.CategoryID], transaction)
	}

	// Return the grouped transactions
	return c.JSON(fiber.Map{
		"status":      "success",
		"message":     "Transactions retrieved",
		"groupedData": groupedTransactions,
	})
}
