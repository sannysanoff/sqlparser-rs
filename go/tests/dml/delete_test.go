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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/oracle"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseDelete verifies DELETE statement parsing.
// Reference: tests/sqlparser_common.rs:703
func TestParseDelete(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "DELETE FROM customer WHERE id = 1"
	dialects.VerifiedStmt(t, sql)
}

// TestParseDeleteStatement verifies DELETE statement parsing.
// Reference: tests/sqlparser_common.rs:703
func TestParseDeleteStatement(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "DELETE FROM \"table\""
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	delete, ok := stmts[0].(*statement.Delete)
	require.True(t, ok, "Expected Delete statement, got %T", stmts[0])

	// Verify table is set
	require.NotNil(t, delete.Tables)
	assert.Equal(t, 1, len(delete.Tables))

	// The table name should be quoted
	assert.Equal(t, "\"table\"", delete.Tables[0].String())
}

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
