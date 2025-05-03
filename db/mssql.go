package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// MSSQLConnection MSSQL数据库连接
type MSSQLConnection struct {
	config *DbConfig
	db     *sql.DB
}

// NewMSSQLConnection 创建MSSQL数据库连接
func NewMSSQLConnection(config *DbConfig) (*MSSQLConnection, error) {
	return &MSSQLConnection{
		config: config,
	}, nil
}

// Connect 连接到MSSQL数据库
func (m *MSSQLConnection) Connect() error {
	// MSSQL连接字符串格式: sqlserver://username:password@host:port?database=dbname
	connectionString := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		m.config.User, m.config.Password, m.config.Host, m.config.Port, m.config.DbName)

	db, err := sql.Open("sqlserver", connectionString)
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
func (m *MSSQLConnection) Disconnect() error {
	if m.db != nil {
		return m.db.Close()
	}
	return nil
}

// formatMSSQLTime 格式化MSSQL时间为标准格式
func formatMSSQLTime(val interface{}) interface{} {
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
func (m *MSSQLConnection) Query(query string) ([]map[string]interface{}, error) {
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
				row[col] = formatMSSQLTime(val)
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
func (m *MSSQLConnection) QueryWithParams(query string, args ...interface{}) ([]map[string]interface{}, error) {
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
				row[col] = formatMSSQLTime(val)
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
func (m *MSSQLConnection) Execute(query string) (int64, error) {
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
func (m *MSSQLConnection) ExecuteWithParams(query string, args ...interface{}) (int64, error) {
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
func (m *MSSQLConnection) GetTables() ([]string, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// MSSQL中查询所有用户表的SQL
	query := `SELECT TABLE_NAME 
              FROM INFORMATION_SCHEMA.TABLES 
              WHERE TABLE_TYPE = 'BASE TABLE' AND TABLE_CATALOG = ?
              ORDER BY TABLE_NAME`

	rows, err := m.db.Query(query, m.config.DbName)
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
func (m *MSSQLConnection) DescribeTable(tableName string) ([]map[string]interface{}, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 查询表字段信息
	query := `
    SELECT 
        c.COLUMN_NAME AS "字段名",
        c.DATA_TYPE + 
            CASE 
                WHEN c.DATA_TYPE IN ('char', 'varchar', 'nchar', 'nvarchar') THEN '(' + CAST(c.CHARACTER_MAXIMUM_LENGTH AS VARCHAR) + ')'
                WHEN c.DATA_TYPE IN ('decimal', 'numeric') THEN '(' + CAST(c.NUMERIC_PRECISION AS VARCHAR) + ',' + CAST(c.NUMERIC_SCALE AS VARCHAR) + ')'
                ELSE ''
            END AS "数据类型",
        CAST(CASE 
            WHEN c.CHARACTER_MAXIMUM_LENGTH IS NOT NULL THEN c.CHARACTER_MAXIMUM_LENGTH
            WHEN c.NUMERIC_PRECISION IS NOT NULL THEN c.NUMERIC_PRECISION
            ELSE NULL
        END AS VARCHAR) AS "长度",
        c.IS_NULLABLE AS "可空",
        CASE 
            WHEN pk.COLUMN_NAME IS NOT NULL THEN 'PRIMARY KEY'
            WHEN fk.COLUMN_NAME IS NOT NULL THEN 'FOREIGN KEY'
            ELSE ''
        END AS "约束",
        ep.value AS "描述"
    FROM 
        INFORMATION_SCHEMA.COLUMNS c
    LEFT JOIN (
        SELECT 
            ku.TABLE_CATALOG,
            ku.TABLE_SCHEMA,
            ku.TABLE_NAME,
            ku.COLUMN_NAME
        FROM 
            INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
        JOIN 
            INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
                ON tc.CONSTRAINT_TYPE = 'PRIMARY KEY'
                AND tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
        ) pk 
            ON c.TABLE_CATALOG = pk.TABLE_CATALOG
            AND c.TABLE_SCHEMA = pk.TABLE_SCHEMA
            AND c.TABLE_NAME = pk.TABLE_NAME
            AND c.COLUMN_NAME = pk.COLUMN_NAME
    LEFT JOIN (
        SELECT 
            ku.TABLE_CATALOG,
            ku.TABLE_SCHEMA,
            ku.TABLE_NAME,
            ku.COLUMN_NAME
        FROM 
            INFORMATION_SCHEMA.TABLE_CONSTRAINTS tc
        JOIN 
            INFORMATION_SCHEMA.KEY_COLUMN_USAGE ku
                ON tc.CONSTRAINT_TYPE = 'FOREIGN KEY'
                AND tc.CONSTRAINT_NAME = ku.CONSTRAINT_NAME
        ) fk 
            ON c.TABLE_CATALOG = fk.TABLE_CATALOG
            AND c.TABLE_SCHEMA = fk.TABLE_SCHEMA
            AND c.TABLE_NAME = fk.TABLE_NAME
            AND c.COLUMN_NAME = fk.COLUMN_NAME
    LEFT JOIN (
        SELECT 
            t.name AS TableName,
            c.name AS ColumnName,
            p.value
        FROM 
            sys.tables t
        JOIN 
            sys.columns c ON t.object_id = c.object_id
        JOIN 
            sys.extended_properties p ON p.major_id = c.object_id AND p.minor_id = c.column_id AND p.name = 'MS_Description'
        ) ep 
            ON c.TABLE_NAME = ep.TableName
            AND c.COLUMN_NAME = ep.ColumnName
    WHERE 
        c.TABLE_CATALOG = ? AND c.TABLE_NAME = ?
    ORDER BY 
        c.ORDINAL_POSITION
    `

	rows, err := m.db.Query(query, m.config.DbName, tableName)
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
func (m *MSSQLConnection) GetTableColumns(tableName string) ([]string, error) {
	if m.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	query := `SELECT COLUMN_NAME 
              FROM INFORMATION_SCHEMA.COLUMNS 
              WHERE TABLE_CATALOG = ? AND TABLE_NAME = ? 
              ORDER BY ORDINAL_POSITION`

	rows, err := m.db.Query(query, m.config.DbName, tableName)
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