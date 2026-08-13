package models
import "time"

type TaskInput struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	DueDate     time.Time `json:"due_date"`
	UserID      uint      `json:"user_id"`
}
