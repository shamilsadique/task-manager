package models

import "gorm.io/gorm"


type User struct {
	gorm.Model
	Name string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Role string `json:"role"`
}

type LoginInput struct {
	Email string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type ResetInput struct {
	Current       string `json:"currentpass" binding:"required"`
	New 		string `json:"newpass" binding:"required,min=8"`
}

type AdminUserResponse struct {
    ID    uint   `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
    Role  string `json:"role"`
}