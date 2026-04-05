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
// This file contains SHOW/EXPLAIN/ANALYZE tests.
package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/statement"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestShowDbsSchemasTablesViews verifies SHOW statements.
// Reference: tests/sqlparser_common.rs:14744
func TestShowDbsSchemasTablesViews(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic SHOW statements
	stmts := []string{
		"SHOW DATABASES",
		"SHOW SCHEMAS",
		"SHOW TABLES",
		"SHOW VIEWS",
		"SHOW TABLES IN db1",
		"SHOW VIEWS FROM db1",
		"SHOW MATERIALIZED VIEWS",
		"SHOW MATERIALIZED VIEWS IN db1",
		"SHOW MATERIALIZED VIEWS FROM db1",
	}
	for _, sql := range stmts {
		dialects.VerifiedStmt(t, sql)
	}

	// SHOW with LIKE (dialect-dependent)
	likeStmts := []string{
		"SHOW DATABASES LIKE '%abc'",
		"SHOW SCHEMAS LIKE '%abc'",
	}
	for _, sql := range likeStmts {
		// Dialects that support LIKE before IN
		dialectsLikeBeforeIn := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
			return d.SupportsShowLikeBeforeIn()
		})
		dialectsLikeBeforeIn.VerifiedStmt(t, sql)

		// Dialects that don't support LIKE before IN
		dialectsNoLikeBeforeIn := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
			return !d.SupportsShowLikeBeforeIn()
		})
		dialectsNoLikeBeforeIn.VerifiedStmt(t, sql)
	}

	// SHOW with LIKE in suffix (only for dialects that support it)
	suffixLikeStmts := []string{
		"SHOW TABLES IN db1 'abc'",
		"SHOW VIEWS IN db1 'abc'",
		"SHOW VIEWS FROM db1 'abc'",
		"SHOW MATERIALIZED VIEWS IN db1 'abc'",
		"SHOW MATERIALIZED VIEWS FROM db1 'abc'",
	}
	for _, sql := range suffixLikeStmts {
		dialectsNoLikeBeforeIn := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
			return !d.SupportsShowLikeBeforeIn()
		})
		dialectsNoLikeBeforeIn.VerifiedStmt(t, sql)
	}
}

// TestParseShowFunctions verifies SHOW FUNCTIONS statement parsing.
// Reference: tests/sqlparser_common.rs:10851
func TestParseShowFunctions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	stmt := dialects.VerifiedStmt(t, "SHOW FUNCTIONS LIKE 'pattern'")
	showFunc, ok := stmt.(*statement.ShowFunctions)
	require.True(t, ok, "Expected ShowFunctions statement, got %T", stmt)
	require.NotNil(t, showFunc.Filter, "Expected filter to be present")
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

// runExplainAnalyzeWithOptions is a helper function for EXPLAIN with options tests
func runExplainAnalyzeWithOptions(t *testing.T, dialects *utils.TestedDialects, sql string, expectedOptions []*expr.UtilityOption) {
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	explain, ok := stmts[0].(*statement.Explain)
	require.True(t, ok, "Expected Explain statement, got %T", stmts[0])

	if expectedOptions != nil {
		assert.Equal(t, len(expectedOptions), len(explain.Options))
	}
}

// TestParseExplainWithOptionList verifies EXPLAIN with option list parsing.
// Reference: tests/sqlparser_common.rs:14072
func TestParseExplainWithOptionList(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsExplainWithUtilityOptions()
	})

	// Test various EXPLAIN options
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (ANALYZE false, VERBOSE true) SELECT sqrt(id) FROM foo", nil)
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (ANALYZE ON, VERBOSE OFF) SELECT sqrt(id) FROM foo", nil)
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (FORMAT1 TEXT, FORMAT2 'JSON', FORMAT3 \"XML\", FORMAT4 YAML) SELECT sqrt(id) FROM foo", nil)
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (NUM1 10, NUM2 +10.1, NUM3 -10.2) SELECT sqrt(id) FROM foo", nil)
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (ANALYZE, VERBOSE true, WAL OFF, FORMAT YAML, USER_DEF_NUM -100.1) SELECT sqrt(id) FROM foo", nil)
}

// TestParseAnalyze verifies ANALYZE statement parsing.
// Reference: tests/sqlparser_common.rs:683
func TestParseAnalyze(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "ANALYZE TABLE test_table")
	dialects.VerifiedStmt(t, "ANALYZE test_table")
}
