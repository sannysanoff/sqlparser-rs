# SQL Parser for Go

A complete SQL parser for Go, transpiled from the popular Rust sqlparser-rs library.

## Features

- **14 SQL Dialects**: PostgreSQL, MySQL, SQLite, BigQuery, Snowflake, DuckDB, ClickHouse, Hive, MSSQL, Redshift, Databricks, Oracle, ANSI SQL, and Generic
- **Complete SQL Support**: SELECT, INSERT, UPDATE, DELETE, CREATE, ALTER, DROP, MERGE, transactions, and more
- **Robust Tokenizer**: Handles all SQL literal formats, operators, and comments
- **Pratt Parser**: Efficient expression parsing with correct operator precedence
- **Source Locations**: Accurate error reporting with line and column numbers
- **SQL Regeneration**: Parse SQL and regenerate it from the AST
- **Fuzz Tested**: Comprehensive fuzz testing for robustness

## Installation

```bash
go get github.com/user/sqlparser
```

## Quick Start

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/user/sqlparser/parser"
    "github.com/user/sqlparser/dialects/generic"
)

func main() {
    sql := `SELECT id, name, email 
            FROM users 
            WHERE active = true 
            ORDER BY created_at DESC 
            LIMIT 10`
    
    dialect := generic.NewGenericDialect()
    statements, err := parser.ParseSQL(dialect, sql)
    if err != nil {
        log.Fatal(err)
    }
    
    for _, stmt := range statements {
        // Print the regenerated SQL
        fmt.Println(stmt.String())
    }
}
```

## Dialect-Specific Example

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/user/sqlparser/parser"
    "github.com/user/sqlparser/dialects/postgresql"
    "github.com/user/sqlparser/dialects/mysql"
    "github.com/user/sqlparser/dialects/bigquery"
)

func main() {
    // PostgreSQL with array syntax
    pg := postgresql.NewPostgreSqlDialect()
    _, err := parser.ParseSQL(pg, "SELECT ARRAY[1, 2, 3]")
    if err != nil {
        log.Fatal(err)
    }
    
    // MySQL with backtick identifiers
    mysql := mysql.NewMySqlDialect()
    _, err = parser.ParseSQL(mysql, "SELECT * FROM `my-table`")
    if err != nil {
        log.Fatal(err)
    }
    
    // BigQuery with STRUCT
    bq := bigquery.NewBigQueryDialect()
    _, err = parser.ParseSQL(bq, "SELECT STRUCT(1 AS x, 'a' AS y)")
    if err != nil {
        log.Fatal(err)
    }
    
    fmt.Println("All dialects work!")
}
```

## Working with the AST

```go
package main

import (
    "fmt"
    "log"
    
    "github.com/user/sqlparser/parser"
    "github.com/user/sqlparser/dialects/generic"
)

func main() {
    sql := "SELECT id, name FROM users WHERE active = true"
    
    dialect := generic.NewGenericDialect()
    statements, err := parser.ParseSQL(dialect, sql)
    if err != nil {
        log.Fatal(err)
    }
    
    // Access the parsed statement
    stmt := statements[0]
    fmt.Printf("Statement type: %T\n", stmt)
    fmt.Printf("SQL: %s\n", stmt.String())
    
    // Use type assertion or switch to access specific statement details
    // See GOLANG.md for detailed AST traversal examples
}
```

## Supported SQL

### SELECT Statements
```sql
SELECT * FROM table_name
SELECT col1, col2 FROM table_name WHERE condition
SELECT DISTINCT col1 FROM table_name ORDER BY col1 DESC
SELECT * FROM t1 JOIN t2 ON t1.id = t2.id
SELECT * FROM table_name LIMIT 10 OFFSET 5
SELECT * FROM table_name WITH (NOLOCK)
WITH cte AS (SELECT * FROM t) SELECT * FROM cte
SELECT * FROM table_name WINDOW w AS (PARTITION BY col1)
SELECT * FROM table_name PIVOT (AVG(val) FOR col1 IN ('a', 'b'))
```

### DDL Statements
```sql
CREATE TABLE users (
    id INT PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    email VARCHAR(255) UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)

CREATE INDEX idx_name ON users(name)

ALTER TABLE users ADD COLUMN age INT
ALTER TABLE users DROP COLUMN age
ALTER TABLE users RENAME TO customers

DROP TABLE users
DROP INDEX idx_name

TRUNCATE TABLE users
```

### DML Statements
```sql
INSERT INTO users (name, email) VALUES ('John', 'john@example.com')
INSERT INTO users VALUES ('John', 'john@example.com')
INSERT INTO users SELECT * FROM new_users

UPDATE users SET name = 'Jane' WHERE id = 1
UPDATE users SET name = 'Jane', email = 'jane@example.com' WHERE id = 1

DELETE FROM users WHERE id = 1
DELETE FROM users

MERGE INTO target USING source ON target.id = source.id
    WHEN MATCHED THEN UPDATE SET target.name = source.name
    WHEN NOT MATCHED THEN INSERT (id, name) VALUES (source.id, source.name)
```

### Transaction Control
```sql
BEGIN TRANSACTION
COMMIT
ROLLBACK
SAVEPOINT my_savepoint
RELEASE SAVEPOINT my_savepoint
```

### Other Statements
```sql
-- PostgreSQL COPY
COPY users FROM '/data/users.csv' WITH (FORMAT CSV)

-- MySQL SHOW
SHOW TABLES
SHOW CREATE TABLE users

-- Analyze and Explain
ANALYZE TABLE users
EXPLAIN SELECT * FROM users
```

## Error Handling

The parser provides detailed error messages with source locations:

```go
sql := "SELECT * FRO users"  // Typo: FRO instead of FROM

_, err := parser.ParseSQL(dialect, sql)
if err != nil {
    // Error: sql ParserError at Line: 1, Column: 10: Expected: FROM, found: FRO
    fmt.Println(err)
}
```

## Testing

### Run Tests
```bash
# Run all tests
cd go
go test ./...

# Run specific package tests
cd go/tests
go test -v

# Run with race detector
go test -race ./...
```

### Run Fuzz Tests
```bash
cd go/fuzz

# Run for 1 hour
go test -fuzz=FuzzParser -fuzztime=1h

# Run specific dialect fuzzer
go test -fuzz=FuzzPostgreSQL -fuzztime=30m
```

## Project Structure

The project is organized into focused packages:

- `token/` - SQL tokenizer and lexer (merged from token/ and tokenizer/)
  - `token.go` - Token types and definitions
  - `lexer.go` - Tokenization logic
  - `keywords.go` - SQL keyword definitions
  - `position.go` - Source location tracking (Span/Position)
  - `state.go` - Lexer state management

- `ast/` - Abstract syntax tree definitions (consolidated structure)
  - `node.go` - Base AST node types
  - `expr_all.go` - Expression types (consolidated)
  - `expr_funcs.go` - Function expressions
  - `operators_all.go` - Operators
  - `query_all.go` - Query types (SELECT, CTEs, etc.)
  - `statement_all.go` - Statement types (DML, DDL, DCL)
  - `types_all.go` - Data types
  - `value.go` - Value/literal types
  - `expr/`, `query/`, `statement/`, `datatype/` - Supporting sub-packages

- `parser/` - Parser implementation (split into focused files)
  - `parser.go` - Main parser entry points
  - `core.go` - Core parsing infrastructure
  - `prefix.go`, `infix.go`, `postfix.go` - Expression parsers (Pratt parsing)
  - `special.go` - Special expression parsers
  - `helpers.go` - Parser utility methods
  - `groupings.go` - GROUPING SETS, CUBE, ROLLUP
  - `create.go`, `alter.go`, `drop.go`, `truncate.go` - DDL parsing
  - `select.go`, `query.go`, `copy.go`, `show.go` - Query parsing
  - `dml.go`, `merge.go`, `transaction.go` - Statement parsing
  - `describe.go`, `misc.go`, `prepared.go` - Other statements

- `dialects/` - SQL dialect implementations
  - `dialect.go` - Core dialect interface and capability interfaces
  - `capabilities.go` - Dialect capability checking helpers
  - Individual dialect packages: `postgresql/`, `mysql/`, `sqlite/`, `bigquery/`, 
    `snowflake/`, `duckdb/`, `clickhouse/`, `hive/`, `mssql/`, `redshift/`, 
    `databricks/`, `oracle/`, `ansi/`, `generic/`

- `parseriface/` - Parser interface definitions (resolves circular imports)
  - Defines ParserInterface for use by dialects and parser

- `errors/` - Error types and handling

- `tests/` - Test suite organized by functionality
  - `parse_test.go` - Core parsing tests
  - `expr_test.go`, `func_test.go` - Expression tests
  - `query/` - SELECT, JOIN, CTE tests
  - `ddl/` - CREATE, ALTER, DROP, TRUNCATE tests
  - `dml/` - INSERT, UPDATE, DELETE tests
  - Dialect-specific test folders: `postgres/`, `mysql/`, `bigquery/`, etc.
  - `regression/` - TPC-H regression tests

- `fuzz/` - Fuzz testing for robustness

- `examples/` - Usage examples

## Dialects

All major SQL dialects are supported:

| Dialect | Package | Notes |
|---------|---------|-------|
| Generic | `dialects/generic` | Most permissive, supports union of all features |
| PostgreSQL | `dialects/postgresql` | Arrays, JSON/JSONB, geometric types |
| MySQL | `dialects/mysql` | Backtick identifiers, LIMIT comma syntax |
| SQLite | `dialects/sqlite` | Lightweight, supports both backticks and double quotes |
| BigQuery | `dialects/bigquery` | STRUCT, ARRAY, QUALIFY, PIVOT |
| Snowflake | `dialects/snowflake` | Semi-structured data, stages, pipes |
| DuckDB | `dialects/duckdb` | List types, lambda functions |
| ClickHouse | `dialects/clickhouse` | Nested types, materialized views |
| Hive | `dialects/hive` | LATERAL VIEW, TRANSFORM |
| MSSQL | `dialects/mssql` | TOP, square bracket identifiers |
| Redshift | `dialects/redshift` | Based on PostgreSQL, COPY/UNLOAD |
| Databricks | `dialects/databricks` | Spark SQL + Delta Lake |
| Oracle | `dialects/oracle` | CONNECT BY, (+) outer join |
| ANSI | `dialects/ansi` | Strict ANSI SQL:2011 |

## Performance

The parser is designed for performance:
- Single-pass tokenization
- Pratt parsing for expressions (linear time)
- Zero-allocation where possible
- Efficient string handling

## License

Apache License 2.0 - See [LICENSE](../LICENSE) for details.

## Credits

This is a Go port of the excellent [sqlparser-rs](https://github.com/sqlparser-rs/sqlparser-rs) Rust library, originally created by the Apache DataFusion project.

## Contributing

Contributions are welcome! Please see [GOLANG.md](../GOLANG.md) for the implementation details and architecture.

## Support

For bugs and feature requests, please open an issue on GitHub.
