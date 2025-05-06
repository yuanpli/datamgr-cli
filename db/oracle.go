//go:build !linux && !windows
// +build !linux,!windows

package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/godror/godror"
)

// OracleConnection Oracle数据库连接
type OracleConnection struct {
	config *DbConfig
	db     *sql.DB
}

// NewOracleConnection 创建Oracle数据库连接
func NewOracleConnection(config *DbConfig) (*OracleConnection, error) {
	return &OracleConnection{
		config: config,
	}, nil
}

// Connect 连接到Oracle数据库
func (o *OracleConnection) Connect() error {
	// Oracle连接字符串格式: user/password@host:port/service_name
	connectionString := fmt.Sprintf("%s/%s@%s:%d/%s",
		o.config.User, o.config.Password, o.config.Host, o.config.Port, o.config.DbName)

	db, err := sql.Open("godror", connectionString)
	if err != nil {
		return err
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		return err
	}

	o.db = db
	return nil
}

// Disconnect 断开连接
func (o *OracleConnection) Disconnect() error {
	if o.db != nil {
		return o.db.Close()
	}
	return nil
}

// formatOracleTime 格式化Oracle时间为标准格式
func formatOracleTime(val interface{}) interface{} {
	if val == nil {
		return val
	}
	
	// 如果是字符串，处理成标准格式
	if strVal, ok := val.(string); ok && (strings.Contains(strVal, "-") || strings.Contains(strVal, "/")) {
		// 处理日期格式
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
func (o *OracleConnection) Query(query string) ([]map[string]interface{}, error) {
	if o.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	rows, err := o.db.Query(query)
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
				row[col] = formatOracleTime(val)
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
func (o *OracleConnection) QueryWithParams(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if o.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	rows, err := o.db.Query(query, args...)
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
				row[col] = formatOracleTime(val)
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
func (o *OracleConnection) Execute(query string) (int64, error) {
	if o.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result, err := o.db.Exec(query)
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
func (o *OracleConnection) ExecuteWithParams(query string, args ...interface{}) (int64, error) {
	if o.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result, err := o.db.Exec(query, args...)
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
func (o *OracleConnection) GetTables() ([]string, error) {
	if o.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// Oracle中查询所有用户表的SQL
	query := `SELECT table_name FROM user_tables ORDER BY table_name`

	rows, err := o.db.Query(query)
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

	return tables, nil
}

// DescribeTable 获取表结构
func (o *OracleConnection) DescribeTable(tableName string) ([]map[string]interface{}, error) {
	if o.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 查询表字段信息
	query := `
    SELECT 
        col.column_name AS "字段名",
        col.data_type || 
            CASE 
                WHEN col.data_type IN ('VARCHAR2', 'CHAR', 'NVARCHAR2', 'NCHAR') THEN '(' || col.char_length || ')'
                WHEN col.data_type = 'NUMBER' AND col.data_precision IS NOT NULL THEN '(' || col.data_precision || 
                    CASE WHEN col.data_scale > 0 THEN ',' || col.data_scale ELSE '' END || ')'
                ELSE ''
            END AS "数据类型",
        col.char_length AS "长度",
        CASE col.nullable WHEN 'Y' THEN 'Y' ELSE 'N' END AS "可空",
        CASE 
            WHEN pk.constraint_name IS NOT NULL THEN 'PRIMARY KEY'
            WHEN fk.constraint_name IS NOT NULL THEN 'FOREIGN KEY'
            WHEN unq.constraint_name IS NOT NULL THEN 'UNIQUE'
            ELSE ''
        END AS "约束",
        col.data_default AS "默认值",
        com.comments AS "描述"
    FROM 
        user_tab_columns col
    LEFT JOIN 
        user_col_comments com ON col.table_name = com.table_name AND col.column_name = com.column_name
    LEFT JOIN (
        SELECT 
            cons.constraint_name, 
            cols.column_name,
            cols.table_name
        FROM 
            user_constraints cons
        JOIN 
            user_cons_columns cols ON cons.constraint_name = cols.constraint_name
        WHERE 
            cons.constraint_type = 'P'
    ) pk ON col.column_name = pk.column_name AND col.table_name = pk.table_name
    LEFT JOIN (
        SELECT 
            cons.constraint_name, 
            cols.column_name,
            cols.table_name
        FROM 
            user_constraints cons
        JOIN 
            user_cons_columns cols ON cons.constraint_name = cols.constraint_name
        WHERE 
            cons.constraint_type = 'R'
    ) fk ON col.column_name = fk.column_name AND col.table_name = fk.table_name
    LEFT JOIN (
        SELECT 
            cons.constraint_name, 
            cols.column_name,
            cols.table_name
        FROM 
            user_constraints cons
        JOIN 
            user_cons_columns cols ON cons.constraint_name = cols.constraint_name
        WHERE 
            cons.constraint_type = 'U'
    ) unq ON col.column_name = unq.column_name AND col.table_name = unq.table_name
    WHERE 
        col.table_name = UPPER(?)
    ORDER BY 
        col.column_id
    `

	rows, err := o.db.Query(query, tableName)
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

// GetTableColumns 获取表列名
func (o *OracleConnection) GetTableColumns(tableName string) ([]string, error) {
	if o.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	query := `SELECT column_name FROM user_tab_columns WHERE table_name = UPPER(?) ORDER BY column_id`

	rows, err := o.db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var columns []string
	for rows.Next() {
		var colName string
		if err := rows.Scan(&colName); err != nil {
			return nil, err
		}
		columns = append(columns, colName)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return columns, nil
} 