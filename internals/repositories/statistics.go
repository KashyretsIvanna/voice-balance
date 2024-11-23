package repositories

import (

	"github.com/KashyretsIvanna/voice-balance/database"
	models "github.com/KashyretsIvanna/voice-balance/internals/model"
	"github.com/google/uuid"
)

func FindTransactionsByCategoryAndDate(userID uuid.UUID, startDate, endDate string) ([]models.Transaction, error) {
	db := database.DB

	var transactions []models.Transaction
	err := db.Where("user_id = ? AND date BETWEEN ? AND ?", userID, startDate, endDate).
		Find(&transactions).Error
	return transactions, err
}
