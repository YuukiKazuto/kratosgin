package cli

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/YuukiKazuto/kratosgin/internal/generator"
	"github.com/YuukiKazuto/kratosgin/internal/parser"
	"github.com/YuukiKazuto/kratosgin/internal/templates"
	"github.com/spf13/cobra"
)

// GenCommand 生成命令
func GenCommand() *cobra.Command {
	var (
		templateFile        string
		serviceOutputDir    string
		middlewareOutputDir string
	)

	cmd := &cobra.Command{
		Use:   "gen",
		Short: "生成 API 代码",
		Long:  "根据 .gin 模板文件生成 Kratos API 代码",
		Run: func(cmd *cobra.Command, args []string) {
			runGen(templateFile, serviceOutputDir, middlewareOutputDir)
		},
	}

	cmd.Flags().StringVarP(&templateFile, "file", "f", "", "模板文件路径 (.gin 文件)")
	cmd.Flags().StringVarP(&serviceOutputDir, "service", "s", "", "Service 实现输出目录")
	cmd.Flags().StringVarP(&middlewareOutputDir, "middleware", "m", "", "Middleware 实现输出目录")
	cmd.MarkFlagRequired("file")

	return cmd
}

// NewCommand 新建命令
func NewCommand() *cobra.Command {
	var outputPath string

	cmd := &cobra.Command{
		Use:   "new [name]",
		Short: "创建新的 .gin 模板文件",
		Long:  "创建一个新的 .gin 模板文件，包含基本的服务定义和示例接口",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			runNew(args[0], outputPath)
		},
	}

	cmd.Flags().StringVarP(&outputPath, "output", "o", "", "指定输出路径（可选）")

	return cmd
}

// runGen 执行生成命令
func runGen(templateFile, serviceOutputDir, middlewareOutputDir string) {
	// 检查文件是否存在
	if _, err := os.Stat(templateFile); os.IsNotExist(err) {
		log.Fatalf("模板文件不存在: %s", templateFile)
	}

	// 读取文件内容
	content, err := os.ReadFile(templateFile)
	if err != nil {
		log.Fatalf("读取模板文件失败: %v", err)
	}

	// 解析模板
	template, err := parser.ParseGinTemplate(string(content))
	if err != nil {
		log.Fatalf("解析模板失败: %v", err)
	}

	// 设置生成选项
	// 如果用户指定了 service 输出目录，则生成 service 实现
	if serviceOutputDir != "" {
		template.Options.GenerateService = true
		template.Options.ServiceOutputDir = serviceOutputDir
	}

	// 设置中间件生成选项
	// 如果用户指定了 middleware 输出目录，则生成 middleware 实现
	if middlewareOutputDir != "" {
		template.Options.GenerateMiddleware = true
		template.Options.MiddlewareOutputDir = middlewareOutputDir
	}

	// 生成代码
	gen := generator.NewCodeGenerator(template)
	if err := gen.Generate(); err != nil {
		log.Fatalf("生成代码失败: %v", err)
	}

	fmt.Printf("代码生成成功! 输出目录: %s, 包名: %s\n", template.Options.OutputDir, template.Options.PackageName)
}

// runNew 执行新建命令
func runNew(name, outputPath string) {
	var templateContent string
	var err error

	if outputPath != "" {
		// 如果指定了输出路径，使用带路径的模板处理
		templateContent, err = templates.ProcessNewTemplateWithPath(name, outputPath)
	} else {
		// 默认使用当前目录的模板处理
		templateContent, err = templates.ProcessNewTemplate(name)
	}

	if err != nil {
		log.Fatalf("生成模板内容失败: %v", err)
	}

	// 确定输出路径
	var filename string
	if outputPath != "" {
		// 如果指定了输出路径，使用指定的路径
		if strings.HasSuffix(outputPath, ".gin") {
			filename = outputPath
		} else {
			// 如果路径是目录，在目录下创建文件
			if err := os.MkdirAll(outputPath, 0755); err != nil {
				log.Fatalf("创建目录失败: %v", err)
			}
			filename = filepath.Join(outputPath, name+".gin")
		}
	} else {
		// 默认在当前目录创建
		filename = name + ".gin"
	}

	// 创建模板文件
	if err := os.WriteFile(filename, []byte(templateContent), 0644); err != nil {
		log.Fatalf("创建模板文件失败: %v", err)
	}

	fmt.Printf("模板文件创建成功: %s\n", filename)
}
