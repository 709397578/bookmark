package middleware

import (
	"pintree-backend/config"
	"pintree-backend/utils"
	"strings"

	"github.com/gin-gonic/gin"
)

// OptionalAuthMiddleware 可选认证中间件
// 如果提供了token，则验证并设置用户信息；如果没有提供token，则继续执行
func OptionalAuthMiddleware(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")

		// 如果没有提供认证头，直接继续执行
		if authHeader == "" {
			c.Next()
			return
		}

		// Bearer Token格式: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// 格式错误，但不阻止请求继续执行
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := utils.ValidateToken(tokenString, cfg)
		if err != nil {
			// token无效，但不阻止请求继续执行
			c.Next()
			return
		}

		// 将用户信息存储到上下文
		c.Set("userID", claims.UserID)
		c.Set("userEmail", claims.Email)
		c.Set("userRole", claims.Role)

		c.Next()
	}
}
