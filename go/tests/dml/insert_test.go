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
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseInsertValues verifies INSERT with VALUES clause.
// Reference: tests/sqlparser_common.rs:97
func TestParseInsertValues(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "INSERT INTO customer VALUES (1, 2, 3)"
	dialects.VerifiedStmt(t, sql)

	sql2 := "INSERT INTO customer (id, name, active) VALUES (1, 2, 3)"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "INSERT INTO customer VALUES (1, 2, 3), (1, 2, 3)"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseInsertDefaultValues verifies INSERT with DEFAULT VALUES.
// Reference: tests/sqlparser_common.rs:198
func TestParseInsertDefaultValues(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "INSERT INTO test_table DEFAULT VALUES"
	dialects.VerifiedStmt(t, sql)
}

// TestParseInsertValuesFull verifies INSERT with VALUES clause parsing (comprehensive).
// Reference: tests/sqlparser_common.rs:96
func TestParseInsertValuesFull(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test various INSERT forms
	testCases := []struct {
		sql          string
		tableName    string
		columns      []string
		rowCount     int
		valueKeyword bool
	}{
		{
			sql:          "INSERT customer VALUES (1, 2, 3)",
			tableName:    "customer",
			columns:      []string{},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO customer VALUES (1, 2, 3)",
			tableName:    "customer",
			columns:      []string{},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO customer VALUES (1, 2, 3), (1, 2, 3)",
			tableName:    "customer",
			columns:      []string{},
			rowCount:     2,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO public.customer VALUES (1, 2, 3)",
			tableName:    "public.customer",
			columns:      []string{},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO db.public.customer VALUES (1, 2, 3)",
			tableName:    "db.public.customer",
			columns:      []string{},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO public.customer (id, name, active) VALUES (1, 2, 3)",
			tableName:    "public.customer",
			columns:      []string{"id", "name", "active"},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          `INSERT INTO t (id, name, active) VALUE (1, 2, 3)`,
			tableName:    "t",
			columns:      []string{"id", "name", "active"},
			rowCount:     1,
			valueKeyword: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, tc.sql)
			require.Len(t, stmts, 1)

			insert, ok := stmts[0].(*statement.Insert)
			require.True(t, ok, "Expected Insert statement, got %T", stmts[0])

			// Verify table name
			assert.Equal(t, tc.tableName, insert.Table.String())

			// Verify columns
			assert.Equal(t, len(tc.columns), len(insert.Columns))
			for i, col := range tc.columns {
				assert.Equal(t, col, insert.Columns[i].String())
			}

			// Verify source exists (VALUES clause)
			require.NotNil(t, insert.Source)
			require.NotNil(t, insert.Source.Body)

			// Check that body is a ValuesSetExpr
			valuesExpr, ok := insert.Source.Body.(*query.ValuesSetExpr)
			require.True(t, ok, "Expected ValuesSetExpr, got %T", insert.Source.Body)

			// Verify row count
			assert.Equal(t, tc.rowCount, len(valuesExpr.Values.Rows))

			// Verify VALUE keyword vs VALUES keyword
			assert.Equal(t, tc.valueKeyword, valuesExpr.Values.ValueKeyword)
		})
	}

	// Test INSERT with CTE
	cteSQL := "INSERT INTO customer WITH foo AS (SELECT 1) SELECT * FROM foo UNION VALUES (1)"
	dialects.VerifiedStmt(t, cteSQL)
}

// TestParseInsertSet verifies INSERT with SET clause parsing (MySQL-specific).
// Reference: tests/sqlparser_common.rs:180
func TestParseInsertSet(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsInsertSet()
	})

	sql := "INSERT INTO tbl1 SET col1 = 1, col2 = 'abc', col3 = current_date()"
	dialects.VerifiedStmt(t, sql)
}

// TestParseReplaceInto verifies REPLACE INTO parsing error (not supported by PostgreSQL).
// Reference: tests/sqlparser_common.rs:186
func TestParseReplaceInto(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()
	sql := "REPLACE INTO public.customer (id, name, active) VALUES (1, 2, 3)"

	_, err := parser.ParseSQL(dialect, sql)
	require.Error(t, err)
	// The error should indicate REPLACE is not supported
	assert.Contains(t, err.Error(), "REPLACE")
}

// TestParseInsertDefaultValuesFull verifies INSERT with DEFAULT VALUES (comprehensive).
// Reference: tests/sqlparser_common.rs:197
func TestParseInsertDefaultValuesFull(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test basic INSERT with DEFAULT VALUES
	t.Run("basic default values", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES"
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		insert, ok := stmts[0].(*statement.Insert)
		require.True(t, ok, "Expected Insert statement")

		// Verify table name
		assert.Equal(t, "test_table", insert.Table.String())

		// Verify columns is empty
		assert.Empty(t, insert.Columns)

		// Verify source is nil (DEFAULT VALUES has no source)
		assert.Nil(t, insert.Source)

		// Verify no partitioned clause
		assert.Empty(t, insert.Partitioned)

		// Verify no returning clause
		assert.Empty(t, insert.Returning)

		// Verify no on conflict clause
		assert.Nil(t, insert.On)
	})

	// Test INSERT with DEFAULT VALUES and RETURNING
	t.Run("default values with returning", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES RETURNING test_column"
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		insert, ok := stmts[0].(*statement.Insert)
		require.True(t, ok, "Expected Insert statement")

		assert.Equal(t, "test_table", insert.Table.String())
		assert.Empty(t, insert.Columns)
		assert.Nil(t, insert.Source)
		assert.NotEmpty(t, insert.Returning, "Should have RETURNING clause")
	})

	// Test INSERT with DEFAULT VALUES and ON CONFLICT
	t.Run("default values with on conflict", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES ON CONFLICT DO NOTHING"
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		insert, ok := stmts[0].(*statement.Insert)
		require.True(t, ok, "Expected Insert statement")

		assert.Equal(t, "test_table", insert.Table.String())
		assert.Empty(t, insert.Columns)
		assert.Nil(t, insert.Source)
		assert.NotNil(t, insert.On, "Should have ON CONFLICT clause")
	})

	// Test error: columns with DEFAULT VALUES should fail
	t.Run("columns with default values error", func(t *testing.T) {
		sql := "INSERT INTO test_table (test_col) DEFAULT VALUES"
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DEFAULT")
	})

	// Test error: DEFAULT VALUES with Hive after columns should fail
	t.Run("default values with after columns error", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES (some_column)"
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err)
	})

	// Test error: DEFAULT VALUES with Hive partition should fail
	t.Run("default values with partition error", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES PARTITION (some_column)"
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err)
	})

	// Test error: DEFAULT VALUES with values list should fail
	t.Run("default values with values list error", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES (1)"
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err)
	})
}

// TestParseInsertSelectReturning verifies INSERT with SELECT and RETURNING clause.
// Reference: tests/sqlparser_common.rs:312
func TestParseInsertSelectReturning(t *testing.T) {
	// Note: The original Rust test filters dialects that treat RETURNING as column alias.
	// For now, we test with all dialects that support RETURNING.
	dialects := utils.NewTestedDialects()

	// Basic INSERT ... SELECT ... RETURNING
	dialects.VerifiedStmt(t, "INSERT INTO t SELECT 1 RETURNING 2")

	// INSERT with SELECT and RETURNING with alias
	stmts := dialects.ParseSQL(t, "INSERT INTO t SELECT x RETURNING x AS y")
	require.Len(t, stmts, 1)

	insert, ok := stmts[0].(*statement.Insert)
	require.True(t, ok, "Expected Insert statement, got %T", stmts[0])

	// Verify source is present
	require.NotNil(t, insert.Source, "Expected source (SELECT) to be present")

	// Verify returning is present with 1 item
	require.NotNil(t, insert.Returning, "Expected RETURNING clause to be present")
	assert.Equal(t, 1, len(insert.Returning), "Expected 1 RETURNING item")
}

// TestParseInsertSelectFromReturning verifies INSERT with SELECT FROM and RETURNING clause.
// Reference: tests/sqlparser_common.rs:331
func TestParseInsertSelectFromReturning(t *testing.T) {
	sql := "INSERT INTO table1 SELECT * FROM table2 RETURNING id"
	stmts := utils.NewTestedDialects().ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	insert, ok := stmts[0].(*statement.Insert)
	require.True(t, ok, "Expected Insert statement, got %T", stmts[0])

	// Verify table name
	assert.Equal(t, "table1", insert.Table.String())

	// Verify source is present (SELECT)
	require.NotNil(t, insert.Source, "Expected source to be present")

	// Verify RETURNING clause is present with 1 item
	require.NotNil(t, insert.Returning, "Expected RETURNING clause to be present")
	assert.Equal(t, 1, len(insert.Returning))
}

// TestParseReturningAsColumnAlias verifies that RETURNING can be used as a column alias.
// Reference: tests/sqlparser_common.rs:352
func TestParseReturningAsColumnAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT 1 AS RETURNING")
}

// TestParseInsertSqlite verifies SQLite-specific INSERT syntax with ON CONFLICT clauses.
// Reference: tests/sqlparser_common.rs:357
func TestParseInsertSqlite(t *testing.T) {
	dialect := sqlite.NewSQLiteDialect()

	check := func(t *testing.T, sql string, expectedAction int) {
		stmts, err := parser.ParseSQL(dialect, sql)
		require.NoError(t, err, "Failed to parse: %s", sql)
		require.Len(t, stmts, 1)

		insert, ok := stmts[0].(*statement.Insert)
		require.True(t, ok, "Expected Insert statement, got %T", stmts[0])

		if expectedAction == 0 {
			assert.Nil(t, insert.Or, "Expected no OR clause for: %s", sql)
		} else {
			require.NotNil(t, insert.Or, "Expected OR clause for: %s", sql)
		}
	}

	// Standard INSERT without OR clause
	check(t, "INSERT INTO test_table(id) VALUES(1)", 0)

	// REPLACE INTO
	check(t, "INSERT INTO test_table(id) VALUES(1)", 1)

	// INSERT OR REPLACE
	check(t, "INSERT OR REPLACE INTO test_table(id) VALUES(1)", 1)

	// INSERT OR ROLLBACK
	check(t, "INSERT OR ROLLBACK INTO test_table(id) VALUES(1)", 1)

	// INSERT OR ABORT
	check(t, "INSERT OR ABORT INTO test_table(id) VALUES(1)", 1)

	// INSERT OR FAIL
	check(t, "INSERT OR FAIL INTO test_table(id) VALUES(1)", 1)

	// INSERT OR IGNORE
	check(t, "INSERT OR IGNORE INTO test_table(id) VALUES(1)", 1)
}
