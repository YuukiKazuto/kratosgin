.PHONY: build install clean test help

# 构建二进制文件
build:
	@echo "构建 kratosgin 工具..."
	@go build -o bin/kratosgin .

# 安装到系统
install: build
	@echo "安装 kratosgin 到系统..."
	@sudo cp bin/kratosgin /usr/local/bin/

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf bin/

# 运行测试
test:
	@echo "运行测试..."
	@go test ./...

# 显示帮助
help:
	@echo "Kratos Gin 代码生成器"
	@echo ""
	@echo "可用命令:"
	@echo "  build    - 构建二进制文件"
	@echo "  install  - 安装到系统"
	@echo "  clean    - 清理构建文件"
	@echo "  test     - 运行测试"
	@echo "  help     - 显示此帮助信息"
	@echo ""
	@echo "使用示例:"
	@echo "  make build"
	@echo "  ./bin/kratosgin new user"
	@echo "  ./bin/kratosgin gen -f user.gin"
	@echo "  ./bin/kratosgin gen -f user.gin -s internal/service -m internal/middleware"
	@echo ""
	@echo "测试:"
	@echo "  cd test"
	@echo "  ../bin/kratosgin new product"
	@echo "  ../bin/kratosgin gen -f product.gin"