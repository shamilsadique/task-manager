package middleware
import (
	"github.com/gin-gonic/gin"
	"strings"
	"github.com/golang-jwt/jwt/v5"
)
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(401, gin.H{"error": "Authorization header is required"})
			c.Abort()
			return
		}
		tokenString = strings.TrimPrefix(tokenString, "Bearer ")

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("my-secret-key"), nil
		})

		if err != nil {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid token claims"})
			c.Abort()
			return
		}
		userIDFloat,ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(401, gin.H{"error": "Invalid user ID in token claims",
			})
			c.Abort()
			return
		}
		userID := uint(userIDFloat)
		c.Set("user_id", userID)
		c.Next()
	}
}
