# BWTY 通用数据管理工具

一个支持多种数据库的通用数据管理命令行工具，提供统一的表管理操作接口。

## 特性

- 支持多种数据库类型（达梦/MySQL/SQLite/PostgreSQL等），默认为达梦数据库
- 提供统一的表管理操作接口
- 类SQL的命令语法
- 交互式命令行界面，带有自动补全和命令历史记录
- 支持直接通过命令行参数指定连接信息

## 安装

### 从源码构建

```bash
git clone https://github.com/bwty/[DB_TOOL].git
cd [DB_TOOL]
go build
```

## 使用方法

### 启动程序

```bash
./[DB_TOOL]
```

### 连接数据库

通过命令行参数连接：

```bash
./[DB_TOOL] connect -H <host> -P <port> -u <user> -p <password> -D <dbname>
```

或者在交互式命令行中连接：

```
db> connect
```

然后按照提示输入连接信息。

### 可用命令

#### 系统命令

- `help` - 显示所有可用命令列表
- `connect` - 连接到指定数据库
- `status` - 显示当前连接状态
- `exit/quit` - 退出程序
- `clear` - 清屏

#### 表清单交互命令

- `show tables` - 列出所有可用数据表
- `desc table <table_name>` - 显示表结构详情

#### 通用数据操作命令

```sql
-- 查询操作
SELECT [字段列表] FROM <table> [WHERE 条件] [LIMIT 数量]

-- 新增记录
INSERT INTO <table> (字段名1,字段名2,....)VALUES (值1,值2,.....)

-- 更新记录
UPDATE <table> SET 字段=值 [WHERE 条件]

-- 删除记录
DELETE FROM <table> [WHERE 条件]

-- 批量导入（尚未实现）
IMPORT <table> FROM <file> [FORMAT csv/excel]

-- 导出（尚未实现）
EXPORT <table> [WHERE 条件] <file> [FORMAT csv/excel]
```

## 示例

```
db> connect -H localhost -u SYSDBA -p SYSDBA -D DAMENG
已成功连接到 dameng 数据库: DAMENG

db[DAMENG]> show tables
表列表:
  1) EMPLOYEES
  2) DEPARTMENTS
  3) PRODUCTS

db[DAMENG]> desc table EMPLOYEES
表 EMPLOYEES 的结构:
字段名                 数据类型           长度         可空         约束          
---------------------------------------------------------------------------
ID                   NUMBER             22          N           PRIMARY KEY  
NAME                 VARCHAR2           100         N                       
DEPARTMENT_ID        NUMBER             22          Y                       
SALARY               NUMBER             22          Y                       
HIRE_DATE            DATE               7           Y                       

db[DAMENG]> SELECT * FROM EMPLOYEES WHERE DEPARTMENT_ID = 1
ID                  NAME                DEPARTMENT_ID       SALARY              HIRE_DATE           
----------------------------------------------------------------------------------------------------
1                   张三                 1                   10000               2022-01-01          
2                   李四                 1                   12000               2022-02-15          

共 2 行结果
```

## 开发环境要求

- Go 1.18 或更高版本
- 达梦数据库驱动 `gitee.com/chunanyong/dm` 