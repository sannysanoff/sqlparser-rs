# Statement Parsers

This package contains all SQL statement-specific parsing methods transpiled from sqlparser-rs.

## Files

- **doc.go** - Package documentation and common imports
- **package.go** - Package-level types and interfaces
- **query.go** - SELECT, Query, WITH parsing methods
- **dml.go** - INSERT, UPDATE, DELETE parsing methods
- **merge.go** - MERGE statement parsing (complete implementation)
- **ddl.go** - CREATE statements (TABLE, VIEW, INDEX, etc.)
- **alter.go** - ALTER statements (TABLE, VIEW, INDEX, ROLE, USER, etc.)
- **transaction.go** - BEGIN, COMMIT, ROLLBACK, SAVEPOINT
- **other.go** - DROP, TRUNCATE, EXPLAIN, SET, SHOW, GRANT, REVOKE, COPY, DECLARE, FETCH, CLOSE, USE, CALL, ANALYZE

## Statement Coverage

### Query Statements
- `parseQuery()` - Full query with CTEs, ORDER BY, LIMIT, FOR clauses
- `parseSelect()` - Core SELECT with all clauses (WHERE, GROUP BY, HAVING, etc.)
- `parseWith()` - Common Table Expression definitions
- `parseProjection()` - SELECT list parsing

### DML Statements
- `ParseInsert()` - INSERT with VALUES or subquery, PARTITION, RETURNING, ON CONFLICT
- `ParseUpdate()` - UPDATE with SET, WHERE, FROM, RETURNING
- `ParseDelete()` - DELETE with WHERE, USING, RETURNING, ORDER BY, LIMIT
- `ParseMerge()` - MERGE with MATCHED/NOT MATCHED clauses, UPDATE/DELETE/INSERT actions

### DDL Statements
- `ParseCreate()` - CREATE TABLE, VIEW, INDEX, FUNCTION, ROLE, DATABASE, SCHEMA
- `ParseAlter()` - ALTER TABLE (ADD/DROP/RENAME columns), ALTER VIEW, ALTER INDEX, ALTER ROLE, ALTER USER
- `ParseDrop()` - DROP TABLE, VIEW, INDEX, FUNCTION, ROLE, DATABASE, SCHEMA
- `ParseTruncate()` - TRUNCATE TABLE

### Transaction Statements
- `ParseStartTransaction()` / `ParseBegin()` - Transaction start
- `ParseCommit()` - Transaction commit
- `ParseRollback()` - Transaction rollback
- `ParseSavepoint()` - Savepoint creation
- `ParseRelease()` - Release savepoint

### Other Statements
- `ParseExplain()` - EXPLAIN / DESCRIBE
- `ParseAnalyze()` - ANALYZE table
- `ParseSet()` - SET variables / SET TIME ZONE / SET ROLE
- `ParseShow()` - SHOW TABLES, COLUMNS, DATABASES, VARIABLES
- `ParseGrant()` - GRANT privileges
- `ParseRevoke()` - REVOKE privileges
- `ParseCopy()` - PostgreSQL COPY
- `ParseDeclare()` - DECLARE cursor
- `ParseFetch()` - FETCH from cursor
- `ParseClose()` - CLOSE cursor
- `ParseUse()` - USE database
- `ParseCall()` - CALL procedure

## Usage

```go
import (
    "github.com/user/sqlparser-parser"
    "github.com/user/sqlparser-dialects/generic"
)

// Create a parser
dialect := generic.New()
p := parser.New(dialect)

// Parse SQL
stmts, err := p.TryWithSQL("SELECT * FROM users WHERE id = 1").ParseStatements()
```

## Implementation Notes

- All parsers follow the Rust implementation from sqlparser-rs
- Dialect-specific syntax is handled via dialect hooks
- Proper error messages with location information
- Support for optional clauses and keyword variations
- Recursion limits to prevent stack overflow

## License

Apache License 2.0 - See LICENSE file for details.
