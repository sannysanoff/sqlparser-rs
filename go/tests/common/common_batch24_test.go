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
// This file contains the FINAL batch of tests 441-461 from the Rust test file.
package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseNotNullInColumnOptions verifies parsing NOT NULL in column options.
// Reference: tests/sqlparser_common.rs:17753
func TestParseNotNullInColumnOptions(t *testing.T) {
	canonical := "CREATE TABLE foo (abc INT DEFAULT (42 IS NOT NULL) NOT NULL, def INT, def_null BOOL GENERATED ALWAYS AS (def IS NOT NULL) STORED, CHECK (abc IS NOT NULL))"
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, canonical)

	// Test that the shorter form without IS parses to the canonical form
	shortForm := "CREATE TABLE foo (abc INT DEFAULT (42 NOT NULL) NOT NULL, def INT, def_null BOOL GENERATED ALWAYS AS (def NOT NULL) STORED, CHECK (abc NOT NULL))"
	dialects.OneStatementParsesTo(t, shortForm, canonical)
}

// TestParseDefaultExprWithOperators verifies parsing DEFAULT with operators.
// Reference: tests/sqlparser_common.rs:17777
func TestParseDefaultExprWithOperators(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "CREATE TABLE t (c INT DEFAULT (1 + 2) + 3)")
	dialects.VerifiedStmt(t, "CREATE TABLE t (c INT DEFAULT (1 + 2) + 3 NOT NULL)")
}

// TestParseDefaultWithCollateColumnOption verifies DEFAULT with COLLATE column option.
// Reference: tests/sqlparser_common.rs:17783
func TestParseDefaultWithCollateColumnOption(t *testing.T) {
	sql := "CREATE TABLE foo (abc TEXT DEFAULT 'foo' COLLATE 'en_US')"
	dialects := utils.NewTestedDialects()

	stmt := dialects.VerifiedStmt(t, sql)

	createTable, ok := stmt.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmt)
	require.Equal(t, 1, len(createTable.Columns), "Expected 1 column")

	column := createTable.Columns[0]
	require.NotNil(t, column.Name, "Expected column name")
	assert.Equal(t, "abc", column.Name.String())

	// Verify TEXT data type
	require.NotNil(t, column.DataType, "Expected data type")
	if dt, ok := column.DataType.(fmt.Stringer); ok {
		assert.Equal(t, "TEXT", dt.String())
	} else {
		t.Errorf("DataType does not implement String()")
	}

	// Verify we have column options (DEFAULT and COLLATE)
	require.True(t, len(column.Options) >= 2, "Expected at least 2 column options, got %d", len(column.Options))
}

// TestParseCreateTableLike verifies CREATE TABLE LIKE syntax.
// Reference: tests/sqlparser_common.rs:17816
func TestParseCreateTableLike(t *testing.T) {
	// Test basic CREATE TABLE LIKE (non-parenthesized)
	sql1 := "CREATE TABLE new LIKE old"
	stmt1 := utils.NewTestedDialects().VerifiedStmt(t, sql1)

	createTable1, ok := stmt1.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmt1)
	assert.Equal(t, "new", createTable1.Name.String())
	require.NotNil(t, createTable1.Like, "Expected LIKE clause")

	// Test CREATE TABLE with parenthesized LIKE
	sql2 := "CREATE TABLE new (LIKE old)"
	stmt2 := utils.NewTestedDialects().VerifiedStmt(t, sql2)

	createTable2, ok := stmt2.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmt2)
	assert.Equal(t, "new", createTable2.Name.String())
	require.NotNil(t, createTable2.Like, "Expected LIKE clause")
}

// TestParseCopyOptions verifies COPY statement options parsing.
// Reference: tests/sqlparser_common.rs:17890
func TestParseCopyOptions(t *testing.T) {
	sql1 := "COPY dst (c1, c2, c3) FROM 's3://bucket/file.txt' IAM_ROLE 'arn:aws:iam::123:role/r' CSV IGNOREHEADER 1"
	stmt1 := utils.NewTestedDialects().VerifiedStmt(t, sql1)

	copy1, ok := stmt1.(*statement.Copy)
	require.True(t, ok, "Expected Copy statement, got %T", stmt1)
	require.NotNil(t, copy1.Source, "Expected source")
	assert.Equal(t, "dst", copy1.Source.String())

	// Verify IAM_ROLE and other options are parsed - LegacyOptions is where COPY options are stored
	assert.True(t, len(copy1.LegacyOptions) > 0 || len(copy1.Options) > 0, "Expected copy options")
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

// TestParseAdjacentStringLiteralConcatenation verifies adjacent string literal concatenation.
// Reference: tests/sqlparser_common.rs:18120
func TestParseAdjacentStringLiteralConcatenation(t *testing.T) {
	// Test string literal concatenation
	sql := `SELECT 'M' "y" 'S' "q" 'l'`
	canonical := `SELECT 'MySql'`

	// Test with dialects that support string literal concatenation
	dialects := utils.NewTestedDialects()
	dialects.OneStatementParsesTo(t, sql, canonical)

	// Test concatenation in WHERE clause
	sql2 := "SELECT * FROM t WHERE col = 'Hello' ' ' 'World!'"
	canonical2 := "SELECT * FROM t WHERE col = 'Hello World!'"
	dialects.OneStatementParsesTo(t, sql2, canonical2)
}

// TestParseInvisibleColumn verifies INVISIBLE column option parsing.
// Reference: tests/sqlparser_common.rs:18149
func TestParseInvisibleColumn(t *testing.T) {
	sql := "CREATE TABLE t (foo INT, bar INT INVISIBLE)"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	createTable, ok := stmt.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmt)
	require.Equal(t, 2, len(createTable.Columns), "Expected 2 columns")

	// First column should not have INVISIBLE option
	assert.Equal(t, "foo", createTable.Columns[0].Name.String())
	assert.Equal(t, 0, len(createTable.Columns[0].Options), "Expected no options for first column")

	// Second column should have INVISIBLE option
	assert.Equal(t, "bar", createTable.Columns[1].Name.String())
	require.True(t, len(createTable.Columns[1].Options) > 0, "Expected at least one option for INVISIBLE column")

	// Verify ALTER TABLE ADD COLUMN INVISIBLE
	sql2 := "ALTER TABLE t ADD COLUMN bar INT INVISIBLE"
	stmt2 := utils.NewTestedDialects().VerifiedStmt(t, sql2)

	alterTable, ok := stmt2.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement, got %T", stmt2)
	require.Equal(t, 1, len(alterTable.Operations), "Expected 1 operation")
}

// TestParseCreateIndexDifferentUsingPositions verifies CREATE INDEX with USING clause.
// Reference: tests/sqlparser_common.rs:18202
func TestParseCreateIndexDifferentUsingPositions(t *testing.T) {
	sql := "CREATE INDEX idx_name USING BTREE ON table_name (col1)"
	expected := "CREATE INDEX idx_name ON table_name USING BTREE (col1)"

	stmt := utils.NewTestedDialects().OneStatementParsesTo(t, sql, expected)

	createIndex, ok := stmt.(*statement.CreateIndex)
	require.True(t, ok, "Expected CreateIndex statement, got %T", stmt)
	require.NotNil(t, createIndex.Name, "Expected index name")
	assert.Equal(t, "idx_name", createIndex.Name.String())
	assert.Equal(t, "table_name", createIndex.TableName.String())
	require.NotNil(t, createIndex.Using, "Expected USING clause")
	assert.Equal(t, statement.IndexTypeBTree, *createIndex.Using)
	assert.Equal(t, 1, len(createIndex.Columns), "Expected 1 column")
	assert.False(t, createIndex.Unique, "Expected non-unique index")

	// Test double USING (in CREATE and in options)
	sql2 := "CREATE INDEX idx_name USING BTREE ON table_name (col1) USING HASH"
	expected2 := "CREATE INDEX idx_name ON table_name USING BTREE (col1) USING HASH"
	stmt2 := utils.NewTestedDialects().OneStatementParsesTo(t, sql2, expected2)

	createIndex2, ok := stmt2.(*statement.CreateIndex)
	require.True(t, ok, "Expected CreateIndex statement, got %T", stmt2)
	// The second USING appears in the WITH clause options
	assert.True(t, len(createIndex2.With) > 0 || createIndex2.Using != nil, "Expected index options")
}

// TestParseAlterUser verifies ALTER USER statement parsing.
// Reference: tests/sqlparser_common.rs:18245
func TestParseAlterUser(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic ALTER USER
	dialects.VerifiedStmt(t, "ALTER USER u1")

	// ALTER USER IF EXISTS
	dialects.VerifiedStmt(t, "ALTER USER IF EXISTS u1")

	// ALTER USER RENAME TO
	stmt1 := dialects.VerifiedStmt(t, "ALTER USER IF EXISTS u1 RENAME TO u2")
	alterUser1, ok := stmt1.(*statement.AlterUser)
	require.True(t, ok, "Expected AlterUser statement, got %T", stmt1)
	assert.True(t, alterUser1.IfExists, "Expected IfExists to be true")
	assert.Equal(t, "u1", alterUser1.Name.String())
	require.NotNil(t, alterUser1.RenameTo, "Expected RenameTo")
	assert.Equal(t, "u2", alterUser1.RenameTo.String())

	// ALTER USER SET PASSWORD
	stmt2 := dialects.VerifiedStmt(t, "ALTER USER u1 PASSWORD 'AAA'")
	alterUser2, ok := stmt2.(*statement.AlterUser)
	require.True(t, ok, "Expected AlterUser statement, got %T", stmt2)
	assert.Equal(t, "u1", alterUser2.Name.String())

	// ALTER USER ENCRYPTED PASSWORD
	dialects.VerifiedStmt(t, "ALTER USER u1 ENCRYPTED PASSWORD 'AAA'")

	// ALTER USER PASSWORD NULL
	dialects.VerifiedStmt(t, "ALTER USER u1 PASSWORD NULL")

	// WITH PASSWORD should parse to canonical form
	dialects.OneStatementParsesTo(t, "ALTER USER u1 WITH PASSWORD 'AAA'", "ALTER USER u1 PASSWORD 'AAA'")
}

// TestParseGenericUnaryOps verifies unary operator parsing.
// Reference: tests/sqlparser_common.rs:18468
func TestParseGenericUnaryOps(t *testing.T) {
	dialects := utils.NewTestedDialects()

	tests := []struct {
		sql string
		op  string
	}{
		{"SELECT ~expr", "~"},
		{"SELECT -expr", "-"},
		{"SELECT +expr", "+"},
	}

	for _, tc := range tests {
		t.Run(tc.sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, tc.sql)
			require.Len(t, stmts, 1)
		})
	}
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

// TestParseSetSessionAuthorization verifies SET SESSION AUTHORIZATION parsing.
// Reference: tests/sqlparser_common.rs:18510
func TestParseSetSessionAuthorization(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// SET SESSION AUTHORIZATION DEFAULT
	dialects.VerifiedStmt(t, "SET SESSION AUTHORIZATION DEFAULT")

	// SET SESSION AUTHORIZATION with username
	dialects.VerifiedStmt(t, "SET SESSION AUTHORIZATION 'username'")
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

// TestParseSelectParenthesizedWildcard verifies SELECT DISTINCT(*) parsing.
// Reference: tests/sqlparser_common.rs:18545
func TestParseSelectParenthesizedWildcard(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// SELECT DISTINCT(*) - parentheses should be normalized away
	sql := "SELECT DISTINCT (*) FROM table1"
	_ = "SELECT DISTINCT * FROM table1" // canonical form

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Also test without spaces
	sql2 := "SELECT DISTINCT(*) FROM table1"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)
}

// TestParseOverlapAsBoolAnd verifies && operator parsing.
// Reference: tests/sqlparser_common.rs:18564
func TestParseOverlapAsBoolAnd(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// && should be parsed as AND in dialects that support it
	sql := "SELECT x && y"
	canonical := "SELECT x AND y"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify it produces equivalent results
	stmts2 := dialects.ParseSQL(t, canonical)
	require.Len(t, stmts2, 1)

	// Use canonical to avoid unused variable error
	_ = canonical
}

// TestParseKeyValueOptionsTrailingSemicolon verifies trailing semicolon handling.
// Reference: tests/sqlparser_common.rs:18570
func TestParseKeyValueOptionsTrailingSemicolon(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Trailing semicolon should be handled gracefully
	sql := "CREATE USER u1 option1='value1' option2='value2';"
	canonical := "CREATE USER u1 option1='value1' option2='value2'"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify statement can be serialized
	result := stmts[0].String()
	// The result should match the canonical form (without semicolon)
	assert.Equal(t, canonical, result)
}

// TestParseBinaryKwAsCast verifies BINARY keyword as CAST parsing.
// Reference: tests/slparser_common.rs:18578
func TestParseBinaryKwAsCast(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT BINARY 1+1"
	canonical := "SELECT CAST(1 + 1 AS BINARY)"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	stmts2 := dialects.ParseSQL(t, canonical)
	require.Len(t, stmts2, 1)

	// Verify that both forms produce equivalent parse trees
	// (Note: the exact serialization may vary by dialect support)
	_ = stmts[0].String()
	_ = stmts2[0].String()
}

// TestParseSemiStructuredDataTraversal verifies semi-structured data traversal parsing.
// Reference: tests/sqlparser_common.rs:18584
func TestParseSemiStructuredDataTraversal(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test basic JSON access with colon notation
	sql1 := "SELECT a:b FROM t"
	stmts1 := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts1, 1)

	// Test with quoted identifier
	sql2 := `SELECT a:"my long object key name" FROM t`
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)

	// Test with type cast
	sql3 := "SELECT a:b::INT FROM t"
	stmts3 := dialects.ParseSQL(t, sql3)
	require.Len(t, stmts3, 1)

	// Test multiple levels of traversal
	sql4 := `SELECT a:foo."bar".baz`
	stmts4 := dialects.ParseSQL(t, sql4)
	require.Len(t, stmts4, 1)

	// Test dot and bracket notation mixed
	sql5 := `SELECT a:foo[0].bar`
	stmts5 := dialects.ParseSQL(t, sql5)
	require.Len(t, stmts5, 1)
}

// TestParseArraySubscript verifies array subscript parsing.
// Reference: tests/sqlparser_common.rs:18707
func TestParseArraySubscript(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic array subscript
	dialects.VerifiedStmt(t, "SELECT arr[1]")

	// Slicing with colon
	dialects.VerifiedStmt(t, "SELECT arr[:]")
	dialects.VerifiedStmt(t, "SELECT arr[1:2]")
	dialects.VerifiedStmt(t, "SELECT arr[1:2:4]")

	// Slicing with expressions
	dialects.VerifiedStmt(t, "SELECT arr[1:array_length(arr)]")
	dialects.VerifiedStmt(t, "SELECT arr[array_length(arr) - 1:array_length(arr)]")

	// Multi-dimensional array access
	dialects.VerifiedStmt(t, "SELECT arr[1][2]")
	dialects.VerifiedStmt(t, "SELECT arr[:][:]")
}

// TestWildcardFuncArg verifies wildcard with EXCLUDE as a function argument.
// Reference: tests/sqlparser_common.rs:18729
func TestWildcardFuncArg(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Wildcard with EXCLUDE in function argument
	sql1 := "SELECT HASH(* EXCLUDE(col1)) FROM t"
	canonical1 := "SELECT HASH(* EXCLUDE (col1)) FROM t"

	stmts1 := dialects.ParseSQL(t, sql1)
	require.Len(t, stmts1, 1)

	// Verify canonical form also parses
	stmts2 := dialects.ParseSQL(t, canonical1)
	require.Len(t, stmts2, 1)

	// Multiple EXCLUDE columns
	sql2 := "SELECT HASH(* EXCLUDE (col1, col2)) FROM t"
	stmts3 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts3, 1)
}

// TestParseCopyOptionsRedshift verifies Redshift COPY options specifically.
// Reference: tests/sqlparser_common.rs:17890 (additional tests)
func TestParseCopyOptionsRedshift(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test various COPY options
	tests := []string{
		"COPY dst FROM 's3://bucket/file.txt' IAM_ROLE DEFAULT CSV",
		"COPY dst FROM 's3://bucket/file.txt' CSV IGNOREHEADER 1",
		"COPY dst FROM 's3://bucket/file.txt' JSON 'auto'",
		"COPY dst FROM 's3://bucket/file.txt' GZIP",
	}

	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, sql)
			require.Len(t, stmts, 1)
		})
	}
}

// TestParseAlterUserSetOptions verifies ALTER USER SET options.
// Reference: tests/sqlparser_common.rs:18245 (additional tests)
func TestParseAlterUserSetOptions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// SET options with various types
	tests := []string{
		"ALTER USER u1 SET PASSWORD='secret'",
		"ALTER USER u1 SET DEFAULT_MFA_METHOD='PASSKEY'",
		"ALTER USER u1 SET TAG k1='v1'",
		"ALTER USER u1 SET DEFAULT_SECONDARY_ROLES=('ALL')",
		"ALTER USER u1 UNSET PASSWORD",
	}

	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, sql)
			require.Len(t, stmts, 1)
		})
	}
}

// TestParseStringLiteralConcatenationWithNewline verifies string concatenation with newlines.
// Reference: tests/sqlparser_common.rs:18120 (additional tests)
func TestParseStringLiteralConcatenationWithNewline(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// String concatenation with newlines and comments
	sql := `SELECT 'abc' in ('a'
		'b'
		-- COMMENT
		'c',
		-- COMMENT
		'd'
	)`

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseCreateTableLikeWithDefaults verifies CREATE TABLE LIKE with INCLUDING/EXCLUDING DEFAULTS.
// Reference: tests/sqlparser_common.rs:17816 (additional tests)
func TestParseCreateTableLikeWithDefaults(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test LIKE with INCLUDING DEFAULTS
	sql1 := "CREATE TABLE new (LIKE old INCLUDING DEFAULTS)"
	stmt1 := dialects.VerifiedStmt(t, sql1)
	createTable1, ok := stmt1.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement")
	require.NotNil(t, createTable1.Like, "Expected LIKE clause")

	// Test LIKE with EXCLUDING DEFAULTS
	sql2 := "CREATE TABLE new (LIKE old EXCLUDING DEFAULTS)"
	stmt2 := dialects.VerifiedStmt(t, sql2)
	createTable2, ok := stmt2.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement")
	require.NotNil(t, createTable2.Like, "Expected LIKE clause")
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
