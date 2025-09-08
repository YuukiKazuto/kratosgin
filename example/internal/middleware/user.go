package middleware

import (
	userV1 "example/api/user/v1"

	"github.com/gin-gonic/gin"
	"github.com/go-kratos/kratos/v2/log"
)

type UserMiddleware struct {
	log *log.Helper
}

func NewUserMiddleware() userV1.Middleware {
	return &UserMiddleware{}
}

func (m *UserMiddleware) Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.log.Infof("调用 Auth 方法")
		// TODO: 实现 Auth 中间件逻辑
		c.Next()
	}
}

func (m *UserMiddleware) Logging() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.log.Infof("调用 Logging 方法")
		// TODO: 实现 Logging 中间件逻辑
		c.Next()
	}
}

func (m *UserMiddleware) Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		m.log.Infof("调用 Admin 方法")
		// TODO: 实现 Admin 中间件逻辑
		c.Next()
	}
}
