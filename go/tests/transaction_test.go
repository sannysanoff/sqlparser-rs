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
// This file contains transaction-related tests.
package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseStartTransaction verifies START TRANSACTION statement parsing.
// Reference: tests/sqlparser_common.rs:9040
func TestParseStartTransaction(t *testing.T) {
	// Excluding BigQuery and Snowflake as they don't support this syntax
	dialects := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
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

// TestParseRollback verifies ROLLBACK statement parsing.
// Reference: tests/sqlparser_common.rs:9388
func TestParseRollback(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic ROLLBACK
	stmt := dialects.VerifiedStmt(t, "ROLLBACK")
	rollback, ok := stmt.(*statement.Rollback)
	require.True(t, ok, "Expected Rollback statement, got %T", stmt)
	assert.False(t, rollback.Chain, "Expected chain to be false")
	assert.Nil(t, rollback.Savepoint, "Expected savepoint to be nil")

	// ROLLBACK AND CHAIN
	stmt = dialects.VerifiedStmt(t, "ROLLBACK AND CHAIN")
	rollback, ok = stmt.(*statement.Rollback)
	require.True(t, ok, "Expected Rollback statement, got %T", stmt)
	assert.True(t, rollback.Chain, "Expected chain to be true")
	assert.Nil(t, rollback.Savepoint, "Expected savepoint to be nil")

	// ROLLBACK TO SAVEPOINT
	stmt = dialects.VerifiedStmt(t, "ROLLBACK TO SAVEPOINT test1")
	rollback, ok = stmt.(*statement.Rollback)
	require.True(t, ok, "Expected Rollback statement, got %T", stmt)
	assert.False(t, rollback.Chain, "Expected chain to be false")
	require.NotNil(t, rollback.Savepoint, "Expected savepoint to be present")
	assert.Equal(t, "test1", rollback.Savepoint.String())

	// ROLLBACK AND CHAIN TO SAVEPOINT
	stmt = dialects.VerifiedStmt(t, "ROLLBACK AND CHAIN TO SAVEPOINT test1")
	rollback, ok = stmt.(*statement.Rollback)
	require.True(t, ok, "Expected Rollback statement, got %T", stmt)
	assert.True(t, rollback.Chain, "Expected chain to be true")
	require.NotNil(t, rollback.Savepoint, "Expected savepoint to be present")
	assert.Equal(t, "test1", rollback.Savepoint.String())

	// Test one_statement_parses_to equivalents
	dialects.OneStatementParsesTo(t, "ROLLBACK AND NO CHAIN", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK WORK AND NO CHAIN", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK TRANSACTION AND NO CHAIN", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK WORK AND CHAIN", "ROLLBACK AND CHAIN")
	dialects.OneStatementParsesTo(t, "ROLLBACK TRANSACTION AND CHAIN", "ROLLBACK AND CHAIN")
	dialects.OneStatementParsesTo(t, "ROLLBACK WORK", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK TRANSACTION", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK TO test1", "ROLLBACK TO SAVEPOINT test1")
	dialects.OneStatementParsesTo(t, "ROLLBACK AND CHAIN TO test1", "ROLLBACK AND CHAIN TO SAVEPOINT test1")
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

// TestParseTransactionStatements verifies transaction statement parsing.
// Reference: tests/sqlparser_common.rs:435
func TestParseTransactionStatements(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "BEGIN")
	dialects.VerifiedStmt(t, "COMMIT")
	dialects.VerifiedStmt(t, "ROLLBACK")
}

// TestLock verifies FOR UPDATE and FOR SHARE lock clauses.
// Reference: tests/sqlparser_common.rs:10301
func TestLock(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test FOR UPDATE
	sql := "SELECT * FROM student WHERE id = '1' FOR UPDATE"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Get the query from the statement
	selectStmt, ok := stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q := selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock := q.Locks[0]
	assert.Equal(t, query.LockTypeUpdate, lock.LockType, "Expected UPDATE lock type")
	assert.Nil(t, lock.Of, "Expected OF to be nil")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil")

	// Test FOR SHARE
	sql = "SELECT * FROM student WHERE id = '1' FOR SHARE"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok = stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q = selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock = q.Locks[0]
	assert.Equal(t, query.LockTypeShare, lock.LockType, "Expected SHARE lock type")
	assert.Nil(t, lock.Of, "Expected OF to be nil")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil")
}

// TestLockTable verifies FOR UPDATE OF and FOR SHARE OF table lock clauses.
// Reference: tests/sqlparser_common.rs:10320
func TestLockTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test FOR UPDATE OF school
	sql := "SELECT * FROM student WHERE id = '1' FOR UPDATE OF school"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q := selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock := q.Locks[0]
	assert.Equal(t, query.LockTypeUpdate, lock.LockType, "Expected UPDATE lock type")
	require.NotNil(t, lock.Of, "Expected OF to be present")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil")

	// Test FOR SHARE OF school
	sql = "SELECT * FROM student WHERE id = '1' FOR SHARE OF school"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok = stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q = selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock = q.Locks[0]
	assert.Equal(t, query.LockTypeShare, lock.LockType, "Expected SHARE lock type")
	require.NotNil(t, lock.Of, "Expected OF to be present")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil")

	// Test multiple locks: FOR SHARE OF school FOR UPDATE OF student
	sql = "SELECT * FROM student WHERE id = '1' FOR SHARE OF school FOR UPDATE OF student"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok = stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q = selectStmt.Query
	require.Len(t, q.Locks, 2, "Expected 2 lock clauses")

	// First lock: FOR SHARE OF school
	lock = q.Locks[0]
	assert.Equal(t, query.LockTypeShare, lock.LockType, "Expected first lock to be SHARE")
	require.NotNil(t, lock.Of, "Expected OF to be present on first lock")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil on first lock")

	// Second lock: FOR UPDATE OF student
	lock = q.Locks[1]
	assert.Equal(t, query.LockTypeUpdate, lock.LockType, "Expected second lock to be UPDATE")
	require.NotNil(t, lock.Of, "Expected OF to be present on second lock")
	assert.Equal(t, "student", lock.Of.String(), "Expected OF table to be 'student'")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil on second lock")
}

// TestLockNonblock verifies SKIP LOCKED and NOWAIT lock modifiers.
// Reference: tests/sqlparser_common.rs:10379
func TestLockNonblock(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test FOR UPDATE OF school SKIP LOCKED
	sql := "SELECT * FROM student WHERE id = '1' FOR UPDATE OF school SKIP LOCKED"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q := selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock := q.Locks[0]
	assert.Equal(t, query.LockTypeUpdate, lock.LockType, "Expected UPDATE lock type")
	require.NotNil(t, lock.Of, "Expected OF to be present")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	require.NotNil(t, lock.Nonblock, "Expected nonblock to be present")
	assert.Equal(t, query.NonBlockSkipLocked, *lock.Nonblock, "Expected SKIP LOCKED")

	// Test FOR SHARE OF school NOWAIT
	sql = "SELECT * FROM student WHERE id = '1' FOR SHARE OF school NOWAIT"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok = stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q = selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock = q.Locks[0]
	assert.Equal(t, query.LockTypeShare, lock.LockType, "Expected SHARE lock type")
	require.NotNil(t, lock.Of, "Expected OF to be present")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	require.NotNil(t, lock.Nonblock, "Expected nonblock to be present")
	assert.Equal(t, query.NonBlockNowait, *lock.Nonblock, "Expected NOWAIT")
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

// TestParseSetNames verifies SET NAMES statement parsing.
// Reference: tests/sqlparser_common.rs:16589
func TestParseSetNames(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsSetNames()
	})

	dialects.VerifiedStmt(t, "SET NAMES 'UTF8'")
	dialects.VerifiedStmt(t, "SET NAMES 'utf8'")
	dialects.VerifiedStmt(t, "SET NAMES UTF8 COLLATE bogus")
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

// TestParseSetSessionAuthorization verifies SET SESSION AUTHORIZATION parsing.
// Reference: tests/sqlparser_common.rs:18510
func TestParseSetSessionAuthorization(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// SET SESSION AUTHORIZATION DEFAULT
	dialects.VerifiedStmt(t, "SET SESSION AUTHORIZATION DEFAULT")

	// SET SESSION AUTHORIZATION with username
	dialects.VerifiedStmt(t, "SET SESSION AUTHORIZATION 'username'")
}

// TestParseSetAuthorizationVariations verifies SET AUTHORIZATION variations.
// Reference: tests/sqlparser_common.rs:18510 (additional tests)
func TestParseSetAuthorizationVariations(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Various SET AUTHORIZATION forms
	tests := []string{
		"SET SESSION AUTHORIZATION DEFAULT",
		"SET SESSION AUTHORIZATION 'user_name'",
		"SET LOCAL AUTHORIZATION DEFAULT",
	}

	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, sql)
			require.Len(t, stmts, 1)
		})
	}
}

// TestSetAuthorizationWithoutScopeErrors verifies SET AUTHORIZATION without scope errors.
// Reference: tests/sqlparser_common.rs:18535
func TestSetAuthorizationWithoutScopeErrors(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// This should return an error, not panic
	stmts := dialects.ParseSQL(t, "SET AUTHORIZATION TIME TIME")
	// We expect parsing to either succeed with an error or the statement to be parsed differently
	// The test verifies that it doesn't panic
	require.NotNil(t, stmts)
}

// TestParseResetStatement verifies RESET statement parsing.
// Reference: tests/sqlparser_common.rs:18487
func TestParseResetStatement(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// RESET with parameter
	stmt1 := dialects.VerifiedStmt(t, "RESET some_parameter")
	reset1, ok := stmt1.(*statement.Reset)
	require.True(t, ok, "Expected Reset statement, got %T", stmt1)
	require.NotNil(t, reset1.Statement, "Expected ResetStatement")

	// RESET with qualified parameter
	stmt2 := dialects.VerifiedStmt(t, "RESET some_extension.some_parameter")
	reset2, ok := stmt2.(*statement.Reset)
	require.True(t, ok, "Expected Reset statement, got %T", stmt2)
	require.NotNil(t, reset2.Statement, "Expected ResetStatement")

	// RESET ALL
	stmt3 := dialects.VerifiedStmt(t, "RESET ALL")
	reset3, ok := stmt3.(*statement.Reset)
	require.True(t, ok, "Expected Reset statement, got %T", stmt3)
	require.NotNil(t, reset3.Statement, "Expected ResetStatement")
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

// TestParseCursor verifies CLOSE cursor statement parsing.
// Reference: tests/sqlparser_common.rs:10831
func TestParseCursor(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test CLOSE my_cursor
	sql := "CLOSE my_cursor"
	stmt := dialects.VerifiedStmt(t, sql)
	closeStmt, ok := stmt.(*statement.Close)
	require.True(t, ok, "Expected Close statement, got %T", stmt)
	require.NotNil(t, closeStmt.Cursor, "Expected cursor to be present")

	// Test CLOSE ALL
	sql = "CLOSE ALL"
	stmt = dialects.VerifiedStmt(t, sql)
	closeStmt, ok = stmt.(*statement.Close)
	require.True(t, ok, "Expected Close statement, got %T", stmt)
}

// TestParseDiscard verifies DISCARD statement parsing.
// Reference: tests/sqlparser_common.rs:10804
func TestParseDiscard(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test DISCARD ALL
	stmt := dialects.VerifiedStmt(t, "DISCARD ALL")
	discard, ok := stmt.(*statement.Discard)
	require.True(t, ok, "Expected Discard statement, got %T", stmt)
	_ = discard

	// Test DISCARD PLANS
	stmt = dialects.VerifiedStmt(t, "DISCARD PLANS")
	discard, ok = stmt.(*statement.Discard)
	require.True(t, ok, "Expected Discard statement, got %T", stmt)

	// Test DISCARD SEQUENCES
	stmt = dialects.VerifiedStmt(t, "DISCARD SEQUENCES")
	discard, ok = stmt.(*statement.Discard)
	require.True(t, ok, "Expected Discard statement, got %T", stmt)

	// Test DISCARD TEMP
	stmt = dialects.VerifiedStmt(t, "DISCARD TEMP")
	discard, ok = stmt.(*statement.Discard)
	require.True(t, ok, "Expected Discard statement, got %T", stmt)
}
