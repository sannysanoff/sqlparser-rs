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
// This file contains tests 241-260 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseCreateOrReplaceMaterializedView verifies CREATE OR REPLACE MATERIALIZED VIEW parsing.
// Reference: tests/sqlparser_common.rs:8507
func TestParseCreateOrReplaceMaterializedView(t *testing.T) {
	// Supported in BigQuery (Beta) and Snowflake
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			bigquery.NewBigQueryDialect(),
			snowflake.NewSnowflakeDialect(),
		},
	}

	sql := "CREATE OR REPLACE MATERIALIZED VIEW v AS SELECT 1"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createView, ok := stmts[0].(*statement.CreateView)
	require.True(t, ok, "Expected CreateView statement, got %T", stmts[0])

	// Verify OR REPLACE is set
	assert.True(t, createView.OrReplace, "Expected or_replace to be true")

	// Verify MATERIALIZED is set
	assert.True(t, createView.Materialized, "Expected materialized to be true")

	// Verify name
	assert.Equal(t, "v", createView.Name.String())

	// Verify columns is empty
	assert.Empty(t, createView.Columns)

	// Verify query is present
	require.NotNil(t, createView.Query, "Expected query to be present")

	// Verify or_alter is false
	assert.False(t, createView.OrAlter, "Expected or_alter to be false")

	// Verify cluster_by is empty
	assert.Empty(t, createView.ClusterBy)

	// Verify comment is none
	assert.Nil(t, createView.Comment, "Expected comment to be nil")

	// Verify with_no_schema_binding (late_binding) is false
	assert.False(t, createView.WithNoSchemaBinding, "Expected with_no_schema_binding to be false")

	// Verify if_not_exists is false
	assert.False(t, createView.IfNotExists, "Expected if_not_exists to be false")

	// Verify temporary is false
	assert.False(t, createView.Temporary, "Expected temporary to be false")

	// Verify to is none
	assert.Nil(t, createView.To, "Expected to to be nil")

	// Verify params is none
	assert.Nil(t, createView.Params, "Expected params to be nil")
}

// TestParseCreateMaterializedView verifies CREATE MATERIALIZED VIEW parsing.
// Reference: tests/sqlparser_common.rs:8553
func TestParseCreateMaterializedView(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "CREATE MATERIALIZED VIEW myschema.myview AS SELECT foo FROM bar"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createView, ok := stmts[0].(*statement.CreateView)
	require.True(t, ok, "Expected CreateView statement, got %T", stmts[0])

	// Verify or_alter is false
	assert.False(t, createView.OrAlter, "Expected or_alter to be false")

	// Verify name
	assert.Equal(t, "myschema.myview", createView.Name.String())

	// Verify columns is empty
	assert.Empty(t, createView.Columns)

	// Verify or_replace is false
	assert.False(t, createView.OrReplace, "Expected or_replace to be false")

	// Verify query is present
	require.NotNil(t, createView.Query, "Expected query to be present")

	// Verify materialized is true
	assert.True(t, createView.Materialized, "Expected materialized to be true")

	// Verify options is None
	assert.Nil(t, createView.Options, "Expected options to be nil")

	// Verify cluster_by is empty
	assert.Empty(t, createView.ClusterBy)

	// Verify comment is none
	assert.Nil(t, createView.Comment, "Expected comment to be nil")

	// Verify with_no_schema_binding (late_binding) is false
	assert.False(t, createView.WithNoSchemaBinding, "Expected with_no_schema_binding to be false")

	// Verify if_not_exists is false
	assert.False(t, createView.IfNotExists, "Expected if_not_exists to be false")

	// Verify temporary is false
	assert.False(t, createView.Temporary, "Expected temporary to be false")

	// Verify to is none
	assert.Nil(t, createView.To, "Expected to to be nil")

	// Verify params is none
	assert.Nil(t, createView.Params, "Expected params to be nil")
}

// TestParseCreateMaterializedViewWithClusterBy verifies CREATE MATERIALIZED VIEW with CLUSTER BY parsing.
// Reference: tests/sqlparser_common.rs:8595
func TestParseCreateMaterializedViewWithClusterBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "CREATE MATERIALIZED VIEW myschema.myview CLUSTER BY (foo) AS SELECT foo FROM bar"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createView, ok := stmts[0].(*statement.CreateView)
	require.True(t, ok, "Expected CreateView statement, got %T", stmts[0])

	// Verify or_alter is false
	assert.False(t, createView.OrAlter, "Expected or_alter to be false")

	// Verify name
	assert.Equal(t, "myschema.myview", createView.Name.String())

	// Verify or_replace is false
	assert.False(t, createView.OrReplace, "Expected or_replace to be false")

	// Verify columns is empty
	assert.Empty(t, createView.Columns)

	// Verify query is present
	require.NotNil(t, createView.Query, "Expected query to be present")

	// Verify materialized is true
	assert.True(t, createView.Materialized, "Expected materialized to be true")

	// Verify options is None
	assert.Nil(t, createView.Options, "Expected options to be nil")

	// Verify cluster_by contains "foo"
	require.Equal(t, 1, len(createView.ClusterBy))
	assert.Equal(t, "foo", createView.ClusterBy[0].String())

	// Verify comment is none
	assert.Nil(t, createView.Comment, "Expected comment to be nil")

	// Verify with_no_schema_binding (late_binding) is false
	assert.False(t, createView.WithNoSchemaBinding, "Expected with_no_schema_binding to be false")

	// Verify if_not_exists is false
	assert.False(t, createView.IfNotExists, "Expected if_not_exists to be false")

	// Verify temporary is false
	assert.False(t, createView.Temporary, "Expected temporary to be false")

	// Verify to is none
	assert.Nil(t, createView.To, "Expected to to be nil")

	// Verify params is none
	assert.Nil(t, createView.Params, "Expected params to be nil")
}

// TestParseDropTable verifies DROP TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:8637
func TestParseDropTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test basic DROP TABLE
	sql := "DROP TABLE foo"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	dropStmt, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])

	// Verify if_exists is false
	assert.False(t, dropStmt.IfExists, "Expected if_exists to be false")

	// Verify object_type is Table
	assert.Equal(t, "TABLE", dropStmt.ObjectType.String())

	// Verify names
	require.Equal(t, 1, len(dropStmt.Names))
	assert.Equal(t, "foo", dropStmt.Names[0].String())

	// Verify cascade is false
	assert.False(t, dropStmt.Cascade, "Expected cascade to be false")

	// Verify temporary is false
	assert.False(t, dropStmt.Temporary, "Expected temporary to be false")

	// Test DROP TABLE IF EXISTS with multiple tables and CASCADE
	sql2 := "DROP TABLE IF EXISTS foo, bar CASCADE"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)

	dropStmt2, ok := stmts2[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts2[0])

	// Verify if_exists is true
	assert.True(t, dropStmt2.IfExists, "Expected if_exists to be true")

	// Verify object_type is Table
	assert.Equal(t, "TABLE", dropStmt2.ObjectType.String())

	// Verify names
	require.Equal(t, 2, len(dropStmt2.Names))
	assert.Equal(t, "foo", dropStmt2.Names[0].String())
	assert.Equal(t, "bar", dropStmt2.Names[1].String())

	// Verify cascade is true
	assert.True(t, dropStmt2.Cascade, "Expected cascade to be true")

	// Verify temporary is false
	assert.False(t, dropStmt2.Temporary, "Expected temporary to be false")

	// Test DROP TABLE without table name should fail
	sql3 := "DROP TABLE"
	_, err := parser.ParseSQL(dialects.Dialects[0], sql3)
	require.Error(t, err, "Expected error for DROP TABLE without table name")
	assert.Contains(t, err.Error(), "identifier", "Error should mention identifier")

	// Test DROP TABLE with both CASCADE and RESTRICT should fail
	sql4 := "DROP TABLE IF EXISTS foo, bar CASCADE RESTRICT"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql4)
	require.Error(t, err, "Expected error for both CASCADE and RESTRICT")
	assert.Contains(t, err.Error(), "CASCADE", "Error should mention CASCADE")
	assert.Contains(t, err.Error(), "RESTRICT", "Error should mention RESTRICT")
}

// TestParseDropView verifies DROP VIEW statement parsing.
// Reference: tests/sqlparser_common.rs:8697
func TestParseDropView(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test DROP VIEW
	sql := "DROP VIEW myschema.myview"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	dropStmt, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])

	// Verify names
	require.Equal(t, 1, len(dropStmt.Names))
	assert.Equal(t, "myschema.myview", dropStmt.Names[0].String())

	// Verify object_type is View
	assert.Equal(t, "VIEW", dropStmt.ObjectType.String())

	// Test DROP MATERIALIZED VIEW
	sql2 := "DROP MATERIALIZED VIEW a.b.c"
	dialects.VerifiedStmt(t, sql2)

	// Test DROP MATERIALIZED VIEW IF EXISTS
	sql3 := "DROP MATERIALIZED VIEW IF EXISTS a.b.c"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseDropUser verifies DROP USER statement parsing.
// Reference: tests/sqlparser_common.rs:8717
func TestParseDropUser(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test DROP USER
	sql := "DROP USER u1"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	dropStmt, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])

	// Verify names
	require.Equal(t, 1, len(dropStmt.Names))
	assert.Equal(t, "u1", dropStmt.Names[0].String())

	// Verify object_type is User
	assert.Equal(t, "USER", dropStmt.ObjectType.String())

	// Test DROP USER IF EXISTS
	sql2 := "DROP USER IF EXISTS u1"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseInvalidSubqueryWithoutParens verifies that subqueries without parentheses fail.
// Reference: tests/sqlparser_common.rs:8735
func TestParseInvalidSubqueryWithoutParens(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT SELECT 1 FROM bar WHERE 1=1 FROM baz"
	_, err := parser.ParseSQL(dialects.Dialects[0], sql)
	require.Error(t, err, "Expected error for subquery without parens")
	assert.Contains(t, err.Error(), "end of statement", "Error should mention end of statement")
}

// TestParseOffset verifies OFFSET clause parsing.
// Reference: tests/sqlparser_common.rs:8745
func TestParseOffset(t *testing.T) {
	// Dialects that support `OFFSET` as column identifiers don't support this syntax
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			generic.NewGenericDialect(),
			postgresql.NewPostgreSqlDialect(),
			mssql.NewMsSqlDialect(),
			snowflake.NewSnowflakeDialect(),
			duckdb.NewDuckDbDialect(),
			redshift.NewRedshiftSqlDialect(),
			sqlite.NewSQLiteDialect(),
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

// TestParseStartTransaction verifies START TRANSACTION statement parsing.
// Reference: tests/sqlparser_common.rs:9040
func TestParseStartTransaction(t *testing.T) {
	// Excluding BigQuery and Snowflake as they don't support this syntax
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			generic.NewGenericDialect(),
			postgresql.NewPostgreSqlDialect(),
			mssql.NewMsSqlDialect(),
			sqlite.NewSQLiteDialect(),
			mysql.NewMySqlDialect(),
		},
	}

	// Test START TRANSACTION with modes
	sql1 := "START TRANSACTION READ ONLY, READ WRITE, ISOLATION LEVEL SERIALIZABLE"
	stmts := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts, 1)

	startTxn, ok := stmts[0].(*statement.StartTransaction)
	require.True(t, ok, "Expected StartTransaction statement, got %T", stmts[0])

	// Verify modes
	require.Equal(t, 3, len(startTxn.Modes))

	// Test START TRANSACTION (without modes)
	sql2 := "START TRANSACTION"
	dialects.VerifiedStmt(t, sql2)

	// Test BEGIN
	sql3 := "BEGIN"
	dialects.VerifiedStmt(t, sql3)

	// Test BEGIN WORK
	sql4 := "BEGIN WORK"
	dialects.VerifiedStmt(t, sql4)

	// Test BEGIN TRANSACTION
	sql5 := "BEGIN TRANSACTION"
	dialects.VerifiedStmt(t, sql5)

	// Test isolation levels
	dialects.VerifiedStmt(t, "START TRANSACTION ISOLATION LEVEL READ UNCOMMITTED")
	dialects.VerifiedStmt(t, "START TRANSACTION ISOLATION LEVEL READ COMMITTED")
	dialects.VerifiedStmt(t, "START TRANSACTION ISOLATION LEVEL REPEATABLE READ")
	dialects.VerifiedStmt(t, "START TRANSACTION ISOLATION LEVEL SERIALIZABLE")

	// Test multiple statements separated by semicolon
	sql6 := "START TRANSACTION; SELECT 1"
	stmts6, err := parser.ParseSQL(dialects.Dialects[0], sql6)
	require.NoError(t, err, "Failed to parse multiple statements")
	require.Equal(t, 2, len(stmts6))

	// Test invalid: ISOLATION LEVEL BAD
	sql7 := "START TRANSACTION ISOLATION LEVEL BAD"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql7)
	require.Error(t, err, "Expected error for invalid isolation level")

	// Test invalid: BAD mode
	sql8 := "START TRANSACTION BAD"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql8)
	require.Error(t, err, "Expected error for invalid transaction mode")

	// Test invalid: trailing comma
	sql9 := "START TRANSACTION READ ONLY,"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql9)
	require.Error(t, err, "Expected error for trailing comma")
}

// TestParseStartTransactionMssql verifies MS-SQL specific START TRANSACTION syntax.
// Reference: tests/sqlparser_common.rs:9040 (MS-SQL specific)
func TestParseStartTransactionMssql(t *testing.T) {
	dialect := mssql.NewMsSqlDialect()

	// Test BEGIN TRY
	sql1 := "BEGIN TRY"
	stmts, err := parser.ParseSQL(dialect, sql1)
	require.NoError(t, err, "Failed to parse BEGIN TRY")
	require.Len(t, stmts, 1)

	// Test BEGIN CATCH
	sql2 := "BEGIN CATCH"
	stmts2, err := parser.ParseSQL(dialect, sql2)
	require.NoError(t, err, "Failed to parse BEGIN CATCH")
	require.Len(t, stmts2, 1)

	// Test complex MS-SQL transaction block
	sql3 := `
		BEGIN TRY;
			SELECT 1/0;
		END TRY;
		BEGIN CATCH;
			EXECUTE foo;
		END CATCH;
	`
	stmts3, err := parser.ParseSQL(dialect, sql3)
	require.NoError(t, err, "Failed to parse MS-SQL transaction block")
	require.Equal(t, 4, len(stmts3))
}

// TestParseSetTransaction verifies SET TRANSACTION statement parsing.
// Reference: tests/sqlparser_common.rs:9140
func TestParseSetTransaction(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SET TRANSACTION READ ONLY, READ WRITE, ISOLATION LEVEL SERIALIZABLE"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	setStmt, ok := stmts[0].(*statement.Set)
	require.True(t, ok, "Expected Set statement, got %T", stmts[0])

	// Verify it's a SET TRANSACTION
	require.NotNil(t, setStmt, "Expected Set statement")
}

// TestParseSetVariable verifies SET variable statement parsing.
// Reference: tests/sqlparser_common.rs:9166
func TestParseSetVariable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test SET SOMETHING = '1'
	sql1 := "SET SOMETHING = '1'"
	stmts := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts, 1)

	_, ok := stmts[0].(*statement.Set)
	require.True(t, ok, "Expected Set statement, got %T", stmts[0])

	// Test SET GLOBAL VARIABLE = 'Value'
	sql2 := "SET GLOBAL VARIABLE = 'Value'"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)

	_, ok = stmts2[0].(*statement.Set)
	require.True(t, ok, "Expected Set statement, got %T", stmts2[0])

	// Test parenthesized assignments
	sql3 := "SET (a, b, c) = (1, 2, 3)"
	dialects.VerifiedStmt(t, sql3)

	// Test SET TO syntax
	sql4 := "SET SOMETHING TO '1'"
	canonical4 := "SET SOMETHING = '1'"
	dialects.OneStatementParsesTo(t, sql4, canonical4)
}

// TestParseSetVariableSubquery verifies SET with subquery expression parsing.
// Reference: tests/sqlparser_common.rs:9166 (subquery tests)
func TestParseSetVariableSubquery(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test subquery expressions
	testCases := []struct {
		sql      string
		expected string
	}{
		{
			sql:      "SET (a) = (SELECT 22 FROM tbl1)",
			expected: "SET (a) = ((SELECT 22 FROM tbl1))",
		},
		{
			sql:      "SET (a) = (SELECT 22 FROM tbl1, (SELECT 1 FROM tbl2))",
			expected: "SET (a) = ((SELECT 22 FROM tbl1, (SELECT 1 FROM tbl2)))",
		},
		{
			sql:      "SET (a) = ((SELECT 22 FROM tbl1, (SELECT 1 FROM tbl2)))",
			expected: "SET (a) = ((SELECT 22 FROM tbl1, (SELECT 1 FROM tbl2)))",
		},
	}

	for _, tc := range testCases {
		dialects.OneStatementParsesTo(t, tc.sql, tc.expected)
	}
}

// TestParseSetVariableErrors verifies SET variable error cases.
// Reference: tests/sqlparser_common.rs:9166 (error tests)
func TestParseSetVariableErrors(t *testing.T) {
	dialects := utils.NewTestedDialects()

	errorCases := []struct {
		sql   string
		error string
	}{
		{
			sql:   "SET (a, b, c) = (1, 2, 3",
			error: ")",
		},
		{
			sql:   "SET (a, b, c) = 1, 2, 3",
			error: "(",
		},
	}

	for _, tc := range errorCases {
		t.Run(tc.sql, func(t *testing.T) {
			_, err := parser.ParseSQL(dialects.Dialects[0], tc.sql)
			require.Error(t, err, "Expected error for: %s", tc.sql)
			assert.Contains(t, err.Error(), tc.error, "Error should contain: %s", tc.error)
		})
	}
}

// TestParseSetRoleAsVariable verifies SET role as variable parsing.
// Reference: tests/sqlparser_common.rs:9278
func TestParseSetRoleAsVariable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SET role = 'foobar'"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	setStmt, ok := stmts[0].(*statement.Set)
	require.True(t, ok, "Expected Set statement, got %T", stmts[0])
	require.NotNil(t, setStmt, "Expected Set statement")
}

// TestParseDoubleColonCastAtTimezone verifies double colon cast with AT TIME ZONE parsing.
// Reference: tests/sqlparser_common.rs:9301
func TestParseDoubleColonCastAtTimezone(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT '2001-01-01T00:00:00.000Z'::TIMESTAMP AT TIME ZONE 'Europe/Brussels' FROM t"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSetTimeZone verifies SET TIME ZONE statement parsing.
// Reference: tests/sqlparser_common.rs:9327
func TestParseSetTimeZone(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test SET TIMEZONE = 'UTC'
	sql1 := "SET TIMEZONE = 'UTC'"
	stmts := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts, 1)

	setStmt, ok := stmts[0].(*statement.Set)
	require.True(t, ok, "Expected Set statement, got %T", stmts[0])
	require.NotNil(t, setStmt, "Expected Set statement")

	// Test SET TIME ZONE TO 'UTC' should be equivalent to SET TIMEZONE = 'UTC'
	sql2 := "SET TIME ZONE TO 'UTC'"
	canonical2 := "SET TIMEZONE = 'UTC'"
	dialects.OneStatementParsesTo(t, sql2, canonical2)
}

// TestParseCommit verifies COMMIT statement parsing.
// Reference: tests/sqlparser_common.rs:9352
func TestParseCommit(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test basic COMMIT
	sql1 := "COMMIT"
	stmts := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts, 1)

	commit, ok := stmts[0].(*statement.Commit)
	require.True(t, ok, "Expected Commit statement, got %T", stmts[0])
	assert.False(t, commit.Chain, "Expected chain to be false")

	// Test COMMIT AND CHAIN
	sql2 := "COMMIT AND CHAIN"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)

	commit2, ok := stmts2[0].(*statement.Commit)
	require.True(t, ok, "Expected Commit statement, got %T", stmts2[0])
	assert.True(t, commit2.Chain, "Expected chain to be true")

	// Test variations that should parse to COMMIT
	dialects.OneStatementParsesTo(t, "COMMIT AND NO CHAIN", "COMMIT")
	dialects.OneStatementParsesTo(t, "COMMIT WORK AND NO CHAIN", "COMMIT")
	dialects.OneStatementParsesTo(t, "COMMIT TRANSACTION AND NO CHAIN", "COMMIT")
	dialects.OneStatementParsesTo(t, "COMMIT WORK AND CHAIN", "COMMIT AND CHAIN")
	dialects.OneStatementParsesTo(t, "COMMIT TRANSACTION AND CHAIN", "COMMIT AND CHAIN")
	dialects.OneStatementParsesTo(t, "COMMIT WORK", "COMMIT")
	dialects.OneStatementParsesTo(t, "COMMIT TRANSACTION", "COMMIT")
}

// TestParseEnd verifies END statement parsing.
// Reference: tests/sqlparser_common.rs:9373
func TestParseEnd(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test END variations
	dialects.OneStatementParsesTo(t, "END AND NO CHAIN", "END")
	dialects.OneStatementParsesTo(t, "END WORK AND NO CHAIN", "END")
	dialects.OneStatementParsesTo(t, "END TRANSACTION AND NO CHAIN", "END")
	dialects.OneStatementParsesTo(t, "END WORK AND CHAIN", "END AND CHAIN")
	dialects.OneStatementParsesTo(t, "END TRANSACTION AND CHAIN", "END AND CHAIN")
	dialects.OneStatementParsesTo(t, "END WORK", "END")
	dialects.OneStatementParsesTo(t, "END TRANSACTION", "END")
}

// TestParseEndMssql verifies MS-SQL specific END syntax.
// Reference: tests/sqlparser_common.rs:9373 (MS-SQL specific)
func TestParseEndMssql(t *testing.T) {
	dialect := mssql.NewMsSqlDialect()

	// Test END TRY
	sql1 := "END TRY"
	stmts, err := parser.ParseSQL(dialect, sql1)
	require.NoError(t, err, "Failed to parse END TRY")
	require.Len(t, stmts, 1)

	// Test END CATCH
	sql2 := "END CATCH"
	stmts2, err := parser.ParseSQL(dialect, sql2)
	require.NoError(t, err, "Failed to parse END CATCH")
	require.Len(t, stmts2, 1)
}
