package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yuanpli/datamgr-cli/db"
	"github.com/yuanpli/datamgr-cli/pkg/handler"
	"github.com/yuanpli/datamgr-cli/pkg/prompt"
)

var (
	dbType   string
	host     string
	port     int
	user     string
	password string
	dbName   string
)

var connectCmd = &cobra.Command{
	Use:   "connect",
	Short: "连接到数据库",
	Long:  `连接到指定的数据库。支持达梦、MySQL、SQLite、PostgreSQL、Oracle、MS SQL Server等。`,
	Run: func(cmd *cobra.Command, args []string) {
		// 如果没有提供命令行参数，则启动交互式连接向导
		if !cmd.Flags().Changed("host") && !cmd.Flags().Changed("user") && 
		   !cmd.Flags().Changed("password") && !cmd.Flags().Changed("dbname") {
			// 使用交互式连接向导
			if err := handler.HandleInteractiveConnect(); err != nil {
				fmt.Printf("连接失败: %v\n", err)
				return
			}
			// 成功连接后启动交互式命令行
			prompt.Start()
		} else {
			// 检查必要参数
			if host == "" || user == "" || password == "" || dbName == "" {
				fmt.Println("连接参数不完整，请提供主机、用户名、密码和数据库名")
				fmt.Println("将启动交互式连接向导...")
				if err := handler.HandleInteractiveConnect(); err != nil {
					fmt.Printf("连接失败: %v\n", err)
					return
				}
				// 成功连接后启动交互式命令行
				prompt.Start()
				return
			}
			
			// 使用命令行参数连接
			if err := db.Connect(dbType, host, port, user, password, dbName); err != nil {
				fmt.Printf("连接失败: %v\n", err)
				return
			}
			fmt.Printf("已成功连接到 %s 数据库: %s\n", dbType, dbName)
			
			// 成功连接后启动交互式命令行
			prompt.Start()
		}
	},
}

func init() {
	// 默认为达梦数据库
	connectCmd.Flags().StringVar(&dbType, "type", "dameng", "数据库类型 (dameng, mysql, sqlite, postgresql, oracle, mssql)")
	connectCmd.Flags().StringVarP(&host, "host", "H", "", "数据库主机地址")
	connectCmd.Flags().IntVarP(&port, "port", "P", 5236, "数据库端口")
	connectCmd.Flags().StringVarP(&user, "user", "u", "", "数据库用户名")
	connectCmd.Flags().StringVarP(&password, "password", "p", "", "数据库密码")
	connectCmd.Flags().StringVarP(&dbName, "dbname", "D", "", "数据库名称")
} 