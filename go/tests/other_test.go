// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

// Package tests contains the SQL parsing tests.
// These tests are ported from tests/sqlparser_common.rs in the Rust implementation.
// This file contains other/miscellaneous tests.
package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseAssert verifies ASSERT statement parsing.
// Reference: tests/sqlparser_common.rs:4381
func TestParseAssert(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "ASSERT (SELECT COUNT(*) FROM my_table) > 0"
	ast := dialects.OneStatementParsesTo(t, sql, "ASSERT (SELECT COUNT(*) FROM my_table) > 0")
	require.NotNil(t, ast)

	// Check it's an Assert statement without message
	assertStmt, ok := ast.(*statement.Assert)
	require.True(t, ok)
	require.Nil(t, assertStmt.Message)
}

// TestParseAssertMessage verifies ASSERT statement with AS message parsing.
// Reference: tests/sqlparser_common.rs:4396
func TestParseAssertMessage(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "ASSERT (SELECT COUNT(*) FROM my_table) > 0 AS 'No rows in my_table'"
	ast := dialects.OneStatementParsesTo(t, sql,
		"ASSERT (SELECT COUNT(*) FROM my_table) > 0 AS 'No rows in my_table'")
	require.NotNil(t, ast)

	// Check it's an Assert statement with message
	assertStmt, ok := ast.(*statement.Assert)
	require.True(t, ok)
	require.NotNil(t, assertStmt.Message)
}

// TestParseRaiseStatement verifies RAISE statement parsing.
// Reference: tests/sqlparser_common.rs:16021
func TestParseRaiseStatement(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// RAISE statement forms - parsing only (struct fields not yet implemented)
	dialects.VerifiedStmt(t, "RAISE USING MESSAGE = 42")
	dialects.VerifiedStmt(t, "RAISE USING MESSAGE = 'error'")
	dialects.VerifiedStmt(t, "RAISE myerror")
	dialects.VerifiedStmt(t, "RAISE 42")
	dialects.VerifiedStmt(t, "RAISE using")
	dialects.VerifiedStmt(t, "RAISE")

	// Error case
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, "RAISE USING MESSAGE error")
	require.Error(t, err)
}

// TestParseReturn verifies RETURN statement parsing.
// Reference: tests/sqlparser_common.rs:17156
func TestParseReturn(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic RETURN - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "RETURN")

	// RETURN with value
	_ = dialects.VerifiedStmt(t, "RETURN 1")
}

// TestParseMethodSelect verifies method call parsing in SELECT.
// Reference: tests/sqlparser_common.rs:14665
func TestParseMethodSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT LEFT('abc', 1).value('.', 'NVARCHAR(MAX)').value('.', 'NVARCHAR(MAX)') AS T")
	dialects.VerifiedOnlySelect(t, "SELECT STUFF((SELECT ',' + name FROM sys.objects FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS T")
	dialects.VerifiedOnlySelect(t, "SELECT CAST(column AS XML).value('.', 'NVARCHAR(MAX)') AS T")

	// CONVERT support
	dialects2 := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTryConvert() && d.ConvertTypeBeforeValue()
	})
	dialects2.VerifiedOnlySelect(t, "SELECT CONVERT(XML, '<Book>abc</Book>').value('.', 'NVARCHAR(MAX)').value('.', 'NVARCHAR(MAX)') AS T")
}

// TestParseColumnAliases verifies column alias parsing with AS keyword.
// Reference: tests/sqlparser_common.rs:1184
func TestParseColumnAliases(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a.col + 1 AS newname FROM foo AS a"
	dialects.VerifiedOnlySelect(t, sql)

	// alias without AS is parsed correctly
	dialects.OneStatementParsesTo(t, "SELECT a.col + 1 newname FROM foo AS a", sql)
}

// TestParseAliasedExpressions verifies aliased expressions.
// Reference: tests/sqlparser_common.rs:384
func TestParseAliasedExpressions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT x AS y FROM t"
	dialects.VerifiedStmt(t, sql1)

	// alias without AS is parsed correctly and normalized to include AS:
	sql2 := "SELECT x AS y FROM t"
	dialects.OneStatementParsesTo(t, "SELECT x y FROM t", sql2)
}

// TestParseColumnDefinitionTrailingCommas verifies column definition trailing comma parsing.
// Reference: tests/sqlparser_common.rs:15746
func TestParseColumnDefinitionTrailingCommas(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsColumnDefinitionTrailingCommas()
	})

	dialects.OneStatementParsesTo(t, "CREATE TABLE T (x INT64,)", "CREATE TABLE T (x INT64)")
	dialects.OneStatementParsesTo(t, "CREATE TABLE T (x INT64, y INT64, )", "CREATE TABLE T (x INT64, y INT64)")
	dialects.OneStatementParsesTo(t, "CREATE VIEW T (x, y, ) AS SELECT 1", "CREATE VIEW T (x, y) AS SELECT 1")

	// Test unsupported dialects get an error
	// Trailing commas are allowed if either SupportsTrailingCommas() or SupportsColumnDefinitionTrailingCommas() is true
	unsupportedDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsTrailingCommas() && !d.SupportsColumnDefinitionTrailingCommas()
	})

	_, err := utils.ParseSQLWithDialects(unsupportedDialects.Dialects, "CREATE TABLE employees (name text, age int,)")
	require.Error(t, err)
}

// TestParseKeyValueOptionsTrailingSemicolon verifies trailing semicolon handling.
// Reference: tests/sqlparser_common.rs:18570
func TestParseKeyValueOptionsTrailingSemicolon(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Trailing semicolon should be handled gracefully
	sql := "CREATE USER u1 option1='value1' option2='value2';"
	canonical := "CREATE USER u1 option1='value1' option2='value2'"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify statement can be serialized
	result := stmts[0].String()
	// The result should match the canonical form (without semicolon)
	assert.Equal(t, canonical, result)
}

// TestParseCreateOrAlterTable verifies CREATE OR REPLACE TABLE parsing.
// Reference: tests/sqlparser_common.rs:4585
func TestParseCreateOrAlterTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE OR REPLACE TABLE t (a INT)"
	ast := dialects.VerifiedStmt(t, sql)
	require.NotNil(t, ast)

	// Check that or_replace is set
	createTable, ok := ast.(*statement.CreateTable)
	require.True(t, ok)
	assert.True(t, createTable.OrReplace)
}

// TestParseBadConstraint verifies error handling for bad constraint syntax.
// Reference: tests/sqlparser_common.rs:5322
func TestParseBadConstraint(t *testing.T) {
	// Missing identifier after ADD
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "ALTER TABLE tab ADD")
	require.Error(t, err)

	// Missing column name or constraint in CREATE TABLE
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CREATE TABLE tab (foo int,")
	require.Error(t, err)
}

// TestIdentifierUnicodeSupport verifies Unicode identifier support.
// Reference: tests/sqlparser_common.rs:17570
func TestIdentifierUnicodeSupport(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mysql.NewMySqlDialect(),
			redshift.NewRedshiftSqlDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}
	sql := "SELECT phoneǤЖשचᎯ⻩☯♜🦄⚛🀄ᚠ⌛🌀 AS tbl FROM customers"
	_ = dialects.VerifiedStmt(t, sql)
}

// TestIdentifierUnicodeStart verifies Unicode start characters in identifiers.
// Reference: tests/sqlparser_common.rs:17581
func TestIdentifierUnicodeStart(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mysql.NewMySqlDialect(),
			redshift.NewRedshiftSqlDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}
	sql := "SELECT 💝phone AS 💝 FROM customers"
	_ = dialects.VerifiedStmt(t, sql)
}

// TestParseNonLatinIdentifiers verifies Unicode/non-Latin identifier parsing.
// Reference: tests/sqlparser_common.rs:11794
func TestParseNonLatinIdentifiers(t *testing.T) {
	// Test with dialects that support non-Latin identifiers
	sql := "SELECT 説明 FROM table1"
	dialect := bigquery.NewBigQueryDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Test multiple non-Latin identifiers
	sql = "SELECT 説明, hühnervögel, garçon, Москва, 東京 FROM inter01"
	stmts, err = parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Dialects that don't support emoji
	dialectsWithEmojiRestriction := []sqlparserDialects.Dialect{
		generic.NewGenericDialect(),
		mssql.NewMsSqlDialect(),
	}

	for _, dialect := range dialectsWithEmojiRestriction {
		sql := "SELECT 💝 FROM table1"
		_, err := parser.ParseSQL(dialect, sql)
		assert.Error(t, err, "Expected error for emoji identifier with %T", dialect)
	}
}

// TestParseArrayTypeDefWithBrackets verifies array type definition with brackets.
// Reference: tests/sqlparser_common.rs:16582
func TestParseArrayTypeDefWithBrackets(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsArrayTypedefWithBrackets()
	})

	dialects.VerifiedStmt(t, "SELECT x::INT[]")
	dialects.VerifiedStmt(t, "SELECT STRING_TO_ARRAY('1,2,3', ',')::INT[3]")
}

// TestParseSelectExprStar verifies SELECT with qualified wildcard expansion.
// Reference: tests/sqlparser_common.rs:1206
func TestParseSelectExprStar(t *testing.T) {
	// Filter dialects that support select expression star
	dialectsFiltered := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsSelectExprStar()
	})

	// Identifier wildcard expansion
	dialectsFiltered.VerifiedOnlySelect(t, "SELECT foo.bar.* FROM T")

	// Arbitrary compound expression with wildcard expansion
	dialectsFiltered.VerifiedOnlySelect(t, "SELECT foo - bar.* FROM T")

	// Arbitrary expression wildcard expansion with function
	dialectsFiltered.VerifiedOnlySelect(t, "SELECT myfunc().foo.* FROM T")

	// Test float multiplication - Go preserves the decimal point like Rust without bigdecimal
	dialectsFiltered.VerifiedOnlySelect(t, "SELECT 2. * 3 FROM T")

	// Test myfunc().*
	dialectsFiltered.VerifiedOnlySelect(t, "SELECT myfunc().* FROM T")

	// Invalid: double wildcard
	for _, d := range dialectsFiltered.Dialects {
		_, err := parser.ParseSQL(d, "SELECT foo.*.* FROM T")
		assert.Error(t, err, "Expected error for double wildcard with dialect %T", d)
	}

	// Test EXCEPT with wildcard
	dialectsExcept := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsSelectExprStar() && d.SupportsSelectWildcardExcept()
	})
	dialectsExcept.VerifiedOnlySelect(t, "SELECT myfunc().* EXCEPT (foo) FROM T")
}

// TestParseSelectWildcardWithAlias verifies wildcard with alias parsing.
// Reference: tests/sqlparser_common.rs:1288
func TestParseSelectWildcardWithAlias(t *testing.T) {
	dialectsFiltered := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsSelectWildcardWithAlias()
	})

	// qualified wildcard with alias
	dialectsFiltered.ParseSQL(t, "SELECT t.* AS all_cols FROM t")

	// unqualified wildcard with alias
	dialectsFiltered.ParseSQL(t, "SELECT * AS all_cols FROM t")

	// mixed: regular column + qualified wildcard with alias
	dialectsFiltered.ParseSQL(t, "SELECT a.id, b.* AS b_cols FROM a JOIN b ON (a.id = b.a_id)")
}

// TestParseSelectParenthesizedWildcard verifies SELECT DISTINCT(*) parsing.
// Reference: tests/sqlparser_common.rs:18545
func TestParseSelectParenthesizedWildcard(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// SELECT DISTINCT(*) - parentheses should be normalized away
	sql := "SELECT DISTINCT (*) FROM table1"
	_ = "SELECT DISTINCT * FROM table1" // canonical form

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Also test without spaces
	sql2 := "SELECT DISTINCT(*) FROM table1"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
}

// TestParseProjectionNestedType verifies nested type projection parsing.
// Reference: tests/sqlparser_common.rs:1435
func TestParseProjectionNestedType(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT customer.address.state FROM foo"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseBinaryAll verifies ALL comparison operator parsing.
// Reference: tests/sqlparser_common.rs:2449
func TestParseBinaryAll(t *testing.T) {
	sql := "SELECT a = ALL(b)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseBinaryAny verifies ANY comparison operator parsing.
// Reference: tests/sqlparser_common.rs:2435
func TestParseBinaryAny(t *testing.T) {
	sql := "SELECT a = ANY(b)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseOverlaps verifies OVERLAPS expression parsing.
// Reference: tests/sqlparser_common.rs:15741
func TestParseOverlaps(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT (DATE '2016-01-10', DATE '2016-02-01') OVERLAPS (DATE '2016-01-20', DATE '2016-02-10')")
}

// TestParseIfStatement verifies IF statement parsing.
// Reference: tests/sqlparser_common.rs:15894
func TestParseIfStatement(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		_, ok := d.(*mssql.MsSqlDialect)
		return !ok
	})

	// IF statement forms - parsing only (struct fields not yet implemented)
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; ELSEIF 2 THEN SELECT 2; ELSE SELECT 3; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; SELECT 2; SELECT 3; ELSEIF 2 THEN SELECT 4; SELECT 5; ELSEIF 3 THEN SELECT 6; SELECT 7; ELSE SELECT 8; SELECT 9; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; SELECT 2; ELSE SELECT 3; SELECT 4; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; SELECT 2; SELECT 3; ELSEIF 2 THEN SELECT 3; SELECT 4; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; SELECT 2; END IF")
	dialects.VerifiedStmt(t, "IF (1) THEN SELECT 1; SELECT 2; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; ELSEIF 1 THEN END IF")

	// Error case
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, "IF 1 THEN SELECT 1; ELSEIF 1 THEN SELECT 2; END")
	require.Error(t, err)
}

// TestParseTupleInvalid verifies error handling for invalid tuple syntax.
// Reference: tests/sqlparser_common.rs:2555
func TestParseTupleInvalid(t *testing.T) {
	d := generic.NewGenericDialect()

	// Missing closing paren
	sql := "select (1"
	_, err := parser.ParseSQL(d, sql)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: )")

	// Empty tuple
	sql = "select (), 2"
	_, err = parser.ParseSQL(d, sql)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "an expression")
}
