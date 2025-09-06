package generator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/YuukiKazuto/kratosgin/internal/parser"
)

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
	var content strings.Builder

	// 包声明
	content.WriteString("package service\n\n")

	// 导入
	content.WriteString("import (\n")
	content.WriteString("\t\"context\"\n")
	content.WriteString("\t\n")
	content.WriteString("\t\"github.com/go-kratos/kratos/v2/log\"\n")

	// 导入 API 包
	moduleName := g.inferModuleName()
	apiPath, packageName := g.inferAPIPathAndPackage()
	packageAlias := g.generatePackageAlias(apiPath, packageName)
	content.WriteString(fmt.Sprintf("\t%s \"%s/api/%s/%s\"\n", packageAlias, moduleName, apiPath, packageName))
	content.WriteString(")\n\n")

	// 生成服务实现结构体
	content.WriteString(fmt.Sprintf("// %s %s 服务实现\n", service.Name, service.Name))
	content.WriteString(fmt.Sprintf("type %s struct {\n", service.Name))
	content.WriteString("\tlog *log.Helper\n")
	content.WriteString("}\n\n")

	// 构造函数
	content.WriteString(fmt.Sprintf("// New%s 创建 %s 服务\n", service.Name, service.Name))
	content.WriteString(fmt.Sprintf("func New%s(logger log.Logger) *%s {\n", service.Name, service.Name))
	content.WriteString(fmt.Sprintf("\treturn &%s{\n", service.Name))
	content.WriteString("\t\tlog: log.NewHelper(logger),\n")
	content.WriteString("\t}\n")
	content.WriteString("}\n\n")

	// 生成方法实现
	for _, method := range service.Methods {
		content.WriteString(fmt.Sprintf("// %s %s\n", strings.Title(method.Name), method.Description))
		content.WriteString(fmt.Sprintf("func (s *%s) %s(ctx context.Context, req *%s.%s) (*%s.%s, error) {\n",
			service.Name, strings.Title(method.Name), packageAlias, method.Request, packageAlias, method.Response))
		content.WriteString(fmt.Sprintf("\ts.log.Infof(\"调用 %s 方法\")\n", strings.Title(method.Name)))
		content.WriteString("\t\n")
		content.WriteString("\t// TODO: 实现具体的业务逻辑\n")
		content.WriteString(fmt.Sprintf("\tresp := &%s.%s{}\n", packageAlias, method.Response))
		content.WriteString("\t\n")
		content.WriteString("\treturn resp, nil\n")
		content.WriteString("}\n\n")
	}

	// 写入文件
	filename := fmt.Sprintf("%s.go", strings.ToLower(service.Name))
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

	_, err = file.WriteString(content.String())
	return err
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
