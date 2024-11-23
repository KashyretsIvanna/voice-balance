package userHandler

import (
	"github.com/KashyretsIvanna/voice-balance/database"
	"github.com/KashyretsIvanna/voice-balance/internals/model"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// GetUsers func gets all existing users
// @Description Get all existing users
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {array} model.User
// @Router /api/user [get]
func GetUsers(c *fiber.Ctx) error {
	db := database.DB
	var users []model.User

	// Find all users in the database
	db.Find(&users)

	// If no user is present return an error
	if len(users) == 0 {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "No users present", "data": nil})
	}

	// Else return users
	return c.JSON(fiber.Map{"status": "success", "message": "Users Found", "data": users})
}

// GetUser func gets one user by ID
// @Description Get one user by ID
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} model.User
// @Router /api/user/{id} [get]
func GetUser(c *fiber.Ctx) error {
	db := database.DB
	var user model.User

	// Read the user ID from the URL parameter
	id := c.Params("userId")

	// Find the user with the given ID
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	// Return the user
	return c.JSON(fiber.Map{"status": "success", "message": "User Found", "data": user})
}

// CreateUser func creates a user
// @Description Create a User
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Tags Users
// @Accept json
// @Produce json
// @Param name body string true "Name"
// @Param email body string true "Email"
// @Success 200 {object} model.User
// @Router /api/user [post]
func CreateUser(c *fiber.Ctx) error {
	db := database.DB
	user := new(model.User)

	// Parse the request body into the User struct
	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{"status": "error", "message": "Invalid input", "data": err})
	}

	// Add a UUID to the user
	user.ID = uuid.New()

	// Create the User and return error if encountered
	if err := db.Create(&user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Could not create user", "data": err})
	}

	// Return the created user
	return c.JSON(fiber.Map{"status": "success", "message": "Created User", "data": user})
}

// DeleteUser deletes a user by ID
// @Description Delete a user by ID
// @Param Authorization header string true "Insert your access token" default(Bearer <Add access token here>)
// @Tags user
// @Accept json
// @Produce json
// @Success 200
// @Router /api/user/{id} [delete]
func DeleteUser(c *fiber.Ctx) error {
	db := database.DB
	var user model.User

	// Read the user ID from the URL parameter
	id := c.Params("userId")

	// Find the user with the given ID
	if err := db.First(&user, "id = ?", id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"status": "error", "message": "User not found", "data": nil})
	}

	// Delete the user
	if err := db.Delete(&user, "id = ?", id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"status": "error", "message": "Failed to delete user", "data": err})
	}

	// Return success message
	return c.JSON(fiber.Map{"status": "success", "message": "User Deleted"})
}