package cmd

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/yuanpli/datamgr-cli/db"
	"github.com/yuanpli/datamgr-cli/pkg/utils"
)

var (
	saveFlag    bool
	clearFlag   bool
	showFlag    bool
	configType  string
	configHost  string
	configPort  string
	configUser  string
	configPwd   string
	configDbName string
)

var configCmd = &cobra.Command{
	Use:   "config",
	Short: "管理默认连接配置",
	Long:  `管理数据库默认连接配置，支持保存、查看、修改和清除操作。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 根据标志执行不同的操作
		if clearFlag {
			// 清除配置
			if err := utils.ClearConfig(); err != nil {
				fmt.Printf("清除配置失败: %v\n", err)
				return
			}
			fmt.Println("已清除默认配置")
			return
		}

		if showFlag || (!saveFlag && configType == "" && configHost == "" && 
			configPort == "" && configUser == "" && configPwd == "" && configDbName == "") {
			// 显示配置
			config, err := utils.LoadConfig()
			if err != nil {
				fmt.Printf("加载配置失败: %v\n", err)
				return
			}
			fmt.Println("当前默认配置:")
			fmt.Printf("  数据库类型: %s\n", config.Type)
			fmt.Printf("  主机地址: %s\n", config.Host)
			fmt.Printf("  端口: %d\n", config.Port)
			fmt.Printf("  用户名: %s\n", config.User)
			fmt.Printf("  密码: %s\n", "********") // 不直接显示密码
			fmt.Printf("  数据库名: %s\n", config.DbName)
			return
		}

		// 保存当前连接或修改配置
		if saveFlag {
			// 使用当前连接信息保存
			currentConfig := db.GetCurrentConfig()
			if currentConfig == nil {
				fmt.Println("当前未连接到任何数据库，无法保存配置")
				return
			}

			config := &utils.Config{
				Type:     currentConfig.Type,
				Host:     currentConfig.Host,
				Port:     currentConfig.Port,
				User:     currentConfig.User,
				Password: currentConfig.Password,
				DbName:   currentConfig.DbName,
			}

			if err := utils.SaveConfig(config); err != nil {
				fmt.Printf("保存配置失败: %v\n", err)
				return
			}

			fmt.Println("已将当前连接信息保存为默认配置")
			return
		}

		// 修改配置
		// 先尝试加载现有配置
		var config *utils.Config
		existingConfig, err := utils.LoadConfig()
		if err == nil {
			// 有现有配置，以它为基础修改
			config = existingConfig
		} else {
			// 没有现有配置，创建新的
			config = &utils.Config{
				Type: "dameng", // 默认使用达梦数据库
				Port: 5236,     // 默认端口
			}
		}

		// 更新指定的字段
		if configType != "" {
			config.Type = configType
		}
		if configHost != "" {
			config.Host = configHost
		}
		if configPort != "" {
			port, err := strconv.Atoi(configPort)
			if err != nil {
				fmt.Printf("端口格式错误: %v\n", err)
				return
			}
			config.Port = port
		}
		if configUser != "" {
			config.User = configUser
		}
		if configPwd != "" {
			config.Password = configPwd
		}
		if configDbName != "" {
			config.DbName = configDbName
		}

		// 保存修改后的配置
		if err := utils.SaveConfig(config); err != nil {
			fmt.Printf("保存配置失败: %v\n", err)
			return
		}

		fmt.Println("配置已更新")
	},
}

func init() {
	configCmd.Flags().BoolVar(&saveFlag, "save", false, "保存当前连接为默认配置")
	configCmd.Flags().BoolVar(&clearFlag, "clear", false, "清除默认配置")
	configCmd.Flags().BoolVar(&showFlag, "show", false, "显示当前默认配置")
	configCmd.Flags().StringVar(&configType, "type", "", "设置数据库类型")
	configCmd.Flags().StringVar(&configHost, "host", "", "设置主机地址")
	configCmd.Flags().StringVar(&configPort, "port", "", "设置端口")
	configCmd.Flags().StringVar(&configUser, "user", "", "设置用户名")
	configCmd.Flags().StringVar(&configPwd, "password", "", "设置密码")
	configCmd.Flags().StringVar(&configDbName, "dbname", "", "设置数据库名")
} 