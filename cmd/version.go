package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// 版本信息变量
var (
	Version   = "0.1.0"
	BuildTime = "未知"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "显示版本信息",
	Run: func(cmd *cobra.Command, args []string) {
		programName := filepath.Base(os.Args[0])
		fmt.Printf("%s 版本: %s\n", programName, Version)
		fmt.Printf("构建时间: %s\n", BuildTime)
	},
} 