package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"github.com/HYuCN666/kite/internal/auth"
)

// Auth 返回校验 Bearer token 的中间件。
func Auth(m *auth.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		parts := strings.SplitN(c.GetHeader("Authorization"), " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40100, "message": "unauthorized"})
			return
		}
		claims, err := m.Parse(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"code": 40100, "message": "unauthorized"})
			return
		}
		c.Set("claims", claims)
		c.Next()
	}
}
