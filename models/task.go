package models

import (
	"time"
	"gorm.io/gorm"
)

type Task struct {
	gorm.Model
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	DueDate     time.Time `json:"due_date"`
	UserID      uint
}
