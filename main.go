package main

import (
	"fmt"
	"time"
	"welcome/config"
	"welcome/middleware"
	"welcome/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	config.ConnectDatabase()
	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		AllowCredentials: true,
	}))

	router.POST("/register", func(c *gin.Context) {
		var user models.User
		if err := c.BindJSON(&user); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			firstError := validationErrors[0]
			switch firstError.Field() {
			case "Name":
				c.JSON(400, gin.H{
					"error": "Name is required.",
				})
			case "Email":
				c.JSON(400, gin.H{
					"error": "Please enter a valid email address.",
				})
			case "Password":
				switch firstError.Tag() {
				case "required":
					c.JSON(400, gin.H{
						"error": "Password is required.",
					})
				case "min":
					c.JSON(400, gin.H{
						"error": "Password must be at least 8 characters.",
					})
				default:
					c.JSON(400, gin.H{
						"error": err.Error(),
					})
				}
			default:
				c.JSON(400, gin.H{
					"error": err.Error(),
				})
			}
			return
		}
		var existingUser models.User
		if err := config.DB.Where("email = ?", user.Email).First(&existingUser).Error; err == nil {
			c.JSON(409, gin.H{
				"error": "Email already exists.",
			})
			return
		}
		hashedPassword, err := bcrypt.GenerateFromPassword(([]byte)(user.Password), bcrypt.DefaultCost)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to hash password.",
			})
			return
		}
		user.Password = string(hashedPassword)

		if err := config.DB.Create(&user).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to create user.",
			})
			return
		}
		c.JSON(201, gin.H{
			"message": "User registered successfully.",
			"user":    user,
		})

	})

	router.POST("/login", func(c *gin.Context) {
		var logindata models.LoginInput
		if err := c.BindJSON(&logindata); err != nil {
			validationErrors := err.(validator.ValidationErrors)
			firstError := validationErrors[0]
			switch firstError.Field() {
			case "Email":
				switch firstError.Tag() {
				case "required":
					c.JSON(400, gin.H{
						"error": "Email is required.",
					})
				case "email":
					c.JSON(400, gin.H{
						"error": "Please enter a valid email address.",
					})
				}
			case "Password":
				switch firstError.Tag() {
				case "required":
					c.JSON(400, gin.H{
						"error": "Password is required.",
					})
				default:
					c.JSON(400, gin.H{
						"error": err.Error(),
					})
				}
			default:
				c.JSON(400, gin.H{
					"error": err.Error(),
				})
			}
			return
		}
		var user models.User
		result := config.DB.Where("email = ?", logindata.Email).First(&user)
		if result.Error != nil {
			c.JSON(401, gin.H{
				"error": "Invalid email or password.",
			})
			return
		}
		err := bcrypt.CompareHashAndPassword(([]byte)(user.Password), ([]byte)(logindata.Password))
		if err != nil {
			c.JSON(401, gin.H{
				"error": "Invalid email or password.",
			})
			return
		}
		claims := jwt.MapClaims{
			"user_id": user.ID,
			"email":   user.Email,
			"exp":     time.Now().Add(time.Hour * 72).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte("my-secret-key"))
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to generate token.",
			})
			return
		}
		c.JSON(200, gin.H{
			"login": "login successful",
			"token": tokenString,
		})
	})

	router.GET("/dashboard", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"error": "User ID not found in token.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Welcome to your dashboard!",
			"user_id": userID,
		})

	})

	router.POST("/tasks", middleware.AuthMiddleware(), func(c *gin.Context) {
		var input models.TaskInput
		if err := c.BindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"error": "User ID not found in token.",
			})
			return
		}
		userIDUint, ok := userID.(uint)
		if !ok {
			c.JSON(500, gin.H{
				"error": "Invalid user ID type.",
			})
			return
		}
		task := models.Task{
			Title:       input.Title,
			Description: input.Description,
			Status:      "Pending",
			DueDate:     input.DueDate,
			UserID:      userIDUint,
		}
		if err := config.DB.Create(&task).Error; err != nil {
			fmt.Println("Error creating task:", err)
			c.JSON(500, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(201, gin.H{
			"message": "Task created successfully.",
			"task":    task,
		})

	})

	router.GET("/tasks", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"error": "User ID not found in token.",
			})
			return
		}
		userIDUint, ok := userID.(uint)
		if !ok {
			c.JSON(500, gin.H{
				"error": "Invalid user ID type.",
			})
			return
		}
		var tasks []models.Task
		if err := config.DB.Where("user_id = ?", userIDUint).Find(&tasks).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to retrieve tasks.",
			})
			return
		}
		c.JSON(200, gin.H{
			"tasks": tasks,
		})

	})

	router.GET("/tasks/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"error": "User ID not found in token.",
			})
			return
		}
		userIDUint, ok := userID.(uint)
		if !ok {
			c.JSON(500, gin.H{
				"error": "Invalid user ID type.",
			})
			return
		}
		taskID := c.Param("id")
		var task models.Task
		if err := config.DB.Where("id = ? AND user_id = ?", taskID, userIDUint).First(&task).Error; err != nil {
			c.JSON(404, gin.H{
				"error": "Task not found.",
			})
			return
		}
		c.JSON(200, gin.H{
			"task": task,
		})
	})

	router.PUT("/tasks/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"error": "User ID not found in token.",
			})
			return
		}
		userIDUint, ok := userID.(uint)
		if !ok {
			c.JSON(500, gin.H{
				"error": "Invalid user ID type.",
			})
			return
		}
		taskID := c.Param("id")
		var task models.Task
		if err := config.DB.Where("id = ? AND user_id = ?", taskID, userIDUint).First(&task).Error; err != nil {
			c.JSON(404, gin.H{
				"error": "Task not found.",
			})
			return
		}
		var input models.TaskInput
		if err := c.BindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		task.Title = input.Title
		task.Description = input.Description
		task.DueDate = input.DueDate

		if err := config.DB.Save(&task).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to update task.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Task updated successfully.",
			"task":    task,
		})

	})

	router.DELETE("/tasks/:id", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"error": "User ID not found in token.",
			})
			return
		}
		userIDUint, ok := userID.(uint)
		if !ok {
			c.JSON(500, gin.H{
				"error": "Invalid user ID type.",
			})
			return
		}
		taskID := c.Param("id")
		var task models.Task
		if err := config.DB.Where("id = ? AND user_id = ?", taskID, userIDUint).First(&task).Error; err != nil {
			c.JSON(404, gin.H{
				"error": "Task not found.",
			})
			return
		}
		if err := config.DB.Delete(&task).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to delete task.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Task deleted successfully.",
		})

	})

	router.Run(":8080")
}
