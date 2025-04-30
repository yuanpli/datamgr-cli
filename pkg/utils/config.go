package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

// Config 存储数据库连接的配置信息
type Config struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DbName   string `json:"dbname"`
}

const (
	configFileName = "datamgr-cli-config.json"
)

// GetConfigDir 获取配置文件目录
func GetConfigDir() (string, error) {
	var configDir string

	switch runtime.GOOS {
	case "windows":
		// Windows: %USERPROFILE%\.datamgr-cli
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".datamgr-cli")
	case "darwin", "linux":
		// macOS/Linux: ~/.datamgr-cli
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		configDir = filepath.Join(homeDir, ".datamgr-cli")
	default:
		return "", errors.New("不支持的操作系统")
	}

	// 确保目录存在
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return "", err
	}

	return configDir, nil
}

// GetConfigFilePath 获取配置文件路径
func GetConfigFilePath() (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, configFileName), nil
}

// SaveConfig 保存配置到文件
func SaveConfig(config *Config) error {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

// LoadConfig 从文件加载配置
func LoadConfig() (*Config, error) {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return nil, err
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, errors.New("默认配置不存在")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// ClearConfig 清除配置文件
func ClearConfig() error {
	configPath, err := GetConfigFilePath()
	if err != nil {
		return err
	}

	// 检查文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return errors.New("默认配置不存在")
	}

	// 删除配置文件
	return os.Remove(configPath)
}

// DisplayConfig 显示当前配置信息
func DisplayConfig() (*Config, error) {
	config, err := LoadConfig()
	if err != nil {
		return nil, err
	}
	
	fmt.Println("当前默认配置:")
	fmt.Printf("  数据库类型: %s\n", config.Type)
	fmt.Printf("  主机地址: %s\n", config.Host)
	fmt.Printf("  端口: %d\n", config.Port)
	fmt.Printf("  用户名: %s\n", config.User)
	fmt.Printf("  密码: %s\n", "********") // 不直接显示密码
	fmt.Printf("  数据库名: %s\n", config.DbName)
	
	return config, nil
} 