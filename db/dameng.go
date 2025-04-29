package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "gitee.com/chunanyong/dm"
)

// DamengConnection 达梦数据库连接
type DamengConnection struct {
	config *DbConfig
	db     *sql.DB
}

// NewDamengConnection 创建达梦数据库连接
func NewDamengConnection(config *DbConfig) (*DamengConnection, error) {
	return &DamengConnection{
		config: config,
	}, nil
}

// Connect 连接到达梦数据库
func (d *DamengConnection) Connect() error {
	connectionString := fmt.Sprintf("dm://%s:%s@%s:%d/%s?charset=utf8",
		d.config.User, d.config.Password, d.config.Host, d.config.Port, d.config.DbName)

	db, err := sql.Open("dm", connectionString)
	if err != nil {
		return err
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		return err
	}

	d.db = db
	return nil
}

// Disconnect 断开连接
func (d *DamengConnection) Disconnect() error {
	if d.db != nil {
		return d.db.Close()
	}
	return nil
}

// formatDateTime 格式化时间为标准格式，只保留到秒
func formatDateTime(val interface{}) interface{} {
	if val == nil {
		return val
	}
	
	// 如果是字符串，处理成标准格式
	if strVal, ok := val.(string); ok && (strings.Contains(strVal, "-") || strings.Contains(strVal, ":")) {
		// 处理带时区的格式
		if strings.Contains(strVal, "+") {
			parts := strings.Split(strVal, " ")
			if len(parts) >= 2 {
				strVal = parts[0] + " " + parts[1]
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
func (d *DamengConnection) Query(query string) ([]map[string]interface{}, error) {
	if d.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	rows, err := d.db.Query(query)
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
				row[col] = formatDateTime(val)
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
func (d *DamengConnection) QueryWithParams(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if d.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	rows, err := d.db.Query(query, args...)
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
				row[col] = formatDateTime(val)
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
func (d *DamengConnection) Execute(query string) (int64, error) {
	if d.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result, err := d.db.Exec(query)
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
func (d *DamengConnection) ExecuteWithParams(query string, args ...interface{}) (int64, error) {
	if d.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result, err := d.db.Exec(query, args...)
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
func (d *DamengConnection) GetTables() ([]string, error) {
	if d.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	query := `SELECT TABLE_NAME FROM DBA_TABLES WHERE OWNER = UPPER(?)`
	rows, err := d.db.Query(query, d.config.User)
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
func (d *DamengConnection) DescribeTable(tableName string) ([]map[string]interface{}, error) {
	if d.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 查询表结构和字段注释
	query := `
		SELECT 
			C.COLUMN_NAME,
			C.DATA_TYPE,
			C.DATA_LENGTH,
			C.NULLABLE,
			DECODE(C.COLUMN_NAME, 
				(SELECT CC.COLUMN_NAME FROM USER_CONS_COLUMNS CC 
				 JOIN USER_CONSTRAINTS uc ON CC.CONSTRAINT_NAME = uc.CONSTRAINT_NAME 
				 WHERE uc.TABLE_NAME = ? AND uc.CONSTRAINT_TYPE = 'P' AND CC.COLUMN_NAME = C.COLUMN_NAME 
				 AND ROWNUM = 1), 'PRIMARY KEY', '') AS CONSTRAINT_TYPE,
			NVL((SELECT COMMENTS FROM USER_COL_COMMENTS WHERE TABLE_NAME = ? AND COLUMN_NAME = C.COLUMN_NAME), '') AS DESCRIPTION,
			CASE WHEN C.DATA_TYPE LIKE '%IDENTITY%' THEN 'IDENTITY' ELSE '' END AS IDENTITY_INFO
		FROM 
			USER_TAB_COLUMNS C
		WHERE 
			C.TABLE_NAME = ?
		ORDER BY 
			C.COLUMN_ID
	`

	rows, err := d.db.Query(query, strings.ToUpper(tableName), strings.ToUpper(tableName), strings.ToUpper(tableName))
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
				row[col] = val
			}
		}
		results = append(results, row)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

// GetTableColumns 获取表字段列表，按字段在表中的顺序排序
func (d *DamengConnection) GetTableColumns(tableName string) ([]string, error) {
	if d.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 查询表结构，按列ID排序
	query := `
		SELECT 
			C.COLUMN_NAME
		FROM 
			USER_TAB_COLUMNS C
		WHERE 
			C.TABLE_NAME = ?
		ORDER BY 
			C.COLUMN_ID
	`

	rows, err := d.db.Query(query, strings.ToUpper(tableName))
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