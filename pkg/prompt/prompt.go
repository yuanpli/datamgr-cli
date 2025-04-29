package prompt

import (
	"fmt"
	"os"
	"os/signal"
	"sort"
	"strings"
	"syscall"

	"github.com/bwty/bwty-data-cli/db"
	"github.com/bwty/bwty-data-cli/pkg/handler"
	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
)

var (
	successPrefix = "✓ "
	errorPrefix   = "✗ "
	dbPrompt      = "db> "
)

// ExecuteCommand 执行命令
func ExecuteCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}

	// 判断是否为退出命令
	if strings.ToLower(cmd) == "exit" || strings.ToLower(cmd) == "quit" {
		fmt.Println("再见！")
		
		// 清理资源
		if conn := db.GetCurrentConnection(); conn != nil {
			conn.Disconnect()
		}
		
		os.Exit(0)
	}

	// 删除末尾的分号（如果有）
	cmd = strings.TrimSuffix(cmd, ";")

	// 根据命令的第一个词来确定要调用的处理函数
	cmdParts := strings.Fields(cmd)
	if len(cmdParts) == 0 {
		return
	}

	// 命令处理
	var err error
	switch strings.ToLower(cmdParts[0]) {
	case "help":
		handler.HandleHelp()
	case "clear":
		handler.HandleClear()
	case "status":
		err = handler.HandleStatus()
	case "connect":
		err = handler.HandleConnect(cmd)
	case "show":
		if len(cmdParts) > 1 && strings.ToLower(cmdParts[1]) == "tables" {
			err = handler.HandleShowTables()
		} else {
			fmt.Println("未知的 show 命令。尝试使用 'show tables'。")
		}
	case "desc", "describe":
		if len(cmdParts) > 2 && strings.ToLower(cmdParts[1]) == "table" {
			err = handler.HandleDescribeTable(cmdParts[2])
		} else {
			fmt.Println("用法: desc table <表名>")
		}
	case "select", "insert", "update", "delete":
		err = executeSQL(cmd)
	case "import":
		err = handler.HandleImport(cmd)
	case "export":
		err = handler.HandleExport(cmd)
	default:
		fmt.Printf("未知命令: %s\n", cmd)
	}

	if err != nil {
		fmt.Printf("%s%v\n", color.RedString(errorPrefix), err)
	}
}

// completer 命令自动补全
func completer(d prompt.Document) []prompt.Suggest {
	s := []prompt.Suggest{
		{Text: "help", Description: "显示帮助信息"},
		{Text: "exit", Description: "退出程序"},
		{Text: "quit", Description: "退出程序"},
		{Text: "clear", Description: "清屏"},
		{Text: "status", Description: "显示连接状态"},
		{Text: "connect", Description: "连接到数据库"},
		{Text: "show tables", Description: "列出所有表"},
		{Text: "desc table", Description: "显示表结构"},
		{Text: "select", Description: "查询数据"},
		{Text: "insert", Description: "插入数据"},
		{Text: "update", Description: "更新数据"},
		{Text: "delete", Description: "删除数据"},
		{Text: "import", Description: "导入数据"},
		{Text: "export", Description: "导出数据"},
	}

	// 添加表名补全
	if strings.HasPrefix(d.TextBeforeCursor(), "desc table ") ||
		strings.HasPrefix(d.TextBeforeCursor(), "select * from ") {
		conn := db.GetCurrentConnection()
		if conn != nil {
			tables, err := conn.GetTables()
			if err == nil {
				for _, table := range tables {
					s = append(s, prompt.Suggest{
						Text:        table,
						Description: "表名",
					})
				}
			}
		}
	}

	return prompt.FilterHasPrefix(s, d.GetWordBeforeCursor(), true)
}

// getPrompt 获取命令提示符
func getPrompt() (string, bool) {
	config := db.GetCurrentConfig()
	if config != nil {
		return fmt.Sprintf("db[%s]> ", config.DbName), true
	}
	return dbPrompt, true
}

// setupSignalHandler 设置信号处理器
func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		fmt.Println("\n收到终止信号，程序退出...")
		// 清理资源，如断开数据库连接
		if conn := db.GetCurrentConnection(); conn != nil {
			conn.Disconnect()
		}
		os.Exit(0)
	}()
}

// Start 启动交互式命令行
func Start() {
	fmt.Println("欢迎使用通用数据管理工具！输入 'help' 查看帮助信息。")
	fmt.Println("输入 'exit' 或 'quit' 退出程序")
	fmt.Println("按 Ctrl+C 也可以终止程序")
	
	// 设置信号处理
	setupSignalHandler()

	p := prompt.New(
		ExecuteCommand,
		completer,
		prompt.OptionPrefix(dbPrompt),
		prompt.OptionTitle("BWTY 数据管理工具"),
		prompt.OptionLivePrefix(getPrompt),
		prompt.OptionInputTextColor(prompt.Blue),
		prompt.OptionPrefixTextColor(prompt.Blue),
		// 下面的选项提高了终端兼容性
		prompt.OptionMaxSuggestion(8),
		prompt.OptionSuggestionBGColor(prompt.LightGray),
		prompt.OptionSuggestionTextColor(prompt.Black),
		prompt.OptionDescriptionBGColor(prompt.White),
		prompt.OptionDescriptionTextColor(prompt.Black),
		// 启用中断处理
		prompt.OptionAddKeyBind(
			prompt.KeyBind{
				Key: prompt.ControlC,
				Fn: func(buf *prompt.Buffer) {
					fmt.Println("\n程序被中断，正在退出...")
					if conn := db.GetCurrentConnection(); conn != nil {
						conn.Disconnect()
					}
					os.Exit(0)
				},
			},
		),
	)

	p.Run()
}

// executeSQL 执行SQL语句
func executeSQL(sql string) error {
	conn := db.GetCurrentConnection()
	if conn == nil {
		return fmt.Errorf("当前未连接到任何数据库")
	}

	// 判断是查询语句还是更新语句
	sqlLower := strings.ToLower(sql)
	if strings.HasPrefix(sqlLower, "select") {
		// 尝试解析SQL获取表名
		var tableName string
		fromIndex := strings.Index(sqlLower, "from")
		if fromIndex > 0 {
			afterFrom := sqlLower[fromIndex+4:]
			parts := strings.Fields(afterFrom)
			if len(parts) > 0 {
				tableName = parts[0]
			}
		}

		// 获取表的列顺序
		var orderedColumns []string
		var err error
		if tableName != "" {
			orderedColumns, err = conn.GetTableColumns(tableName)
			if err != nil {
				// 如果获取列顺序失败，不中断执行，只是不排序
				orderedColumns = nil
			}
		}

		// 执行查询
		results, err := conn.Query(sql)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("查询没有返回结果")
			return nil
		}

		// 获取所有列名
		var columns []string
		if orderedColumns != nil {
			// 使用表的原始顺序
			for _, col := range orderedColumns {
				// 检查查询结果中是否包含此列
				if _, exists := results[0][col]; exists {
					columns = append(columns, col)
				}
			}
			
			// 处理查询结果中可能存在但未在获取的列顺序中的列
			for col := range results[0] {
				found := false
				for _, orderedCol := range columns {
					if orderedCol == col {
						found = true
						break
					}
				}
				if !found {
					columns = append(columns, col)
				}
			}
		} else {
			// 没有获取到表的列顺序，使用map的键
			for col := range results[0] {
				columns = append(columns, col)
			}
			// 字母排序以保证每次显示顺序一致
			sort.Strings(columns)
		}

		// 打印表头
		headerFmt := ""
		for range columns {
			headerFmt += fmt.Sprintf("%%-%ds ", 20)
		}
		headerFmt += "\n"

		fmt.Printf(headerFmt, interfaceSlice(columns)...)
		fmt.Println(strings.Repeat("-", 20*len(columns)))

		// 打印每一行
		rowFmt := ""
		for range columns {
			rowFmt += fmt.Sprintf("%%-%dv ", 20)
		}
		rowFmt += "\n"

		for _, row := range results {
			values := make([]interface{}, len(columns))
			for i, col := range columns {
				values[i] = row[col]
			}
			fmt.Printf(rowFmt, values...)
		}

		fmt.Printf("\n共 %d 行结果\n", len(results))

	} else {
		// 执行更新操作
		affected, err := conn.Execute(sql)
		if err != nil {
			return err
		}
		
		fmt.Printf("操作成功，影响了 %d 行数据\n", affected)
	}

	return nil
}

// interfaceSlice 将字符串切片转换为接口切片
func interfaceSlice(slice []string) []interface{} {
	iSlice := make([]interface{}, len(slice))
	for i, v := range slice {
		iSlice[i] = v
	}
	return iSlice
} 