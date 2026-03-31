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
// This file contains tests 161-180 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseAlterTableDropConstraint verifies ALTER TABLE DROP CONSTRAINT parsing.
// Reference: tests/sqlparser_common.rs:5291
func TestParseAlterTableDropConstraint(t *testing.T) {
	dialects := utils.NewTestedDialects()

	checkOne := func(t *testing.T, constraintText string) {
		sql := "ALTER TABLE tab " + constraintText
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		alterTable, ok := stmts[0].(*statement.AlterTable)
		require.True(t, ok, "Expected AlterTable statement, got %T", stmts[0])

		assert.Equal(t, "tab", alterTable.Name.String())
		require.NotEmpty(t, alterTable.Operations, "Expected at least one operation")
	}

	checkOne(t, "DROP CONSTRAINT IF EXISTS constraint_name")
	checkOne(t, "DROP CONSTRAINT IF EXISTS constraint_name RESTRICT")
	checkOne(t, "DROP CONSTRAINT IF EXISTS constraint_name CASCADE")

	// Test parsing error for invalid syntax
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "ALTER TABLE tab DROP CONSTRAINT is_active TEXT")
	require.Error(t, err)
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

// TestParseScalarFunctionInProjection verifies scalar function parsing in SELECT.
// Reference: tests/sqlparser_common.rs:5339
func TestParseScalarFunctionInProjection(t *testing.T) {
	dialects := utils.NewTestedDialects()
	names := []string{"sqrt", "foo"}

	for _, functionName := range names {
		sql := "SELECT " + functionName + "(id) FROM foo"
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		selectStmt, ok := stmts[0].(*parser.SelectStatement)
		require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
		require.NotNil(t, selectStmt)
	}
}

// runExplainAnalyze is a helper function for EXPLAIN ANALYZE tests
func runExplainAnalyze(t *testing.T, sql string, expectedVerbose, expectedAnalyze bool) {
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	explain, ok := stmts[0].(*statement.Explain)
	require.True(t, ok, "Expected Explain statement, got %T", stmts[0])

	assert.Equal(t, expectedVerbose, explain.Verbose, "VERBOSE mismatch")
	assert.Equal(t, expectedAnalyze, explain.Analyze, "ANALYZE mismatch")
	assert.False(t, explain.QueryPlan, "QueryPlan should be false")
}

// TestParseExplainTable verifies EXPLAIN TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:5385
func TestParseExplainTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	validateExplain := func(t *testing.T, sql string) {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		explainTable, ok := stmts[0].(*statement.ExplainTable)
		require.True(t, ok, "Expected ExplainTable statement, got %T", stmts[0])

		assert.Equal(t, "test_identifier", explainTable.TableName.String())
	}

	validateExplain(t, "EXPLAIN test_identifier")
	validateExplain(t, "DESCRIBE test_identifier")
	validateExplain(t, "DESC test_identifier")
}

// TestExplainDescribe verifies DESCRIBE statement parsing.
// Reference: tests/sqlparser_common.rs:5410
func TestExplainDescribe(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "DESCRIBE test.table")
}

// TestExplainDesc verifies DESC statement parsing.
// Reference: tests/sqlparser_common.rs:5415
func TestExplainDesc(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "DESC test.table")
}

// TestParseExplainAnalyzeWithSimpleSelect verifies EXPLAIN ANALYZE parsing.
// Reference: tests/sqlparser_common.rs:5420
func TestParseExplainAnalyzeWithSimpleSelect(t *testing.T) {
	// DESCRIBE is an alias for EXPLAIN
	runExplainAnalyze(t, "DESCRIBE SELECT sqrt(id) FROM foo", false, false)
	runExplainAnalyze(t, "EXPLAIN SELECT sqrt(id) FROM foo", false, false)
	runExplainAnalyze(t, "EXPLAIN VERBOSE SELECT sqrt(id) FROM foo", true, false)
	runExplainAnalyze(t, "EXPLAIN ANALYZE SELECT sqrt(id) FROM foo", false, true)
	runExplainAnalyze(t, "EXPLAIN ANALYZE VERBOSE SELECT sqrt(id) FROM foo", true, true)
}

// TestParseExplainQueryPlan verifies EXPLAIN QUERY PLAN parsing.
// Reference: tests/sqlparser_common.rs:5538
func TestParseExplainQueryPlan(t *testing.T) {
	dialects := utils.NewTestedDialects()

	stmts := dialects.ParseSQL(t, "EXPLAIN QUERY PLAN SELECT sqrt(id) FROM foo")
	require.Len(t, stmts, 1)

	explain, ok := stmts[0].(*statement.Explain)
	require.True(t, ok, "Expected Explain statement, got %T", stmts[0])

	assert.True(t, explain.QueryPlan)
	assert.False(t, explain.Analyze)
	assert.False(t, explain.Verbose)

	// Omit QUERY PLAN should be good
	dialects.VerifiedStmt(t, "EXPLAIN SELECT sqrt(id) FROM foo")

	// Missing PLAN keyword should return error
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "EXPLAIN QUERY SELECT sqrt(id) FROM foo")
	require.Error(t, err)
}

// TestParseExplainEstimate verifies EXPLAIN ESTIMATE parsing.
// Reference: tests/sqlparser_common.rs:5568
func TestParseExplainEstimate(t *testing.T) {
	dialects := utils.NewTestedDialects()

	stmts := dialects.ParseSQL(t, "EXPLAIN ESTIMATE SELECT sqrt(id) FROM foo")
	require.Len(t, stmts, 1)

	explain, ok := stmts[0].(*statement.Explain)
	require.True(t, ok, "Expected Explain statement, got %T", stmts[0])

	assert.True(t, explain.Estimate)
	assert.False(t, explain.QueryPlan)
	assert.False(t, explain.Analyze)
	assert.False(t, explain.Verbose)
}

// TestParseNamedArgumentFunction verifies named function arguments with => operator.
// Reference: tests/sqlparser_common.rs:5596
func TestParseNamedArgumentFunction(t *testing.T) {
	// Use dialects that support named fn args with => but not with expression names
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
		},
	}

	sql := "SELECT FUN(a => '1', b => '2') FROM foo"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.NotNil(t, selectStmt)
}

// TestParseNamedArgumentFunctionWithEqOperator verifies named function arguments with = operator.
// Reference: tests/sqlparser_common.rs:5639
func TestParseNamedArgumentFunctionWithEqOperator(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT FUN(a = '1', b = '2') FROM foo"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.NotNil(t, selectStmt)

	// Test iff function parses for all dialects
	sql2 := "iff(1 = 1, 1, 0)"
	stmts2 := dialects.ParseSQL(t, "SELECT "+sql2+" FROM t")
	require.Len(t, stmts2, 1)
}

// TestParseWindowFunctionsAdvanced verifies advanced window function parsing.
// This is a more comprehensive version of window function tests.
// Reference: tests/sqlparser_common.rs:5698
func TestParseWindowFunctionsAdvanced(t *testing.T) {
	sql := "SELECT row_number() OVER (ORDER BY dt DESC), " +
		"sum(foo) OVER (PARTITION BY a, b ORDER BY c, d ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW), " +
		"avg(bar) OVER (ORDER BY a RANGE BETWEEN 1 PRECEDING AND 1 FOLLOWING), " +
		"sum(bar) OVER (ORDER BY a RANGE BETWEEN INTERVAL '1' DAY PRECEDING AND INTERVAL '1 MONTH' FOLLOWING), " +
		"COUNT(*) OVER (ORDER BY a RANGE BETWEEN INTERVAL '1 DAY' PRECEDING AND INTERVAL '1 DAY' FOLLOWING), " +
		"max(baz) OVER (ORDER BY a ROWS UNBOUNDED PRECEDING), " +
		"sum(qux) OVER (ORDER BY a GROUPS BETWEEN 1 PRECEDING AND 1 FOLLOWING) " +
		"FROM foo"

	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.NotNil(t, selectStmt)
}

// TestParseNamedWindowFunctions verifies named window function parsing.
// Reference: tests/sqlparser_common.rs:5764
func TestParseNamedWindowFunctions(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
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

// TestParseWindowClause verifies WINDOW clause parsing.
// Reference: tests/sqlparser_common.rs:5809
func TestParseWindowClause(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM mytable WINDOW " +
		"window1 AS (ORDER BY 1 ASC, 2 DESC, 3 NULLS FIRST), " +
		"window2 AS (window1), " +
		"window3 AS (PARTITION BY a, b, c), " +
		"window4 AS (ROWS UNBOUNDED PRECEDING), " +
		"window5 AS (window1 PARTITION BY a), " +
		"window6 AS (window1 ORDER BY a), " +
		"window7 AS (window1 ROWS UNBOUNDED PRECEDING), " +
		"window8 AS (window1 PARTITION BY a ORDER BY b ROWS UNBOUNDED PRECEDING) " +
		"ORDER BY C3"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Test error case
	_, err := parser.ParseSQL(bigquery.NewBigQueryDialect(), "SELECT * from mytable WINDOW window1 AS window2")
	require.Error(t, err)
}

// TestParseNamedWindow verifies named window reference parsing.
// Reference: tests/sqlparser_common.rs:5839
func TestParseNamedWindow(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT MIN(c12) OVER window1 AS min1, MAX(c12) OVER window2 AS max1 " +
		"FROM aggregate_test_100 " +
		"WINDOW window1 AS (ORDER BY C12), window2 AS (PARTITION BY C11) " +
		"ORDER BY C3"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.NotNil(t, selectStmt)
}

// TestParseWindowAndQualifyClause verifies WINDOW and QUALIFY clause parsing.
// Reference: tests/sqlparser_common.rs:5999
func TestParseWindowAndQualifyClause(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT MIN(c12) OVER window1 AS min1 " +
		"FROM aggregate_test_100 " +
		"QUALIFY ROW_NUMBER() OVER my_window " +
		"WINDOW window1 AS (ORDER BY C12), window2 AS (PARTITION BY C11) " +
		"ORDER BY C3"

	stmts := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts, 1)

	sql2 := "SELECT MIN(c12) OVER window1 AS min1 " +
		"FROM aggregate_test_100 " +
		"WINDOW window1 AS (ORDER BY C12), window2 AS (PARTITION BY C11) " +
		"QUALIFY ROW_NUMBER() OVER my_window " +
		"ORDER BY C3"

	stmts = dialects.ParseSQL(t, sql2)
	require.Len(t, stmts, 1)
}

// TestParseWindowClauseNamedWindow verifies WINDOW clause with named window reference.
// Reference: tests/sqlparser_common.rs:6024
func TestParseWindowClauseNamedWindow(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM mytable WINDOW window1 AS window2"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseAggregateWithGroupBy verifies aggregate function with GROUP BY parsing.
// Reference: tests/sqlparser_common.rs:6039
func TestParseAggregateWithGroupBy(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT a, COUNT(1), MIN(b), MAX(b) FROM foo GROUP BY a"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.NotNil(t, selectStmt)
}

// TestParseLiteralInteger verifies integer literal parsing.
// Reference: tests/sqlparser_common.rs:6046
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
// Reference: tests/sqlparser_common.rs:6073
func TestParseLiteralDecimal(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// These numbers were explicitly chosen to not roundtrip if represented as
	// f64s (i.e., as 64-bit binary floating point numbers).
	sql := "SELECT 0.300000000000000004, 9007199254740993.0"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected SelectStatement, got %T", stmts[0])
	require.Equal(t, 2, len(selectStmt.Projection))
}

// Note: Test 181 (parse_literal_string) starts at line 6089
// and would be included in the next batch.
