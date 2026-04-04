# CREATE TABLE Implementation Summary

## Current Status
The CREATE TABLE implementation is STUBBED OUT in `/Users/san/Fun/sqlparser-rs/go/parser/ddl.go` at line 89-95:

```go
func parseCreateTable(p *Parser, orReplace, temporary bool) (ast.Statement, error) {
	if _, err := p.ExpectKeyword("TABLE"); err != nil {
		return nil, err
	}
	return nil, p.expectedRef("CREATE TABLE not yet fully implemented", p.PeekTokenRef())
}
```

This causes 42+ test failures with the error: "CREATE TABLE not yet fully implemented"

## Required Implementation

### 1. Main parseCreateTable Function
Replace the stub with a full implementation that:
- Parses CREATE [OR REPLACE] [TEMPORARY] [EXTERNAL] TABLE
- Handles IF NOT EXISTS clause
- Parses table name
- Supports LIKE clause (CREATE TABLE new LIKE old)
- Supports CLONE clause (Snowflake syntax)
- Parses column definitions in parentheses
- Parses table constraints
- Handles WITH options (table properties)
- Supports AS SELECT clause (CREATE TABLE AS SELECT)

### 2. Required Helper Functions to Add

#### parseCreateTableColumns(p *Parser)
- Parses column definitions and table constraints within parentheses
- Returns ([]*expr.ColumnDef, []*expr.TableConstraint, error)
- Handles trailing commas (if dialect supports it)
- Differentiates between column definitions and table constraints

#### isTableConstraint(p *Parser) bool
- Detects if next token is a table constraint keyword:
  - CONSTRAINT
  - PRIMARY KEY
  - FOREIGN KEY
  - UNIQUE
  - CHECK

#### parseTableConstraint(p *Parser, exprParser)
- Parses named constraints: CONSTRAINT <name> <constraint_type>
- Supports PRIMARY KEY (col1, col2, ...)
- Supports FOREIGN KEY (cols) REFERENCES table (cols) [ON DELETE/UPDATE action]
- Supports UNIQUE (col1, col2, ...)
- Supports CHECK (<expression>)

#### parseColumnDef(p *Parser, exprParser)
- Parses column name and data type
- Supports column constraints:
  - NULL / NOT NULL
  - DEFAULT <value>
  - PRIMARY KEY
  - UNIQUE
  - REFERENCES <table>[(cols)] [ON DELETE/UPDATE action]
  - CHECK (<expression>)
  - CONSTRAINT <name> (for named constraints)

#### parseReferentialAction(p *Parser)
- Returns action type for ON DELETE/ON UPDATE:
  - CASCADE
  - RESTRICT
  - SET NULL
  - SET DEFAULT
  - NO ACTION

### 3. AST Types Required in expr/ddl.go

The following constraint types need to be defined:

```go
// PrimaryKeyConstraint represents a PRIMARY KEY constraint
type PrimaryKeyConstraint struct {
    Name    *ast.Ident
    Columns []*ast.Ident
}

// UniqueConstraint represents a UNIQUE constraint  
type UniqueConstraint struct {
    Name    *ast.Ident
    Columns []*ast.Ident
}

// ForeignKeyConstraint represents a FOREIGN KEY constraint
type ForeignKeyConstraint struct {
    Name            *ast.Ident
    Columns         []*ast.Ident
    ForeignTable    *ast.ObjectName
    ReferredColumns []*ast.Ident
    OnDelete        ReferentialAction
    OnUpdate        ReferentialAction
}

// CheckConstraint represents a CHECK constraint
type CheckConstraint struct {
    Name *ast.Ident
    Expr Expr
}
```

### 4. Known Issues to Fix

#### Import Cycle Issue
The datatype package imports expr, and attempts to use datatype types in the parser cause import cycles. The solution is to:
- Either fix the datatype package to properly implement ast.DataType
- Or avoid using concrete datatype types in the parser and use interface types

#### DataType Interface Implementation
The types in ast/datatype/datatype.go don't implement ast.DataType because they're missing the IsDataType() method. They need to either:
- Embed ast.DataTypeBase
- Or add the IsDataType() method manually

### 5. Test Coverage

Failing tests to fix:
- TestParseCreateTableSelect
- TestParseCreateTableWithBitTypes
- TestParseCreateTableWithEnumTypes
- TestParseCreateTableLike
- TestParseCreateTableLikeWithDefaults
- TestParseCreateTable
- TestParseCreateTableWithConstraintCharacteristics
- TestParseCreateTableColumnConstraintCharacteristics
- TestParseCreateTableHiveArray (tokenizer issue)
- TestParseCreateTableAs
- TestParseCreateTableAsTable
- TestParseCreateTableOnCluster
- TestParseCreateOrReplaceTable
- TestParseCreateTableWithOnDeleteOnUpdateInAnyOrder
- TestParseCreateTableWithOptions
- TestParseCreateTableClone
- TestParseCreateTableTrailingComma
- TestParseCreateExternalTable
- TestParseCreateOrReplaceExternalTable
- TestParseCreateExternalTableLowercase
- TestParseCreateTableHiveFormatsNoneWhenNoOptions

### 6. Implementation Priority

1. **Critical (for basic CREATE TABLE):**
   - parseCreateTable with column definitions
   - parseColumnDef with data types
   - parseCreateTableColumns

2. **High Priority:**
   - Table constraints (PRIMARY KEY, FOREIGN KEY, UNIQUE, CHECK)
   - Column constraints (NOT NULL, DEFAULT, REFERENCES)
   - IF NOT EXISTS, OR REPLACE, TEMPORARY

3. **Medium Priority:**
   - LIKE clause
   - WITH options
   - EXTERNAL tables

4. **Lower Priority:**
   - CLONE clause (Snowflake)
   - AS SELECT
   - AS TABLE

## Files to Modify

1. `/Users/san/Fun/sqlparser-rs/go/parser/ddl.go`
   - Replace parseCreateTable stub
   - Add helper functions

2. `/Users/san/Fun/sqlparser-rs/go/ast/expr/ddl.go`
   - Add constraint type definitions

3. `/Users/san/Fun/sqlparser-rs/go/ast/datatype/datatype.go`
   - Fix DataType interface implementation

## Code Structure

The implementation should follow the pattern of parse_create_table in the Rust source:
`src/parser/mod.rs` starting at line 8339

Key points from Rust implementation:
- Uses dialect-specific parsing for certain features
- Supports extensive PostgreSQL-specific syntax
- Handles complex constraint characteristics
- Supports multiple table options (PARTITION BY, CLUSTER BY, etc.)

