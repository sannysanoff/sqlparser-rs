# MySQL Dialect Tests - Porting Status Report

## Summary
- **Total Tests**: 130
- **Passing**: 16 (12.3%)
- **Failing**: 114 (87.7%)
- **Target**: 52 tests passing (40%)
- **Gap**: Need 36 more tests to pass

## Test Files
1. `go/tests/mysql/mysql_test.go` - 51 tests
2. `go/tests/mysql/mysql_batch2_test.go` - 79 tests

## Passing Tests (16)
1. TestParseRlikeAndRegexp
2. TestParseCreateTableConstraintCheckWithoutName
3. TestParseCreateTableWithFulltextDefinitionShouldNotAcceptConstraintName
4. TestParseLongblobType
5. TestParseCastArray
6. TestParseJsonMemberOf
7. TestParseCreateDatabaseWithCharsetErrors
8. TestParseLiteralString
9. TestParseEscapedQuoteIdentifiersWithEscape
10. TestParseEscapedQuoteIdentifiersWithNoEscape
11. TestParseEscapedBackticksWithEscape
12. TestParseEscapedBackticksWithNoEscape
13. TestParseUnterminatedEscape
14. TestParseSimpleInsert
15. TestParseQualifiedIdentifiersWithNumericPrefix
16. TestParseInsertWithNumericPrefixColumnName

## Failing Tests by Category

### 1. FLUSH Statements (1 test)
- TestParseFlush - FLUSH statement not implemented

### 2. SHOW Statements (6 tests)
- TestParseShowColumns - SHOW COLUMNS/FIELDS/EXTENDED/FULL
- TestParseShowStatus - SHOW STATUS
- TestParseShowTables - SHOW TABLES
- TestParseShowExtendedFull - SHOW EXTENDED/FULL restrictions
- TestParseShowCreate - SHOW CREATE TABLE/TRIGGER/EVENT/FUNCTION/PROCEDURE/VIEW
- TestParseShowCollation - SHOW COLLATION

### 3. CREATE TABLE Features (18 tests)
- TestParseCreateTableAutoIncrement - AUTO_INCREMENT column option
- TestParseCreateTablePrimaryAndUniqueKey - PRIMARY KEY/UNIQUE KEY constraints
- TestParseCreateTablePrimaryAndUniqueKeyWithIndexOptions - Index options in constraints
- TestParsePrefixKeyPart - Prefix key parts (textcol(10))
- TestFunctionalKeyPart - Functional key parts with expressions
- TestParseCreateTablePrimaryAndUniqueKeyWithIndexType - USING BTREE/HASH
- TestParseCreateTablePrimaryAndUniqueKeyCharacteristic - Constraint characteristics
- TestParseCreateTableColumnKeyOptions - Column-level KEY options
- TestParseCreateTableComment - Table COMMENT option
- TestParseCreateTableAutoIncrementOffset - AUTO_INCREMENT table option
- TestParseCreateTableMultipleOptionsOrderIndependent - Table options ordering
- TestParseCreateTableSetEnum - SET and ENUM types
- TestParseCreateTableEngineDefaultCharset - ENGINE and DEFAULT CHARSET
- TestParseCreateTableCollate - COLLATE option
- TestParseCreateTableBothOptionsAndAsQuery - CREATE TABLE AS with options
- TestParseCreateTableCommentCharacterSet - Column CHARACTER SET/COMMENT
- TestParseCreateTableGencol - Generated columns (VIRTUAL/STORED)
- TestParseCreateTableOptionsCommaSeparated - Comma-separated options

### 4. ALTER TABLE - MySQL Extensions (14 tests)
- TestParseAlterTableAddColumn - ADD COLUMN with position (FIRST/AFTER)
- TestParseAlterTableAddColumns - Multiple ADD COLUMN operations
- TestParseAlterTableDropPrimaryKey - DROP PRIMARY KEY
- TestParseAlterTableDropForeignKey - DROP FOREIGN KEY
- TestParseAlterTableChangeColumn - CHANGE COLUMN syntax
- TestParseAlterTableChangeColumnWithColumnPosition - CHANGE with position
- TestParseAlterTableModifyColumn - MODIFY COLUMN syntax
- TestParseAlterTableWithAlgorithm - ALGORITHM option
- TestParseAlterTableWithLock - LOCK option
- TestParseAlterTableAutoIncrement - AUTO_INCREMENT option
- TestParseAlterTableModifyColumnWithColumnPosition - MODIFY with position
- TestParseAlterTableDropIndex - DROP INDEX
- TestDDLWithIndexUsing - INDEX USING clause
- TestCreateIndexOptions - CREATE INDEX with LOCK/ALGORITHM

### 5. UPDATE/DELETE with MySQL Features (3 tests)
- TestParseUpdateWithJoins - UPDATE with JOINs
- TestParseDeleteWithOrderBy - DELETE with ORDER BY
- TestParseDeleteWithLimit - DELETE with LIMIT

### 6. INSERT Features (5 tests)
- TestParseIgnoreInsert - INSERT IGNORE
- TestParsePriorityInsert - INSERT with HIGH/LOW_PRIORITY
- TestParseInsertAs - INSERT with AS alias
- TestParseReplaceInsert - REPLACE statement
- TestParseEmptyRowInsert - INSERT with empty row ()
- TestParseInsertWithOnDuplicateUpdate - ON DUPLICATE KEY UPDATE

### 7. String/Identifier Handling (7 tests)
- TestParseIdentifiers - MySQL identifier syntax ($a$, Unicode)
- TestParseQuoteIdentifiers - Quoted reserved words
- TestParseEscapedQuoteIdentifiersWithEscape - Escaped quotes
- TestParseCreateTableWithMinimumDisplayWidth - TINYINT(3), etc.
- TestParseCreateTableUnsigned - UNSIGNED data types
- TestParseSignedDataTypes - SIGNED data types
- TestParseDeprecatedMySQLUnsignedDataTypes - Deprecated UNSIGNED types

### 8. SELECT Features (8 tests)
- TestParseSelectWithNumericPrefixColumnName - Numeric prefix columns
- TestParseSelectWithConcatenationOfExpNumberAndNumericPrefixColumn - Complex identifiers
- TestParseSubstringInSelect - SUBSTRING/SUBSTR variants
- TestParseLimitMySqlSyntax - LIMIT offset,count syntax
- TestParseStraightJoin - STRAIGHT_JOIN
- TestParseDistinctrowToDistinct - DISTINCTROW
- TestParseSelectStraightJoin - SELECT STRAIGHT_JOIN modifier
- TestParseSelectModifiers - HIGH_PRIORITY, SQL_SMALL_RESULT, etc.
- TestParseSelectModifiersAnyOrder - Modifier ordering
- TestParseSelectModifiersCanBeRepeated - Duplicate modifiers
- TestParseSelectModifiersCanonicalOrdering - Canonical form
- TestParseSelectModifiersErrors - Invalid modifier combinations

### 9. String/Value Literals (5 tests)
- TestParseHexStringIntroducer - X'...' and 0x... hex strings
- TestParseStringIntroducers - _utf8'...', _binary'...' introducers
- TestParseBitstringLiteral - b'1010', B'1010', 0b1010
- TestParseCastIntegers - CAST as UNSIGNED/SIGNED
- TestParseValues - VALUES statement

### 10. Other Statements (12 tests)
- TestParseUse - USE statement
- TestParseSetVariables - SET variables
- TestParseKill - KILL statement
- TestParseSetNames - SET NAMES
- TestParseShowVariables - SHOW VARIABLES
- TestParseDropTemporaryTable - DROP TEMPORARY TABLE
- TestParseLockTables - LOCK/UNLOCK TABLES
- TestParseJsonTable - JSON_TABLE function
- TestGroupConcat - GROUP_CONCAT
- TestParseLogicalXor - XOR operator
- TestParseGrant - GRANT
- TestParseRevoke - REVOKE
- TestParseDropTrigger - DROP TRIGGER
- TestParseDropIndex - DROP INDEX
- TestParseMatchAgainstWithAlias - MATCH AGAINST
- TestVariableAssignmentUsingColonEqual - @var := expr
- TestMySQLForeignKeyWithIndexName - FOREIGN KEY with index name
- TestParseCreateDatabaseWithCharset - CREATE DATABASE
- TestParseCreateDatabaseWithCharsetOptionOrdering - Option ordering
- TestParseLooksLikeSingleLineComment - -- comment handling
- TestParseConvertUsing - CONVERT USING
- TestParseCreateTableWithColumnCollate - Column COLLATE
- TestParseTableColumnOptionOnUpdate - ON UPDATE CURRENT_TIMESTAMP
- TestParseFulltextExpression - MATCH AGAINST
- TestParseCreateTableWithIndexDefinition - Inline KEY definition
- TestParseCreateTableUnallowConstraintThenIndex - Constraint ordering
- TestParseCreateTableWithFulltextDefinition - FULLTEXT index
- TestParseCreateTableWithSpatialDefinition - SPATIAL index
- TestParseBeginWithoutTransaction - BEGIN
- TestParseGeometricTypesSridOption - Geometry with SRID
- TestParseDoublePrecision - DOUBLE PRECISION type
- TestParseCreateTrigger - CREATE TRIGGER
- TestParseCreateTriggerCompoundStatement - Compound statement triggers
- TestParseOptimizerHints - Optimizer hints (/*+ ... */)

### 11. CREATE VIEW (5 tests)
- TestParseCreateViewAlgorithmParam - ALGORITHM = {UNDEFINED|MERGE|TEMPTABLE}
- TestParseCreateViewDefinerParam - DEFINER = user
- TestParseCreateViewSecurityParam - SQL SECURITY {DEFINER|INVOKER}
- TestParseCreateViewMultipleParams - Multiple view options

### 12. LIKE/Regex (2 tests)
- TestParseRlikeAndRegexp - RLIKE, REGEXP operators
- TestParseLikeWithEscape - LIKE ... ESCAPE

## Most Common Parser Errors
1. `Expected: )` - 27 failures (mostly CREATE TABLE options)
2. `Expected: end of statement` - 24 failures (extra clauses not supported)
3. `Expected: ADD` - 7 failures (ALTER TABLE operations)
4. `Expected: TABLE` - 5 failures (non-table CREATE statements)
5. `Expected: (` - 5 failures (function call syntax issues)

## Required Parser Implementations

### High Priority (for 40% target)
To reach 40% pass rate, the following features would provide the most benefit:

1. **SHOW statements** - SHOW COLUMNS, TABLES, STATUS, VARIABLES, COLLATION (6 tests)
2. **FLUSH statement** - Various FLUSH types (1 test)
3. **CREATE TABLE options** - Multiple table/column options (15+ tests)
4. **String literals** - Hex strings (X'...', 0x...), introducers (_utf8'...') (3 tests)
5. **USE statement** - Simple identifier handling (1 test)
6. **Basic ALTER TABLE** - ADD/DROP COLUMN, basic operations (5 tests)
7. **Simpler INSERT** - INSERT IGNORE, basic priority (2 tests)
8. **SELECT modifiers** - LIMIT MySQL syntax, basic modifiers (3 tests)

### Medium Priority
- CREATE VIEW parameters (5 tests)
- ALTER TABLE MySQL extensions (8 tests)
- UPDATE/DELETE with JOINs (3 tests)
- SUBSTRING variants (1 test)
- GROUP_CONCAT (1 test)

### Lower Priority (more complex)
- CREATE TRIGGER (2 tests)
- CREATE DATABASE (3 tests)
- KILL statement (1 test)
- Optimizer hints (1 test)
- LOCK TABLES (1 test)
- JSON_TABLE (1 test)
- Complex string/identifier handling (5+ tests)

## Recommendations

1. **Immediate wins**: Implement SHOW, FLUSH, USE statements (8 tests)
2. **Next phase**: Fix CREATE TABLE options parsing (15+ tests)
3. **Medium effort**: String literal enhancements (5 tests)
4. **Complex**: ALTER TABLE MySQL extensions, CREATE TRIGGER, CREATE DATABASE

## Test Coverage Notes

All 130 MySQL tests from the Rust implementation have been successfully ported to Go. The tests are organized into two files:
- `mysql_test.go`: Core MySQL tests (51 tests)
- `mysql_batch2_test.go`: Extended MySQL tests (79 tests)

Each test includes a reference comment pointing to the original Rust test location in `tests/sqlparser_mysql.rs`.
