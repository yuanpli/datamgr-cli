package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/yuanpli/datamgr-cli/cmd"
	"github.com/yuanpli/datamgr-cli/db"
	"github.com/yuanpli/datamgr-cli/pkg/handler"
	"github.com/yuanpli/datamgr-cli/pkg/utils"
)

func main() {
	// 创建一个带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	
	// 确保在程序结束时关闭readline
	defer handler.Close()
	
	// 仅当需要时才连接数据库
	shouldAutoConnect := shouldConnectDatabase()
	if shouldAutoConnect {
		tryAutoConnect()
	}
	
	// 执行主程序逻辑
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "执行出错:", err)
		os.Exit(1)
	}
}

// shouldConnectDatabase 判断是否应该自动连接数据库
func shouldConnectDatabase() bool {
	// 检查命令行参数
	cmdArgs := os.Args
	if len(cmdArgs) <= 1 {
		// 如果没有命令参数，将进入交互模式，此时不需要预先连接
		return false
	}

	// 这些命令不需要连接数据库
	noConnectCommands := map[string]bool{
		"help":    true,
		"version": true,
		"connect": true, // connect命令会自己处理连接
		"config":  true, // config命令通常不需要连接
		"-h":      true,
		"--help":  true,
	}
	
	// 获取主命令（第一个参数）
	command := cmdArgs[1]
	if noConnectCommands[command] {
		return false
	}
	
	// 检查是否有帮助标志
	for _, arg := range cmdArgs[1:] {
		if arg == "--help" || arg == "-h" {
			return false
		}
	}
	
	// 如果命令是查询相关的SQL命令（如SELECT, SHOW等），则需要连接
	sqlCommands := []string{"select", "show", "desc", "insert", "update", "delete", "export", "import"}
	command = strings.ToLower(command)
	for _, sqlCmd := range sqlCommands {
		if command == sqlCmd || strings.HasPrefix(command, sqlCmd+" ") {
			return true
		}
	}
	
	// 默认情况下，其他命令可能需要连接数据库
	return true
}

// tryAutoConnect 尝试使用默认配置自动连接数据库
func tryAutoConnect() {
	// 尝试加载默认配置
	defaultConfig, err := utils.LoadConfig()
	if err != nil {
		// 没有默认配置，使用程序的普通流程
		return
	}

	// 使用默认配置尝试连接
	err = db.Connect(defaultConfig.Type, defaultConfig.Host, defaultConfig.Port, 
		defaultConfig.User, defaultConfig.Password, defaultConfig.DbName)
	if err != nil {
		fmt.Printf("使用默认配置连接失败: %v\n", err)
		fmt.Println("请使用 'connect' 命令手动连接数据库")
		return
	}

	fmt.Printf("已使用默认配置连接到 %s 数据库: %s\n", defaultConfig.Type, defaultConfig.DbName)
} 