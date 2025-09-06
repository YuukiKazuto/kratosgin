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
	// 将模板内容中的占位符替换为实际值
	content := strings.ReplaceAll(newTemplateContent, "{{.Name}}", strings.ToLower(name))
	content = strings.ReplaceAll(content, "{{.NameTitle}}", strings.Title(name))
	content = strings.ReplaceAll(content, "{{.ServiceName}}", strings.Title(name)+"Service")

	return content, nil
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
