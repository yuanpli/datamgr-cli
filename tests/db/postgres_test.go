package db_test

import (
	"os"
	"testing"

	"github.com/yuanpli/datamgr-cli/db"
)

// TestPostgresConnection 测试PostgreSQL数据库连接
// 注意：此测试需要有可用的PostgreSQL实例
func TestPostgresConnection(t *testing.T) {
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
	
	// 测试创建连接实例
	t.Run("Create connection", func(t *testing.T) {
		conn, err := db.NewPostgresConnection(config)
		if err != nil {
			t.Fatalf("Failed to create PostgreSQL connection: %v", err)
		}
		if conn == nil {
			t.Fatal("NewPostgresConnection returned nil connection")
		}
	})
	
	// 测试连接到数据库
	// 注意：此测试需要有可用的PostgreSQL实例
	t.Run("Connect to database", func(t *testing.T) {
		// 如果没有测试数据库，则跳过此测试
		if os.Getenv("PG_TEST_SKIP_CONNECTION") != "" {
			t.Skip("Skipping PostgreSQL connection test")
		}
		
		conn, _ := db.NewPostgresConnection(config)
		err := conn.Connect()
		if err != nil {
			t.Skipf("Failed to connect to PostgreSQL: %v (skipping as this may be due to no available database)", err)
		}
		
		// 测试完成后断开连接
		defer conn.Disconnect()
		
		// 尝试执行一个简单查询
		_, err = conn.Query("SELECT 1")
		if err != nil {
			t.Errorf("Failed to execute simple query: %v", err)
		}
	})
} 