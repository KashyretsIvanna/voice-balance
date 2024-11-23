package repositories

import (
	"github.com/KashyretsIvanna/voice-balance/internals/model"
	"gorm.io/gorm"
)

// AddCategory saves a new category to the database
func AddCategory(db *gorm.DB, category *model.Category) error {
	return db.Create(category).Error
}
