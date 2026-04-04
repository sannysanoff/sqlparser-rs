# AST to JSON Example

This example demonstrates how to parse SQL from stdin and output the Abstract Syntax Tree (AST) as JSON.

## Building

### Debug Build
```bash
cargo build --features serde,serde_json --example ast_to_json
```

### Release Build (Optimized)
```bash
cargo build --release --features serde,serde_json --example ast_to_json
```

The binary will be located at:
- Debug: `target/debug/examples/ast_to_json`
- Release: `target/release/examples/ast_to_json`

## Usage

### Basic Usage
```bash
echo "SELECT * FROM users" | cargo run --features serde,serde_json --example ast_to_json -- generic
```

### Using the Compiled Binary
```bash
echo "SELECT * FROM users" | ./target/release/examples/ast_to_json postgres
```

### Reading from a File
```bash
cat query.sql | ./target/release/examples/ast_to_json mysql
```

### Piping SQL
```bash
./target/release/examples/ast_to_json generic < query.sql
```

## Supported Dialects

- `ansi` - ANSI SQL
- `bigquery` - Google BigQuery
- `clickhouse` - ClickHouse
- `databricks` - Databricks SQL
- `duckdb` - DuckDB
- `generic` - Generic SQL (default-like behavior)
- `hive` - Apache Hive
- `mysql` - MySQL
- `mssql` - Microsoft SQL Server
- `oracle` - Oracle Database
- `postgres` or `postgresql` - PostgreSQL
- `redshift` - Amazon Redshift
- `snowflake` - Snowflake
- `sqlite` - SQLite

## Examples

### Parse a SELECT statement with PostgreSQL dialect
```bash
echo "SELECT id, name FROM users WHERE age > 18" | ./target/release/examples/ast_to_json postgres
```

### Parse a complex query with MySQL dialect
```bash
echo "SELECT u.id, COUNT(*) as total FROM users u JOIN orders o ON u.id = o.user_id GROUP BY u.id" | ./target/release/examples/ast_to_json mysql
```

### Parse from a file
```bash
./target/release/examples/ast_to_json postgres < my_query.sql > ast_output.json
```

## Output Format

The output is a JSON array containing one or more statement objects. Each statement is fully serialized with all its properties, including:

- Statement type (Query, Insert, Update, Delete, etc.)
- All expressions and identifiers
- Source span information (line and column numbers)
- All SQL syntax elements

This detailed JSON output can be used for:
- SQL analysis tools
- Query visualization
- AST manipulation
- Educational purposes
- Integration with other tools

## Error Handling

If the SQL cannot be parsed, the program will:
1. Print an error message to stderr
2. Exit with code 1

Example:
```bash
echo "SELECT * FORM users" | ./target/release/examples/ast_to_json generic
# Output: Parse error: sql parser error: Expected FROM, found: users at Line: 1, Column: 10
```
