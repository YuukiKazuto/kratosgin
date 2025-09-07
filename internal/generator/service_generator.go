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

//go:embed templates/service_impl.tmpl
var serviceImplTemplate string

// generateServiceImplementations 生成 service 实现
func (g *CodeGenerator) generateServiceImplementations() error {
	// 确定 service 输出目录
	serviceOutputDir := g.template.Options.ServiceOutputDir
	if serviceOutputDir == "" {
		// 默认输出到 internal/service 目录
		serviceOutputDir = "internal/service"
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
	absoluteServiceDir := filepath.Join(projectRoot, serviceOutputDir)

	// 创建输出目录
	if err := os.MkdirAll(absoluteServiceDir, 0755); err != nil {
		return fmt.Errorf("failed to create service output directory: %w", err)
	}

	// 为每个服务生成实现
	for _, service := range g.template.Services {
		if err := g.generateSingleServiceImplementation(service, absoluteServiceDir); err != nil {
			return fmt.Errorf("failed to generate service implementation for %s: %w", service.Name, err)
		}
	}

	return nil
}

// generateSingleServiceImplementation 生成单个服务的实现
func (g *CodeGenerator) generateSingleServiceImplementation(service parser.Service, outputDir string) error {
	// 准备模板数据
	moduleName := g.inferModuleName()
	apiPath, packageName := g.inferAPIPathAndPackage()
	packageAlias := g.generatePackageAlias(apiPath, packageName)

	// 收集所有方法（包括直接方法和路由分组中的方法）
	allMethods := make([]parser.Method, 0)
	allMethods = append(allMethods, service.Methods...)
	for _, group := range service.RouteGroups {
		allMethods = append(allMethods, group.Methods...)
	}

	templateData := struct {
		ServiceName  string
		ModuleName   string
		APIPath      string
		PackageName  string
		PackageAlias string
		Methods      []parser.Method
	}{
		ServiceName:  service.Name,
		ModuleName:   moduleName,
		APIPath:      apiPath,
		PackageName:  packageName,
		PackageAlias: packageAlias,
		Methods:      allMethods,
	}

	// 使用模板生成文件
	t, err := template.New("service_impl.tmpl").Funcs(template.FuncMap{
		"title": strings.Title,
	}).Parse(serviceImplTemplate)
	if err != nil {
		return err
	}

	// 写入文件
	// 去掉 Service 后缀
	serviceName := service.Name
	serviceName = strings.TrimSuffix(serviceName, "Service")
	filename := fmt.Sprintf("%s.go", strings.ToLower(serviceName))
	filepath := filepath.Join(outputDir, filename)

	// 检查文件是否已存在
	if _, err := os.Stat(filepath); err == nil {
		fmt.Printf("Service 文件已存在，跳过生成: %s\n", filepath)
		return nil
	}

	file, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer file.Close()

	return t.Execute(file, templateData)
}

// inferAPIPathAndPackage 推断 API 路径和包名
func (g *CodeGenerator) inferAPIPathAndPackage() (string, string) {
	outputDir := g.template.Options.OutputDir
	packageName := g.template.Options.PackageName

	// 如果 outputDir 是 "."，需要从当前工作目录推断
	if outputDir == "." {
		// 从当前工作目录推断，例如 test/api/product/v2 -> product, v2
		cwd, err := os.Getwd()
		if err == nil {
			parts := strings.Split(cwd, string(filepath.Separator))
			// 查找 api 目录
			for i, part := range parts {
				if part == "api" && i+2 < len(parts) {
					return parts[i+1], parts[i+2] // 返回 apiPath 和 packageName
				}
			}
		}
		return "user", packageName // 使用模板中的包名
	}

	// 例如: api/product/v1 -> product, v1
	parts := strings.Split(outputDir, string(filepath.Separator))
	if len(parts) >= 3 && parts[0] == "api" {
		return parts[1], parts[2] // 返回 apiPath 和 packageName
	}
	return "user", packageName // 使用模板中的包名
}

// generatePackageAlias 生成包别名
func (g *CodeGenerator) generatePackageAlias(apiPath, packageName string) string {
	// 例如: user + v1 -> userV1
	return apiPath + strings.Title(packageName)
}
