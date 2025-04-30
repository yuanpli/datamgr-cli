package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yuanpli/datamgr-cli/cmd"
	"github.com/yuanpli/datamgr-cli/db"
	"github.com/yuanpli/datamgr-cli/pkg/utils"
)

func main() {
	// 创建一个带取消的上下文
	ctx, cancel := context.WithCancel(context.Background())
	
	// 设置信号处理
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, syscall.SIGTERM)
	
	// 在后台处理信号
	go func() {
		sig := <-signalChan
		fmt.Printf("\n收到信号 %s，程序正在退出...\n", sig)
		
		// 断开数据库连接
		if conn := db.GetCurrentConnection(); conn != nil {
			fmt.Println("正在断开数据库连接...")
			conn.Disconnect()
		}
		
		// 触发取消
		cancel()
		
		// 给一点时间进行清理
		os.Exit(0)
	}()
	
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