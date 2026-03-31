# SQL Expression Parser - Implementation Summary

## Overview
This implementation provides comprehensive SQL expression parsing methods for the sqlparser Go library, transpiled from the Rust sqlparser-rs crate.

## Files Created

### Core Files (in `go/parser/expressions/`)

1. **doc.go** - Package documentation explaining the structure and usage

2. **core.go** - Core expression parsing infrastructure:
   - `ExpressionParser` struct and constructor
   - `ParserInterface` for avoiding circular dependencies
   - `ParseExpr()` - Main entry point for expression parsing
   - `ParseExprWithPrecedence()` - Pratt parsing with precedence climbing
   - `GetNextPrecedence()` / `GetNextPrecedenceDefault()` - Operator precedence
   - `parseCompoundExpr()` - Field access chain parsing (e.g., a.b.c)
   - Helper methods for identifier handling and subscript parsing

3. **prefix.go** - Prefix expression parsers:
   - `parsePrefix()` - Entry point for prefix expressions
   - `parseIdentifier()` - Column/table names
   - `parseValue()` - Literals (strings, numbers, booleans, null)
   - `parseFunction()` - Function calls
   - `parseParenthesizedPrefix()` - Parenthesized expressions and tuples
   - Word-based prefix parsing (reserved keywords)
   - Unary operators (+, -, NOT, ~, !, etc.)
   - PostgreSQL geometric operators

4. **infix.go** - Infix expression parsers:
   - `parseInfix()` - Entry point for infix operators
   - `tokenToBinaryOperator()` - Token to operator mapping
   - `parseBinaryOp()` - Binary operations with ALL/ANY/SOME support
   - IS [NOT] NULL/TRUE/FALSE/DISTINCT FROM
   - IN expressions (list, subquery, UNNEST)
   - BETWEEN expressions
   - LIKE/ILIKE/SIMILAR TO/RLIKE/REGEXP
   - AT TIME ZONE
   - PostgreSQL custom operators

5. **postfix.go** - Postfix expression parsers:
   - `parseArraySubscript()` - Array indexing and slicing
   - `parseCollate()` - Collation expressions

6. **special.go** - Special expression types:
   - `parseFunctionWithName()` - Function calls with full argument parsing
   - `parseWindowSpec()` - Window specifications for OVER clause
   - `parseCaseExpr()` - CASE WHEN THEN ELSE END
   - `parseSubqueryExpr()` - Subqueries
   - `parseExistsExpr()` - EXISTS(subquery)
   - Aggregate function support (FILTER, OVER, WITHIN GROUP)

7. **helpers.go** - Utility methods:
   - `parseCommaSeparatedExprs()` - Comma-separated expression lists
   - `parseOptionalAlias()` - [AS] alias parsing
   - `parseArrayExpr()` - ARRAY[...] literals
   - `parseIntervalExpr()` - INTERVAL expressions with temporal units
   - `parseCastExpr()` - CAST/TRY_CAST/SAFE_CAST
   - `parseConvertExpr()` - CONVERT with multiple syntaxes
   - `parseExtractExpr()` - EXTRACT(field FROM expr)
   - `parseCeilFloorExpr()` - CEIL/FLOOR functions
   - `parsePositionExpr()` - POSITION(substr IN str)
   - `parseSubstringExpr()` - SUBSTRING/SUBSTR variants
   - `parseOverlayExpr()` - OVERLAY(...)
   - `parseTrimExpr()` - TRIM with WHERE/FROM options
   - `parseStructLiteral()` - STRUCT<...>(...) literals
   - `parseMapLiteral()` - MAP literals (DuckDB)
   - `parseDictionaryExpr()` - Dictionary/struct literals
   - `parseLambdaExpr()` - Lambda functions (x -> expr)

8. **groupings.go** - GROUP BY expression parsers:
   - `ParseGroupingSets()` - GROUPING SETS (...)
   - `ParseCube()` - CUBE (...)
   - `ParseRollup()` - ROLLUP (...)

## Key Features

### Precedence Climbing (Pratt Parsing)
The parser uses precedence climbing for handling operator precedence correctly:
- Highest: Period (.) for field access
- High: Multiplicative (*, /, %)
- Medium: Additive (+, -)
- Low: Comparison (=, <>, <, >, etc.)
- Lowest: Logical (AND, OR)

### Dialect Support
The parser supports multiple SQL dialects through the `dialects.Dialect` interface:
- PostgreSQL-specific operators and types
- MySQL/Oracle/MS SQL Server specific syntax
- BigQuery lambda functions
- DuckDB map literals
- Geometric types (PostgreSQL)

### Expression Types Supported

**Basic Expressions:**
- Identifiers (column/table names)
- Literals (strings, numbers, booleans, NULL)
- Qualified wildcards (table.*)

**Operators:**
- Arithmetic: +, -, *, /, %
- Comparison: =, <>, <, >, <=, >=, <=> (spaceship)
- Logical: AND, OR, NOT, XOR
- Bitwise: |, &, ^, ~, <<, >>
- Pattern matching: LIKE, ILIKE, SIMILAR TO, RLIKE, REGEXP
- JSON operators: ->, ->>, #>, #>>, @>, <@, etc.
- PostgreSQL: ~, ~*, !~, !~*, && (overlap), etc.

**Special Expressions:**
- CASE WHEN THEN ELSE END
- CAST(expr AS type), TRY_CAST, SAFE_CAST
- CONVERT(type, expr) / CONVERT(expr, type)
- EXTRACT(field FROM expr)
- POSITION(substr IN str)
- SUBSTRING(expr [FROM start] [FOR length])
- TRIM([[WHERE] [chars] FROM] expr)
- OVERLAY(expr PLACING what FROM start [FOR length])
- INTERVAL 'value' unit
- ARRAY[expr1, expr2, ...]
- STRUCT<...>(...) / STRUCT(...)
- EXISTS(subquery)
- Window functions with OVER clause

## Architecture Notes

### Type System
The parser works with `expr.Expr` interface types from the `expr` package. There's a noted design issue where some types like `OrderByExpr` don't implement the full `Expr` interface, requiring workarounds in the current implementation.

### Dialect Integration
The dialect system was designed for the main `ast` package, creating an architectural mismatch with the `expr` package. The current implementation works around this by:
1. Using dialect feature flags (e.g., `dialect.SupportsLambdaFunctions()`)
2. Checking dialect names (e.g., `dialect.Dialect() == "postgresql"`)
3. Temporarily disabling dialect hooks for ParsePrefix/ParseInfix (marked as TODO)

## TODOs for Future Work

1. **Resolve AST Type Design Issues:**
   - Make `OrderByExpr` implement the `Expr` interface
   - Fix type mismatches between `[]*T` and `[]T` in AST definitions
   - Enable proper custom operator name storage

2. **Complete Dialect Integration:**
   - Re-enable ParsePrefix/ParseInfix dialect hooks
   - Resolve `ast.Expr` vs `expr.Expr` type mismatch

3. **Enhance Operator Support:**
   - Full PostgreSQL custom operator support with names
   - Dialect-specific binary operator handling

4. **Add Missing Expression Types:**
   - Typed string literals (DATE '2020-01-01')
   - More geometric types
   - Full data type parsing in CAST/CONVERT

## Usage Example

```go
dialect := dialects.NewGenericDialect()
parser := parser.New(dialect)
exprParser := expressions.NewExpressionParser(parser)

// Parse an expression
expr, err := exprParser.ParseExpr()
if err != nil {
    // Handle error
}

// expr is now an expr.Expr containing the parsed AST
```

## Testing

The implementation compiles successfully and is ready for integration testing with actual SQL parsing scenarios.

## License

All files include the Apache 2.0 license header as required.
