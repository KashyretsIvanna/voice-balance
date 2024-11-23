package handlers

import (
	"github.com/KashyretsIvanna/voice-balance/database"
	"github.com/KashyretsIvanna/voice-balance/internals/model"
	"github.com/KashyretsIvanna/voice-balance/internals/repositories"
	"github.com/google/uuid"

	"github.com/gofiber/fiber/v2"
)

// AddCategoryHandler godoc
// @Summary Add a new category
// @Description Add a new category to the database
// @Tags categories
// @Accept json
// @Produce json
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Param category body model.Category true "Category to add"
// @Success 201 {object} model.Category
// @Failure 400 {object} map[string]string "Invalid input"
// @Failure 500 {object} map[string]string "Failed to add category"
// @Router /api/categories [post]
func AddCategoryHandler(c *fiber.Ctx) error {
	db := database.DB

	// Retrieve user ID from context and handle potential errors
	userID, ok := c.Locals("ID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User ID not found in context",
		})
	}

	// Parse JSON body into the Category struct
	category := new(model.Category)
	if err := c.BodyParser(category); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	// Set the UserID for the category
	category.UserID = userID

	// Save the category using the repository function
	if err := repositories.AddCategory(db, category); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to add category",
		})
	}

	// Return the created category
	return c.Status(fiber.StatusCreated).JSON(category)
}

// GetCategoriesByUserID retrieves categories for the authenticated user
// @Description Get categories for the authenticated user
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Tags categories
// @Accept json
// @Produce json
// @Success 200 {array} model.Category
// @router /api/categories [get]
func GetCategoriesByUserID(c *fiber.Ctx) error {
	db := database.DB
	userID, ok := c.Locals("ID").(uuid.UUID)
	if !ok {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "User ID not found in context",
		})
	}

	var categories []model.Category

	// Retrieve categories by UserID
	if err := db.Where("user_id = ?", userID).Find(&categories).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to retrieve categories",
		})
	}

	// If no categories are found, you can return an empty array or a 404 status
	if len(categories) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"status":  "error",
			"message": "No categories found for this user",
			"data":    nil,
		})
	}

	// Return the categories
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Categories found",
		"data":    categories,
	})
}
