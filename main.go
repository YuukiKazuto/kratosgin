package main

import (
	"fmt"
	"os"

	"github.com/YuukiKazuto/kratosgin/internal/cli"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "kratosgin",
	Short: "Kratos Gin 代码生成器",
	Long:  "一个用于生成 Kratos 框架使用 Gin 路由的 API 代码生成器",
}

func init() {
	// 添加子命令
	rootCmd.AddCommand(cli.GenCommand())
	rootCmd.AddCommand(cli.NewCommand())
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
