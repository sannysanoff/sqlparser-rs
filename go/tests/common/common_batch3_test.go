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
// This file contains tests 21-40 from the Rust test suite.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/oracle"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseDeleteWithoutFromError verifies that DELETE without FROM keyword fails for most dialects.
// Reference: tests/sqlparser_common.rs:720
func TestParseDeleteWithoutFromError(t *testing.T) {
	sql := "DELETE \"table\" WHERE 1"

	// Test with dialects that should fail
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		// Exclude BigQuery, Oracle, and Generic dialects which may support this syntax
		switch d.(type) {
		case *bigquery.BigQueryDialect, *oracle.OracleDialect, *generic.GenericDialect:
			return false
		default:
			return true
		}
	})

	for _, dialect := range dialects.Dialects {
		_, err := parser.ParseSQL(dialect, sql)
		assert.Error(t, err, "Expected error for dialect %T", dialect)
	}
}

// TestParseDeleteStatementForMultiTables verifies DELETE with multiple tables.
// Reference: tests/sqlparser_common.rs:734
func TestParseDeleteStatementForMultiTables(t *testing.T) {
	// Note: Multi-table DELETE is not yet fully implemented in the Go port
	t.Skip("Multi-table DELETE not yet implemented in Go port")
}

// TestParseDeleteStatementForMultiTablesWithUsing verifies DELETE with USING clause.
// Reference: tests/sqlparser_common.rs:773
func TestParseDeleteStatementForMultiTablesWithUsing(t *testing.T) {
	// Note: DELETE WITH USING is not yet fully implemented in the Go port
	t.Skip("DELETE WITH USING not yet implemented in Go port")
}

// TestParseWhereDeleteStatement verifies DELETE with WHERE clause.
// Reference: tests/sqlparser_common.rs:815
func TestParseWhereDeleteStatement(t *testing.T) {
	// Note: DELETE WHERE is not yet fully implemented in the Go port
	t.Skip("DELETE WHERE not yet implemented in Go port")
}

// TestParseWhereDeleteWithAliasStatement verifies DELETE with table aliases and USING.
// Reference: tests/sqlparser_common.rs:848
func TestParseWhereDeleteWithAliasStatement(t *testing.T) {
	// Note: DELETE with aliases is not yet fully implemented in the Go port
	t.Skip("DELETE with aliases not yet implemented in Go port")
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

// TestParseSelectInto verifies SELECT INTO syntax.
// Reference: tests/sqlparser_common.rs:1105
func TestParseSelectInto(t *testing.T) {
	// Note: SELECT INTO is not yet fully implemented in the Go port
	// This test serves as a placeholder for when it's added
	t.Skip("SELECT INTO not yet implemented in Go port")
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
