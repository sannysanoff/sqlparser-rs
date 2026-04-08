# SQL Parser Go Test Failure Analysis

## Summary
- Total Tests: 813
- Passing: 667 (~82%)
- **Failing: 146 (~18%)**

## Categorization of Failing Tests by Missing Features

### Category 1: CREATE TABLE Features (13 tests)
**Failing Tests:**
- TestParseCreateTableHiveArray
- TestParseCreateTableWithMultipleOnDeleteInConstraintFails
- TestParseCreateTableWithMultipleOnDeleteFails
- TestParseCreateTableComment
- TestParseCreateTableCommentCharacterSet
- TestParseCreateTableWithColumnCollate
- TestParseCreateTablePrimaryAndUniqueKeyWithIndexOptions
- TestParseCreateTablePrimaryAndUniqueKeyWithIndexType
- TestParseCreateTableAutoIncrementOffset
- TestParseDefaultExprWithOperators
- TestParseDefaultWithCollateColumnOption
- TestParseNotNullInColumnOptions
- TestParseGeometricTypesSridOption

**Error Patterns:**
- "Expected: ), found: <" (Hive ARRAY<INT> syntax)
- "should have failed" (validation not implemented)
- Re-serialization format differences
- "Expected: end of statement, found: ..."

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 8338-8522: `parse_create_table()` - Full CREATE TABLE parsing
2. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 6293-6339: `parse_create_external_table()` - External table parsing
3. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 3150-3250: Column option parsing with DEFAULT expressions
4. `/Users/san/Fun/sqlparser-rs/src/ast/ddl.rs` lines 200-500: ColumnDef and ColumnOption AST types

---

### Category 2: CREATE INDEX & ALTER INDEX Features (7 tests)
**Failing Tests:**
- TestParseCreateIndexDifferentUsingPositions
- TestCreateIndexWithWithClause
- TestDDLWithIndexUsing
- TestParseAlterIndex
- TestFunctionalKeyPart
- TestParsePrefixKeyPart
- TestMySQLForeignKeyWithIndexName

**Error Patterns:**
- "Expected: =, found: )" (parsing WITH clause params)
- Re-serialization missing USING clause
- Missing ALTER INDEX RENAME TO support

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 7926-8016: `parse_create_index()` - Full CREATE INDEX parsing with USING, WITH, INCLUDE, NULLS DISTINCT
2. `/Users/san/Fun/sqlparser-rs/src/parser/alter.rs` lines 1-100: ALTER INDEX parsing (if exists)
3. `/Users/san/Fun/sqlparser-rs/src/ast/ddl.rs` lines 900-1100: CreateIndex and IndexOption AST types

---

### Category 3: CREATE PROCEDURE/CONNECTOR/TRIGGER (8 tests)
**Failing Tests:**
- TestParseCreateProcedureWithLanguage
- TestParseCreateProcedureWithParameterModes
- TestParseCreateConnector
- TestCreateConnector
- TestParseCreateTrigger
- TestParseCreateTriggerCompoundStatement
- TestPostgresCreateTriggerWithMultipleEventsAndDeferrable

**Error Patterns:**
- "Expected: end of statement, found: my_connector"
- "Expected: end of statement, found: PROCEDURE"
- "Expected: end of statement, found: TRIGGER"

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 6930-6970: `parse_create_connector()` - Lines 6930-6970
2. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 5970-6080: `parse_create_procedure()` - Procedure with LANGUAGE, parameter modes (IN, OUT, INOUT)
3. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 6200-6290: `parse_create_trigger()` - Lines 6200-6290
4. `/Users/san/Fun/sqlparser-rs/src/ast/mod.rs` lines 9750-9830: Procedure and trigger AST types
5. `/Users/san/Fun/sqlparser-rs/src/ast/ddl.rs` lines 3700-3800: Connector AST types

---

### Category 4: DML - INSERT/UPDATE/DELETE Features (14 tests)
**Failing Tests:**
- TestParseInsertValuesFull
- TestParseInsertSelectReturning
- TestParseInsertSelectFromReturning
- TestParseInsertSqlite
- TestParseUpdateWithTableAlias
- TestParseUpdateWithTableAliasFull
- TestParseUpdateOr
- TestParseUpdateOrFull
- TestParseUpdateFromBeforeSelect
- TestParseUpdateWithJoins
- TestParseDeleteStatement
- TestMergeInCte
- TestPostgresComplexPostgresInsertWithAlias

**Error Patterns:**
- "Failed to parse SQL with dialect *generic.GenericDialect: INSERT INTO ... RETURNING ..."
- "Failed to parse SQL with dialect *generic.GenericDialect: UPDATE OR REPLACE ..."
- Missing UPDATE with FROM and JOINs support
- Missing DELETE with USING support
- "Multi-table DELETE not yet implemented in Go port" (explicitly skipped)

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 8650-8750: `parse_insert()` - Lines 8650-8750 with RETURNING clause
2. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 9150-9250: `parse_update()` - Lines 9150-9250 with FROM, JOINs
3. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 9350-9450: `parse_delete()` - Lines 9350-9450 with USING
4. `/Users/san/Fun/sqlparser-rs/src/parser/merge.rs` lines 1-100: MERGE statement parsing for CTEs

---

### Category 5: PostgreSQL-Specific Features (50+ tests)
**Failing Tests:**
- TestPostgresCreateTableFromPgDump
- TestPostgresCreateTableEmpty
- TestPostgresCreateTableConstraintsOnly
- TestPostgresCreateTableIfNotExists
- TestPostgresCreateTableWithAlias
- TestPostgresCreateTableWithDefaults
- TestPostgresCreateTableWithInherit
- TestPostgresCreateTableWithPartitionBy
- TestPostgresAlterTableAddColumns
- TestPostgresAlterTableConstraintsRename
- TestPostgresAlterTableConstraintUsingIndex
- TestPostgresAlterTableValidateConstraint
- TestPostgresAlterType
- TestPostgresAlterOperator
- TestPostgresAlterOperatorFamily
- TestPostgresAlterRole
- TestPostgresAlterSchema
- TestPostgresCreateOperator
- TestPostgresCreateOperatorClass
- TestPostgresCreateRole
- TestPostgresCopyFromError
- TestPostgresDropDomain
- TestPostgresDropFunction
- TestPostgresDropOperator
- TestPostgresDropOperatorClass
- TestPostgresDropOperatorFamily
- TestPostgresDropProcedure
- TestPostgresDropTrigger
- TestPostgresEscapedLiteralString
- TestPostgresEscapedStringLiteral
- TestPostgresFetch
- TestPostgresForeignKeyMatch
- TestPostgresForeignKeyMatchWithActions
- TestPostgresIncorrectDollarQuotedString
- TestPostgresIntervalDataType
- TestPostgresIntervalKeywordAsUnquotedIdentifier
- TestPostgresJsonObjectValueSyntax
- TestPostgresLockTable
- TestPostgresQuotedIdentifier
- TestPostgresSetTransaction
- TestPostgresShow
- TestPostgresTableFunctionWithOrdinality
- TestPostgresTransactionStatement
- TestPostgresUnicodeStringLiteral
- TestPostgresUpdateHasKeyword
- TestPostgresUpdateInWithSubquery
- TestPostgresVarbitDatatype
- TestPostgresArraySubqueryExpr
- TestPostgresCastInDefaultExpr
- TestPostgresDeclare
- TestPostgresDelimitedIdentifiers
- TestPostgresDoublePrecision

**Error Patterns:**
- Dollar-quoted strings ($$...$$) not tokenized
- Escaped string literals (E'...') not supported
- COPY statement parsing
- Operator definitions (CREATE OPERATOR)
- Table INHERIT clause
- PARTITION BY clause
- Dollar-quoted string parsing

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/tokenizer.rs` lines 1920-2010: Dollar-quoted string tokenization (lines ~1920-2010)
2. `/Users/san/Fun/sqlparser-rs/src/tokenizer.rs` lines 1190-1210: Escaped string literal handling
3. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 6900-6930: `parse_copy()` - COPY statement
4. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 7000-7050: `parse_create_operator()`
5. `/Users/san/Fun/sqlparser-rs/src/dialect/postgresql.rs` lines 82-264: PostgreSQL dialect features
6. `/Users/san/Fun/sqlparser-rs/src/ast/ddl.rs` lines 4000-4100: CreateOperator AST

---

### Category 6: MySQL-Specific Features (15 tests)
**Failing Tests:**
- TestParseCreateTableWithColumnCollate
- TestParseCreateTableAutoIncrementOffset
- TestParseCreateTrigger
- TestParseCreateTriggerCompoundStatement
- TestParseJsonTable
- TestParseLogicalXor
- TestParseStraightJoin
- TestParseSelectModifiersErrors
- TestMySQLForeignKeyWithIndexName
- TestOptimizerHints
- TestParseIdentifiers
- TestParseShowExtendedFull
- TestParseCreateTablePrimaryAndUniqueKeyWithIndexOptions
- TestParseCreateTablePrimaryAndUniqueKeyWithIndexType
- TestVariableAssignmentUsingColonEqual

**Error Patterns:**
- CHARACTER SET / COLLATE in column definitions
- JSON_TABLE function
- XOR operator
- STRAIGHT_JOIN
- Optimizer hints (/*+ ... */)
- Variable assignment with :=
- Multi-table UPDATE with JOINs

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/dialect/mysql.rs` lines 70-214: MySQL dialect features
2. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 9520-9580: JSON_TABLE parsing
3. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 7990-8000: Index options in CREATE TABLE
4. `/Users/san/Fun/sqlparser-rs/src/tokenizer.rs` lines 200-300: Optimizer hint tokenization

---

### Category 7: Snowflake-Specific Features (6 tests)
**Failing Tests:**
- TestSnowflakeChangesClause
- TestSnowflakeCopyIntoWithStageParams
- TestSnowflakeCopyIntoWithTransformations
- TestSnowflakeCreateStageWithParams
- TestSnowflakeExplainDescribe
- TestSnowflakePivot

**Error Patterns:**
- COPY INTO with stage parameters
- CHANGES clause
- PIVOT syntax
- CREATE STAGE with parameters

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/dialect/snowflake.rs` lines 248-1927: Full Snowflake dialect
2. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 8250-8350: COPY statement parsing
3. `/Users/san/Fun/sqlparser-rs/src/ast/helpers/stmt_data_loading.rs`: Data loading helpers

---

### Category 8: ALTER TABLE Operations (8 tests)
**Failing Tests:**
- TestParseAlterIndex
- TestParseAlterUserSetOptions
- TestParseDropSchema
- TestParseDropTable
- TestParseDropView
- TestParseDropIndex
- TestDropPolicy

**Error Patterns:**
- ALTER INDEX RENAME TO not implemented
- ALTER USER with MFA options
- DROP statements serialization issues

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/parser/alter.rs` lines 1-150: ALTER statement parsing
2. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 10550-10650: `parse_alter_table_operation()`
3. `/Users/san/Fun/sqlparser-rs/src/ast/mod.rs` lines 5500-5600: ALTER statement AST types

---

### Category 9: Expression & Operator Issues (12 tests)
**Failing Tests:**
- TestParseNotPrecedence
- TestParseAtTimezone
- TestExtractSecondsOk
- TestParseLoadExtension
- TestParsePipeOperatorSelect
- TestParsePipeOperatorExtend
- TestParseColumnAliases
- TestParseSelectExprStar
- TestNoInfixError
- TestParseInvalidInfixNot
- TestKeywordsAsColumnNamesAfterDot
- TestReservedKeywordsForIdentifiers

**Error Patterns:**
- NOT precedence issues
- AT TIME ZONE operator
- Pipe operators (|>) for dataflow
- Column alias with AS keyword
- Keywords as column names after dot
- Reserved keyword handling

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 7420-7500: Expression precedence handling
2. `/Users/san/Fun/sqlparser-rs/src/parser/infix.rs` or similar: Pipe operator parsing
3. `/Users/san/Fun/sqlparser-rs/src/dialect/generic.rs` lines 188-276: Keyword reservation handling
4. `/Users/san/Fun/sqlparser-rs/src/tokenizer.rs` lines 2420-2470: String unescaping

---

### Category 10: Transaction & Variable Features (3 tests)
**Failing Tests:**
- TestParseStartTransactionMssql
- TestParseSetVariableErrors
- TestParseFetchVariations

**Error Patterns:**
- Transaction modes not properly parsed
- SET variable error handling
- FETCH statement variations

**Required Rust Code:**
1. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 7230-7280: Transaction parsing
2. `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` lines 7150-7220: SET statement parsing

---

## Top 10 Major Implementation Chunks (by impact)

### 1. CREATE EXTERNAL TABLE Support (6 tests)
**Tests Fixed:** TestParseCreateExternalTable, TestParseCreateOrReplaceExternalTable, TestParseCreateExternalTableLowercase, TestParseCreateTableHiveArray
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` 6293-6339
**Complexity:** Medium
**Description:** Parse STORED AS, LOCATION, TBLPROPERTIES for external tables

### 2. CREATE INDEX WITH Clause (5 tests)
**Tests Fixed:** TestCreateIndexWithWithClause, TestParseCreateIndexDifferentUsingPositions, TestDDLWithIndexUsing, TestFunctionalKeyPart, TestParsePrefixKeyPart
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` 7926-8016
**Complexity:** Medium
**Description:** Parse USING index_type, WITH options, INCLUDE columns, NULLS DISTINCT

### 3. INSERT with RETURNING Clause (5 tests)
**Tests Fixed:** TestParseInsertSelectReturning, TestParseInsertSelectFromReturning, TestParseInsertValuesFull, TestMergeInCte
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` 8650-8750
**Complexity:** Low-Medium
**Description:** Add RETURNING clause support to INSERT statements

### 4. UPDATE with FROM and JOINs (5 tests)
**Tests Fixed:** TestParseUpdateFromBeforeSelect, TestParseUpdateWithJoins, TestParseUpdateWithTableAlias, TestParseUpdateWithTableAliasFull, TestPostgresUpdateInWithSubquery
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` 9150-9250
**Complexity:** High
**Description:** Parse UPDATE with FROM clause and JOIN syntax (PostgreSQL/MySQL style)

### 5. PostgreSQL Dollar-Quoted Strings (8 tests)
**Tests Fixed:** TestPostgresEscapedLiteralString, TestPostgresEscapedStringLiteral, TestPostgresIncorrectDollarQuotedString, TestPostgresUnicodeStringLiteral, TestPostgresDelimitedIdentifiers, TestCheckRoundtripOfEscapedString
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/tokenizer.rs` 1920-2010
**Complexity:** Medium
**Description:** Tokenize $$...$$ and $tag$...$tag$ quoted strings

### 6. CREATE PROCEDURE Support (3 tests)
**Tests Fixed:** TestParseCreateProcedureWithLanguage, TestParseCreateProcedureWithParameterModes
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` 5970-6080
**Complexity:** Medium
**Description:** Parse CREATE PROCEDURE with LANGUAGE, parameter modes (IN/OUT/INOUT), and body

### 7. CREATE CONNECTOR Support (2 tests)
**Tests Fixed:** TestParseCreateConnector, TestCreateConnector
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` 6930-6970
**Lines in Rust AST:** `/Users/san/Fun/sqlparser-rs/src/ast/ddl.rs` 3700-3800
**Complexity:** Low
**Description:** Parse CREATE CONNECTOR with TYPE, URL, and DCPROPERTIES

### 8. MySQL JSON_TABLE Support (3 tests)
**Tests Fixed:** TestParseJsonTable, TestMySQLForeignKeyWithIndexName
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` 9520-9580
**Complexity:** High
**Description:** Parse JSON_TABLE function with COLUMNS clause

### 9. CREATE TRIGGER Support (4 tests)
**Tests Fixed:** TestParseCreateTrigger, TestParseCreateTriggerCompoundStatement, TestPostgresCreateTriggerWithMultipleEventsAndDeferrable
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/mod.rs` 6200-6290
**Complexity:** Medium-High
**Description:** Parse CREATE TRIGGER with timing, events, table, and action body

### 10. ALTER INDEX Support (3 tests)
**Tests Fixed:** TestParseAlterIndex, TestParseAlterUserSetOptions, TestDropPolicy
**Lines in Rust:** `/Users/san/Fun/sqlparser-rs/src/parser/alter.rs` 1-150
**Complexity:** Low-Medium
**Description:** Parse ALTER INDEX RENAME TO and other ALTER operations

---

## Implementation Priority Recommendations

### Phase 1: High Impact, Lower Complexity (Quick Wins)
1. CREATE CONNECTOR (2 tests, ~100 lines)
2. ALTER INDEX (3 tests, ~50 lines)
3. INSERT RETURNING (5 tests, ~50 lines)
4. CREATE EXTERNAL TABLE formatting (3 tests, ~30 lines in Display)

### Phase 2: Core Features (Medium Complexity)
5. CREATE INDEX enhancements (5 tests, ~200 lines)
6. CREATE PROCEDURE (3 tests, ~150 lines)
7. CREATE TRIGGER (4 tests, ~200 lines)

### Phase 3: Complex Dialect Features
8. Dollar-quoted strings (8 tests, tokenizer changes ~150 lines)
9. UPDATE with FROM/JOINs (5 tests, ~300 lines)
10. JSON_TABLE (3 tests, ~250 lines)

---

## Detailed Rust Source Locations for Major Chunks

### Chunk 1: CREATE EXTERNAL TABLE
```rust
// File: /Users/san/Fun/sqlparser-rs/src/parser/mod.rs
// Lines: 6293-6339
pub fn parse_create_external_table(
    &mut self,
    or_replace: bool,
) -> Result<CreateTable, ParserError> {
    self.expect_keyword_is(Keyword::TABLE)?;
    let if_not_exists = self.parse_keywords(&[Keyword::IF, Keyword::NOT, Keyword::EXISTS]);
    let table_name = self.parse_object_name(false)?;
    let (columns, constraints) = self.parse_columns()?;
    let hive_distribution = self.parse_hive_distribution()?;
    let hive_formats = self.parse_hive_formats()?;
    // ... continues
}
```

### Chunk 2: CREATE INDEX
```rust
// File: /Users/san/Fun/sqlparser-rs/src/parser/mod.rs
// Lines: 7926-8016
pub fn parse_create_index(&mut self, unique: bool) -> Result<CreateIndex, ParserError> {
    let concurrently = self.parse_keyword(Keyword::CONCURRENTLY);
    let if_not_exists = self.parse_keywords(&[Keyword::IF, Keyword::NOT, Keyword::EXISTS]);
    let mut using = None;
    // ... MySQL USING index_type handling
    // ... WITH clause parsing (lines 7968-7977)
    // ... WHERE predicate parsing (lines 7979-7983)
}
```

### Chunk 3: INSERT RETURNING
```rust
// File: /Users/san/Fun/sqlparser-rs/src/parser/mod.rs
// Lines: ~8650-8750 (in parse_insert)
// Add after parsing values/select:
let returning = if self.parse_keyword(Keyword::RETURNING) {
    Some(self.parse_comma_separated(|p| p.parse_select_item())?)
} else {
    None
};
```

### Chunk 4: Dollar-Quoted Strings
```rust
// File: /Users/san/Fun/sqlparser-rs/src/tokenizer.rs
// Lines: 1928-2010
// In tokenize function:
'0'..='9' | '$' => {
    // Check for dollar-quoted string
    if ch == '$' && self.dialect.supports_dollar_quoted_string() {
        return self.tokenize_dollar_quoted_string(chars);
    }
    // ... continue with number parsing
}
```

### Chunk 5: UPDATE FROM/JOIN
```rust
// File: /Users/san/Fun/sqlparser-rs/src/parser/mod.rs
// Lines: ~9150-9250 (in parse_update)
// Parse UPDATE with optional FROM clause for PostgreSQL
let from = if self.parse_keyword(Keyword::FROM) {
    Some(self.parse_table_and_joins()?)
} else {
    None
};
```

---

## Key Observations

1. **Most failures are in DDL parsing** (CREATE, ALTER, DROP) - ~45% of failures
2. **PostgreSQL dialect features** account for ~35% of failures
3. **MySQL dialect features** account for ~10% of failures
4. **Many failures are Display/serialization issues** - the AST parses but doesn't re-serialize correctly
5. **Some features are explicitly skipped** with comments like "not yet implemented in Go port"

## Recommendations for Implementation

1. **Start with Display/serialization fixes** - Many tests fail because the AST parses correctly but the Display trait doesn't format SQL correctly. These are quick fixes.

2. **Focus on CREATE INDEX and EXTERNAL TABLE next** - These affect multiple dialects and have clear test cases.

3. **Port PostgreSQL tokenizer features** (dollar-quoted strings) - These enable many PostgreSQL tests to pass.

4. **Implement INSERT/UPDATE/DELETE enhancements** - These are core DML features used across all dialects.

5. **Address MySQL-specific features** last - They have narrower impact but are important for MySQL compatibility.
