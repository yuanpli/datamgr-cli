package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// MySQLConnection MySQL数据库连接
type MySQLConnection struct {
	config *DbConfig
	db     *sql.DB
}

// NewMySQLConnection 创建MySQL数据库连接
func NewMySQLConnection(config *DbConfig) (*MySQLConnection, error) {
	return &MySQLConnection{
		config: config,
	}, nil
}

// Connect 连接到MySQL数据库
func (m *MySQLConnection) Connect() error {
	connectionString := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s",
		m.config.User, m.config.Password, m.config.Host, m.config.Port, m.config.DbName)

	db, err := sql.Open("mysql", connectionString)
	if err != nil {
		return err
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		return err
	}

	m.db = db
	return nil
}

// Disconnect 断开连接
func (m *MySQLConnection) Disconnect() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// formatMySQLTime 格式化MySQL时间为标准格式
func formatMySQLTime(val interface{}) interface{} {
	if val == nil {
		return val
	}
	
	// 如果是字符串，处理成标准格式
	if strVal, ok := val.(string); ok && (strings.Contains(strVal, "-") || strings.Contains(strVal, ":")) {
		// 处理带时区的格式
		if strings.Contains(strVal, "+") || strings.Contains(strVal, "Z") {
			parts := strings.Fields(strVal)
			if len(parts) >= 1 {
				strVal = parts[0]
				if len(parts) >= 2 {
					strVal += " " + parts[1]
				}
				if strings.Contains(strVal, ".") {
					timeParts := strings.Split(strVal, ".")
					strVal = timeParts[0]  // 只保留到秒
				}
			}
		}
		return strVal
	}
	
	// 如果是time.Time类型，格式化为标准格式
	if timeVal, ok := val.(time.Time); ok {
		return timeVal.Format("2006-01-02 15:04:05")
	}
	
	return val
}

// Query 执行查询语句
func (m *MySQLConnection) Query(query string) ([]map[string]interface{}, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				// 格式化时间类型值
				row[col] = formatMySQLTime(val)
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// QueryWithParams 执行带参数的查询语句
func (m *MySQLConnection) QueryWithParams(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	rows, err := m.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		for i, col := range columns {
			val := values[i]
			if b, ok := val.([]byte); ok {
				row[col] = string(b)
			} else {
				// 格式化时间类型值
				row[col] = formatMySQLTime(val)
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// Execute 执行更新/插入/删除语句
func (m *MySQLConnection) Execute(query string) (int64, error) {
	if m.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result, err := m.db.Exec(query)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

// ExecuteWithParams 执行带参数的更新/插入/删除语句
func (m *MySQLConnection) ExecuteWithParams(query string, args ...interface{}) (int64, error) {
	if m.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result, err := m.db.Exec(query, args...)
	if err != nil {
		return 0, err
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return affected, nil
}

// GetTables 获取所有表
func (m *MySQLConnection) GetTables() ([]string, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	query := `SHOW TABLES`
	rows, err := m.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			return nil, err
		}
		tables = append(tables, tableName)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return tables, nil
}

// DescribeTable 获取表结构
func (m *MySQLConnection) DescribeTable(tableName string) ([]map[string]interface{}, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 查询表结构
	descQuery := fmt.Sprintf("DESCRIBE %s", tableName)
	rows, err := m.db.Query(descQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// 获取表的主键信息
	pkQuery := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
		AND CONSTRAINT_NAME = 'PRIMARY'
	`
	pkRows, err := m.db.Query(pkQuery, m.config.DbName, tableName)
	if err != nil {
		return nil, err
	}
	defer pkRows.Close()

	// 存储主键列名
	primaryKeys := make(map[string]bool)
	for pkRows.Next() {
		var pkColumn string
		if err := pkRows.Scan(&pkColumn); err != nil {
			return nil, err
		}
		primaryKeys[pkColumn] = true
	}

	// 获取列注释信息
	commentQuery := `
		SELECT COLUMN_NAME, COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
	`
	commentRows, err := m.db.Query(commentQuery, m.config.DbName, tableName)
	if err != nil {
		return nil, err
	}
	defer commentRows.Close()

	// 存储列注释
	comments := make(map[string]string)
	for commentRows.Next() {
		var colName, comment string
		if err := commentRows.Scan(&colName, &comment); err != nil {
			return nil, err
		}
		comments[colName] = comment
	}

	// 处理DESCRIBE结果
	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]interface{}, 0)
	values := make([]interface{}, len(columns))
	scanArgs := make([]interface{}, len(columns))

	for i := range values {
		scanArgs[i] = &values[i]
	}

	for rows.Next() {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}

		row := make(map[string]interface{})
		
		// 获取列名 (DESCRIBE结果的第一列)
		var columnName string
		if b, ok := values[0].([]byte); ok {
			columnName = string(b)
		} else if s, ok := values[0].(string); ok {
			columnName = s
		} else {
			columnName = fmt.Sprintf("%v", values[0])
		}
		
		// 统一转换为我们需要的列名格式
		row["column_name"] = columnName
		
		// 获取数据类型 (DESCRIBE结果的第二列)
		var dataType string
		if b, ok := values[1].([]byte); ok {
			dataType = string(b)
		} else if s, ok := values[1].(string); ok {
			dataType = s
		} else {
			dataType = fmt.Sprintf("%v", values[1])
		}
		row["data_type"] = dataType
		
		// 数据长度 (从类型中提取)
		dataLength := ""
		if strings.Contains(dataType, "(") && strings.Contains(dataType, ")") {
			start := strings.Index(dataType, "(")
			end := strings.Index(dataType, ")")
			if start > 0 && end > start {
				dataLength = dataType[start+1 : end]
				dataType = dataType[:start] // 更新数据类型为不带长度的类型名
				row["data_type"] = dataType
			}
		}
		row["data_length"] = dataLength
		
		// 获取可空信息 (DESCRIBE结果的第三列)
		var nullable string
		if b, ok := values[2].([]byte); ok {
			nullable = string(b)
		} else if s, ok := values[2].(string); ok {
			nullable = s
		} else {
			nullable = fmt.Sprintf("%v", values[2])
		}
		row["is_nullable"] = nullable
		
		// 设置主键约束
		constraintType := ""
		if primaryKeys[columnName] {
			constraintType = "PRIMARY KEY"
		}
		row["constraint_type"] = constraintType
		
		// 设置注释
		description := comments[columnName]
		row["description"] = description
		
		// 检查是否是自增列
		var extra string
		if len(values) >= 6 {
			if b, ok := values[5].([]byte); ok {
				extra = string(b)
			} else if s, ok := values[5].(string); ok {
				extra = s
			} else {
				extra = fmt.Sprintf("%v", values[5])
			}
		}
		identityInfo := ""
		if strings.Contains(strings.ToLower(extra), "auto_increment") {
			identityInfo = "IDENTITY"
		}
		row["identity_info"] = identityInfo
		
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetTableColumns 获取表字段列表，按字段在表中的顺序排序
func (m *MySQLConnection) GetTableColumns(tableName string) ([]string, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 查询表结构，按列顺序获取字段名
	query := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ?
		AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := m.db.Query(query, m.config.DbName, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var columnName string
		if err := rows.Scan(&columnName); err != nil {
			return nil, err
		}
		columns = append(columns, columnName)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
} 