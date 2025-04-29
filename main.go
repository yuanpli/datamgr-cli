package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bwty/bwty-data-cli/cmd"
	"github.com/bwty/bwty-data-cli/db"
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
	
	// 执行主程序逻辑
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, "执行出错:", err)
		os.Exit(1)
	}
} 