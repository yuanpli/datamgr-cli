# DATAMGR-CLI Universal Data Management Tool

A command-line tool for universal data management that supports multiple database types, providing a unified interface for table operations.

## Features

- Supports multiple database types (DM/MySQL/SQLite/PostgreSQL, etc.), with DM as the default
- Provides a unified interface for table operations
- SQL-like command syntax
- Interactive CLI with auto-completion and command history
- Supports direct connection via command-line parameters

## Installation

### Build from Source

```bash
git clone https://github.com/yuanpli/datamgr-cli.git
cd datamgr-cli
go build
```

## Usage

### Start the Program

```bash
./datamgr-cli
```

The program will automatically try to load default configuration and connect to the database on startup.

### Connect to a Database

Connect via command-line parameters:

```bash
./datamgr-cli connect -H <host> -P <port> -u <user> -p <password> -D <dbname>
```

Or connect interactively:

```
datamgr> connect
```

Follow the prompts to enter connection details.

### Default Configuration Management

You can save the current connection as the default configuration:

```
datamgr> config save
```

The program will automatically use this configuration to connect to the database the next time you start it.

### Available Commands

#### System Commands

- `help` - Show a list of available commands
- `connect` - Connect to a specified database
- `status` - Show current connection status
- `exit/quit` - Exit the program
- `clear` - Clear the screen

#### Configuration Commands

- `config` - Show current default configuration
- `config save` - Save current connection as default configuration
- `config set <item> <value>` - Set configuration items (type/host/port/user/password/dbname)
- `config clear` - Clear default configuration

#### Table Interaction Commands

- `show tables` - List all available tables
- `desc table <table_name>` - Show table structure details

#### Universal Data Operation Commands

```sql
-- Query
SELECT [field_list] FROM <table> [WHERE condition] [LIMIT count]

-- Insert
INSERT INTO <table> (field1, field2, ...) VALUES (value1, value2, ...)

-- Update
UPDATE <table> SET field=value [WHERE condition]

-- Delete
DELETE FROM <table> [WHERE condition]

-- Bulk import (Not yet implemented)
IMPORT <table> FROM <file> [FORMAT csv/excel]

-- Export (Not yet implemented)
EXPORT <table> [WHERE condition] <file> [FORMAT csv/excel]
```

## Examples

```
datamgr> connect -H localhost -u SYSDBA -p SYSDBA -D DAMENG
Successfully connected to dameng database: DAMENG

datamgr[DAMENG]> show tables
Table list:
  1) EMPLOYEES
  2) DEPARTMENTS
  3) PRODUCTS

datamgr[DAMENG]> desc table EMPLOYEES
Structure of table EMPLOYEES:
Field Name            Data Type          Length        Nullable      Constraints          
---------------------------------------------------------------------------
ID                   NUMBER             22            N            PRIMARY KEY  
NAME                 VARCHAR2           100           N                       
DEPARTMENT_ID        NUMBER             22            Y                       
SALARY               NUMBER             22            Y                       
HIRE_DATE            DATE               7             Y                       

datamgr[DAMENG]> SELECT * FROM EMPLOYEES WHERE DEPARTMENT_ID = 1
ID                  NAME                DEPARTMENT_ID       SALARY              HIRE_DATE           
----------------------------------------------------------------------------------------------------
1                   Zhang San           1                   10000               2022-01-01          
2                   Li Si               1                   12000               2022-02-15          

Total 2 rows
```

## Development Requirements

- Go 1.18 or higher
- DM database driver `gitee.com/chunanyong/dm`

## README Links

- [English Version](README.md)
- [中文版本](README_zh.md)

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

Copyright © 2023 datamgr-cli Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.