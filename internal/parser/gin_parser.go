package parser

import (
	"fmt"
	"regexp"
	"strings"
)

// GinTemplate 表示解析后的 gin 模板
type GinTemplate struct {
	Info             Info
	Types            []Type
	Services         []Service
	RouteGroups      []RouteGroup
	StandaloneRoutes []StandaloneRoute
	Options          Options
}

// Info 包含 API 基本信息
type Info struct {
	Title   string
	Version string
	Desc    string
}

// Type 表示数据类型定义
type Type struct {
	Name   string
	Fields []Field
}

// Field 表示字段定义
type Field struct {
	Name     string
	Type     string
	Tag      string
	Comment  string
	Required bool
}

// Service 表示服务定义
type Service struct {
	Name        string
	Prefix      string   // 服务前缀，如 v1, v2 等
	Middleware  []string // 服务级别中间件
	Methods     []Method
	RouteGroups []RouteGroup
}

// Method 表示方法定义
type Method struct {
	Name           string
	HTTPMethod     string
	Path           string
	Request        string
	Response       string
	Description    string
	WithGinContext bool     // 是否在 context 中传递 gin.Context
	Middleware     []string // 中间件列表
}

// RouteGroup 表示路由分组
type RouteGroup struct {
	Name       string
	Path       string
	Middleware []string
	Methods    []Method
}

// StandaloneRoute 表示独立路由
type StandaloneRoute struct {
	Method
}

// Options 表示模板选项
type Options struct {
	WithGinContext      bool
	OutputDir           string
	PackageName         string
	ServiceOutputDir    string // Service 实现输出目录
	GenerateService     bool   // 是否生成 service 实现
	MiddlewareOutputDir string // Middleware 实现输出目录
	GenerateMiddleware  bool   // 是否生成 middleware 实现
}

// ParseGinTemplate 解析 gin 模板文件
func ParseGinTemplate(content string) (*GinTemplate, error) {
	lines := strings.Split(content, "\n")
	template := &GinTemplate{
		Types:            make([]Type, 0),
		Services:         make([]Service, 0),
		RouteGroups:      make([]RouteGroup, 0),
		StandaloneRoutes: make([]StandaloneRoute, 0),
		Options:          Options{},
	}

	var currentService *Service
	var currentType *Type
	var currentRouteGroup *RouteGroup
	var inType bool
	var inTypeGroup bool

	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// 解析 info 块
		if strings.HasPrefix(line, "info") {
			if err := parseInfo(lines, i, &template.Info); err != nil {
				return nil, err
			}
			continue
		}

		// 解析 options 块
		if strings.HasPrefix(line, "options") {
			if err := parseOptions(lines, i, &template.Options); err != nil {
				return nil, err
			}
			continue
		}

		// 解析 type 定义
		if strings.HasPrefix(line, "type ") {
			// 检查是否是 type ( ) 格式
			if strings.Contains(line, "(") {
				// 解析 type ( xxxReq { } ) 格式
				nextIndex, err := parseTypeGroup(lines, i, template)
				if err != nil {
					return nil, err
				}
				// 跳过已解析的行
				i = nextIndex
				continue
			}

			// 原有的单个 type 定义格式
			typeName := strings.TrimSpace(strings.TrimPrefix(line, "type "))
			// 移除可能的大括号和 struct 关键字
			typeName = strings.TrimSpace(strings.Trim(typeName, "{}"))
			// 如果包含 struct 关键字，只取前面的部分
			if strings.Contains(typeName, " struct") {
				typeName = strings.TrimSpace(strings.Split(typeName, " struct")[0])
			}
			currentType = &Type{Name: typeName, Fields: make([]Field, 0)}
			template.Types = append(template.Types, *currentType)
			inType = true
			continue
		}

		// 解析 service 定义
		if strings.HasPrefix(line, "service ") {
			serviceLine := strings.TrimSpace(strings.TrimPrefix(line, "service "))
			// 移除可能的大括号
			serviceLine = strings.TrimSpace(strings.Trim(serviceLine, "{}"))

			// 解析服务名和可选的 prefix
			parts := strings.Fields(serviceLine)
			serviceName := parts[0]
			prefix := ""

			// 检查是否有 prefix 参数
			if len(parts) >= 3 && parts[1] == "prefix" {
				prefix = parts[2]
			}

			currentService = &Service{
				Name:        serviceName,
				Prefix:      prefix,
				Middleware:  make([]string, 0),
				Methods:     make([]Method, 0),
				RouteGroups: make([]RouteGroup, 0),
			}
			template.Services = append(template.Services, *currentService)
			inType = false
			inTypeGroup = false
			continue
		}

		// 解析路由分组定义（在 service 内部）
		if strings.HasPrefix(line, "group ") && currentService != nil {
			// 解析 group @admin /v1/admin 格式
			groupLine := strings.TrimSpace(strings.TrimPrefix(line, "group "))
			groupLine = strings.TrimSpace(strings.Trim(groupLine, "{}"))

			parts := strings.Fields(groupLine)
			var groupName, groupPath string
			if len(parts) >= 2 {
				groupName = strings.TrimPrefix(parts[0], "@")
				groupPath = parts[1]
			} else if len(parts) == 1 {
				groupPath = parts[0]
				groupName = strings.ReplaceAll(strings.TrimPrefix(groupPath, "/"), "/", "")
			}

			currentRouteGroup = &RouteGroup{
				Name:    groupName,
				Path:    groupPath,
				Methods: make([]Method, 0),
			}

			// 添加到当前服务的路由分组
			for i, s := range template.Services {
				if s.Name == currentService.Name {
					template.Services[i].RouteGroups = append(template.Services[i].RouteGroups, *currentRouteGroup)
					break
				}
			}
			inType = false
			continue
		}

		// 检查是否结束当前路由分组
		if currentRouteGroup != nil && strings.HasPrefix(line, "}") {
			currentRouteGroup = nil
			continue
		}

		// 解析独立的路由分组定义（不在 service 内部）
		if strings.HasPrefix(line, "group ") && currentService == nil {
			groupPath := strings.TrimSpace(strings.TrimPrefix(line, "group "))
			groupPath = strings.TrimSpace(strings.Trim(groupPath, "{}"))
			currentRouteGroup = &RouteGroup{Path: groupPath, Methods: make([]Method, 0)}
			template.RouteGroups = append(template.RouteGroups, *currentRouteGroup)
			inType = false
			continue
		}

		// 解析独立路由（以 @ 开头但不在 service 或 group 中）
		if strings.HasPrefix(line, "@") && currentService == nil && currentRouteGroup == nil {
			method, err := parseMethod(line)
			if err != nil {
				return nil, err
			}
			template.StandaloneRoutes = append(template.StandaloneRoutes, StandaloneRoute{Method: method})
			continue
		}

		// 解析字段定义
		if inType && !inTypeGroup && currentType != nil && strings.Contains(line, " ") && !strings.HasPrefix(line, "}") {
			field, err := parseField(line)
			if err != nil {
				return nil, err
			}
			// 更新当前类型
			for i, t := range template.Types {
				if t.Name == currentType.Name {
					template.Types[i].Fields = append(template.Types[i].Fields, field)
					break
				}
			}
			continue
		}

		// 解析服务级别中间件
		if currentService != nil && currentRouteGroup == nil && strings.HasPrefix(line, "middleware:") {
			middlewareStr := strings.TrimSpace(strings.TrimPrefix(line, "middleware:"))
			middlewareStr = strings.TrimSpace(strings.Trim(middlewareStr, "[]"))
			if middlewareStr != "" {
				middlewareParts := strings.Split(middlewareStr, ",")
				for _, part := range middlewareParts {
					// 忽略 // 后面的注释内容
					if commentIndex := strings.Index(part, "//"); commentIndex != -1 {
						part = part[:commentIndex]
					}
					part = strings.TrimSpace(strings.Trim(part, `"'`))
					// 过滤中文注释和空字符串
					if part != "" && !strings.Contains(part, "只包含") {
						// 更新当前服务的中间件
						for i, s := range template.Services {
							if s.Name == currentService.Name {
								template.Services[i].Middleware = append(template.Services[i].Middleware, part)
								break
							}
						}
					}
				}
			}
			continue
		}

		// 解析路由分组中间件
		if currentRouteGroup != nil && strings.HasPrefix(line, "middleware:") {
			middlewareStr := strings.TrimSpace(strings.TrimPrefix(line, "middleware:"))
			middlewareStr = strings.TrimSpace(strings.Trim(middlewareStr, "[]"))
			if middlewareStr != "" {
				middlewareParts := strings.Split(middlewareStr, ",")
				for _, part := range middlewareParts {
					// 忽略 // 后面的注释内容
					if commentIndex := strings.Index(part, "//"); commentIndex != -1 {
						part = part[:commentIndex]
					}
					part = strings.TrimSpace(strings.Trim(part, `"'`))
					// 过滤中文注释和空字符串
					if part != "" && !strings.Contains(part, "只包含") {
						// 更新当前路由分组的中间件
						if currentService != nil {
							// 在服务内部的分组
							for i, s := range template.Services {
								if s.Name == currentService.Name {
									for j, g := range s.RouteGroups {
										if g.Name == currentRouteGroup.Name && g.Path == currentRouteGroup.Path {
											template.Services[i].RouteGroups[j].Middleware = append(template.Services[i].RouteGroups[j].Middleware, part)
											break
										}
									}
									break
								}
							}
						} else {
							// 独立的分组
							for i, g := range template.RouteGroups {
								if g.Path == currentRouteGroup.Path {
									template.RouteGroups[i].Middleware = append(template.RouteGroups[i].Middleware, part)
									break
								}
							}
						}
					}
				}
			}
			continue
		}

		// 解析方法定义
		if (currentService != nil || currentRouteGroup != nil) && !inType {
			if strings.Contains(line, "@") {
				method, err := parseMethod(line)
				if err != nil {
					return nil, err
				}

				// 更新当前服务（只有在不在路由分组内时才添加）
				if currentService != nil && currentRouteGroup == nil {
					for i, s := range template.Services {
						if s.Name == currentService.Name {
							template.Services[i].Methods = append(template.Services[i].Methods, method)
							break
						}
					}
				}

				// 更新当前路由分组
				if currentRouteGroup != nil {
					if currentService != nil {
						// 在服务内部的分组
						for i, s := range template.Services {
							if s.Name == currentService.Name {
								for j, g := range s.RouteGroups {
									if g.Name == currentRouteGroup.Name && g.Path == currentRouteGroup.Path {
										template.Services[i].RouteGroups[j].Methods = append(template.Services[i].RouteGroups[j].Methods, method)
										break
									}
								}
								break
							}
						}
					} else {
						// 独立的分组
						for i, g := range template.RouteGroups {
							if g.Path == currentRouteGroup.Path {
								template.RouteGroups[i].Methods = append(template.RouteGroups[i].Methods, method)
								break
							}
						}
					}
				}
				continue
			}
		}
	}

	return template, nil
}

func parseInfo(lines []string, start int, info *Info) error {
	for i := start + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "}") {
			break
		}
		if strings.HasPrefix(line, "//") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(strings.Trim(parts[1], `"`))

		switch key {
		case "title":
			info.Title = value
		case "version":
			info.Version = value
		case "desc":
			info.Desc = value
		}
	}
	return nil
}

func parseOptions(lines []string, start int, options *Options) error {
	for i := start + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" || strings.HasPrefix(line, "}") {
			break
		}
		if strings.HasPrefix(line, "//") {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(strings.TrimSuffix(parts[0], ":"))
		value := strings.TrimSpace(strings.Trim(parts[1], `"`))

		switch key {
		case "withGinContext":
			options.WithGinContext = value == "true"
		case "outputDir":
			options.OutputDir = value
		case "packageName":
			options.PackageName = value
		case "serviceOutputDir":
			options.ServiceOutputDir = value
		case "generateService":
			options.GenerateService = value == "true"
		}
	}
	return nil
}

func parseField(line string) (Field, error) {
	// 解析字段格式: name type `tag` // comment
	// 支持数组类型如 []User, map[string]interface{} 等
	// 支持指针类型如 *User 等
	re := regexp.MustCompile(`(\w+)\s+([\w\[\]{}.,*]+)\s+` + "`([^`]*)`" + `\s*(?://\s*(.*))?`)
	matches := re.FindStringSubmatch(line)
	if len(matches) < 4 {
		return Field{}, fmt.Errorf("invalid field format: %s", line)
	}

	field := Field{
		Name: matches[1],
		Type: matches[2],
		Tag:  matches[3],
	}

	// 检查是否必填
	field.Required = strings.Contains(field.Tag, "required")

	if len(matches) > 4 {
		field.Comment = strings.TrimSpace(matches[4])
	}

	return field, nil
}

func parseMethod(line string) (Method, error) {
	// 解析方法格式: @method httpMethod path [WithGinContext] request response [middleware: ["mw1", "mw2"]] // comment

	// 提取中间件部分
	middleware := []string{}
	if strings.Contains(line, "middleware:") {
		reMiddleware := regexp.MustCompile(`middleware:\s*\[([^\]]+)\]`)
		matches := reMiddleware.FindStringSubmatch(line)
		if len(matches) > 1 {
			middlewareStr := strings.TrimSpace(matches[1])
			if middlewareStr != "" {
				// 解析中间件列表
				middlewareParts := strings.Split(middlewareStr, ",")
				for _, part := range middlewareParts {
					part = strings.TrimSpace(strings.Trim(part, `"'`))
					if part != "" {
						middleware = append(middleware, part)
					}
				}
			}
		}
		// 移除中间件部分以便后续解析
		line = reMiddleware.ReplaceAllString(line, "")
	}

	// 先尝试解析包含 WithGinContext 的格式
	reWithGin := regexp.MustCompile(`@(\w+)\s+(\w+)\s+(\S+)\s+WithGinContext\s+(\w+)\s+(\w+)\s*(?://\s*(.*))?`)
	matches := reWithGin.FindStringSubmatch(line)
	if len(matches) >= 6 {
		method := Method{
			Name:           matches[1],
			HTTPMethod:     strings.ToUpper(matches[2]),
			Path:           matches[3],
			Request:        matches[4],
			Response:       matches[5],
			WithGinContext: true,
			Middleware:     middleware,
		}
		if len(matches) > 6 {
			method.Description = strings.TrimSpace(matches[6])
		}
		return method, nil
	}

	// 如果没有 WithGinContext，解析普通格式
	reNormal := regexp.MustCompile(`@(\w+)\s+(\w+)\s+(\S+)\s+(\w+)\s+(\w+)\s*(?://\s*(.*))?`)
	matches = reNormal.FindStringSubmatch(line)
	if len(matches) >= 6 {
		method := Method{
			Name:           matches[1],
			HTTPMethod:     strings.ToUpper(matches[2]),
			Path:           matches[3],
			Request:        matches[4],
			Response:       matches[5],
			WithGinContext: false,
			Middleware:     middleware,
		}
		if len(matches) > 6 {
			method.Description = strings.TrimSpace(matches[6])
		}
		return method, nil
	}

	return Method{}, fmt.Errorf("invalid method format: %s", line)
}

// parseTypeGroup 解析 type ( ) 格式的类型定义组
func parseTypeGroup(lines []string, start int, template *GinTemplate) (int, error) {
	// 找到 type ( 行
	line := strings.TrimSpace(lines[start])
	if !strings.HasPrefix(line, "type ") || !strings.Contains(line, "(") {
		return start, fmt.Errorf("invalid type group format")
	}

	// 从下一行开始解析，直到找到对应的 )
	for i := start + 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// 如果遇到 )，说明类型组结束
		if line == ")" {
			return i, nil
		}

		// 解析单个类型定义
		if strings.Contains(line, "{") {
			// 解析 xxxReq { } 格式
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				typeName := parts[0]
				// 移除可能的大括号和 struct 关键字
				typeName = strings.TrimSpace(strings.Trim(typeName, "{}"))
				// 如果包含 struct 关键字，只取前面的部分
				if strings.Contains(typeName, " struct") {
					typeName = strings.TrimSpace(strings.Split(typeName, " struct")[0])
				}

				// 创建类型
				currentType := &Type{Name: typeName, Fields: make([]Field, 0)}

				// 检查是否在同一行有字段定义
				if strings.Contains(line, "{") && strings.Contains(line, "}") {
					// 单行类型定义，没有字段
					template.Types = append(template.Types, *currentType)
				} else if strings.Contains(line, "{") && !strings.Contains(line, "}") {
					// 多行类型定义，继续解析字段
					for j := i + 1; j < len(lines); j++ {
						fieldLine := strings.TrimSpace(lines[j])

						// 跳过空行和注释
						if fieldLine == "" || strings.HasPrefix(fieldLine, "//") {
							continue
						}

						// 如果遇到 }，说明字段定义结束
						if fieldLine == "}" {
							break
						}

						// 解析字段
						if strings.Contains(fieldLine, " ") && !strings.HasPrefix(fieldLine, "}") {
							field, err := parseField(fieldLine)
							if err != nil {
								return start, err
							}
							currentType.Fields = append(currentType.Fields, field)
						}
					}
					template.Types = append(template.Types, *currentType)
				}
			}
		}
	}

	return start, nil
}
