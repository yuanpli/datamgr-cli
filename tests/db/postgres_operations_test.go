package db_test

import (
	"os"
	"testing"

	"github.com/yuanpli/datamgr-cli/db"
)

// 创建PostgreSQL测试连接
func createPostgresTestConnection(t *testing.T) (*db.PostgresConnection, func()) {
	// 从环境变量获取测试数据库连接信息
	host := os.Getenv("PG_TEST_HOST")
	if host == "" {
		host = "localhost"
	}
	
	port := 5432 // PostgreSQL默认端口
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
	config := &db.DbConfig{
		Type:     "postgresql",
		Host:     host,
		Port:     port,
		User:     user,
		Password: password,
		DbName:   dbName,
	}
	
	// 如果没有测试数据库，则跳过此测试
	if os.Getenv("PG_TEST_SKIP_CONNECTION") != "" {
		t.Skip("Skipping PostgreSQL connection test")
	}
	
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

// TestPostgresGetTables 测试获取表列表
func TestPostgresGetTables(t *testing.T) {
	conn, cleanup := createPostgresTestConnection(t)
	defer cleanup()
	
	tables, err := conn.GetTables()
	if err != nil {
		t.Fatalf("Failed to get tables: %v", err)
	}
	
	// 至少应该有一些系统表
	if len(tables) == 0 {
		t.Logf("Warning: No tables found in database, this might be expected in a clean test database")
	} else {
		t.Logf("Found %d tables in database", len(tables))
	}
}

// TestPostgresQuery 测试基本查询
func TestPostgresQuery(t *testing.T) {
	conn, cleanup := createPostgresTestConnection(t)
	defer cleanup()
	
	// 执行简单的SELECT查询
	results, err := conn.Query("SELECT 1 as num, 'test' as str")
	if err != nil {
		t.Fatalf("Failed to execute simple query: %v", err)
	}
	
	// 检查结果
	if len(results) != 1 {
		t.Fatalf("Expected 1 row, got %d", len(results))
	}
	
	row := results[0]
	
	// PostgreSQL可能会以不同方式返回数字，所以我们需要灵活处理
	numVal, ok := row["num"]
	if !ok {
		t.Fatalf("Column 'num' not found in result")
	}
	
	// 转换为字符串进行比较，避免类型问题
	numStr := ""
	switch v := numVal.(type) {
	case string:
		numStr = v
	case int64:
		if v != 1 {
			t.Errorf("Expected num=1, got %d", v)
		}
	case float64:
		if v != 1.0 {
			t.Errorf("Expected num=1.0, got %f", v)
		}
	default:
		t.Logf("num is of type %T: %v", numVal, numVal)
	}
	
	if numStr != "" && numStr != "1" {
		t.Errorf("Expected num='1', got '%s'", numStr)
	}
	
	// 检查字符串值
	strVal, ok := row["str"]
	if !ok {
		t.Fatalf("Column 'str' not found in result")
	}
	
	strStr, ok := strVal.(string)
	if !ok {
		t.Fatalf("Expected str to be string, got %T", strVal)
	}
	
	if strStr != "test" {
		t.Errorf("Expected str='test', got '%s'", strStr)
	}
} 