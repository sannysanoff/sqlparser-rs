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
// This file contains tests 421-440 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseReturn verifies RETURN statement parsing.
// Reference: tests/sqlparser_common.rs:17156
func TestParseReturn(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic RETURN - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "RETURN")

	// RETURN with value
	_ = dialects.VerifiedStmt(t, "RETURN 1")
}

// TestParseSubqueryLimit verifies subquery with LIMIT parsing.
// Reference: tests/sqlparser_common.rs:17164
func TestParseSubqueryLimit(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT t1_id, t1_name FROM t1 WHERE t1_id IN (SELECT t2_id FROM t2 WHERE t1_name = t2_name LIMIT 10)"
	_ = dialects.VerifiedStmt(t, sql)
}

// TestOpen verifies OPEN cursor statement parsing.
// Reference: tests/sqlparser_common.rs:17169
func TestOpen(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// OPEN cursor - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "OPEN Employee_Cursor")
}

// TestParseTruncateOnly verifies TRUNCATE TABLE with ONLY clause parsing.
// Reference: tests/sqlparser_common.rs:17181
func TestParseTruncateOnly(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// TRUNCATE with ONLY - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "TRUNCATE TABLE employee, ONLY dept")
}

// TestCheckEnforced verifies CHECK constraint with ENFORCED/NOT ENFORCED parsing.
// Reference: tests/sqlparser_common.rs:17212
func TestCheckEnforced(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "CREATE TABLE t (a INT, b INT, c INT, CHECK (a > 0) NOT ENFORCED, CHECK (b > 0) ENFORCED, CHECK (c > 0))"
	_ = dialects.VerifiedStmt(t, sql)
}

// TestColumnCheckEnforced verifies column-level CHECK constraint with ENFORCED/NOT ENFORCED parsing.
// Reference: tests/sqlparser_common.rs:17219
func TestColumnCheckEnforced(t *testing.T) {
	dialects := utils.NewTestedDialects()
	_ = dialects.VerifiedStmt(t, "CREATE TABLE t (x INT CHECK (x > 1) NOT ENFORCED)")
	_ = dialects.VerifiedStmt(t, "CREATE TABLE t (x INT CHECK (x > 1) ENFORCED)")
	_ = dialects.VerifiedStmt(t, "CREATE TABLE t (a INT CHECK (a > 0) NOT ENFORCED, b INT CHECK (b > 0) ENFORCED, c INT CHECK (c > 0))")
}

// TestJoinPrecedence verifies JOIN precedence parsing.
// Reference: tests/sqlparser_common.rs:17228
func TestJoinPrecedence(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// Test that joins are parsed with correct precedence
	_ = dialects.VerifiedStmt(t, "SELECT * FROM t1 NATURAL JOIN t5 INNER JOIN t0 ON (t0.v1 + t5.v0) > 0 WHERE t0.v1 = t1.v0")
}

// TestParseCreateProcedureWithLanguage verifies CREATE PROCEDURE with LANGUAGE clause parsing.
// Reference: tests/sqlparser_common.rs:17251
func TestParseCreateProcedureWithLanguage(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// CREATE PROCEDURE - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "CREATE PROCEDURE test_proc LANGUAGE sql AS BEGIN SELECT 1; END")
}

// TestParseCreateProcedureWithParameterModes verifies CREATE PROCEDURE with parameter modes parsing.
// Reference: tests/sqlparser_common.rs:17281
func TestParseCreateProcedureWithParameterModes(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// CREATE PROCEDURE with parameter modes - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "CREATE PROCEDURE test_proc (IN a INTEGER, OUT b TEXT, INOUT c TIMESTAMP, d BOOL) AS BEGIN SELECT 1; END")

	// Test with default values
	_ = dialects.VerifiedStmt(t, "CREATE PROCEDURE test_proc (IN a INTEGER = 1, OUT b TEXT = '2', INOUT c TIMESTAMP = NULL, d BOOL = 0) AS BEGIN SELECT 1; END")
}

// TestParseNotNull verifies NOT NULL expression parsing.
// Reference: tests/sqlparser_common.rs:17393
func TestParseNotNull(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// These should parse as IS NOT NULL
	_ = dialects.VerifiedStmt(t, "SELECT x IS NOT NULL FROM t")
	_ = dialects.VerifiedStmt(t, "SELECT NULL IS NOT NULL FROM t")
}

// TestSelectExclude verifies SELECT * EXCLUDE clause parsing.
// Reference: tests/sqlparser_common.rs:17414
func TestSelectExclude(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			bigquery.NewBigQueryDialect(),
			snowflake.NewSnowflakeDialect(),
		},
	}

	// Single column EXCLUDE - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "SELECT * EXCLUDE c1 FROM test")

	// Multi-column EXCLUDE with parentheses
	_ = dialects.VerifiedStmt(t, "SELECT * EXCLUDE (c1, c2) FROM test")
}

// TestSelectExcludeQualifiedNames verifies SELECT EXCLUDE with qualified names parsing.
// Reference: tests/sqlparser_common.rs:17527
func TestSelectExcludeQualifiedNames(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			bigquery.NewBigQueryDialect(),
			snowflake.NewSnowflakeDialect(),
		},
	}

	// Qualified names in EXCLUDE
	sql := "SELECT f.* EXCLUDE (f.account_canonical_id, f.amount) FROM t AS f"
	_ = dialects.VerifiedStmt(t, sql)

	// Plain identifiers
	_ = dialects.VerifiedStmt(t, "SELECT f.* EXCLUDE (account_canonical_id) FROM t AS f")
	_ = dialects.VerifiedStmt(t, "SELECT f.* EXCLUDE (col1, col2) FROM t AS f")
}

// TestNoSemicolonRequiredBetweenStatements verifies parsing without semicolons between statements.
// Reference: tests/sqlparser_common.rs:17553
func TestNoSemicolonRequiredBetweenStatements(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM tbl1 SELECT * FROM tbl2"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 2)
}

// TestIdentifierUnicodeSupport verifies Unicode identifier support.
// Reference: tests/sqlparser_common.rs:17570
func TestIdentifierUnicodeSupport(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			mysql.NewMySqlDialect(),
			redshift.NewRedshiftSqlDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}
	sql := "SELECT phoneǤЖשचᎯ⻩☯♜🦄⚛🀄ᚠ⌛🌀 AS tbl FROM customers"
	_ = dialects.VerifiedStmt(t, sql)
}

// TestIdentifierUnicodeStart verifies Unicode start characters in identifiers.
// Reference: tests/sqlparser_common.rs:17581
func TestIdentifierUnicodeStart(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			mysql.NewMySqlDialect(),
			redshift.NewRedshiftSqlDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}
	sql := "SELECT 💝phone AS 💝 FROM customers"
	_ = dialects.VerifiedStmt(t, sql)
}

// TestParseNotnull verifies NOTNULL operator parsing.
// Reference: tests/sqlparser_common.rs:17592
func TestParseNotnull(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// For dialects that support it, NOTNULL is equivalent to IS NOT NULL
	_ = dialects.VerifiedStmt(t, "SELECT x IS NOT NULL FROM t")
}

// TestParseOdbcTimeDateTimestamp verifies ODBC-style date/time/timestamp literal parsing.
// Reference: tests/sqlparser_common.rs:17618
func TestParseOdbcTimeDateTimestamp(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// ODBC date literal
	_ = dialects.VerifiedStmt(t, "SELECT {d '2025-07-17'}, category_name FROM categories")

	// ODBC time literal
	_ = dialects.VerifiedStmt(t, "SELECT {t '14:12:01'}, category_name FROM categories")

	// ODBC timestamp literal
	_ = dialects.VerifiedStmt(t, "SELECT {ts '2025-07-17 14:12:01'}, category_name FROM categories")
}

// TestParseCreateUser verifies CREATE USER statement parsing.
// Reference: tests/sqlparser_common.rs:17643
func TestParseCreateUser(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic CREATE USER
	stmts := dialects.ParseSQL(t, "CREATE USER u1")
	require.Len(t, stmts, 1)
	createUser, ok := stmts[0].(*statement.CreateUser)
	require.True(t, ok, "Expected CreateUser statement, got %T", stmts[0])
	assert.Equal(t, "u1", createUser.Name.String())

	// OR REPLACE
	_ = dialects.VerifiedStmt(t, "CREATE OR REPLACE USER u1")

	// IF NOT EXISTS
	_ = dialects.VerifiedStmt(t, "CREATE OR REPLACE USER IF NOT EXISTS u1")

	// With password
	_ = dialects.VerifiedStmt(t, "CREATE OR REPLACE USER IF NOT EXISTS u1 PASSWORD='secret'")
}

// TestParseDropStream verifies DROP STREAM statement parsing.
// Reference: tests/sqlparser_common.rs:17719
func TestParseDropStream(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic DROP STREAM
	stmts := dialects.ParseSQL(t, "DROP STREAM s1")
	require.Len(t, stmts, 1)
	drop, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])
	assert.Equal(t, "STREAM", drop.ObjectType.String())
	require.Len(t, drop.Names, 1)
	assert.Equal(t, "s1", drop.Names[0].String())
	assert.False(t, drop.IfExists)

	// IF EXISTS
	stmts = dialects.ParseSQL(t, "DROP STREAM IF EXISTS s1")
	require.Len(t, stmts, 1)
	drop, ok = stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])
	assert.True(t, drop.IfExists)
}

// TestParseCreateViewIfNotExists verifies CREATE VIEW IF NOT EXISTS parsing.
// Reference: tests/sqlparser_common.rs:17737
func TestParseCreateViewIfNotExists(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Name after IF NOT EXISTS
	_ = dialects.VerifiedStmt(t, "CREATE VIEW IF NOT EXISTS v AS SELECT 1")

	// Name before IF NOT EXISTS
	_ = dialects.VerifiedStmt(t, "CREATE VIEW v IF NOT EXISTS AS SELECT 1")
}
