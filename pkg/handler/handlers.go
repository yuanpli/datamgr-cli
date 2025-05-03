package handler

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"encoding/csv"

	"github.com/chzyer/readline"
	"github.com/xuri/excelize/v2"
	"github.com/yuanpli/datamgr-cli/db"
	"github.com/yuanpli/datamgr-cli/pkg/utils"
)

// HandleHelp 帮助命令处理
func HandleHelp() {
	helpText := `
可用命令:
  系统命令:
    help                   - 显示此帮助信息
    connect                - 连接数据库
    status                 - 显示连接状态
    exit, quit             - 退出程序
    clear                  - 清屏

  配置管理:
    config                 - 显示当前默认配置
    config save            - 保存当前连接为默认配置
    config set <项> <值>   - 设置默认配置项
    config clear           - 清除默认配置

  表管理命令:
    show tables            - 列出所有表
    desc table <表名>      - 显示表结构

  数据操作命令:
    SELECT [字段] FROM <表> [WHERE 条件] [LIMIT 数量]  - 查询数据
    INSERT INTO <表> SET 字段1=值1, 字段2=值2...      - 插入数据
    UPDATE <表> SET 字段=值 [WHERE 条件]             - 更新数据
    DELETE FROM <表> [WHERE 条件]                    - 删除数据
    IMPORT <表> FROM <文件> [FORMAT csv/excel]       - 导入数据
    EXPORT <表> [WHERE 条件] <文件> [FORMAT csv/excel] - 导出数据
`
	fmt.Println(helpText)
}

// HandleClear 清屏命令处理
func HandleClear() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}

// HandleStatus 状态命令处理
func HandleStatus() error {
	config := db.GetCurrentConfig()
	if config == nil {
		return errors.New("当前未连接到任何数据库")
	}

	fmt.Println("当前连接状态:")
	fmt.Printf("  数据库类型: %s\n", config.Type)
	fmt.Printf("  主机地址: %s\n", config.Host)
	fmt.Printf("  端口: %d\n", config.Port)
	fmt.Printf("  用户名: %s\n", config.User)
	fmt.Printf("  数据库名: %s\n", config.DbName)
	return nil
}

// HandleConnect 处理连接命令
func HandleConnect(cmdStr string) error {
	// 简单解析connect命令
	args := strings.Fields(cmdStr)
	if len(args) == 1 {
		// 交互式连接向导
		return handleInteractiveConnect()
	}

	// 解析命令行参数
	var dbType, host, user, password, dbName string
	var port int = 5236 // 默认端口

	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--type":
			if i+1 < len(args) {
				dbType = args[i+1]
				i++
			}
		case "-H", "--host":
			if i+1 < len(args) {
				host = args[i+1]
				i++
			}
		case "-P", "--port":
			if i+1 < len(args) {
				fmt.Sscanf(args[i+1], "%d", &port)
				i++
			}
		case "-u", "--user":
			if i+1 < len(args) {
				user = args[i+1]
				i++
			}
		case "-p", "--password":
			if i+1 < len(args) {
				password = args[i+1]
				i++
			}
		case "-D", "--dbname":
			if i+1 < len(args) {
				dbName = args[i+1]
				i++
			}
		}
	}

	// 如果没有指定数据库类型，默认为达梦
	if dbType == "" {
		dbType = "dameng"
	}

	// 验证必要参数
	if host == "" || user == "" || password == "" || dbName == "" {
		return errors.New("连接参数不完整，请提供主机、用户名、密码和数据库名")
	}

	// 执行连接
	err := db.Connect(dbType, host, port, user, password, dbName)
	if err != nil {
		return err
	}

	fmt.Printf("已成功连接到 %s 数据库: %s\n", dbType, dbName)
	return nil
}

// 初始化一个全局的readline实例
var rl *readline.Instance

// 初始化readline
func init() {
	var err error
	rl, err = readline.New("")
	if err != nil {
		fmt.Println("初始化输入处理失败:", err)
		os.Exit(1)
	}
}

// readInput 使用readline库读取用户输入
func readInput(prompt string) string {
	if rl == nil {
		// 如果无法使用readline，回退到简单输入
		fmt.Print(prompt)
		var input string
		fmt.Scanln(&input)
		return input
	}
	
	// 设置提示
	rl.SetPrompt(prompt)
	
	// 读取一行
	line, err := rl.Readline()
	if err != nil {
		if err == readline.ErrInterrupt {
			fmt.Println("^C")
			os.Exit(0)
		} else if err == io.EOF {
			return ""
		}
		fmt.Println("读取错误:", err)
		return ""
	}
	
	return strings.TrimSpace(line)
}

// readPassword 读取密码但不显示
func readPassword(prompt string) string {
	// 创建一个临时的readline实例，用于密码输入
	rlTemp, err := readline.New(prompt)
	if err != nil {
		fmt.Println("初始化密码读取失败:", err)
		return ""
	}
	defer rlTemp.Close()
	
	// 设置为密码模式
	rlTemp.Config.EnableMask = true
	
	// 读取密码
	password, err := rlTemp.Readline()
	if err != nil {
		if err == readline.ErrInterrupt {
			fmt.Println("^C")
			os.Exit(0)
		}
		fmt.Println("读取密码错误:", err)
		return ""
	}
	
	return strings.TrimSpace(password)
}

// handleInteractiveConnect 交互式连接向导
func handleInteractiveConnect() error {
	fmt.Println("请输入连接信息:")

	// 尝试加载默认配置
	defaultConfig, err := utils.LoadConfig()
	var dbType, host, user, password, dbName string
	var port int

	// 默认使用达梦数据库
	dbType = "dameng"
	port = 5236

	// 如果有默认配置，使用默认值，但允许用户修改
	if err == nil && defaultConfig != nil {
		dbType = defaultConfig.Type
		port = defaultConfig.Port
		
		// 提示用户是否使用默认配置
		fmt.Println("发现默认配置:")
		fmt.Printf("  数据库类型: %s\n", defaultConfig.Type)
		fmt.Printf("  主机地址: %s\n", defaultConfig.Host)
		fmt.Printf("  端口: %d\n", defaultConfig.Port)
		fmt.Printf("  用户名: %s\n", defaultConfig.User)
		fmt.Printf("  数据库名: %s\n", defaultConfig.DbName)
		
		// 读取用户选择
		useDefault := readInput("是否使用默认配置? (y/n): ")
		
		// 如果用户选择使用默认配置
		if strings.HasPrefix(strings.ToLower(useDefault), "y") {
			fmt.Println("使用默认配置连接...")
			return db.Connect(defaultConfig.Type, defaultConfig.Host, defaultConfig.Port, 
				defaultConfig.User, defaultConfig.Password, defaultConfig.DbName)
		}
		
		// 否则让用户输入新配置
		fmt.Println("请输入新的连接信息 (直接回车使用默认值):")
	}

	// 获取数据库类型
	fmt.Println("支持的数据库类型: dameng, mysql, postgresql, sqlite, oracle, mssql")
	var dbTypePrompt string
	if defaultConfig != nil {
		dbTypePrompt = fmt.Sprintf("数据库类型 (默认 %s): ", defaultConfig.Type)
	} else {
		dbTypePrompt = fmt.Sprintf("数据库类型 (默认 %s): ", dbType)
	}
	
	dbTypeInput := readInput(dbTypePrompt)
	
	// 验证数据库类型
	if dbTypeInput != "" {
		dbTypeInput = strings.ToLower(dbTypeInput)
		supportedDbTypes := []string{"dameng", "mysql", "postgresql", "sqlite", "oracle", "mssql"}
		isValidType := false
		for _, supportedType := range supportedDbTypes {
			if dbTypeInput == supportedType {
				isValidType = true
				break
			}
		}
		if isValidType {
			dbType = dbTypeInput
		} else {
			fmt.Printf("无效的数据库类型: %s, 将使用默认类型: %s\n", dbTypeInput, dbType)
		}
	} else if defaultConfig != nil {
		dbType = defaultConfig.Type
	}

	// 获取主机地址
	var hostPrompt string
	if defaultConfig != nil {
		hostPrompt = fmt.Sprintf("主机地址 (默认 %s): ", defaultConfig.Host)
	} else {
		hostPrompt = "主机地址: "
	}
	
	host = readInput(hostPrompt)
	if host == "" && defaultConfig != nil {
		host = defaultConfig.Host
	}
	
	// 获取端口
	var portPrompt string
	if defaultConfig != nil {
		portPrompt = fmt.Sprintf("端口 (默认 %d): ", defaultConfig.Port)
	} else {
		portPrompt = fmt.Sprintf("端口 (默认 %d): ", port)
	}
	
	portStr := readInput(portPrompt)
	if portStr != "" {
		fmt.Sscanf(portStr, "%d", &port)
	} else if defaultConfig != nil && defaultConfig.Port != 0 {
		port = defaultConfig.Port
	}
	
	// 获取用户名
	var userPrompt string
	if defaultConfig != nil {
		userPrompt = fmt.Sprintf("用户名 (默认 %s): ", defaultConfig.User)
	} else {
		userPrompt = "用户名: "
	}
	
	user = readInput(userPrompt)
	if user == "" && defaultConfig != nil {
		user = defaultConfig.User
	}
	
	// 获取密码
	password = readPassword("密码: ")
	if password == "" && defaultConfig != nil {
		password = defaultConfig.Password
	}
	
	// 获取数据库名
	var dbNamePrompt string
	if defaultConfig != nil {
		dbNamePrompt = fmt.Sprintf("数据库名 (默认 %s): ", defaultConfig.DbName)
	} else {
		dbNamePrompt = "数据库名: "
	}
	
	dbName = readInput(dbNamePrompt)
	if dbName == "" && defaultConfig != nil {
		dbName = defaultConfig.DbName
	}

	if host == "" || user == "" || password == "" || dbName == "" {
		return errors.New("连接参数不完整，请提供主机、用户名、密码和数据库名")
	}

	// 连接数据库
	return db.Connect(dbType, host, port, user, password, dbName)
}

// HandleInteractiveConnect 交互式连接向导（公开版本）
func HandleInteractiveConnect() error {
	return handleInteractiveConnect()
}

// HandleShowTables 显示表列表
func HandleShowTables() error {
	conn := db.GetCurrentConnection()
	if conn == nil {
		return errors.New("当前未连接到任何数据库")
	}

	tables, err := conn.GetTables()
	if err != nil {
		return err
	}

	if len(tables) == 0 {
		fmt.Println("数据库中没有找到表")
		return nil
	}

	fmt.Println("表列表:")
	for i, table := range tables {
		fmt.Printf("%3d) %s\n", i+1, table)
	}

	return nil
}

// HandleDescribeTable 显示表结构
func HandleDescribeTable(tableName string) error {
	conn := db.GetCurrentConnection()
	if conn == nil {
		return errors.New("当前未连接到任何数据库")
	}

	columns, err := conn.DescribeTable(tableName)
	if err != nil {
		return err
	}

	if len(columns) == 0 {
		return fmt.Errorf("表 %s 不存在或没有字段", tableName)
	}

	// 打印表头
	fmt.Printf("\n表 %s 的结构:\n", tableName)
	fmt.Printf("%-20s %-15s %-10s %-10s %-15s %-30s\n", "字段名", "数据类型", "长度", "可空", "约束", "描述")
	fmt.Println(strings.Repeat("-", 105))

	for _, col := range columns {
		// 处理不同数据库返回的列名大小写差异
		// 尝试先获取大写列名(达梦数据库风格)，如果不存在则尝试小写列名(PostgreSQL风格)
		var colName, dataType, dataLength, nullable, constraint, description string

		// 列名处理
		if val, ok := col["COLUMN_NAME"]; ok && val != nil {
			colName = fmt.Sprintf("%v", val)
		} else if val, ok := col["column_name"]; ok && val != nil {
			colName = fmt.Sprintf("%v", val)
		} else {
			colName = "<未知>"
		}

		// 数据类型处理
		if val, ok := col["DATA_TYPE"]; ok && val != nil {
			dataType = fmt.Sprintf("%v", val)
		} else if val, ok := col["data_type"]; ok && val != nil {
			dataType = fmt.Sprintf("%v", val)
		} else {
			dataType = "<未知>"
		}

		// 数据长度处理
		if val, ok := col["DATA_LENGTH"]; ok && val != nil {
			dataLength = fmt.Sprintf("%v", val)
		} else if val, ok := col["data_length"]; ok && val != nil {
			dataLength = fmt.Sprintf("%v", val)
		} else {
			dataLength = ""
		}

		// 可空处理
		if val, ok := col["NULLABLE"]; ok && val != nil {
			nullable = fmt.Sprintf("%v", val)
		} else if val, ok := col["is_nullable"]; ok && val != nil {
			nullable = fmt.Sprintf("%v", val)
		} else {
			nullable = "<未知>"
		}

		// 约束处理
		if val, ok := col["CONSTRAINT_TYPE"]; ok && val != nil {
			constraint = fmt.Sprintf("%v", val)
		} else if val, ok := col["constraint_type"]; ok && val != nil {
			constraint = fmt.Sprintf("%v", val)
		} else {
			constraint = ""
		}

		// 描述处理
		if val, ok := col["DESCRIPTION"]; ok && val != nil {
			description = fmt.Sprintf("%v", val)
		} else if val, ok := col["description"]; ok && val != nil {
			description = fmt.Sprintf("%v", val)
		} else {
			description = ""
		}

		fmt.Printf("%-20s %-15s %-10s %-10s %-15s %-30s\n",
			colName, dataType, dataLength, nullable, constraint, description)
	}
	fmt.Println()

	return nil
}

// 导入模式
const (
	ImportModeInsert = "insert" // 仅插入，遇到冲突会报错
	ImportModeUpsert = "upsert" // 更新插入，存在则更新，不存在则插入
)

// HandleImport 处理导入命令
func HandleImport(cmdStr string) error {
	// 解析命令
	// IMPORT <table> FROM <file> [FORMAT csv/excel] [MODE insert/upsert]
	parts := strings.Fields(cmdStr)
	if len(parts) < 4 || strings.ToUpper(parts[2]) != "FROM" {
		return errors.New("用法: IMPORT <表名> FROM <文件路径> [FORMAT csv/excel] [MODE insert/upsert]")
	}

	// 获取表名和文件路径
	tableName := parts[1]
	filePath := parts[3]
	
	// 默认格式为CSV，默认模式为INSERT
	format := FormatCSV
	importMode := ImportModeInsert
	
	// 解析其他参数
	for i := 4; i < len(parts); i++ {
		switch strings.ToUpper(parts[i]) {
		case "FORMAT":
			if i+1 < len(parts) {
				specifiedFormat := strings.ToLower(parts[i+1])
				if specifiedFormat != FormatCSV && specifiedFormat != FormatExcel {
					return fmt.Errorf("不支持的导入格式: %s，支持的格式为: csv, excel", specifiedFormat)
				}
				format = specifiedFormat
				i++
			}
		case "MODE":
			if i+1 < len(parts) {
				specifiedMode := strings.ToLower(parts[i+1])
				if specifiedMode != ImportModeInsert && specifiedMode != ImportModeUpsert {
					return fmt.Errorf("不支持的导入模式: %s，支持的模式为: insert, upsert", specifiedMode)
				}
				importMode = specifiedMode
				i++
			}
		}
	}
	
	// 根据文件扩展名判断格式（如果未明确指定）
	if strings.HasSuffix(strings.ToLower(filePath), ".xlsx") {
		format = FormatExcel
	} else if strings.HasSuffix(strings.ToLower(filePath), ".csv") {
		format = FormatCSV
	}
	
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return fmt.Errorf("文件不存在: %s", filePath)
	}
	
	// 获取数据库连接
	conn := db.GetCurrentConnection()
	if conn == nil {
		return errors.New("当前未连接到任何数据库")
	}
	
	// 获取表结构信息
	tableInfo, err := conn.DescribeTable(tableName)
	if err != nil {
		return fmt.Errorf("获取表结构失败: %v", err)
	}
	
	// 创建表字段映射
	fieldMap := make(map[string]string) // 列名/描述 -> 字段名
	dbFieldsInfo := make(map[string]map[string]interface{}) // 字段名 -> 字段信息
	primaryKeyField := "" // 记录主键字段名
	isIdentityField := make(map[string]bool) // 记录自增字段
	
	for _, col := range tableInfo {
		// 处理不同数据库返回的列名大小写差异
		var colName, description, constraint, identityInfo string
		
		// 获取列名
		if val, ok := col["COLUMN_NAME"]; ok && val != nil {
			colName = fmt.Sprintf("%v", val)
		} else if val, ok := col["column_name"]; ok && val != nil {
			colName = fmt.Sprintf("%v", val)
		} else {
			continue // 跳过无效列
		}
		
		colNameUpper := strings.ToUpper(colName)
		
		// 获取描述
		if val, ok := col["DESCRIPTION"]; ok && val != nil {
			description = fmt.Sprintf("%v", val)
		} else if val, ok := col["description"]; ok && val != nil {
			description = fmt.Sprintf("%v", val)
		} else {
			description = ""
		}
		
		// 获取约束
		if val, ok := col["CONSTRAINT_TYPE"]; ok && val != nil {
			constraint = fmt.Sprintf("%v", val)
		} else if val, ok := col["constraint_type"]; ok && val != nil {
			constraint = fmt.Sprintf("%v", val)
		} else {
			constraint = ""
		}
		
		// 获取自增信息
		if val, ok := col["IDENTITY_INFO"]; ok && val != nil {
			identityInfo = fmt.Sprintf("%v", val)
		} else if val, ok := col["identity_info"]; ok && val != nil {
			identityInfo = fmt.Sprintf("%v", val)
		} else {
			identityInfo = ""
		}
		
		// 使用字段名和描述作为可能的映射键
		fieldMap[colNameUpper] = colName 
		if description != "" && description != "<nil>" {
			fieldMap[strings.ToUpper(description)] = colName
		}
		
		// 记录主键字段
		if strings.Contains(strings.ToUpper(constraint), "PRIMARY KEY") {
			primaryKeyField = colName
		}
		
		// 检查是否为自增字段 (IDENTITY)
		if identityInfo != "" && identityInfo != "<nil>" && strings.Contains(strings.ToUpper(identityInfo), "IDENTITY") {
			isIdentityField[colNameUpper] = true
		}
		
		// 存储字段信息，以便后续处理
		dbFieldsInfo[colNameUpper] = col
	}
	
	// 如果是upsert模式，需要确保存在主键
	if importMode == ImportModeUpsert && primaryKeyField == "" {
		return errors.New("更新插入模式(upsert)需要表有主键字段，但未找到主键")
	}

	// 根据格式读取数据
	var records [][]string
	var headers []string
	
	switch format {
	case FormatCSV:
		records, headers, err = readCSV(filePath)
	case FormatExcel:
		records, headers, err = readExcel(filePath)
	}
	
	if err != nil {
		return err
	}
	
	if len(records) == 0 {
		return errors.New("文件中没有数据")
	}
	
	// 列映射: 文件列索引 -> 数据库字段名
	columnMapping := make(map[int]string)
	unmappedColumns := []string{}
	
	// 确认主键字段是否在文件中
	hasPrimaryKeyInFile := false

	// 尝试将文件列头映射到数据库字段
	for i, header := range headers {
		headerUpper := strings.ToUpper(strings.TrimSpace(header))
		
		// 尝试直接匹配
		if dbFieldName, ok := fieldMap[headerUpper]; ok {
			columnMapping[i] = dbFieldName
			// 检查是否有主键字段
			if dbFieldName == primaryKeyField {
				hasPrimaryKeyInFile = true
			}
			continue
		}
		
		// 如果不匹配，记录未映射的列
		unmappedColumns = append(unmappedColumns, header)
	}

	// 如果是upsert模式但文件中没有主键列，则报错
	if importMode == ImportModeUpsert && !hasPrimaryKeyInFile {
		return fmt.Errorf("更新插入模式(upsert)需要文件中包含主键列(%s)，但未找到", primaryKeyField)
	}

	// 开始导入数据
	successCount := 0
	updateCount := 0
	insertCount := 0
	errorCount := 0
	
	for i, record := range records {
		// 准备字段和值
		var columns []string
		var placeholders []string
		var values []interface{}
		var pkValue interface{} = nil
		var pkColumn string = ""
		
		// 判断是否存在ID字段
		idColumnIndex := -1
		for colIndex, dbFieldName := range columnMapping {
			if strings.EqualFold(dbFieldName, "id") || strings.EqualFold(dbFieldName, primaryKeyField) {
				idColumnIndex = colIndex
				pkColumn = dbFieldName
				break
			}
		}
		
		// 简化处理：如果存在ID字段，则使用其值判断是更新还是插入，而不直接插入ID值
		isUpdate := false
		if idColumnIndex >= 0 && idColumnIndex < len(record) {
			idValue := strings.TrimSpace(record[idColumnIndex])
			if idValue != "" {
				// 有ID值，记录为更新
				if idVal, err := strconv.ParseInt(idValue, 10, 64); err == nil {
					pkValue = idVal
					isUpdate = true
				} else if floatVal, err := strconv.ParseFloat(idValue, 64); err == nil {
					pkValue = floatVal
					isUpdate = true
				}
			}
		}
		
		for colIndex, dbFieldName := range columnMapping {
			if colIndex < len(record) {
				// 跳过ID字段，不直接插入
				if isUpdate && (strings.EqualFold(dbFieldName, "id") || strings.EqualFold(dbFieldName, primaryKeyField)) {
					continue
				}
				
				// 获取并转换值
				value := strings.TrimSpace(record[colIndex])
				dbFieldNameUpper := strings.ToUpper(dbFieldName)
				
				// 对于自增字段，仅当有值时才使用它
				if isIdentityField[dbFieldNameUpper] && value == "" {
					continue
				}
				
				if value != "" { // 忽略空值
					columns = append(columns, dbFieldName)
					placeholders = append(placeholders, "?")
					
					// 根据字段类型转换值
					fieldInfoUpper := dbFieldsInfo[dbFieldNameUpper]
					dataType := fmt.Sprintf("%v", fieldInfoUpper["DATA_TYPE"])
					dataType = strings.ToUpper(dataType)
					
					var convertedValue interface{}
					
					switch {
					case strings.Contains(dataType, "INT") || strings.Contains(dataType, "NUMBER"):
						// 尝试转换为数字
						if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
							convertedValue = intVal
						} else if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
							convertedValue = floatVal
						} else {
							// 无法解析为数字，跳过此列
							columns = columns[:len(columns)-1]
							placeholders = placeholders[:len(placeholders)-1]
							continue
						}
					case strings.Contains(dataType, "DATE") || strings.Contains(dataType, "TIME"):
						// 处理日期时间格式
						// 达梦数据库支持的标准格式是 YYYY-MM-DD HH24:MI:SS 或 YYYY-MM-DD
						dateTimeStr := value
						
						// 检测并处理常见的日期时间格式错误
						if strings.Contains(dateTimeStr, "+0800") {
							// 处理带时区信息的日期时间，比如: 2025-04-28 15:00:13.727014 +0800 +0800
							parts := strings.Split(dateTimeStr, " ")
							if len(parts) >= 2 {
								// 保留日期和时间部分，去掉时区信息
								dateTimeStr = parts[0] + " " + parts[1]
								
								// 处理可能的毫秒部分
								if strings.Contains(dateTimeStr, ".") {
									timeParts := strings.Split(dateTimeStr, ".")
									if len(timeParts) > 1 {
										// 只保留到秒
										dateTimeStr = timeParts[0]
									}
								}
							}
						} else if strings.Contains(dateTimeStr, "T") && strings.Contains(dateTimeStr, "Z") {
							// 处理ISO格式: 2025-04-28T15:00:13Z
							dateTimeStr = strings.Replace(dateTimeStr, "T", " ", 1)
							dateTimeStr = strings.Replace(dateTimeStr, "Z", "", 1)
						} else if len(dateTimeStr) == 10 && strings.Count(dateTimeStr, "-") == 2 {
							// 只有日期部分，保持不变
						} else if len(dateTimeStr) > 19 {
							// 如果字符串长度超过标准的日期时间格式(YYYY-MM-DD HH:MM:SS)，截取前19个字符
							dateTimeStr = dateTimeStr[:19]
						}
						
						// 检查最终格式是否符合标准
						isDateOnly := len(dateTimeStr) == 10 && strings.Count(dateTimeStr, "-") == 2
						isDateTime := len(dateTimeStr) == 19 && strings.Count(dateTimeStr, "-") == 2 && strings.Count(dateTimeStr, ":") == 2
						
						if !isDateOnly && !isDateTime {
							// 格式仍然不符合，尝试解析为更通用的格式
							t, err := time.Parse(time.RFC3339, value)
							if err != nil {
								t, err = time.Parse(time.RFC1123, value)
							}
							if err != nil {
								t, err = time.Parse("2006-01-02 15:04:05.999999999 -0700 MST", value)
							}
							if err != nil {
								// 无法解析，使用当前时间
								fmt.Printf("警告: 第 %d 行日期时间格式不正确: %s，将被忽略\n", i+1, value)
								columns = columns[:len(columns)-1]
								placeholders = placeholders[:len(placeholders)-1]
								continue
							}
							// 格式化为达梦数据库接受的格式
							dateTimeStr = t.Format("2006-01-02 15:04:05")
						}
						
						convertedValue = dateTimeStr
					default:
						// 字符串或其他类型
						convertedValue = value
					}
					
					values = append(values, convertedValue)
					
					// 如果是主键字段，记录主键值
					if dbFieldName == primaryKeyField {
						pkValue = convertedValue
					}
				}
			}
		}
		
		// 如果没有有效列，跳过此行
		if len(columns) == 0 {
			fmt.Printf("警告: 第 %d 行没有有效数据，已跳过\n", i+1)
			continue
		}
		
		var err error
		
		if isUpdate && pkValue != nil {
			// 执行更新操作
			updateSql := fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?",
				tableName,
				strings.Join(columns, " = ?, ") + " = ?",
				pkColumn)
			
			// 添加所有参数，最后一个是WHERE条件
			updateValues := append(values, pkValue)
			
			_, err = conn.ExecuteWithParams(updateSql, updateValues...)
			if err == nil {
				updateCount++
				successCount++
			} else {
				errorCount++
				fmt.Printf("错误: 第 %d 行更新失败: %v\n", i+1, err)
			}
		} else {
			// 执行插入操作 - 不包含ID字段
			insertSql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s)",
				tableName,
				strings.Join(columns, ", "),
				strings.Join(placeholders, ", "))
			
			_, err = conn.ExecuteWithParams(insertSql, values...)
			if err == nil {
				insertCount++
				successCount++
			} else {
				errorCount++
				fmt.Printf("错误: 第 %d 行插入失败: %v\n", i+1, err)
			}
		}
		
		if err != nil {
			fmt.Printf("错误: 第 %d 行操作失败: %v\n", i+1, err)
		}
	}

	fmt.Printf("成功导入 %d 条记录到 %s\n", successCount, filePath)
	return nil
}

// readCSV 读取CSV文件
func readCSV(filePath string) ([][]string, []string, error) {
	// 打开文件
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	
	// 创建带缓冲的读取器，处理BOM
	reader := bufio.NewReader(file)
	
	// 检测并跳过BOM
	bom := make([]byte, 3)
	_, _ = reader.Read(bom)
	if bom[0] != 0xEF || bom[1] != 0xBB || bom[2] != 0xBF {
		// 不是BOM，将文件指针重置
		file.Seek(0, 0)
		reader = bufio.NewReader(file)
	}
	
	// 创建CSV读取器
	csvReader := csv.NewReader(reader)
	
	// 读取所有记录
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, nil, err
	}
	
	if len(records) < 1 {
		return nil, nil, errors.New("CSV文件为空")
	}
	
	// 提取表头和数据
	headers := records[0]
	data := records[1:]
	
	return data, headers, nil
}

// readExcel 读取Excel文件
func readExcel(filePath string) ([][]string, []string, error) {
	// 打开Excel文件
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	
	// 获取第一个工作表名称
	sheets := f.GetSheetList()
	if len(sheets) == 0 {
		return nil, nil, errors.New("Excel文件不包含工作表")
	}
	sheetName := sheets[0]
	
	// 读取所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		return nil, nil, err
	}
	
	if len(rows) < 1 {
		return nil, nil, errors.New("Excel工作表为空")
	}
	
	// 提取表头和数据
	headers := rows[0]
	data := rows[1:]
	
	return data, headers, nil
}

// 导出格式
const (
	FormatCSV   = "csv"
	FormatExcel = "excel"
)

// HandleExport 处理导出命令
func HandleExport(cmdStr string) error {
	// 解析命令
	// EXPORT <table> [WHERE 条件] <file> [FORMAT csv/excel]
	parts := strings.Fields(cmdStr)
	if len(parts) < 3 {
		return errors.New("用法: EXPORT <表名> [WHERE 条件] <文件名> [FORMAT csv/excel]")
	}

	// 获取表名
	tableName := parts[1]
	
	var whereClause string
	var filePath string
	var format string = FormatCSV // 默认为CSV
	
	// 解析命令参数
	i := 2
	for i < len(parts) {
		switch strings.ToUpper(parts[i]) {
		case "WHERE":
			// 寻找WHERE子句的结束位置
			start := i + 1
			end := start
			for end < len(parts) && strings.ToUpper(parts[end]) != "FORMAT" && !strings.HasSuffix(parts[end], ".csv") && !strings.HasSuffix(parts[end], ".xlsx") {
				end++
			}
			if start < end {
				whereClause = strings.Join(parts[start:end], " ")
			}
			i = end
		case "FORMAT":
			if i+1 < len(parts) {
				format = strings.ToLower(parts[i+1])
				i += 2
			} else {
				i++
			}
		default:
			// 检查是否是文件路径
			if strings.HasSuffix(strings.ToLower(parts[i]), ".csv") || strings.HasSuffix(strings.ToLower(parts[i]), ".xlsx") {
				filePath = parts[i]
				// 根据文件扩展名判断格式
				if strings.HasSuffix(strings.ToLower(parts[i]), ".xlsx") {
					format = FormatExcel
				} else {
					format = FormatCSV
				}
			}
			i++
		}
	}
	
	if filePath == "" {
		return errors.New("必须指定导出文件路径")
	}
	
	// 验证格式
	if format != FormatCSV && format != FormatExcel {
		return fmt.Errorf("不支持的导出格式: %s，支持的格式为: csv, excel", format)
	}
	
	// 先获取表结构信息
	conn := db.GetCurrentConnection()
	if conn == nil {
		return errors.New("当前未连接到任何数据库")
	}
	
	tableInfo, err := conn.DescribeTable(tableName)
	if err != nil {
		return fmt.Errorf("获取表结构失败: %v", err)
	}
	
	// 创建字段名到描述的映射
	fieldDescriptions := make(map[string]string)
	for _, col := range tableInfo {
		// 处理不同数据库返回的列名大小写差异
		var colName, description string
		
		// 获取列名
		if val, ok := col["COLUMN_NAME"]; ok && val != nil {
			colName = fmt.Sprintf("%v", val)
		} else if val, ok := col["column_name"]; ok && val != nil {
			colName = fmt.Sprintf("%v", val)
		} else {
			continue // 跳过无效列
		}
		
		// 获取描述
		if val, ok := col["DESCRIPTION"]; ok && val != nil {
			description = fmt.Sprintf("%v", val)
		} else if val, ok := col["description"]; ok && val != nil {
			description = fmt.Sprintf("%v", val)
		} else {
			description = ""
		}
		
		fieldDescriptions[strings.ToUpper(colName)] = description
	}
	
	// 获取表的列顺序
	orderedColumns, err := conn.GetTableColumns(tableName)
	if err != nil {
		// 如果获取列顺序失败，记录错误但不中断执行
		fmt.Printf("警告: 无法获取表的列顺序: %v，将使用默认排序\n", err)
		orderedColumns = nil
	}
	
	// 构建查询SQL
	sql := fmt.Sprintf("SELECT * FROM %s", tableName)
	if whereClause != "" {
		sql += " WHERE " + whereClause
	}
	
	// 执行查询
	results, err := conn.Query(sql)
	if err != nil {
		return err
	}
	
	if len(results) == 0 {
		return errors.New("没有找到符合条件的数据")
	}
	
	// 导出数据
	switch format {
	case FormatCSV:
		err = exportToCSV(results, filePath, fieldDescriptions, orderedColumns)
	case FormatExcel:
		err = exportToExcel(results, filePath, fieldDescriptions, orderedColumns)
	}
	
	if err != nil {
		return err
	}
	
	fmt.Printf("成功导出 %d 条记录到 %s\n", len(results), filePath)
	return nil
}

// exportToCSV 导出数据到CSV文件
func exportToCSV(data []map[string]interface{}, filePath string, fieldDescriptions map[string]string, orderedColumns []string) error {
	if len(data) == 0 {
		return errors.New("没有数据可导出")
	}
	
	// 创建文件
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	
	// 写入UTF-8 BOM，解决中文乱码问题
	_, err = file.WriteString("\xEF\xBB\xBF")
	if err != nil {
		return err
	}
	
	// 创建CSV写入器
	writer := csv.NewWriter(file)
	defer writer.Flush()
	
	// 获取列顺序
	var headers []string
	if orderedColumns != nil {
		// 使用表的原始顺序
		for _, col := range orderedColumns {
			// 检查数据中是否包含此列
			if _, exists := data[0][col]; exists {
				headers = append(headers, col)
			}
		}
		
		// 处理数据中可能存在但未在获取的列顺序中的列
		for col := range data[0] {
			found := false
			for _, orderedCol := range headers {
				if strings.EqualFold(orderedCol, col) {
					found = true
					break
				}
			}
			if !found {
				headers = append(headers, col)
			}
		}
	} else {
		// 没有获取到表的列顺序，使用map的键并排序
		for col := range data[0] {
			headers = append(headers, col)
		}
		sort.Strings(headers) // 确保列顺序一致
	}
	
	// 准备显示标题 - 优先使用描述，如果没有描述则使用字段名
	var displayHeaders []string
	for _, header := range headers {
		// 尝试获取字段描述
		description := fieldDescriptions[strings.ToUpper(header)]
		if description != "" && description != "<nil>" {
			displayHeaders = append(displayHeaders, description)
		} else {
			displayHeaders = append(displayHeaders, header)
		}
	}
	
	// 写入表头
	if err := writer.Write(displayHeaders); err != nil {
		return err
	}
	
	// 写入数据行
	for _, record := range data {
		var row []string
		for _, header := range headers {
			// 将各种类型转换为字符串
			var cell string
			if record[header] == nil {
				cell = ""
			} else {
				cell = fmt.Sprintf("%v", record[header])
				// 处理日期时间格式，只保留到秒
				if strings.Contains(cell, "-") && strings.Contains(cell, ":") {
					// 判断是否是日期时间格式
					if strings.Contains(cell, "+") || strings.Contains(cell, ".") {
						parts := strings.Split(cell, " ")
						if len(parts) >= 2 {
							// 提取日期和时间部分，截断到秒
							timeStr := parts[1]
							if strings.Contains(timeStr, ".") {
								timeStr = strings.Split(timeStr, ".")[0]
							}
							// 重建日期时间字符串，不包含时区和毫秒部分
							cell = parts[0] + " " + timeStr
						}
					}
				}
			}
			row = append(row, cell)
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	
	return nil
}

// exportToExcel 导出数据到Excel文件
func exportToExcel(data []map[string]interface{}, filePath string, fieldDescriptions map[string]string, orderedColumns []string) error {
	if len(data) == 0 {
		return errors.New("没有数据可导出")
	}
	
	// 创建一个新的Excel文件
	f := excelize.NewFile()
	
	// 默认的工作表名称
	sheetName := "Sheet1"
	
	// 获取列顺序
	var headers []string
	if orderedColumns != nil {
		// 使用表的原始顺序
		for _, col := range orderedColumns {
			// 检查数据中是否包含此列
			if _, exists := data[0][col]; exists {
				headers = append(headers, col)
			}
		}
		
		// 处理数据中可能存在但未在获取的列顺序中的列
		for col := range data[0] {
			found := false
			for _, orderedCol := range headers {
				if strings.EqualFold(orderedCol, col) {
					found = true
					break
				}
			}
			if !found {
				headers = append(headers, col)
			}
		}
	} else {
		// 没有获取到表的列顺序，使用map的键并排序
		for col := range data[0] {
			headers = append(headers, col)
		}
		sort.Strings(headers) // 确保列顺序一致
	}
	
	// 准备显示标题 - 优先使用描述，如果没有描述则使用字段名
	var displayHeaders []string
	for _, header := range headers {
		// 尝试获取字段描述
		description := fieldDescriptions[strings.ToUpper(header)]
		if description != "" && description != "<nil>" {
			displayHeaders = append(displayHeaders, description)
		} else {
			displayHeaders = append(displayHeaders, header)
		}
	}
	
	// 写入表头
	for i, displayHeader := range displayHeaders {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheetName, cell, displayHeader)
	}
	
	// 写入数据行
	for rowIdx, record := range data {
		for colIdx, header := range headers {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, rowIdx+2)
			
			// 获取值并处理时间格式
			var value interface{} = record[header]
			if value != nil {
				strValue := fmt.Sprintf("%v", value)
				// 处理日期时间格式，只保留到秒
				if strings.Contains(strValue, "-") && strings.Contains(strValue, ":") {
					// 判断是否是日期时间格式
					if strings.Contains(strValue, "+") || strings.Contains(strValue, ".") {
						parts := strings.Split(strValue, " ")
						if len(parts) >= 2 {
							// 提取日期和时间部分，截断到秒
							timeStr := parts[1]
							if strings.Contains(timeStr, ".") {
								timeStr = strings.Split(timeStr, ".")[0]
							}
							// 重建日期时间字符串，不包含时区和毫秒部分
							strValue = parts[0] + " " + timeStr
							value = strValue
						}
					}
				}
			}
			
			f.SetCellValue(sheetName, cell, value)
		}
	}
	
	// 保存文件
	if err := f.SaveAs(filePath); err != nil {
		return err
	}
	
	// 关闭文件
	if err := f.Close(); err != nil {
		return err
	}
	
	return nil
}

// interfaceSlice 将字符串切片转换为接口切片
func interfaceSlice(slice []string) []interface{} {
	iSlice := make([]interface{}, len(slice))
	for i, v := range slice {
		iSlice[i] = v // 不再使用颜色包装，避免显示问题
	}
	return iSlice
}

// Close 关闭readline实例
func Close() {
	if rl != nil {
		rl.Close()
		rl = nil
	}
} 