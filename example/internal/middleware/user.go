package middleware

import (
	userV1 "example/api/user/v1"
	"github.com/gin-gonic/gin"
)

type UserMiddleware struct {
}

func NewUserMiddleware() userV1.Middleware {
	return &UserMiddleware{}
}

func (m *UserMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现 Auth 中间件逻辑
		c.Next()
	}
}

func (m *UserMiddleware) Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现 Logging 中间件逻辑
		c.Next()
	}
}

func (m *UserMiddleware) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// TODO: 实现 Admin 中间件逻辑
		c.Next()
	}
}
