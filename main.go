package main

import (
	"context"
	"fmt"
	"os"

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
	
	// 尝试加载默认配置并自动连接
	tryAutoConnect()
	
	// 执行主程序逻辑
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "执行出错:", err)
		os.Exit(1)
	}
}

// tryAutoConnect 尝试使用默认配置自动连接数据库
func tryAutoConnect() {
	// 检查是否已经通过命令行参数指定了连接（如果是，则跳过自动连接）
	cmdArgs := os.Args
	if len(cmdArgs) > 1 && cmdArgs[1] == "connect" {
		return
	}

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