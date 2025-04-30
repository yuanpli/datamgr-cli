package db_test

import (
	"os"
	"strconv"
	"testing"

	"github.com/yuanpli/datamgr-cli/db"
)

// PostgresTestConfig 返回PostgreSQL测试配置
func PostgresTestConfig() *db.DbConfig {
	// 从环境变量获取测试数据库连接信息
	host := os.Getenv("PG_TEST_HOST")
	if host == "" {
		host = "localhost"
	}
	
	port := 5432 // PostgreSQL默认端口
	if portStr := os.Getenv("PG_TEST_PORT"); portStr != "" {
		// 忽略错误，使用默认端口
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			port = p
		}
	}
	
	user := os.Getenv("PG_TEST_USER")
	if user == "" {
		user = "postgres"
	}
	
	password := os.Getenv("PG_TEST_PASSWORD")
	if password == "" {
		password = "postgres"
	}
	
	dbName := os.Getenv("PG_TEST_DBNAME")
	if dbName == "" {
		dbName = "postgres"
	}
	
	// 创建配置
	return &db.DbConfig{
		Type:     "postgresql",
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DbName:   dbName,
	}
}

// SkipIfNoPostgres 如果没有可用的PostgreSQL测试环境，则跳过测试
func SkipIfNoPostgres(t *testing.T) {
	if os.Getenv("PG_TEST_SKIP_CONNECTION") != "" {
		t.Skip("Skipping PostgreSQL connection test")
	}
}

// SetupPostgresConnection 设置PostgreSQL测试连接并返回清理函数
func SetupPostgresConnection(t *testing.T) (db.Connection, func()) {
	SkipIfNoPostgres(t)
	
	config := PostgresTestConfig()
	
	conn, err := db.NewPostgresConnection(config)
	if err != nil {
		t.Skipf("Failed to create PostgreSQL connection: %v", err)
	}
	
	err = conn.Connect()
	if err != nil {
		t.Skipf("Failed to connect to PostgreSQL: %v (skipping as this may be due to no available database)", err)
	}
	
	// 返回清理函数
	cleanup := func() {
		conn.Disconnect()
	}
	
	return conn, cleanup
} 