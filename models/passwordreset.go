package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordReset struct {
	gorm.Model
	UserID    uint      `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"-"`
}