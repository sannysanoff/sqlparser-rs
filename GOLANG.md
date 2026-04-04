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
        if !ep.parser.ConsumeToken(tokenizer.TokenPeriod{}) {
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
    word := p.PeekToken().Token.(tokenizer.TokenWord)  // Panic if not TokenWord
    return word.Value
}

// CORRECT - safe type assertion with ok check
func parseKeyword(p *Parser) (string, error) {
    tok := p.PeekToken()
    if word, ok := tok.Token.(tokenizer.TokenWord); ok {
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
    p.ExpectToken(tokenizer.TokenLParen{})  // Consumes it
    if p.peekIsToken(tokenizer.TokenRParen{}) {
        // Error: we already consumed '('
        return &expr.WindowSpec{}, nil
    }
    // Now we need '(' again but it's gone
}

// CORRECT - peek first, consume only after validation
func parseWindowSpec(p *Parser) (*expr.WindowSpec, error) {
    // Check for empty spec: OVER ()
    tok := p.PeekToken()
    if _, ok := tok.Token.(tokenizer.TokenLParen); ok {
        // Check if next token after ( is )
        nextTok := p.PeekNthToken(1)
        if _, ok := nextTok.Token.(tokenizer.TokenRParen); ok {
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
    if word, ok := tok.Token.(tokenizer.TokenWord); ok {
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

## Current Status

**Overall Progress: 39% Test Pass Rate** (318/816 tests passing)

| Test Suite       | Status           | Passing | Total | Pass Rate |
| ---------------- | ---------------- | ------- | ----- | --------- |
| **TPC-H**        | ✅ Perfect        | 44      | 44    | **100%**  |
| **Common Tests** | 🔄 In Progress   | 203     | 435   | **47%**   |
| **PostgreSQL**   | 🔄 In Progress   | ~40     | 157   | **~25%**  |
| **MySQL**        | 🔄 In Progress   | ~31     | 125   | **~25%**  |
| **Snowflake**    | 🔄 In Progress   | ~0      | 97    | **~0%**   |
| **TOTAL**        | **39% Complete** | **318** | 816   | **39%**   |

**Line Counts:**

- Rust Source: 67,345 lines
- Go Source: 56,064 lines (83% of Rust)
- Go Tests: 14,492 lines

---

## Major Missing Parser Chunks (Priority Order)

Based on test failures analysis, the following major parser chunks need implementation:

1. **SHOW Statement Extensions** (~40+ test failures)
   - SHOW DATABASES, SHOW SCHEMAS
   - SHOW VIEWS, SHOW MATERIALIZED VIEWS
   - SHOW FUNCTIONS
   - TERSE, HISTORY, EXTERNAL modifiers
   - Filter position variations (LIKE/WHERE before IN/FROM)

2. **CREATE TABLE AS/LIKE/CLONE** (~25+ test failures)
   - CREATE TABLE ... AS (query)
   - CREATE TABLE ... LIKE (table)
   - CREATE TABLE ... CLONE (table)
   - PARTITION OF, ON CLUSTER

3. **Named Arguments with => operator** (~15+ test failures)
   - PostgreSQL named argument syntax: func(arg => value)

4. **OUTPUT/RETURNING in MERGE** (~10+ test failures)
   - SQL Server OUTPUT clause in MERGE
   - PostgreSQL RETURNING clause in MERGE

5. **CREATE/DROP/ALTER Extensions** (~50+ "not yet implemented" errors)
   - CREATE TRIGGER, CREATE FUNCTION, CREATE DATABASE
   - CREATE POLICY, CREATE CONNECTOR, CREATE OPERATOR
   - ALTER VIEW, ALTER INDEX, ALTER POLICY, etc.

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
├── token/                      # 800+ SQL keywords
├── span/                       # Source location tracking
├── errors/                     # Error types
├── tokenizer/                  # Lexer (~4,500 lines, 29 tests ✅)
├── ast/                        # Abstract Syntax Tree
│   ├── statement/             # 131 Statement types
│   ├── expr/                  # 69 Expression types
│   ├── datatype/              # 117 DataType variants
│   └── query/                 # Query-related types
├── parser/                    # Parser (~10,000 lines)
├── dialects/                  # 14 SQL dialects
│   ├── generic, postgresql, mysql, sqlite, bigquery
│   ├── snowflake, duckdb, clickhouse, hive, mssql
│   └── redshift, databricks, oracle, ansi
└── tests/                     # Test suites (TPC-H 100% ✅)
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

go test ./tokenizer/... -v           # 29 tests passing
go test ./tests/... -v              # TPC-H 100% passing
go test ./tests/mysql/... -v        # MySQL dialect
go test ./tests/postgres/... -v     # PostgreSQL dialect
go build ./...                      # Build everything
```

---

**Version:** 1.0  
**Last Updated:** April 4, 2026  
**Status:** TPC-H 100%, Common 40%, PostgreSQL 22%, MySQL 46%, Snowflake 14%, **Total 324/858 (40%)**
