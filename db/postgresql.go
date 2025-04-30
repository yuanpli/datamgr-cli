package db

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

// PostgresConnection PostgreSQL数据库连接
type PostgresConnection struct {
	config *DbConfig
	db     *sql.DB
}

// NewPostgresConnection 创建PostgreSQL数据库连接
func NewPostgresConnection(config *DbConfig) (*PostgresConnection, error) {
	return &PostgresConnection{
		config: config,
	}, nil
}

// Connect 连接到PostgreSQL数据库
func (p *PostgresConnection) Connect() error {
	connectionString := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		p.config.Host, p.config.Port, p.config.User, p.config.Password, p.config.DbName)

	db, err := sql.Open("postgres", connectionString)
	if err != nil {
		return err
	}

	// 测试连接
	if err = db.Ping(); err != nil {
		return err
	}

	p.db = db
	return nil
}

// Disconnect 断开连接
func (p *PostgresConnection) Disconnect() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// formatPostgresTime 格式化PostgreSQL时间为标准格式
func formatPostgresTime(val interface{}) interface{} {
	if val == nil {
		return val
	}
	
	// 如果是字符串，处理成标准格式
	if strVal, ok := val.(string); ok && (strings.Contains(strVal, "-") || strings.Contains(strVal, ":")) {
		// 处理带时区的格式
		if strings.Contains(strVal, "+") || strings.Contains(strVal, "-") {
			parts := strings.Fields(strVal)
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
func (p *PostgresConnection) Query(query string) ([]map[string]interface{}, error) {
	if p.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	rows, err := p.db.Query(query)
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
				row[col] = formatPostgresTime(val)
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
func (p *PostgresConnection) QueryWithParams(query string, args ...interface{}) ([]map[string]interface{}, error) {
	if p.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	rows, err := p.db.Query(query, args...)
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
				row[col] = formatPostgresTime(val)
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
func (p *PostgresConnection) Execute(query string) (int64, error) {
	if p.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result, err := p.db.Exec(query)
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
func (p *PostgresConnection) ExecuteWithParams(query string, args ...interface{}) (int64, error) {
	if p.db == nil {
		return 0, fmt.Errorf("数据库未连接")
	}

	result, err := p.db.Exec(query, args...)
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
func (p *PostgresConnection) GetTables() ([]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	query := `
		SELECT table_name 
		FROM information_schema.tables 
		WHERE table_schema = 'public' 
		AND table_type = 'BASE TABLE'
		ORDER BY table_name
	`
	rows, err := p.db.Query(query)
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
func (p *PostgresConnection) DescribeTable(tableName string) ([]map[string]interface{}, error) {
	if p.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 查询表结构和字段注释
	query := `
		SELECT 
			c.column_name, 
			c.data_type, 
			c.character_maximum_length AS data_length, 
			c.is_nullable, 
			CASE 
				WHEN pk.column_name IS NOT NULL THEN 'PRIMARY KEY' 
				ELSE '' 
			END AS constraint_type,
			COALESCE(pgd.description, '') AS description,
			CASE 
				WHEN c.column_default LIKE 'nextval%' THEN 'IDENTITY' 
				ELSE '' 
			END AS identity_info
		FROM 
			information_schema.columns c
		LEFT JOIN 
			(
				SELECT tc.table_schema, tc.table_name, kcu.column_name
				FROM information_schema.table_constraints tc
				JOIN information_schema.key_column_usage kcu 
					ON tc.constraint_name = kcu.constraint_name 
					AND tc.table_schema = kcu.table_schema
				WHERE tc.constraint_type = 'PRIMARY KEY'
			) pk 
			ON c.table_schema = pk.table_schema 
			AND c.table_name = pk.table_name 
			AND c.column_name = pk.column_name
		LEFT JOIN 
			pg_catalog.pg_statio_all_tables st 
			ON st.schemaname = c.table_schema 
			AND st.relname = c.table_name
		LEFT JOIN 
			pg_catalog.pg_description pgd 
			ON pgd.objoid = st.relid 
			AND pgd.objsubid = c.ordinal_position
		WHERE 
			c.table_schema = 'public' 
			AND c.table_name = $1
		ORDER BY 
			c.ordinal_position
	`

	rows, err := p.db.Query(query, strings.ToLower(tableName))
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
func (p *PostgresConnection) GetTableColumns(tableName string) ([]string, error) {
	if p.db == nil {
		return nil, fmt.Errorf("数据库未连接")
	}

	// 查询表结构，按列ID排序
	query := `
		SELECT 
			column_name
		FROM 
			information_schema.columns
		WHERE 
			table_schema = 'public' 
			AND table_name = $1
		ORDER BY 
			ordinal_position
	`

	rows, err := p.db.Query(query, strings.ToLower(tableName))
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