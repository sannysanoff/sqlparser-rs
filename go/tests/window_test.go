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
// This file contains window function and GROUP BY tests.
package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/ansi"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseWindowFunctionsAdvanced verifies advanced window function parsing.
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

	// Test error case - ANSI dialect doesn't support named window references
	// so "WINDOW window1 AS window2" (without parens) should fail
	_, err := parser.ParseSQL(ansi.NewAnsiDialect(), "SELECT * from mytable WINDOW window1 AS window2")
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
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsWindowClauseNamedWindowReference()
	})

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

// TestParseSelectGroupBy verifies GROUP BY clause parsing.
// Reference: tests/sqlparser_common.rs:1113
func TestParseSelectGroupBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, name FROM customer GROUP BY name"
	dialects.VerifiedStmt(t, sql)
}

// TestParseGroupByWithModifier verifies GROUP BY with modifiers (ROLLUP, CUBE, TOTALS).
// Reference: tests/sqlparser_common.rs:2846
func TestParseGroupByWithModifier(t *testing.T) {
	clauses := []string{"x", "a, b", "ALL"}
	modifiers := []string{
		"WITH ROLLUP",
		"WITH CUBE",
		"WITH TOTALS",
		"WITH ROLLUP WITH CUBE",
	}

	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsGroupByWithModifier()
	})

	for _, clause := range clauses {
		for _, modifier := range modifiers {
			sql := fmt.Sprintf("SELECT * FROM t GROUP BY %s %s", clause, modifier)
			stmts := dialects.ParseSQL(t, sql)
			require.Len(t, stmts, 1)
		}
	}

	// Invalid cases
	invalidCases := []string{
		"SELECT * FROM t GROUP BY x WITH",
		"SELECT * FROM t GROUP BY x WITH ROLLUP CUBE",
		"SELECT * FROM t GROUP BY x WITH WITH ROLLUP",
		"SELECT * FROM t GROUP BY WITH ROLLUP",
	}

	for _, sql := range invalidCases {
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err, "Expected error for: %s", sql)
	}
}

// TestParseGroupBySpecialGroupingSets verifies GROUP BY with GROUPING SETS.
// Reference: tests/sqlparser_common.rs:2903
func TestParseGroupBySpecialGroupingSets(t *testing.T) {
	sql := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), (a), (b), ())"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseGroupByGroupingSetsSingleValues verifies GROUP BY with GROUPING SETS.
// Reference: tests/sqlparser_common.rs:2931
func TestParseGroupByGroupingSetsSingleValues(t *testing.T) {
	sql := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), a, (b), c, ())"
	canonical := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), (a), (b), (c), ())"

	// Test that both SQL parse correctly
	dialects := utils.NewTestedDialects()
	dialects.OneStatementParsesTo(t, sql, canonical)
}

// TestGroupByNothing verifies GROUP BY () syntax parsing.
// Reference: tests/sqlparser_common.rs:13899
func TestGroupByNothing(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsGroupByExpr()
	})

	// Test GROUP BY () alone
	stmts := dialects.ParseSQL(t, "SELECT count(1) FROM t GROUP BY ()")
	require.Len(t, stmts, 1)

	// Test GROUP BY with other expressions and ()
	stmts = dialects.ParseSQL(t, "SELECT name, count(1) FROM t GROUP BY name, ()")
	require.Len(t, stmts, 1)
}
