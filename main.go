package main

import (
	"fmt"
	"os"
	"time"
	"welcome/config"
	"welcome/middleware"
	"welcome/models"
	"welcome/utils"

	"github.com/joho/godotenv"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"

	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
)

func generateResetToken() (string, string, error) {
	bytes := make([]byte, 32)

	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	token := hex.EncodeToString(bytes)

	hash := sha256.Sum256([]byte(token))
	tokenHash := hex.EncodeToString(hash[:])

	return token, tokenHash, nil
}



func main() {
	if err := godotenv.Load(); err != nil {
        fmt.Println("Warning: .env file not found")
    }

    router := gin.Default()

	config.ConnectDatabase()

	secret := os.Getenv("JWT_SECRET")

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

		user.Role = "user"

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
			"role":    user.Role,
			"exp":     time.Now().Add(time.Hour * 72).Unix(),
		}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString([]byte(secret))
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

	router.GET("/info", middleware.AuthMiddleware(), func(c *gin.Context) {
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(401, gin.H{
				"error": "User ID not found in token.",
			})
			return
		}
		var user models.User
		if err := config.DB.Where("id = ?", userID).First(&user).Error; err != nil {
			c.JSON(404, gin.H{
				"error": "User not found.",
			})
			return
		}

		var completedTasksCount int64
		if err := config.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "Completed").Count(&completedTasksCount).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to retrieve tasks count.",
			})
			return
		}
		var pendingTasksCount int64
		if err := config.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "Pending").Count(&pendingTasksCount).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to retrieve tasks count.",
			})
			return
		}
		var inProgressTasksCount int64
		if err := config.DB.Model(&models.Task{}).Where("user_id = ? AND status = ?", userID, "In Progress").Count(&inProgressTasksCount).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to retrieve tasks count.",
			})
			return
		}

		c.JSON(200, gin.H{
			"username":          user.Name,
			"email":             user.Email,
			"tasks_completed":   completedTasksCount,
			"tasks_pending":     pendingTasksCount,
			"tasks_in_progress": inProgressTasksCount,
		})
	})

	router.POST("/reset-password", middleware.AuthMiddleware(), func(c *gin.Context) {
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

		var input models.ResetInput
		if err := c.BindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		var user models.User
		if err := config.DB.First(&user, userIDUint).Error; err != nil {
			c.JSON(404, gin.H{
				"error": "User not found.",
			})
			return
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(user.Password),
			[]byte(input.Current),
		); err != nil {
			c.JSON(401, gin.H{
				"error": "Current password is incorrect.",
			})
			return
		}

		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(input.New),
			bcrypt.DefaultCost,
		)
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to hash new password.",
			})
			return
		}

		user.Password = string(hashedPassword)
		if err := config.DB.Save(&user).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to update password.",
			})
			return
		}

		c.JSON(200, gin.H{
			"message": "Password updated successfully.",
		})
	})

	router.GET("/admin/test", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "Welcome to the admin area!",
		})
	})

	router.GET("/admin/users", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) {
		var users []models.User
		if err := config.DB.Find(&users).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to retrieve users.",
			})
			return
		}
		userResponses := make([]models.AdminUserResponse, 0, len(users))
		for _, user := range users {
			userResponses = append(userResponses, models.AdminUserResponse{
				ID:    user.ID,
				Name:  user.Name,
				Email: user.Email,
				Role:  user.Role,
			})
		}
		c.JSON(200, gin.H{
			"users": userResponses,
		})
	})

	router.GET("/admin/users/:id/tasks", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) {
		userID := c.Param("id")
		var tasks []models.Task
		if err := config.DB.Where("user_id = ?", userID).Find(&tasks).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to retrieve tasks.",
			})
			return
		}
		c.JSON(200, gin.H{
			"tasks": tasks,
		})
	})

	router.PUT("/admin/tasks/:id/status", middleware.AuthMiddleware(), middleware.AdminMiddleware(), func(c *gin.Context) {
		taskID := c.Param("id")
		var task models.Task
		if err := config.DB.First(&task, taskID).Error; err != nil {
			c.JSON(404, gin.H{
				"error": "Task not found.",
			})
			return
		}
		var input struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.BindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}
		if input.Status != "Pending" && input.Status != "In Progress" && input.Status != "Completed" {
			c.JSON(400, gin.H{
				"error": "Invalid status. Allowed values are: Pending, In Progress, Completed.",
			})
			return
		}
		task.Status = input.Status
		if err := config.DB.Save(&task).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to update task status.",
			})
			return
		}
		c.JSON(200, gin.H{
			"message": "Task status updated successfully.",
			"task":    task,
		})
	})

	router.POST("/admin/tasks",middleware.AuthMiddleware(),middleware.AdminMiddleware(),func(c *gin.Context) {
		var input models.TaskInput
		if err := c.BindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": err.Error(),
			})
			return
		}

		var user models.User
		if err := config.DB.First(&user, input.UserID).Error; err != nil {
			c.JSON(404, gin.H{
				"error": "User not found.",
			})
			return
		}

		task := models.Task{
			Title:       input.Title,
			Description: input.Description,
			Status:      "Pending",
			DueDate:     input.DueDate,
			UserID:      input.UserID,
		}
		if err := config.DB.Create(&task).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to create task.",
			})
			return
		}
		c.JSON(201, gin.H{
			"message": "Task created and assigned successfully.",
			"task":    task,
		})
	})

	router.POST("/forgot-password", func(c *gin.Context) {

		var input struct {
			Email string `json:"email" binding:"required,email"`
		}
	
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": "Please enter a valid email address.",
			})
			return
		}
	
		var user models.User
	
		if err := config.DB.Where("email = ?", input.Email).First(&user).Error; err != nil {
			c.JSON(200, gin.H{
				"message": "If an account with that email exists, a password reset link has been sent.",
			})
			return
		}
	
		config.DB.
			Where("user_id = ?", user.ID).
			Delete(&models.PasswordReset{})
	
		token, tokenHash, err := generateResetToken()
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to generate reset token.",
			})
			return
		}
	
		reset := models.PasswordReset{
			UserID:    user.ID,
			TokenHash: tokenHash,
			ExpiresAt: time.Now().Add(15 * time.Minute),
		}
	
		if err := config.DB.Create(&reset).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to create password reset request.",
			})
			return
		}
	

		resetLink := os.Getenv("FRONTEND_URL") +"/reset-password?token=" + token
		if err := utils.SendPasswordResetEmail(
			user.Email,
			resetLink,
		); err != nil {

			fmt.Println("Failed to send password reset email:", err)

			c.JSON(500, gin.H{
				"error": "Failed to send password reset email.",
			})
			return
		}
	
		c.JSON(200, gin.H{
			"message": "a password reset link has been sent to your mail.",
		})
	})

	router.POST("/reset-password-token", func(c *gin.Context) {

		var input struct {
			Token    string `json:"token" binding:"required"`
			Password string `json:"password" binding:"required,min=8"`
		}
	
		if err := c.ShouldBindJSON(&input); err != nil {
			c.JSON(400, gin.H{
				"error": "Password must be at least 8 characters.",
			})
			return
		}
	
		tokenHashBytes := sha256.Sum256([]byte(input.Token))
		tokenHash := hex.EncodeToString(tokenHashBytes[:])
	
		var reset models.PasswordReset
	
		if err := config.DB.
			Where("token_hash = ?", tokenHash).
			First(&reset).Error; err != nil {
	
			c.JSON(400, gin.H{
				"error": "Invalid or expired reset link.",
			})
			return
		}
	
		if time.Now().After(reset.ExpiresAt) {
	
			config.DB.Delete(&reset)
	
			c.JSON(400, gin.H{
				"error": "Invalid or expired reset link.",
			})
			return
		}
	
		var user models.User
	
		if err := config.DB.First(&user, reset.UserID).Error; err != nil {
			c.JSON(404, gin.H{
				"error": "User not found.",
			})
			return
		}
	
		hashedPassword, err := bcrypt.GenerateFromPassword(
			[]byte(input.Password),
			bcrypt.DefaultCost,
		)
	
		if err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to hash password.",
			})
			return
		}
	
		user.Password = string(hashedPassword)
	
		if err := config.DB.Save(&user).Error; err != nil {
			c.JSON(500, gin.H{
				"error": "Failed to update password.",
			})
			return
		}
	
		config.DB.Delete(&reset)
	
		c.JSON(200, gin.H{
			"message": "Password reset successfully.",
		})
	})

	router.Run(":8080")
}