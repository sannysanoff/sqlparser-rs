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
// This file contains tests 361-380 from the Rust test file.
package common

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
	"github.com/user/sqlparser/token"
)

// TestParseListenChannel verifies LISTEN statement parsing.
// Reference: tests/sqlparser_common.rs:14784
func TestParseListenChannel(t *testing.T) {
	// Test with dialects that support LISTEN/NOTIFY
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsListenNotify()
	})

	stmts := dialects.ParseSQL(t, "LISTEN test1")
	require.Len(t, stmts, 1)

	listen, ok := stmts[0].(*statement.Listen)
	require.True(t, ok, "Expected Listen statement, got %T", stmts[0])
	assert.Equal(t, "test1", listen.Channel.String())

	// Test parsing error for wildcard
	errDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsListenNotify()
	})
	for _, d := range errDialects.Dialects {
		_, err := parser.ParseSQL(d, "LISTEN *")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Expected: identifier")
	}

	// Test with dialects that don't support LISTEN/NOTIFY
	notSupportDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsListenNotify()
	})
	for _, d := range notSupportDialects.Dialects {
		_, err := parser.ParseSQL(d, "LISTEN test1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Expected: an SQL statement")
	}
}

// TestParseUnlistenChannel verifies UNLISTEN statement parsing.
// Reference: tests/sqlparser_common.rs:14808
func TestParseUnlistenChannel(t *testing.T) {
	// Test with dialects that support LISTEN/NOTIFY
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsListenNotify()
	})

	stmts := dialects.ParseSQL(t, "UNLISTEN test1")
	require.Len(t, stmts, 1)

	unlisten, ok := stmts[0].(*statement.Unlisten)
	require.True(t, ok, "Expected Unlisten statement, got %T", stmts[0])
	assert.Equal(t, "test1", unlisten.Channel.String())

	// Test wildcard
	stmts2 := dialects.ParseSQL(t, "UNLISTEN *")
	require.Len(t, stmts2, 1)

	unlisten2, ok := stmts2[0].(*statement.Unlisten)
	require.True(t, ok, "Expected Unlisten statement, got %T", stmts2[0])
	assert.Equal(t, "*", unlisten2.Channel.String())

	// Test with dialects that don't support LISTEN/NOTIFY
	notSupportDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsListenNotify()
	})
	for _, d := range notSupportDialects.Dialects {
		_, err := parser.ParseSQL(d, "UNLISTEN test1")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Expected: an SQL statement")
	}
}

// TestParseNotifyChannel verifies NOTIFY statement parsing.
// Reference: tests/sqlparser_common.rs:14839
func TestParseNotifyChannel(t *testing.T) {
	// Test with dialects that support LISTEN/NOTIFY
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsListenNotify()
	})

	// Basic NOTIFY
	stmts := dialects.ParseSQL(t, "NOTIFY test1")
	require.Len(t, stmts, 1)

	notify, ok := stmts[0].(*statement.Notify)
	require.True(t, ok, "Expected Notify statement, got %T", stmts[0])
	assert.Equal(t, "test1", notify.Channel.String())
	assert.Nil(t, notify.Payload)

	// NOTIFY with payload
	stmts2 := dialects.ParseSQL(t, "NOTIFY test1, 'this is a test notification'")
	require.Len(t, stmts2, 1)

	notify2, ok := stmts2[0].(*statement.Notify)
	require.True(t, ok, "Expected Notify statement, got %T", stmts2[0])
	assert.Equal(t, "test1", notify2.Channel.String())
	require.NotNil(t, notify2.Payload)
	assert.Equal(t, "this is a test notification", *notify2.Payload)

	// Test with dialects that don't support LISTEN/NOTIFY
	notSupportDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsListenNotify()
	})
	testStatements := []string{
		"NOTIFY test1",
		"NOTIFY test1, 'this is a test notification'",
	}
	for _, sql := range testStatements {
		for _, d := range notSupportDialects.Dialects {
			_, err := parser.ParseSQL(d, sql)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Expected: an SQL statement")
		}
	}
}

// TestParseLoadData verifies LOAD DATA statement parsing.
// Reference: tests/sqlparser_common.rs:14887
func TestParseLoadData(t *testing.T) {
	// Test with dialects that support LOAD DATA
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsLoadData()
	})

	sql := "LOAD DATA INPATH '/local/path/to/data.txt' INTO TABLE test.my_table"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	loadData, ok := stmts[0].(*statement.LoadData)
	require.True(t, ok, "Expected LoadData statement, got %T", stmts[0])
	assert.False(t, loadData.Local)
	assert.Equal(t, "/local/path/to/data.txt", loadData.Inpath)
	assert.False(t, loadData.Overwrite)
	assert.Equal(t, "test.my_table", loadData.TableName.String())

	// Test with OVERWRITE
	sql2 := "LOAD DATA INPATH '/local/path/to/data.txt' OVERWRITE INTO TABLE my_table"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)

	loadData2, ok := stmts2[0].(*statement.LoadData)
	require.True(t, ok, "Expected LoadData statement, got %T", stmts2[0])
	assert.True(t, loadData2.Overwrite)
}

// TestParseBangNot verifies the bang (!) NOT operator parsing.
// Reference: tests/sqlparser_common.rs:15149
func TestParseBangNot(t *testing.T) {
	// Test with dialects that support bang not operator
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsBangNotOperator()
	})

	sql := "SELECT !a, !(b > 3)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Test that unsupported dialects fail
	notSupportDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsBangNotOperator()
	})
	testCases := []string{
		"SELECT !a",
		"SELECT !a b",
		"SELECT !a as b",
	}
	for _, sql := range testCases {
		for _, d := range notSupportDialects.Dialects {
			_, err := parser.ParseSQL(d, sql)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "Expected: an expression")
		}
	}
}

// TestParseFactorialOperator verifies the factorial (!) operator parsing.
// Reference: tests/sqlparser_common.rs:15197
func TestParseFactorialOperator(t *testing.T) {
	// Test with dialects that support factorial operator
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsFactorialOperator()
	})

	sql := "SELECT a!, (b + c)!"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Test that unsupported dialects fail with factorial syntax
	notSupportDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsFactorialOperator() && !d.SupportsBangNotOperator()
	})
	testCases := []string{
		"SELECT a!",
		"SELECT a ! b",
		"SELECT a ! as b",
	}
	for _, sql := range testCases {
		for _, d := range notSupportDialects.Dialects {
			_, err := parser.ParseSQL(d, sql)
			require.Error(t, err)
		}
	}
}

// TestParseComments verifies COMMENT statement parsing.
// Reference: tests/sqlparser_common.rs:15261
func TestParseComments(t *testing.T) {
	// Test with dialects that support COMMENT ON
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsCommentOn()
	})

	stmts := dialects.ParseSQL(t, "COMMENT ON COLUMN tab.name IS 'comment'")
	require.Len(t, stmts, 1)

	comment, ok := stmts[0].(*statement.Comment)
	require.True(t, ok, "Expected Comment statement, got %T", stmts[0])
	require.NotNil(t, comment.Comment)
	assert.Equal(t, "comment", *comment.Comment)
	assert.Equal(t, "tab.name", comment.ObjectName.String())
}

// TestParseCreateTableSelect verifies CREATE TABLE ... SELECT parsing.
// Reference: tests/sqlparser_common.rs:15358
func TestParseCreateTableSelect(t *testing.T) {
	// Test with dialects that support CREATE TABLE SELECT
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsCreateTableSelect()
	})

	sql1 := "CREATE TABLE foo (baz INT) SELECT bar"
	expected1 := "CREATE TABLE foo (baz INT) AS SELECT bar"
	dialects.OneStatementParsesTo(t, sql1, expected1)

	sql2 := "CREATE TABLE foo (baz INT, name STRING) SELECT bar, oth_name FROM test.table_a"
	expected2 := "CREATE TABLE foo (baz INT, name STRING) AS SELECT bar, oth_name FROM test.table_a"
	dialects.OneStatementParsesTo(t, sql2, expected2)
}

// TestReservedKeywordsForIdentifiers verifies reserved keywords behavior.
// Reference: tests/sqlparser_common.rs:15379
func TestReservedKeywordsForIdentifiers(t *testing.T) {
	// Test with dialects that reserve INTERVAL keyword
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.IsReservedForIdentifier(token.INTERVAL)
	})

	sql := "SELECT MAX(interval) FROM tbl"
	for _, d := range dialects.Dialects {
		_, err := parser.ParseSQL(d, sql)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Expected: an expression")
	}

	// Test with dialects that don't reserve INTERVAL keyword
	notReservedDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.IsReservedForIdentifier(token.INTERVAL)
	})

	sql2 := "SELECT MAX(interval) FROM tbl"
	notReservedDialects.ParseSQL(t, sql2)
}

// TestKeywordsAsColumnNamesAfterDot verifies keywords as column names after dot.
// Reference: tests/sqlparser_common.rs:15397
func TestKeywordsAsColumnNamesAfterDot(t *testing.T) {
	keywords := []string{
		"interval",
		"case",
		"cast",
		"extract",
		"trim",
		"substring",
		"left",
		"right",
	}

	dialects := utils.NewTestedDialects()

	for _, kw := range keywords {
		sql := fmt.Sprintf("SELECT T.%s FROM T", kw)
		dialects.VerifiedStmt(t, sql)

		sql2 := fmt.Sprintf("SELECT SUM(x) OVER (PARTITION BY T.%s ORDER BY T.id) FROM T", kw)
		dialects.VerifiedStmt(t, sql2)

		sql3 := fmt.Sprintf("SELECT T.%s, S.%s FROM T, S WHERE T.%s = S.%s", kw, kw, kw, kw)
		dialects.VerifiedStmt(t, sql3)
	}
}

// TestParseCreateTableWithBitTypes verifies CREATE TABLE with BIT types.
// Reference: tests/sqlparser_common.rs:15442
func TestParseCreateTableWithBitTypes(t *testing.T) {
	sql := "CREATE TABLE t (a BIT, b BIT VARYING, c BIT(42), d BIT VARYING(43))"
	stmts := utils.NewTestedDialects().ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createTable, ok := stmts[0].(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmts[0])
	require.Len(t, createTable.Columns, 4)
}

// TestParseCompositeAccessExpr verifies composite field access expressions.
// Reference: tests/sqlparser_common.rs:15461
func TestParseCompositeAccessExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Simple composite access
	stmts := dialects.ParseSQL(t, "SELECT f(a).b FROM t")
	require.Len(t, stmts, 1)

	// Nested composite access
	stmts2 := dialects.ParseSQL(t, "SELECT f(a).b.c FROM t")
	require.Len(t, stmts2, 1)

	// Composite access in WHERE clause
	stmts3 := dialects.ParseSQL(t, "SELECT * FROM t WHERE f(a).b IS NOT NULL")
	require.Len(t, stmts3, 1)
}

// TestParseCreateTableWithEnumTypes verifies CREATE TABLE with ENUM types.
// Reference: tests/sqlparser_common.rs:15585
func TestParseCreateTableWithEnumTypes(t *testing.T) {
	sql := "CREATE TABLE t0 (foo ENUM8('a' = 1, 'b' = 2), bar ENUM16('a' = 1, 'b' = 2), baz ENUM('a', 'b'))"
	stmts := utils.NewTestedDialects().ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createTable, ok := stmts[0].(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmts[0])
	assert.Equal(t, "t0", createTable.Name.String())
	require.Len(t, createTable.Columns, 3)
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

// TestOverflow verifies overflow handling with deeply nested expressions.
// Reference: tests/sqlparser_common.rs:15683
func TestOverflow(t *testing.T) {
	expr := strings.Repeat("1 + ", 999) + "1"
	sql := fmt.Sprintf("SELECT %s", expr)

	dialect := generic.NewGenericDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// TestParseDeeplyNestedBooleanExprDoesNotStackoverflow verifies deep nesting doesn't overflow.
// Reference: tests/sqlparser_common.rs:15695
func TestParseDeeplyNestedBooleanExprDoesNotStackoverflow(t *testing.T) {
	var buildNestedExpr func(depth int) string
	buildNestedExpr = func(depth int) string {
		if depth == 0 {
			return "x = 1"
		}
		return fmt.Sprintf("(%s OR %s AND (%s))", buildNestedExpr(0), buildNestedExpr(0), buildNestedExpr(depth-1))
	}

	depth := 200
	whereClause := buildNestedExpr(depth)
	sql := fmt.Sprintf("SELECT pk FROM tab0 WHERE %s", whereClause)

	// Note: This may need recursion limit adjustment in the parser
	dialect := generic.NewGenericDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err, "Parsing deeply nested boolean expression should not overflow")
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// TestParseSelectWithoutProjection verifies SELECT without projection parsing.
// Reference: tests/sqlparser_common.rs:15722
func TestParseSelectWithoutProjection(t *testing.T) {
	// Test with dialects that support empty projections
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsEmptyProjections()
	})

	dialects.VerifiedStmt(t, "SELECT FROM users")
}

// TestParseUpdateFromBeforeSelect verifies UPDATE ... FROM before SELECT parsing.
// Reference: tests/sqlparser_common.rs:15728
func TestParseUpdateFromBeforeSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "UPDATE t1 FROM (SELECT name, id FROM t1 GROUP BY id) AS t2 SET name = t2.name WHERE t1.id = t2.id"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "UPDATE t1 FROM U, (SELECT id FROM V) AS W SET a = b WHERE 1 = 1"
	dialects.VerifiedStmt(t, sql2)
}
