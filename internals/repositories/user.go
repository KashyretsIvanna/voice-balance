package repositories

import (
	"errors"
	"github.com/KashyretsIvanna/voice-balance/database"
	"github.com/KashyretsIvanna/voice-balance/internals/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AddUser saves a new user to the database
func AddUser(user *model.User) error {
	db := database.DB

	user.ID = uuid.New()

	return db.Create(user).Error
}

// GetUserByEmail tries to find a user by email.
// Returns the user if found, or nil if not found or an error occurs.
func GetUserByEmail(email string) (*model.User, error) {
	user := &model.User{}
	DB := database.DB
	if err := DB.Where("email = ?", email).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

