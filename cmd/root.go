package cmd

import (
	"context"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/yuanpli/datamgr-cli/pkg/prompt"
)

var rootCmd = &cobra.Command{
	Use:   filepath.Base(os.Args[0]),
	Short: "通用CLI数据管理工具",
	Long:  `一个支持多种数据库的通用数据管理命令行工具，提供统一的表管理操作接口。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 启动交互式命令行
		prompt.Start()
	},
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteContext 带上下文执行根命令
func ExecuteContext(ctx context.Context) error {
	rootCmd.SetContext(ctx)
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(connectCmd)
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(configCmd)
} 