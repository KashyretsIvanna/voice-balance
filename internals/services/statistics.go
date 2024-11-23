package services

import (

	models "github.com/KashyretsIvanna/voice-balance/internals/model"
	repositories "github.com/KashyretsIvanna/voice-balance/internals/repositories"
	"github.com/google/uuid"
)

func GetStatistics(userID uuid.UUID, startDate, endDate string) ([]models.Transaction, error) {
	return repositories.FindTransactionsByCategoryAndDate(userID, startDate, endDate)
}
