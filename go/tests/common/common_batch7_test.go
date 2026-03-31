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

// Package common contains the common SQL parsing tests.
// These tests are ported from tests/sqlparser_common.rs in the Rust implementation.
// This file contains tests 101-120 from the Rust test suite.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/ansi"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/hive"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseGroupByGroupingSetsSingleValues verifies GROUP BY with GROUPING SETS.
// Reference: tests/sqlparser_common.rs:2931
func TestParseGroupByGroupingSetsSingleValues(t *testing.T) {
	sql := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), a, (b), c, ())"
	canonical := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), (a), (b), (c), ())"

	// Test that both SQL parse correctly
	dialects := utils.NewTestedDialects()
	dialects.OneStatementParsesTo(t, sql, canonical)
}

// TestParseSelectHavingClause verifies HAVING clause parsing.
// Reference: tests/sqlparser_common.rs:2963
func TestParseSelectHavingClause(t *testing.T) {
	sql := "SELECT foo FROM bar GROUP BY foo HAVING COUNT(*) > 1"

	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify the statement round-trips
	assert.Equal(t, sql, stmts[0].String())

	// Second test case - HAVING without GROUP BY
	sql2 := "SELECT 'foo' HAVING 1 = 1"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseSelectQualify verifies QUALIFY clause parsing.
// Reference: tests/sqlparser_common.rs:2994
func TestParseSelectQualify(t *testing.T) {
	sql := "SELECT i, p, o FROM qt QUALIFY ROW_NUMBER() OVER (PARTITION BY p ORDER BY o) = 1"

	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify the statement round-trips
	assert.Equal(t, sql, stmts[0].String())

	// Second test case - QUALIFY with aliased window function
	sql2 := "SELECT i, p, o, ROW_NUMBER() OVER (PARTITION BY p ORDER BY o) AS row_num FROM qt QUALIFY row_num = 1"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
	assert.Equal(t, sql2, stmts2[0].String())
}

// TestParseLimitAcceptsAll verifies LIMIT ALL syntax.
// Reference: tests/sqlparser_common.rs:3044
func TestParseLimitAcceptsAll(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// LIMIT ALL should be equivalent to no LIMIT
	dialects.OneStatementParsesTo(
		t,
		"SELECT id, fname, lname FROM customer WHERE id = 1 LIMIT ALL",
		"SELECT id, fname, lname FROM customer WHERE id = 1",
	)

	dialects.OneStatementParsesTo(
		t,
		"SELECT id, fname, lname FROM customer WHERE id = 1 LIMIT ALL OFFSET 1",
		"SELECT id, fname, lname FROM customer WHERE id = 1 OFFSET 1",
	)

	dialects.OneStatementParsesTo(
		t,
		"SELECT id, fname, lname FROM customer WHERE id = 1 OFFSET 1 LIMIT ALL",
		"SELECT id, fname, lname FROM customer WHERE id = 1 OFFSET 1",
	)
}

// TestParseCast verifies CAST expression parsing.
// Reference: tests/sqlparser_common.rs:3060
func TestParseCast(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic CAST expressions
	dialects.VerifiedStmt(t, "SELECT CAST(id AS BIGINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS TINYINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS MEDIUMINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS NUMERIC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS DEC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS DECIMAL) FROM customer")

	// CAST with length
	dialects.VerifiedStmt(t, "SELECT CAST(id AS NVARCHAR(50)) FROM customer")

	// CLOB variants
	dialects.VerifiedStmt(t, "SELECT CAST(id AS CLOB) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS CLOB(50)) FROM customer")

	// BINARY and VARBINARY
	dialects.VerifiedStmt(t, "SELECT CAST(id AS BINARY(50)) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS VARBINARY(50)) FROM customer")

	// BLOB variants
	dialects.VerifiedStmt(t, "SELECT CAST(id AS BLOB) FROM customer")
	dialects.VerifiedStmt(t, "SELECT CAST(id AS BLOB(50)) FROM customer")

	// JSONB
	dialects.VerifiedStmt(t, "SELECT CAST(details AS JSONB) FROM customer")
}

// TestParseTryCast verifies TRY_CAST expression parsing.
// Reference: tests/sqlparser_common.rs:3212
func TestParseTryCast(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic TRY_CAST expressions
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS BIGINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS NUMERIC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS DEC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS DECIMAL) FROM customer")
}

// TestParseExtract verifies EXTRACT expression parsing.
// Reference: tests/sqlparser_common.rs:3235
func TestParseExtract(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic EXTRACT expressions
	dialects.VerifiedStmt(t, "SELECT EXTRACT(YEAR FROM d)")

	// Case-insensitive
	dialects.OneStatementParsesTo(t, "SELECT EXTRACT(year from d)", "SELECT EXTRACT(YEAR FROM d)")

	// Various date/time fields
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

// TestParseCeilNumber verifies CEIL function parsing.
// Reference: tests/sqlparser_common.rs:3294
func TestParseCeilNumber(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT CEIL(1.5)")
	dialects.VerifiedStmt(t, "SELECT CEIL(float_column) FROM my_table")
}

// TestParseFloorNumber verifies FLOOR function parsing.
// Reference: tests/sqlparser_common.rs:3300
func TestParseFloorNumber(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT FLOOR(1.5)")
	dialects.VerifiedStmt(t, "SELECT FLOOR(float_column) FROM my_table")
}

// TestParseCeilNumberScale verifies CEIL with scale parameter.
// Reference: tests/sqlparser_common.rs:3306
func TestParseCeilNumberScale(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT CEIL(1.5, 1)")
	dialects.VerifiedStmt(t, "SELECT CEIL(float_column, 3) FROM my_table")
}

// TestParseFloorNumberScale verifies FLOOR with scale parameter.
// Reference: tests/sqlparser_common.rs:3312
func TestParseFloorNumberScale(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT FLOOR(1.5, 1)")
	dialects.VerifiedStmt(t, "SELECT FLOOR(float_column, 3) FROM my_table")
}

// TestParseCeilScale verifies CEIL with scale parameter (detailed).
// Reference: tests/sqlparser_common.rs:3318
func TestParseCeilScale(t *testing.T) {
	sql := "SELECT CEIL(d, 2)"

	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify the statement round-trips
	assert.Equal(t, sql, stmts[0].String())
}

// TestParseFloorScale verifies FLOOR with scale parameter (detailed).
// Reference: tests/sqlparser_common.rs:3344
func TestParseFloorScale(t *testing.T) {
	sql := "SELECT FLOOR(d, 2)"

	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify the statement round-trips
	assert.Equal(t, sql, stmts[0].String())
}

// TestParseCeilDatetime verifies CEIL with datetime field.
// Reference: tests/sqlparser_common.rs:3370
func TestParseCeilDatetime(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic CEIL TO DAY
	sql := "SELECT CEIL(d TO DAY)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())

	// Case-insensitive
	dialects.OneStatementParsesTo(t, "SELECT CEIL(d to day)", "SELECT CEIL(d TO DAY)")

	// Various datetime fields
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO HOUR) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO MINUTE) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO SECOND) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO MILLISECOND) FROM df")
}

// TestParseFloorDatetime verifies FLOOR with datetime field.
// Reference: tests/sqlparser_common.rs:3397
func TestParseFloorDatetime(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic FLOOR TO DAY
	sql := "SELECT FLOOR(d TO DAY)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())

	// Case-insensitive
	dialects.OneStatementParsesTo(t, "SELECT FLOOR(d to day)", "SELECT FLOOR(d TO DAY)")

	// Various datetime fields
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO HOUR) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO MINUTE) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO SECOND) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO MILLISECOND) FROM df")
}

// TestParseListagg verifies LISTAGG function parsing.
// Reference: tests/sqlparser_common.rs:3424
func TestParseListagg(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Complex LISTAGG with all options
	sql := "SELECT LISTAGG(DISTINCT dateid, ', ' ON OVERFLOW TRUNCATE '%' WITHOUT COUNT) WITHIN GROUP (ORDER BY id, username)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())

	// Simpler variants
	dialects.VerifiedStmt(t, "SELECT LISTAGG(sellerid) WITHIN GROUP (ORDER BY dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(DISTINCT dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW ERROR)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW TRUNCATE N'...' WITH COUNT)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW TRUNCATE X'deadbeef' WITH COUNT)")
}

// TestParseArrayAggFunc verifies ARRAY_AGG function parsing.
// Reference: tests/sqlparser_common.rs:3496
func TestParseArrayAggFunc(t *testing.T) {
	// Test with specific dialects that support ARRAY_AGG
	supportedDialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
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

// TestParseAggWithOrderBy verifies aggregate functions with ORDER BY.
// Reference: tests/sqlparser_common.rs:3518
func TestParseAggWithOrderBy(t *testing.T) {
	// Test with specific dialects that support aggregate with ORDER BY
	supportedDialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
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

// TestParseWindowRankFunction verifies window rank functions parsing.
// Reference: tests/sqlparser_common.rs:3538
func TestParseWindowRankFunction(t *testing.T) {
	// Test with dialects that support window functions
	supportedDialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
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

	// Test with dialects that support IGNORE/RESPECT NULLS
	supportedDialectsNulls := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
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
// Reference: tests/sqlparser_common.rs:3574
func TestParseWindowFunctionNullTreatmentArg(t *testing.T) {
	// Test with dialects that support window function null treatment
	supportedDialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
		},
	}

	sql := "SELECT FIRST_VALUE(a IGNORE NULLS) OVER (), FIRST_VALUE(b RESPECT NULLS) OVER () FROM mytable"
	stmts := supportedDialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify the statement round-trips
	assert.Equal(t, sql, stmts[0].String())

	// Test error case - double IGNORE NULLS should fail
	errorSql := "SELECT LAG(1 IGNORE NULLS) IGNORE NULLS OVER () FROM t1"
	for _, d := range supportedDialects.Dialects {
		_, err := parser.ParseSQL(d, errorSql)
		assert.Error(t, err, "Expected error for double IGNORE NULLS with dialect %T", d)
	}
}
