package templates

import (
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed new_template.gin
var newTemplateContent string

// ProcessNewTemplate 处理新建模板
func ProcessNewTemplate(name string) (string, error) {
	// 检测当前目录是否包含版本号
	prefix := detectVersionPrefix()

	// 检测包名和输出目录
	packageName, outputDir := detectPackageAndOutputDir(name, prefix)

	// 将模板内容中的占位符替换为实际值
	content := strings.ReplaceAll(newTemplateContent, "{{.Name}}", strings.ToLower(name))
	content = strings.ReplaceAll(content, "{{.NameTitle}}", strings.Title(name))
	content = strings.ReplaceAll(content, "{{.ServiceName}}", strings.Title(name)+"Service")
	content = strings.ReplaceAll(content, "{{.Prefix}}", prefix)
	content = strings.ReplaceAll(content, "{{.PackageName}}", packageName)
	content = strings.ReplaceAll(content, "{{.OutputDir}}", outputDir)

	return content, nil
}

// ProcessNewTemplateWithPath 处理新建模板（带路径）
func ProcessNewTemplateWithPath(name, outputPath string) (string, error) {
	// 检测当前目录是否包含版本号
	prefix := detectVersionPrefix()

	// 根据输出路径检测包名和输出目录
	packageName, outputDir := detectPackageAndOutputDirFromPath(name, outputPath, prefix)

	// 将模板内容中的占位符替换为实际值
	content := strings.ReplaceAll(newTemplateContent, "{{.Name}}", strings.ToLower(name))
	content = strings.ReplaceAll(content, "{{.NameTitle}}", strings.Title(name))
	content = strings.ReplaceAll(content, "{{.ServiceName}}", strings.Title(name)+"Service")
	content = strings.ReplaceAll(content, "{{.Prefix}}", prefix)
	content = strings.ReplaceAll(content, "{{.PackageName}}", packageName)
	content = strings.ReplaceAll(content, "{{.OutputDir}}", outputDir)

	return content, nil
}

// detectPackageAndOutputDir 检测包名和输出目录
func detectPackageAndOutputDir(name, prefix string) (string, string) {
	dir, err := os.Getwd()
	if err != nil {
		return "v1", "."
	}

	// 检查当前目录结构
	dirName := filepath.Base(dir)
	parentDir := filepath.Base(filepath.Dir(dir))

	// 如果当前目录是版本目录（如 v1, v2）
	if strings.HasPrefix(dirName, "v") && len(dirName) >= 2 && isNumeric(dirName[1:]) {
		// 如果父目录是服务名目录
		if parentDir == strings.ToLower(name) {
			return dirName, "."
		}
		// 如果当前目录就是版本目录，使用它作为包名
		return dirName, "."
	}

	// 如果父目录是版本目录
	if strings.HasPrefix(parentDir, "v") && len(parentDir) >= 2 && isNumeric(parentDir[1:]) {
		return parentDir, "."
	}

	// 默认情况
	return "v1", "."
}

// detectPackageAndOutputDirFromPath 根据输出路径检测包名和输出目录
func detectPackageAndOutputDirFromPath(name, outputPath, prefix string) (string, string) {
	// 如果输出路径是文件路径，提取目录部分
	if strings.HasSuffix(outputPath, ".gin") {
		outputPath = filepath.Dir(outputPath)
	}

	// 标准化路径分隔符
	outputPath = strings.ReplaceAll(outputPath, "\\", "/")

	// 解析路径组件
	parts := strings.Split(strings.Trim(outputPath, "/"), "/")

	// 查找版本目录
	var versionDir string
	for i, part := range parts {
		if strings.HasPrefix(part, "v") && len(part) >= 2 && isNumeric(part[1:]) {
			versionDir = part
			// 检查前面的部分是否是服务名
			if i > 0 && parts[i-1] == strings.ToLower(name) {
				// 使用当前目录作为输出目录
				return versionDir, "."
			}
		}
	}

	// 如果找到版本目录但没有找到服务名，使用默认结构
	if versionDir != "" {
		return versionDir, "."
	}

	// 默认情况
	return "v1", "."
}

// detectVersionPrefix 检测当前目录的版本前缀
func detectVersionPrefix() string {
	dir, err := os.Getwd()
	if err != nil {
		return "v1" // 默认返回 v1
	}

	// 检查当前目录名是否匹配版本模式 (v1, v2, v3, etc.)
	dirName := filepath.Base(dir)
	if strings.HasPrefix(dirName, "v") && len(dirName) >= 2 {
		// 检查是否是有效的版本号
		version := dirName[1:]
		if isNumeric(version) {
			return dirName
		}
	}

	// 检查父目录
	parentDir := filepath.Base(filepath.Dir(dir))
	if strings.HasPrefix(parentDir, "v") && len(parentDir) >= 2 {
		version := parentDir[1:]
		if isNumeric(version) {
			return parentDir
		}
	}

	return "v1" // 默认返回 v1
}

// isNumeric 检查字符串是否为纯数字
func isNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return len(s) > 0
}

// findTemplatePath 查找模板文件路径
func findTemplatePath() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		templatePath := filepath.Join(dir, "internal", "templates", "new_template.gin")
		if _, err := os.Stat(templatePath); err == nil {
			return templatePath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("找不到模板文件")
}
