# Kratos Gin 代码生成器

一个基于 Kratos 框架和 Gin 路由的 API 代码生成器，支持自定义模板语法和验证规则。

## 功能特性

- 🚀 **模板驱动**: 使用 `.gin` 模板文件定义 API，类似 go-zero 的 `.api` 文件
- 🎯 **方法级控制**: 支持在方法级别选择是否在 context 中传递 `gin.Context`
- 🔧 **自定义验证**: 支持 Gin 的自定义验证规则，可扩展验证器
- 📦 **自动生成**: 自动生成类型定义、服务接口、HTTP 处理器等代码
- 🏗️ **Kratos 集成**: 完美集成 Kratos 框架的 HTTP 服务器
- 📁 **智能路径**: 自动从文件路径推断输出目录和包名


## 快速开始

### 1. 安装命令行工具

```bash
go install github.com/YuukiKazuto/kratosgin@latest
```

### 2. 创建模板文件

```bash
# 创建 user.gin 模板文件
kratosgin new user
```

### 3. 生成代码

```bash
# 生成基本 API 代码
kratosgin gen -f user.gin

# 生成 API 代码 + Service 实现
kratosgin gen -f user.gin -s internal/service

# 生成 API 代码 + Middleware 实现
kratosgin gen -f user.gin -m internal/middleware

# 生成所有代码
kratosgin gen -f user.gin -s internal/service -m internal/middleware
```

## 命令行工具

### 安装

```bash
go install github.com/YuukiKazuto/kratosgin@latest
```

### 命令说明

#### `kratosgin gen` - 生成代码

```bash
kratosgin gen [flags]
```

**参数：**
- `-f, --file string`: 指定 `.gin` 模板文件路径（必需）
- `-s, --service string`: 指定 Service 实现输出目录（可选）
- `-m, --middleware string`: 指定 Middleware 实现输出目录（可选）

**示例：**
```bash
# 只生成 API 代码
kratosgin gen -f api/user/v1/user.gin

# 生成 API 代码 + Service 实现
kratosgin gen -f api/user/v1/user.gin -s internal/service

# 生成 API 代码 + Middleware 实现
kratosgin gen -f api/user/v1/user.gin -m internal/middleware

# 生成所有代码
kratosgin gen -f api/user/v1/user.gin -s internal/service -m internal/middleware
```

#### `kratosgin new` - 创建模板

```bash
kratosgin new <name>
```

**参数：**
- `name`: 模板名称，会生成 `{name}.gin` 文件

**示例：**
```bash
# 创建 user.gin 模板文件
kratosgin new user

# 创建 product.gin 模板文件
kratosgin new product
```

## Gin 文件语法

### 基本结构

`.gin` 文件是 Kratos Gin 代码生成器的模板文件，采用类似 go-zero `.api` 文件的语法。一个完整的 `.gin` 文件包含以下部分：

```gin
info {
    title: "API 标题"
    version: "v1.0.0"
    desc: "API 描述"
}

options {
    outputDir: "api/user/v1"  // 输出目录，默认为当前目录
    packageName: "v1"         // 包名，默认为 "v1"
}

type RequestType {
    Field1 string `json:"field1" binding:"required"` // 字段注释
    Field2 int    `json:"field2" binding:"min=1"`
}

service ServiceName {
    @method POST /api/path RequestType ResponseType // 方法注释
    @method GET /api/path WithGinContext RequestType ResponseType // 使用 gin context
}
```

### 语法详解

#### 1. info 块
定义 API 的基本信息：
```gin
info {
    title: "用户服务 API"      // API 标题
    version: "v1.0.0"         // API 版本
    desc: "用户相关的 API 接口" // API 描述
}
```

#### 2. options 块
配置代码生成选项：
```gin
options {
    outputDir: "api/user/v1"  // 输出目录，相对于项目根目录
    packageName: "v1"         // 生成的包名
}
```

#### 3. type 定义
定义数据结构，支持两种格式：

**单个类型定义：**
```gin
type UserRequest {
    ID       int64  `json:"id" binding:"required,min=1"`           // 必填，最小值为1
    Username string `json:"username" binding:"required,username"`  // 必填，用户名格式
    Email    string `json:"email" binding:"required,email"`        // 必填，邮箱格式
    Age      int    `json:"age" binding:"min=0,max=150"`           // 年龄范围
}
```

**类型组定义：**
```gin
type (
    UserRequest {
        ID       int64  `json:"id" binding:"required,min=1"`
        Username string `json:"username" binding:"required,username"`
    }
    
    UserResponse {
        ID       int64  `json:"id"`
        Username string `json:"username"`
        Email    string `json:"email"`
    }
    
    ProductRequest {
        Name  string  `json:"name" binding:"required"`
        Price float64 `json:"price" binding:"required,min=0"`
    }
    
    ProductResponse {
        ID    int64   `json:"id"`
        Name  string  `json:"name"`
        Price float64 `json:"price"`
    }
)
```

**字段标签说明：**
- `json:"field_name"`: JSON 序列化标签
- `binding:"rules"`: Gin 验证规则，支持多个规则用逗号分隔

#### 4. service 定义
定义服务接口和路由：
```gin
service UserService {
    // 普通方法：@方法名 HTTP方法 路径 请求类型 响应类型
    @getUser GET /api/v1/user/:id GetUserRequest GetUserResponse
    
    // 带 Gin Context 的方法：@方法名 HTTP方法 路径 WithGinContext 请求类型 响应类型
    @createUser POST /api/v1/user WithGinContext CreateUserRequest CreateUserResponse
    
    // 路由组：@组名 HTTP方法 路径 请求类型 响应类型
    @admin {
        @getUsers GET /api/v1/admin/users GetUsersRequest GetUsersResponse
        @deleteUser DELETE /api/v1/admin/user/:id DeleteUserRequest DeleteUserResponse
    }
}
```

**方法定义格式：**
- **普通方法**: `@方法名 HTTP方法 路径 请求类型 响应类型`
- **带 Gin Context**: `@方法名 HTTP方法 路径 WithGinContext 请求类型 响应类型`
- **路由组**: 使用 `@组名 { ... }` 定义一组相关路由

#### 5. middleware 定义
定义中间件：
```gin
middleware {
    Auth    // 认证中间件
    Admin   // 管理员中间件
    Logging // 日志中间件
}
```

### 支持的 HTTP 方法

- `GET`: 获取资源
- `POST`: 创建资源
- `PUT`: 更新资源
- `DELETE`: 删除资源
- `PATCH`: 部分更新资源

### 路径参数

支持路径参数，使用 `:参数名` 格式：
```gin
@getUser GET /api/v1/user/:id GetUserRequest GetUserResponse
@updateUser PUT /api/v1/user/:id UpdateUserRequest UpdateUserResponse
```

### 验证规则

支持 Gin 的所有内置验证规则：

**常用验证规则：**
- `required`: 必填字段
- `min=n`: 最小值
- `max=n`: 最大值
- `len=n`: 长度
- `email`: 邮箱格式
- `url`: URL 格式
- `numeric`: 数字
- `alpha`: 字母
- `alphanum`: 字母数字

**自定义验证规则：**
- `mobile`: 手机号验证
- `password`: 密码强度验证
- `username`: 用户名验证
- `idcard`: 身份证号验证

### 注释

支持行内注释，使用 `//` 开头：
```gin
type UserRequest {
    ID int64 `json:"id" binding:"required"` // 用户ID，必填
    // 用户名，必填且符合用户名格式
    Username string `json:"username" binding:"required,username"`
}
```

## 实现示例

### Service 实现

当使用 `-s` 参数时，工具会生成 Service 实现模板：

```go
package service

import (
    "context"
    "github.com/go-kratos/kratos/v2/log"
    userV1 "your-project/api/user/v1"
)

// UserService UserService 服务实现
type UserService struct {
    log *log.Helper
}

func NewUserService(logger log.Logger) *UserService {
    return &UserService{
        log: log.NewHelper(logger),
    }
}

func (s *UserService) Login(ctx context.Context, req *userV1.LoginRequest) (*userV1.LoginResponse, error) {
    // 实现登录逻辑
    return &userV1.LoginResponse{
        Token: "jwt_token",
        User:  &userV1.User{ID: 1, Username: req.Username},
    }, nil
}

func (s *UserService) Register(ctx context.Context, req *userV1.LoginRequest) (*userV1.LoginResponse, error) {
    // 实现注册逻辑
    return &userV1.LoginResponse{
        Token: "jwt_token",
        User:  &userV1.User{ID: 2, Username: req.Username},
    }, nil
}

func (s *UserService) GetUser(ctx context.Context, req *userV1.GetUserRequest) (*userV1.GetUserResponse, error) {
    // 实现获取用户逻辑
    return &userV1.GetUserResponse{
        User: &userV1.User{ID: req.ID, Username: "test"},
    }, nil
}
```

### Middleware 实现

当使用 `-m` 参数时，工具会生成 Middleware 实现模板：

```go
package middleware

import "github.com/gin-gonic/gin"

type UserMiddleware struct {
}

func NewUserMiddleware() *UserMiddleware {
    return &UserMiddleware{}
}

func (m *UserMiddleware) Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 实现认证逻辑
        c.Next()
    }
}

func (m *UserMiddleware) Admin() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 实现管理员权限检查逻辑
        c.Next()
    }
}
```

## 生成的文件

代码生成器会根据 `.gin` 文件内容生成以下文件：

### API 文件（总是生成）
- `types.go`: 类型定义，包含所有 `type` 块中定义的结构体
- `service.go`: 服务接口，包含所有 `service` 块中定义的方法
- `handlers.go`: HTTP 处理器，包含路由注册和请求处理逻辑
- `ginutil.go`: Gin Context 工具（仅当使用了 `WithGinContext` 时生成）

### Service 实现文件（使用 `-s` 参数时生成）
- `{service_name}.go`: Service 实现模板，包含结构体定义和空方法实现

### Middleware 实现文件（使用 `-m` 参数时生成）
- `{middleware_name}.go`: Middleware 实现模板，包含结构体定义和空方法实现

## 自定义验证器

在 `internal/validate/validator.go` 中注册自定义验证规则：

```go
func registerCustomValidators(v *validator.Validate) {
    v.RegisterValidation("custom", validateCustom)
}

func validateCustom(fl validator.FieldLevel) bool {
    // 自定义验证逻辑
    return true
}
```

## 开发指南

### 构建项目

```bash
# 构建 kratosgin 工具
go build -o kratosgin main.go

# 或者使用 make
make build
```

### 测试

```bash
# 运行测试
go test ./...

# 测试代码生成
kratosgin new test-user
kratosgin gen -f test-user.gin -s internal/service -m internal/middleware
```

### 项目结构

```
kratosgin/
├── main.go                    # 主入口文件
├── go.mod                     # Go 模块文件
├── internal/
│   ├── cli/                   # 命令行接口
│   │   └── commands.go        # Cobra 命令定义
│   ├── generator/             # 代码生成器
│   │   ├── code_generator.go  # 主生成器
│   │   ├── service_generator.go # Service 生成器
│   │   ├── middleware_generator.go # Middleware 生成器
│   │   ├── simple_handlers.go # Handler 生成器
│   │   └── templates/         # 代码模板
│   │       ├── types.tmpl
│   │       ├── service.tmpl
│   │       ├── handlers.tmpl
│   │       └── ginutil.tmpl
│   ├── parser/                # 模板解析器
│   │   └── gin_parser.go      # .gin 文件解析
│   └── templates/             # 模板文件
│       └── new_template.gin   # 新模板生成器
└── README.md
```

## 完整示例

以下是一个完整的 `.gin` 文件示例，展示了所有语法特性：

```gin
info {
    title: "电商系统 API"
    version: "v1.0.0"
    desc: "电商系统的用户和商品管理 API"
}

options {
    outputDir: "api/ecommerce/v1"
    packageName: "v1"
}

// 用户相关类型（使用类型组语法）
type (
    User {
        ID       int64  `json:"id"`
        Username string `json:"username"`
        Email    string `json:"email"`
        Phone    string `json:"phone"`
    }
    
    CreateUserRequest {
        Username string `json:"username" binding:"required,username"`
        Email    string `json:"email" binding:"required,email"`
        Phone    string `json:"phone" binding:"required,mobile"`
        Password string `json:"password" binding:"required,password"`
    }
    
    CreateUserResponse {
        User  User   `json:"user"`
        Token string `json:"token"`
    }
    
    GetUserRequest {
        ID int64 `json:"id" binding:"required,min=1"`
    }
    
    GetUserResponse {
        User User `json:"user"`
    }
)

// 商品相关类型（使用类型组语法）
type (
    Product {
        ID          int64   `json:"id"`
        Name        string  `json:"name"`
        Description string  `json:"description"`
        Price       float64 `json:"price"`
        Stock       int     `json:"stock"`
    }
    
    CreateProductRequest {
        Name        string  `json:"name" binding:"required,min=1,max=100"`
        Description string  `json:"description" binding:"max=500"`
        Price       float64 `json:"price" binding:"required,min=0"`
        Stock       int     `json:"stock" binding:"required,min=0"`
    }
    
    CreateProductResponse {
        Product Product `json:"product"`
    }
    
    GetProductRequest {
        ID int64 `json:"id" binding:"required,min=1"`
    }
    
    GetProductResponse {
        Product Product `json:"product"`
    }
)

// 用户服务
service UserService {
    @createUser POST /api/v1/users WithGinContext CreateUserRequest CreateUserResponse
    @getUser    GET  /api/v1/users/:id GetUserRequest GetUserResponse
}

// 商品服务
service ProductService {
    @createProduct POST /api/v1/products CreateProductRequest CreateProductResponse
    @getProduct    GET  /api/v1/products/:id GetProductRequest GetProductResponse
    
    // 管理员路由组
    @admin {
        @deleteProduct DELETE /api/v1/admin/products/:id GetProductRequest GetProductResponse
    }
}

// 中间件定义
middleware {
    Auth    // 认证中间件
    Admin   // 管理员权限中间件
    Logging // 日志中间件
}
```

## 许可证

MIT License
