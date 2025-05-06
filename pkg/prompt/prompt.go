package prompt

import (
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"

	"github.com/c-bata/go-prompt"
	"github.com/fatih/color"
	"github.com/yuanpli/datamgr-cli/db"
	"github.com/yuanpli/datamgr-cli/pkg/handler"
	"github.com/yuanpli/datamgr-cli/pkg/utils"
)

var (
	successPrefix = "✓ "
	errorPrefix   = "✗ "
	dbPrompt      = "datamgr> "
)

// cleanExit 处理程序退出前的清理工作
func cleanExit(message string, exitCode int) {
	if message != "" {
		fmt.Println(message)
	}
	
	// 清理资源
	if conn := db.GetCurrentConnection(); conn != nil {
		conn.Disconnect()
	}
	
	// 恢复终端状态
	handler.Close()
	
	os.Exit(exitCode)
}

// ExecuteCommand 执行命令
func ExecuteCommand(cmd string) {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return
	}

	// 判断是否为退出命令
	if strings.ToLower(cmd) == "exit" || strings.ToLower(cmd) == "quit" {
		cleanExit("再见！", 0)
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
	case "config":
		err = handleConfig(cmdParts[1:])
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

// 保存当前连接为配置
func saveCurrentConnectionAsConfig() error {
	currentConfig := db.GetCurrentConfig()
	if currentConfig == nil {
		return fmt.Errorf("当前未连接到任何数据库，无法保存配置")
	}

	config := &utils.Config{
		Type:     currentConfig.Type,
		Host:     currentConfig.Host,
		Port:     currentConfig.Port,
		User:     currentConfig.User,
		Password: currentConfig.Password,
		DbName:   currentConfig.DbName,
	}

	if err := utils.SaveConfig(config); err != nil {
		return fmt.Errorf("保存配置失败: %v", err)
	}

	fmt.Println("已将当前连接信息保存为默认配置")
	return nil
}

// 处理配置相关命令
func handleConfig(args []string) error {
	if len(args) == 0 {
		// 无参数，显示当前配置
		_, err := utils.DisplayConfig()
		if err != nil {
			return fmt.Errorf("加载配置失败: %v", err)
		}
		return nil
	}

	switch strings.ToLower(args[0]) {
	case "save":
		// 保存当前连接为默认配置
		return saveCurrentConnectionAsConfig()

	case "clear":
		// 清除配置
		if err := utils.ClearConfig(); err != nil {
			return fmt.Errorf("清除配置失败: %v", err)
		}
		fmt.Println("已清除默认配置")
		return nil

	case "set":
		// 更新配置
		if len(args) < 2 {
			return fmt.Errorf("用法: config set [type|host|port|user|password|dbname] 值")
		}

		// 先尝试加载现有配置
		var config *utils.Config
		existingConfig, err := utils.LoadConfig()
		if err == nil {
			// 有现有配置，以它为基础修改
			config = existingConfig
		} else {
			// 没有现有配置，创建新的
			config = &utils.Config{
				Type: "dameng", // 默认使用达梦数据库
				Port: 5236,     // 默认端口
			}
		}

		// 根据参数设置不同的字段
		if len(args) >= 3 {
			switch strings.ToLower(args[1]) {
			case "type":
				config.Type = args[2]
			case "host":
				config.Host = args[2]
			case "port":
				port, err := strconv.Atoi(args[2])
				if err != nil {
					return fmt.Errorf("端口格式错误: %v", err)
				}
				config.Port = port
			case "user":
				config.User = args[2]
			case "password":
				config.Password = args[2]
			case "dbname":
				config.DbName = args[2]
			default:
				return fmt.Errorf("未知的配置项: %s", args[1])
			}

			// 保存配置
			if err := utils.SaveConfig(config); err != nil {
				return fmt.Errorf("保存配置失败: %v", err)
			}
			fmt.Printf("配置已更新: %s = %s\n", args[1], args[2])
		} else {
			return fmt.Errorf("需要指定配置项的值")
		}
		return nil

	default:
		return fmt.Errorf("未知的config子命令: %s", args[0])
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
		{Text: "config", Description: "管理默认连接配置"},
		{Text: "config save", Description: "保存当前连接为默认配置"},
		{Text: "config set", Description: "修改默认配置"},
		{Text: "config clear", Description: "清除默认配置"},
		{Text: "show tables", Description: "列出所有表"},
		{Text: "desc table", Description: "显示表结构"},
		{Text: "select", Description: "查询数据"},
		{Text: "insert", Description: "插入数据"},
		{Text: "update", Description: "更新数据"},
		{Text: "delete", Description: "删除数据"},
		{Text: "import", Description: "导入数据"},
		{Text: "export", Description: "导出数据"},
	}

	// 添加config set子命令补全
	if strings.HasPrefix(d.TextBeforeCursor(), "config set ") {
		configItems := []prompt.Suggest{
			{Text: "type", Description: "设置数据库类型"},
			{Text: "host", Description: "设置主机地址"},
			{Text: "port", Description: "设置端口"},
			{Text: "user", Description: "设置用户名"},
			{Text: "password", Description: "设置密码"},
			{Text: "dbname", Description: "设置数据库名"},
		}
		return prompt.FilterHasPrefix(configItems, d.GetWordBeforeCursor(), true)
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
		return fmt.Sprintf("datamgr[%s]> ", config.DbName), true
	}
	return dbPrompt, true
}

// setupSignalHandler 设置信号处理器
func setupSignalHandler() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-c
		cleanExit("\n收到终止信号，程序退出...", 0)
	}()
}

// Start 启动交互式命令行
func Start() {
	fmt.Println("欢迎使用通用数据管理工具！输入 'help' 查看帮助信息。")
	fmt.Println("输入 'exit' 或 'quit' 退出程序")
	fmt.Println("按 Ctrl+C 也可以终止程序")
	
	// 检查是否已经连接到数据库（通过默认配置）
	if config := db.GetCurrentConfig(); config != nil {
		fmt.Printf("当前已连接到 %s 数据库: %s\n", config.Type, config.DbName)
	} else {
		// 提示用户连接数据库
		fmt.Println("当前未连接到数据库，请使用 'connect' 命令连接")
		
		// 检查是否有默认配置可用
		defaultConfig, err := utils.LoadConfig()
		if err == nil && defaultConfig != nil {
			fmt.Println("发现默认配置信息，可以使用 'connect' 命令快速连接")
		}
	}
	
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
					cleanExit("\n程序被中断，正在退出...", 0)
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
		// 直接执行查询
		results, err := conn.Query(sql)
		if err != nil {
			return err
		}

		if len(results) == 0 {
			fmt.Println("查询没有返回结果")
			return nil
		}

		// 处理列排序
		isSelectStar := isSelectAllQuery(sqlLower)
		columns := getOrderedColumns(results[0], sqlLower, conn, isSelectStar)

		// 打印表头
		headerFmt := ""
		for range columns {
			headerFmt += fmt.Sprintf("%%-%ds ", 20)
		}
		headerFmt += "\n"

		// 将列名转换为接口切片
		headerValues := make([]interface{}, len(columns))
		for i, v := range columns {
			headerValues[i] = v
		}
		
		fmt.Printf(headerFmt, headerValues...)
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
		// 直接执行更新操作
		affected, err := conn.Execute(sql)
		if err != nil {
			return err
		}
		
		fmt.Printf("操作成功，影响了 %d 行数据\n", affected)
	}

	return nil
}

// isSelectAllQuery 判断是否为SELECT *查询
func isSelectAllQuery(sqlLower string) bool {
	// 去除多余空格
	sqlLower = strings.TrimSpace(sqlLower)
	
	// 检查是否以SELECT *开头
	if strings.HasPrefix(sqlLower, "select *") {
		return true
	}
	
	// 检查是否有SELECT和FROM之间只有*（可能有空格）
	selectIndex := strings.Index(sqlLower, "select")
	fromIndex := strings.Index(sqlLower, "from")
	
	if selectIndex >= 0 && fromIndex > selectIndex {
		between := strings.TrimSpace(sqlLower[selectIndex+6:fromIndex])
		if between == "*" {
			return true
		}
	}
	
	return false
}

// getTableNameFromSQL 从SQL语句中提取表名
func getTableNameFromSQL(sqlLower string) string {
	fromIndex := strings.Index(sqlLower, "from")
	if fromIndex < 0 {
		return ""
	}
	
	afterFrom := sqlLower[fromIndex+4:]
	parts := strings.Fields(afterFrom)
	if len(parts) == 0 {
		return ""
	}
	
	// 处理表名可能有的别名、WHERE子句等
	tableName := parts[0]
	// 移除可能的逗号、括号等
	tableName = strings.TrimRight(tableName, ",();")
	
	return tableName
}

// getOrderedColumns 根据查询类型获取有序的列名
func getOrderedColumns(resultRow map[string]interface{}, sqlLower string, conn db.Connection, isSelectStar bool) []string {
	// 如果不是SELECT *查询，保持原始顺序
	if !isSelectStar {
		var columns []string
		for col := range resultRow {
			columns = append(columns, col)
		}
		return columns
	}
	
	// 对于SELECT *查询，尝试按表结构排序
	tableName := getTableNameFromSQL(sqlLower)
	if tableName == "" {
		// 无法确定表名，使用原始顺序
		var columns []string
		for col := range resultRow {
			columns = append(columns, col)
		}
		return columns
	}
	
	// 获取表的列顺序
	tableColumns, err := conn.GetTableColumns(tableName)
	if err != nil || len(tableColumns) == 0 {
		// 获取列顺序失败，使用原始顺序
		var columns []string
		for col := range resultRow {
			columns = append(columns, col)
		}
		return columns
	}
	
	// 使用表结构顺序排序结果列
	var orderedColumns []string
	
	// 首先添加按表结构顺序的列
	for _, col := range tableColumns {
		if _, exists := resultRow[col]; exists {
			orderedColumns = append(orderedColumns, col)
		}
	}
	
	// 添加可能的额外列（不在表结构中的列）
	for col := range resultRow {
		found := false
		for _, orderedCol := range orderedColumns {
			if col == orderedCol {
				found = true
				break
			}
		}
		if !found {
			orderedColumns = append(orderedColumns, col)
		}
	}
	
	return orderedColumns
} 