# 测试说明

本目录包含datamgr-cli项目的测试代码，按照TDD（测试驱动开发）风格组织。

## 目录结构

- `db/` - 数据库相关测试
  - `main_test.go` - 测试辅助函数和通用配置
  - `postgres_test.go` - PostgreSQL连接测试
  - `postgres_operations_test.go` - PostgreSQL基本操作测试
  - `integration_test.go` - 数据库集成测试

## 运行测试

### 运行所有测试

```bash
go test ./tests/... -v
```

### 运行特定模块测试

```bash
go test ./tests/db -v
```

### 运行特定测试

```bash
go test ./tests/db -run TestPostgresQuery -v
```

## 配置PostgreSQL测试环境

测试会尝试连接到PostgreSQL数据库，如果没有可用的数据库，则会跳过需要实际连接的测试。你可以通过以下环境变量配置PostgreSQL测试环境：

- `PG_TEST_HOST` - PostgreSQL服务器主机名（默认为localhost）
- `PG_TEST_PORT` - PostgreSQL服务器端口（默认为5432）
- `PG_TEST_USER` - 用户名（默认为postgres）
- `PG_TEST_PASSWORD` - 密码（默认为postgres）
- `PG_TEST_DBNAME` - 数据库名（默认为postgres）
- `PG_TEST_SKIP_CONNECTION` - 设置此变量将跳过所有需要实际连接的测试

示例：

```bash
export PG_TEST_HOST=localhost
export PG_TEST_PORT=5432
export PG_TEST_USER=postgres
export PG_TEST_PASSWORD=mysecretpassword
export PG_TEST_DBNAME=testdb

go test ./tests/db -v
```

如果你没有可用的PostgreSQL环境，但仍然想运行测试，可以这样做：

```bash
export PG_TEST_SKIP_CONNECTION=1
go test ./tests/db -v
```

这将只运行不需要实际连接的测试。 