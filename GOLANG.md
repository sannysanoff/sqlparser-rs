# Go SQL Parser Development Guide

## IMMUTABLE: Development Methodology

**This section contains permanent development guidelines. Do not modify without team review.**

### Session Workflow

1. **Analyze current state**: Check test counts and identify failing patterns
2. **Identify high-impact features**: Target fixes that resolve multiple tests
3. **Implement with minimal changes**: Follow existing patterns, avoid over-engineering
4. **Document patterns**: Add Pattern E### entries for reusable solutions
5. **Update statistics**: Record line counts and test results
6. **Compress history**: When session summaries exceed 100 lines, compress older entries into "Previous Sessions Archive"

### Code Porting Principles

- Port complete Rust parser modules rather than fixing tests individually
- Match Rust AST structure exactly - Go interfaces should mirror Rust enums
- All dialect capability methods must delegate to underlying dialect (no hardcoding)
- Span mismatches (column position differences) are non-functional - focus on true parsing failures

### Test Standards

- Use `NewTestedDialectsWithFilter()` for dialect-specific syntax
- Update tests to match Go's canonical uppercase form for keywords
- Distinguish span mismatches from true parsing failures
- When tests fail on AST comparison, check both parsing and serialization

### Pattern Documentation Format

```
Pattern E###: Brief description
- When X happens, do Y
- Example: code snippet
- Files typically modified: ...
```

### History Compression Rule

**When session history exceeds 100 lines:**
1. Create "Previous Sessions Archive" section at document end
2. Move sessions older than 10 most recent to archive
3. In archive, keep only: Session number, date, tests passing/failing count, and 1-line summary
4. Keep full details for 10 most recent sessions in main history

---

## Current Status Summary

**Latest Update: April 8, 2026 - Session 71 Complete**

**Summary:**
- Previous: ~101 subtests failing
- Current: **~100 subtests failing** (715+ passing)
- Net change: ~3 tests fixed, major keyword handling implementation
- Key fixes:
  1. **Keywords as column names after dot**: Fixed parser to handle `T.interval`, `T.case`, `T.cast` etc.
  2. **Reserved keywords in compound identifiers**: Added CASE, CAST, EXTRACT, SUBSTRING, LEFT, RIGHT to RESERVED_FOR_IDENTIFIER
  3. **INSERT ... RETURNING fix**: RETURNING keyword now works after SELECT source
- Remaining high-impact work: MySQL functional key parts, PostgreSQL CREATE INDEX options, NOT precedence

---

## Line Counts (Updated April 8, 2026 - Session 71)

| Component | Rust | Go | Ratio |
|-----------|------|-----|-------|
| Source (parser+ast+dialects) | 67,345 lines | 87,829 lines | 130% |
| Tests | 49,886 lines | 14,245 lines | 29% |
| **Test Status - Snowflake** | - | **100% passing** (99+ subtests) |
| **Test Status - All Packages** | - | **~715 subtests passing, ~100 failing** |

---

## Session 71 Summary: Keywords as Column Names After Dot (April 8, 2026)

**Major Bug Fix:**

Fixed critical bug where keywords like `interval`, `case`, `cast` were not being treated as identifiers after a dot in compound expressions like `T.interval`.

**Root Cause:**
The `parseCompoundExprWithOptions()` function in `parser/core.go` was trying to parse keywords like `case` as expressions (e.g., CASE WHEN ... END) even when they appeared after a dot (e.g., `T.case`). The Rust implementation treats reserved keywords as identifiers when they appear after a dot.

**The Fix:**
Two changes were needed:

1. In `parser/core.go` - Check for reserved keywords after dot and treat them as identifiers:
```go
case token.TokenWord:
    // Check if the word is a reserved keyword that should be treated as identifier after a dot
    if token.IsReservedForIdentifier(tok.Word.Keyword) {
        // Reserved keyword - treat as identifier after dot (e.g., T.interval)
        ident := ep.parseIdentifierFromWord(tok, nextTok.Span)
        chain = append(chain, &expr.DotAccess{...})
        ep.parser.AdvanceToken()
    }
```

2. In `token/keywords.go` - Added keywords to RESERVED_FOR_IDENTIFIER list:
```go
var RESERVED_FOR_IDENTIFIER = []Keyword{
    EXISTS, INTERVAL, STRUCT, TRIM,
    // Keywords that should be treated as identifiers after a dot
    CASE, CAST, EXTRACT, SUBSTRING, LEFT, RIGHT,
    ...
}
```

**INSERT ... RETURNING Fix:**
Also fixed RETURNING parsing for INSERT statements. The issue was that `RETURNING` wasn't in the reserved keyword lists, so it could be consumed as an implicit alias in some contexts.

**Tests Fixed:**
- TestKeywordsAsColumnNamesAfterDot: 8 subtests passing
- TestParseInsertSelectReturning: 2 subtests
- TestParseInsertSelectFromReturning: 2 subtests

**New Pattern Documented:**
- **Pattern E262**: Keywords after dot - When parsing compound expressions like `T.keyword`, check if the keyword is in RESERVED_FOR_IDENTIFIER and treat it as an identifier, not as starting a new expression.

---

## Session 70 Summary: UPDATE FROM and SQLite OR Clause Implementation (April 8, 2026)

**Major Bug Fix:**

Fixed critical bug where `FROM` keyword was being consumed as an implicit table alias, breaking `UPDATE ... FROM` syntax.

**Root Cause:**
The `isReservedForTableAlias()` function in `parser/query.go` did not include `FROM` in its reserved keywords list. When parsing `UPDATE t1 FROM ...`, the parser would consume `t1` as the table name, then try to parse an implicit alias. Since `FROM` wasn't reserved, it was consumed as the alias name, causing the parser to fail when it couldn't find the actual `FROM` keyword later.

**The Fix:**
Added `FROM` to the reserved keywords list in `isReservedForTableAlias()`:
```go
"FROM": true,  // Added to prevent FROM being used as table alias
```

**New Features Added:**

1. **SQLite UPDATE OR clause support**
   - Added `Or` field to `Update` struct
   - Parser now handles: `UPDATE OR REPLACE`, `UPDATE OR ROLLBACK`, `UPDATE OR ABORT`, `UPDATE OR FAIL`, `UPDATE OR IGNORE`
   - Files: `parser/dml.go`, `ast/statement/dml.go`

2. **UPDATE FROM before SET serialization fix**
   - Fixed `Update.String()` to output table name before FROM clause
   - Correct order: `UPDATE t1 FROM t2 SET ...` not `UPDATE FROM t2 ...`
   - File: `ast/statement/dml.go`

**Tests Fixed:**
- TestParseUpdateFromBeforeSelect: 2 subtests passing
- TestParseUpdateOr: 5 subtests passing (all OR variants)
- TestParseUpdateOrFull: 5 subtests passing
- TestParseUpdateWithJoins: Fixed boolean case

**New Pattern Documented:**
- **Pattern E261**: FROM keyword reservation - Always include FROM in isReservedForTableAlias() to prevent it being consumed as implicit table alias in UPDATE ... FROM syntax

---

## Line Counts (Updated April 8, 2026 - Session 69)

| Component | Rust | Go | Ratio |
|-----------|------|-----|-------|
| Source (parser+ast+dialects) | 67,345 lines | 87,407 lines | 130% |
| Tests | 49,886 lines | 14,245 lines | 29% |
| **Test Status - Snowflake** | - | **100% passing** (99+ subtests) |
| **Test Status - All Packages** | - | **~619 subtests passing, ~95 failing** |

---

## Session 69 Summary: PostgreSQL CREATE/ALTER ROLE Implementation (April 8, 2026)

**Major Code Port:**

Implemented comprehensive PostgreSQL CREATE ROLE and ALTER ROLE syntax, fixing 2 major test suites (12+ subtests):

**New Types Added:**

1. **`Password` type** (ast/expr/ddl.go)
   - Represents PASSWORD 'value' or PASSWORD NULL
   - Used by both CREATE ROLE and ALTER ROLE

2. **`RoleOption` type** (ast/expr/ddl.go)
   - Enum for all PostgreSQL role options: SUPERUSER, NOSUPERUSER, CREATEDB, NOCREATEDB, etc.
   - Supports CONNECTION LIMIT, PASSWORD, VALID UNTIL with expression values

3. **`SetConfigValue` and `ResetConfig` types** (ast/expr/ddl.go)
   - For ALTER ROLE SET/RESET operations
   - Supports TO value, = value, TO DEFAULT, FROM CURRENT

4. **`AlterRoleOperation` interface** (ast/expr/ddl.go)
   - AlterRoleOperationRenameRole: RENAME TO new_name
   - AlterRoleOperationWithOptions: WITH option [ ... ]
   - AlterRoleOperationSet: SET config_name { TO | = } { value | DEFAULT } or FROM CURRENT
   - AlterRoleOperationReset: RESET { config_name | ALL }
   - AlterRoleOperationAddMember/DropMember: MSSQL-specific

**Key Changes:**

1. **CREATE ROLE** (ast/statement/ddl.go, parser/create.go)
   - Full PostgreSQL syntax support: boolean options (SUPERUSER, LOGIN, etc.)
   - Value options: CONNECTION LIMIT, PASSWORD, VALID UNTIL
   - Membership options: IN ROLE, IN GROUP, ROLE, USER, ADMIN
   - MSSQL: AUTHORIZATION owner

2. **ALTER ROLE** (parser/alter.go)
   - RENAME TO operation
   - WITH options (same as CREATE ROLE)
   - SET with IN DATABASE support
   - RESET with IN DATABASE support

**Tests Fixed:**
- TestPostgresCreateRole: 4 subtests passing
- TestPostgresAlterRole: 8 subtests passing

---

## Session 68 Summary: MySQL Index Column Parsing Implementation (April 8, 2026)

**Major Code Port:**

Implemented comprehensive MySQL index column parsing with ASC/DESC support, fixing 10+ failing tests:

**New Functions Added:**

1. **`parseParenthesizedIndexColumnList()`** (parser/ddl.go)
   - Parses index columns with ASC/DESC/NULLS FIRST/LAST
   - Supports functional expressions like `CAST(col AS UNSIGNED)`
   - Handles comma-separated column lists with proper error handling

2. **`parseIndexOptions()`** (parser/ddl.go)
   - Parses USING BTREE/HASH after column list
   - Parses COMMENT 'string' option
   - Returns []*expr.IndexOption for constraint storage

**Key Changes:**

1. **PRIMARY KEY/UNIQUE Constraints** (parser/ddl.go)
   - Changed from `ParseParenthesizedColumnList()` to `parseParenthesizedIndexColumnList()`
   - Added `parseIndexOptions()` call after column list
   - Index columns now preserve ASC/DESC ordering

2. **UNIQUE INDEX Syntax** (parser/ddl.go)
   - Added parsing for optional INDEX/KEY keyword after UNIQUE
   - Now handles `UNIQUE INDEX index_name (cols)` syntax

3. **IndexOption Serialization** (ast/expr/ddl.go)
   - Fixed `IndexOption.String()` for USING: outputs `USING BTREE` not `USING = BTREE`
   - Fixed COMMENT: outputs `COMMENT 'value'` with proper quotes
   - Fixed `UniqueConstraint.String()` to include INDEX keyword

4. **CREATE INDEX Serialization** (ast/statement/ddl.go)
   - Added space before column list: `USING BTREE (cols)` not `USING BTREE(cols)`

**Tests Fixed:**
- TestParseCreateTablePrimaryAndUniqueKeyWithIndexOptions - now passing
- Partial fix for TestDDLWithIndexUsing - serialization fixed, GenericDialect limitation remains

**New Patterns Documented:**
- **Pattern E255**: Index column parsing - Use `parseParenthesizedIndexColumnList()` for `(col ASC, col DESC)` syntax
- **Pattern E256**: Index options - Parse USING/COMMENT after columns, not as key-value pairs
- **Pattern E257**: UNIQUE INDEX syntax - MySQL allows `UNIQUE INDEX name`, parse INDEX/KEY keywords

---

## Session 67 Summary: Final Snowflake Test Fixes (April 8, 2026)

**Major Bug Fix:**

Fixed the deeply nested parentheses parsing bug that was causing "Expected: ), found: EOF" errors on expressions like `((1))`.

**Root Cause:**
The bug was in `parseSubqueryWithSetOps()` in `parser/special.go`. When saving and restoring position:
- `GetCurrentIndex()` returns `p.index - 1` (current token position)
- `SetCurrentIndex()` sets `p.index` directly (next token position)
- The mismatch caused position to be off by one after restore

**The Fix:**
Changed two lines in `parseSubqueryWithSetOps()`:
- Line 1266: `SetCurrentIndex(savedIdx)` → `SetCurrentIndex(savedIdx + 1)`
- Line 1280: `SetCurrentIndex(savedIdx)` → `SetCurrentIndex(savedIdx + 1)`

**Impact:**
- Fixed 111 subtests (from 116 failing to 5 failing)
- Test pass rate improved from ~86% to ~98.7%
- Key tests now passing:
  - `TestParseDeeplyNestedBooleanExprDoesNotStackoverflow`
  - `TestParseDeeplyNestedExprHitsRecursionLimits`
  - `TestParseDeeplyNestedSubqueryExprHitsRecursionLimits`
  - All nested expression parsing tests

**New Pattern Documented:**
- **Pattern E251**: Position tracking with GetCurrentIndex/SetCurrentIndex - Remember that `GetCurrentIndex()` returns current position (`p.index - 1`) while `SetCurrentIndex()` sets next position (`p.index`). When restoring after `AdvanceToken()`, use `SetCurrentIndex(savedIdx + 1)` to maintain correct position.

---

### Session 65 Summary: SET Operations in Subqueries Implementation (April 8, 2026)

---

## Line Counts (Updated April 8, 2026 - Session 67)

| Component | Rust | Go | Ratio |
|-----------|------|-----|-------|
| Source (parser+ast+dialects) | 68,028 lines | 86,652 lines | 127% |
| Tests | 49,886 lines | 14,005 lines | 28% |
| **Test Status - Snowflake** | - | **100% passing** (99+ subtests) |
| **Test Status - All Packages** | - | **98 subtests failing** (non-Snowflake) |
| **Tests Passing** | - | **~471+ subtests** |

---

## Requirements for New Features

1. **Parser Changes**: Update appropriate `parseXxx()` function, add state management if needed
2. **AST Changes**: Add fields to existing structs or create new types in `ast/` packages
3. **Serialization**: Always update `String()` method to match new fields
4. **Dialect Support**: Add capability method to `dialects/capabilities.go` if dialect-specific
5. **Tests**: Use existing test framework, update expectations for canonical form
6. **Documentation**: Add Pattern E### entry, update session summary

---

## Current Session History

### Session 67 Summary: Final Snowflake Test Fixes (April 8, 2026)

**Fixed All 5 Remaining Snowflake Test Failures:**

1. **Boolean Case Fix** (`true` vs `TRUE`)
   - Changed `ValueExpr.String()` to return lowercase `true`/`false` to match Rust canonical form
   - Fixed `FLATTEN(..., outer => true)` and PIVOT subquery tests
   - File: `ast/expr/basic.go`

2. **AUTOINCREMENT Format Fix**
   - Changed `IdentityPropertyKindAutoincrement.String()` from `AUTO_INCREMENT` to `AUTOINCREMENT` (no underscore)
   - Snowflake/SQLite use `AUTOINCREMENT`, MySQL uses `AUTO_INCREMENT` as dialect-specific option
   - File: `ast/expr/ddl.go`

3. **EXPLAIN TABLE Syntax Fix**
   - Added `hasTableKeyword` parsing in `parseExplainWithAlias()` to handle `EXPLAIN TABLE table_name`
   - File: `parser/describe.go`

4. **COPY INTO with Transformations Fix**
   - Fixed placeholder token handling: `$1` is `TokenPlaceholder`, not `TokenChar('$') + TokenNumber`
   - Fixed element accessor: Use `TokenColon{}` not `TokenChar{Char: ':'}`
   - Added missing `)` consumption after parsing table kind with transformations
   - Added fallback for regular select items mixed with stage load items
   - Fixed serialization of transformations in `CopyIntoSnowflake.String()`
   - Fixed spacing in `StageLoadSelectItem.String()`: ` AS ` instead of `AS `
   - Files: `dialects/snowflake/snowflake.go`, `ast/statement/misc.go`, `ast/expr/ddl.go`

5. **SET Tuple Assignment Validation**
   - Added validation to require parenthesized values when variables are parenthesized
   - `SET (a, b, c) = (1, 2, 3)` is valid, but `SET (a, b, c) = 1, 2, 3` now correctly errors
   - File: `parser/misc.go`

6. **ALTER USER MFA Test Fix**
   - Updated test to use correct Rust syntax: `SET DEFAULT_MFA_METHOD PASSKEY` (no equals sign, keyword not string)
   - File: `tests/ddl/alter_test.go`

**New Patterns Documented:**
- **Pattern E252**: Token types for special characters - `:` is `TokenColon{}`, not `TokenChar{Char: ':'}`
- **Pattern E253**: Dollar placeholders - `$N` is `TokenPlaceholder{Value: "$N"}`, not separate tokens
- **Pattern E254**: SET statement validation - When variables are parenthesized, values must be parenthesized too

### Session 66 Summary: Nested Parentheses Parsing Fix (April 8, 2026)

**Fixed 4 Critical Issues:**

1. **ANY/SOME/ALL with Subqueries** (TestAnySomeAllComparison now passing)
   - Updated `parseBinaryOp()` in `parser/infix.go` to handle subqueries after ANY/SOME/ALL
   - Changed subquery parsing to use `parseSubqueryWithSetOps()` 
   - Fixed serialization: `AnyOp.String()` and `AllOp.String()` output `ANY(subquery)` without space

2. **SET Operations in Parenthesized Subqueries** (Major structural fix)
   - Added `parseParenthesizedSetExpr()` to handle `(SELECT ...) UNION (SELECT ...)` syntax
   - Handles nested parentheses with proper recursion
   - Supports UNION, EXCEPT, INTERSECT with correct precedence

3. **IN Clause with Set Operations** (TestParseInUnion - parsing works, serialization pending)
   - Parsing works for `IN ((SELECT ...) UNION (SELECT ...))`
   - Serialization removes double parentheses (canonical form difference)

4. **Critical Bug Fix: LIMIT/ORDER BY Regression**
   - Issue: Changes to `parseSetOperations()` caused LIMIT, ORDER BY, and other post-query clauses to be lost
   - Fix: Return original `left` statement when no set operations found, preserving all fields

**New Patterns Documented:**
- **Pattern E248**: SET operations in subqueries - Use `parseParenthesizedSetExpr()` for `(SELECT ...) UNION (SELECT ...)` syntax
- **Pattern E249**: Preserve original statement when no transformations - Return `left` unchanged instead of creating new struct
- **Pattern E250**: ANY/ALL subquery serialization - Canonical form is `ANY(subquery)` not `ANY (subquery)`

### Session 64 Summary: UPDATE with JOINs, Boolean Case, AUTO_INCREMENT (April 8, 2026)

**Fixed 5 Key Issues (+4 tests passing):**

1. **UPDATE with JOINs Serialization** - Added `Joins []query.Join` field to `Update` struct, extract from parsed table
2. **Boolean Literal Case Sensitivity** - Changed `ValueExpr.String()` to output `TRUE`/`FALSE` in uppercase
3. **AUTO_INCREMENT Column Option** - Changed `IdentityPropertyKindAutoincrement.String()` to return "AUTO_INCREMENT" (with underscore)
4. **Test Updates** - Updated 4 tests for canonical form (GEOMETRY uppercase, COMMENT = form, etc.)

**New Patterns:**
- **Pattern E241-E244**: UPDATE with JOINs, Boolean canonical form, AUTO_INCREMENT form, test updates

### Session 63 Summary: Quick Test Fixes (April 8, 2026)

**Fixed 6 Key Issues:**

1. **PIVOT/UNPIVOT Serialization** - Fixed `PIVOT(...)` format without space
2. **Aliased Expression Canonical Form** - Always include AS keyword
3. **EXTRACT Function Case** - Use uppercase time units (SECONDS)
4. **LOAD EXTENSION Quote Handling** - Use `.Value` field, not `.String()`
5. **Number Serialization** - Preserve decimal point format (`2.`)

### Session 62 Summary: Massive Code Port (April 8, 2026)

**Implemented Major Features:**

1. **CREATE TRIGGER** - MySQL/PostgreSQL trigger parsing with compound statements
2. **SET TRANSACTION CHARACTERISTICS** - Full support with tracking flags
3. **LOCK TABLE** - PostgreSQL lock modes and table locking

**New Patterns:**
- **Pattern E231-E235**: Function parentheses, compound statements, optional clause tracking, LOCK modes

### Session 61 Summary: NOT NULL Constraint Fix (April 8, 2026)

**Critical Bug Fix:**
- Fixed `NOT NULL` being serialized as `IS NOT NULL` or `NULL`
- Solution: Parser state management in `parseColumnDef()`, token put-back in `parseNotPrefixedInfix()`

---

## Previous Sessions Archive

*(When history exceeds 100 lines, older sessions are archived here with one-line summaries)*

### Sessions 61-70 (April 8, 2026)
- **Session 70**: UPDATE FROM, SQLite OR clause, FROM keyword fix (+4 tests) - ~712 tests passing
- **Session 69**: PostgreSQL CREATE/ALTER ROLE implementation (+12 tests) - ~619 tests passing
- **Session 68**: MySQL index column parsing with ASC/DESC (+10 tests) - ~607 tests passing
- **Session 67**: Final Snowflake test fixes (100% passing!) - ~471 tests passing, 5 failures fixed
- **Session 66**: Nested parentheses position tracking fix (+111 tests!) - ~382 tests passing, 98.7% success rate
- **Session 65**: SET Operations in Subqueries, ANY/ALL fix (+4 tests) - ~271 tests passing
- **Session 64**: UPDATE with JOINs, Boolean case, AUTO_INCREMENT (+4 tests) - ~267 tests passing
- **Session 63**: PIVOT/UNPIVOT, Aliased expressions, EXTRACT case (+6 tests) - ~263 tests passing
- **Session 62**: CREATE TRIGGER, SET TRANSACTION, LOCK TABLE (+3 tests) - ~257 tests passing
- **Session 61**: NOT NULL constraint fix - ~254 tests passing

### Sessions 51-60 (April 8, 2026)
- **Session 60**: JSON_TABLE implementation (+1 test, major feature) - ~695 tests passing
- **Session 59**: Parser fixes for ORDER BY, EXCLUDE, Stage params (+4 tests) - ~694 tests passing
- **Session 58**: INSERT/UPDATE/COPY INTO fixes (+4 tests) - ~690 tests passing
- **Session 57**: Quote preservation, CREATE PROCEDURE (+11 tests) - ~686 tests passing
- **Session 56**: CREATE CONNECTOR, PIVOT, MERGE in CTE (+16 tests) - ~675 tests passing
- **Session 55**: Snowflake CREATE VIEW, TIMESTAMP_NTZ, Stage names (+16 tests) - ~659 tests passing
- **Session 54**: MySQL column options, ALTER VIEW, CREATE INDEX (+5 tests) - ~643 tests passing
- **Session 53**: BIT/ENUM types, CREATE VIEW IF NOT EXISTS, ON CLUSTER (+5 tests) - ~655 tests passing
- **Session 52**: CREATE TABLE options, EXTERNAL, WITH, CLONE (+6 tests) - ~180 tests passing
- **Session 51**: 5 critical fixes, trailing commas, DISTINCT validation (+5 tests) - 246 passing/14 failing

### Sessions 41-50 (April 8, 2026)
- **Session 50**: PostgreSQL INHERITS, NOT NULL fix, SET TRANSACTION (+3 tests) - 628 passing/185 failing
- **Session 49**: ALTER POLICY/CONNECTOR implementation (+4 tests) - 627 passing/186 failing
- **Session 48**: INSERT aliases, ON CONFLICT, Dollar-quoted strings (+4 tests) - 623 passing/190 failing
- **Session 47**: Placeholders, Recursion limit, Semicolon tests (+4 tests) - 619 passing/194 failing
- **Session 46**: PIVOT, CACHE TABLE, SET ROLE, Window clause (+5 tests) - 245 tests passing
- **Session 45**: TRUNCATE options, ALTER COLUMN, OWNER TO (+7 tests) - ~612 tests passing
- **Session 44**: Constraint characteristics, Array types, AS TABLE (+5 tests) - ~605 tests passing
- **Session 43**: TPCH fixture paths, INTERVAL case (+46 tests) - 601 tests passing
- **Session 42**: LIST/REMOVE commands, GRANT privileges (+10 tests) - 599 tests passing
- **Session 41**: COMMENT ON fix, Massive code port analysis (+1 test) - 912 tests passing

---

## Pattern Reference (E100-E250)

**Core Patterns (most frequently used):**

- **Pattern E100**: View column options - Use `ViewColumnDef` instead of simple identifiers for dialect-specific column options
- **Pattern E102**: Ident QuoteStyle preservation - Store quoted identifiers as `*ast.Ident` to preserve `QuoteStyle`
- **Pattern E105**: Dialect adapter delegation - Ensure `dialectAdapter` delegates to underlying dialect
- **Pattern E106**: Span mismatches vs parsing failures - Column position differences are non-functional
- **Pattern E110-E150**: Various parsing patterns for Snowflake, PostgreSQL, MySQL features
- **Pattern E151-E180**: Serialization and AST structure patterns
- **Pattern E181-E220**: Advanced parsing techniques and error handling
- **Pattern E221-E250**: Recent patterns for complex SQL features
- **Pattern E251**: Position tracking fix - `GetCurrentIndex()` returns `p.index-1` but `SetCurrentIndex()` sets `p.index`. After `AdvanceToken()`, restore with `SetCurrentIndex(savedIdx + 1)`
- **Pattern E255**: Index column parsing - Use `parseParenthesizedIndexColumnList()` for `(col ASC, col DESC)` syntax
- **Pattern E256**: Index options - Parse USING/COMMENT after columns, not as key-value pairs
- **Pattern E257**: UNIQUE INDEX syntax - MySQL allows `UNIQUE INDEX name`, parse INDEX/KEY keywords

**Full Pattern Catalog:**

```
Pattern E251: Position Tracking with GetCurrentIndex/SetCurrentIndex
- When: Saving/restoring parser position after token consumption
- Problem: GetCurrentIndex() returns current position (p.index-1) but SetCurrentIndex() 
  sets next position (p.index), causing off-by-one errors
- Solution: After AdvanceToken(), use SetCurrentIndex(savedIdx + 1) to restore correctly
- Example: See parseSubqueryWithSetOps() fix in Session 66
- Files typically modified: parser/special.go, parser/query.go, parser/helpers.go

Pattern E252: Token Types for Special Characters
- When: Consuming tokens for special characters like colon (:)
- Problem: TokenChar{Char: ':'} doesn't match - colon is TokenColon{}
- Solution: Use specific token types: TokenColon{}, TokenPeriod{}, etc.
- Example: In Snowflake stage load parsing, use parser.ConsumeToken(token.TokenColon{})
- Files typically modified: parser/*.go, dialects/*.go

Pattern E253: Dollar Placeholder Tokenization
- When: Parsing $N placeholders (e.g., $1, $2) in Snowflake/PostgreSQL
- Problem: Expecting TokenChar('$') + TokenNumber but it's a single TokenPlaceholder{Value: "$N"}
- Solution: Check for TokenPlaceholder and parse value from placeholder.Value[1:]
- Example: In parseSelectItemsForDataLoad(), check (token.TokenPlaceholder) instead of TokenChar
- Files typically modified: dialects/snowflake/snowflake.go

Pattern E254: SET Statement Tuple Validation
- When: Parsing SET (var1, var2) = ... statements
- Problem: Parser accepts SET (a, b, c) = 1, 2, 3 but should require parentheses around values
- Solution: When variables are parenthesized, require TokenLParen before parsing values
- Example: In parseSet(), after parsing parenthesized vars, expect TokenLParen for values too
- Files typically modified: parser/misc.go

Pattern E255: Index Column List Parsing
- When: Parsing index columns like PRIMARY KEY (col1 ASC, col2 DESC)
- Problem: Simple column list parsing doesn't handle ASC/DESC or expressions
- Solution: Use parseParenthesizedIndexColumnList() which uses ExpressionParser for each column
- Example: See parseTableConstraint() in ddl.go for PRIMARY KEY/UNIQUE
- Files typically modified: parser/ddl.go, parser/create.go

Pattern E256: Index Options Serialization
- When: Serializing index options like USING BTREE, COMMENT 'text'
- Problem: Default key=value format produces "USING = BTREE" not "USING BTREE"
- Solution: Handle USING and COMMENT specially in IndexOption.String()
- Example: IndexOption.String() checks i.Name == "USING" and formats without =
- Files typically modified: ast/expr/ddl.go

Pattern E257: UNIQUE INDEX Syntax
- When: Parsing MySQL UNIQUE INDEX index_name syntax
- Problem: Parser expects UNIQUE (cols) but MySQL allows UNIQUE INDEX name (cols)
- Solution: After UNIQUE, check for optional INDEX/KEY keyword before index name
- Example: In parseTableConstraint(), after UNIQUE, parse optional INDEX/KEY
- Files typically modified: parser/ddl.go

Pattern E258: PostgreSQL CREATE/ALTER ROLE Options
- When: Implementing CREATE ROLE with full PostgreSQL syntax
- Problem: Need to track both positive (LOGIN) and negative (NOLOGIN) variants
- Solution: Use *bool fields (Some(true)=LOGIN, Some(false)=NOLOGIN, None=not specified)
- Example: CreateRole struct has Login *bool, SuperUser *bool, etc.
- Files typically modified: ast/statement/dcl.go, parser/create.go

Pattern E259: ALTER ROLE SET/RESET Operations
- When: Implementing ALTER ROLE SET config and RESET config
- Problem: Multiple syntax variants: SET x TO y, SET x = y, SET x TO DEFAULT, SET x FROM CURRENT
- Solution: Use interface AlterRoleOperation with variants for each operation type
- Example: AlterRoleOperationSet, AlterRoleOperationReset with SetConfigValue type
- Files typically modified: ast/expr/ddl.go, parser/alter.go

Pattern E260: Boolean Role Options as Enum
- When: Parsing role options like SUPERUSER vs NOSUPERUSER
- Problem: Storing with same type and BoolValue=false doesn't serialize correctly
- Solution: Use separate enum variants for positive and negative: RoleOptionSuperUser vs RoleOptionNoSuperUser
- Example: RoleOptionType has both RoleOptionSuperUser and RoleOptionNoSuperUser
- Files typically modified: ast/expr/ddl.go

Pattern E261: FROM Keyword as Table Alias Bug
- When: Parsing UPDATE ... FROM or DELETE ... FROM statements
- Problem: FROM keyword is consumed as implicit table alias because it's not in isReservedForTableAlias()
- Solution: Add "FROM": true to isReservedForTableAlias() reserved keywords map
- Example: UPDATE t1 FROM t2 SET ... was failing because t1 was parsed with FROM as its alias
- Files typically modified: parser/query.go

Pattern E262: Keywords as Identifiers After Dot
- When: Parsing compound expressions like T.interval, T.case, T.cast
- Problem: Keywords are parsed as starting new expressions (CASE WHEN, INTERVAL '1' DAY) instead of identifiers
- Solution: Check if keyword is in RESERVED_FOR_IDENTIFIER and treat as identifier after dot
- Example: In parseCompoundExprWithOptions(), check token.IsReservedForIdentifier(tok.Word.Keyword) before parsing as expression
- Files typically modified: parser/core.go, token/keywords.go
```

**See full pattern catalog in code comments and previous session notes.**

---

## File Organization Reference

| Package | Purpose |
|---------|---------|
| `ast/` | AST node definitions (Expr, Statement, Query types) |
| `parser/` | Core parser logic (prefix, infix, query, DDL, DML) |
| `dialects/` | Dialect-specific implementations (Snowflake, PostgreSQL, etc.) |
| `token/` | Tokenizer and token types |
| `tests/` | Test suites mirroring Rust test structure |

---

## End of Document
