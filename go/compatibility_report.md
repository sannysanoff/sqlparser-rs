# Go ast_to_json Compatibility Report

## Overview
Built a Go equivalent of the Rust `ast_to_json` binary. The Go binary reads SQL from stdin, parses with the chosen dialect, and outputs pretty JSON.

## Files Created
- `go/astjson/marshal.go` – Reflection-based JSON serialization layer for Go AST types
- `go/cmd/ast_to_json/main.go` – CLI entry point (reads SQL from stdin, parses, outputs JSON)
- Binary installed to: `~/bin/go_sql_ast_to_json`

## Usage
```
echo "SELECT * FROM users" | go_sql_ast_to_json clickhouse
echo "CREATE TABLE t (id UInt64) ENGINE = MergeTree()" | go_sql_ast_to_json clickhouse
```

## Test Results

### 23 queries parsed successfully by BOTH parsers:
| # | Query Type | Status |
|---|------------|--------|
| 1 | Basic SELECT * | OK |
| 2 | SELECT with WHERE | OK |
| 3 | SELECT FINAL | OK (see note) |
| 4 | SELECT with WHERE condition | OK |
| 5 | arrayJoin function | OK |
| 6 | INNER JOIN with ON condition | OK |
| 7 | Nested column access | OK |
| 8 | CREATE TABLE with columns and ENGINE | OK |
| 9 | GROUP BY / HAVING / ORDER BY / LIMIT | OK |
| 10 | Functions (toYear, sum) | OK |
| 11 | ORDER BY ASC | OK |
| 12 | ORDER BY DESC NULLS LAST | OK |
| 13 | SELECT DISTINCT | OK |
| 14 | BETWEEN condition | OK |
| 15 | LIKE condition | OK |
| 16 | LIMIT with OFFSET | OK |
| 17 | Aggregate functions (count, min, max, avg) | OK |
| 18 | UNION ALL | OK |
| 19 | Subquery in FROM | OK |
| 20 | **PREWHERE clause** | **OK** |
| 21 | **LIMIT ... BY clause** | **OK** |
| 22 | **Parametric aggregate functions** | **OK** |
| 23 | **SELECT ... SETTINGS** | **OK** |

All 23 queries now pass in both parsers.

## Structural Differences

### JSON Output Structure
- **Rust** uses deeply nested enums: `Query → body → Select { select_token, optimizer_hints, ... }`
- **Go** flattens the nesting: `Query { flavor, from, projection, ... }` (the `body → Select` nesting is collapsed)

### Key Incompatibilities
1. **FINAL keyword**: Go parser treats `FINAL` as a table alias (`events AS FINAL`), not a table modifier
2. **queryExprWrapper**: Expression types from `ast/expr`/`ast/query` packages are bridged via `queryExprWrapper` structs. Serialized as `{"queryExprWrapper": {"expr": {...}}}` which adds an extra nesting layer vs Rust output
3. **No token/span info**: Go output omits source location metadata (which Rust includes from serde serialization)
4. **Enum values**: Some Go enum types lack `String()` methods, so integer values appear (e.g., `flavor: 0` instead of `"Standard"`)

### What Works Well
- Full expression tree serialization (WHERE, PREWHERE, GROUP BY, HAVING, ORDER BY, JOIN ON)
- Complex column references (compound identifiers, qualified wildcards)
- Function calls with parameters and arguments (parametric aggregates like quantile(0.5)(x))
- CREATE TABLE with column definitions and types
- UNION/UNION ALL set operations
- Subqueries in FROM
- DISTINCT, LIMIT, LIMIT BY, OFFSET
- SETTINGS clause on SELECT
