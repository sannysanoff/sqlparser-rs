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

// Package ddl contains the DDL (Data Definition Language) SQL parsing tests.
// These tests are ported from tests/sqlparser_common.rs in the Rust implementation.
package ddl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseRenameTable verifies RENAME TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:4924
func TestParseRenameTable(t *testing.T) {
	d := utils.NewTestedDialects()

	// Single table rename
	sql := "RENAME TABLE test.test1 TO test_db.test2"
	stmt := d.VerifiedStmt(t, sql)

	_, ok := stmt.(*statement.RenameTable)
	require.True(t, ok, "Expected RenameTable statement")

	// Multiple table rename
	multiSql := "RENAME TABLE old_table1 TO new_table1, old_table2 TO new_table2, old_table3 TO new_table3"
	stmt = d.VerifiedStmt(t, multiSql)
	_, ok = stmt.(*statement.RenameTable)
	require.True(t, ok)

	// Test error: extra token after statement
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "RENAME TABLE old_table TO new_table a")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "end of statement")

	// Test error: wrong keyword
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "RENAME TABLE1 old_table TO new_table a")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "TABLE")
}

// TestParsePipeOperatorDrop verifies pipe operator DROP syntax.
// Reference: tests/sqlparser_common.rs:16633
func TestParsePipeOperatorDrop(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> DROP id")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> DROP id, name")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> DROP c |> RENAME a AS x")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> DROP a, b |> SELECT c")
}

// TestParsePipeOperatorRename verifies pipe operator RENAME syntax.
// Reference: tests/sqlparser_common.rs:16696
func TestParsePipeOperatorRename(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> RENAME old_name AS new_name")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> RENAME id AS user_id, name AS user_name")
	dialects.OneStatementParsesTo(t, "SELECT * FROM tbl |> RENAME id user_id", "SELECT * FROM tbl |> RENAME id AS user_id")
}
