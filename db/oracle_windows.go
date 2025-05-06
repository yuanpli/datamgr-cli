//go:build windows && !cgo
// +build windows,!cgo

package db

import (
	"database/sql"
	"fmt"
)

// OracleConnectionWindows Windows平台Oracle数据库连接
type OracleConnectionWindows struct {
	config *DbConfig
	db     *sql.DB
}

// NewOracleConnection 创建Oracle数据库连接
func NewOracleConnection(config *DbConfig) (Connection, error) {
	return &OracleConnectionWindows{
		config: config,
	}, nil
}

// Connect 连接到Oracle数据库
func (o *OracleConnectionWindows) Connect() error {
	return fmt.Errorf("在当前构建的Windows版本中不支持Oracle数据库")
}

// Disconnect 断开连接
func (o *OracleConnectionWindows) Disconnect() error {
	return nil
}

// Query 执行查询语句
func (o *OracleConnectionWindows) Query(query string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("在当前构建的Windows版本中不支持Oracle数据库")
}

// QueryWithParams 执行带参数的查询语句
func (o *OracleConnectionWindows) QueryWithParams(query string, args ...interface{}) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("在当前构建的Windows版本中不支持Oracle数据库")
}

// Execute 执行更新/插入/删除语句
func (o *OracleConnectionWindows) Execute(query string) (int64, error) {
	return 0, fmt.Errorf("在当前构建的Windows版本中不支持Oracle数据库")
}

// ExecuteWithParams 执行带参数的更新/插入/删除语句
func (o *OracleConnectionWindows) ExecuteWithParams(query string, args ...interface{}) (int64, error) {
	return 0, fmt.Errorf("在当前构建的Windows版本中不支持Oracle数据库")
}

// GetTables 获取所有表
func (o *OracleConnectionWindows) GetTables() ([]string, error) {
	return nil, fmt.Errorf("在当前构建的Windows版本中不支持Oracle数据库")
}

// DescribeTable 获取表结构
func (o *OracleConnectionWindows) DescribeTable(tableName string) ([]map[string]interface{}, error) {
	return nil, fmt.Errorf("在当前构建的Windows版本中不支持Oracle数据库")
}

// GetTableColumns 获取表列名
func (o *OracleConnectionWindows) GetTableColumns(tableName string) ([]string, error) {
	return nil, fmt.Errorf("在当前构建的Windows版本中不支持Oracle数据库")
} 