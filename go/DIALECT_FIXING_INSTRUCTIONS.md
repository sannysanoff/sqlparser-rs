# Dialect Fixing Instructions for sqlparser-rs Go Implementation

## CRITICAL: Always Reference Original Rust Code

Before making ANY changes, you MUST examine the original Rust source code:
- Location: `/Users/san/Fun/sqlparser-rs/src/dialects/[dialect_name].rs`
- Example: For BigQuery, check `/Users/san/Fun/sqlparser-rs/src/dialects/bigquery.rs`

**DO NOT invent types, methods, or logic.** Always verify against the original Rust implementation.

## Common Pitfalls Found in SQLite (Apply to All Dialects)

### 1. Duplicate Method Declarations
**Problem**: Methods declared twice (likely copy-paste errors during transpilation)
**Example from SQLite**:
```go
// Lines 110-147 had duplicate declarations of:
// - IsSelectItemAlias
// - IsTableFactor  
// - IsTableAlias
// - IsTableFactorAlias
```
**Fix**: Remove duplicate method declarations, keeping only the first occurrence.

### 2. Duplicate Case Statements in Precedence Switches
**Problem**: Go doesn't allow duplicate case values, even from different constant names
**Example**:
```go
// WRONG - PrecedencePipe and PrecedenceColon both = 21
case dialects.PrecedencePipe, dialects.PrecedenceColon:
    return 21
    
// WRONG - PrecedenceBetween and PrecedenceEq both = 20
case dialects.PrecedenceBetween, dialects.PrecedenceEq:
    return 20
```

**Correct Fix** (use explicit comparisons):
```go
func (d *Dialect) PrecValue(prec dialects.Precedence) uint8 {
    switch {
    case prec == dialects.PrecedencePipe:
        return 21
    case prec == dialects.PrecedenceColon:
        return 21
    case prec == dialects.PrecedenceBetween:
        return 20
    case prec == dialects.PrecedenceEq:
        return 20
    // ... other cases
    default:
        return d.PrecUnknown()
    }
}
```

**IMPORTANT**: Different dialects handle precedence differently:
- Some use `switch prec {` (standard)
- Some use `switch {` with `case prec == ...`
- PostgreSQL uses custom logic with `if prec == dialects.PrecedencePipe || prec == dialects.PrecedenceColon`

**ALWAYS check the Rust source to see the pattern used!**

### 3. Token Type Mismatches
**Problem**: Using wrong token type references
**Example from BigQuery**:
```go
// WRONG
token.Semicolon  // undefined
token.EOF        // undefined

// CORRECT
tokenizer.TokenSemicolon{}  // struct type
tokenizer.EOF               // struct type
```

**Check**: Look at `/Users/san/Fun/sqlparser-rs/go/tokenizer/tokens.go` for available token types.

### 4. Interface Compatibility Issues (CRITICAL)
**Problem**: `ast.Expr` and `expr.Expr` are DIFFERENT interfaces

**The Issue**:
- `ast.Expr` (in `ast/node.go`) requires: `expr()`, `IsExpr()`, `node()`, `Span()`, `String()`
- `expr.Expr` (in `ast/expr/expr.go`) requires: `exprNode()`, `Span()`, `String()`
- They are NOT compatible!

**Example from SQLite**:
```go
// WRONG - This doesn't work!
func ParseInfix(...) (ast.Expr, bool, error) {
    return &expr.BinaryOp{  // expr.BinaryOp implements expr.Expr, NOT ast.Expr
        Left:  expr,        // ast.Expr doesn't implement expr.Expr
        Op:    operator.BOpMatch,
        Right: right,
    }, true, nil
}
```

**Solution**: If you encounter this, the dialect method needs to:
1. Return results that match the interface, OR
2. Comment out the implementation with a TODO noting the interface mismatch

**For now**: Comment out and add TODO - this requires architectural alignment

### 5. ParserAccessor Interface Mismatches
**Problem**: The ParserAccessor interface in `dialects/dialect.go` may not match what dialects try to call

**Check the interface**:
```go
type ParserAccessor interface {
    PeekToken() tokenizer.TokenWithSpan
    ParseKeyword(expected string) bool
    PeekNthKeyword(n int, expected string) bool  // Note: returns bool, not token.Keyword
    ParseExpression() (ast.Expr, error)
    ParseInsert() (ast.Statement, error)
    // ... other methods
}
```

**Common mistakes**:
- `parser.PeekNthKeyword(1, "TABLES")` - Correct
- `parser.PeekNthKeyword(1)` expecting a token - Wrong

### 6. Import Path Issues
**Common missing imports**:
```go
import (
    "github.com/user/sqlparser/ast"
    "github.com/user/sqlparser/ast/expr"        // For BinaryOp, etc.
    "github.com/user/sqlparser/ast/operator"    // For BOpMatch, etc.
    "github.com/user/sqlparser/dialects"
    "github.com/user/sqlparser/token"
    "github.com/user/sqlparser/tokenizer"
)
```

### 7. Missing AST Types
**Problem**: Referencing types that don't exist in Go
**Example**: `ast.LockTablesStmt`, `ast.LockTable`, etc.

**Solution**: Add the missing types to the appropriate package:
- Statement types → `ast/statement/dcl.go` or appropriate file
- Expression types → `ast/expr/`
- Make sure to reference Rust source for exact structure

## Step-by-Step Fixing Process

1. **Build the dialect alone**:
   ```bash
   cd /Users/san/Fun/sqlparser-rs/go
   go build ./dialects/[dialect_name]
   ```

2. **Check original Rust source**:
   ```bash
   cat /Users/san/Fun/sqlparser-rs/src/dialects/[dialect_name].rs | head -100
   ```

3. **Fix errors in this order**:
   - Duplicate method declarations
   - Duplicate case statements (precedence)
   - Import path issues
   - Token type mismatches
   - ParserAccessor method mismatches
   - Missing AST types (move to appropriate package)
   - Interface compatibility issues (comment out if needed)

4. **Verify**:
   ```bash
   go build ./dialects/[dialect_name]
   ```

5. **Test the whole project**:
   ```bash
   go build ./...
   ```

## Dialects to Fix (Priority Order)

### High Priority (Most Used):
1. **bigquery** - Has TokenSemicolon issue
2. **mysql** - May have remaining issues
3. **postgresql** - Complex dialect, check carefully

### Medium Priority:
4. **snowflake** 
5. **duckdb**
6. **clickhouse**
7. **sqlite** - ✅ COMPLETED

### Lower Priority:
8. **hive**
9. **mssql**
10. **redshift**
11. **databricks**
12. **oracle**
13. **ansi**
14. **generic** - Likely already working

## Example: Fixing BigQuery

**File**: `/Users/san/Fun/sqlparser-rs/go/dialects/bigquery/bigquery.go`

**Current Error**:
```
dialects/bigquery/bigquery.go:883:45: undefined: tokenizer.TokenSemicolon
```

**Step 1**: Check the code around line 883
```bash
sed -n '880,890p' /Users/san/Fun/sqlparser-rs/go/dialects/bigquery/bigquery.go
```

**Step 2**: Check how SQLite does it (reference implementation)

**Step 3**: Check Rust source
```bash
grep -A5 -B5 "Semicolon\|semicolon" /Users/san/Fun/sqlparser-rs/src/dialects/bigquery.rs
```

**Step 4**: Fix the token reference
```go
// WRONG
if token.Token == token.Semicolon

// CORRECT  
if _, isSemicolon := token.Token.(tokenizer.TokenSemicolon); isSemicolon
```

**Step 5**: Verify build
```bash
go build ./dialects/bigquery
```

## Critical Reminders

1. **ALWAYS check Rust source first** - Don't guess, verify
2. **Run builds incrementally** - Fix one error at a time
3. **Test the whole project after each dialect** - `go build ./...`
4. **Comment rather than delete** - If something needs architectural fix, leave a TODO
5. **Be consistent** - If you fix a pattern in one place, apply it everywhere

## Success Criteria

- `go build ./dialects/[dialect_name]` succeeds with no errors
- `go build ./...` succeeds for the entire project
- Tokenizer tests still pass: `go test ./tokenizer -v`

## Contact

If you encounter an architectural issue (like the ast.Expr vs expr.Expr mismatch):
1. Document it with a TODO comment
2. Note it in your summary
3. Move on to fixable issues

Good luck!
