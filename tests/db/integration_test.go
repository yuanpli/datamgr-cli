package db_test

import (
	"testing"

	"github.com/yuanpli/datamgr-cli/db"
)

// TestIntegrationPostgresConnect 测试通过Connect函数连接到PostgreSQL数据库
func TestIntegrationPostgresConnect(t *testing.T) {
	SkipIfNoPostgres(t)
	
	config := PostgresTestConfig()
	
	// 测试Connect函数
	err := db.Connect(
		config.Type,
		config.Host,
		config.Port,
		config.User,
		config.Password,
		config.DbName,
	)
	if err != nil {
		t.Skipf("Failed to connect to PostgreSQL: %v (skipping as this may be due to no available database)", err)
	}
	
	// 测试结束后断开连接
	defer db.Disconnect()
	
	// 获取当前连接
	conn := db.GetCurrentConnection()
	if conn == nil {
		t.Fatal("GetCurrentConnection returned nil")
	}
	
	// 获取当前配置
	currentConfig := db.GetCurrentConfig()
	if currentConfig == nil {
		t.Fatal("GetCurrentConfig returned nil")
	}
	
	// 验证配置
	if currentConfig.Type != config.Type {
		t.Errorf("Expected Type='%s', got '%s'", config.Type, currentConfig.Type)
	}
	
	if currentConfig.Host != config.Host {
		t.Errorf("Expected Host='%s', got '%s'", config.Host, currentConfig.Host)
	}
	
	if currentConfig.Port != config.Port {
		t.Errorf("Expected Port=%d, got %d", config.Port, currentConfig.Port)
	}
	
	if currentConfig.User != config.User {
		t.Errorf("Expected User='%s', got '%s'", config.User, currentConfig.User)
	}
	
	if currentConfig.DbName != config.DbName {
		t.Errorf("Expected DbName='%s', got '%s'", config.DbName, currentConfig.DbName)
	}
	
	// 测试执行简单查询
	results, err := conn.Query("SELECT 1 as test")
	if err != nil {
		t.Errorf("Failed to execute simple query: %v", err)
	}
	
	if len(results) != 1 {
		t.Errorf("Expected 1 row, got %d", len(results))
	}
} 