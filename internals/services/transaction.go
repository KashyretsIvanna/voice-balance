package services

import (
	models "github.com/KashyretsIvanna/voice-balance/internals/model"
	repositories "github.com/KashyretsIvanna/voice-balance/internals/repositories"
)

func CreateTransaction(transaction *models.Transaction) error {
	return repositories.SaveTransaction(transaction)
}
