# Go Implementation Plan for sqlparser-rs

Complete re-implementation of sqlparser-rs in Go using automated transpilation.

**Project Scope:** ~38,000 lines of Rust → Go  
**Target:** Full feature parity with all 14 dialects and 1,260+ tests  
**Approach:** Automated transpilation with interface-based AST design

---

## Critical Implementation Rule

⚠️ **ALWAYS USE RUST IMPLEMENTATION AS REFERENCE** ⚠️

When implementing parser functionality:

1. **First, examine the Rust source** (`src/parser/mod.rs`, `src/ast/*.rs`, etc.)
2. **Port the logic directly** - do not reinvent or redesign
3. **Follow Rust naming conventions** where possible (e.g., `parse_create_view` → `parseCreateView`)
4. **Preserve behavior exactly** - edge cases, error messages, dialect-specific handling
5. **Comment with references** - cite the Rust file/line for complex logic

---

## Import Structure

The project uses a consolidated import structure after refactoring:

```go
import (
    "github.com/datafuselabs/sqlparser-go/ast"
    "github.com/datafuselabs/sqlparser-go/ast/expr"
    "github.com/datafuselabs/sqlparser-go/ast/statement"
    "github.com/datafuselabs/sqlparser-go/ast/query"
    "github.com/datafuselabs/sqlparser-go/parseriface"
    "github.com/datafuselabs/sqlparser-go/token"
    "github.com/datafuselabs/sqlparser-go/dialects"
)
```

The `token` package now contains both keywords and lexer/tokenizer functionality (merged from the old `tokenizer` and `span` packages). The `parseriface` package provides shared interfaces to avoid circular dependencies between `parser` and `dialects` packages.

---

## Typical Code Editing Errors and Patterns

This section documents common mistakes encountered when porting Rust to Go, categorized by pattern type. Understanding these patterns helps prevent similar bugs during implementation.

### Pattern A: Using Literal Tokens Instead of Constants

**Problem:** Hardcoding token values instead of using the token constant types.

**Example - parseObjectName (Dots vs Commas):**

```go
// INCORRECT - parsed comma-separated instead of dot-separated
func (ep *ExpressionParser) parseObjectName() (*expr.ObjectName, error) {
    idents, err := ep.parseCommaSeparatedIdents()  // Wrong! Uses commas
    ...
}

// CORRECT - parse dot-separated qualified names
func (ep *ExpressionParser) parseObjectName() (*expr.ObjectName, error) {
    var parts []*expr.ObjectNamePart
    for {
        ident, err := ep.parseIdentifier()
        if err != nil {
            return nil, err
        }
        parts = append(parts, &expr.ObjectNamePart{...})
        // Check for DOT token constant, not comma
        if !ep.parser.ConsumeToken(token.TokenPeriod{}) {
            break
        }
    }
    return &expr.ObjectName{Parts: parts}, nil
}
```

**Key Lesson:** Always use the proper token constant (`TokenPeriod`, `TokenComma`, etc.) rather than assuming separator behavior. Check Rust reference `src/parser/mod.rs:12715-12740` for `parse_object_name_inner`.

---

### Pattern B: Using Wrong Parser Function for Context

**Problem:** Using a generic identifier parser when the context requires qualified names.

**Example - JOIN USING Clause:**

```go
// INCORRECT - only parses simple identifiers
func parseJoinConstraint(p *Parser) (query.JoinConstraint, error) {
    cols, err := parseCommaSeparatedQueryIdents(p)  // Can only parse "col1"
    // FAILS on: JOIN tbl2 USING(t2.col1)
}

// CORRECT - parses qualified identifiers using proper function
func parseJoinConstraint(p *Parser) (query.JoinConstraint, error) {
    objNames, err := parseCommaSeparatedObjectNames(p)  // Can parse "t2.col1"
    // Convert []*ast.ObjectName to []query.ObjectName
    attrs := make([]query.ObjectName, len(objNames))
    for i, objName := range objNames {
        // ... conversion logic
    }
}
```

**Key Lesson:** When the SQL construct can contain qualified names (table.column), use `parseObjectName()` or `parseCommaSeparatedObjectNames()`, not simple identifier parsers. Reference: `src/parser/mod.rs:16687` for `parse_parenthesized_qualified_column_list`.

---

### Pattern C: Incorrect Token Type in Switch/Case

**Problem:** Using literal rune/byte comparisons instead of token type assertions.

**Example - Array Subscript Tokenizer Bug:**

```go
// INCORRECT - treated '[' as quoted identifier start
func (t *Tokenizer) NextToken() (Token, error) {
    switch state.Current() {
    case '`', '[':  // Wrong! '[' is not a quote character
        return t.tokenizeQuotedIdentifier(state)
    }
}

// CORRECT - '[' is a separate bracket token
func (t *Tokenizer) NextToken() (Token, error) {
    switch state.Current() {
    case '`':
        return t.tokenizeQuotedIdentifier(state)
    case '[':
        state.Next()
        return TokenLBracket{}, nil  // Return proper token type
    }
}
```

**Key Lesson:** Don't assume characters have the same token semantics as Rust. `[` is `Token::LBracket` in Rust (`src/tokenizer.rs:1691`), not a quoted identifier delimiter. Always verify token type mappings.

---

### Pattern D: Missing Dialect Capability Checks

**Problem:** Parsing dialect-specific syntax without checking if the dialect supports it.

```go
// INCORRECT - always tries to parse PostgreSQL syntax
func parseFunctionArgs(p *Parser) ([]expr.FunctionArg, error) {
    // Tries => syntax even for MySQL dialect
    if p.peekIsOperator("=>") {
        // parse named argument
    }
}

// CORRECT - check dialect first
func parseFunctionArgs(p *Parser) ([]expr.FunctionArg, error) {
    if p.GetDialect().SupportsNamedFnArgsWithRArrowOperator() {
        if p.peekIsOperator("=>") {
            // parse => syntax
        }
    }
    // ... fallback to regular args
}
```

**Key Lesson:** Always guard dialect-specific features with `dialect.SupportsXxx()` checks. Reference: `src/parser/mod.rs:17788-17836` for `parse_function_args()`.

---

### Pattern E: Unsafe Type Assertions

**Problem:** Direct type assertions without `ok` check cause panics.

```go
// INCORRECT - will panic if token is wrong type
func parseKeyword(p *Parser) string {
    word := p.PeekToken().Token.(token.TokenWord)  // Panic if not TokenWord
    return word.Value
}

// CORRECT - safe type assertion with ok check
func parseKeyword(p *Parser) (string, error) {
    tok := p.PeekToken()
    if word, ok := tok.Token.(token.TokenWord); ok {
        return word.Value, nil
    }
    return "", fmt.Errorf("expected keyword, got %T", tok.Token)
}
```

**Key Lesson:** Always use the `if x, ok := y.(Type); ok {` pattern. Go type assertions fail hard without the ok check.

---

### Pattern F: Case Sensitivity Assumptions

**Problem:** Assuming all dialects lowercase identifiers.

```go
// INCORRECT - always lowercases
func (i *Ident) String(dialect Dialect) string {
    return strings.ToLower(i.Value)  // Breaks MySQL tests
}

// CORRECT - preserve original case
func (i *Ident) String(dialect Dialect) string {
    // Rust preserves original case for ALL dialects to ensure AST comparison
    return i.Value  // Preserve original
}
```

**Key Lesson:** The Rust implementation preserves original case for ALL dialects to ensure consistent AST comparison. Don't normalize case in the parser. Reference: MySQL dialect in `src/dialect/mysql.rs`.

---

### Pattern G: Premature Token Consumption

**Problem:** Consuming tokens before validating the full pattern.

```go
// INCORRECT - consumes '(' then checks for empty
func parseWindowSpec(p *Parser) (*expr.WindowSpec, error) {
    p.ExpectToken(token.TokenLParen{})  // Consumes it
    if p.peekIsToken(token.TokenRParen{}) {
        // Error: we already consumed '('
        return &expr.WindowSpec{}, nil
    }
    // Now we need '(' again but it's gone
}

// CORRECT - peek first, consume only after validation
func parseWindowSpec(p *Parser) (*expr.WindowSpec, error) {
    // Check for empty spec: OVER ()
    tok := p.PeekToken()
    if _, ok := tok.Token.(token.TokenLParen); ok {
        // Check if next token after ( is )
        nextTok := p.PeekNthToken(1)
        if _, ok := nextTok.Token.(token.TokenRParen); ok {
            // Empty window specification
            p.AdvanceToken() // consume (
            p.AdvanceToken() // consume )
            return &expr.WindowSpec{}, nil
        }
    }
    // Continue parsing non-empty spec...
}
```

**Key Lesson:** Use `PeekToken()` and `PeekNthToken(n)` to look ahead before consuming. Only call `AdvanceToken()` or `ExpectToken()` after validation.

---

### Pattern H: Wrong Expression Return Type

**Problem:** Using `ast.Expr` interface instead of `expr.Expr` for statement fields.

```go
// INCORRECT - type mismatch
stmt := &statement.Pragma{
    Value: &expr.ValueExpr{...},  // *expr.ValueExpr doesn't implement ast.Expr
}

// CORRECT - use expr.Expr for expression fields
type Pragma struct {
    Value expr.Expr  // Not ast.Expr
}

stmt := &statement.Pragma{
    Value: &expr.ValueExpr{Value: ...},  // Works!
}
```

**Key Lesson:** The Go AST has two expression interfaces:

- `expr.Expr` - for expression types in `ast/expr/` (requires: `exprNode()`, `Span()`, `String()`)
- `ast.Expr` - sealed interface for all AST expressions (requires additional methods)

Use `expr.Expr` for statement fields that hold expressions.

---

### Pattern I: Stub Implementation Returns Empty

**Problem:** Stub functions return empty structs instead of parsing.

```go
// INCORRECT - silent failure with empty data
func parseIdentifier(p *Parser) (*ast.Ident, error) {
    return &ast.Ident{}, nil  // Returns empty! No error!
}

// Result: Table="", Alias="t" for "LOCK TABLES t READ"

// CORRECT - actually parse the token
func parseIdentifier(p *Parser) (*ast.Ident, error) {
    tok := p.PeekToken()
    if word, ok := tok.Token.(token.TokenWord); ok {
        p.AdvanceToken()
        return &ast.Ident{Value: word.Word.Value}, nil
    }
    return nil, fmt.Errorf("expected identifier, found %v", tok.Token)
}
```

**Key Lesson:** Always verify helper functions actually parse data. Empty returns without errors are silent bugs that show up as missing data later.

---

### Pattern J: Missing Statement Type Handling

**Problem:** Only handling one statement type when multiple are possible.

```go
// INCORRECT - only handles SelectStatement
func parseCTE(p *Parser) (*query.CTE, error) {
    innerQuery, err := p.parseQuery()
    if err != nil {
        return nil, err
    }

    if selStmt, ok := innerQuery.(*SelectStatement); ok {
        cte.Query = &query.Query{...}
    }
    // BUG: Query is nil if innerQuery is *QueryStatement (nested CTE)
}

// CORRECT - handle all possible query types
func parseCTE(p *Parser) (*query.CTE, error) {
    innerQuery, err := p.parseQuery()
    if err != nil {
        return nil, err
    }

    if selStmt, ok := innerQuery.(*SelectStatement); ok {
        cte.Query = &query.Query{...}
    } else if qStmt, ok := innerQuery.(*QueryStatement); ok {
        // Nested CTE (WITH clause inside CTE)
        cte.Query = qStmt.Query
    }
    // ... handle other cases
}
```

**Key Lesson:** When WITH clause is present, `parseQuery` can return `QueryStatement` instead of `SelectStatement`. All statement types that contain queries must handle both. Reference: `src/parser/mod.rs:13599-13610` for `parse_query()`.

---

### Pattern K: Hardcoded Case in Tokenizer Prefix Functions

**Problem:** Tokenizer functions that handle character-specific cases (like 'U' for Unicode strings) using hardcoded uppercase letters instead of preserving the original case.

**Example - tokenizeUnicodeStringLiteral:**

```go
// INCORRECT - hardcodes uppercase "U"
func (t *Tokenizer) tokenizeUnicodeStringLiteral(state *State) (Token, error) {
    // ...
    state.Next() // consume U/u
    // ...
    word := t.tokenizeWord(state, "U")  // Wrong! Always uses uppercase U
    return MakeWord(word, nil), nil
}

// Result: "uk_cities" becomes "Uk_cities" in the token value
```

```go
// CORRECT - preserve original case
func (t *Tokenizer) tokenizeUnicodeStringLiteral(state *State) (Token, error) {
    ch, _ := state.Peek()  // Save original character with its case
    state.Next() // consume U/u
    // ...
    word := t.tokenizeWord(state, string(ch))  // Use original case
    return MakeWord(word, nil), nil
}

// Result: "uk_cities" stays "uk_cities" - original case preserved
```

**Key Lesson:** When handling character-specific tokenizer cases (like Unicode string literals U&'...', byte strings B'...', etc.), always preserve the original character case. The tokenizer must not normalize case - that's the parser/AST's responsibility if needed. This affects identifiers starting with U, B, R, N, Q, E, X, etc. Reference: `src/tokenizer.rs` for how Rust handles these prefixes.

---

### Pattern L: Hardcoded Prefix Characters in Tokenizer Special Cases

**Problem:** Tokenizer functions that handle special character prefixes (like Q'...', E'...', X'...') using hardcoded uppercase characters instead of preserving the actual input character.

**Example - tokenizeQuoteDelimitedLiteral:**

```go
// INCORRECT - hardcodes uppercase "Q"
func (t *Tokenizer) tokenizeQuoteDelimitedLiteral(state *State) (Token, error) {
    state.Next() // consume Q/q
    next, ok := state.Peek()
    if !ok || next != '\'' {
        word := t.tokenizeWord(state, "Q")  // Wrong! Always uses uppercase Q
        return MakeWord(word, nil), nil
    }
    // ...
}

// Result: "quaternion" becomes "Quaternion" when not followed by quote
```

```go
// CORRECT - preserves original character case
func (t *Tokenizer) tokenizeQuoteDelimitedLiteral(state *State) (Token, error) {
    ch, _ := state.Peek()  // Save original character (Q or q)
    state.Next()           // consume Q/q
    next, ok := state.Peek()
    if !ok || next != '\'' {
        word := t.tokenizeWord(state, string(ch))  // Use original case
        return MakeWord(word, nil), nil
    }
    // ...
}

// Result: "quaternion" stays "quaternion" - original case preserved
```

**Key Lesson:** When tokenizer functions check for special prefixes (Q/q, E/e, X/x, etc.) and need to fall back to regular identifier tokenization, always use the actual consumed character rather than a hardcoded literal. The pattern of:
1. Check current character with `state.Peek()`
2. Save the character before consuming
3. Use `string(ch)` instead of hardcoded "Q"

This affects functions like `tokenizeQuoteDelimitedLiteral`, `tokenizeEscapedStringLiteral`, `tokenizeHexStringLiteral`, and others that handle special prefix characters. Reference: `src/tokenizer.rs` lines 544-628 for Rust reference implementation.

---

### Pattern M: Using Full ObjectName Parser for Simple Qualified Names

**Problem:** Using `ParseObjectName()` for qualified wildcards like `table.*` incorrectly consumes the dot and tries to parse additional identifier parts.

**Example - parseSelectItem (Qualified Wildcards):**

```go
// INCORRECT - consumes table.* incorrectly
func parseSelectItem(p *Parser) (query.SelectItem, error) {
    if isQualifiedWildcard(p) {
        tableName, err := p.ParseObjectName()  // Wrong! Consumes 'table.' and expects more
        // ... fails on 'inserted.*' because it tries to parse another identifier after '.'
    }
}

// CORRECT - parse single identifier only
func parseSelectItem(p *Parser) (query.SelectItem, error) {
    if isQualifiedWildcard(p) {
        // Don't use ParseObjectName() - it consumes the period and tries to parse more parts
        tableIdent, err := p.ParseIdentifier()  // Only parse the identifier
        if err != nil {
            return nil, err
        }
        p.ConsumeToken(token.TokenPeriod{}) // consume the . explicitly
        p.ConsumeToken(token.TokenMul{})    // consume the * explicitly
        
        // Build ObjectName from single identifier
        queryName := query.ObjectName{
            Parts: []query.Ident{{Value: tableIdent.Value}},
        }
        return &query.QualifiedWildcard{
            Kind: &query.ObjectNameWildcard{Name: queryName},
        }, nil
    }
}
```

**Key Lesson:** `ParseObjectName()` is designed for multi-part names like `schema.table.column` and will consume periods greedily. For qualified wildcards like `table.*`, use `ParseIdentifier()` directly and handle the period and star tokens explicitly. Reference: `src/parser/mod.rs` for how Rust handles select item parsing.

---

### Pattern N: Incorrect Empty String vs nil Pointer in AST Fields

**Problem:** Setting a pointer to an empty string `&""` instead of leaving it as `nil` causes incorrect String() output.

**Example - CEIL/FLOOR parsing:**

```go
// INCORRECT - creates "CEIL(1.5 TO )" output
func parseCeilFloorExpr(isCeil bool) (expr.Expr, error) {
    // ...
    } else {
        // CEIL/FLOOR(expr) - simple case
        ceilExpr.Field.Kind = expr.CeilFloorDateTime
        empty := ""  // Wrong! Creates pointer to empty string
        ceilExpr.Field.DateTimeField = &empty
    }
    // String() checks: if Kind == CeilFloorDateTime && DateTimeField != nil
    // This passes because DateTimeField is not nil (points to empty string)
    // Result: fmt.Sprintf("CEIL(%s TO %s)", expr, *DateTimeField) outputs "CEIL(1.5 TO )"
}

// CORRECT - leaves DateTimeField as nil
func parseCeilFloorExpr(isCeil bool) (expr.Expr, error) {
    // ...
    } else {
        // CEIL/FLOOR(expr) - simple case, no DateTimeField
        ceilExpr.Field.Kind = expr.CeilFloorDateTime
        // Don't set DateTimeField - leave it nil so String() outputs simple form
    }
    // String() checks: if Kind == CeilFloorDateTime && DateTimeField != nil
    // This fails because DateTimeField is nil
    // Falls through to: fmt.Sprintf("CEIL(%s)", expr) outputs "CEIL(1.5)"
}
```

**Key Lesson:** When an AST field is optional (pointer type), leave it as `nil` when not present, don't set it to a pointer to an empty value. The String() methods typically check `!= nil` to determine if the field should be rendered. Reference: `src/parser/mod.rs` and AST String() implementations.

---

## Current Status

**Overall Progress: 38% Test Pass Rate** (459/1207 tests passing)

| Test Suite       | Status           | Passing | Total | Pass Rate |
| ---------------- | ---------------- | ------- | ----- | --------- |
| **TPC-H**        | ⚠️ Fixture issue  | 0       | 44    | **0%**    |
| **Common Tests** | 🔄 In Progress   | ~200    | ~435  | **46%**   |
| **PostgreSQL**   | 🔄 In Progress   | ~40     | ~157  | **25%**   |
| **MySQL**        | 🔄 In Progress   | ~57     | ~125  | **46%**   |
| **Snowflake**    | 🔄 In Progress   | ~16     | ~97   | **16%**   |
| **TOTAL**        | **38% Complete** | **459** | 1207  | **38%**   |

**Line Counts:**

- Rust Source: 67,345 lines
- Go Source: 64,156 lines (95% of Rust)
- Go Tests: 14,112 lines (28% of Rust tests)

---

## Recent Progress

### April 5, 2026 - UPDATE FROM, CEIL/FLOOR Fix, Array Subscript Parser

Implemented major missing parser chunks following Rust reference:

1. **UPDATE FROM Clause** - Added full support per Rust `parse_update` (src/parser/mod.rs:17715):
   - PostgreSQL style: `UPDATE t1 SET ... FROM t2 WHERE ...`
   - Snowflake/MSSQL style: `UPDATE FROM t1 SET ... WHERE ...`
   - Support for multiple tables in FROM clause (comma-separated)
   - Updated `Update` AST struct to use `UpdateTableFromKind` (BeforeSet/AfterSet)
   - **+2 tests passing** (UPDATE FROM tests)

2. **CEIL/FLOOR String() Fix** - Fixed incorrect output format:
   - Problem: `CEIL(1.5)` was outputting as `CEIL(1.5 TO )` 
   - Root cause: Parser was setting `DateTimeField = &""` (empty string pointer) instead of leaving it nil
   - Fix: Removed the empty string assignment in `parseCeilFloorExpr` when no TO clause present
   - **+2 tests passing** (CEIL/FLOOR tests)

3. **Array Subscript Infix Parser** - Added missing `[` token handler:
   - Added `TokenLBracket` case to `parseInfix` in infix.go
   - Connected to existing `parseArraySubscript` function in postfix.go
   - Enables parsing of expressions like `arr[1]` or `matrix[i][j]`
   - **+1 test passing** (Array subscript test)

**Implementation Pattern:** When adding infix operators, always:
1. Add precedence case in `getPrecedence` (core.go)
2. Add handler case in `parseInfix` (infix.go) 
3. Ensure the handler consumes the token and returns proper expression type

---

### April 5, 2026 - MERGE OUTPUT/RETURNING and TRUNCATE Implementation

Implemented two major missing parser chunks:

1. **MERGE OUTPUT/RETURNING Clause** - Added support for:
   - SQL Server OUTPUT clause: `MERGE ... OUTPUT inserted.* INTO log_table`
   - PostgreSQL RETURNING clause: `MERGE ... RETURNING merge_action(), *`
   - Fixed qualified wildcard parsing (e.g., `inserted.*`, `w.*`) in projection parser
   - Added `parseOutputClauseInternal()` and `parseSelectIntoInternal()` functions
   - **+8 tests passing** (MERGE OUTPUT/RETURNING tests)

2. **TRUNCATE Statement** - Implemented full parser:
   - TRUNCATE [TABLE] name [, ...] syntax
   - IF EXISTS modifier
   - PARTITION clause support
   - ON CLUSTER (ClickHouse) support
   - **+2 tests passing** (TRUNCATE tests)

**Key Bug Fix:** The qualified wildcard parsing was incorrectly using `ParseObjectName()` which consumed the dot and tried to parse more identifiers. Fixed by parsing just a single identifier for the table name in qualified wildcards.

---

## Major Missing Parser Chunks (Priority Order)

Based on test failures analysis, the following major parser chunks need implementation:

1. **SET Statement Variants** (~10+ test failures)
   - SET TRANSACTION READ ONLY/READ WRITE/ISOLATION LEVEL
   - SET (a, b, c) = (1, 2, 3) parenthesized assignment syntax
   - SET TIME ZONE TO 'value' syntax
   - SET SESSION AUTHORIZATION
   - SET NAMES (MySQL)

2. **SHOW Statement Extensions** (~40+ test failures)
   - SHOW DATABASES, SHOW SCHEMAS
   - SHOW VIEWS, SHOW MATERIALIZED VIEWS
   - SHOW FUNCTIONS
   - TERSE, HISTORY, EXTERNAL modifiers
   - Filter position variations (LIKE/WHERE before IN/FROM)

3. **CREATE Statement Extensions** (~30+ "not yet implemented" errors)
   - CREATE MATERIALIZED VIEW
   - CREATE PROCEDURE
   - CREATE USER
   - CREATE ASSERT
   - CREATE TRIGGER, CREATE FUNCTION, CREATE DATABASE
   - CREATE POLICY, CREATE CONNECTOR, CREATE OPERATOR

4. **UPDATE FROM Clause** (~5+ test failures)
   - UPDATE ... SET ... FROM ... WHERE syntax (MSSQL/PostgreSQL)

5. **FETCH Statement** (~3+ test failures)
   - FETCH [FIRST|NEXT] n [ROW|ROWS] [ONLY|WITH TIES]

---

## Recent Progress (Concise)

### April 5, 2026 - SHOW Statement Extensions and CREATE TABLE AS/LIKE

Implemented major missing parser chunks to bring maximum test coverage:

1. **SHOW Statement Extensions** - Added support for:
   - SHOW DATABASES, SHOW SCHEMAS with TERSE/HISTORY modifiers
   - SHOW VIEWS, SHOW MATERIALIZED VIEWS
   - SHOW FUNCTIONS, SHOW OBJECTS
   - EXTERNAL modifier for SHOW TABLES
   - Bare string literal suffix filters (Snowflake-style: `SHOW TABLES IN db1 'abc'`)
   - Added `SuffixString` field to `ShowStatementFilter` AST type

2. **CREATE TABLE AS/LIKE** - Added support for:
   - CREATE TABLE ... AS (query) - with proper query extraction from SelectStatement/QueryStatement
   - CREATE TABLE ... LIKE (table) - with CreateTableLikeKind
   - Parsing of LOCAL/GLOBAL/TRANSIENT modifiers before CREATE TABLE
   - Updated `ParseCreate` to handle modifier parsing order per Rust reference

**Result:** SHOW statements now working for Snowflake, MySQL, and Generic dialects. CREATE TABLE AS basic functionality operational. Common tests pass rate improved.

### April 4, 2026 - Analysis of Missing Parser Chunks

Implemented CREATE ROLE and DROP ROLE parsers, added DENY statement support, and fixed critical tokenizer case preservation bugs. The tokenizer was hardcoding uppercase prefixes in `tokenizeQuoteDelimitedLiteral`, `tokenizeEscapedStringLiteral`, and `tokenizeHexStringLiteral`, causing identifiers like "quaternion" to become "Quaternion" in GenericDialect. Added Pattern L documenting this widespread issue. **22 more tests passing** (356/858 total, 41% pass rate).

### April 4, 2026 - Tokenizer Case Preservation Fix

Fixed critical bug in `tokenizeUnicodeStringLiteral()` that was hardcoding uppercase "U" instead of preserving original case. This caused identifiers starting with 'u' (like `uk_cities`) to be incorrectly tokenized as "Uk_cities". Added Pattern K to "Typical Code Editing Errors" section documenting this case. **10 more tests passing** (334/858 total, 39% pass rate).

### April 4, 2026 - Named Argument Test Fix

Fixed `TestParseNamedArgumentFunctionWithEqOperator` to use `NewTestedDialectsWithFilter()` with predicate for `SupportsNamedFnArgsWithEqOperator()`, ensuring only dialects that support `=` for named arguments are tested. This aligns with the Rust reference test using `all_dialects_where()`.

### April 4, 2026 - Array Subscripts and CTEs

Implemented array subscript parsing, CTE in CREATE VIEW, UNNEST table factors, and VALUES as table factors. **6 more tests passing** (40% common tests).

### April 4, 2026 - Transaction Statements

Implemented COMMIT, ROLLBACK, SAVEPOINT, RELEASE, START TRANSACTION, LISTEN/NOTIFY, SQLite PRAGMA, and transaction modes. **52 more tests passing** (308 total).

### April 4, 2026 - INTERVAL, Typed Strings, TABLE Function

Fixed INTERVAL in window frames for dialects requiring qualifiers, added TIMESTAMPTZ/JSON/BIGNUMERIC support, implemented TABLE function parsing. **5 more tests passing**.

### April 4, 2026 - MERGE Statement

Implemented MERGE INTO statement with UPDATE SET, DELETE, INSERT actions and compound identifiers. **17 more tests passing** (251 total, 31%).

### April 4, 2026 - COPY and Window Functions

Implemented PostgreSQL COPY FROM/TO, fixed OVER clause for PARTITION BY/ORDER BY, added QUALIFY clause. **4 more tests passing**.

### April 4, 2026 - GRANT/REVOKE and LOCK TABLES

Implemented GRANT/REVOKE statements and MySQL LOCK TABLES with implicit AS. **3 more MySQL tests passing** (42%).

### April 4, 2026 - Named Arguments and CTE

Implemented PostgreSQL named arguments (`=>` syntax) and basic CTE (WITH clause) parsing. **2 more tests passing**.

---

## Project Structure

```
sqlparser-go/
├── token/                      # 800+ SQL keywords, lexer, position/span
├── errors/                     # Error types
├── ast/                        # Abstract Syntax Tree
│   ├── statement/             # Statement types (consolidated)
│   ├── expr/                  # Expression types
│   ├── datatype/              # DataType variants
│   ├── query/                 # Query-related types
│   ├── node.go                # Base Node types
│   ├── expr.go, expr_all.go   # Expression interfaces and all expressions
│   ├── statement.go, statement_all.go # Statement interfaces and types
│   ├── query.go, query_all.go # Query interfaces and types
│   └── types_all.go, operators_all.go # Consolidated type definitions
├── parser/                     # Parser (~10,000 lines, split by function)
│   ├── core.go, parser.go     # Core parser types and main entry points
│   ├── helpers.go, utils.go   # Parser utilities
│   ├── prefix.go, infix.go, postfix.go # Expression parsing
│   ├── create.go, alter.go, drop.go, truncate.go # DDL statements
│   ├── dml.go                 # INSERT, UPDATE, DELETE
│   ├── query.go               # SELECT, CTE parsing
│   ├── merge.go, copy.go, show.go, describe.go, misc.go # Other statements
│   └── transaction.go, prepared.go, special.go, groupings.go, options.go, state.go
├── parseriface/               # Parser interface definitions (resolves circular deps)
├── dialects/                  # 14 SQL dialects
│   ├── dialect.go, capabilities.go # Core dialect interfaces
│   ├── generic, postgresql, mysql, sqlite, bigquery
│   ├── snowflake, duckdb, clickhouse, hive, mssql
│   └── redshift, databricks, oracle, ansi
├── tests/                     # Test suites (TPC-H 100% ✅)
│   ├── ddl/                   # CREATE, ALTER, DROP, TRUNCATE tests
│   ├── query/                 # JOIN, CTE, set operations tests
│   └── snowflake/, mysql/, postgres/ # Dialect-specific tests
├── fuzz/                      # Fuzz testing
└── examples/                  # Usage examples
```

---

## Test Porting Status

| Phase | Test Suite     | Tests | Status                    |
| ----- | -------------- | ----- | ------------------------- |
| 1     | Common         | 435   | ✅ 97% ported, 40% passing |
| 2     | PostgreSQL     | 157   | 🔄 22% passing            |
| 2     | MySQL          | 125   | 🔄 46% passing            |
| 2     | Snowflake      | 97    | 🔄 14% passing            |
| 3-4   | Other dialects | 444   | ⏳ Pending                 |

**Completed:** Tokenizer (29/29), TPC-H (44/44, 100%), AST types (131 statements, 69 expressions, 117 data types), 14 dialects, Fuzz testing.

---

## AST Interface Design

```go
// Sealed interface pattern

type Node interface {
    node() // Unexported prevents external implementation
}

type Statement interface {
    Node
    statementNode()
    String() string  // SQL regeneration
}

type Expr interface {
    Node
    exprNode()
    String() string
}

// Type assertion pattern (replaces Rust match)
func processStatement(stmt Statement) error {
    switch s := stmt.(type) {
    case *SelectStmt:
        return handleSelect(s)
    case *InsertStmt:
        return handleInsert(s)
    default:
        return fmt.Errorf("unknown statement: %T", stmt)
    }
}
```

---

## Running Tests

```bash
cd /Users/san/Fun/sqlparser-rs/go  # Must run from go/ directory

go test ./token/... -v               # Tokenizer tests (29 passing)
go test ./tests/... -v              # All test suites
go test ./tests/ddl/... -v          # DDL tests (CREATE, ALTER, DROP)
go test ./tests/query/... -v         # Query tests (JOIN, CTE)
go test ./tests/mysql/... -v         # MySQL dialect
go test ./tests/postgres/... -v      # PostgreSQL dialect
go build ./...                      # Build everything
```

---

**Version:** 1.0  
**Last Updated:** April 5, 2026  
**Status:** TPC-H fixture issue, Common 46%, PostgreSQL 25%, MySQL 46%, Snowflake 16%, **Total 459/1207 (38%)**
