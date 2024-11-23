package repositories

import (
	"github.com/KashyretsIvanna/voice-balance/database"
	models "github.com/KashyretsIvanna/voice-balance/internals/model"
)


func SaveTransaction(transaction *models.Transaction) error {
	db := database.DB

	return db.Create(transaction).Error
}
