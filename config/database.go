package config
import (
	"fmt"
	"welcome/models"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB
func ConnectDatabase(){
	dsn := "root:@tcp(127.0.0.1:3306)/task_manager?charset=utf8mb4&parseTime=True&loc=Local"

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database!")
	}
	fmt.Println("✅ Connected to MySQL successfully!")
	DB = db
	DB.AutoMigrate(&models.User{}, &models.Task{}, &models.PasswordReset{})
}