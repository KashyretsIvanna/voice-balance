package handlers

import (
	"fmt"

	services "github.com/KashyretsIvanna/voice-balance/internals/services"
	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
)

// GetStatisticsByCategory godoc
// @Summary      Get statistics by category
// @Description  Returns income and expense statistics by category and date range
// @Tags         statistics
// @Produce      json
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Param        start_date   query     string false  "Start Date (YYYY-MM-DD)"
// @Param        end_date     query     string false  "End Date (YYYY-MM-DD)"
// @Success      200          {array}   model.Transaction
// @Failure      400          {object}  interface{} // This tells Swagger to expect any object as a response
// @Failure      500          {object}  interface{}
// @Router       /api/statistics/category [get]
func GetStatisticsByCategory(c *fiber.Ctx) error {
	userID, ok := c.Locals("ID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User ID not found in context",
		})
	}
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	fmt.Print(endDate)
	stats, err := services.GetStatistics(userID, startDate, endDate)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(stats)
}
