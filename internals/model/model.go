package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	// Adds some metadata fields to the table
	ID           uuid.UUID  `gorm:"type:uuid;primary_key"` // Use PostgreSQL's uuid_generate_v4 function
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	DeletedAt    *time.Time `json:"deleted_at,omitempty" gorm:"index"` // Soft delete
	Email        string     `json:"email" gorm:"unique;not null"`      // Email field with JSON and unique constraint
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Password     string     `gorm:"not null"` // Only for email/password login
	RefreshToken string     `gorm:""`         // To store the refresh token
}

type Reminder struct {
	// Adds some metadata fields to the table
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"` // Use PostgreSQL's uuid_generate_v4 function
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"` // Soft delete
	Title       string     `gorm:"size:100;not null"`
	Amount      float64    `gorm:"not null"`
	DueDate     time.Time  `gorm:"not null"`
	IsCompleted bool       `gorm:"default:false"`
	UserID      uuid.UUID  `gorm:"not null"` // Foreign key to User

}

type Transaction struct {
	// Adds some metadata fields to the table
	ID          uuid.UUID  `gorm:"type:uuid;primary_key"` // Use PostgreSQL's uuid_generate_v4 function
	Amount      float64    `gorm:"not null"`
	Description string     `gorm:"size:255"`
	Date        time.Time  `gorm:"not null"`
	UserID      uuid.UUID  `gorm:"not null"` // Foreign key to User
	Category   	Category   `gorm:"foreignKey:CategoryID"`
	CategoryID  uuid.UUID  `gorm:"not null"` // Foreign key to Category
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `json:"deleted_at,omitempty" gorm:"index"` // Soft delete
}

type Category struct {
	// Adds some metadata fields to the table
	ID        uuid.UUID  `gorm:"type:uuid;primary_key"` // Use PostgreSQL's uuid_generate_v4 function
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty" gorm:"index"` // Soft delete
	Name      string     `gorm:"size:100;not null;unique"`
	Type      string     `gorm:"size:20;not null"` // 'income' or 'expense'
	UserID    uuid.UUID  `gorm:"not null"`         // Foreign key to User
}

func (category *Category) BeforeCreate(tx *gorm.DB) (err error) {
	if category.ID == uuid.Nil {
		category.ID = uuid.New() // Generate a new UUID
	}
	return
}

func (user *User) BeforeCreate(tx *gorm.DB) (err error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New() // Generate a new UUID
	}
	return
}

func (transaction *Transaction) BeforeCreate(tx *gorm.DB) (err error) {
	if transaction.ID == uuid.Nil {
		transaction.ID = uuid.New() // Generate a new UUID
	}
	return
}

func (reminder *Reminder) BeforeCreate(tx *gorm.DB) (err error) {
	if reminder.ID == uuid.Nil {
		reminder.ID = uuid.New() // Generate a new UUID
	}
	return
}
