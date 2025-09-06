package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// generateMiddlewareImplementations 生成中间件实现
func (g *CodeGenerator) generateMiddlewareImplementations() error {
	if !g.template.Options.GenerateMiddleware {
		return nil
	}

	// 确定中间件输出目录
	middlewareOutputDir := g.template.Options.MiddlewareOutputDir
	if middlewareOutputDir == "" {
		middlewareOutputDir = "internal/middleware"
	}

	// 计算绝对路径
	apiOutputDir := g.template.Options.OutputDir
	var projectRoot string
	if apiOutputDir == "." {
		// 如果输出目录是当前目录，从当前目录向上查找项目根目录
		projectRoot = g.findProjectRoot(".")
	} else {
		// 从 api/user/v1 回到项目根目录，需要向上两级
		projectRoot = filepath.Dir(filepath.Dir(filepath.Dir(apiOutputDir)))
	}
	absoluteMiddlewareDir := filepath.Join(projectRoot, middlewareOutputDir)

	// 创建目录
	if err := os.MkdirAll(absoluteMiddlewareDir, 0755); err != nil {
		return fmt.Errorf("创建中间件目录失败: %w", err)
	}

	// 收集所有中间件名称
	middlewareNames := make(map[string]bool)
	for _, service := range g.template.Services {
		for _, group := range service.RouteGroups {
			for _, middleware := range group.Middleware {
				middlewareNames[middleware] = true
			}
		}
	}

	// 生成中间件实现
	if len(middlewareNames) > 0 {
		if err := g.generateMiddlewareFile(absoluteMiddlewareDir, middlewareNames); err != nil {
			return fmt.Errorf("生成中间件文件失败: %w", err)
		}
	}

	return nil
}

// generateMiddlewareFile 生成中间件文件
func (g *CodeGenerator) generateMiddlewareFile(outputDir string, middlewareNames map[string]bool) error {
	// 获取服务名称用于生成中间件结构体名称
	serviceName := "User" // 默认值
	if len(g.template.Services) > 0 {
		serviceName = g.template.Services[0].Name
		// 移除 Service 后缀
		if strings.HasSuffix(serviceName, "Service") {
			serviceName = strings.TrimSuffix(serviceName, "Service")
		}
	}

	filename := strings.ToLower(serviceName) + ".go"
	filepath := filepath.Join(outputDir, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("中间件文件已存在，跳过生成: %s\n", filepath)
		return nil
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	// 生成包声明
	file.WriteString("package middleware\n\n")

	// 生成导入
	file.WriteString("import \"github.com/gin-gonic/gin\"\n\n")

	// 生成中间件结构体
	middlewareStructName := serviceName + "Middleware"
	file.WriteString(fmt.Sprintf("type %s struct {\n", middlewareStructName))
	file.WriteString("}\n\n")

	// 生成构造函数
	file.WriteString(fmt.Sprintf("func New%s() *%s {\n", middlewareStructName, middlewareStructName))
	file.WriteString(fmt.Sprintf("\treturn &%s{}\n", middlewareStructName))
	file.WriteString("}\n\n")

	// 生成中间件方法
	for middlewareName := range middlewareNames {
		// 清理中间件名称中的引号
		cleanName := strings.Trim(middlewareName, `"'`)
		titleName := strings.Title(cleanName)
		file.WriteString(fmt.Sprintf("func (m *%s) %s() gin.HandlerFunc {\n", middlewareStructName, titleName))
		file.WriteString("\treturn func(c *gin.Context) {\n")
		file.WriteString(fmt.Sprintf("\t\t// TODO: 实现 %s 中间件逻辑\n", titleName))
		file.WriteString("\t\tc.Next()\n")
		file.WriteString("\t}\n")
		file.WriteString("}\n\n")
	}

	return nil
}
