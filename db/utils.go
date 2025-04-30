package db

import (
	"strconv"
)

// ParsePort 解析端口号字符串为整数
func ParsePort(portStr string) (int, error) {
	return strconv.Atoi(portStr)
} 