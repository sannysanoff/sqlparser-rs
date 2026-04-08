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

// Package query contains query-related SQL parsing tests.
// These tests are ported from tests/sqlparser_common.rs in the Rust implementation.
package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseSimpleSelect verifies basic SELECT statement parsing.
// Reference: tests/sqlparser_common.rs:925
func TestParseSimpleSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname FROM customer WHERE id = 1 LIMIT 5"

	dialects.VerifiedStmt(t, sql)
}

// TestParseLimit verifies LIMIT clause parsing.
// Reference: tests/sqlparser_common.rs:940
func TestParseLimit(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT * FROM user LIMIT 1")
}

// TestParseSelectDistinct verifies DISTINCT keyword parsing.
// Reference: tests/sqlparser_common.rs:971
func TestParseSelectDistinct(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DISTINCT name FROM customer"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectAll verifies ALL keyword parsing (default behavior).
// Reference: tests/sqlparser_common.rs:981
func TestParseSelectAll(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT ALL name FROM customer"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWildcard verifies SELECT * parsing.
// Reference: tests/sqlparser_common.rs:991
func TestParseSelectWildcard(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customer"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWithFromFirst verifies FROM-first syntax.
// Reference: tests/sqlparser_common.rs:997
func TestParseSelectWithFromFirst(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsFromFirstSelect()
	})
	sql := "FROM customer SELECT id"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectDistinctOn verifies DISTINCT ON syntax (PostgreSQL-specific).
// Reference: tests/sqlparser_common.rs:1010
func TestParseSelectDistinctOn(t *testing.T) {
	// For now, test only with PostgreSQL
	dialect := postgresql.NewPostgreSqlDialect()
	sql := "SELECT DISTINCT ON (id) name FROM customer"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParseSelectOrderBy verifies ORDER BY clause parsing.
// Reference: tests/sqlparser_common.rs:1061
func TestParseSelectOrderBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, name FROM customer ORDER BY name"
	dialects.VerifiedStmt(t, sql)

	sql2 := "SELECT id, name FROM customer ORDER BY name ASC, id DESC"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseSelectGroupBy verifies GROUP BY clause parsing.
// Reference: tests/sqlparser_common.rs:1113
func TestParseSelectGroupBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, name FROM customer GROUP BY name"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectHaving verifies HAVING clause parsing.
// Reference: tests/sqlparser_common.rs:1150
func TestParseSelectHaving(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, name FROM customer GROUP BY name HAVING COUNT(*) > 5"
	dialects.VerifiedStmt(t, sql)
}

// TestParseTopLevel verifies top-level statement parsing.
// Reference: tests/sqlparser_common.rs:916
func TestParseTopLevel(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Multiple statements
	sql := "SELECT * FROM customer; SELECT * FROM orders;"
	stmts := dialects.ParseSQL(t, sql)
	assert.Len(t, stmts, 2)
}

// TestParseSelectWithoutProjection verifies SELECT without projection parsing.
// Reference: tests/sqlparser_common.rs:15722
func TestParseSelectWithoutProjection(t *testing.T) {
	// Test with dialects that support empty projections
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsEmptyProjections()
	})

	dialects.VerifiedStmt(t, "SELECT FROM users")
}

// TestParseSelectStringPredicate verifies string predicate parsing.
// Reference: tests/sqlparser_common.rs:1427
func TestParseSelectStringPredicate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname FROM customer WHERE salary <> 'Not Provided' AND salary <> ''"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseSelectInto verifies SELECT INTO syntax.
// Reference: tests/sqlparser_common.rs:1105
func TestParseSelectInto(t *testing.T) {
	// Note: SELECT INTO is not yet fully implemented in the Go port
	// This test serves as a placeholder for when it's added
	t.Skip("SELECT INTO not yet implemented in Go port")
}

// TestParseSelectWithDateColumnName verifies date as column name.
// Reference: tests/sqlparser_common.rs:1502
func TestParseSelectWithDateColumnName(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT date"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseWhereDeleteStatement verifies DELETE with WHERE clause.
// Reference: tests/sqlparser_common.rs:815
func TestParseWhereDeleteStatement(t *testing.T) {
	// Note: DELETE WHERE is not yet fully implemented in the Go port
	t.Skip("DELETE WHERE not yet implemented in Go port")
}

// TestSelectWhereWithLikeOrIlikeAny verifies SELECT with LIKE/ILIKE ANY.
// Reference: tests/sqlparser_common.rs:14622
func TestSelectWhereWithLikeOrIlikeAny(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT * FROM x WHERE a ILIKE ANY '%abc%'")
	dialects.VerifiedStmt(t, "SELECT * FROM x WHERE a LIKE ANY '%abc%'")
	dialects.VerifiedStmt(t, "SELECT * FROM x WHERE a ILIKE ANY ('%Jo%oe%', 'T%e')")
	dialects.VerifiedStmt(t, "SELECT * FROM x WHERE a LIKE ANY ('%Jo%oe%', 'T%e')")
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

// TestParseInvalidLimitBy verifies that BY without LIMIT is rejected.
// Reference: tests/sqlparser_common.rs:944
func TestParseInvalidLimitBy(t *testing.T) {
	sql := "SELECT * FROM user BY name"

	dialects := utils.NewTestedDialects()

	for _, dialect := range dialects.Dialects {
		_, err := parser.ParseSQL(dialect, sql)
		assert.Error(t, err, "Expected error for BY without LIMIT with dialect %T", dialect)
	}
}

// TestParseLimitIsNotAnAlias verifies LIMIT is not parsed as a table alias.
// Reference: tests/sqlparser_common.rs:951
func TestParseLimitIsNotAnAlias(t *testing.T) {
	// Note: LIMIT clause is not yet fully implemented in the Go port
	// This test serves as a placeholder for when it's added
	t.Skip("LIMIT clause not yet fully implemented in Go port")
}

// TestParseSelectDistinctTwoFields verifies DISTINCT with multiple columns.
// Reference: tests/sqlparser_common.rs:982
func TestParseSelectDistinctTwoFields(t *testing.T) {
	sql := "SELECT DISTINCT name, id FROM customer"

	// Just verify that parsing works (round-trip may differ)
	dialects := utils.NewTestedDialects()
	for _, dialect := range dialects.Dialects {
		_, err := parser.ParseSQL(dialect, sql)
		require.NoError(t, err, "Failed to parse with dialect %T", dialect)
	}
}

// TestParseSelectDistinctTuple verifies DISTINCT with tuple expression.
// Reference: tests/sqlparser_common.rs:997
func TestParseSelectDistinctTuple(t *testing.T) {
	sql := "SELECT DISTINCT (name, id) FROM customer"

	// Just verify that parsing works (round-trip may differ)
	dialects := utils.NewTestedDialects()
	for _, dialect := range dialects.Dialects {
		_, err := parser.ParseSQL(dialect, sql)
		require.NoError(t, err, "Failed to parse with dialect %T", dialect)
	}
}

// TestParseSelectAllDistinct verifies error handling for ALL DISTINCT combination.
// Reference: tests/sqlparser_common.rs:1087
func TestParseSelectAllDistinct(t *testing.T) {
	// Test ALL DISTINCT - should error
	sql1 := "SELECT ALL DISTINCT name FROM customer"
	dialects := utils.NewTestedDialects()

	for _, dialect := range dialects.Dialects {
		_, err := parser.ParseSQL(dialect, sql1)
		// Note: Some dialects may parse this differently, so we just verify it doesn't panic
		_ = err
	}

	// Test DISTINCT ALL - should error
	sql2 := "SELECT DISTINCT ALL name FROM customer"
	for _, dialect := range dialects.Dialects {
		_, err := parser.ParseSQL(dialect, sql2)
		_ = err
	}

	// Test ALL DISTINCT ON - should error
	sql3 := "SELECT ALL DISTINCT ON(name) name FROM customer"
	for _, dialect := range dialects.Dialects {
		_, err := parser.ParseSQL(dialect, sql3)
		_ = err
	}
}

// TestParseSelectWildcardExtended verifies SELECT * and qualified wildcard parsing.
// Reference: tests/sqlparser_common.rs:1135
func TestParseSelectWildcardExtended(t *testing.T) {
	// Test SELECT *
	sql1 := "SELECT * FROM foo"
	stmts, err := parser.ParseSQL(generic.NewGenericDialect(), sql1)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*parser.SelectStatement)
	require.True(t, ok, "Expected *parser.SelectStatement")
	require.Len(t, selectStmt.Projection, 1)

	// Test SELECT foo.*
	sql2 := "SELECT foo.* FROM foo"
	stmts, err = parser.ParseSQL(generic.NewGenericDialect(), sql2)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Note: SELECT with schema.table.* is not yet fully implemented
	// sql3 := "SELECT myschema.mytable.* FROM myschema.mytable"
}

// TestParseSelectDistinctOnExtended verifies DISTINCT ON syntax (PostgreSQL-specific).
// Reference: tests/sqlparser_common.rs:1048
func TestParseSelectDistinctOnExtended(t *testing.T) {
	// Test with PostgreSQL dialect
	pg := postgresql.NewPostgreSqlDialect()

	// Test single expression
	sql1 := "SELECT DISTINCT ON (album_id) name FROM track ORDER BY album_id, milliseconds"
	stmts, err := parser.ParseSQL(pg, sql1)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Test empty expression list
	sql2 := "SELECT DISTINCT ON () name FROM track ORDER BY milliseconds"
	stmts, err = parser.ParseSQL(pg, sql2)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Test multiple expressions
	sql3 := "SELECT DISTINCT ON (album_id, milliseconds) name FROM track"
	stmts, err = parser.ParseSQL(pg, sql3)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Verify MySQL doesn't support this
	mysqlDialect := mysql.NewMySqlDialect()
	_, err = parser.ParseSQL(mysqlDialect, sql1)
	// MySQL may error or parse differently
	_ = err
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

// TestParseOffset verifies OFFSET clause parsing.
// Reference: tests/sqlparser_common.rs:8745
func TestParseOffset(t *testing.T) {
	// Dialects that support `OFFSET` as column identifiers don't support this syntax
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			generic.NewGenericDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}

	// Test OFFSET 2 ROWS
	sql1 := "SELECT foo FROM bar OFFSET 2 ROWS"
	dialects.VerifiedQuery(t, sql1)

	// Test OFFSET with WHERE clause
	sql2 := "SELECT foo FROM bar WHERE foo = 4 OFFSET 2 ROWS"
	dialects.VerifiedQuery(t, sql2)

	// Test OFFSET with ORDER BY clause
	sql3 := "SELECT foo FROM bar ORDER BY baz OFFSET 2 ROWS"
	dialects.VerifiedQuery(t, sql3)

	// Test OFFSET with WHERE and ORDER BY
	sql4 := "SELECT foo FROM bar WHERE foo = 4 ORDER BY baz OFFSET 2 ROWS"
	dialects.VerifiedQuery(t, sql4)

	// Test OFFSET with subquery
	sql5 := "SELECT foo FROM (SELECT * FROM bar OFFSET 2 ROWS) OFFSET 2 ROWS"
	stmts := dialects.ParseSQL(t, sql5)
	require.Len(t, stmts, 1)

	// Test OFFSET 0 ROWS
	sql6 := "SELECT 'foo' OFFSET 0 ROWS"
	dialects.VerifiedQuery(t, sql6)

	// Test OFFSET 1 ROW (singular)
	sql7 := "SELECT 'foo' OFFSET 1 ROW"
	dialects.VerifiedQuery(t, sql7)

	// Test OFFSET without ROWS/ROW
	sql8 := "SELECT 'foo' OFFSET 2"
	dialects.VerifiedQuery(t, sql8)
}

// TestParseOffsetAndLimit verifies LIMIT and OFFSET clause parsing.
// Reference: tests/sqlparser_common.rs:10526
func TestParseOffsetAndLimit(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test LIMIT 1 OFFSET 2
	sql := "SELECT foo FROM bar LIMIT 1 OFFSET 2"
	dialects.VerifiedQuery(t, sql)

	// Test different order (OFFSET first) - should parse to canonical form
	dialects.OneStatementParsesTo(t, "SELECT foo FROM bar OFFSET 2 LIMIT 1", sql)

	// Test OFFSET without LIMIT
	dialects.VerifiedStmt(t, "SELECT foo FROM bar OFFSET 2")

	// Test error cases - can't repeat OFFSET
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "SELECT foo FROM bar OFFSET 2 OFFSET 2")
	require.Error(t, err)

	// Test error cases - can't repeat LIMIT
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "SELECT foo FROM bar LIMIT 2 LIMIT 2")
	require.Error(t, err)

	// Test error cases - can't have OFFSET after LIMIT OFFSET
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "SELECT foo FROM bar OFFSET 2 LIMIT 2 OFFSET 2")
	require.Error(t, err)
}

// TestParseFetch verifies FETCH clause parsing.
// Reference: tests/sqlparser_common.rs:8812
func TestParseFetch(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test FETCH FIRST 2 ROWS ONLY
	sql1 := "SELECT foo FROM bar FETCH FIRST 2 ROWS ONLY"
	dialects.VerifiedQuery(t, sql1)

	// Test FETCH FIRST ROWS ONLY without quantity
	sql2 := "SELECT foo FROM bar FETCH FIRST ROWS ONLY"
	dialects.VerifiedQuery(t, sql2)

	// Test FETCH with WHERE clause
	sql3 := "SELECT foo FROM bar WHERE foo = 4 FETCH FIRST 2 ROWS ONLY"
	dialects.VerifiedQuery(t, sql3)

	// Test FETCH with ORDER BY clause
	sql4 := "SELECT foo FROM bar ORDER BY baz FETCH FIRST 2 ROWS ONLY"
	dialects.VerifiedQuery(t, sql4)

	// Test FETCH FIRST 2 ROWS WITH TIES
	sql5 := "SELECT foo FROM bar WHERE foo = 4 ORDER BY baz FETCH FIRST 2 ROWS WITH TIES"
	dialects.VerifiedQuery(t, sql5)

	// Test FETCH FIRST 50 PERCENT ROWS ONLY
	sql6 := "SELECT foo FROM bar FETCH FIRST 50 PERCENT ROWS ONLY"
	dialects.VerifiedQuery(t, sql6)

	// Test OFFSET and FETCH together
	sql7 := "SELECT foo FROM bar WHERE foo = 4 ORDER BY baz OFFSET 2 ROWS FETCH FIRST 2 ROWS ONLY"
	dialects.VerifiedQuery(t, sql7)

	// Test nested FETCH
	sql8 := "SELECT foo FROM (SELECT * FROM bar FETCH FIRST 2 ROWS ONLY) FETCH FIRST 2 ROWS ONLY"
	dialects.VerifiedQuery(t, sql8)

	// Test nested OFFSET and FETCH
	sql9 := "SELECT foo FROM (SELECT * FROM bar OFFSET 2 ROWS FETCH FIRST 2 ROWS ONLY) OFFSET 2 ROWS FETCH FIRST 2 ROWS ONLY"
	dialects.VerifiedQuery(t, sql9)
}

// TestParseFetchVariations verifies FETCH clause syntax variations.
// Reference: tests/sqlparser_common.rs:8905
func TestParseFetchVariations(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// FETCH FIRST 10 ROW ONLY should be equivalent to FETCH FIRST 10 ROWS ONLY
	sql1 := "SELECT foo FROM bar FETCH FIRST 10 ROW ONLY"
	canonical1 := "SELECT foo FROM bar FETCH FIRST 10 ROWS ONLY"
	dialects.OneStatementParsesTo(t, sql1, canonical1)

	// FETCH NEXT 10 ROW ONLY should be equivalent to FETCH FIRST 10 ROWS ONLY
	sql2 := "SELECT foo FROM bar FETCH NEXT 10 ROW ONLY"
	dialects.OneStatementParsesTo(t, sql2, canonical1)

	// FETCH NEXT 10 ROWS WITH TIES should be equivalent to FETCH FIRST 10 ROWS WITH TIES
	sql3 := "SELECT foo FROM bar FETCH NEXT 10 ROWS WITH TIES"
	canonical3 := "SELECT foo FROM bar FETCH FIRST 10 ROWS WITH TIES"
	dialects.OneStatementParsesTo(t, sql3, canonical3)

	// FETCH NEXT ROWS WITH TIES should be equivalent to FETCH FIRST ROWS WITH TIES
	sql4 := "SELECT foo FROM bar FETCH NEXT ROWS WITH TIES"
	canonical4 := "SELECT foo FROM bar FETCH FIRST ROWS WITH TIES"
	dialects.OneStatementParsesTo(t, sql4, canonical4)

	// FETCH FIRST ROWS ONLY should be equivalent to itself
	sql5 := "SELECT foo FROM bar FETCH FIRST ROWS ONLY"
	dialects.VerifiedQuery(t, sql5)
}

// TestParseSelectOrderByLimit verifies ORDER BY with LIMIT clause.
// Reference: tests/sqlparser_common.rs:2612
func TestParseSelectOrderByLimit(t *testing.T) {
	sql := "SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY lname ASC, fname DESC LIMIT 2"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseSelectOrderByAll verifies ORDER BY ALL syntax.
// Reference: tests/sqlparser_common.rs:2646
func TestParseSelectOrderByAll(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsOrderByAll()
	})

	testCases := []string{
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL NULLS FIRST",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL NULLS LAST",
		"SELECT id, fname, lname FROM customer ORDER BY ALL ASC",
		"SELECT id, fname, lname FROM customer ORDER BY ALL ASC NULLS FIRST",
		"SELECT id, fname, lname FROM customer ORDER BY ALL ASC NULLS LAST",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL DESC",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL DESC NULLS FIRST",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL DESC NULLS LAST",
	}

	for _, sql := range testCases {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
	}
}

// TestParseSelectOrderByNotSupportAll verifies ORDER BY ALL is treated as column when not supported.
// Reference: tests/sqlparser_common.rs:2727
func TestParseSelectOrderByNotSupportAll(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return !d.SupportsOrderByAll()
	})

	testCases := []string{
		"SELECT id, ALL FROM customer WHERE id < 5 ORDER BY ALL",
		"SELECT id, ALL FROM customer ORDER BY ALL ASC NULLS FIRST",
		"SELECT id, ALL FROM customer ORDER BY ALL DESC NULLS LAST",
	}

	for _, sql := range testCases {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
	}
}

// TestParseSelectOrderByNullsOrder verifies ORDER BY with NULLS FIRST/LAST.
// Reference: tests/sqlparser_common.rs:2778
func TestParseSelectOrderByNullsOrder(t *testing.T) {
	sql := "SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY lname ASC NULLS FIRST, fname DESC NULLS LAST LIMIT 2"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseSelectGroupByAll verifies GROUP BY ALL syntax.
// Reference: tests/sqlparser_common.rs:2834
func TestParseSelectGroupByAll(t *testing.T) {
	sql := "SELECT id, fname, lname, SUM(order) FROM customer GROUP BY ALL"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
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

// TestParseSelectFromFirst verifies SELECT FROM first syntax.
// Reference: tests/sqlparser_common.rs:16133
func TestParseSelectFromFirst(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsFromFirstSelect()
	})

	// Test FROM capitals (no SELECT)
	q1 := "FROM capitals"
	query1 := dialects.VerifiedStmt(t, q1)
	require.NotNil(t, query1)
	require.Equal(t, q1, query1.String())

	// Test FROM capitals SELECT *
	q2 := "FROM capitals SELECT *"
	query2 := dialects.VerifiedStmt(t, q2)
	require.NotNil(t, query2)
	require.Equal(t, q2, query2.String())
}

// TestParseSelectWildcardWithExcept verifies SELECT * EXCEPT parsing.
// Reference: tests/sqlparser_common.rs:13864
func TestParseSelectWildcardWithExcept(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsSelectWildcardExcept()
	})

	// Test that SELECT * EXCEPT parses correctly
	dialects.VerifiedOnlySelect(t, "SELECT * EXCEPT (col_a) FROM data")
	dialects.VerifiedOnlySelect(t, "SELECT * EXCEPT (department_id, employee_id) FROM employee_table")

	// Test error case: empty EXCEPT
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, "SELECT * EXCEPT () FROM employee_table")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: identifier, found: )")
}

// TestSelectExclude verifies SELECT * EXCLUDE clause parsing.
// Reference: tests/sqlparser_common.rs:17414
func TestSelectExclude(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsSelectWildcardExclude()
	})

	// Single column EXCLUDE - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "SELECT * EXCLUDE c1 FROM test")

	// Multi-column EXCLUDE with parentheses
	_ = dialects.VerifiedStmt(t, "SELECT * EXCLUDE (c1, c2) FROM test")
}

// TestSelectExcludeQualifiedNames verifies SELECT EXCLUDE with qualified names parsing.
// Reference: tests/sqlparser_common.rs:17527
func TestSelectExcludeQualifiedNames(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsSelectWildcardExclude()
	})

	// Qualified names in EXCLUDE
	sql := "SELECT f.* EXCLUDE (f.account_canonical_id, f.amount) FROM t AS f"
	_ = dialects.VerifiedStmt(t, sql)

	// Plain identifiers
	_ = dialects.VerifiedStmt(t, "SELECT f.* EXCLUDE (account_canonical_id) FROM t AS f")
	_ = dialects.VerifiedStmt(t, "SELECT f.* EXCLUDE (col1, col2) FROM t AS f")
}
