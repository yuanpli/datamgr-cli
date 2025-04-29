# 通用数据管理工具需求清单

## 1. 核心功能架构
- 支持多种数据库类型（达梦/MySQL/SQLite/PostgreSQL等），默认为达梦数据库
- 提供统一的表管理操作接口


## 2. 连接管理

### 2.1 连接方式
- 支持直接通过命令行参数指定连接信息：`[DB_TOOL] connect --type <db_type> -H <host> -P <port> -u <user> -p <password> -D <dbname>`
- 交互式连接向导：逐步提示输入连接参数
- 当前只需要支持db_type为达梦数据库，预留其他数据库。默认可以不指定type，默认为达梦数据库

## 3. 命令行界面

### 3.1 提示符设计
- 提示符显示当前连接的数据库名称：`db[dp_data]>`
- 提示符显示命令执行状态（成功/失败）的颜色提示

### 3.2 命令格式
- 采用类SQL的命令语法
- 支持以分号(;)结束命令
- 支持命令自动补全和历史记录

## 4. 核心命令集

### 4.1 系统命令
- `help` - 显示所有可用命令列表
- `connect` - 连接到指定数据库（切换数据库连接）
- `status` - 显示当前连接状态
- `exit/quit` - 退出程序
- `clear` - 清屏

### 4.2 表清单交互命令
1. `show tables` - 列出所有可用数据表
2. `desc table <table_name>` 显示表结构详情（字段名/类型/约束）

### 4.3 通用数据操作管理命令集
```sql
-- 查询操作
SELECT [字段列表] FROM <table> [WHERE 条件] [LIMIT 数量]

-- 新增记录
INSERT INTO <table> (字段名1,字段名2,....)VALUES (值1,值2,.....)

-- 更新记录
UPDATE <table> SET 字段=值 [WHERE 条件]

-- 删除记录
DELETE FROM <table> [WHERE 条件]

-- 批量导入
IMPORT <table> FROM <file> [FORMAT csv/excel]

-- 导出
EXPORT <table> [WHERE 条件] <file> [FORMAT csv/excel]