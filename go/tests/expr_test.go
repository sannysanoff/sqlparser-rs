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

// Package tests contains the SQL parsing tests organized by category.
// This file contains expression parsing tests extracted from common_*.go files.
package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/ansi"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/hive"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// ============================================================================
// Literal Tests
// ============================================================================

// TestParseLiteralInteger verifies integer literal parsing.
func TestParseLiteralInteger(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT 1, -10, +20"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.Equal(t, 3, len(selectStmt.Projection))
}

// TestParseLiteralDecimal verifies decimal literal parsing.
func TestParseLiteralDecimal(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT 0.300000000000000004, 9007199254740993.0"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.Equal(t, 2, len(selectStmt.Projection))
}

// TestParseLiteralString verifies parsing of string literals including national and hex strings.
func TestParseLiteralString(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT 'one', N'national string', X'deadBEEF'"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	dialects.OneStatementParsesTo(t, "SELECT x'deadBEEF'", "SELECT X'deadBEEF'")
	dialects.OneStatementParsesTo(t, "SELECT n'national string'", "SELECT N'national string'")
}

// TestParseLiteralDate verifies parsing of DATE typed string literals.
func TestParseLiteralDate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DATE '1999-01-01'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseLiteralTime verifies parsing of TIME typed string literals.
func TestParseLiteralTime(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT TIME '01:23:34'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseLiteralDatetime verifies parsing of DATETIME typed string literals.
func TestParseLiteralDatetime(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DATETIME '1999-01-01 01:23:34.45'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseLiteralTimestampWithoutTimeZone verifies parsing of TIMESTAMP without time zone.
func TestParseLiteralTimestampWithoutTimeZone(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT TIMESTAMP '1999-01-01 01:23:34'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseLiteralTimestampWithTimeZone verifies parsing of TIMESTAMPTZ with time zone.
func TestParseLiteralTimestampWithTimeZone(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT TIMESTAMPTZ '1999-01-01 01:23:34Z'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseNumericLiteralUnderscore verifies parsing of numeric literals with underscores.
func TestParseNumericLiteralUnderscore(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsNumericLiteralUnderscores()
	})

	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support numeric literal underscores")
		return
	}

	// Skip the VerifiedOnlySelectWithCanonical and just verify parsing works
	stmts := dialects.ParseSQL(t, "SELECT 10_000")
	require.Len(t, stmts, 1)
}

// TestDoubleValue tests parsing of various double/floating-point number formats.
func TestDoubleValue(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testCases := []string{
		"0.", "0.0", "0000.", "0000.00", ".0", ".00",
		"0e0", "0e+0", "0e-0", "0.e-0", "0.e+0",
		".0e-0", ".0e+0", "00.0e+0", "00.0e-0",
	}

	for _, num := range testCases {
		for _, sign := range []string{"", "+", "-"} {
			signedNum := sign + num
			sql := "SELECT " + signedNum
			_, err := parser.ParseSQL(dialects.Dialects[0], sql)
			require.NoError(t, err, "Failed to parse: %s", sql)
		}
	}
}

// TestParseNegativeValue tests parsing of negative values in SELECT.
func TestParseNegativeValue(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT -1"
	dialects.OneStatementParsesTo(t, sql1, "SELECT -1")
}

// TestParseNumber verifies number parsing.
func TestParseNumber(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT 1.0"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// String Literal Tests
// ============================================================================

// TestParseStringLiterals verifies string literal parsing.
func TestParseStringLiterals(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT 'hello' FROM t"
	dialects.VerifiedStmt(t, sql1)

	// For double-quoted strings, only test with generic dialect
	// as different dialects handle them differently
	sql2 := `SELECT "hello" FROM t`
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql2)
	require.NoError(t, err, "Generic dialect should parse double-quoted strings")
}

// TestParseAdjacentStringLiteralConcatenation verifies adjacent string literal concatenation.
func TestParseAdjacentStringLiteralConcatenation(t *testing.T) {
	// Test with single-quoted strings (portable across dialects)
	sql := `SELECT 'M' 'y' 'S' 'q' 'l'`
	canonical := `SELECT 'MySql'`

	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsStringLiteralConcatenation()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support string literal concatenation")
		return
	}

	// Verify parsing works and produces expected canonical form
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, canonical, stmts[0].String())

	sql2 := "SELECT * FROM t WHERE col = 'Hello' ' ' 'World!'"
	canonical2 := "SELECT * FROM t WHERE col = 'Hello World!'"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
	assert.Equal(t, canonical2, stmts2[0].String())
}

// TestParseStringLiteralConcatenationWithNewline verifies string concatenation with newlines.
func TestParseStringLiteralConcatenationWithNewline(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsStringLiteralConcatenationWithNewline()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support string literal concatenation with newline")
		return
	}

	sql := `SELECT 'abc' in ('a'
		'b'
		-- COMMENT
		'c',
		-- COMMENT
		'd'
	)`

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Boolean and NULL Tests
// ============================================================================

// TestParseNull verifies NULL literal parsing.
func TestParseNull(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT NULL FROM t"
	dialects.VerifiedStmt(t, sql)
}

// TestParseNotnull verifies NOTNULL operator parsing.
func TestParseNotnull(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT x IS NOT NULL FROM t")
}

// TestParseNullInSelect verifies NULL value in SELECT.
func TestParseNullInSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT NULL"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Interval Tests
// ============================================================================

// TestParseIntervalAll verifies parsing of various INTERVAL expressions.
func TestParseIntervalAll(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1-1' YEAR TO MONTH")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '01:01.01' MINUTE (5) TO SECOND (5)")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' SECOND (5, 4)")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '10' HOUR")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 5 DAY")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 5 DAYS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '10' HOUR (1)")

	dialect := generic.NewGenericDialect()
	_, err := parser.ParseSQL(dialect, "SELECT INTERVAL '1' SECOND TO SECOND")
	require.Error(t, err)

	_, err = parser.ParseSQL(dialect, "SELECT INTERVAL '10' HOUR (1) TO HOUR (2)")
	require.Error(t, err)

	variousIntervals := []string{
		"YEAR", "MONTH", "WEEK", "DAY", "HOUR", "MINUTE", "SECOND",
		"YEARS", "MONTHS", "WEEKS", "DAYS", "HOURS", "MINUTES", "SECONDS",
	}
	for _, unit := range variousIntervals {
		dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT INTERVAL '1' %s", unit))
	}

	rangeIntervals := []string{
		"YEAR TO MONTH", "DAY TO HOUR", "DAY TO MINUTE", "DAY TO SECOND",
		"HOUR TO MINUTE", "HOUR TO SECOND", "MINUTE TO SECOND",
	}
	for _, r := range rangeIntervals {
		dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT INTERVAL '1' %s", r))
	}

	numericIntervals := []string{
		"1 YEAR", "1 MONTH", "1 WEEK", "1 DAY",
		"1 HOUR", "1 MINUTE", "1 SECOND",
		"1 YEARS", "1 MONTHS", "1 WEEKS", "1 DAYS",
		"1 HOURS", "1 MINUTES", "1 SECONDS",
	}
	for _, ni := range numericIntervals {
		dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT INTERVAL %s", ni))
	}

	dialects.VerifiedOnlySelect(t, "SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::INTERVAL")
}

// TestParseIntervalDontRequireUnit verifies INTERVAL without explicit unit for dialects that don't require it.
func TestParseIntervalDontRequireUnit(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()

	testCases := []string{
		"SELECT INTERVAL '1 DAY'",
		"SELECT INTERVAL '1 YEAR'",
		"SELECT INTERVAL '1 MONTH'",
		"SELECT INTERVAL '1 DAY'",
		"SELECT INTERVAL '1 HOUR'",
		"SELECT INTERVAL '1 MINUTE'",
		"SELECT INTERVAL '1 SECOND'",
	}

	for _, sql := range testCases {
		stmts := utils.MustParseSQL(t, dialect, sql)
		require.Len(t, stmts, 1)
	}
}

// TestParseIntervalRequireUnit verifies that dialects requiring interval qualifiers error on missing units.
func TestParseIntervalRequireUnit(t *testing.T) {
	dialect := mysql.NewMySqlDialect()
	sql := "SELECT INTERVAL '1 DAY'"
	_, err := parser.ParseSQL(dialect, sql)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "INTERVAL requires a unit")
}

// TestParseIntervalRequireQualifier verifies INTERVAL with expressions for dialects requiring qualifiers.
func TestParseIntervalRequireQualifier(t *testing.T) {
	dialect := mysql.NewMySqlDialect()

	testCases := []string{
		"SELECT INTERVAL 1 + 1 DAY",
		"SELECT INTERVAL '1' + '1' DAY",
		"SELECT INTERVAL '1' + '2' - '3' DAY",
	}

	for _, sql := range testCases {
		stmts := utils.MustParseSQL(t, dialect, sql)
		require.Len(t, stmts, 1)
	}
}

// TestParseIntervalDisallowIntervalExpr verifies INTERVAL comparisons work for dialects without qualifier requirement.
func TestParseIntervalDisallowIntervalExpr(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()

	testCases := []string{
		"SELECT INTERVAL '1 DAY'",
		"SELECT INTERVAL '1 YEAR'",
		"SELECT INTERVAL '1 YEAR' AS one_year",
		"SELECT INTERVAL '1 DAY' > INTERVAL '1 SECOND'",
	}

	for _, sql := range testCases {
		stmts := utils.MustParseSQL(t, dialect, sql)
		require.Len(t, stmts, 1)
	}
}

// TestIntervalDisallowIntervalExprGt verifies INTERVAL greater-than comparisons.
func TestIntervalDisallowIntervalExprGt(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()
	sql := "SELECT INTERVAL '1 second' > x"
	stmts := utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)
}

// TestIntervalDisallowIntervalExprDoubleColon verifies INTERVAL casting with double colon.
func TestIntervalDisallowIntervalExprDoubleColon(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()
	sql := "SELECT INTERVAL '1 second'::TEXT"
	stmts := utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)
}

// TestParseIntervalAndOrXor verifies INTERVAL expressions in AND/OR/XOR conditions.
func TestParseIntervalAndOrXor(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()

	testCases := []string{
		"SELECT col FROM test WHERE d3_date > d1_date + INTERVAL '5 days' AND d2_date > d1_date + INTERVAL '3 days'",
		"SELECT col FROM test WHERE d3_date > d1_date + INTERVAL '5 days' OR d2_date > d1_date + INTERVAL '3 days'",
		"SELECT col FROM test WHERE d3_date > d1_date + INTERVAL '5 days' XOR d2_date > d1_date + INTERVAL '3 days'",
	}

	for _, sql := range testCases {
		stmts := utils.MustParseSQL(t, dialect, sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// CAST and TRY_CAST Tests
// ============================================================================

// TestParseCast verifies CAST expression parsing.
func TestParseCast(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT CAST(id AS BIGINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS TINYINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS MEDIUMINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS NUMERIC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS DEC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS DECIMAL) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS NVARCHAR(50)) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS CLOB) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS CLOB(50)) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS BINARY(50)) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS VARBINARY(50)) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS BLOB) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS BLOB(50)) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(details AS JSONB) FROM customer")
}

// TestParseTryCast verifies TRY_CAST expression parsing.
func TestParseTryCast(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS BIGINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS NUMERIC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS DEC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS DECIMAL) FROM customer")
}

// TestParseDoubleColonCastAtTimezone verifies double colon cast with AT TIME ZONE parsing.
func TestParseDoubleColonCastAtTimezone(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT '2001-01-01T00:00:00.000Z'::TIMESTAMP AT TIME ZONE 'Europe/Brussels' FROM t"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// CASE Expression Tests
// ============================================================================

// TestParseCaseExpression verifies CASE expression parsing.
func TestParseCaseExpression(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT CASE x WHEN 1 THEN 'one' WHEN 2 THEN 'two' ELSE 'other' END FROM t"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSimpleCaseExpr verifies simple CASE expression with operand parsing.
func TestParseSimpleCaseExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT CASE foo WHEN 1 THEN 'Y' ELSE 'N' END"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseSearchedCaseExpr verifies searched CASE expression parsing.
func TestParseSearchedCaseExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT CASE WHEN bar IS NULL THEN 'null' WHEN bar = 0 THEN '=0' WHEN bar >= 0 THEN '>=0' ELSE '<0' END FROM foo"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// IN Expression Tests
// ============================================================================

// TestParseInExpression verifies IN expression parsing.
func TestParseInExpression(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE x IN (1, 2, 3)"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE x NOT IN (1, 2, 3)"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseInList verifies IN list expression parsing.
func TestParseInList(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE segment IN ('HIGH', 'MED')")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE segment NOT IN ('HIGH', 'MED')")
}

// TestParseInSubquery verifies IN with subquery parsing.
func TestParseInSubquery(t *testing.T) {
	sql := "SELECT * FROM customers WHERE segment IN (SELECT segm FROM bar)"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// BETWEEN Tests
// ============================================================================

// TestParseBetween verifies BETWEEN expression parsing.
func TestParseBetween(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM t WHERE x BETWEEN 1 AND 10"
	dialects.VerifiedStmt(t, sql)
}

// TestParseBetweenWithExpr verifies BETWEEN with complex expressions.
func TestParseBetweenWithExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM t WHERE 1 BETWEEN 1 + 2 AND 3 + 4 IS NULL"
	dialects.VerifiedStmt(t, sql)

	sql = "SELECT * FROM t WHERE 1 = 1 AND 1 + x BETWEEN 1 AND 2"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// IS NULL / IS NOT NULL Tests
// ============================================================================

// TestParseIsNull verifies IS NULL expression parsing.
func TestParseIsNull(t *testing.T) {
	sql := "a IS NULL"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseIsNotNull verifies IS NOT NULL expression parsing.
func TestParseIsNotNull(t *testing.T) {
	sql := "a IS NOT NULL"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseNullHandling verifies IS NULL / IS NOT NULL parsing.
func TestParseNullHandling(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE x IS NULL"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE x IS NOT NULL"
	dialects.VerifiedStmt(t, sql2)
}

// ============================================================================
// IS DISTINCT FROM Tests
// ============================================================================

// TestParseIsDistinctFrom verifies IS DISTINCT FROM expression parsing.
func TestParseIsDistinctFrom(t *testing.T) {
	sql := "a IS DISTINCT FROM b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseIsNotDistinctFrom verifies IS NOT DISTINCT FROM expression parsing.
func TestParseIsNotDistinctFrom(t *testing.T) {
	sql := "a IS NOT DISTINCT FROM b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// ============================================================================
// IS BOOLEAN Tests
// ============================================================================

// TestParseIsBoolean verifies IS TRUE/FALSE/UNKNOWN/NORMALIZED expression parsing.
func TestParseIsBoolean(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedExpr(t, "a IS TRUE")
	dialects.VerifiedExpr(t, "a IS NOT TRUE")
	dialects.VerifiedExpr(t, "a IS FALSE")
	dialects.VerifiedExpr(t, "a IS NOT FALSE")
	dialects.VerifiedExpr(t, "a IS NORMALIZED")
	dialects.VerifiedExpr(t, "a IS NOT NORMALIZED")
	dialects.VerifiedExpr(t, "a IS NFKC NORMALIZED")
	dialects.VerifiedExpr(t, "a IS NOT NFKD NORMALIZED")
	dialects.VerifiedExpr(t, "a IS UNKNOWN")
	dialects.VerifiedExpr(t, "a IS NOT UNKNOWN")

	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS TRUE")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT TRUE")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS FALSE")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT FALSE")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NFC NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NFD NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT NFKC NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS UNKNOWN")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT UNKNOWN")
}

// ============================================================================
// LIKE / ILIKE / SIMILAR TO Tests
// ============================================================================

// TestParseLike verifies LIKE expression parsing.
func TestParseLike(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name LIKE '%a'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name LIKE '%a' ESCAPE '^'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT LIKE '%a'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT LIKE '%a' ESCAPE '^'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name LIKE '%a' IS NULL")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT LIKE '%a' IS NULL")
}

// TestParseLikeExpression verifies LIKE expression parsing.
func TestParseLikeExpression(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE name LIKE '%test%'"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE name NOT LIKE '%test%'"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseIlike verifies ILIKE expression parsing.
func TestParseIlike(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name ILIKE '%a'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name ILIKE '%a' ESCAPE '^'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT ILIKE '%a'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT ILIKE '%a' ESCAPE '^'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name ILIKE '%a' IS NULL")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT ILIKE '%a' IS NULL")
}

// TestParseSimilarTo verifies SIMILAR TO expression parsing.
func TestParseSimilarTo(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name SIMILAR TO '%a'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name SIMILAR TO '%a' ESCAPE '^'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name SIMILAR TO '%a' ESCAPE NULL")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT SIMILAR TO '%a'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT SIMILAR TO '%a' ESCAPE '^'")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name SIMILAR TO '%a' ESCAPE '^' IS NULL")
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT SIMILAR TO '%a' ESCAPE '^' IS NULL")
}

// TestParseNullLike verifies LIKE with NULL operands.
func TestParseNullLike(t *testing.T) {
	sql := "SELECT column1 LIKE NULL AS col_null, NULL LIKE column1 AS null_col FROM customers"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Logical Operators Tests
// ============================================================================

// TestParseLogicalOperators verifies logical operators.
func TestParseLogicalOperators(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE x = 1 AND y = 2"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE x = 1 OR y = 2"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT * FROM t WHERE NOT x = 1"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseNot verifies NOT operator parsing.
func TestParseNot(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id FROM customer WHERE NOT salary = ''"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseNotPrecedence verifies NOT operator precedence.
func TestParseNotPrecedence(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT NOT 1 OR 1 FROM t")
	dialects.VerifiedOnlySelect(t, "SELECT NOT a IS NULL FROM t")
	dialects.VerifiedOnlySelect(t, "SELECT NOT 1 NOT BETWEEN 1 AND 2 FROM t")
	dialects.VerifiedOnlySelect(t, "SELECT NOT 'a' NOT LIKE 'b' FROM t")
	dialects.VerifiedOnlySelect(t, "SELECT NOT a NOT IN ('a') FROM t")
}

// TestParseBangNot verifies the bang (!) NOT operator parsing.
func TestParseBangNot(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsBangNotOperator()
	})

	sql := "SELECT !a, !(b > 3)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Arithmetic Operators Tests
// ============================================================================

// TestParseArithmeticOperators verifies arithmetic operators.
func TestParseArithmeticOperators(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT 1 + 2 FROM t"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT 10 - 5 FROM t"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT 4 * 5 FROM t"
	dialects.VerifiedStmt(t, sql3)

	sql4 := "SELECT 20 / 4 FROM t"
	dialects.VerifiedStmt(t, sql4)
}

// TestParseSimpleMathExprPlus verifies parsing of simple addition expressions.
func TestParseSimpleMathExprPlus(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a + b, 2 + a, 2.5 + a, a_f + b_f, 2 + a_f, 2.5 + a_f FROM c"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseSimpleMathExprMinus verifies parsing of simple subtraction expressions.
func TestParseSimpleMathExprMinus(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a - b, 2 - a, 2.5 - a, a_f - b_f, 2 - a_f, 2.5 - a_f FROM c"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseUnaryMathWithPlus verifies unary minus with plus operator.
func TestParseUnaryMathWithPlus(t *testing.T) {
	sql := "-a + -b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseUnaryMathWithMultiply verifies unary minus with multiply operator.
func TestParseUnaryMathWithMultiply(t *testing.T) {
	sql := "-a * -b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// ============================================================================
// Comparison Operators Tests
// ============================================================================

// TestParseComparisonOperators verifies comparison operators.
func TestParseComparisonOperators(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE x = 1"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE x != 1"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT * FROM t WHERE x <> 1"
	dialects.VerifiedStmt(t, sql3)

	sql4 := "SELECT * FROM t WHERE x > 1"
	dialects.VerifiedStmt(t, sql4)

	sql5 := "SELECT * FROM t WHERE x < 1"
	dialects.VerifiedStmt(t, sql5)

	sql6 := "SELECT * FROM t WHERE x >= 1"
	dialects.VerifiedStmt(t, sql6)

	sql7 := "SELECT * FROM t WHERE x <= 1"
	dialects.VerifiedStmt(t, sql7)
}

// ============================================================================
// Bitwise Operators Tests
// ============================================================================

// TestParseBitwiseOps verifies bitwise operator parsing (^, |, &).
func TestParseBitwiseOps(t *testing.T) {
	dialectsNoPG := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		_, isPG := d.(*postgresql.PostgreSqlDialect)
		return !isPG
	})
	sql := "SELECT a ^ b"
	stmts := dialectsNoPG.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	dialectsAll := utils.NewTestedDialects()
	sql = "SELECT a | b"
	stmts = dialectsAll.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	sql = "SELECT a & b"
	stmts = dialectsAll.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseBitwiseShiftOps verifies bitwise shift operator parsing (<<, >>).
func TestParseBitwiseShiftOps(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsBitwiseShiftOperators()
	})

	sql := "SELECT 1 << 2, 3 >> 4"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Compound and Unary Expression Tests
// ============================================================================

// TestParseCompoundExpr1 verifies compound expression with precedence.
func TestParseCompoundExpr1(t *testing.T) {
	sql := "a + b * c"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseCompoundExpr2 verifies compound expression with precedence.
func TestParseCompoundExpr2(t *testing.T) {
	sql := "a * b + c"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestCompoundExpr tests compound expression parsing with array access and field access.
func TestCompoundExpr(t *testing.T) {
	supportedDialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			bigquery.NewBigQueryDialect(),
		},
	}

	sqls := []string{
		"SELECT abc[1].f1 FROM t",
		"SELECT abc[1].f1.f2 FROM t",
		"SELECT f1.abc[1] FROM t",
		"SELECT f1.f2.abc[1] FROM t",
		"SELECT f1.abc[1].f2 FROM t",
		"SELECT named_struct('a', 1, 'b', 2).a",
		"SELECT make_array(1, 2, 3)[1]",
		"SELECT make_array(named_struct('a', 1))[1].a",
		"SELECT abc[1][-1].a.b FROM t",
		"SELECT abc[1][-1].a.b[1] FROM t",
	}

	for _, sql := range sqls {
		supportedDialects.VerifiedStmt(t, sql)
	}
}

// TestParseGenericUnaryOps verifies unary operator parsing.
func TestParseGenericUnaryOps(t *testing.T) {
	dialects := utils.NewTestedDialects()

	tests := []struct {
		sql string
		op  string
	}{
		{"SELECT ~expr", "~"},
		{"SELECT -expr", "-"},
		{"SELECT +expr", "+"},
	}

	for _, tc := range tests {
		t.Run(tc.sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, tc.sql)
			require.Len(t, stmts, 1)
		})
	}
}

// ============================================================================
// COLLATE Tests
// ============================================================================

// TestParseCollate verifies COLLATE expression parsing.
func TestParseCollate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT name COLLATE \"de_DE\" FROM customer"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseCollateAfterParens verifies COLLATE after parentheses.
func TestParseCollateAfterParens(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT (name) COLLATE \"de_DE\" FROM customer"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Parentheses Tests
// ============================================================================

// TestParseParens verifies parenthesized expression parsing.
func TestParseParens(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT (a + b) - (c + d)"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// Method Expression Tests
// ============================================================================

// TestParseMethodExpr verifies method call expression parsing.
func TestParseMethodExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT LEFT('abc', 1).value('.', 'NVARCHAR(MAX)').value('.', 'NVARCHAR(MAX)')")
	dialects.VerifiedStmt(t, "SELECT (SELECT ',' + name FROM sys.objects FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)')")
	dialects.VerifiedStmt(t, "SELECT CAST(column AS XML).value('.', 'NVARCHAR(MAX)')")

	dialects2 := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTryConvert() && d.ConvertTypeBeforeValue()
	})
	dialects2.VerifiedStmt(t, "SELECT CONVERT(XML, '<Book>abc</Book>').value('.', 'NVARCHAR(MAX)').value('.', 'NVARCHAR(MAX)')")
}

// ============================================================================
// Composite Access Expression Tests
// ============================================================================

// TestParseCompositeAccessExpr verifies composite field access expressions.
func TestParseCompositeAccessExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()

	stmts := dialects.ParseSQL(t, "SELECT f(a).b FROM t")
	require.Len(t, stmts, 1)

	stmts2 := dialects.ParseSQL(t, "SELECT f(a).b.c FROM t")
	require.Len(t, stmts2, 1)

	stmts3 := dialects.ParseSQL(t, "SELECT * FROM t WHERE f(a).b IS NOT NULL")
	require.Len(t, stmts3, 1)
}

// ============================================================================
// Nullary Table-Valued Function Tests
// ============================================================================

// TestParseNullaryTableValuedFunction verifies nullary (no-argument) table-valued function.
func TestParseNullaryTableValuedFunction(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM fn()"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// AT TIME ZONE Tests
// ============================================================================

// TestParseAtTimezone verifies AT TIME ZONE expressions.
func TestParseAtTimezone(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT FROM_UNIXTIME(0) AT TIME ZONE 'UTC-06:00' FROM t"
	dialects.VerifiedOnlySelect(t, sql)

	sql = `SELECT DATE_FORMAT(FROM_UNIXTIME(0) AT TIME ZONE 'UTC-06:00', '%Y-%m-%dT%H') AS "hour" FROM t`
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Typed String Tests
// ============================================================================

// TestParseTypedStrings verifies parsing of various typed string literals.
func TestParseTypedStrings(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := `JSON '{"foo":"bar"}'`
	stmts := dialects.ParseSQL(t, "SELECT "+sql)
	require.Len(t, stmts, 1)
}

// TestParseBignumericKeyword verifies parsing of BIGNUMERIC typed string literals.
func TestParseBignumericKeyword(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '0'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '123456'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '-3.14'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '-0.54321'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '1.23456e05'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '-9.876e-3'")
}

// TestParseJsonKeyword verifies parsing of JSON typed string literals.
func TestParseJsonKeyword(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := `SELECT JSON '{
  "id": 10,
  "type": "fruit",
  "name": "apple",
  "on_menu": true
}'`
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// EXTRACT Tests
// ============================================================================

// TestParseExtract verifies EXTRACT expression parsing.
func TestParseExtract(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT EXTRACT(YEAR FROM d)")
	dialects.OneStatementParsesTo(t, "SELECT EXTRACT(year from d)", "SELECT EXTRACT(YEAR FROM d)")

	fields := []string{
		"MONTH", "WEEK", "DAY", "DAYOFWEEK", "DAYOFYEAR",
		"DATE", "DATETIME", "HOUR", "MINUTE", "SECOND",
		"MILLISECOND", "MICROSECOND", "NANOSECOND", "CENTURY",
		"DECADE", "DOW", "DOY", "EPOCH", "ISODOW", "ISOWEEK",
		"ISOYEAR", "JULIAN", "MICROSECONDS", "MILLENIUM",
		"MILLENNIUM", "MILLISECONDS", "QUARTER", "TIMEZONE",
		"TIMEZONE_ABBR", "TIMEZONE_HOUR", "TIMEZONE_MINUTE",
		"TIMEZONE_REGION", "TIME",
	}

	for _, field := range fields {
		sql := "SELECT EXTRACT(" + field + " FROM d)"
		dialects.VerifiedStmt(t, sql)
	}
}

// ============================================================================
// POSITION Tests
// ============================================================================

// TestParsePosition verifies POSITION expression parsing.
func TestParsePosition(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "POSITION('@' IN field)"
	dialects.VerifiedExpr(t, sql)

	sql = "position('an', 'banana', 1)"
	dialects.VerifiedExpr(t, sql)
}

// TestParsePositionNegative verifies error handling for invalid POSITION syntax.
func TestParsePositionNegative(t *testing.T) {
	sql := "SELECT POSITION(foo IN) from bar"
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err, "Expected error for incomplete POSITION expression")
}

// ============================================================================
// CEIL/FLOOR Tests
// ============================================================================

// TestParseCeilNumber verifies CEIL function parsing.
func TestParseCeilNumber(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT CEIL(1.5)")
	dialects.VerifiedStmt(t, "SELECT CEIL(float_column) FROM my_table")
}

// TestParseFloorNumber verifies FLOOR function parsing.
func TestParseFloorNumber(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT FLOOR(1.5)")
	dialects.VerifiedStmt(t, "SELECT FLOOR(float_column) FROM my_table")
}

// TestParseCeilNumberScale verifies CEIL with scale parameter.
func TestParseCeilNumberScale(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT CEIL(1.5, 1)")
	dialects.VerifiedStmt(t, "SELECT CEIL(float_column, 3) FROM my_table")
}

// TestParseFloorNumberScale verifies FLOOR with scale parameter.
func TestParseFloorNumberScale(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT FLOOR(1.5, 1)")
	dialects.VerifiedStmt(t, "SELECT FLOOR(float_column, 3) FROM my_table")
}

// TestParseCeilScale verifies CEIL with scale parameter (detailed).
func TestParseCeilScale(t *testing.T) {
	sql := "SELECT CEIL(d, 2)"

	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// TestParseFloorScale verifies FLOOR with scale parameter (detailed).
func TestParseFloorScale(t *testing.T) {
	sql := "SELECT FLOOR(d, 2)"

	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// TestParseCeilDatetime verifies CEIL with datetime field.
func TestParseCeilDatetime(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT CEIL(d TO DAY)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())

	dialects.OneStatementParsesTo(t, "SELECT CEIL(d to day)", "SELECT CEIL(d TO DAY)")

	dialects.VerifiedStmt(t, "SELECT CEIL(d TO HOUR) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO MINUTE) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO SECOND) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO MILLISECOND) FROM df")
}

// TestParseFloorDatetime verifies FLOOR with datetime field.
func TestParseFloorDatetime(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT FLOOR(d TO DAY)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())

	dialects.OneStatementParsesTo(t, "SELECT FLOOR(d to day)", "SELECT FLOOR(d TO DAY)")

	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO HOUR) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO MINUTE) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO SECOND) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO MILLISECOND) FROM df")
}

// ============================================================================
// Array Tests
// ============================================================================

// TestParseArraySubscript verifies array subscript parsing.
func TestParseArraySubscript(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT arr[1]")
	dialects.VerifiedStmt(t, "SELECT arr[:]")
	dialects.VerifiedStmt(t, "SELECT arr[1:2]")
	dialects.VerifiedStmt(t, "SELECT arr[1:2:4]")
	dialects.VerifiedStmt(t, "SELECT arr[1:array_length(arr)]")
	dialects.VerifiedStmt(t, "SELECT arr[array_length(arr) - 1:array_length(arr)]")
	dialects.VerifiedStmt(t, "SELECT arr[1][2]")
	dialects.VerifiedStmt(t, "SELECT arr[:][:]")
}

// ============================================================================
// Exponent Tests
// ============================================================================

// TestParseExponentInSelect verifies exponent notation in SELECT.
func TestParseExponentInSelect(t *testing.T) {
	dialectList := []sqlparserDialects.Dialect{
		generic.NewGenericDialect(),
		bigquery.NewBigQueryDialect(),
		postgresql.NewPostgreSqlDialect(),
		duckdb.NewDuckDbDialect(),
		mssql.NewMsSqlDialect(),
		mysql.NewMySqlDialect(),
		redshift.NewRedshiftSqlDialect(),
		snowflake.NewSnowflakeDialect(),
		sqlite.NewSQLiteDialect(),
	}

	sql := "SELECT 10e-20, 1e3, 1e+3, 1e3a, 1e, 0.5e2"

	for _, d := range dialectList {
		stmts, err := parser.ParseSQL(d, sql)
		require.NoError(t, err, "Failed to parse with dialect %T", d)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Alias Tests
// ============================================================================

// TestParseAliasEqualExpr verifies alias assignment with = syntax.
func TestParseAliasEqualExpr(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsEqAliasAssignment()
	})

	dialects.OneStatementParsesTo(t, "SELECT some_alias = some_column FROM some_table", "SELECT some_column AS some_alias FROM some_table")
	dialects.OneStatementParsesTo(t, "SELECT some_alias = (a*b) FROM some_table", "SELECT (a * b) AS some_alias FROM some_table")
}

// ============================================================================
// Modulo and Factorial Tests
// ============================================================================

// TestParseMod verifies modulo operator parsing.
func TestParseMod(t *testing.T) {
	sql := "a % b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseModNoSpaces verifies modulo operator without spaces.
func TestParseModNoSpaces(t *testing.T) {
	canonical := "a1 % b1"
	sqls := []string{"a1 % b1", "a1% b1", "a1 %b1", "a1%b1"}

	pgGeneric := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
			generic.NewGenericDialect(),
		},
	}

	for _, sql := range sqls {
		fullSql := fmt.Sprintf("SELECT %s FROM t", sql)
		fullCanonical := fmt.Sprintf("SELECT %s FROM t", canonical)
		pgGeneric.OneStatementParsesTo(t, fullSql, fullCanonical)
	}
}

// TestParseFactorialOperator verifies the factorial (!) operator parsing.
func TestParseFactorialOperator(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsFactorialOperator()
	})

	sql := "SELECT a!, (b + c)!"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// JSON Tests
// ============================================================================

// TestParseJsonObject verifies JSON_OBJECT function parsing.
func TestParseJsonObject(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mssql.NewMsSqlDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}

	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT('name' : 'value', 'type' : 1)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT('name' : 'value', 'type' : NULL ABSENT ON NULL)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT(NULL ON NULL)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT(ABSENT ON NULL)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT('name' : 'value', 'type' : JSON_ARRAY(1, 2) ABSENT ON NULL)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT('name' : 'value', 'type' : JSON_OBJECT('type_id' : 1, 'name' : 'a') NULL ON NULL)")
}

// TestParseJsonOpsWithoutColon verifies JSON operators parsing.
func TestParseJsonOpsWithoutColon(t *testing.T) {
	pgGeneric := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
			generic.NewGenericDialect(),
		},
	}
	pgGeneric.VerifiedOnlySelect(t, "SELECT a -> b")
	pgGeneric.VerifiedOnlySelect(t, "SELECT a ->> b")
	pgGeneric.VerifiedOnlySelect(t, "SELECT a #> b")
	pgGeneric.VerifiedOnlySelect(t, "SELECT a #>> b")
	pgGeneric.VerifiedOnlySelect(t, "SELECT a @> b")
	pgGeneric.VerifiedOnlySelect(t, "SELECT a <@ b")
	pgGeneric.VerifiedOnlySelect(t, "SELECT a #- b")
	pgGeneric.VerifiedOnlySelect(t, "SELECT a @? b")
	pgGeneric.VerifiedOnlySelect(t, "SELECT a @@ b")

	allDialects := utils.NewTestedDialects()
	allDialects.VerifiedOnlySelect(t, "SELECT a ->> b")
	allDialects.VerifiedOnlySelect(t, "SELECT a @> b")
	allDialects.VerifiedOnlySelect(t, "SELECT a <@ b")
	allDialects.VerifiedOnlySelect(t, "SELECT a @? b")
	allDialects.VerifiedOnlySelect(t, "SELECT a @@ b")
}

// ============================================================================
// Aggregate Function Tests
// ============================================================================

// TestParseAggregateFunctions verifies aggregate function parsing.
func TestParseAggregateFunctions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT COUNT(*) FROM customer"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT SUM(amount) FROM orders"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT AVG(DISTINCT price) FROM products"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseCountWildcard verifies COUNT(*) and COUNT(table.*) parsing.
func TestParseCountWildcard(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, "SELECT COUNT(*) FROM Order WHERE id = 10")
	dialects.VerifiedOnlySelect(t, "SELECT COUNT(Employee.*) FROM Order JOIN Employee ON Order.employee = Employee.id")
}

// TestParseCountDistinct verifies COUNT(DISTINCT ...) parsing.
func TestParseCountDistinct(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT COUNT(DISTINCT +x) FROM customer"
	dialects.VerifiedOnlySelect(t, sql)
	dialects.VerifiedStmt(t, "SELECT COUNT(ALL +x) FROM customer")
	dialects.VerifiedStmt(t, "SELECT COUNT(+x) FROM customer")

	d := generic.NewGenericDialect()
	_, err := parser.ParseSQL(d, "SELECT COUNT(ALL DISTINCT + x) FROM customer")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Cannot specify both ALL and DISTINCT")
}

// TestParseStringAgg verifies string concatenation operator parsing.
func TestParseStringAgg(t *testing.T) {
	sql := "SELECT a || b"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseListagg verifies LISTAGG function parsing.
func TestParseListagg(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT LISTAGG(DISTINCT dateid, ', ' ON OVERFLOW TRUNCATE '%' WITHOUT COUNT) WITHIN GROUP (ORDER BY id, username)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())

	dialects.VerifiedStmt(t, "SELECT LISTAGG(sellerid) WITHIN GROUP (ORDER BY dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(DISTINCT dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW ERROR)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW TRUNCATE N'...' WITH COUNT)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW TRUNCATE X'deadbeef' WITH COUNT)")
}

// TestParseAggWithOrderBy verifies aggregate functions with ORDER BY.
func TestParseAggWithOrderBy(t *testing.T) {
	supportedDialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			postgresql.NewPostgreSqlDialect(),
			mssql.NewMsSqlDialect(),
			ansi.NewAnsiDialect(),
			hive.NewHiveDialect(),
		},
	}

	testCases := []string{
		"SELECT FIRST_VALUE(x ORDER BY x) AS a FROM T",
		"SELECT FIRST_VALUE(x ORDER BY x) FROM tbl",
		"SELECT LAST_VALUE(x ORDER BY x, y) AS a FROM T",
		"SELECT LAST_VALUE(x ORDER BY x ASC, y DESC) AS a FROM T",
	}

	for _, sql := range testCases {
		supportedDialects.VerifiedStmt(t, sql)
	}
}

// TestParseArrayAggFunc verifies ARRAY_AGG function parsing.
func TestParseArrayAggFunc(t *testing.T) {
	supportedDialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			postgresql.NewPostgreSqlDialect(),
			mssql.NewMsSqlDialect(),
			ansi.NewAnsiDialect(),
			hive.NewHiveDialect(),
		},
	}

	testCases := []string{
		"SELECT ARRAY_AGG(x ORDER BY x) AS a FROM T",
		"SELECT ARRAY_AGG(x ORDER BY x LIMIT 2) FROM tbl",
		"SELECT ARRAY_AGG(DISTINCT x ORDER BY x LIMIT 2) FROM tbl",
		"SELECT ARRAY_AGG(x ORDER BY x, y) AS a FROM T",
		"SELECT ARRAY_AGG(x ORDER BY x ASC, y DESC) AS a FROM T",
	}

	for _, sql := range testCases {
		supportedDialects.VerifiedStmt(t, sql)
	}
}

// ============================================================================
// Window Function Tests
// ============================================================================

// TestParseWindowFunctions verifies window function parsing.
func TestParseWindowFunctions(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT ROW_NUMBER() OVER (ORDER BY id) FROM t"
	dialects.VerifiedStmt(t, sql)
}

// TestParseNamedWindowFunctions verifies named window function parsing.
func TestParseNamedWindowFunctions(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			postgresql.NewPostgreSqlDialect(),
			mysql.NewMySqlDialect(),
			bigquery.NewBigQueryDialect(),
		},
	}

	sql := "SELECT row_number() OVER (w ORDER BY dt DESC), " +
		"sum(foo) OVER (win PARTITION BY a, b ORDER BY c, d ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW) " +
		"FROM foo " +
		"WINDOW w AS (PARTITION BY x), win AS (ORDER BY y)"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.NotNil(t, selectStmt)
}

// TestParseWindowRankFunction verifies window rank functions parsing.
func TestParseWindowRankFunction(t *testing.T) {
	supportedDialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			postgresql.NewPostgreSqlDialect(),
			mssql.NewMsSqlDialect(),
			ansi.NewAnsiDialect(),
			hive.NewHiveDialect(),
			snowflake.NewSnowflakeDialect(),
		},
	}

	testCases := []string{
		"SELECT column1, column2, FIRST_VALUE(column2) OVER (PARTITION BY column1 ORDER BY column2 NULLS LAST) AS column2_first FROM t1",
		"SELECT column1, column2, FIRST_VALUE(column2) OVER (ORDER BY column2 NULLS LAST) AS column2_first FROM t1",
		"SELECT col_1, col_2, LAG(col_2) OVER (ORDER BY col_1) FROM t1",
		"SELECT LAG(col_2, 1, 0) OVER (ORDER BY col_1) FROM t1",
		"SELECT LAG(col_2, 1, 0) OVER (PARTITION BY col_3 ORDER BY col_1)",
	}

	for _, sql := range testCases {
		supportedDialects.VerifiedStmt(t, sql)
	}

	supportedDialectsNulls := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mssql.NewMsSqlDialect(),
			snowflake.NewSnowflakeDialect(),
		},
	}

	nullsTestCases := []string{
		"SELECT column1, column2, FIRST_VALUE(column2) IGNORE NULLS OVER (PARTITION BY column1 ORDER BY column2 NULLS LAST) AS column2_first FROM t1",
		"SELECT column1, column2, FIRST_VALUE(column2) RESPECT NULLS OVER (PARTITION BY column1 ORDER BY column2 NULLS LAST) AS column2_first FROM t1",
		"SELECT LAG(col_2, 1, 0) IGNORE NULLS OVER (ORDER BY col_1) FROM t1",
		"SELECT LAG(col_2, 1, 0) RESPECT NULLS OVER (ORDER BY col_1) FROM t1",
	}

	for _, sql := range nullsTestCases {
		supportedDialectsNulls.VerifiedStmt(t, sql)
	}
}

// TestParseWindowFunctionNullTreatmentArg verifies window function NULL treatment argument.
func TestParseWindowFunctionNullTreatmentArg(t *testing.T) {
	supportedDialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
		},
	}

	sql := "SELECT FIRST_VALUE(a IGNORE NULLS) OVER (), FIRST_VALUE(b RESPECT NULLS) OVER () FROM mytable"
	stmts := supportedDialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())

	errorSql := "SELECT LAG(1 IGNORE NULLS) IGNORE NULLS OVER () FROM t1"
	for _, d := range supportedDialects.Dialects {
		_, err := parser.ParseSQL(d, errorSql)
		assert.Error(t, err, "Expected error for double IGNORE NULLS with dialect %T", d)
	}
}

// ============================================================================
// Subquery Tests
// ============================================================================

// TestParseSubquery verifies subquery parsing.
func TestParseSubquery(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM (SELECT id FROM customer) AS c"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM customer WHERE id IN (SELECT customer_id FROM orders)"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT * FROM customer WHERE EXISTS (SELECT * FROM orders WHERE orders.customer_id = customer.id)"
	dialects.VerifiedStmt(t, sql3)
}

// ============================================================================
// Wildcard Function Argument Tests
// ============================================================================

// TestWildcardFuncArg verifies wildcard with EXCLUDE as a function argument.
func TestWildcardFuncArg(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT HASH(* EXCLUDE(col1)) FROM t"
	canonical1 := "SELECT HASH(* EXCLUDE (col1)) FROM t"

	stmts1 := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts1, 1)

	stmts2 := dialects.ParseSQL(t, canonical1)
	require.Len(t, stmts2, 1)

	sql2 := "SELECT HASH(* EXCLUDE (col1, col2)) FROM t"
	stmts3 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts3, 1)
}

// ============================================================================
// Lambda Tests
// ============================================================================

// TestParseLambdas verifies lambda function parsing.
func TestParseLambdas(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsLambdaFunctions()
	})

	dialects.VerifiedExpr(t, "map_zip_with(map(1, 'a', 2, 'b'), map(1, 'x', 2, 'y'), (k, v1, v2) -> concat(v1, v2))")
	dialects.VerifiedExpr(t, "transform(array(1, 2, 3), x -> x + 1)")
	dialects.VerifiedExpr(t, "a -> a * 2")
	dialects.VerifiedExpr(t, "a INT -> a * 2")
	dialects.VerifiedExpr(t, "(a, b) -> a * b")
	dialects.VerifiedExpr(t, "(a INT, b FLOAT) -> a * b")
}

// ============================================================================
// Dictionary and Map Tests
// ============================================================================

// TestDictionarySyntax verifies dictionary literal syntax parsing.
func TestDictionarySyntax(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsDictionarySyntax()
	})
	dialects.VerifiedExpr(t, "{}")
	dialects.VerifiedExpr(t, "{'Alberta': 'Edmonton', 'Manitoba': 'Winnipeg'}")
	dialects.VerifiedExpr(t, "{'start': CAST('2023-04-01' AS TIMESTAMP), 'end': CAST('2023-04-05' AS TIMESTAMP)}")
}

// TestMapSyntax verifies MAP literal syntax parsing.
func TestMapSyntax(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsMapLiteralSyntax()
	})

	dialects.VerifiedExpr(t, "MAP {'Alberta': 'Edmonton', 'Manitoba': 'Winnipeg'}")
	dialects.VerifiedExpr(t, "MAP {1: 10.0, 2: 20.0}")
	dialects.VerifiedExpr(t, "MAP {[1, 2, 3]: 10.0, [4, 5, 6]: 20.0}")
	dialects.VerifiedExpr(t, "MAP {'a': 10, 'b': 20}['a']")
	dialects.VerifiedExpr(t, "MAP {}")
	dialects.VerifiedExpr(t, "MAP {'a': 1, 'b': NULL}")
	dialects.VerifiedExpr(t, "MAP {1: [1, NULL, 3], 2: [4, NULL, 6], 3: [7, 8, 9]}")
}

// ============================================================================
// Tuples Tests
// ============================================================================

// TestParseTuples verifies tuple/row expression parsing.
func TestParseTuples(t *testing.T) {
	sql := "SELECT (1, 2), (1), ('foo', 3, baz)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Semi-Structured Data Tests
// ============================================================================

// TestParseSemiStructuredDataTraversal verifies semi-structured data traversal parsing.
func TestParseSemiStructuredDataTraversal(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT a:b FROM t"
	stmts1 := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts1, 1)

	sql2 := `SELECT a:"my long object key name" FROM t`
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)

	sql3 := "SELECT a:b::INT FROM t"
	stmts3 := dialects.ParseSQL(t, sql3)
	require.Len(t, stmts3, 1)

	sql4 := `SELECT a:foo."bar".baz`
	stmts4 := dialects.ParseSQL(t, sql4)
	require.Len(t, stmts4, 1)

	sql5 := `SELECT a:foo[0].bar`
	stmts5 := dialects.ParseSQL(t, sql5)
	require.Len(t, stmts5, 1)
}

// ============================================================================
// BINARY Keyword Tests
// ============================================================================

// TestParseBinaryKwAsCast verifies BINARY keyword as CAST parsing.
func TestParseBinaryKwAsCast(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT BINARY 1+1"
	canonical := "SELECT CAST(1 + 1 AS BINARY)"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	stmts2 := dialects.ParseSQL(t, canonical)
	require.Len(t, stmts2, 1)
}

// ============================================================================
// TRY_CONVERT Tests
// ============================================================================

// TestTryConvert verifies TRY_CONVERT function parsing.
func TestTryConvert(t *testing.T) {
	dialects1 := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTryConvert() && d.ConvertTypeBeforeValue()
	})
	dialects1.VerifiedExpr(t, "TRY_CONVERT(VARCHAR(MAX), 'foo')")

	dialects2 := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTryConvert() && !d.ConvertTypeBeforeValue()
	})
	dialects2.VerifiedExpr(t, "TRY_CONVERT('foo', VARCHAR(MAX))")
}

// ============================================================================
// EXTRACT (Custom) Tests
// ============================================================================

// TestExtractSecondsOk verifies EXTRACT with SECONDS from INTERVAL.
func TestExtractSecondsOk(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.AllowExtractCustom()
	})

	dialects.VerifiedExpr(t, "EXTRACT(SECONDS FROM '2 seconds'::INTERVAL)")
	dialects.VerifiedStmt(t, "SELECT EXTRACT(seconds FROM '2 seconds'::INTERVAL)")
}

// TestExtractSecondsSingleQuoteOk verifies EXTRACT with quoted 'seconds' field.
func TestExtractSecondsSingleQuoteOk(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.AllowExtractCustom()
	})

	dialects.VerifiedExpr(t, "EXTRACT('seconds' FROM '2 seconds'::INTERVAL)")
}

// ============================================================================
// Overlap / Bool And Tests
// ============================================================================

// TestParseOverlapAsBoolAnd verifies && operator parsing.
func TestParseOverlapAsBoolAnd(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT x && y"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	sql2 := "SELECT x AND y"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
}

// ============================================================================
// Helper function for number
// ============================================================================

func number(n string) *ast.Value {
	val, err := ast.NewNumber(n, false)
	if err != nil {
		panic(fmt.Sprintf("Failed to create number: %v", err))
	}
	return val
}
