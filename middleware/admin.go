package middleware

import "github.com/gin-gonic/gin"

func AdminMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {

        role, exists := c.Get("role")

        if !exists {
            c.JSON(401, gin.H{
                "error": "Role not found.",
            })
            c.Abort()
            return
        }

        if role != "admin" {
            c.JSON(403, gin.H{
                "error": "Admin access required.",
            })
            c.Abort()
            return
        }

        c.Next()
    }
}