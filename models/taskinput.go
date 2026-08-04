package models
import "time"

type TaskInput struct {
	Title       string    `json:"title" binding:"required"`
	Description string    `json:"description"`
	Status	  string    `json:"status"`
	DueDate     time.Time `json:"due_date"`
}
