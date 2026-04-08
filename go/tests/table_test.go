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
// This file contains table-related tests.
package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseCacheTable verifies CACHE TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:10861
func TestParseCacheTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT a, b, c FROM foo"
	dialects.VerifiedQuery(t, sql)

	// Test CACHE TABLE 'cache_table_name'
	stmt := dialects.VerifiedStmt(t, "CACHE TABLE 'cache_table_name'")
	cache, ok := stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.NotNil(t, cache.TableName, "Expected table name to be present")
	assert.False(t, cache.HasAs, "Expected HasAs to be false")
	assert.Equal(t, 0, len(cache.Options), "Expected no options")
	assert.Nil(t, cache.Query, "Expected no query")

	// Test CACHE flag TABLE 'cache_table_name'
	stmt = dialects.VerifiedStmt(t, "CACHE flag TABLE 'cache_table_name'")
	cache, ok = stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.NotNil(t, cache.TableFlag, "Expected table flag to be present")
	require.NotNil(t, cache.TableName, "Expected table name to be present")

	// Test with OPTIONS
	stmt = dialects.VerifiedStmt(t, "CACHE flag TABLE 'cache_table_name' OPTIONS('K1' = 'V1', 'K2' = 0.88)")
	cache, ok = stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.Equal(t, 2, len(cache.Options), "Expected 2 options")

	// Test with query (no AS)
	stmt = dialects.VerifiedStmt(t, "CACHE flag TABLE 'cache_table_name' OPTIONS('K1' = 'V1', 'K2' = 0.88) SELECT a, b, c FROM foo")
	cache, ok = stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.NotNil(t, cache.Query, "Expected query to be present")
	assert.False(t, cache.HasAs, "Expected HasAs to be false without AS")

	// Test with query and AS
	stmt = dialects.VerifiedStmt(t, "CACHE flag TABLE 'cache_table_name' OPTIONS('K1' = 'V1', 'K2' = 0.88) AS SELECT a, b, c FROM foo")
	cache, ok = stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.NotNil(t, cache.Query, "Expected query to be present")
	assert.True(t, cache.HasAs, "Expected HasAs to be true with AS")

	// Error cases
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "CACHE TABLE 'table_name' foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE flag TABLE 'table_name' OPTIONS('K1'='V1') foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE TABLE 'table_name' AS foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE flag TABLE 'table_name' OPTIONS('K1'='V1') AS foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE 'table_name'")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE 'table_name' OPTIONS('K1'='V1')")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE flag 'table_name' OPTIONS('K1'='V1')")
	require.Error(t, err)
}

// TestParseUncacheTable verifies UNCACHE TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:11038
func TestParseUncacheTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test UNCACHE TABLE 'table_name'
	stmt := dialects.VerifiedStmt(t, "UNCACHE TABLE 'table_name'")
	uncache, ok := stmt.(*statement.Uncache)
	require.True(t, ok, "Expected Uncache statement, got %T", stmt)
	require.NotNil(t, uncache.TableName, "Expected table name to be present")
	assert.False(t, uncache.IfExists, "Expected IfExists to be false")

	// Test UNCACHE TABLE IF EXISTS 'table_name'
	stmt = dialects.VerifiedStmt(t, "UNCACHE TABLE IF EXISTS 'table_name'")
	uncache, ok = stmt.(*statement.Uncache)
	require.True(t, ok, "Expected Uncache statement, got %T", stmt)
	require.NotNil(t, uncache.TableName, "Expected table name to be present")
	assert.True(t, uncache.IfExists, "Expected IfExists to be true")

	// Error cases
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "UNCACHE TABLE 'table_name' foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "UNCACHE 'table_name' foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "UNCACHE IF EXISTS 'table_name' foo")
	require.Error(t, err)
}

// TestParsePivotTable verifies PIVOT table expression parsing.
// Reference: tests/sqlparser_common.rs:11256
func TestParsePivotTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM monthly_sales PIVOT(SUM(amount) FOR month IN ('JAN', 'FEB'))"
	dialects.VerifiedStmt(t, sql)
}

// TestParseUnpivotTable verifies UNPIVOT table expression parsing.
// Reference: tests/sqlparser_common.rs:11434
func TestParseUnpivotTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM quarterly_sales UNPIVOT (amount FOR quarter IN (q1, q2, q3, q4))"
	dialects.VerifiedStmt(t, sql)
}

// TestParsePivotUnpivotTable verifies combined PIVOT after UNPIVOT.
// Reference: tests/sqlparser_common.rs:11713
func TestParsePivotUnpivotTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM (SELECT * FROM sales UNPIVOT (amount FOR month IN (jan, feb))) PIVOT(SUM(amount) FOR month IN ('JAN', 'FEB'))"
	dialects.VerifiedStmt(t, sql)
}

// TestParseTableSample verifies TABLESAMPLE clause parsing.
// Reference: tests/sqlparser_common.rs:15670
func TestParseTableSample(t *testing.T) {
	// Test with dialects that support TABLESAMPLE before alias
	dialectsBeforeAlias := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTableSampleBeforeAlias()
	})

	testCasesBefore := []string{
		"SELECT * FROM tbl TABLESAMPLE (50) AS t",
		"SELECT * FROM tbl TABLESAMPLE (50 ROWS) AS t",
		"SELECT * FROM tbl TABLESAMPLE (50 PERCENT) AS t",
	}
	for _, sql := range testCasesBefore {
		dialectsBeforeAlias.VerifiedStmt(t, sql)
	}

	// Test with dialects that support TABLESAMPLE after alias
	dialectsAfterAlias := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsTableSampleBeforeAlias()
	})

	testCasesAfter := []string{
		"SELECT * FROM tbl AS t TABLESAMPLE BERNOULLI (50)",
		"SELECT * FROM tbl AS t TABLESAMPLE SYSTEM (50)",
		"SELECT * FROM tbl AS t TABLESAMPLE SYSTEM (50) REPEATABLE (10)",
	}
	for _, sql := range testCasesAfter {
		dialectsAfterAlias.VerifiedStmt(t, sql)
	}
}

// TestTableSample verifies TABLESAMPLE clause parsing.
// Reference: tests/sqlparser_common.rs:15670
func TestTableSample(t *testing.T) {
	// Test with dialects that support TABLESAMPLE before alias
	dialectsBeforeAlias := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTableSampleBeforeAlias()
	})

	testCasesBefore := []string{
		"SELECT * FROM tbl TABLESAMPLE (50) AS t",
		"SELECT * FROM tbl TABLESAMPLE (50 ROWS) AS t",
		"SELECT * FROM tbl TABLESAMPLE (50 PERCENT) AS t",
	}
	for _, sql := range testCasesBefore {
		dialectsBeforeAlias.VerifiedStmt(t, sql)
	}

	// Test with dialects that support TABLESAMPLE after alias
	dialectsAfterAlias := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsTableSampleBeforeAlias()
	})

	testCasesAfter := []string{
		"SELECT * FROM tbl AS t TABLESAMPLE BERNOULLI (50)",
		"SELECT * FROM tbl AS t TABLESAMPLE SYSTEM (50)",
		"SELECT * FROM tbl AS t TABLESAMPLE SYSTEM (50) REPEATABLE (10)",
	}
	for _, sql := range testCasesAfter {
		dialectsAfterAlias.VerifiedStmt(t, sql)
	}
}

// TestParseUnnest verifies UNNEST function parsing.
// Reference: tests/sqlparser_common.rs:7000
func TestParseUnnest(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT UNNEST(make_array(1, 2, 3))"
	dialects.VerifiedStmt(t, sql)

	sql2 := "SELECT UNNEST(make_array(1, 2, 3), make_array(4, 5))"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseInUnnest verifies IN UNNEST expression parsing.
// Reference: tests/sqlparser_common.rs:2332
func TestParseInUnnest(t *testing.T) {
	chk := func(negated bool) {
		sql := fmt.Sprintf("SELECT * FROM customers WHERE segment %sIN UNNEST(expr)",
			func() string {
				if negated {
					return "NOT "
				}
				return ""
			}())
		dialects := utils.NewTestedDialects()
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
	}
	chk(false)
	chk(true)
}

// TestLateralDerived verifies LATERAL derived table parsing.
// Reference: tests/sqlparser_common.rs:8929
func TestLateralDerived(t *testing.T) {
	dialects := utils.NewTestedDialects()

	chk := func(t *testing.T, lateralIn bool) {
		lateralStr := ""
		if lateralIn {
			lateralStr = "LATERAL "
		}
		sql := "SELECT * FROM customer LEFT JOIN " + lateralStr +
			"(SELECT * FROM orders WHERE orders.customer = customer.id LIMIT 3) AS orders ON 1"
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}

	// Test without LATERAL
	chk(t, false)

	// Test with LATERAL
	chk(t, true)

	// Test invalid: LATERAL UNNEST with WITH OFFSET should fail
	sql3 := "SELECT * FROM LATERAL UNNEST ([10,20,30]) as numbers WITH OFFSET"
	_, err := parser.ParseSQL(dialects.Dialects[0], sql3)
	require.Error(t, err, "Expected error for LATERAL UNNEST with WITH OFFSET")

	// Test invalid: LEFT JOIN LATERAL with cross join should fail
	sql4 := "SELECT * FROM a LEFT JOIN LATERAL (b CROSS JOIN c)"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql4)
	require.Error(t, err, "Expected error for invalid LATERAL join")
}

// TestLateralFunction verifies LATERAL function call parsing.
// Reference: tests/sqlparser_common.rs:8985
func TestLateralFunction(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM customer LEFT JOIN LATERAL generate_series(1, customer.id)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify the statement parses correctly
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseUnnestInFromClause verifies UNNEST in FROM clause with various options.
// Reference: tests/sqlparser_common.rs:7008
func TestParseUnnestInFromClause(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			bigquery.NewBigQueryDialect(),
			generic.NewGenericDialect(),
		},
	}

	// Test various UNNEST configurations
	testCases := []string{
		"SELECT * FROM UNNEST(expr) AS numbers WITH OFFSET",
		"SELECT * FROM UNNEST(expr)",
		"SELECT * FROM UNNEST(expr) WITH OFFSET",
		"SELECT * FROM UNNEST(make_array(1, 2, 3))",
		"SELECT * FROM UNNEST(make_array(1, 2, 3), make_array(5, 6))",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedOnlySelect(t, sql)
		})
	}
}

// TestParseSemanticViewTableFactor verifies SEMANTIC_VIEW table factor parsing.
// Reference: tests/sqlparser_common.rs:17992
// Note: SEMANTIC_VIEW is a specialized feature, basic parsing is tested here.
func TestParseSemanticViewTableFactor(t *testing.T) {
	// Test basic SEMANTIC_VIEW parsing if supported
	sql := "SELECT * FROM SEMANTIC_VIEW(model)"

	// Parse and verify - may need dialect filtering if dialect-specific
	stmts := utils.NewTestedDialects().ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseTableFunction verifies TABLE() function in FROM clause.
// Reference: tests/sqlparser_common.rs:6933
func TestParseTableFunction(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic TABLE function with alias
	dialects.VerifiedOnlySelect(t, "SELECT * FROM TABLE(FUN('1')) AS a")
}

// TestParseFromAdvanced verifies complex FROM clause with table-valued function and hints.
// Reference: tests/sqlparser_common.rs:7253
func TestParseFromAdvanced(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM fn(1, 2) AS foo, schema.bar AS bar WITH (NOLOCK)"
	dialects.VerifiedStmt(t, sql)
}

// TestTrailingCommasInFrom verifies trailing commas in FROM clause parsing.
// Reference: tests/sqlparser_common.rs:15772
func TestTrailingCommasInFrom(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsFromTrailingCommas()
	})

	dialects.VerifiedOnlySelectWithCanonical(t, "SELECT 1, 2 FROM t,", "SELECT 1, 2 FROM t")
	dialects.VerifiedOnlySelectWithCanonical(t, "SELECT 1, 2 FROM t1, t2,", "SELECT 1, 2 FROM t1, t2")

	sql := "SELECT a, FROM b, LIMIT 1"
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	sql = "INSERT INTO a SELECT b FROM c,"
	_, err = utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	sql = "SELECT a FROM b, HAVING COUNT(*) > 1"
	_, err = utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	sql = "SELECT a FROM b, WHERE c = 1"
	_, err = utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	// nested
	sql = "SELECT 1, 2 FROM (SELECT * FROM t,),"
	_, err = utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	// multiple_subqueries
	dialects.VerifiedOnlySelectWithCanonical(t, "SELECT 1, 2 FROM (SELECT * FROM t1), (SELECT * FROM t2),", "SELECT 1, 2 FROM (SELECT * FROM t1), (SELECT * FROM t2)")
}

// TestParseSelectTableWithIndexHints verifies MySQL index hints parsing.
// Reference: tests/sqlparser_common.rs:11649
func TestParseSelectTableWithIndexHints(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsIndexHints()
	})

	sql := "SELECT * FROM t USE INDEX (idx1)"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWithTableAlias verifies SELECT with table alias parsing.
// Reference: tests/sqlparser_common.rs:631
func TestParseSelectWithTableAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id FROM customer AS c"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWithTableAliasWithoutAs verifies table alias without AS keyword.
// Reference: tests/sqlparser_common.rs:644
func TestParseSelectWithTableAliasWithoutAs(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id FROM customer c"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWithTableAliasAs verifies SELECT with table alias and column list.
// Reference: tests/sqlparser_common.rs:630
func TestParseSelectWithTableAliasAs(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// With AS keyword
	dialects.VerifiedStmt(t, "SELECT a, b, c FROM lineitem AS l (A, B, C)")

	// Without AS keyword (AS is optional)
	dialects.VerifiedStmt(t, "SELECT a, b, c FROM lineitem l (A, B, C)")
}

// TestSelectTop verifies SELECT TOP clause parsing.
// Reference: tests/sqlparser_common.rs:15140
func TestSelectTop(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTopBeforeDistinct()
	})

	dialects.VerifiedStmt(t, "SELECT ALL * FROM tbl")
	dialects.VerifiedStmt(t, "SELECT TOP 3 * FROM tbl")
	dialects.VerifiedStmt(t, "SELECT TOP 3 ALL * FROM tbl")
	dialects.VerifiedStmt(t, "SELECT TOP 3 DISTINCT * FROM tbl")
	dialects.VerifiedStmt(t, "SELECT TOP 3 DISTINCT a, b, c FROM tbl")
}
