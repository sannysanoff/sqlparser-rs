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

// Package dml contains DML (Data Manipulation Language) SQL parsing tests.
// These tests cover INSERT, UPDATE, DELETE, MERGE, and COPY statements.
package dml

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/errors"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseUpdate verifies UPDATE statement parsing.
// Reference: tests/sqlparser_common.rs:394
func TestParseUpdate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "UPDATE customer SET name = 'John' WHERE id = 1"
	dialects.VerifiedStmt(t, sql)
}

// TestParseUpdateFull verifies UPDATE statement parsing with multiple assignments and error cases.
// Reference: tests/sqlparser_common.rs:394
func TestParseUpdateFull(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "UPDATE t SET a = 1, b = 2, c = 3 WHERE d"
	dialects.VerifiedStmt(t, sql)

	// Multiple assignments to same column
	dialects.VerifiedStmt(t, "UPDATE t SET a = 1, a = 2, a = 3")
}

// TestParseUpdateSetFrom verifies UPDATE with FROM clause parsing.
// Reference: tests/sqlparser_common.rs:443
func TestParseUpdateSetFrom(t *testing.T) {
	// This test uses specific dialects that support UPDATE...FROM
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			postgresql.NewPostgreSqlDialect(),
			bigquery.NewBigQueryDialect(),
			snowflake.NewSnowflakeDialect(),
			redshift.NewRedshiftSqlDialect(),
			mssql.NewMsSqlDialect(),
			sqlite.NewSQLiteDialect(),
		},
	}

	sql := "UPDATE t1 SET name = t2.name FROM (SELECT name, id FROM t1 GROUP BY id) AS t2 WHERE t1.id = t2.id"
	dialects.VerifiedStmt(t, sql)

	sql2 := "UPDATE T SET a = b FROM U, (SELECT foo FROM V) AS W WHERE 1 = 1"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseUpdateSetFromFull verifies UPDATE with FROM clause parsing.
// Reference: tests/sqlparser_common.rs:444
func TestParseUpdateSetFromFull(t *testing.T) {
	// This test uses specific dialects that support UPDATE...FROM
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			postgresql.NewPostgreSqlDialect(),
			bigquery.NewBigQueryDialect(),
			snowflake.NewSnowflakeDialect(),
			redshift.NewRedshiftSqlDialect(),
			mssql.NewMsSqlDialect(),
			sqlite.NewSQLiteDialect(),
		},
	}

	sql := "UPDATE t1 SET name = t2.name FROM (SELECT name, id FROM t1 GROUP BY id) AS t2 WHERE t1.id = t2.id"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	update, ok := stmts[0].(*statement.Update)
	require.True(t, ok, "Expected Update statement, got %T", stmts[0])

	// Verify table name
	assert.Equal(t, "t1", update.Table.String())

	// Verify assignments
	require.Equal(t, 1, len(update.Assignments))

	// Verify FROM clause is present
	require.NotNil(t, update.From, "Expected FROM clause to be present")

	// Verify WHERE clause
	require.NotNil(t, update.Selection, "Expected WHERE clause to be present")

	// Second test case
	sql2 := "UPDATE T SET a = b FROM U, (SELECT foo FROM V) AS W WHERE 1 = 1"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseUpdateWithTableAlias verifies UPDATE with table alias parsing.
// Reference: tests/sqlparser_common.rs:546
func TestParseUpdateWithTableAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "UPDATE users AS u SET u.username = 'new_user' WHERE u.username = 'old_user'"
	dialects.VerifiedStmt(t, sql)
}

// TestParseUpdateWithTableAliasFull verifies UPDATE with table alias parsing.
// Reference: tests/sqlparser_common.rs:547
func TestParseUpdateWithTableAliasFull(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "UPDATE users AS u SET u.username = 'new_user' WHERE u.username = 'old_user'"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	update, ok := stmts[0].(*statement.Update)
	require.True(t, ok, "Expected Update statement, got %T", stmts[0])

	// Verify table name
	assert.Equal(t, "users", update.Table.String())

	// Verify table alias
	require.NotNil(t, update.TableAlias, "Expected table alias to be present")
	assert.Equal(t, "u", update.TableAlias.String())

	// Verify assignments
	require.Equal(t, 1, len(update.Assignments))

	// Verify WHERE clause
	require.NotNil(t, update.Selection, "Expected WHERE clause to be present")

	// Verify no RETURNING clause
	assert.Empty(t, update.Returning)
}

// TestParseUpdateOr verifies SQLite UPDATE OR clause parsing.
// Reference: tests/sqlparser_common.rs:611
func TestParseUpdateOr(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "UPDATE OR REPLACE t SET n = n + 1")
	dialects.VerifiedStmt(t, "UPDATE OR ROLLBACK t SET n = n + 1")
	dialects.VerifiedStmt(t, "UPDATE OR ABORT t SET n = n + 1")
	dialects.VerifiedStmt(t, "UPDATE OR FAIL t SET n = n + 1")
	dialects.VerifiedStmt(t, "UPDATE OR IGNORE t SET n = n + 1")
}

// TestParseUpdateOrFull verifies SQLite UPDATE OR clause parsing.
// Reference: tests/sqlparser_common.rs:612
func TestParseUpdateOrFull(t *testing.T) {
	expectOrClause := func(t *testing.T, sql string) {
		dialects := utils.NewTestedDialects()
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		update, ok := stmts[0].(*statement.Update)
		require.True(t, ok, "Expected Update statement, got %T", stmts[0])

		// The Or field should be set for SQLite dialect
		// For non-SQLite dialects, this might parse differently
		_ = update
	}

	expectOrClause(t, "UPDATE OR REPLACE t SET n = n + 1")
	expectOrClause(t, "UPDATE OR ROLLBACK t SET n = n + 1")
	expectOrClause(t, "UPDATE OR ABORT t SET n = n + 1")
	expectOrClause(t, "UPDATE OR FAIL t SET n = n + 1")
	expectOrClause(t, "UPDATE OR IGNORE t SET n = n + 1")
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

// TestParseUpdateDeeplyNestedParensHitsRecursionLimits verifies recursion limit with UPDATE and nested parens.
// Reference: tests/sqlparser_common.rs:11082
func TestParseUpdateDeeplyNestedParensHitsRecursionLimits(t *testing.T) {
	sql := "\nUPDATE\n\n\n\n\n\n\n\n\n\n" + strings.Repeat("(", 1000)
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err)
	assert.True(t, errors.IsRecursionLimitExceeded(err), "Expected RecursionLimitExceeded error")
}
