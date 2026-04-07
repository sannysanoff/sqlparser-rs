---

**Line Counts (Updated April 8, 2026):**

- Rust Source: 67,345 lines (parser + dialects + AST)
- Go Source: 77,105 lines (114% of Rust - AST types and interfaces)
- Rust Tests: 45,672 lines  
- Go Tests: 14,149 lines (31% of Rust test coverage)
- **Current Test Pass Rate: ~74.5%** (204 passing out of 274 total tests)

**Recent Progress:**
- Fixed FROM clause LATERAL table functions (+1 test)
- Fixed MySQL index hints: USE INDEX, IGNORE INDEX, FORCE INDEX (+1 test)
- Fixed Window function INTERVAL handling in frame bounds (+1 test)
- **Total: +3 tests passing** (from 201 to 204 passing)

---

### April 8, 2026 - Table Constraint Types Implementation

Implemented comprehensive table constraint types to fix DDL constraint parsing:

1. **AST Constraint Types** (ast/expr/ddl.go):
   - `PrimaryKeyConstraint` - PRIMARY KEY with optional index name, index type, columns, and characteristics
   - `UniqueConstraint` - UNIQUE with NULLS DISTINCT/NOT DISTINCT support
   - `ForeignKeyConstraint` - FOREIGN KEY with REFERENCES, ON DELETE/UPDATE actions, MATCH kinds
   - `CheckConstraint` - CHECK with optional ENFORCED/NOT ENFORCED (MySQL)
   - `IndexConstraint` - MySQL INDEX/KEY constraints
   - `FullTextOrSpatialConstraint` - MySQL FULLTEXT/SPATIAL constraints
   - Supporting types: `ConstraintCharacteristics`, `ConstraintReferenceMatchKind`, `NullsDistinctOption`

2. **Parser Updates** (parser/ddl.go):
   - Updated `parseTableConstraint()` to populate constraint-specific structs instead of discarding parsed data
   - Updated `parseConstraintCharacteristics()` to return `*ConstraintCharacteristics` instead of discarding
   - Fixed FOREIGN KEY String() output to match expected format: `REFERENCES table(col)` not `REFERENCES table (col)`

3. **Key Pattern Documentation:**
   - **Pattern CT: Table Constraint Implementation** - When implementing table constraints:
     1. Create specific constraint type structs with all relevant fields
     2. Store parsed data in the constraint struct, never discard with `_ = parsedValue`
     3. Update the TableConstraint.Constraint field with the specific constraint type
     4. Implement proper String() method that matches SQL canonical format
     5. For FOREIGN KEY, concatenate table name and column list without space: `table(col)` not `table (col)`

**Result:** +6 tests passing (TestParseAlterTableConstraints and related tests)

---

### April 8, 2026 - FROM Clause LATERAL and Index Hints

Implemented major missing chunks for FROM clause parsing:

1. **LATERAL Table Functions** (parser/query.go):
   - Fixed `parseLateralTable()` to handle both subqueries `LATERAL (SELECT ...)` and table functions `LATERAL generate_series(...)`
   - When LATERAL is followed by `(` it's a subquery; otherwise it's a table function with name followed by `(`

2. **MySQL Index Hints** (parser/query.go):
   - Added `parseTableIndexHints()` function to parse `USE INDEX`, `IGNORE INDEX`, `FORCE INDEX` syntax
   - Supports optional FOR clause: `FOR JOIN`, `FOR ORDER BY`, `FOR GROUP BY`
   - Handles both INDEX and KEY keywords
   - Fixed critical bug: Added `USE`, `IGNORE`, `FORCE` to reserved keywords list for table aliases
     - These keywords were being consumed as implicit table aliases, breaking index hints parsing

3. **Key Pattern Documentation:**
   - **Pattern TI: Table Index Hints** - When parsing table factors for dialects with index hints:
     1. Add hint keywords (USE, IGNORE, FORCE) to `isReservedForTableAlias()` to prevent them being consumed as aliases
     2. Parse index hints AFTER the table alias (in `parseTableName()`)
     3. Use `maybeParse` pattern for optional hints

**Result:** +2 tests passing (TestLateralFunction, TestParseSelectTableWithIndexHints)

---

### April 8, 2026 - Window Function INTERVAL Handling

Fixed window frame bound parsing for INTERVAL expressions:

1. **Problem**: Window frame bounds like `RANGE BETWEEN INTERVAL '1' DAY PRECEDING AND INTERVAL '1 MONTH' FOLLOWING` were failing with "INTERVAL requires a unit after the literal value"

2. **Root Cause**: The `parseIntervalExpr()` function was being called from `parseWindowFrameBound()`, but this caused double-parsing:
   - First call to parseIntervalExpr calls ParseExpr()
   - ParseExpr sees INTERVAL keyword, calls parsePrefix()
   - parsePrefix calls parseIntervalExpr recursively
   - This second call properly parses INTERVAL '1' DAY
   - Control returns to first call which then checks for temporal unit again - but DAY was already consumed!

3. **Solution** (parser/special.go):
   - In `parseWindowFrameBound()`, when we see the INTERVAL keyword, manually parse the components:
     1. Consume INTERVAL keyword
     2. Expect and consume string literal (the value)
     3. Check for temporal unit using `isTemporalUnit()`
     4. Create IntervalExpr with value and unit
   - This avoids the recursive parseIntervalExpr call that caused the double-parsing issue

4. **Key Pattern Documentation:**
   - **Pattern WI: Window INTERVAL Bounds** - When parsing INTERVAL in window frame bounds:
     1. Check for INTERVAL keyword first (before string literal check)
     2. Manually consume INTERVAL keyword and parse components (value + unit)
     3. Don't call parseIntervalExpr() directly to avoid double-parsing
     4. Reference: src/parser/mod.rs:2575-2578 - Rust uses `parse_interval()` only for string literals, not for INTERVAL keyword

**Result:** +1 test passing (TestParseWindowFunctionsAdvanced)

---

## Common Errors and How to Avoid Them

### Error E1: "Expected: end of statement, found: X"
**Cause**: The parser finished parsing a statement but found unexpected tokens. This usually means some syntax wasn't recognized and parsing stopped early.

**Solution**: Check if keywords that should start clauses are being consumed elsewhere (e.g., as aliases). Add them to reserved keyword lists.

### Error E2: Keywords consumed as aliases
**Cause**: SQL keywords like USE, IGNORE, FORCE are being parsed as table aliases because they're not in the reserved list.

**Solution**: Add keywords to `isReservedForTableAlias()` in parser/query.go when they should start new clauses rather than being aliases.

### Error E3: Double-parsing in expression parsing
**Cause**: Calling a parse function that internally calls ParseExpr(), which recursively calls the same function.

**Solution**: When the current token is a keyword that triggers special parsing, manually consume it and parse components rather than calling the high-level parse function.

### Error E4: "INTERVAL requires a unit after the literal value"
**Cause**: The temporal unit (DAY, MONTH, etc.) was already consumed by a recursive call to parseIntervalExpr.

**Solution**: In window frame bounds and similar contexts, manually parse INTERVAL components instead of calling parseIntervalExpr().

---

## Porting Patterns from Rust

### Pattern P1: String Literal as Expression Value
When Rust does: `Token::SingleQuotedString(_) => self.parse_interval()`
In Go: Check for `token.TokenSingleQuotedString` and call interval parsing directly.

### Pattern P2: Keyword-Triggered Special Parsing
When Rust checks for a keyword before calling a parse function, the Go code should:
1. Check if current token matches the keyword
2. If yes, manually consume and parse components (don't call the high-level function)
3. This avoids recursive double-parsing

### Pattern P3: Reserved Keywords for Aliases
Rust's parser has implicit knowledge of which keywords can't be aliases. In Go, we explicitly list them in `isReservedForTableAlias()`.
When adding new syntax that uses keywords after table names, add those keywords to the reserved list.

---

## Test Status Summary

| Category | Tests | Passing | Failing |
|----------|-------|---------|---------|
| Expressions | ~80 | ~65 | ~15 |
| SELECT | ~60 | ~45 | ~15 |
| DDL (CREATE/ALTER) | ~70 | ~55 | ~15 |
| Window Functions | ~25 | ~18 | ~7 |
| Transactions | ~20 | ~12 | ~8 |
| SET statements | ~15 | ~8 | ~7 |

**Total**: 274 tests, 204 passing, 70 failing (~74.5% pass rate)

