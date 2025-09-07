package v1

import (
	"github.com/gin-gonic/gin"
	kgin "github.com/go-kratos/gin"
	"github.com/go-kratos/kratos/v2/log"
	"net/http"
)

// Middleware 中间件接口
type Middleware interface {
	Auth() gin.HandlerFunc
	Logging() gin.HandlerFunc
	Admin() gin.HandlerFunc
}

// UserServiceHandler UserService 处理器
type UserServiceHandler struct {
	log         *log.Helper
	middleware  Middleware
	UserService UserService
}

// NewUserServiceHandler 创建 UserService 处理器
func NewUserServiceHandler(logger log.Logger, middleware Middleware, userservice UserService) *UserServiceHandler {
	return &UserServiceHandler{
		log:         log.NewHelper(logger),
		middleware:  middleware,
		UserService: userservice,
	}
}

// RegisterRoutes 注册路由
func (h *UserServiceHandler) RegisterRoutes(r *gin.Engine) {
	PrefixGroup := r.Group("/v1")
	{
		PrefixGroup.Use(h.middleware.Auth())
		PrefixGroup.Use(h.middleware.Logging())
		PrefixGroup.GET("/users/:id", h.GetUser)
		PrefixGroup.POST("/users", h.CreateUser)
		PrefixGroup.PUT("/users/:id", h.UpdateUser)
		PrefixGroup.DELETE("/users/:id", h.DeleteUser)

		AdminGroup := PrefixGroup.Group("/admin")
		{
			AdminGroup.Use(h.middleware.Admin())
			AdminGroup.GET("/users", h.GetAllUsers)
			AdminGroup.DELETE("/users", h.BulkDeleteUsers)
		}

		PublicGroup := PrefixGroup.Group("/public")
		{
			PublicGroup.GET("/users/:id", h.GetPublicUser)
			PublicGroup.GET("/users/search", h.SearchUsers)
		}
	}
}

// GetUser
func (h *UserServiceHandler) GetUser(c *gin.Context) {
	req := &UserReq{}
	if err := c.ShouldBind(req); err != nil {
		h.log.Errorw("Struct", "UserServiceHandler", "method", "GetUser", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.UserService.GetUser(ctx, req)
	if err != nil {
		kgin.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// CreateUser
func (h *UserServiceHandler) CreateUser(c *gin.Context) {
	req := &CreateUserReq{}
	if err := c.ShouldBind(req); err != nil {
		h.log.Errorw("Struct", "UserServiceHandler", "method", "CreateUser", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.UserService.CreateUser(ctx, req)
	if err != nil {
		kgin.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// UpdateUser
func (h *UserServiceHandler) UpdateUser(c *gin.Context) {
	req := &UpdateUserReq{}
	if err := c.ShouldBind(req); err != nil {
		h.log.Errorw("Struct", "UserServiceHandler", "method", "UpdateUser", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.UserService.UpdateUser(ctx, req)
	if err != nil {
		kgin.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// DeleteUser
func (h *UserServiceHandler) DeleteUser(c *gin.Context) {
	req := &UserReq{}
	if err := c.ShouldBind(req); err != nil {
		h.log.Errorw("Struct", "UserServiceHandler", "method", "DeleteUser", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.UserService.DeleteUser(ctx, req)
	if err != nil {
		kgin.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetAllUsers
func (h *UserServiceHandler) GetAllUsers(c *gin.Context) {
	req := &UserReq{}
	if err := c.ShouldBind(req); err != nil {
		h.log.Errorw("Struct", "UserServiceHandler", "method", "GetAllUsers", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.UserService.GetAllUsers(ctx, req)
	if err != nil {
		kgin.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// BulkDeleteUsers
func (h *UserServiceHandler) BulkDeleteUsers(c *gin.Context) {
	req := &UserReq{}
	if err := c.ShouldBind(req); err != nil {
		h.log.Errorw("Struct", "UserServiceHandler", "method", "BulkDeleteUsers", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.UserService.BulkDeleteUsers(ctx, req)
	if err != nil {
		kgin.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetPublicUser
func (h *UserServiceHandler) GetPublicUser(c *gin.Context) {
	req := &UserReq{}
	if err := c.ShouldBind(req); err != nil {
		h.log.Errorw("Struct", "UserServiceHandler", "method", "GetPublicUser", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.UserService.GetPublicUser(ctx, req)
	if err != nil {
		kgin.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// SearchUsers
func (h *UserServiceHandler) SearchUsers(c *gin.Context) {
	req := &UserReq{}
	if err := c.ShouldBind(req); err != nil {
		h.log.Errorw("Struct", "UserServiceHandler", "method", "SearchUsers", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
		})
		return
	}

	ctx := c.Request.Context()
	resp, err := h.UserService.SearchUsers(ctx, req)
	if err != nil {
		kgin.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}
