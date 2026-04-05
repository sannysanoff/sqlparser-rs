# Code Organization Refactoring Plan

This document outlines a comprehensive plan to address code organization issues in the sqlparser-go project.

## Overview

The current codebase has several structural issues that impact maintainability, readability, and developer productivity. This refactoring plan is designed to be implemented incrementally to minimize disruption.

---

## Phase 1: Package Structure Consolidation

### 1.1 Merge Token Packages

**Problem:** The `token/` and `tokenizer/` packages are confusingly separated with overlapping responsibilities.

**Current State:**

- `token/` contains only `keywords.go` (83KB of keywords)
- `tokenizer/` contains token definitions and tokenization logic

**Solution:**

Merge into a single `token/` package with clear sub-packages:

```
token/
├── token.go           # Core Token interface and types
├── keywords.go        # Keyword definitions
├── lexer.go           # Tokenization logic (moved from tokenizer/)
├── lexer_test.go
└── position.go        # Span/Position types (moved from span/)
```

**Implementation Steps:**

1. Move `tokenizer/tokenizer.go` → `token/lexer.go`
2. Move `tokenizer/tokens.go` → `token/token.go`
3. Move `tokenizer/state.go` → `token/state.go`
4. Merge `span/` package into `token/position.go`
5. Update all imports across the codebase
6. Run tests to verify functionality

**Estimated Effort:** 2-3 hours

---

### 1.2 Consolidate AST Packages

**Problem:** The AST is split into too many sub-packages, creating import overhead and cognitive complexity.

**Current State:**

```
ast/
├── datatype/
├── expr/
├── operator/
├── query/
├── statement/
├── value/
├── expr.go
├── query.go
├── statement.go
└── node.go
```

**Solution:**

Flatten into fewer, more cohesive packages:

```
ast/
├── node.go            # Base Node types
├── expr.go            # Expression types (merge expr/*.go)
├── expr_ops.go        # Expression operators (from expr/operators.go)
├── expr_funcs.go      # Function expressions (from expr/functions.go)
├── statement.go       # Statement types (merge statement/*.go)
├── query.go           # Query types (merge query/*.go)
├── types.go           # Data types (from datatype/)
└── values.go          # Value types (from value/)
```

**Implementation Steps:**

1. Create `ast/expr.go` by merging:
   - `ast/expr.go` (existing)
   - `ast/expr/basic.go`
   - `ast/expr/complex.go`
   - `ast/expr/conditional.go`
   - `ast/expr/subqueries.go`

2. Create `ast/expr_ops.go` from `ast/expr/operators.go` and `ast/operator/operator.go`

3. Create `ast/expr_funcs.go` from `ast/expr/functions.go`

4. Create `ast/statement.go` by merging:
   - `ast/statement/statement.go`
   - `ast/statement/dml.go`
   - `ast/statement/dcl.go`
   - `ast/statement/action.go`
   - `ast/statement/misc.go`
   - `ast/statement/ddl.go`

5. Create `ast/query.go` by merging:
   - `ast/query/query.go`
   - `ast/query/clauses.go`
   - `ast/query/table.go`
   - `ast/query/setops.go`
   - `ast/query/window.go`
   - `ast/query/other.go`

6. Create `ast/types.go` from `ast/datatype/datatype.go`

7. Create `ast/values.go` from `ast/value.go`

8. Delete old sub-packages after verification

**Estimated Effort:** 4-6 hours

---

## Phase 2: Split Monolithic Files

### 2.1 Parser Package Refactoring

**Problem:** Several parser files exceed 1,000 lines with unrelated functionality mixed together.

**Files to Split:**

| File | Lines | Split Into |
|------|-------|------------|
| `parser/other.go` | 2,444 | `parser/copy.go`, `parser/show.go`, `parser/analyze.go`, `parser/describe.go`, `parser/misc.go` |
| `parser/ddl.go` | 1,425 | `parser/create.go`, `parser/alter.go`, `parser/drop.go`, `parser/truncate.go` |
| `parser/query.go` | 2,064 | `parser/select.go`, `parser/cte.go`, `parser/subquery.go` |
| `parser/dml.go` | 699 | Already reasonable size, but consider splitting inserts/updates |
| `parser/merge.go` | 471 | Keep as-is (focused) |

**Implementation Steps for `parser/other.go`:**

1. Create `parser/copy.go` for COPY statement parsing
2. Create `parser/show.go` for SHOW statement parsing
3. Create `parser/analyze.go` for ANALYZE statement parsing
4. Create `parser/describe.go` for DESCRIBE/DESC statement parsing
5. Move remaining miscellaneous parsers to `parser/misc.go`
6. Update `parser/parser.go` to use new files

**Implementation Steps for `parser/ddl.go`:**

1. Create `parser/create.go`:
   - CREATE TABLE
   - CREATE INDEX
   - CREATE VIEW
   - CREATE DATABASE/SCHEMA

2. Create `parser/alter.go`:
   - ALTER TABLE
   - ALTER INDEX
   - ALTER VIEW

3. Create `parser/drop.go`:
   - DROP TABLE
   - DROP INDEX
   - DROP VIEW
   - DROP DATABASE/SCHEMA

4. Create `parser/truncate.go`:
   - TRUNCATE TABLE

5. Keep common DDL utilities in `parser/ddl.go` (reduced size)

**Estimated Effort:** 6-8 hours

---

### 2.2 AST File Refactoring

**After Phase 1.2 consolidation, split large merged files:**

| File (After Merge) | Target Size | Split Strategy |
|-------------------|-------------|----------------|
| `ast/statement.go` | 2,200 lines | Split by statement category |
| `ast/expr.go` | ~800 lines | Keep consolidated |
| `ast/query.go` | ~1,600 lines | Split into `query_select.go`, `query_cte.go` |
| `ast/types.go` | 1,964 lines | Split by type category |

**Implementation:**

1. `ast/statement.go` → Split into:
   - `ast/stmt_dml.go` (INSERT, UPDATE, DELETE)
   - `ast/stmt_ddl.go` (CREATE, ALTER, DROP)
   - `ast/stmt_dcl.go` (GRANT, REVOKE)
   - `ast/stmt_txn.go` (BEGIN, COMMIT, ROLLBACK)

2. `ast/types.go` → Split into:
   - `ast/types_numeric.go` (INT, FLOAT, DECIMAL)
   - `ast/types_text.go` (VARCHAR, CHAR, TEXT)
   - `ast/types_datetime.go` (DATE, TIME, TIMESTAMP)
   - `ast/types_complex.go` (ARRAY, JSON, STRUCT)

**Estimated Effort:** 4-5 hours

---

## Phase 3: Interface Consolidation

### 3.1 Resolve ParserAccessor/ParserInterface Duplication

**Problem:** Two nearly identical interfaces exist to avoid circular imports:

```go
// parser/core.go
interface ParserInterface { ... 35 methods ... }

// dialects/dialect.go
interface ParserAccessor { ... 35 methods ... }
```

**Root Cause:** The `dialects` package needs to reference parser functionality, and the `parser` package needs dialects.

**Solution:** Create a `parseriface` package that both can depend on:

```
parseriface/
├── parser.go          # Core parser interface (35 methods)
├── state.go           # ParserState, ParserOptions
└── token.go           # Token access methods
```

**Implementation Steps:**

1. Create new package `parseriface/`
2. Move common interfaces from `parser/core.go` and `dialects/dialect.go`
3. Update `parser` package to implement `parseriface.Parser`
4. Update `dialects` package to use `parseriface.Parser`
5. Remove duplicate interface definitions
6. Run full test suite

**Estimated Effort:** 3-4 hours

---

### 3.2 Break Up Monolithic Dialect Interface

**Problem:** The `Dialect` interface has 100+ methods, violating Interface Segregation Principle.

**Current Interface:**

```go
type Dialect interface {
    // Identifier Handling (5 methods)
    // Reserved Keywords (6 methods)
    // String Literals (8 methods)
    // Aggregations (5 methods)
    // GROUP BY (2 methods)
    // JOIN Support (3 methods)
    // ... 70+ more methods
}
```

**Solution:** Split into focused interfaces:

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
    IsDelimitedIdentifierStart(ch rune) bool
    IdentifierQuoteStyle(identifier string) *rune
}

// String literal handling
type StringLiteralDialect interface {
    Dialect
    SupportsStringLiteralBackslashEscape() bool
    SupportsUnicodeStringLiteral() bool
    SupportsTripleQuotedString() bool
    // ... other string methods
}

// SELECT clause support
type SelectDialect interface {
    Dialect
    SupportsSelectWildcardExcept() bool
    SupportsFromFirstSelect() bool
    SupportsLimitComma() bool
    // ... other SELECT methods
}

// DDL support
type DDLDialect interface {
    Dialect
    SupportsCreateTableSelect() bool
    SupportsAlterColumnTypeUsing() bool
    // ... other DDL methods
}

// Complete dialect (embeds all sub-interfaces)
type CompleteDialect interface {
    IdentifierDialect
    StringLiteralDialect
    SelectDialect
    DDLDialect
    // ... all other sub-interfaces
}
```

**Implementation Steps:**

1. Define new sub-interfaces in `dialects/` package
2. Create interface composition hierarchy
3. Update all dialect implementations (14 dialects) to use new structure
4. Update parser to check for specific capability interfaces
5. Add helper functions for capability checking:
   ```go
   func SupportsSelectWildcard(d Dialect) bool {
       if sd, ok := d.(SelectDialect); ok {
           return sd.SupportsSelectWildcardExcept()
       }
       return false
   }
   ```

**Estimated Effort:** 6-8 hours (due to 14 dialects)

---

## Phase 4: Test Reorganization

### 4.1 Consolidate Fragmented Test Files

**Problem:** Tests are split into arbitrary "batches" (`common_batch2_test.go` through `common_batch24_test.go`).

**Current State:**

```
tests/common/
├── common_test.go
├── common_batch2_test.go
├── common_batch3_test.go
...
├── common_batch24_test.go
```

**Solution:** Organize by functionality:

```
tests/
├── parse_test.go              # Core parsing tests
├── expr_test.go               # Expression parsing tests
├── query/
│   ├── select_test.go         # SELECT statement tests
│   ├── join_test.go            # JOIN parsing tests
│   └── cte_test.go            # CTE/subquery tests
├── ddl/
│   ├── create_test.go         # CREATE statement tests
│   ├── alter_test.go          # ALTER statement tests
│   └── drop_test.go           # DROP statement tests
├── dml/
│   ├── insert_test.go         # INSERT statement tests
│   ├── update_test.go         # UPDATE statement tests
│   └── delete_test.go         # DELETE statement tests
├── dialects/
│   ├── mysql_test.go          # MySQL-specific tests
│   ├── postgres_test.go       # PostgreSQL-specific tests
│   └── snowflake_test.go      # Snowflake-specific tests
└── utils/
    └── test_helpers.go         # Shared test utilities
```

**Implementation Steps:**

1. Create new directory structure
2. Analyze existing test files to categorize tests
3. Move tests to appropriate functional files
4. Update test helper utilities for new structure
5. Delete old batch files after verification
6. Ensure all tests still pass

**Migration Strategy:**

- Use `grep` to identify test functions by content
- Group tests by the SQL keywords they test
- Keep dialect-specific tests in dialect folders
- Maintain the TPC-H regression tests separately

**Estimated Effort:** 4-5 hours

---

## Phase 5: Naming Convention Standardization

### 5.1 Unify Parser Method Naming

**Problem:** Inconsistent naming: `ParseInsert` vs `parseDelete` vs `parseCreate`

**Solution:** Establish clear conventions:

**Exported Functions (Package-level):**
- `ParseSQL(dialect, sql)` - Main entry point
- `ParseStatements(p *Parser)` - Parse multiple statements
- `ParseStatement(p *Parser)` - Parse single statement
- `ParseExpression(p *Parser)` - Parse expression

**Internal Methods (Parser struct):**
- `parseSelect()` - Parse SELECT statement
- `parseInsert()` - Parse INSERT statement
- `parseCreateTable()` - Parse CREATE TABLE
- `parseAlterTable()` - Parse ALTER TABLE

**Implementation Steps:**

1. Audit all parser methods
2. Rename methods to follow convention
3. Update call sites
4. Update documentation

**Estimated Effort:** 2-3 hours

---

## Phase 6: Dependency Cleanup

### 6.1 Remove Unused Abstractions

**Problem:** Some delegation methods serve no purpose:

```go
func (p *Parser) parseInsert(tok tokenizer.TokenWithSpan) (ast.Statement, error) {
    return ParseInsert(p, tok)  // Just delegates!
}
```

**Solution:**

1. Identify and remove unnecessary delegation methods
2. Call exported functions directly where appropriate
3. Simplify the parser's public API

**Estimated Effort:** 1-2 hours

---

## Phase 7: Documentation Updates

### 7.1 Update Project Documentation

After all refactoring, update:

1. **README.md** - Update project structure description
2. **GOLANG.md** - Update architecture documentation
3. **Package-level doc.go files** - Add proper Go documentation

**Estimated Effort:** 2 hours

---

## Implementation Schedule

| Phase | Task | Estimated Time | Dependencies |
|-------|------|----------------|--------------|
| 1.1 | Merge token packages | 2-3 hours | None |
| 1.2 | Consolidate AST packages | 4-6 hours | None |
| 2.1 | Split parser files | 6-8 hours | None |
| 2.2 | Split AST files | 4-5 hours | 1.2 |
| 3.1 | Create parseriface package | 3-4 hours | None |
| 3.2 | Refactor dialect interfaces | 6-8 hours | 3.1 |
| 4.1 | Reorganize tests | 4-5 hours | 1.1, 1.2 |
| 5.1 | Standardize naming | 2-3 hours | 2.1 |
| 6.1 | Clean up abstractions | 1-2 hours | 2.1 |
| 7.1 | Update documentation | 2 hours | All |
| **Total** | | **34-46 hours** | |

---

## Recommended Execution Order

For minimal disruption, execute phases in this order:

### Sprint 1: Foundation (Week 1)
1. Phase 1.1 - Merge token packages
2. Phase 3.1 - Create parseriface package
3. Phase 5.1 - Standardize naming

### Sprint 2: Parser Restructure (Week 2)
1. Phase 2.1 - Split parser files
2. Phase 6.1 - Clean up abstractions

### Sprint 3: AST Restructure (Week 3)
1. Phase 1.2 - Consolidate AST packages
2. Phase 2.2 - Split large AST files

### Sprint 4: Dialects (Week 4)
1. Phase 3.2 - Refactor dialect interfaces

### Sprint 5: Tests & Docs (Week 5)
1. Phase 4.1 - Reorganize tests
2. Phase 7.1 - Update documentation

---

## Risk Mitigation

### Testing Strategy

- Run full test suite after each file move
- Use `go build ./...` to catch compilation errors
- Use `go vet ./...` for static analysis
- Maintain backward compatibility where possible

### Rollback Plan

- Create feature branches for each phase
- Tag stable points before major changes
- Keep old package structures until new ones are verified
- Use `git mv` for file moves to preserve history

### Compatibility Considerations

- If this is a public API, maintain backward compatibility:
  - Keep old type aliases during transition
  - Add deprecation notices
  - Provide migration guide for users
- If internal-only, can make breaking changes

---

## Success Metrics

After refactoring:

- [ ] No files exceed 800 lines (except generated code)
- [ ] Each package has a clear, single responsibility
- [ ] Interface segregation: no interface with >20 methods
- [ ] Test files organized by functionality
- [ ] All imports use consistent paths
- [ ] `go test ./...` passes
- [ ] `go build ./...` passes
- [ ] `go vet ./...` passes with no issues
- [ ] `gofmt -l .` produces no output

---

## Notes

- This plan assumes the codebase is internal (not a public library)
- If this is a public API, additional deprecation steps are needed
- Consider using automated tools like `gofmt`, `goimports`, `golint` during refactoring
- Document any deviations from this plan in the commit messages
