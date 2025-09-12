package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"

	"github.com/YuukiKazuto/kratosgin/internal/parser"
)

//go:embed templates/handlers.tmpl
var handlersTemplate string

// toCamelCase 将字符串转换为小驼峰命名
func toCamelCase(s string) string {
	if s == "" {
		return s
	}

	runes := []rune(s)
	if len(runes) == 0 {
		return s
	}

	// 第一个字符转小写
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

// generateHTTPHandlers 生成 HTTP 处理器
func (g *CodeGenerator) generateHTTPHandlers() error {
	t, err := template.New("handlers.tmpl").Funcs(template.FuncMap{
		"generateServiceHandlerWithGroups": g.generateServiceHandlerWithGroups,
		"generateRouteGroupHandler":        g.generateRouteGroupHandler,
		"generateStandaloneRoutesHandler":  g.generateStandaloneRoutesHandler,
	}).Parse(handlersTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(g.template.Options.OutputDir, "handlers.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, g.template)
}

// generateServiceHandlerWithGroups 生成带路由组的服务处理器
func (g *CodeGenerator) generateServiceHandlerWithGroups(service parser.Service) string {
	var result strings.Builder

	// 收集所有中间件名称（包括服务级别和路由组级别）
	middlewareSet := make(map[string]bool)

	// 收集服务级别中间件
	for _, middleware := range service.Middleware {
		cleanMiddleware := strings.Trim(middleware, `"'`)
		if cleanMiddleware != "" {
			middlewareSet[cleanMiddleware] = true
		}
	}

	// 收集路由组级别中间件
	for _, group := range service.RouteGroups {
		for _, middleware := range group.Middleware {
			cleanMiddleware := strings.Trim(middleware, `"'`)
			if cleanMiddleware != "" {
				middlewareSet[cleanMiddleware] = true
			}
		}
		for _, method := range group.Methods {
			for _, middleware := range method.Middleware {
				cleanMiddleware := strings.Trim(middleware, `"'`)
				if cleanMiddleware != "" {
					middlewareSet[cleanMiddleware] = true
				}
			}
		}
	}

	// 生成中间件接口
	if len(middlewareSet) > 0 {
		result.WriteString("// Middleware 中间件接口\n")
		result.WriteString("type Middleware interface {\n")
		for middleware := range middlewareSet {
			cleanName := strings.Trim(middleware, `"'`)
			result.WriteString(fmt.Sprintf("\t%s() gin.HandlerFunc\n", strings.Title(cleanName)))
		}
		result.WriteString("}\n\n")
	}

	// 生成处理器结构体
	result.WriteString(fmt.Sprintf("// %sHandler %s 处理器\n", service.Name, service.Name))
	result.WriteString(fmt.Sprintf("type %sHandler struct {\n", service.Name))
	result.WriteString("\tlog *log.Helper\n")
	if len(middlewareSet) > 0 {
		result.WriteString("\tmiddleware Middleware\n")
	}
	result.WriteString(fmt.Sprintf("\t%s %s\n", toCamelCase(service.Name), service.Name))
	result.WriteString("\ttranslator ut.Translator\n")
	result.WriteString("}\n\n")

	// 生成构造函数
	result.WriteString(fmt.Sprintf("// New%sHandler 创建 %s 处理器\n", service.Name, service.Name))
	result.WriteString(fmt.Sprintf("func New%sHandler(logger log.Logger", service.Name))
	if len(middlewareSet) > 0 {
		result.WriteString(", middleware Middleware")
	}
	result.WriteString(fmt.Sprintf(", %s %s", toCamelCase(service.Name), service.Name))
	result.WriteString(", translator ut.Translator")
	result.WriteString(fmt.Sprintf(") *%sHandler {\n", service.Name))
	result.WriteString("\treturn &")
	result.WriteString(fmt.Sprintf("%sHandler{\n", service.Name))
	result.WriteString("\t\tlog: log.NewHelper(logger),\n")
	if len(middlewareSet) > 0 {
		result.WriteString("\t\tmiddleware: middleware,\n")
	}
	result.WriteString(fmt.Sprintf("\t\t%s: %s,\n", toCamelCase(service.Name), toCamelCase(service.Name)))
	result.WriteString("\t\ttranslator: translator,\n")
	result.WriteString("\t}\n")
	result.WriteString("}\n\n")

	// 生成路由注册方法
	result.WriteString("// RegisterRoutes 注册路由\n")
	result.WriteString(fmt.Sprintf("func (h *%sHandler) RegisterRoutes(r *gin.Engine) {\n", service.Name))
	// 创建顶级路由组（如果有 prefix）
	var topLevelGroup string
	if service.Prefix != "" {
		topLevelGroup = "PrefixGroup"
		result.WriteString(fmt.Sprintf("\t%s := r.Group(\"/%s\")\n", topLevelGroup, service.Prefix))
		result.WriteString("\t{\n")

		// 应用服务级别中间件
		for _, middleware := range service.Middleware {
			cleanMiddleware := strings.Trim(middleware, `"'`)
			if cleanMiddleware != "" {
				result.WriteString(fmt.Sprintf("\t\t%s.Use(h.middleware.%s())\n",
					topLevelGroup, strings.Title(cleanMiddleware)))
			}
		}
	} else {
		// 如果没有 prefix，直接在根路由上应用服务级别中间件
		for _, middleware := range service.Middleware {
			cleanMiddleware := strings.Trim(middleware, `"'`)
			if cleanMiddleware != "" {
				result.WriteString(fmt.Sprintf("\tr.Use(h.middleware.%s())\n", strings.Title(cleanMiddleware)))
			}
		}
	}

	// 注册非分组路由
	for _, method := range service.Methods {
		router := "r"
		indent := "\t"
		if topLevelGroup != "" {
			router = topLevelGroup
			indent = "\t\t"
		}

		result.WriteString(fmt.Sprintf("%s%s.%s(\"%s\", h.%s)\n",
			indent, router, strings.ToUpper(method.HTTPMethod), method.Path, method.Name))
	}

	// 注册分组路由
	for _, group := range service.RouteGroups {
		groupVarName := strings.Title(group.Name) + "Group"

		parentRouter := "r"
		indent := "\t"
		if topLevelGroup != "" {
			parentRouter = topLevelGroup
			indent = "\t\t"
		}

		result.WriteString(fmt.Sprintf("\n%s%s := %s.Group(\"%s\")\n", indent, groupVarName, parentRouter, group.Path))
		result.WriteString(fmt.Sprintf("%s{\n", indent))

		// 应用组级中间件（避免重复应用服务级别中间件）
		serviceMiddlewareSet := make(map[string]bool)
		for _, middleware := range service.Middleware {
			cleanMiddleware := strings.Trim(middleware, `"'`)
			if cleanMiddleware != "" {
				serviceMiddlewareSet[cleanMiddleware] = true
			}
		}

		for _, middleware := range group.Middleware {
			cleanMiddleware := strings.Trim(middleware, `"'`)
			// 过滤掉注释和无效的中间件名称，如果组中间件不在服务级别中间件中，才应用
			if cleanMiddleware != "" && !serviceMiddlewareSet[cleanMiddleware] {
				result.WriteString(fmt.Sprintf("%s\t%s.Use(h.middleware.%s())\n",
					indent, groupVarName, strings.Title(cleanMiddleware)))
			}
		}

		// 注册组内路由
		for _, method := range group.Methods {
			middlewareChain := ""
			for _, middleware := range method.Middleware {
				cleanMiddleware := strings.Trim(middleware, `"'`)
				middlewareChain += fmt.Sprintf("h.middleware.%s(), ", strings.Title(cleanMiddleware))
			}
			result.WriteString(fmt.Sprintf("%s\t%s.%s(\"%s\", %sh.%s)\n",
				indent, groupVarName, strings.ToUpper(method.HTTPMethod), method.Path, middlewareChain, method.Name))
		}

		result.WriteString(fmt.Sprintf("%s}\n", indent))
	}

	// 关闭顶级路由组
	if topLevelGroup != "" {
		result.WriteString("\t}\n")
	}

	result.WriteString("}\n\n")

	// 生成处理器方法
	allMethods := append(service.Methods, []parser.Method{}...)
	for _, group := range service.RouteGroups {
		allMethods = append(allMethods, group.Methods...)
	}

	for _, method := range allMethods {
		result.WriteString(fmt.Sprintf("// %s %s\n", method.Name, method.Description))
		result.WriteString(fmt.Sprintf("func (h *%sHandler) %s(c *gin.Context) {\n", service.Name, method.Name))

		// 绑定请求
		result.WriteString(fmt.Sprintf("\treq := &%s{}\n", method.Request))
		result.WriteString("\tif err := c.ShouldBind(req); err != nil {\n")
		result.WriteString("\t\terr = translateValidationError(err, h.translator)\n")
		result.WriteString(fmt.Sprintf("\t\th.log.Errorw(\"Struct\", \"%sHandler\", \"method\", \"%s\", \"error\", err)\n", service.Name, method.Name))
		result.WriteString("\t\tc.JSON(http.StatusBadRequest, gin.H{\n")
		result.WriteString("\t\t\t\"message\": err.Error(),\n")
		result.WriteString("\t\t})\n")
		result.WriteString("\t\treturn\n")
		result.WriteString("\t}\n\n")

		// 调用服务
		if method.WithGinContext {
			result.WriteString("\tctx := SaveToContext(c.Request.Context(), c)\n")
		} else {
			result.WriteString("\tctx := c.Request.Context()\n")
		}
		result.WriteString(fmt.Sprintf("\tresp, err := h.%s.%s(ctx, req)\n", toCamelCase(service.Name), strings.Title(method.Name)))
		result.WriteString("\tif err != nil {\n")
		result.WriteString("\t\tkgin.Error(c, err)\n")
		result.WriteString("\t\treturn\n")
		result.WriteString("\t}\n\n")

		// 返回响应
		result.WriteString("\tc.JSON(http.StatusOK, resp)\n")
		result.WriteString("}\n\n")
	}

	return result.String()
}

// generateRouteGroupHandler 生成路由组处理器
func (g *CodeGenerator) generateRouteGroupHandler(group parser.RouteGroup) string {
	var result strings.Builder

	// 收集所有中间件名称
	middlewareSet := make(map[string]bool)
	for _, middleware := range group.Middleware {
		cleanMiddleware := strings.Trim(middleware, `"'`)
		middlewareSet[cleanMiddleware] = true
	}
	for _, method := range group.Methods {
		for _, middleware := range method.Middleware {
			cleanMiddleware := strings.Trim(middleware, `"'`)
			middlewareSet[cleanMiddleware] = true
		}
	}

	// 生成中间件接口
	if len(middlewareSet) > 0 {
		result.WriteString("// Middleware 中间件接口\n")
		result.WriteString("type Middleware interface {\n")
		for middleware := range middlewareSet {
			cleanName := strings.Trim(middleware, `"'`)
			result.WriteString(fmt.Sprintf("\t%s() gin.HandlerFunc\n", strings.Title(cleanName)))
		}
		result.WriteString("}\n\n")
	}

	// 生成处理器结构体
	result.WriteString(fmt.Sprintf("// %sHandler %s 处理器\n", group.Name, group.Name))
	result.WriteString(fmt.Sprintf("type %sHandler struct {\n", group.Name))
	result.WriteString("\tlog *log.Helper\n")
	if len(middlewareSet) > 0 {
		result.WriteString("\tmiddleware Middleware\n")
	}
	result.WriteString("\ttranslator ut.Translator\n")
	result.WriteString("}\n\n")

	// 生成构造函数
	result.WriteString(fmt.Sprintf("// New%sHandler 创建 %s 处理器\n", group.Name, group.Name))
	result.WriteString(fmt.Sprintf("func New%sHandler(logger log.Logger", group.Name))
	if len(middlewareSet) > 0 {
		result.WriteString(", middleware Middleware")
	}
	result.WriteString(", translator ut.Translator")
	result.WriteString(fmt.Sprintf(") *%sHandler {\n", group.Name))
	result.WriteString("\treturn &")
	result.WriteString(fmt.Sprintf("%sHandler{\n", group.Name))
	result.WriteString("\t\tlog: log.NewHelper(logger),\n")
	if len(middlewareSet) > 0 {
		result.WriteString("\t\tmiddleware: middleware,\n")
	}
	result.WriteString("\t\ttranslator: translator,\n")
	result.WriteString("\t}\n")
	result.WriteString("}\n\n")

	// 生成路由注册方法
	result.WriteString("// RegisterRoutes 注册路由\n")
	result.WriteString(fmt.Sprintf("func (h *%sHandler) RegisterRoutes(r *gin.Engine) {\n", group.Name))

	groupVarName := strings.Title(group.Name) + "Group"
	result.WriteString(fmt.Sprintf("\t%s := r.Group(\"%s\")\n", groupVarName, group.Path))
	result.WriteString("\t{\n")

	// 应用组级中间件
	for _, middleware := range group.Middleware {
		cleanMiddleware := strings.Trim(middleware, `"'`)
		result.WriteString(fmt.Sprintf("\t\t%s.Use(h.middleware.%s())\n",
			groupVarName, strings.Title(cleanMiddleware)))
	}

	// 注册组内路由
	for _, method := range group.Methods {
		middlewareChain := ""
		for _, middleware := range method.Middleware {
			cleanMiddleware := strings.Trim(middleware, `"'`)
			middlewareChain += fmt.Sprintf("h.middleware.%s(), ", strings.Title(cleanMiddleware))
		}
		result.WriteString(fmt.Sprintf("\t\t%s.%s(\"%s\", %sh.%s)\n",
			groupVarName, strings.ToUpper(method.HTTPMethod), method.Path, middlewareChain, method.Name))
	}

	result.WriteString("\t}\n")
	result.WriteString("}\n\n")

	// 生成处理器方法
	for _, method := range group.Methods {
		result.WriteString(fmt.Sprintf("// %s %s\n", method.Name, method.Description))
		result.WriteString(fmt.Sprintf("func (h *%sHandler) %s(c *gin.Context) {\n", group.Name, method.Name))

		// 绑定请求
		result.WriteString(fmt.Sprintf("\treq := &%s{}\n", method.Request))
		result.WriteString("\tif err := c.ShouldBind(req); err != nil {\n")
		result.WriteString("\t\terr = translateValidationError(err, h.translator)\n")
		result.WriteString(fmt.Sprintf("\t\th.log.Errorw(\"Struct\", \"%sHandler\", \"method\", \"%s\", \"error\", err)\n", group.Name, method.Name))
		result.WriteString("\t\tc.JSON(http.StatusBadRequest, gin.H{\n")
		result.WriteString("\t\t\t\"message\": err.Error(),\n")
		result.WriteString("\t\t})\n")
		result.WriteString("\t\treturn\n")
		result.WriteString("\t}\n\n")

		// 调用服务
		if method.WithGinContext {
			result.WriteString("\tctx := SaveToContext(c.Request.Context(), c)\n")
		} else {
			result.WriteString("\tctx := c.Request.Context()\n")
		}
		result.WriteString(fmt.Sprintf("\tresp, err := h.%s.%s(ctx, req)\n", toCamelCase(group.Name), method.Name))
		result.WriteString("\tif err != nil {\n")
		result.WriteString("\t\tkgin.Error(c, err)\n")
		result.WriteString("\t\treturn\n")
		result.WriteString("\t}\n\n")

		// 返回响应
		result.WriteString("\tc.JSON(http.StatusOK, resp)\n")
		result.WriteString("}\n\n")
	}

	return result.String()
}

// generateStandaloneRoutesHandler 生成独立路由处理器
func (g *CodeGenerator) generateStandaloneRoutesHandler(routes []parser.StandaloneRoute) string {
	var result strings.Builder

	// 生成处理器结构体
	result.WriteString("// StandaloneHandler 独立路由处理器\n")
	result.WriteString("type StandaloneHandler struct {\n")
	result.WriteString("\tlog *log.Helper\n")
	result.WriteString("\ttranslator ut.Translator\n")
	result.WriteString("}\n\n")

	// 生成构造函数
	result.WriteString("// NewStandaloneHandler 创建独立路由处理器\n")
	result.WriteString("func NewStandaloneHandler(logger log.Logger, translator ut.Translator) *StandaloneHandler {\n")
	result.WriteString("\treturn &StandaloneHandler{\n")
	result.WriteString("\t\tlog: log.NewHelper(logger),\n")
	result.WriteString("\t\ttranslator: translator,\n")
	result.WriteString("\t}\n")
	result.WriteString("}\n\n")

	// 生成路由注册方法
	result.WriteString("// RegisterRoutes 注册路由\n")
	result.WriteString("func (h *StandaloneHandler) RegisterRoutes(r *gin.Engine) {\n")

	for _, route := range routes {
		result.WriteString(fmt.Sprintf("\tr.%s(\"%s\", h.%s)\n",
			strings.ToUpper(route.HTTPMethod), route.Path, route.Name))
	}

	result.WriteString("}\n\n")

	// 生成处理器方法
	for _, route := range routes {
		result.WriteString(fmt.Sprintf("// %s %s\n", route.Name, route.Description))
		result.WriteString(fmt.Sprintf("func (h *StandaloneHandler) %s(c *gin.Context) {\n", route.Name))

		// 绑定请求
		result.WriteString(fmt.Sprintf("\treq := &%s{}\n", route.Request))
		result.WriteString("\tif err := c.ShouldBind(req); err != nil {\n")
		result.WriteString("\t\terr = translateValidationError(err, h.translator)\n")
		result.WriteString(fmt.Sprintf("\t\th.log.Errorw(\"Struct\", \"StandaloneHandler\", \"method\", \"%s\", \"error\", err)\n", route.Name))
		result.WriteString("\t\tc.JSON(http.StatusBadRequest, gin.H{\n")
		result.WriteString("\t\t\t\"message\": err.Error(),\n")
		result.WriteString("\t\t})\n")
		result.WriteString("\t\treturn\n")
		result.WriteString("\t}\n\n")

		// 调用服务
		if route.WithGinContext {
			result.WriteString("\tctx := SaveToContext(c.Request.Context(), c)\n")
		} else {
			result.WriteString("\tctx := c.Request.Context()\n")
		}
		result.WriteString(fmt.Sprintf("\tresp, err := h.%s.%s(ctx, req)\n", toCamelCase(route.Name), route.Name))
		result.WriteString("\tif err != nil {\n")
		result.WriteString("\t\tkgin.Error(c, err)\n")
		result.WriteString("\t\treturn\n")
		result.WriteString("\t}\n\n")

		// 返回响应
		result.WriteString("\tc.JSON(http.StatusOK, resp)\n")
		result.WriteString("}\n\n")
	}

	return result.String()
}
