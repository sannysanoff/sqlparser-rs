# Go SQL Parser Architecture

This document describes the internal architecture and design patterns of the sqlparser-go project.

## Table of Contents

1. [Overview](#overview)
2. [Package Organization](#package-organization)
3. [Parser Design](#parser-design)
4. [Tokenization](#tokenization)
5. [AST Structure](#ast-structure)
6. [Dialect System](#dialect-system)
7. [Testing Architecture](#testing-architecture)
8. [Key Design Patterns](#key-design-patterns)

---

## Overview

This is a Go port of the Rust sqlparser-rs library. The parser uses:

- **Pratt Parsing** (Top-Down Operator Precedence) for expressions
- **Recursive Descent** for statements
- **Interface-based** dialect system for SQL variations
- **Visitor-friendly** AST design

The codebase is organized to minimize dependencies and circular imports while maintaining clear separation of concerns.

---

## Package Organization

### Core Packages

```
sqlparser/
├── token/           # Tokenization and lexical analysis
├── ast/             # Abstract Syntax Tree definitions
├── parser/          # Parser implementation
├── dialects/        # SQL dialect implementations
├── parseriface/     # Shared parser interfaces
└── errors/          # Error types
```

### Package Dependencies

```
errors (no dependencies)
    ↑
token → ast
    ↑       ↘
    ↑    parseriface ← dialects
    ↑       ↑
    parser ─┘
```

**Key Design Decision:** The `parseriface` package resolves circular imports between `parser` and `dialects`. Both can depend on `parseriface` without depending on each other.

---

## Parser Design

### Pratt Parser Architecture

The parser uses Pratt parsing (Top-Down Operator Precedence) for expressions, which provides:

- **Linear time complexity** O(n)
- **Correct operator precedence** without complex grammar rules
- **Easy extensibility** for new operators

### Core Parser Files

| File | Purpose |
|------|---------|
| `parser.go` | Entry points (`ParseSQL`, `New`), main parsing loop |
| `core.go` | `ParseExpr()`, `GetNextPrecedence()` - Pratt parser core |
| `prefix.go` | Prefix expression parsers (identifiers, literals, functions) |
| `infix.go` | Infix expression parsers (binary ops, `AND`, `OR`, comparisons) |
| `postfix.go` | Postfix parsers (array subscripts, `COLLATE`) |
| `special.go` | Special expressions (window functions, aggregates, subqueries) |
| `helpers.go` | Utility methods (comma-separated lists, optional aliases) |

### Expression Parsing Flow

```
ParseExpr()
    ↓
ParsePrefix() ──→ parseIdentifier(), parseNumber(), parseFunction(), etc.
    ↓
GetNextPrecedence() ──→ Check for operators at current position
    ↓
while precedence < next_precedence:
    ParseInfix(left, precedence) ──→ parseBinaryOp(), parseIn(), etc.
    left = result
    ↓
return left
```

### Statement Parsing

Statements use recursive descent parsing:

```
parseStatement()
    ↓
parseSelect(), parseInsert(), parseUpdate(), parseDelete()
    ↓
parseCreateTable(), parseAlterTable(), parseDropTable()
    ↓
parseCopy(), parseShow(), parseAnalyze()
```

---

## Tokenization

### Token Package Structure

The `token/` package was consolidated from the original `token/` and `tokenizer/` packages:

```
token/
├── token.go        # Token types, Token interface, Tokenizer
├── lexer.go        # Tokenization logic (moved from tokenizer/)
├── keywords.go     # SQL keywords (~83KB of definitions)
├── position.go     # Span, Position (source location tracking)
└── state.go        # Lexer state management
```

### Token Types

Tokens implement the `Token` interface:

```go
type Token interface {
    Type() TokenType
    Value() string
    Span() Span
    String() string
}
```

Special token types:
- `Word` - Identifiers, keywords
- `Number` - Numeric literals
- `SingleQuotedString` - String literals
- `Whitespace` - Spaces, comments (preserved)

---

## AST Structure

### AST Organization

The AST was consolidated during refactoring:

```
ast/
├── node.go            # Base Node interface
├── expr_all.go        # Expression types (consolidated)
├── expr_funcs.go      # Function expressions
├── operators_all.go   # Binary/unary operators
├── query_all.go       # Query types (SELECT, CTEs, subqueries)
├── statement_all.go   # Statement types (DML, DDL, DCL, TCL)
├── types_all.go       # Data types
├── value.go           # Literal values
├── ident.go           # Identifiers
└── Supporting packages: expr/, query/, statement/, datatype/
```

### Node Interface

All AST nodes implement:

```go
type Node interface {
    String() string           // SQL regeneration
    fmt.Stringer
}
```

### Statement Interface

```go
type Statement interface {
    Node
    statementNode()  // Marker method
}
```

### Expression Interface

```go
type Expression interface {
    Node
    expressionNode()  // Marker method
}
```

### AST Traversal

To traverse the AST, use type assertions or switches:

```go
func analyzeStatement(stmt ast.Statement) {
    switch s := stmt.(type) {
    case *statement.SelectStmt:
        analyzeQuery(s.Query)
    case *statement.InsertStmt:
        analyzeInsert(s)
    case *statement.UpdateStmt:
        analyzeUpdate(s)
    }
}
```

---

## Dialect System

### Dialect Interfaces

The dialect system uses interface segregation to avoid 100+ method monolithic interfaces:

```go
// Core dialect identification
type Dialect interface {
    Dialect() string
}

// Identifier handling
type IdentifierDialect interface {
    Dialect
    IsIdentifierStart(ch rune) bool
    IsIdentifierPart(ch rune) bool
    IdentifierQuoteStyle(identifier string) *rune
}

// String literal handling
type StringLiteralDialect interface {
    Dialect
    SupportsStringLiteralBackslashEscape() bool
    SupportsUnicodeStringLiteral() bool
}

// SELECT clause support
type SelectDialect interface {
    Dialect
    SupportsSelectWildcardExcept() bool
    SupportsFromFirstSelect() bool
}

// DDL support
type DDLDialect interface {
    Dialect
    SupportsCreateTableSelect() bool
    SupportsAlterColumnTypeUsing() bool
}
```

### Dialect Capability Checking

Use the helper functions in `dialects/capabilities.go`:

```go
if dialects.SupportsSelectWildcard(d) {
    // Handle EXCEPT/REPLACE clauses
}

if dialects.SupportsStringLiteralBackslashEscape(d) {
    // Handle backslash escapes
}
```

### Adding a New Dialect

1. Create a new file in `dialects/<name>/<name>.go`
2. Implement the required interfaces
3. Add tests in `tests/<name>/`

Example minimal dialect:

```go
package mydialect

type MyDialect struct{}

func NewMyDialect() *MyDialect {
    return &MyDialect{}
}

func (d *MyDialect) Dialect() string {
    return "mydialect"
}

// Implement required capability interfaces...
```

---

## Testing Architecture

### Test Organization

Tests are organized by functionality rather than arbitrary batches:

```
tests/
├── parse_test.go           # Core parsing tests
├── expr_test.go            # Expression parsing
├── func_test.go            # Function parsing
├── transaction_test.go     # Transaction statements
├── table_test.go           # Table operations
├── other_test.go           # Miscellaneous
├── utils/
│   └── test_helpers.go     # Shared test utilities
├── query/                  # Query tests
│   ├── select_test.go
│   ├── join_test.go
│   ├── cte_test.go
│   └── setops_test.go
├── ddl/                    # DDL tests
│   ├── alter_test.go
│   └── truncate_test.go
├── dml/                    # DML tests
│   ├── insert_test.go
│   ├── update_test.go
│   └── delete_test.go
└── <dialect>/              # Dialect-specific tests
    ├── mysql_test.go
    ├── postgres_test.go
    └── ...
```

### Test Helpers

The `tests/utils/test_helpers.go` provides:

```go
// ParseSQL parses SQL and returns statements or fails the test
func ParseSQL(t *testing.T, dialect dialects.Dialect, sql string) []ast.Statement

// ParseOne parses a single statement
func ParseOne(t *testing.T, dialect dialects.Dialect, sql string) ast.Statement

// AssertParseable checks that SQL parses without error
func AssertParseable(t *testing.T, dialect dialects.Dialect, sql string)
```

### Running Tests

```bash
# All tests
go test ./...

# Specific package
go test ./tests/...

# With race detector
go test -race ./...

# Verbose output
go test -v ./tests/query/...
```

---

## Key Design Patterns

### 1. Interface Segregation

Split large interfaces into focused ones (dialect capability interfaces).

### 2. Delegation Pattern

Parser methods often delegate to standalone functions:

```go
func (p *Parser) parseInsert() (ast.Statement, error) {
    return ParseInsert(p)
}
```

This allows:
- Easier testing of individual parsers
- Reuse across different contexts
- Smaller method sets on Parser struct

### 3. Marker Interfaces

Use empty methods for type safety:

```go
type Expression interface {
    Node
    expressionNode()  // Prevents accidental implementation
}
```

### 4. Option Pattern

Parser options use functional options:

```go
parser.New(dialect,
    parser.WithRecursionLimit(100),
    parser.WithUnescape(true),
)
```

### 5. Error Accumulation

Some parsers collect errors and continue:

```go
type ParseResult struct {
    Statements []ast.Statement
    Errors     []error
}
```

---

## Adding New Features

### Adding a New Expression Type

1. Define the type in `ast/expr_all.go` or `ast/expr_funcs.go`
2. Add prefix/infix parser in `parser/prefix.go` or `parser/infix.go`
3. Add precedence in `parser/core.go`
4. Add tests in `tests/expr_test.go`

### Adding a New Statement Type

1. Define the statement in `ast/statement_all.go`
2. Add parser in appropriate file (`parser/dml.go`, `parser/ddl.go`, etc.)
3. Hook into `parseStatement()` in `parser/parser.go`
4. Add tests in appropriate test file

### Adding a New Dialect Feature

1. Add capability method to appropriate dialect interface
2. Implement in dialect structs
3. Use capability check in parser
4. Add dialect-specific tests

---

## Performance Considerations

### Current Optimizations

- **Single-pass tokenization**: Tokens are produced on-demand
- **Pratt parsing**: Linear time expression parsing
- **Pre-allocated slices**: Common pattern in parser helpers
- **String interning**: Keywords and common identifiers

### Memory Considerations

- AST nodes are allocated on the heap
- Large SQL strings are referenced, not copied
- Whitespace is preserved but can be skipped during parsing

---

## Common Tasks

### Regenerating SQL from AST

```go
stmt, _ := parser.ParseSQL(dialect, sql)
regeneratedSQL := stmt[0].String()
```

### Getting Source Locations

```go
token.Span()  // Returns Span with line, column
```

### Customizing Parser Behavior

```go
p := parser.New(dialect,
    parser.WithRecursionLimit(1000),
    parser.WithUnescape(true),
)
statements, err := p.Parse(sql)
```

---

## Debugging Tips

1. **Enable verbose logging**: Add log statements to parser methods
2. **Use `go test -v`**: See individual test output
3. **Check token stream**: Use `tokenizer.New(dialect).Tokenize(sql)` to see tokens
4. **AST inspection**: Use `fmt.Printf("%+v\n", node)` to see AST structure

---

## Contributing

When contributing:

1. Follow existing code organization patterns
2. Add tests for new functionality
3. Update this documentation if architecture changes
4. Run `go fmt`, `go vet`, and all tests before submitting

---

## License

Apache License 2.0 - See [LICENSE](../LICENSE) for details.
