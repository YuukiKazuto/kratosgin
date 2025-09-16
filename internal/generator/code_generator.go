package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/YuukiKazuto/kratosgin/internal/parser"
)

//go:embed templates/types.tmpl
var typesTemplate string

//go:embed templates/service.tmpl
var serviceTemplate string

//go:embed templates/ginutil.tmpl
var ginutilTemplate string

// CodeGenerator 代码生成器
type CodeGenerator struct {
	template *parser.GinTemplate
}

// NewCodeGenerator 创建新的代码生成器
func NewCodeGenerator(template *parser.GinTemplate) *CodeGenerator {
	return &CodeGenerator{
		template: template,
	}
}

// Generate 生成所有代码文件
func (g *CodeGenerator) Generate() error {
	// 创建输出目录
	outputDir := g.template.Options.OutputDir
	if outputDir == "" {
		// 如果没有指定输出目录，使用当前目录
		outputDir = "."
	}

	// 如果 outputDir 是相对路径，确保它相对于当前工作目录（gin 文件所在目录）
	if !filepath.IsAbs(outputDir) {
		// 相对路径直接使用，因为当前工作目录已经是 gin 文件所在目录
		outputDir = filepath.Clean(outputDir)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// 生成类型定义文件
	if err := g.generateTypes(); err != nil {
		return fmt.Errorf("failed to generate types: %w", err)
	}

	// 生成服务接口文件
	if err := g.generateServiceInterface(); err != nil {
		return fmt.Errorf("failed to generate service interface: %w", err)
	}

	// 生成 HTTP 处理器文件
	if err := g.generateHTTPHandlers(); err != nil {
		return fmt.Errorf("failed to generate HTTP handlers: %w", err)
	}

	// 检查是否有方法需要 gin context，如果有则生成 context 工具文件
	hasGinContext := false
	for _, service := range g.template.Services {
		for _, method := range service.Methods {
			if method.WithGinContext {
				hasGinContext = true
				break
			}
		}
		for _, group := range service.RouteGroups {
			for _, method := range group.Methods {
				if method.WithGinContext {
					hasGinContext = true
					break
				}
			}
		}
		if hasGinContext {
			break
		}
	}

	if hasGinContext {
		if err := g.generateGinContext(); err != nil {
			return fmt.Errorf("failed to generate gin context: %w", err)
		}
	}

	// 生成 service 实现
	if g.template.Options.GenerateService {
		if err := g.generateServiceImplementations(); err != nil {
			return fmt.Errorf("failed to generate service implementations: %w", err)
		}
	}

	// 生成中间件实现
	if g.template.Options.GenerateMiddleware {
		if err := g.generateMiddlewareImplementations(); err != nil {
			return fmt.Errorf("生成中间件实现失败: %w", err)
		}
	}

	return nil
}

// generateTypes 生成类型定义
func (g *CodeGenerator) generateTypes() error {
	t, err := template.New("types.tmpl").Funcs(template.FuncMap{
		"title": strings.Title,
	}).Parse(typesTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(g.template.Options.OutputDir, "types.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, g.template)
}

// generateServiceInterface 生成服务接口
func (g *CodeGenerator) generateServiceInterface() error {
	t, err := template.New("service.tmpl").Funcs(template.FuncMap{
		"title": strings.Title,
	}).Parse(serviceTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(g.template.Options.OutputDir, "service.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, g.template)
}

// generateGinContext 生成 gin context 工具
func (g *CodeGenerator) generateGinContext() error {
	t, err := template.New("ginutil.tmpl").Parse(ginutilTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(g.template.Options.OutputDir, "ginutil.go"))
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, g.template)
}

// inferModuleName 从输出目录推断模块名
func (g *CodeGenerator) inferModuleName() string {
	outputDir := g.template.Options.OutputDir

	// 从输出目录向上查找项目根目录
	projectRoot := g.findProjectRoot(outputDir)
	if projectRoot != "" {
		goModPath := filepath.Join(projectRoot, "go.mod")
		if moduleName := g.readModuleName(goModPath); moduleName != "" {
			return moduleName
		}
	}

	// 如果找不到 go.mod，使用默认值
	return "kratos-project"
}

// findProjectRoot 从输出目录向上查找项目根目录
func (g *CodeGenerator) findProjectRoot(startDir string) string {
	dir := startDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// inferAPIPath 从输出目录推断 API 路径
func (g *CodeGenerator) inferAPIPath() string {
	outputDir := g.template.Options.OutputDir
	// 例如: api/product/v1 -> product
	parts := strings.Split(outputDir, string(filepath.Separator))
	if len(parts) >= 2 && parts[0] == "api" {
		return parts[1]
	}
	return "user" // 默认值
}

// findGoModFile 查找 go.mod 文件
func (g *CodeGenerator) findGoModFile(startDir string) string {
	dir := startDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return goModPath
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// readModuleName 从 go.mod 文件读取模块名
func (g *CodeGenerator) readModuleName(goModPath string) string {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return ""
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "module "))
		}
	}
	return ""
}
