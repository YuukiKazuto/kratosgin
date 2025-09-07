package formatter

import (
	"fmt"
	"os"
	"strings"
)

// FormatGinFile 格式化 gin 文件
func FormatGinFile(filePath string) error {
	// 读取文件内容
	content, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("读取文件失败: %v", err)
	}

	// 格式化内容
	formattedContent := formatContent(string(content))

	// 写回文件
	if err := os.WriteFile(filePath, []byte(formattedContent), 0644); err != nil {
		return fmt.Errorf("写入文件失败: %v", err)
	}

	return nil
}

// formatContent 格式化文件内容
func formatContent(content string) string {
	lines := strings.Split(content, "\n")
	var formattedLines []string

	inOptions := false
	inType := false
	inTypeGroup := false
	inService := false
	inGroup := false

	for _, line := range lines {
		originalLine := line
		line = strings.TrimSpace(line)

	processLine:

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "//") {
			formattedLines = append(formattedLines, originalLine)
			continue
		}

		// 处理 options 块
		if strings.HasPrefix(line, "options") {
			inOptions = true
			// 格式化 options 行，确保 { 前有空格
			optionsLine := strings.TrimSpace(line)
			if strings.Contains(optionsLine, "{") && !strings.Contains(optionsLine, " {") {
				optionsLine = strings.Replace(optionsLine, "{", " {", 1)
			} else if !strings.HasSuffix(optionsLine, "{") {
				optionsLine += " {"
			}
			formattedLines = append(formattedLines, optionsLine)
			continue
		}

		if inOptions {
			if line == "}" {
				inOptions = false
				formattedLines = append(formattedLines, "}")
				continue
			}
			// options 内容缩进并格式化
			formattedLine := formatOptionsLine(line)
			formattedLines = append(formattedLines, "\t"+formattedLine)
			continue
		}

		// 处理 type 定义
		if strings.HasPrefix(line, "type ") || strings.HasPrefix(line, "type(") {
			// 检查是否是 type ( ) 格式
			if strings.Contains(line, "(") {
				inTypeGroup = true
				formattedLines = append(formattedLines, "")

				// 格式化 type ( 行，确保 ( 前有空格
				typeLine := strings.TrimSpace(line)
				if !strings.Contains(typeLine, " (") {
					typeLine = strings.Replace(typeLine, "(", " (", 1)
				}

				// 分离 type ( 和后面的内容
				parts := strings.Split(typeLine, "(")
				if len(parts) >= 2 && strings.TrimSpace(parts[1]) != "" {
					// type ( 后面有内容，需要换行
					formattedLines = append(formattedLines, parts[0]+"(")
					// 处理后面的内容
					restContent := strings.TrimSpace(parts[1])
					if strings.Contains(restContent, "{") {
						// 如果包含 {，需要分离 type 名和 {
						if !strings.Contains(restContent, " {") {
							restContent = strings.Replace(restContent, "{", " {", 1)
						}
						// 分离 type 名和 {
						typeParts := strings.Split(restContent, " {")
						if len(typeParts) >= 2 {
							formattedLines = append(formattedLines, "\t"+typeParts[0]+" {")
						} else {
							formattedLines = append(formattedLines, "\t"+restContent)
						}
					} else {
						formattedLines = append(formattedLines, "\t"+restContent)
					}
				} else {
					// type ( 后面没有内容
					formattedLines = append(formattedLines, typeLine)
				}
				continue
			} else {
				inType = true
				formattedLines = append(formattedLines, "")
				// 格式化 type 行，确保 { 前有空格
				typeLine := strings.TrimSpace(line)
				if strings.Contains(typeLine, "{") && !strings.Contains(typeLine, " {") {
					typeLine = strings.Replace(typeLine, "{", " {", 1)
				}
				formattedLines = append(formattedLines, typeLine)
				continue
			}
		}

		if inType {
			if line == "}" {
				inType = false
				formattedLines = append(formattedLines, "}")
				continue
			}
			// type 内容缩进
			formattedLines = append(formattedLines, "\t"+line)
			continue
		}

		// 处理 type ( ) 组
		if inTypeGroup {
			if line == ")" {
				inTypeGroup = false
				formattedLines = append(formattedLines, ")")
				continue
			}
			// 检查是否是新的非类型内容（如 service, group 等）
			if strings.HasPrefix(line, "service ") || strings.HasPrefix(line, "group ") || strings.HasPrefix(line, "type ") {
				// 结束 type 组，但不添加 )，因为可能已经有 )
				inTypeGroup = false
				// 继续处理当前行
				goto processLine
			}
			// type 组内容处理
			if strings.Contains(line, "{") {
				// 格式化 type 行，确保 { 前有空格，并换行缩进
				typeLine := strings.TrimSpace(line)
				if !strings.Contains(typeLine, " {") {
					typeLine = strings.Replace(typeLine, "{", " {", 1)
				}
				formattedLines = append(formattedLines, "\t"+typeLine)
				continue
			}
			if line == "}" {
				formattedLines = append(formattedLines, "\t}")
				continue
			}
			// 处理 }) 的情况，将其分解为 } 和 )
			if line == "})" {
				formattedLines = append(formattedLines, "\t}")
				inTypeGroup = false
				formattedLines = append(formattedLines, ")")
				continue
			}
			// 其他内容（字段定义）需要缩进
			if strings.TrimSpace(line) != "" {
				formattedLines = append(formattedLines, "\t\t"+line)
			} else {
				formattedLines = append(formattedLines, originalLine)
			}
			continue
		}

		// 处理 service 定义
		if strings.HasPrefix(line, "service ") {
			inService = true
			formattedLines = append(formattedLines, "")
			// 格式化 service 行，确保 { 前有空格
			serviceLine := strings.TrimSpace(line)
			if strings.Contains(serviceLine, "{") && !strings.Contains(serviceLine, " {") {
				serviceLine = strings.Replace(serviceLine, "{", " {", 1)
			}
			formattedLines = append(formattedLines, serviceLine)
			continue
		}

		if inService {
			// 先检查是否在 group 中
			if inGroup {
				if line == "}" {
					inGroup = false
					formattedLines = append(formattedLines, "\t}")
					continue
				}
				// group 内容缩进
				formattedLines = append(formattedLines, "\t\t"+line)
				continue
			}

			// 处理 group 定义
			if strings.HasPrefix(line, "group ") {
				inGroup = true
				formattedLines = append(formattedLines, "")
				// 格式化 group 行，确保 { 前有空格
				groupLine := strings.TrimSpace(line)
				if strings.Contains(groupLine, "{") && !strings.Contains(groupLine, " {") {
					groupLine = strings.Replace(groupLine, "{", " {", 1)
				}
				formattedLines = append(formattedLines, "\t"+groupLine)
				continue
			}

			if line == "}" {
				inService = false
				formattedLines = append(formattedLines, "}")
				continue
			}

			// service 内容缩进
			formattedLines = append(formattedLines, "\t"+line)
			continue
		}

		// 其他行保持原样
		formattedLines = append(formattedLines, originalLine)
	}

	// 清理多余的空行
	return cleanupEmptyLines(strings.Join(formattedLines, "\n"))
}

// cleanupEmptyLines 清理多余的空行
func cleanupEmptyLines(content string) string {
	lines := strings.Split(content, "\n")
	var cleanedLines []string

	for i, line := range lines {
		// 如果是空行
		if strings.TrimSpace(line) == "" {
			// 如果前一行也是空行，跳过
			if i > 0 && strings.TrimSpace(lines[i-1]) == "" {
				continue
			}
			// 如果下一行是空行，跳过
			if i < len(lines)-1 && strings.TrimSpace(lines[i+1]) == "" {
				continue
			}
		}
		cleanedLines = append(cleanedLines, line)
	}

	// 确保文件以单个换行符结尾
	result := strings.Join(cleanedLines, "\n")
	if !strings.HasSuffix(result, "\n") {
		result += "\n"
	}

	return result
}

// formatOptionsLine 格式化 options 行
func formatOptionsLine(line string) string {
	line = strings.TrimSpace(line)

	// 处理 packageName: value 格式
	if strings.Contains(line, "packageName:") {
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			return "packageName: " + strings.TrimSpace(parts[1])
		}
	}

	// 处理 outputDir: value 格式
	if strings.Contains(line, "outputDir:") {
		parts := strings.Split(line, ":")
		if len(parts) == 2 {
			return "outputDir: " + strings.TrimSpace(parts[1])
		}
	}

	return line
}

// FormatGinFileWithBackup 格式化 gin 文件并创建备份
func FormatGinFileWithBackup(filePath string) error {
	// 创建备份
	backupPath := filePath + ".bak"
	if err := copyFile(filePath, backupPath); err != nil {
		return fmt.Errorf("创建备份失败: %v", err)
	}

	// 格式化文件
	if err := FormatGinFile(filePath); err != nil {
		// 如果格式化失败，恢复备份
		copyFile(backupPath, filePath)
		os.Remove(backupPath)
		return err
	}

	// 删除备份
	os.Remove(backupPath)
	return nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	content, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, content, 0644)
}
