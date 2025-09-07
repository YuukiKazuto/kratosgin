# Example 示例项目

这是一个使用 Kratos Gin 代码生成器的完整示例项目。

## 项目结构

```
example/
├── go.mod                     # Go 模块文件
├── api/                       # API 定义
│   └── user/                  # 用户服务
│       └── v1/                # v1 版本
│           └── user.gin       # 用户服务定义文件
├── internal/                  # 内部代码
│   ├── service/               # 服务实现
│   └── middleware/            # 中间件实现
└── README.md                  # 说明文档
```

## 使用方法

### 1. 生成 API 代码

```bash
# 生成基本的 API 代码（类型定义、服务接口、HTTP 处理器）
kratosgin gen -f api/user/v1/user.gin
```

### 2. 生成服务实现

```bash
# 生成 API 代码 + 服务实现
kratosgin gen -f api/user/v1/user.gin -s internal/service
```

### 3. 生成中间件实现

```bash
# 生成 API 代码 + 中间件实现
kratosgin gen -f api/user/v1/user.gin -m internal/middleware
```

### 4. 生成所有代码

```bash
# 生成完整的代码
kratosgin gen -f api/user/v1/user.gin -s internal/service -m internal/middleware
```

## 功能特性

这个示例展示了以下功能：

- **多种类型定义**：`type Name {}`、`type Name struct {}` 和 `type ()` 组语法
- **服务前缀**：使用 `prefix v1` 设置服务前缀
- **服务级中间件**：`auth` 和 `logging` 中间件应用到所有路由
- **路由组**：`admin` 和 `public` 两个平级路由组
- **组级中间件**：`admin` 组添加额外的 `admin` 中间件
- **完整的 CRUD 操作**：创建、读取、更新、删除用户

## 生成的路由结构

用户服务会生成以下路由：

- `GET /v1/users/:id` (应用 auth, logging 中间件)
- `POST /v1/users` (应用 auth, logging 中间件)
- `PUT /v1/users/:id` (应用 auth, logging 中间件)
- `DELETE /v1/users/:id` (应用 auth, logging 中间件)
- `GET /v1/admin/users` (应用 auth, logging, admin 中间件)
- `DELETE /v1/admin/users` (应用 auth, logging, admin 中间件)
- `GET /v1/public/users/:id` (应用 auth, logging 中间件)
- `GET /v1/public/users/search` (应用 auth, logging 中间件)
