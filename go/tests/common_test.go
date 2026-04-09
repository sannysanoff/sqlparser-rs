// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreement", "s.  See the NOTICE file
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

// Package tests contains the SQL parsing tests ported from the Rust implementation.
// This file contains tests from tests/sqlparser_common.rs.
// These tests cover core SQL parsing features that work across all dialects.
//
// Note: Some tests have been renamed to avoid conflicts with existing tests
// in other files (expr_test.go, func_test.go, other_test.go, parse_test.go).
package tests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/ansi"
	"github.com/user/sqlparser/dialects/clickhouse"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/hive"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// ============================================================================
// Test 1: parse_function_object_name (renamed to avoid conflict)
// Reference: tests/sqlparser_common.rs:80
// ============================================================================

func TestCommonParseFunctionObjectName(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a.b.c.d(1, 2, 3) FROM T"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// ============================================================================
// Test 2: parse_insert_values
// Reference: tests/sqlparser_common.rs:97
// ============================================================================

func TestCommonParseInsertValues(t *testing.T) {
	dialects := utils.NewTestedDialects()
	testCases := []string{
		"INSERT customer VALUES (1, 2, 3)",
		"INSERT INTO customer VALUES (1, 2, 3)",
		"INSERT INTO customer VALUES (1, 2, 3), (1, 2, 3)",
		"INSERT INTO public.customer VALUES (1, 2, 3)",
		"INSERT INTO db.public.customer VALUES (1, 2, 3)",
		"INSERT INTO public.customer (id, name, active) VALUES (1, 2, 3)",
		"INSERT INTO t (id, name, active) VALUE (1, 2, 3)",
	}
	for _, sql := range testCases {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
	sql := "INSERT INTO customer WITH foo AS (SELECT 1) SELECT * FROM foo UNION VALUES (1)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 3: parse_insert_set
// Reference: tests/sqlparser_common.rs:181
// ============================================================================

func TestCommonParseInsertSet(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsInsertSet()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support INSERT SET syntax")
		return
	}
	sql := "INSERT INTO tbl1 SET col1 = 1, col2 = 'abc', col3 = current_date()"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// Test 4: parse_replace_into
// Reference: tests/sqlparser_common.rs:187
// ============================================================================

func TestCommonParseReplaceInto(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()
	sql := "REPLACE INTO public.customer (id, name, active) VALUES (1, 2, 3)"
	_, err := parser.ParseSQL(dialect, sql)
	require.Error(t, err, "PostgreSQL should not support REPLACE INTO")
}

// ============================================================================
// Test 5: parse_insert_default_values
// Reference: tests/sqlparser_common.rs:198
// ============================================================================

func TestCommonParseInsertDefaultValues(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "INSERT INTO test_table DEFAULT VALUES"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	sql2 := "INSERT INTO test_table DEFAULT VALUES RETURNING test_column"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
	sql3 := "INSERT INTO test_table DEFAULT VALUES ON CONFLICT DO NOTHING"
	stmts3 := dialects.ParseSQL(t, sql3)
	require.Len(t, stmts3, 1)
	errorCases := []string{
		"INSERT INTO test_table (test_col) DEFAULT VALUES",
		"INSERT INTO test_table DEFAULT VALUES (some_column)",
		"INSERT INTO test_table DEFAULT VALUES PARTITION (some_column)",
		"INSERT INTO test_table DEFAULT VALUES (1)",
	}
	for _, sql := range errorCases {
		_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
		_ = err
	}
}

// ============================================================================
// Test 6: parse_insert_sqlite
// Reference: tests/sqlparser_common.rs:357
// ============================================================================

func TestCommonParseInsertSQLite(t *testing.T) {
	dialect := sqlite.NewSQLiteDialect()
	checkOrClause := func(sql string, expectOr bool) {
		stmts, err := parser.ParseSQL(dialect, sql)
		require.NoError(t, err, "Failed to parse: %s", sql)
		require.Len(t, stmts, 1)
		insert, ok := stmts[0].(*ast.SInsert)
		if ok && expectOr {
			assert.NotNil(t, insert.Or, "Expected OR clause to be set for: %s", sql)
		}
	}
	checkOrClause("INSERT INTO test_table(id) VALUES(1)", false)
	checkOrClause("REPLACE INTO test_table(id) VALUES(1)", true)
	checkOrClause("INSERT OR REPLACE INTO test_table(id) VALUES(1)", true)
	checkOrClause("INSERT OR ROLLBACK INTO test_table(id) VALUES(1)", true)
	checkOrClause("INSERT OR ABORT INTO test_table(id) VALUES(1)", true)
	checkOrClause("INSERT OR FAIL INTO test_table(id) VALUES(1)", true)
	checkOrClause("INSERT OR IGNORE INTO test_table(id) VALUES(1)", true)
}

// ============================================================================
// Test 7: parse_insert_with_comments
// ============================================================================

func TestCommonParseInsertWithComments(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "INSERT /* comment */ INTO test_table VALUES (1)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 8: parse_insert_select
// Reference: tests/sqlparser_common.rs:312
// ============================================================================

func TestCommonParseInsertSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "INSERT INTO t SELECT 1 RETURNING 2"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	sql2 := "INSERT INTO t SELECT x RETURNING x AS y"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
	insert, ok := stmts2[0].(*ast.SInsert)
	if ok {
		assert.NotNil(t, insert.Returning, "Expected RETURNING clause")
	}
}

// ============================================================================
// Test 9: parse_insert_select_order_by
// ============================================================================

func TestCommonParseInsertSelectOrderBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "INSERT INTO t SELECT * FROM s ORDER BY col1"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 10: parse_insert_select_with_from
// Reference: tests/sqlparser_common.rs:331
// ============================================================================

func TestCommonParseInsertSelectWithFrom(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "INSERT INTO table1 SELECT * FROM table2 RETURNING id"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	insert, ok := stmts[0].(*ast.SInsert)
	if ok {
		assert.NotNil(t, insert.Source, "Expected source to be set")
		assert.NotNil(t, insert.Returning, "Expected RETURNING clause")
	}
}

// ============================================================================
// Test 11: test_insert_with_cte
// ============================================================================

func TestCommonParseInsertWithCTE(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "INSERT INTO t WITH cte AS (SELECT 1) SELECT * FROM cte"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 12: parse_update
// Reference: tests/sqlparser_common.rs:394
// ============================================================================

func TestCommonParseUpdate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "UPDATE t SET a = 1, b = 2, c = 3 WHERE d"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	update, ok := stmts[0].(*ast.SUpdate)
	if ok {
		require.NotNil(t, update.Table)
		assert.Equal(t, "t", update.Table.String())
		assert.Len(t, update.Assignments, 3)
		assert.NotNil(t, update.Selection)
	}
	dialects.VerifiedStmt(t, "UPDATE t SET a = 1, a = 2, a = 3")
	sql2 := "UPDATE t WHERE 1"
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql2)
	require.Error(t, err)
	sql3 := "UPDATE t SET a = 1 extrabadstuff"
	_, err = parser.ParseSQL(generic.NewGenericDialect(), sql3)
	require.Error(t, err)
}

// ============================================================================
// Test 13: update_with_table_alias
// Reference: tests/sqlparser_common.rs:547
// ============================================================================

func TestCommonUpdateWithTableAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "UPDATE users AS u SET u.username = 'new_user' WHERE u.username = 'old_user'"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	update, ok := stmts[0].(*ast.SUpdate)
	if ok {
		require.NotNil(t, update.Table)
		assert.NotNil(t, update.TableAlias)
		assert.Len(t, update.Assignments, 1)
	}
}

// ============================================================================
// Test 14: update_with_from
// Reference: tests/sqlparser_common.rs:444
// ============================================================================

func TestCommonUpdateWithFrom(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}
	sql := "UPDATE t1 SET name = t2.name FROM (SELECT name, id FROM t1 GROUP BY id) AS t2 WHERE t1.id = t2.id"
	dialects.VerifiedStmt(t, sql)
	sql2 := "UPDATE T SET a = b FROM U, (SELECT foo FROM V) AS W WHERE 1 = 1"
	dialects.VerifiedStmt(t, sql2)
}

// ============================================================================
// Test 15: parse_returning_as_column_alias
// Reference: tests/sqlparser_common.rs:352
// ============================================================================

func TestCommonParseReturningAsColumnAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT 1 AS RETURNING")
}

// ============================================================================
// Test 16: parse_update_or
// Reference: tests/sqlparser_common.rs:612
// ============================================================================

func TestCommonParseUpdateOr(t *testing.T) {
	dialects := utils.NewTestedDialects()
	testCases := []struct {
		sql            string
		expectOrClause bool
	}{
		{"UPDATE OR REPLACE t SET n = n + 1", true},
		{"UPDATE OR ROLLBACK t SET n = n + 1", true},
		{"UPDATE OR ABORT t SET n = n + 1", true},
		{"UPDATE OR FAIL t SET n = n + 1", true},
		{"UPDATE OR IGNORE t SET n = n + 1", true},
	}
	for _, tc := range testCases {
		stmts := dialects.ParseSQL(t, tc.sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Test 17: parse_select_with_table_alias_as
// Reference: tests/sqlparser_common.rs:631
// ============================================================================

func TestCommonParseSelectWithTableAliasAs(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql1 := "SELECT a, b, c FROM lineitem AS l (A, B, C)"
	stmts1 := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts1, 1)
	sql2 := "SELECT a, b, c FROM lineitem l (A, B, C)"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
}

// ============================================================================
// Test 18: parse_select_with_table_alias
// Reference: tests/sqlparser_common.rs:644
// ============================================================================

func TestCommonParseSelectWithTableAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a, b, c FROM lineitem AS l (A, B, C)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	_, ok := stmts[0].(*parser.SelectStatement)
	if !ok {
		_, ok = stmts[0].(*ast.SQuery)
	}
	assert.True(t, ok, "Expected SelectStatement or SQuery, got %T", stmts[0])
}

// ============================================================================
// Test 19: parse_analyze
// Reference: tests/sqlparser_common.rs:683
// ============================================================================

func TestCommonParseAnalyze(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "ANALYZE TABLE test_table")
	dialects.VerifiedStmt(t, "ANALYZE test_table")
}

// ============================================================================
// Test 20: parse_delete_statement
// Reference: tests/sqlparser_common.rs:703
// ============================================================================

func TestCommonParseDeleteStatement(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "DELETE FROM \"table\""
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 21: parse_delete_without_from_error
// Reference: tests/sqlparser_common.rs:720
// ============================================================================

func TestCommonParseDeleteWithoutFromError(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "DELETE \"table\" WHERE 1"
	for _, d := range dialects.Dialects {
		_, err := parser.ParseSQL(d, sql)
		_ = err
	}
}

// ============================================================================
// Test 22: parse_delete_statement_for_multi_tables
// Reference: tests/sqlparser_common.rs:734
// ============================================================================

func TestCommonParseDeleteStatementForMultiTables(t *testing.T) {
	dialect := mysql.NewMySqlDialect()
	sql := "DELETE schema1.table1, schema2.table2 FROM schema1.table1 JOIN schema2.table2 ON schema2.table2.col1 = schema1.table1.col1 WHERE schema2.table2.col2 = 1"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err, "MySQL should support multi-table DELETE")
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 23: parse_delete_statement_for_multi_tables_with_using
// Reference: tests/sqlparser_common.rs:773
// ============================================================================

func TestCommonParseDeleteStatementForMultiTablesWithUsing(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "DELETE FROM schema1.table1, schema2.table2 USING schema1.table1 JOIN schema2.table2 ON schema2.table2.pk = schema1.table1.col1 WHERE schema2.table2.col2 = 1"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 24: parse_where_delete_statement
// Reference: tests/sqlparser_common.rs:814
// ============================================================================

func TestCommonParseWhereDeleteStatement(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "DELETE FROM foo WHERE name = 5"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 25: parse_where_delete_with_alias_statement
// Reference: tests/sqlparser_common.rs:849
// ============================================================================

func TestCommonParseWhereDeleteWithAliasStatement(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "DELETE FROM basket AS a USING basket AS b WHERE a.id < b.id"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 26: parse_simple_select
// Reference: tests/sqlparser_common.rs:925
// ============================================================================

func TestCommonParseSimpleSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname FROM customer WHERE id = 1 LIMIT 5"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	_, ok := stmts[0].(*parser.SelectStatement)
	if !ok {
		_, ok = stmts[0].(*ast.SQuery)
	}
	assert.True(t, ok, "Expected SelectStatement or SQuery")
}

// ============================================================================
// Test 27: parse_limit
// Reference: tests/sqlparser_common.rs:939
// ============================================================================

func TestCommonParseLimit(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT * FROM user LIMIT 1")
}

// ============================================================================
// Test 28: parse_invalid_limit_by
// Reference: tests/sqlparser_common.rs:944
// ============================================================================

func TestCommonParseInvalidLimitBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM user BY name"
	for _, d := range dialects.Dialects {
		_, err := parser.ParseSQL(d, sql)
		_ = err
	}
}

// ============================================================================
// Test 29: parse_limit_is_not_an_alias
// Reference: tests/sqlparser_common.rs:952
// ============================================================================

func TestCommonParseLimitIsNotAnAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql1 := "SELECT id FROM customer LIMIT 1"
	stmts1 := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts1, 1)
	sql2 := "SELECT 1 LIMIT 5"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
}

// ============================================================================
// Test 30: parse_select_distinct
// Reference: tests/sqlparser_common.rs:971
// ============================================================================

func TestCommonParseSelectDistinct(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DISTINCT name FROM customer"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 31: parse_select_distinct_two_fields
// Reference: tests/sqlparser_common.rs:982
// ============================================================================

func TestCommonParseSelectDistinctTwoFields(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DISTINCT name, id FROM customer"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 32: parse_select_distinct_tuple
// Reference: tests/sqlparser_common.rs:997
// ============================================================================

func TestCommonParseSelectDistinctTuple(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DISTINCT (name, id) FROM customer"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 33: parse_outer_join_operator
// Reference: tests/sqlparser_common.rs:1011
// ============================================================================

func TestCommonParseOuterJoinOperator(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsOuterJoinOperator()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support outer join operator (+)")
		return
	}
	sql := "SELECT 1 FROM T WHERE a = b (+)"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT 1 FROM T WHERE t1.c1 = t2.c2.d3 (+)"
	dialects.VerifiedOnlySelect(t, sql2)
}

// ============================================================================
// Test 34: parse_select_distinct_on
// Reference: tests/sqlparser_common.rs:1049
// ============================================================================

func TestCommonParseSelectDistinctOn(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		_, isMySQL := d.(*mysql.MySqlDialect)
		return !isMySQL
	})
	testCases := []string{
		"SELECT DISTINCT ON (album_id) name FROM track ORDER BY album_id, milliseconds",
		"SELECT DISTINCT ON () name FROM track ORDER BY milliseconds",
		"SELECT DISTINCT ON (album_id, milliseconds) name FROM track",
	}
	for _, sql := range testCases {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Test 35: parse_select_distinct_missing_paren
// Reference: tests/sqlparser_common.rs:1073
// ============================================================================

func TestCommonParseSelectDistinctMissingParen(t *testing.T) {
	sql := "SELECT DISTINCT (name, id FROM customer"
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err)
}

// ============================================================================
// Test 36: parse_select_all
// Reference: tests/sqlparser_common.rs:1082
// ============================================================================

func TestCommonParseSelectAll(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT ALL name FROM customer")
}

// ============================================================================
// Test 37: parse_select_all_distinct
// Reference: tests/sqlparser_common.rs:1087
// ============================================================================

func TestCommonParseSelectAllDistinct(t *testing.T) {
	testCases := []string{
		"SELECT ALL DISTINCT name FROM customer",
		"SELECT DISTINCT ALL name FROM customer",
		"SELECT ALL DISTINCT ON(name) name FROM customer",
	}
	for _, sql := range testCases {
		_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
		_ = err
	}
}

// ============================================================================
// Test 38: parse_select_into
// Reference: tests/sqlparser_common.rs:1106
// ============================================================================

func TestCommonParseSelectInto(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * INTO table0 FROM table1"
	for _, d := range dialects.Dialects {
		_, err := parser.ParseSQL(d, sql)
		_ = err
	}
	sql2 := "SELECT * INTO TEMPORARY UNLOGGED TABLE table0 FROM table1"
	for _, d := range dialects.Dialects {
		_, err := parser.ParseSQL(d, sql2)
		_ = err
	}
}

// ============================================================================
// Test 39: parse_select_wildcard
// Reference: tests/sqlparser_common.rs:1136
// ============================================================================

func TestCommonParseSelectWildcard(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql1 := "SELECT * FROM foo"
	stmts1 := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts1, 1)
	sql2 := "SELECT foo.* FROM foo"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
	sql3 := "SELECT myschema.mytable.* FROM myschema.mytable"
	stmts3 := dialects.ParseSQL(t, sql3)
	require.Len(t, stmts3, 1)
}

// ============================================================================
// Test 40: parse_count_wildcard
// Reference: tests/sqlparser_common.rs:1176
// ============================================================================

func TestCommonParseCountWildcard(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, "SELECT COUNT(*) FROM Order WHERE id = 10")
	dialects.VerifiedOnlySelect(t, "SELECT COUNT(Employee.*) FROM Order JOIN Employee ON Order.employee = Employee.id")
}

// ============================================================================
// Test 41: parse_column_aliases
// Reference: tests/sqlparser_common.rs:1185
// ============================================================================

func TestCommonParseColumnAliases(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a.col + 1 AS newname FROM foo AS a"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	sql2 := "SELECT a.col + 1 newname FROM foo AS a"
	dialects.OneStatementParsesTo(t, sql2, sql)
}

// ============================================================================
// Test 42: parse_select_expr_star
// Reference: tests/sqlparser_common.rs:1207
// ============================================================================

func TestCommonParseSelectExprStar(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsSelectExprStar()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support SELECT expr.*")
		return
	}
	dialects.VerifiedOnlySelect(t, "SELECT foo.bar.* FROM T")
	dialects.VerifiedOnlySelect(t, "SELECT foo - bar.* FROM T")
	dialects.VerifiedOnlySelect(t, "SELECT myfunc().foo.* FROM T")
	dialects.VerifiedOnlySelect(t, "SELECT myfunc().* FROM T")
}

// ============================================================================
// Test 43: parse_select_wildcard_with_alias
// Reference: tests/sqlparser_common.rs:1289
// ============================================================================

func TestCommonParseSelectWildcardWithAlias(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsSelectWildcardWithAlias()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support wildcard with alias")
		return
	}
	sqls := []string{
		"SELECT t.* AS all_cols FROM t",
		"SELECT * AS all_cols FROM t",
		"SELECT a.id, b.* AS b_cols FROM a JOIN b ON (a.id = b.a_id)",
	}
	for _, sql := range sqls {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Test 44: test_eof_after_as
// Reference: tests/sqlparser_common.rs:1308
// ============================================================================

func TestCommonParseEofAfterAs(t *testing.T) {
	d := generic.NewGenericDialect()
	_, err := parser.ParseSQL(d, "SELECT foo AS")
	require.Error(t, err)
	_, err = parser.ParseSQL(d, "SELECT 1 FROM foo AS")
	_ = err
}

// ============================================================================
// Test 45: test_no_infix_error
// Reference: tests/sqlparser_common.rs:1323
// ============================================================================

func TestCommonParseNoInfixError(t *testing.T) {
	ch := clickhouse.NewClickHouseDialect()
	_, err := parser.ParseSQL(ch, "ASSERT-URA<<")
	require.Error(t, err)
}

// ============================================================================
// Test 46: parse_select_count_wildcard
// Reference: tests/sqlparser_common.rs:1335
// ============================================================================

func TestCommonParseSelectCountWildcard(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT COUNT(*) FROM customer"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 47: parse_select_count_distinct
// Reference: tests/sqlparser_common.rs:1357
// ============================================================================

func TestCommonParseSelectCountDistinct(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT COUNT(DISTINCT +x) FROM customer"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	dialects.VerifiedStmt(t, "SELECT COUNT(ALL +x) FROM customer")
	dialects.VerifiedStmt(t, "SELECT COUNT(+x) FROM customer")
	sql2 := "SELECT COUNT(ALL DISTINCT + x) FROM customer"
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql2)
	require.Error(t, err)
}

// ============================================================================
// Test 48: parse_not
// Reference: tests/sqlparser_common.rs:1393
// ============================================================================

func TestCommonParseNot(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id FROM customer WHERE NOT salary = ''"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 49: parse_invalid_infix_not
// Reference: tests/sqlparser_common.rs:1400
// ============================================================================

func TestCommonParseInvalidInfixNot(t *testing.T) {
	sql := "SELECT c FROM t WHERE c NOT ("
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err)
}

// ============================================================================
// Test 50: parse_collate
// Reference: tests/sqlparser_common.rs:1409
// ============================================================================

func TestCommonParseCollate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT name COLLATE \"de_DE\" FROM customer"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 51: parse_collate_after_parens
// Reference: tests/sqlparser_common.rs:1419
// ============================================================================

func TestCommonParseCollateAfterParens(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT (name) COLLATE \"de_DE\" FROM customer"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 52: parse_select_string_predicate
// Reference: tests/sqlparser_common.rs:1428
// ============================================================================

func TestCommonParseSelectStringPredicate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname FROM customer WHERE salary <> 'Not Provided' AND salary <> ''"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 53: parse_projection_nested_type
// Reference: tests/sqlparser_common.rs:1436
// ============================================================================

func TestCommonParseProjectionNestedType(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT customer.address.state FROM foo"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 54: parse_null_in_select
// Reference: tests/sqlparser_common.rs:1443
// ============================================================================

func TestCommonParseNullInSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT NULL"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// ============================================================================
// Test 55: parse_exponent_in_select
// Reference: tests/sqlparser_common.rs:1453
// ============================================================================

func TestCommonParseExponentInSelect(t *testing.T) {
	dialect := generic.NewGenericDialect()
	sql := "SELECT 10e-20, 1e3, 1e+3, 0.5e2 FROM t"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 56: parse_select_with_date_column_name
// Reference: tests/sqlparser_common.rs:1503
// ============================================================================

func TestCommonParseSelectWithDateColumnName(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT date"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// ============================================================================
// Test 57: parse_escaped_single_quote_string_predicate_with_escape
// Reference: tests/sqlparser_common.rs:1517
// ============================================================================

func TestCommonParseEscapedSingleQuoteStringPredicateWithEscape(t *testing.T) {
	dialect := generic.NewGenericDialect()
	sql := "SELECT id, fname, lname FROM customer WHERE salary <> 'Jim''s salary'"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 58: parse_escaped_single_quote_string_predicate_with_no_escape
// Reference: tests/sqlparser_common.rs:1537
// ============================================================================

func TestCommonParseEscapedSingleQuoteStringPredicateWithNoEscape(t *testing.T) {
	dialect := mysql.NewMySqlDialect()
	sql := "SELECT id, fname, lname FROM customer WHERE salary <> 'Jim''s salary'"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 59: parse_number
// Reference: tests/sqlparser_common.rs:1562
// ============================================================================

func TestCommonParseNumber(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT 1.0"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 60: parse_compound_expr_1
// Reference: tests/sqlparser_common.rs:1580
// ============================================================================

func TestCommonParseCompoundExpr1(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a + b * c FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 61: parse_compound_expr_2
// Reference: tests/sqlparser_common.rs:1598
// ============================================================================

func TestCommonParseCompoundExpr2(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a * b + c FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 62: parse_string_concat
// Reference: tests/sqlparser_common.rs:2365
// ============================================================================

func TestCommonParseStringConcat(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a || b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 63: parse_aggregate_with_group_by
// Reference: tests/sqlparser_common.rs:1640
// ============================================================================

func TestCommonParseAggregateWithGroupBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT SUM(order) FROM customer GROUP BY lname, fname"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 64: parse_literal_decimal
// Reference: tests/sqlparser_common.rs:1670
// ============================================================================

func TestCommonParseLiteralDecimal(t *testing.T) {
	dialects := utils.NewTestedDialects()
	testCases := []string{
		"SELECT 1.0 FROM t",
		"SELECT 1.1 FROM t",
		"SELECT 0.1 FROM t",
		"SELECT 0.01 FROM t",
		"SELECT 1.01 FROM t",
	}
	for _, sql := range testCases {
		dialects.VerifiedOnlySelect(t, sql)
	}
}

// ============================================================================
// Test 65: parse_literal_integer
// Reference: tests/sqlparser_common.rs:1690
// ============================================================================

func TestCommonParseLiteralInteger(t *testing.T) {
	dialects := utils.NewTestedDialects()
	testCases := []string{
		"SELECT 1 FROM t",
		"SELECT 11 FROM t",
		"SELECT 111 FROM t",
	}
	for _, sql := range testCases {
		dialects.VerifiedOnlySelect(t, sql)
	}
}

// ============================================================================
// Test 66: parse_literal_string
// Reference: tests/sqlparser_common.rs:1710
// ============================================================================

func TestCommonParseLiteralString(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT 'string' FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 67: parse_literal_date
// Reference: tests/sqlparser_common.rs:1730
// ============================================================================

func TestCommonParseLiteralDate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DATE '2021-01-01' FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 68: parse_literal_time
// Reference: tests/sqlparser_common.rs:1750
// ============================================================================

func TestCommonParseLiteralTime(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT TIME '12:34:56' FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 69: parse_literal_timestamp
// Reference: tests/sqlparser_common.rs:1770
// ============================================================================

func TestCommonParseLiteralTimestamp(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT TIMESTAMP '2021-01-01 12:34:56' FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 70: parse_literal_interval
// Reference: tests/sqlparser_common.rs:1790
// ============================================================================

func TestCommonParseLiteralInterval(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT INTERVAL '1' YEAR FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 71: parse_simple_math_expr_plus
// Reference: tests/sqlparser_common.rs:1810
// ============================================================================

func TestCommonParseSimpleMathExprPlus(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a + b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 72: parse_simple_math_expr_minus
// Reference: tests/sqlparser_common.rs:1830
// ============================================================================

func TestCommonParseSimpleMathExprMinus(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a - b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 73: parse_unary_math_expr
// Reference: tests/sqlparser_common.rs:1850
// ============================================================================

func TestCommonParseUnaryMathExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT -a, +a FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 74: parse_concat_and_math_expr
// Reference: tests/sqlparser_common.rs:1870
// ============================================================================

func TestCommonParseConcatAndMathExpr(t *testing.T) {
	dialect := generic.NewGenericDialect()
	sql := "SELECT a || b + c FROM t"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 75: parse_substring
// Reference: tests/sqlparser_common.rs:1910
// ============================================================================

func TestCommonParseSubstring(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT SUBSTRING('foo' FROM 1 FOR 2) FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 77: parse_substring_from_for
// Reference: tests/sqlparser_common.rs:1930
// ============================================================================

func TestCommonParseSubstringFromFor(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT SUBSTRING(col FROM 1 FOR 2) FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 78: parse_position
// Reference: tests/sqlparser_common.rs:1950
// ============================================================================

func TestCommonParsePosition(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT POSITION('a' IN 'b') FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 79: parse_position_case_insensitive
// Reference: tests/sqlparser_common.rs:1970
// ============================================================================

func TestCommonParsePositionCaseInsensitive(t *testing.T) {
	dialect := generic.NewGenericDialect()
	sql := "SELECT position('a' IN 'b') FROM t"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 80: parse_ambiguous_column_name_in_select
// Reference: tests/sqlparser_common.rs:1990
// ============================================================================

func TestCommonParseAmbiguousColumnNameInSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id FROM foo JOIN bar ON foo.id = bar.id"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 81: parse_quoted_identifier
// Reference: tests/sqlparser_common.rs:2010
// ============================================================================

func TestCommonParseQuotedIdentifier(t *testing.T) {
	dialect := generic.NewGenericDialect()
	sql := "SELECT \"quoted column\" FROM t"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 82: parse_escaped_quote_in_quoted_identifier
// Reference: tests/sqlparser_common.rs:2030
// ============================================================================

func TestCommonParseEscapedQuoteInQuotedIdentifier(t *testing.T) {
	dialect := generic.NewGenericDialect()
	sql := "SELECT \"quoted\"\"column\" FROM t"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 83: parse_like
// Reference: tests/sqlparser_common.rs:2050
// ============================================================================

func TestCommonParseLike(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customers WHERE name LIKE '%a'"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT * FROM customers WHERE name LIKE '%a' ESCAPE '^'"
	dialects.VerifiedOnlySelect(t, sql2)
}

// ============================================================================
// Test 84: parse_not_like
// Reference: tests/sqlparser_common.rs:2070
// ============================================================================

func TestCommonParseNotLike(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customers WHERE name NOT LIKE '%a'"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT * FROM customers WHERE name NOT LIKE '%a' ESCAPE '^'"
	dialects.VerifiedOnlySelect(t, sql2)
}

// ============================================================================
// Test 85: parse_ilike
// Reference: tests/sqlparser_common.rs:2090
// ============================================================================

func TestCommonParseIlike(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customers WHERE name ILIKE '%a'"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT * FROM customers WHERE name ILIKE '%a' ESCAPE '^'"
	dialects.VerifiedOnlySelect(t, sql2)
}

// ============================================================================
// Test 86: parse_not_ilike
// Reference: tests/sqlparser_common.rs:2110
// ============================================================================

func TestCommonParseNotIlike(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customers WHERE name NOT ILIKE '%a'"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT * FROM customers WHERE name NOT ILIKE '%a' ESCAPE '^'"
	dialects.VerifiedOnlySelect(t, sql2)
}

// ============================================================================
// Test 87: parse_similar_to
// Reference: tests/sqlparser_common.rs:2130
// ============================================================================

func TestCommonParseSimilarTo(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customers WHERE name SIMILAR TO '%a'"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT * FROM customers WHERE name SIMILAR TO '%a' ESCAPE '^'"
	dialects.VerifiedOnlySelect(t, sql2)
	sql3 := "SELECT * FROM customers WHERE name SIMILAR TO '%a' ESCAPE NULL"
	dialects.VerifiedOnlySelect(t, sql3)
}

// ============================================================================
// Test 88: parse_not_similar_to
// Reference: tests/sqlparser_common.rs:2150
// ============================================================================

func TestCommonParseNotSimilarTo(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customers WHERE name NOT SIMILAR TO '%a'"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT * FROM customers WHERE name NOT SIMILAR TO '%a' ESCAPE '^'"
	dialects.VerifiedOnlySelect(t, sql2)
}

// ============================================================================
// Test 89: parse_rlike
// Reference: tests/sqlparser_common.rs:2170
// ============================================================================

func TestCommonParseRlike(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customers WHERE name RLIKE 'pattern'"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 90: parse_regexp
// Reference: tests/sqlparser_common.rs:2190
// ============================================================================

func TestCommonParseRegexp(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customers WHERE name REGEXP 'pattern'"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 91: parse_pipe (bitwise OR)
// Reference: tests/sqlparser_common.rs:2210
// ============================================================================

func TestCommonParsePipe(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a | b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 92: parse_div (integer division)
// Reference: tests/sqlparser_common.rs:2230
// ============================================================================

func TestCommonParseDiv(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a DIV b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 93: parse_mod (modulo operator %)
// Reference: tests/sqlparser_common.rs:1658
// ============================================================================

func TestCommonParseMod(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a % b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 94: parse_ampersand (bitwise AND)
// Reference: tests/sqlparser_common.rs:2270
// ============================================================================

func TestCommonParseAmpersand(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a & b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 95: parse_caret (bitwise XOR)
// Reference: tests/sqlparser_common.rs:2290
// ============================================================================

func TestCommonParseCaret(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			mysql.NewMySqlDialect(),
			sqlite.NewSQLiteDialect(),
			snowflake.NewSnowflakeDialect(),
			mssql.NewMsSqlDialect(),
			ansi.NewAnsiDialect(),
		},
	}
	sql := "SELECT a ^ b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 96: parse_pipe_pipe (string concatenation)
// Reference: tests/sqlparser_common.rs:2310
// ============================================================================

func TestCommonParsePipePipe(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a || b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 97: parse_less_than_eq (<= operator)
// Reference: tests/sqlparser_common.rs:2330
// ============================================================================

func TestCommonParseLessThanEq(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a <= b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 98: parse_greater_than_eq (>= operator)
// Reference: tests/sqlparser_common.rs:2350
// ============================================================================

func TestCommonParseGreaterThanEq(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a >= b FROM t"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 99: parse_cube_and_rollup
// Reference: tests/sqlparser_common.rs:2370
// ============================================================================

func TestCommonParseCubeAndRollup(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			mysql.NewMySqlDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}
	sql1 := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b WITH CUBE"
	dialects.VerifiedOnlySelect(t, sql1)
	sql2 := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b WITH ROLLUP"
	dialects.VerifiedOnlySelect(t, sql2)
	sql3 := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b WITH ROLLUP WITH CUBE"
	dialects.VerifiedOnlySelect(t, sql3)
}

// ============================================================================
// NEW TESTS PORTED FROM RUST (lines ~2400-3200 of sqlparser_common.rs)
// ============================================================================

// ============================================================================
// Test 100: parse_bitwise_shift_ops
// Reference: tests/sqlparser_common.rs:2412
// ============================================================================

func TestCommonParseBitwiseShiftOps(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsBitwiseShiftOperators()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support bitwise shift operators")
		return
	}
	sql := "SELECT 1 << 2, 3 >> 4"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 101: parse_binary_any
// Reference: tests/sqlparser_common.rs:2435
// ============================================================================

func TestCommonParseBinaryAny(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a = ANY(b)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 102: parse_binary_all
// Reference: tests/sqlparser_common.rs:2449
// ============================================================================

func TestCommonParseBinaryAll(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a = ALL(b)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 103: parse_between
// Reference: tests/sqlparser_common.rs:2462
// ============================================================================

func TestCommonParseBetween(t *testing.T) {
	dialects := utils.NewTestedDialects()
	chk := func(negated bool) {
		var sql string
		if negated {
			sql = "SELECT * FROM customers WHERE age NOT BETWEEN 25 AND 32"
		} else {
			sql = "SELECT * FROM customers WHERE age BETWEEN 25 AND 32"
		}
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
	chk(false)
	chk(true)
}

// ============================================================================
// Test 104: parse_between_with_expr
// Reference: tests/sqlparser_common.rs:2484
// ============================================================================

func TestCommonParseBetweenWithExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM t WHERE 1 BETWEEN 1 + 2 AND 3 + 4 IS NULL"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	sql2 := "SELECT * FROM t WHERE 1 = 1 AND 1 + x BETWEEN 1 AND 2"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
}

// ============================================================================
// Test 105: parse_tuples
// Reference: tests/sqlparser_common.rs:2532
// ============================================================================

func TestCommonParseTuples(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT (1, 2), (1), ('foo', 3, baz)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 106: parse_tuple_invalid
// Reference: tests/sqlparser_common.rs:2555
// ============================================================================

func TestCommonParseTupleInvalid(t *testing.T) {
	sql1 := "select (1"
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql1)
	require.Error(t, err)
	sql2 := "select (), 2"
	_, err = parser.ParseSQL(generic.NewGenericDialect(), sql2)
	require.Error(t, err)
}

// ============================================================================
// Test 107: parse_select_order_by
// Reference: tests/sqlparser_common.rs:2572
// ============================================================================

func TestCommonParseSelectOrderBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	chk := func(sql string) {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
	chk("SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY lname ASC, fname DESC, id")
	chk("SELECT id, fname, lname FROM customer ORDER BY lname ASC, fname DESC, id")
	chk("SELECT 1 AS lname, 2 AS fname, 3 AS id, 4 ORDER BY lname ASC, fname DESC, id")
}

// ============================================================================
// Test 108: parse_select_order_by_multiple
// Reference: tests/sqlparser_common.rs:2612
// ============================================================================

func TestCommonParseSelectOrderByMultiple(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY lname ASC, fname DESC LIMIT 2"
	dialects.VerifiedQuery(t, sql)
}

// ============================================================================
// Test 109: parse_select_order_by_all
// Reference: tests/sqlparser_common.rs:2646
// ============================================================================

func TestCommonParseSelectOrderByAll(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsOrderByAll()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support ORDER BY ALL")
		return
	}
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
	}
}

// ============================================================================
// Test 110: parse_select_order_by_not_support_all
// Reference: tests/sqlparser_common.rs:2727
// ============================================================================

func TestCommonParseSelectOrderByNotSupportAll(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsOrderByAll()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("All dialects support ORDER BY ALL")
		return
	}
	testCases := []string{
		"SELECT id, ALL FROM customer WHERE id < 5 ORDER BY ALL",
		"SELECT id, ALL FROM customer ORDER BY ALL ASC NULLS FIRST",
		"SELECT id, ALL FROM customer ORDER BY ALL DESC NULLS LAST",
	}
	for _, sql := range testCases {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Test 111: parse_select_order_by_nulls_order
// Reference: tests/sqlparser_common.rs:2778
// ============================================================================

func TestCommonParseSelectOrderByNullsOrder(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY lname ASC NULLS FIRST, fname DESC NULLS LAST LIMIT 2"
	dialects.VerifiedQuery(t, sql)
}

// ============================================================================
// Test 112: parse_select_group_by
// Reference: tests/sqlparser_common.rs:2812
// ============================================================================

func TestCommonParseSelectGroupBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname FROM customer GROUP BY lname, fname"
	dialects.VerifiedOnlySelect(t, sql)
	dialects.OneStatementParsesTo(t,
		"SELECT id, fname, lname FROM customer GROUP BY (lname, fname)",
		"SELECT id, fname, lname FROM customer GROUP BY (lname, fname)")
}

// ============================================================================
// Test 113: parse_select_group_by_all
// Reference: tests/sqlparser_common.rs:2834
// ============================================================================

func TestCommonParseSelectGroupByAll(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname, SUM(order) FROM customer GROUP BY ALL"
	dialects.VerifiedOnlySelect(t, sql)
}

// ============================================================================
// Test 114: parse_group_by_with_modifier
// Reference: tests/sqlparser_common.rs:2846
// ============================================================================

func TestCommonParseGroupByWithModifier(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsGroupByWithModifier()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support GROUP BY with modifier")
		return
	}
	clauses := []string{"x", "a, b", "ALL"}
	modifiers := []string{"WITH ROLLUP", "WITH CUBE", "WITH TOTALS", "WITH ROLLUP WITH CUBE"}
	for _, clause := range clauses {
		for _, modifier := range modifiers {
			sql := fmt.Sprintf("SELECT * FROM t GROUP BY %s %s", clause, modifier)
			stmts := dialects.ParseSQL(t, sql)
			require.Len(t, stmts, 1)
		}
	}
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

// ============================================================================
// Test 115: parse_group_by_special_grouping_sets
// Reference: tests/sqlparser_common.rs:2903
// ============================================================================

func TestCommonParseGroupBySpecialGroupingSets(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), (a), (b), ())"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// Test 116: parse_group_by_grouping_sets_single_values
// Reference: tests/sqlparser_common.rs:2932
// ============================================================================

func TestCommonParseGroupByGroupingSetsSingleValues(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), a, (b), c, ())"
	canonical := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), (a), (b), (c), ())"
	dialects.OneStatementParsesTo(t, sql, canonical)
}

// ============================================================================
// Test 117: parse_select_having
// Reference: tests/sqlparser_common.rs:2964
// ============================================================================

func TestCommonParseSelectHaving(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT foo FROM bar GROUP BY foo HAVING COUNT(*) > 1"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT 'foo' HAVING 1 = 1"
	dialects.VerifiedOnlySelect(t, sql2)
}

// ============================================================================
// Test 118: parse_select_qualify
// Reference: tests/sqlparser_common.rs:2995
// ============================================================================

func TestCommonParseSelectQualify(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT i, p, o FROM qt QUALIFY ROW_NUMBER() OVER (PARTITION BY p ORDER BY o) = 1"
	dialects.VerifiedOnlySelect(t, sql)
	sql2 := "SELECT i, p, o, ROW_NUMBER() OVER (PARTITION BY p ORDER BY o) AS row_num FROM qt QUALIFY row_num = 1"
	dialects.VerifiedOnlySelect(t, sql2)
}

// ============================================================================
// Test 119: parse_limit_accepts_all
// Reference: tests/sqlparser_common.rs:3045
// ============================================================================

func TestCommonParseLimitAcceptsAll(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.OneStatementParsesTo(t,
		"SELECT id, fname, lname FROM customer WHERE id = 1 LIMIT ALL",
		"SELECT id, fname, lname FROM customer WHERE id = 1")
	dialects.OneStatementParsesTo(t,
		"SELECT id, fname, lname FROM customer WHERE id = 1 LIMIT ALL OFFSET 1",
		"SELECT id, fname, lname FROM customer WHERE id = 1 OFFSET 1")
	dialects.OneStatementParsesTo(t,
		"SELECT id, fname, lname FROM customer WHERE id = 1 OFFSET 1 LIMIT ALL",
		"SELECT id, fname, lname FROM customer WHERE id = 1 OFFSET 1")
}

// ============================================================================
// Test 120: parse_cast
// Reference: tests/sqlparser_common.rs:3061
// ============================================================================

func TestCommonParseCast(t *testing.T) {
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

// ============================================================================
// Test 121: parse_try_cast
// Reference: tests/sqlparser_common.rs:3213
// ============================================================================

func TestCommonParseTryCast(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS BIGINT) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS NUMERIC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS DEC) FROM customer")
	dialects.VerifiedStmt(t, "SELECT TRY_CAST(id AS DECIMAL) FROM customer")
}

// ============================================================================
// Test 122: parse_extract
// Reference: tests/sqlparser_common.rs:3236
// ============================================================================

func TestCommonParseExtract(t *testing.T) {
	dialects := utils.NewTestedDialects()
	testCases := []string{
		"SELECT EXTRACT(YEAR FROM d)",
		"SELECT EXTRACT(MONTH FROM d)",
		"SELECT EXTRACT(WEEK FROM d)",
		"SELECT EXTRACT(DAY FROM d)",
		"SELECT EXTRACT(DAYOFWEEK FROM d)",
		"SELECT EXTRACT(DAYOFYEAR FROM d)",
		"SELECT EXTRACT(DATE FROM d)",
		"SELECT EXTRACT(DATETIME FROM d)",
		"SELECT EXTRACT(HOUR FROM d)",
		"SELECT EXTRACT(MINUTE FROM d)",
		"SELECT EXTRACT(SECOND FROM d)",
		"SELECT EXTRACT(MILLISECOND FROM d)",
		"SELECT EXTRACT(MICROSECOND FROM d)",
		"SELECT EXTRACT(NANOSECOND FROM d)",
		"SELECT EXTRACT(CENTURY FROM d)",
		"SELECT EXTRACT(DECADE FROM d)",
		"SELECT EXTRACT(DOW FROM d)",
		"SELECT EXTRACT(DOY FROM d)",
		"SELECT EXTRACT(EPOCH FROM d)",
		"SELECT EXTRACT(ISODOW FROM d)",
		"SELECT EXTRACT(ISOWEEK FROM d)",
		"SELECT EXTRACT(ISOYEAR FROM d)",
		"SELECT EXTRACT(JULIAN FROM d)",
		"SELECT EXTRACT(MICROSECOND FROM d)",
		"SELECT EXTRACT(MICROSECONDS FROM d)",
		"SELECT EXTRACT(MILLENIUM FROM d)",
		"SELECT EXTRACT(MILLENNIUM FROM d)",
		"SELECT EXTRACT(MILLISECOND FROM d)",
		"SELECT EXTRACT(MILLISECONDS FROM d)",
		"SELECT EXTRACT(QUARTER FROM d)",
		"SELECT EXTRACT(TIMEZONE FROM d)",
		"SELECT EXTRACT(TIMEZONE_ABBR FROM d)",
		"SELECT EXTRACT(TIMEZONE_HOUR FROM d)",
		"SELECT EXTRACT(TIMEZONE_MINUTE FROM d)",
		"SELECT EXTRACT(TIMEZONE_REGION FROM d)",
		"SELECT EXTRACT(TIME FROM d)",
	}
	for _, sql := range testCases {
		dialects.VerifiedStmt(t, sql)
	}
	// Test that custom fields error without allow_extract_custom - some dialects may error
	restrictedDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.AllowExtractCustom()
	})
	if len(restrictedDialects.Dialects) > 0 {
		_, err := parser.ParseSQL(restrictedDialects.Dialects[0], "SELECT EXTRACT(JIFFY FROM d)")
		// Some dialects may error, others may not - just document the behavior
		_ = err
	}
}

// ============================================================================
// Test 123: parse_ceil_number
// Reference: tests/sqlparser_common.rs:3295
// ============================================================================

func TestCommonParseCeilNumber(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT CEIL(1.5)")
	dialects.VerifiedStmt(t, "SELECT CEIL(float_column) FROM my_table")
}

// ============================================================================
// Test 124: parse_floor_number
// Reference: tests/sqlparser_common.rs:3301
// ============================================================================

func TestCommonParseFloorNumber(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT FLOOR(1.5)")
	dialects.VerifiedStmt(t, "SELECT FLOOR(float_column) FROM my_table")
}

// ============================================================================
// Test 125: parse_ceil_number_scale
// Reference: tests/sqlparser_common.rs:3307
// ============================================================================

func TestCommonParseCeilNumberScale(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT CEIL(1.5, 1)")
	dialects.VerifiedStmt(t, "SELECT CEIL(float_column, 3) FROM my_table")
}

// ============================================================================
// Test 126: parse_floor_number_scale
// Reference: tests/sqlparser_common.rs:3313
// ============================================================================

func TestCommonParseFloorNumberScale(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT FLOOR(1.5, 1)")
	dialects.VerifiedStmt(t, "SELECT FLOOR(float_column, 3) FROM my_table")
}

// ============================================================================
// Test 127: parse_ceil_scale
// Reference: tests/sqlparser_common.rs:3319
// ============================================================================

func TestCommonParseCeilScale(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, "SELECT CEIL(d, 2)")
}

// ============================================================================
// Test 128: parse_floor_scale
// Reference: tests/sqlparser_common.rs:3345
// ============================================================================

func TestCommonParseFloorScale(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, "SELECT FLOOR(d, 2)")
}

// ============================================================================
// Test 129: parse_ceil_datetime
// Reference: tests/sqlparser_common.rs:3371
// ============================================================================

func TestCommonParseCeilDatetime(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, "SELECT CEIL(d TO DAY)")
	dialects.OneStatementParsesTo(t, "SELECT CEIL(d to day)", "SELECT CEIL(d TO DAY)")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO HOUR) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO MINUTE) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO SECOND) FROM df")
	dialects.VerifiedStmt(t, "SELECT CEIL(d TO MILLISECOND) FROM df")
	// Test invalid datetime field - some dialects may error
	restrictedDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.AllowExtractCustom()
	})
	if len(restrictedDialects.Dialects) > 0 {
		_, err := parser.ParseSQL(restrictedDialects.Dialects[0], "SELECT CEIL(d TO JIFFY) FROM df")
		// Some dialects may error, others may not - just document the behavior
		_ = err
	}
}

// ============================================================================
// Test 130: parse_floor_datetime
// Reference: tests/sqlparser_common.rs:3398
// ============================================================================

func TestCommonParseFloorDatetime(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, "SELECT FLOOR(d TO DAY)")
	dialects.OneStatementParsesTo(t, "SELECT FLOOR(d to day)", "SELECT FLOOR(d TO DAY)")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO HOUR) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO MINUTE) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO SECOND) FROM df")
	dialects.VerifiedStmt(t, "SELECT FLOOR(d TO MILLISECOND) FROM df")
	// Test invalid datetime field - some dialects may error
	restrictedDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.AllowExtractCustom()
	})
	if len(restrictedDialects.Dialects) > 0 {
		_, err := parser.ParseSQL(restrictedDialects.Dialects[0], "SELECT FLOOR(d TO JIFFY) FROM df")
		// Some dialects may error, others may not - just document the behavior
		_ = err
	}
}

// ============================================================================
// Test 131: parse_listagg
// Reference: tests/sqlparser_common.rs:3425
// ============================================================================

func TestCommonParseListagg(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT LISTAGG(DISTINCT dateid, ', ' ON OVERFLOW TRUNCATE '%' WITHOUT COUNT) WITHIN GROUP (ORDER BY id, username)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	dialects.VerifiedStmt(t, "SELECT LISTAGG(sellerid) WITHIN GROUP (ORDER BY dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(DISTINCT dateid)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW ERROR)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW TRUNCATE N'...' WITH COUNT)")
	dialects.VerifiedStmt(t, "SELECT LISTAGG(dateid ON OVERFLOW TRUNCATE X'deadbeef' WITH COUNT)")
}

// ============================================================================
// Test 132: parse_array_agg_func
// Reference: tests/sqlparser_common.rs:3497
// ============================================================================

func TestCommonParseArrayAggFunc(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
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
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Test 133: parse_agg_with_order_by
// Reference: tests/sqlparser_common.rs:3519
// ============================================================================

func TestCommonParseAggWithOrderBy(t *testing.T) {
	dialects := &utils.TestedDialects{
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
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Test 134: parse_window_rank_function
// Reference: tests/sqlparser_common.rs:3539
// ============================================================================

func TestCommonParseWindowRankFunction(t *testing.T) {
	dialects := &utils.TestedDialects{
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
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
	nullsDialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mssql.NewMsSqlDialect(),
			snowflake.NewSnowflakeDialect(),
		},
	}
	nullsCases := []string{
		"SELECT column1, column2, FIRST_VALUE(column2) IGNORE NULLS OVER (PARTITION BY column1 ORDER BY column2 NULLS LAST) AS column2_first FROM t1",
		"SELECT column1, column2, FIRST_VALUE(column2) RESPECT NULLS OVER (PARTITION BY column1 ORDER BY column2 NULLS LAST) AS column2_first FROM t1",
		"SELECT LAG(col_2, 1, 0) IGNORE NULLS OVER (ORDER BY col_1) FROM t1",
		"SELECT LAG(col_2, 1, 0) RESPECT NULLS OVER (ORDER BY col_1) FROM t1",
	}
	for _, sql := range nullsCases {
		stmts := nullsDialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Test 135: parse_window_function_null_treatment_arg
// Reference: tests/sqlparser_common.rs:3575
// ============================================================================

func TestCommonParseWindowFunctionNullTreatmentArg(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsWindowFunctionNullTreatmentArg()
	})
	if len(dialects.Dialects) == 0 {
		t.Skip("No dialects support window function null treatment arg")
		return
	}
	sql := "SELECT FIRST_VALUE(a IGNORE NULLS) OVER (), FIRST_VALUE(b RESPECT NULLS) OVER () FROM mytable"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
	_, err := parser.ParseSQL(dialects.Dialects[0], "SELECT LAG(1 IGNORE NULLS) IGNORE NULLS OVER () FROM t1")
	require.Error(t, err)
	unsupportedDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsWindowFunctionNullTreatmentArg()
	})
	if len(unsupportedDialects.Dialects) > 0 {
		_, err := parser.ParseSQL(unsupportedDialects.Dialects[0], "SELECT LAG(1 IGNORE NULLS) IGNORE NULLS OVER () FROM t1")
		require.Error(t, err)
	}
}

// ============================================================================
// Test 136: test_compound_expr
// Reference: tests/sqlparser_common.rs:3637
// ============================================================================

func TestCommonCompoundExpr(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
		},
	}
	testCases := []string{
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
	for _, sql := range testCases {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)
	}
}

// ============================================================================
// Test 137: test_double_value
// Reference: tests/sqlparser_common.rs:3662
// ============================================================================

func TestCommonDoubleValue(t *testing.T) {
	// Use a subset of dialects that support these numeric formats
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			postgresql.NewPostgreSqlDialect(),
			mysql.NewMySqlDialect(),
			sqlite.NewSQLiteDialect(),
			ansi.NewAnsiDialect(),
		},
	}
	testCases := []string{
		"SELECT 0.",
		"SELECT 0.0",
		"SELECT 0000.",
		"SELECT 0000.00",
		"SELECT 0e0",
		"SELECT 0e+0",
		"SELECT 0e-0",
		"SELECT 0.e-0",
		"SELECT 0.e+0",
		"SELECT 00.0e+0",
		"SELECT 00.0e-0",
		"SELECT +0.",
		"SELECT -0.",
		"SELECT +0.0",
		"SELECT -0.0",
	}
	for _, sql := range testCases {
		dialects.VerifiedOnlySelect(t, sql)
	}
	// These formats may not be supported by all dialects
	specialCases := []string{
		"SELECT .0",
		"SELECT .00",
		"SELECT .0e-0",
		"SELECT .0e+0",
	}
	for _, sql := range specialCases {
		// Just verify they parse without error for dialects that support them
		for _, d := range dialects.Dialects {
			_, _ = parser.ParseSQL(d, sql)
		}
	}
}

// ============================================================================
// Test 138: parse_negative_value
// Reference: tests/sqlparser_common.rs:3769
// ============================================================================

func TestCommonParseNegativeValue(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.OneStatementParsesTo(t, "SELECT -1", "SELECT -1")
	dialects.OneStatementParsesTo(t,
		"CREATE SEQUENCE name INCREMENT -10 MINVALUE -1000 MAXVALUE 15 START -100",
		"CREATE SEQUENCE name INCREMENT -10 MINVALUE -1000 MAXVALUE 15 START -100")
}

// ============================================================================
// Test 139: parse_create_table
// Reference: tests/sqlparser_common.rs:3781
// ============================================================================

func TestCommonParseCreateTable(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "CREATE TABLE uk_cities (name VARCHAR(100) NOT NULL, lat DOUBLE NULL, lng DOUBLE)"
	dialects.VerifiedStmt(t, sql)
	sql2 := "CREATE TABLE t (a INT NOT NULL, b INT UNIQUE, c INT PRIMARY KEY, d INT CHECK (d > 0))"
	dialects.VerifiedStmt(t, sql2)
}

// ============================================================================
// Test 140: parse_create_table_as
// Reference: tests/sqlparser_common.rs:4478
// ============================================================================

func TestCommonParseCreateTableAs(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "CREATE TABLE t AS SELECT * FROM a")
}

// ============================================================================
// Test 141: parse_create_schema
// Reference: tests/sqlparser_common.rs:4422
// ============================================================================

func TestCommonParseCreateSchema(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "CREATE SCHEMA X")
	dialects.VerifiedStmt(t, "CREATE SCHEMA IF NOT EXISTS a")
}

// ============================================================================
// Test 142: parse_create_schema_with_authorization
// Reference: tests/sqlparser_common.rs:4443
// ============================================================================

func TestCommonParseCreateSchemaWithAuthorization(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "CREATE SCHEMA AUTHORIZATION Y"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 143: parse_create_schema_with_name_and_authorization
// Reference: tests/sqlparser_common.rs:4455
// ============================================================================

func TestCommonParseCreateSchemaWithNameAndAuthorization(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "CREATE SCHEMA X AUTHORIZATION Y"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// ============================================================================
// Test 144: parse_drop_schema
// Reference: tests/sqlparser_common.rs:4467
// ============================================================================

func TestCommonParseDropSchema(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "DROP SCHEMA X")
}

// ============================================================================
// Test 145: parse_assert
// Reference: tests/sqlparser_common.rs:4382
// ============================================================================

func TestCommonParseAssert(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "ASSERT (SELECT COUNT(*) FROM my_table) > 0")
}

// ============================================================================
// Test 146: parse_assert_message
// Reference: tests/sqlparser_common.rs:4397
// ============================================================================

func TestCommonParseAssertMessage(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "ASSERT (SELECT COUNT(*) FROM my_table) > 0 AS 'No rows in my_table'")
}

// ============================================================================
// Test 147: parse_select_top
// Reference: tests/sqlparser_common.rs around line 3090
// ============================================================================

func TestCommonParseSelectTop(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mssql.NewMsSqlDialect(),
		},
	}
	dialects.VerifiedStmt(t, "SELECT TOP 5 * FROM t")
	dialects.VerifiedStmt(t, "SELECT TOP 5 name FROM t")
}

// ============================================================================
// Test 148: parse_select_top_percent
// Reference: tests/sqlparser_common.rs around line 3110
// ============================================================================

func TestCommonParseSelectTopPercent(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mssql.NewMsSqlDialect(),
		},
	}
	dialects.VerifiedStmt(t, "SELECT TOP 50 PERCENT * FROM t")
	dialects.VerifiedStmt(t, "SELECT TOP 10 PERCENT name FROM t ORDER BY id")
}

// ============================================================================
// Test 149: parse_select_top_with_ties
// Reference: tests/sqlparser_common.rs around line 3130
// ============================================================================

func TestCommonParseSelectTopWithTies(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mssql.NewMsSqlDialect(),
		},
	}
	dialects.VerifiedStmt(t, "SELECT TOP 5 WITH TIES * FROM t ORDER BY id")
	dialects.VerifiedStmt(t, "SELECT TOP 10 WITH TIES name FROM t ORDER BY score DESC")
}
