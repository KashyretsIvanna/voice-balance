package repositories

import (
	"errors"

	"github.com/KashyretsIvanna/voice-balance/database"
	"github.com/KashyretsIvanna/voice-balance/internals/model"
	"gorm.io/gorm"
)

// SaveRefreshToken saves a refresh token for a user in the database
func SaveRefreshToken(email, refreshToken string) error {
	user := &model.User{}
	DB := database.DB

	if err := DB.Where("email = ?", email).First(user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return err
	}

	// Update or create refresh token
	user.RefreshToken = refreshToken
	return DB.Save(user).Error
}

// IsValidRefreshToken checks if a given refresh token is valid for the user
func IsValidRefreshToken(email, refreshToken string) bool {
	user := &model.User{}
	DB := database.DB
	if err := DB.Where("email = ?", email).First(user).Error; err != nil {
		return false
	}

	return user.RefreshToken == refreshToken
}

