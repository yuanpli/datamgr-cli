package db

import (
	"errors"
	"fmt"
	"sync"
)

// DbConfig 数据库配置
type DbConfig struct {
	Type     string
	Host     string
	Port     int
	User     string
	Password string
	DbName   string
}

// Connection 数据库连接接口
type Connection interface {
	Connect() error
	Disconnect() error
	Query(query string) ([]map[string]interface{}, error)
	QueryWithParams(query string, args ...interface{}) ([]map[string]interface{}, error)
	Execute(query string) (int64, error)
	ExecuteWithParams(query string, args ...interface{}) (int64, error)
	GetTables() ([]string, error)
	DescribeTable(tableName string) ([]map[string]interface{}, error)
	GetTableColumns(tableName string) ([]string, error)
}

var (
	currentConnection Connection
	currentConfig     *DbConfig
	mu                sync.Mutex
)

// Connect 连接到数据库
func Connect(dbType, host string, port int, user, password, dbName string) error {
	mu.Lock()
	defer mu.Unlock()

	config := &DbConfig{
		Type:     dbType,
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DbName:   dbName,
	}

	var conn Connection
	var err error

	// 根据数据库类型创建不同的连接实例
	switch dbType {
	case "dameng":
		conn, err = NewDamengConnection(config)
	case "mysql":
		return errors.New("MySQL 数据库支持尚未实现")
	case "sqlite":
		return errors.New("SQLite 数据库支持尚未实现")
	case "postgresql":
		return errors.New("PostgreSQL 数据库支持尚未实现")
	default:
		return fmt.Errorf("不支持的数据库类型: %s", dbType)
	}

	if err != nil {
		return err
	}

	if err := conn.Connect(); err != nil {
		return err
	}

	currentConnection = conn
	currentConfig = config
	return nil
}

// GetCurrentConnection 获取当前连接
func GetCurrentConnection() Connection {
	mu.Lock()
	defer mu.Unlock()
	return currentConnection
}

// GetCurrentConfig 获取当前数据库配置
func GetCurrentConfig() *DbConfig {
	mu.Lock()
	defer mu.Unlock()
	return currentConfig
}

// Disconnect 断开当前连接
func Disconnect() error {
	mu.Lock()
	defer mu.Unlock()
	
	if currentConnection == nil {
		return errors.New("当前没有活动的数据库连接")
	}
	
	err := currentConnection.Disconnect()
	if err != nil {
		return err
	}
	
	currentConnection = nil
	currentConfig = nil
	return nil
} 