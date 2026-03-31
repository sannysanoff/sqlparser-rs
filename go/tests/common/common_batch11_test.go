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
// This file contains tests 181-200 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseLiteralString verifies parsing of string literals including national and hex strings.
// Reference: tests/sqlparser_common.rs:6089
func TestParseLiteralString(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT 'one', N'national string', X'deadBEEF'"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify round-trip for hex string literal variations
	dialects.OneStatementParsesTo(t, "SELECT x'deadBEEF'", "SELECT X'deadBEEF'")
	dialects.OneStatementParsesTo(t, "SELECT n'national string'", "SELECT N'national string'")
}

// TestParseLiteralDate verifies parsing of DATE typed string literals.
// Reference: tests/sqlparser_common.rs:6113
func TestParseLiteralDate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DATE '1999-01-01'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseLiteralTime verifies parsing of TIME typed string literals.
// Reference: tests/sqlparser_common.rs:6130
func TestParseLiteralTime(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT TIME '01:23:34'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseLiteralDatetime verifies parsing of DATETIME typed string literals.
// Reference: tests/sqlparser_common.rs:6147
func TestParseLiteralDatetime(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DATETIME '1999-01-01 01:23:34.45'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseLiteralTimestampWithoutTimeZone verifies parsing of TIMESTAMP without time zone.
// Reference: tests/sqlparser_common.rs:6164
func TestParseLiteralTimestampWithoutTimeZone(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT TIMESTAMP '1999-01-01 01:23:34'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseLiteralTimestampWithTimeZone verifies parsing of TIMESTAMPTZ with time zone.
// Reference: tests/sqlparser_common.rs:6183
func TestParseLiteralTimestampWithTimeZone(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT TIMESTAMPTZ '1999-01-01 01:23:34Z'"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseIntervalAll verifies parsing of various INTERVAL expressions.
// Reference: tests/sqlparser_common.rs:6202
func TestParseIntervalAll(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// INTERVAL with YEAR TO MONTH
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1-1' YEAR TO MONTH")

	// INTERVAL with MINUTE TO SECOND and precision
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '01:01.01' MINUTE (5) TO SECOND (5)")

	// INTERVAL with SECOND and precision
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' SECOND (5, 4)")

	// INTERVAL with HOUR
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '10' HOUR")

	// INTERVAL with numeric expression and DAY
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 5 DAY")

	// INTERVAL with DAYS (plural)
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 5 DAYS")

	// INTERVAL with HOUR and precision
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '10' HOUR (1)")

	// Test invalid interval expressions produce errors
	dialect := generic.NewGenericDialect()

	// SECOND TO SECOND should fail
	_, err := parser.ParseSQL(dialect, "SELECT INTERVAL '1' SECOND TO SECOND")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: end of statement, found: SECOND")

	// HOUR (1) TO HOUR (2) should fail
	_, err = parser.ParseSQL(dialect, "SELECT INTERVAL '10' HOUR (1) TO HOUR (2)")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: end of statement, found: (")

	// Various single unit intervals
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' YEAR")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' MONTH")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' WEEK")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' DAY")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' HOUR")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' MINUTE")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' SECOND")

	// Plural forms
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' YEARS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' MONTHS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' WEEKS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' DAYS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' HOURS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' MINUTES")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' SECONDS")

	// Range intervals
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' YEAR TO MONTH")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' DAY TO HOUR")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' DAY TO MINUTE")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' DAY TO SECOND")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' HOUR TO MINUTE")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' HOUR TO SECOND")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL '1' MINUTE TO SECOND")

	// Numeric intervals
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 YEAR")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 MONTH")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 WEEK")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 DAY")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 HOUR")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 MINUTE")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 SECOND")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 YEARS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 MONTHS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 WEEKS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 DAYS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 HOURS")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 MINUTES")
	dialects.VerifiedOnlySelect(t, "SELECT INTERVAL 1 SECONDS")

	// PostgreSQL-style INTERVAL cast
	dialects.VerifiedOnlySelect(t, "SELECT '2 years 15 months 100 weeks 99 hours 123456789 milliseconds'::INTERVAL")
}

// TestParseIntervalDontRequireUnit verifies INTERVAL without explicit unit for dialects that don't require it.
// Reference: tests/sqlparser_common.rs:6359
func TestParseIntervalDontRequireUnit(t *testing.T) {
	// For dialects that don't require interval qualifier (like PostgreSQL)
	dialect := postgresql.NewPostgreSqlDialect()

	sql := "SELECT INTERVAL '1 DAY'"
	stmts := utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)

	// Test various intervals without explicit qualifiers
	utils.MustParseSQL(t, dialect, "SELECT INTERVAL '1 YEAR'")
	utils.MustParseSQL(t, dialect, "SELECT INTERVAL '1 MONTH'")
	utils.MustParseSQL(t, dialect, "SELECT INTERVAL '1 DAY'")
	utils.MustParseSQL(t, dialect, "SELECT INTERVAL '1 HOUR'")
	utils.MustParseSQL(t, dialect, "SELECT INTERVAL '1 MINUTE'")
	utils.MustParseSQL(t, dialect, "SELECT INTERVAL '1 SECOND'")
}

// TestParseIntervalRequireUnit verifies that dialects requiring interval qualifiers error on missing units.
// Reference: tests/sqlparser_common.rs:6385
func TestParseIntervalRequireUnit(t *testing.T) {
	// MySQL requires interval qualifiers
	dialect := mysql.NewMySqlDialect()

	sql := "SELECT INTERVAL '1 DAY'"
	_, err := parser.ParseSQL(dialect, sql)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "INTERVAL requires a unit")
}

// TestParseIntervalRequireQualifier verifies INTERVAL with expressions for dialects requiring qualifiers.
// Reference: tests/sqlparser_common.rs:6396
func TestParseIntervalRequireQualifier(t *testing.T) {
	// MySQL requires interval qualifiers
	dialect := mysql.NewMySqlDialect()

	// INTERVAL with arithmetic expression
	sql := "SELECT INTERVAL 1 + 1 DAY"
	stmts := utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)

	// INTERVAL with string concatenation
	sql = "SELECT INTERVAL '1' + '1' DAY"
	stmts = utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)

	// INTERVAL with complex arithmetic
	sql = "SELECT INTERVAL '1' + '2' - '3' DAY"
	stmts = utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)
}

// TestParseIntervalDisallowIntervalExpr verifies INTERVAL comparisons work for dialects without qualifier requirement.
// Reference: tests/sqlparser_common.rs:6466
func TestParseIntervalDisallowIntervalExpr(t *testing.T) {
	// For dialects that don't require interval qualifier (like PostgreSQL)
	dialect := postgresql.NewPostgreSqlDialect()

	sql := "SELECT INTERVAL '1 DAY'"
	stmts := utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)

	// Various interval expressions
	utils.MustParseSQL(t, dialect, "SELECT INTERVAL '1 YEAR'")
	utils.MustParseSQL(t, dialect, "SELECT INTERVAL '1 YEAR' AS one_year")

	// INTERVAL comparison
	sql = "SELECT INTERVAL '1 DAY' > INTERVAL '1 SECOND'"
	stmts = utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)
}

// TestIntervalDisallowIntervalExprGt verifies INTERVAL greater-than comparisons.
// Reference: tests/sqlparser_common.rs:6520
func TestIntervalDisallowIntervalExprGt(t *testing.T) {
	// For dialects that don't require interval qualifier
	dialect := postgresql.NewPostgreSqlDialect()

	sql := "SELECT INTERVAL '1 second' > x"
	stmts := utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)
}

// TestIntervalDisallowIntervalExprDoubleColon verifies INTERVAL casting with double colon.
// Reference: tests/sqlparser_common.rs:6546
func TestIntervalDisallowIntervalExprDoubleColon(t *testing.T) {
	// For dialects that don't require interval qualifier (PostgreSQL supports :: cast)
	dialect := postgresql.NewPostgreSqlDialect()

	sql := "SELECT INTERVAL '1 second'::TEXT"
	stmts := utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)
}

// TestParseIntervalAndOrXor verifies INTERVAL expressions in AND/OR/XOR conditions.
// Reference: tests/sqlparser_common.rs:6570
func TestParseIntervalAndOrXor(t *testing.T) {
	// For dialects that don't require interval qualifier
	dialect := postgresql.NewPostgreSqlDialect()

	// AND condition with INTERVAL
	sql := "SELECT col FROM test WHERE d3_date > d1_date + INTERVAL '5 days' AND d2_date > d1_date + INTERVAL '3 days'"
	stmts := utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)

	// OR condition with INTERVAL
	sql = "SELECT col FROM test WHERE d3_date > d1_date + INTERVAL '5 days' OR d2_date > d1_date + INTERVAL '3 days'"
	stmts = utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)

	// XOR condition with INTERVAL
	sql = "SELECT col FROM test WHERE d3_date > d1_date + INTERVAL '5 days' XOR d2_date > d1_date + INTERVAL '3 days'"
	stmts = utils.MustParseSQL(t, dialect, sql)
	require.Len(t, stmts, 1)
}

// TestParseAtTimezone verifies AT TIME ZONE expressions.
// Reference: tests/sqlparser_common.rs:6701
func TestParseAtTimezone(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic AT TIME ZONE
	sql := "SELECT FROM_UNIXTIME(0) AT TIME ZONE 'UTC-06:00' FROM t"
	dialects.VerifiedOnlySelect(t, sql)

	// AT TIME ZONE with nested function and alias
	sql = `SELECT DATE_FORMAT(FROM_UNIXTIME(0) AT TIME ZONE 'UTC-06:00', '%Y-%m-%dT%H') AS "hour" FROM t`
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseJsonKeyword verifies parsing of JSON typed string literals.
// Reference: tests/sqlparser_common.rs:6744
func TestParseJsonKeyword(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := `SELECT JSON '{
  "id": 10,
  "type": "fruit",
  "name": "apple",
  "on_menu": true,
  "recipes":
    {
      "salads":
      [
        { "id": 2001, "type": "Walnut Apple Salad" },
        { "id": 2002, "type": "Apple Spinach Salad" }
      ],
      "desserts":
      [
        { "id": 3001, "type": "Apple Pie" },
        { "id": 3002, "type": "Apple Scones" },
        { "id": 3003, "type": "Apple Crumble" }
      ]
    }
}'`
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseTypedStrings verifies parsing of various typed string literals.
// Reference: tests/sqlparser_common.rs:6802
func TestParseTypedStrings(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := `JSON '{"foo":"bar"}'`
	stmts := dialects.ParseSQL(t, "SELECT "+sql)
	require.Len(t, stmts, 1)
}

// TestParseBignumericKeyword verifies parsing of BIGNUMERIC typed string literals.
// Reference: tests/sqlparser_common.rs:6828
func TestParseBignumericKeyword(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Various BIGNUMERIC literals
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '0'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '123456'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '-3.14'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '-0.54321'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '1.23456e05'")
	dialects.VerifiedOnlySelect(t, "SELECT BIGNUMERIC '-9.876e-3'")
}

// TestParseSimpleMathExprPlus verifies parsing of simple addition expressions.
// Reference: tests/sqlparser_common.rs:6921
func TestParseSimpleMathExprPlus(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a + b, 2 + a, 2.5 + a, a_f + b_f, 2 + a_f, 2.5 + a_f FROM c"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseSimpleMathExprMinus verifies parsing of simple subtraction expressions.
// Reference: tests/sqlparser_common.rs:6927
func TestParseSimpleMathExprMinus(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a - b, 2 - a, 2.5 - a, a_f - b_f, 2 - a_f, 2.5 - a_f FROM c"
	dialects.VerifiedOnlySelect(t, sql)
}
