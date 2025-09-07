package generator

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:embed templates/middleware.tmpl
var middlewareTemplate string

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
		// 收集服务级中间件
		for _, middleware := range service.Middleware {
			cleanMiddleware := strings.Trim(middleware, `"'`)
			if cleanMiddleware != "" {
				middlewareNames[cleanMiddleware] = true
			}
		}
		// 收集路由组中间件
		for _, group := range service.RouteGroups {
			for _, middleware := range group.Middleware {
				cleanMiddleware := strings.Trim(middleware, `"'`)
				if cleanMiddleware != "" {
					middlewareNames[cleanMiddleware] = true
				}
			}
		}
	}

	// 如果没有中间件定义，删除已存在的中间件文件
	if len(middlewareNames) == 0 {
		if err := g.deleteExistingMiddlewareFiles(absoluteMiddlewareDir); err != nil {
			return fmt.Errorf("删除中间件文件失败: %w", err)
		}
		return nil
	}

	// 生成中间件实现
	if err := g.generateMiddlewareFile(absoluteMiddlewareDir, middlewareNames); err != nil {
		return fmt.Errorf("生成中间件文件失败: %w", err)
	}

	return nil
}

// deleteExistingMiddlewareFiles 删除已存在的中间件文件
func (g *CodeGenerator) deleteExistingMiddlewareFiles(middlewareDir string) error {
	// 获取服务名称用于确定要删除的文件名
	serviceName := "User" // 默认值
	if len(g.template.Services) > 0 {
		serviceName = g.template.Services[0].Name
		// 移除 Service 后缀
		serviceName = strings.TrimSuffix(serviceName, "Service")
	}

	filename := strings.ToLower(serviceName) + ".go"
	filepath := filepath.Join(middlewareDir, filename)

	// 检查文件是否存在，如果存在则删除
	if _, err := os.Stat(filepath); err == nil {
		if err := os.Remove(filepath); err != nil {
			return fmt.Errorf("删除中间件文件失败: %w", err)
		}
		fmt.Printf("已删除中间件文件: %s\n", filepath)
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
		serviceName = strings.TrimSuffix(serviceName, "Service")

	}

	filename := strings.ToLower(serviceName) + ".go"
	filepath := filepath.Join(outputDir, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("中间件文件已存在，跳过生成: %s\n", filepath)
		return nil
	}

	// 准备模板数据
	moduleName := g.inferModuleName()
	apiPath, packageName := g.inferAPIPathAndPackage()
	packageAlias := g.generatePackageAlias(apiPath, packageName)

	var middlewareNamesList []string
	for middlewareName := range middlewareNames {
		cleanName := strings.Trim(middlewareName, `"'`)
		if cleanName != "" {
			middlewareNamesList = append(middlewareNamesList, cleanName)
		}
	}

	templateData := struct {
		ServiceName     string
		ModuleName      string
		APIPath         string
		PackageName     string
		PackageAlias    string
		MiddlewareNames []string
	}{
		ServiceName:     serviceName,
		ModuleName:      moduleName,
		APIPath:         apiPath,
		PackageName:     packageName,
		PackageAlias:    packageAlias,
		MiddlewareNames: middlewareNamesList,
	}

	// 使用模板生成文件
	t, err := template.New("middleware.tmpl").Funcs(template.FuncMap{
		"title": strings.Title,
	}).Parse(middlewareTemplate)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, templateData)
}
