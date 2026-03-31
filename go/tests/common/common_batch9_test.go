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
// This file contains tests 141-160.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseCreateTableWithOptions verifies CREATE TABLE with WITH options parsing.
// Reference: tests/sqlparser_common.rs:4614
func TestParseCreateTableWithOptions(t *testing.T) {
	genericDialect := generic.NewGenericDialect()
	d := utils.NewTestedDialects()
	d.Dialects = []dialects.Dialect{genericDialect}

	sql := "CREATE TABLE t (c INT) WITH (foo = 'bar', a = 123)"
	stmt := d.VerifiedStmt(t, sql)

	createTable, ok := stmt.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement")

	// Verify table options exist and are of type WITH
	require.NotNil(t, createTable.TableOptions)
	assert.Equal(t, expr.CreateTableOptionsWith, createTable.TableOptions.Type)
	assert.Len(t, createTable.TableOptions.Options, 2)

	// Verify first option
	opt1 := createTable.TableOptions.Options[0]
	assert.Equal(t, "foo", opt1.Name.String())
	assert.NotNil(t, opt1.Value)

	// Verify second option
	opt2 := createTable.TableOptions.Options[1]
	assert.Equal(t, "a", opt2.Name.String())
}

// TestParseCreateTableClone verifies CREATE TABLE CLONE parsing.
// Reference: tests/sqlparser_common.rs:4645
func TestParseCreateTableClone(t *testing.T) {
	sql := "CREATE OR REPLACE TABLE a CLONE a_tmp"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	createTable, ok := stmt.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement")

	// Verify name is "a"
	assert.Equal(t, "a", createTable.Name.String())

	// Verify clone points to "a_tmp"
	require.NotNil(t, createTable.Clone)
	assert.Equal(t, "a_tmp", createTable.Clone.String())
	assert.True(t, createTable.OrReplace)
}

// TestParseCreateTableTrailingComma verifies CREATE TABLE with trailing comma (DuckDB).
// Reference: tests/sqlparser_common.rs:4657
func TestParseCreateTableTrailingComma(t *testing.T) {
	duckdbDialect := duckdb.NewDuckDbDialect()

	sql := "CREATE TABLE foo (bar int,);"
	canonical := "CREATE TABLE foo (bar INT)"

	stmts, err := parser.ParseSQL(duckdbDialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Verify it re-serializes to canonical form
	assert.Equal(t, canonical, stmts[0].String())
}

// TestParseCreateExternalTable verifies CREATE EXTERNAL TABLE parsing.
// Reference: tests/sqlparser_common.rs:4665
func TestParseCreateExternalTable(t *testing.T) {
	sql := `CREATE EXTERNAL TABLE uk_cities (
		   name VARCHAR(100) NOT NULL,
		   lat DOUBLE NULL,
		   lng DOUBLE)
		   STORED AS TEXTFILE LOCATION '/tmp/example.csv'`

	canonical := `CREATE EXTERNAL TABLE uk_cities (
		 name VARCHAR(100) NOT NULL, 
		 lat DOUBLE NULL, 
		 lng DOUBLE) 
		 STORED AS TEXTFILE LOCATION '/tmp/example.csv'`

	stmt := utils.NewTestedDialects().OneStatementParsesTo(t, sql, canonical)

	createTable, ok := stmt.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement")

	// Verify basic properties
	assert.Equal(t, "uk_cities", createTable.Name.String())
	assert.True(t, createTable.External)
	assert.Len(t, createTable.Columns, 3)

	// Verify columns
	assert.Equal(t, "name", createTable.Columns[0].Name.String())
	assert.Equal(t, "lat", createTable.Columns[1].Name.String())
	assert.Equal(t, "lng", createTable.Columns[2].Name.String())

	// Verify external table properties
	require.NotNil(t, createTable.FileFormat)
	assert.Equal(t, expr.FileFormatTEXTFILE, *createTable.FileFormat)
	require.NotNil(t, createTable.Location)
	assert.Equal(t, "/tmp/example.csv", *createTable.Location)

	// Verify table options are none
	assert.Nil(t, createTable.TableOptions)
	assert.False(t, createTable.IfNotExists)
}

// TestParseCreateOrReplaceExternalTable verifies CREATE OR REPLACE EXTERNAL TABLE parsing.
// Reference: tests/sqlparser_common.rs:4735
func TestParseCreateOrReplaceExternalTable(t *testing.T) {
	sql := `CREATE OR REPLACE EXTERNAL TABLE uk_cities (
		   name VARCHAR(100) NOT NULL)
		   STORED AS TEXTFILE LOCATION '/tmp/example.csv'`

	canonical := `CREATE OR REPLACE EXTERNAL TABLE uk_cities (
		 name VARCHAR(100) NOT NULL) 
		 STORED AS TEXTFILE LOCATION '/tmp/example.csv'`

	stmt := utils.NewTestedDialects().OneStatementParsesTo(t, sql, canonical)

	createTable, ok := stmt.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement")

	// Verify basic properties
	assert.Equal(t, "uk_cities", createTable.Name.String())
	assert.True(t, createTable.External)
	assert.True(t, createTable.OrReplace)
	assert.Len(t, createTable.Columns, 1)

	// Verify external table properties
	require.NotNil(t, createTable.FileFormat)
	assert.Equal(t, expr.FileFormatTEXTFILE, *createTable.FileFormat)
	require.NotNil(t, createTable.Location)
	assert.Equal(t, "/tmp/example.csv", *createTable.Location)

	// Verify table options are none
	assert.Nil(t, createTable.TableOptions)
	assert.False(t, createTable.IfNotExists)
}

// TestParseCreateExternalTableLowercase verifies lowercase CREATE EXTERNAL TABLE parsing.
// Reference: tests/sqlparser_common.rs:4790
func TestParseCreateExternalTableLowercase(t *testing.T) {
	sql := `create external table uk_cities (
		   name varchar(100) not null,
		   lat double null,
		   lng double)
		   stored as parquet location '/tmp/example.csv'`

	canonical := `CREATE EXTERNAL TABLE uk_cities (
		 name VARCHAR(100) NOT NULL, 
		 lat DOUBLE NULL, 
		 lng DOUBLE) 
		 STORED AS PARQUET LOCATION '/tmp/example.csv'`

	stmt := utils.NewTestedDialects().OneStatementParsesTo(t, sql, canonical)

	createTable, ok := stmt.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement")
	assert.True(t, createTable.External)
}

// TestParseCreateTableHiveFormatsNoneWhenNoOptions verifies hive_formats is None when no options.
// Reference: tests/sqlparser_common.rs:4808
func TestParseCreateTableHiveFormatsNoneWhenNoOptions(t *testing.T) {
	sql := "CREATE TABLE simple_table (id INT, name VARCHAR(100))"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	createTable, ok := stmt.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement")
	assert.Nil(t, createTable.HiveFormats)
}

// TestParseAlterTableBatch9 verifies ALTER TABLE operations parsing.
// Reference: tests/sqlparser_common.rs:4819
func TestParseAlterTableBatch9(t *testing.T) {
	d := utils.NewTestedDialects()

	// Test ADD COLUMN
	addColumn := "ALTER TABLE tab ADD COLUMN foo TEXT;"
	stmt := d.OneStatementParsesTo(t, addColumn, "ALTER TABLE tab ADD COLUMN foo TEXT")

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	assert.Equal(t, "tab", alterTable.Name.String())
	require.Len(t, alterTable.Operations, 1)

	// Test RENAME TABLE TO
	renameTable := "ALTER TABLE tab RENAME TO new_tab"
	stmt = d.VerifiedStmt(t, renameTable)
	alterTable, ok = stmt.(*statement.AlterTable)
	require.True(t, ok)
	assert.Equal(t, "tab", alterTable.Name.String())

	// Test RENAME TABLE AS
	renameTableAs := "ALTER TABLE tab RENAME AS new_tab"
	d.VerifiedStmt(t, renameTableAs)

	// Test RENAME COLUMN
	renameColumn := "ALTER TABLE tab RENAME COLUMN foo TO new_foo"
	d.VerifiedStmt(t, renameColumn)

	// Test SET TBLPROPERTIES
	setTblProperties := "ALTER TABLE tab SET TBLPROPERTIES('classification' = 'parquet')"
	d.VerifiedStmt(t, setTblProperties)

	// Test SET with parentheses
	setStorageParams := "ALTER TABLE tab SET (autovacuum_vacuum_scale_factor = 0.01, autovacuum_vacuum_threshold = 500)"
	d.VerifiedStmt(t, setStorageParams)
}

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

// TestAlterTableWithOnCluster verifies ALTER TABLE with ON CLUSTER parsing.
// Reference: tests/sqlparser_common.rs:4983
func TestAlterTableWithOnCluster(t *testing.T) {
	d := utils.NewTestedDialects()

	// Test with quoted cluster name
	sql := "ALTER TABLE t ON CLUSTER 'cluster' ADD CONSTRAINT bar PRIMARY KEY (baz)"
	stmt := d.VerifiedStmt(t, sql)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	assert.Equal(t, "t", alterTable.Name.String())

	// Test with unquoted cluster name
	sql2 := "ALTER TABLE t ON CLUSTER cluster_name ADD CONSTRAINT bar PRIMARY KEY (baz)"
	stmt = d.VerifiedStmt(t, sql2)

	alterTable, ok = stmt.(*statement.AlterTable)
	require.True(t, ok)
	assert.Equal(t, "t", alterTable.Name.String())

	// Test error: numeric cluster name
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "ALTER TABLE t ON CLUSTER 123 ADD CONSTRAINT bar PRIMARY KEY (baz)")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "identifier")
}

// TestParseAlterIndex verifies ALTER INDEX statement parsing.
// Reference: tests/sqlparser_common.rs:5017
func TestParseAlterIndex(t *testing.T) {
	sql := "ALTER INDEX idx RENAME TO new_idx"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	alterIndex, ok := stmt.(*statement.AlterIndex)
	require.True(t, ok, "Expected AlterIndex statement")
	assert.Equal(t, "idx", alterIndex.Name.String())
	require.NotNil(t, alterIndex.Operation)
}

// TestParseAlterView verifies ALTER VIEW statement parsing.
// Reference: tests/sqlparser_common.rs:5032
func TestParseAlterView(t *testing.T) {
	sql := "ALTER VIEW myschema.myview AS SELECT foo FROM bar"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	alterView, ok := stmt.(*statement.AlterView)
	require.True(t, ok, "Expected AlterView statement")
	assert.Equal(t, "myschema.myview", alterView.Name.String())
	assert.Empty(t, alterView.Columns)
	require.NotNil(t, alterView.Query)
	assert.Empty(t, alterView.WithOptions)
}

// TestParseAlterViewWithOptions verifies ALTER VIEW with options parsing.
// Reference: tests/sqlparser_common.rs:5051
func TestParseAlterViewWithOptions(t *testing.T) {
	sql := "ALTER VIEW v WITH (foo = 'bar', a = 123) AS SELECT 1"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	alterView, ok := stmt.(*statement.AlterView)
	require.True(t, ok, "Expected AlterView statement")
	require.Len(t, alterView.WithOptions, 2)

	// Verify first option
	opt1 := alterView.WithOptions[0]
	assert.Equal(t, "foo", opt1.Name.String())

	// Verify second option
	opt2 := alterView.WithOptions[1]
	assert.Equal(t, "a", opt2.Name.String())
}

// TestParseAlterViewWithColumns verifies ALTER VIEW with columns parsing.
// Reference: tests/sqlparser_common.rs:5076
func TestParseAlterViewWithColumns(t *testing.T) {
	sql := "ALTER VIEW v (has, cols) AS SELECT 1, 2"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	alterView, ok := stmt.(*statement.AlterView)
	require.True(t, ok, "Expected AlterView statement")
	assert.Equal(t, "v", alterView.Name.String())
	require.Len(t, alterView.Columns, 2)
	assert.Equal(t, "has", alterView.Columns[0].String())
	assert.Equal(t, "cols", alterView.Columns[1].String())
	require.NotNil(t, alterView.Query)
	assert.Empty(t, alterView.WithOptions)
}

// TestParseAlterTableAddColumn verifies ALTER TABLE ADD COLUMN variations.
// Reference: tests/sqlparser_common.rs:5095
func TestParseAlterTableAddColumn(t *testing.T) {
	d := utils.NewTestedDialects()

	// Without COLUMN keyword
	sql1 := "ALTER TABLE tab ADD foo TEXT"
	stmt := d.VerifiedStmt(t, sql1)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	require.Len(t, alterTable.Operations, 1)

	// With COLUMN keyword
	sql2 := "ALTER TABLE tab ADD COLUMN foo TEXT"
	_ = d.VerifiedStmt(t, sql2)
}

// TestParseAlterTableAddColumnIfNotExists verifies ALTER TABLE ADD COLUMN IF NOT EXISTS.
// Reference: tests/sqlparser_common.rs:5112
func TestParseAlterTableAddColumnIfNotExists(t *testing.T) {
	// Test with specific dialects that support this feature
	filteredDialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
		},
	}

	// Without COLUMN keyword
	sql1 := "ALTER TABLE tab ADD IF NOT EXISTS foo TEXT"
	stmt := filteredDialects.VerifiedStmt(t, sql1)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	require.Len(t, alterTable.Operations, 1)

	// With COLUMN keyword
	sql2 := "ALTER TABLE tab ADD COLUMN IF NOT EXISTS foo TEXT"
	stmt = filteredDialects.VerifiedStmt(t, sql2)

	alterTable, ok = stmt.(*statement.AlterTable)
	require.True(t, ok)
	require.Len(t, alterTable.Operations, 1)
}

// TestParseAlterTableConstraints verifies ALTER TABLE ADD CONSTRAINT parsing.
// Reference: tests/sqlparser_common.rs:5143
func TestParseAlterTableConstraints(t *testing.T) {
	d := utils.NewTestedDialects()

	constraints := []string{
		"CONSTRAINT address_pkey PRIMARY KEY (address_id)",
		"CONSTRAINT uk_task UNIQUE (report_date, task_id)",
		"CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES public.address(address_id)",
		"CONSTRAINT ck CHECK (rtrim(ltrim(REF_CODE)) <> '')",
		"PRIMARY KEY (foo, bar)",
		"UNIQUE (id)",
		"FOREIGN KEY (foo, bar) REFERENCES AnotherTable(foo, bar)",
		"CHECK (end_date > start_date OR end_date IS NULL)",
		"CONSTRAINT fk FOREIGN KEY (lng) REFERENCES othertable4",
	}

	for _, constraint := range constraints {
		sql := "ALTER TABLE tab ADD " + constraint
		stmt := d.VerifiedStmt(t, sql)

		alterTable, ok := stmt.(*statement.AlterTable)
		require.True(t, ok, "Expected AlterTable statement for: %s", constraint)
		require.Len(t, alterTable.Operations, 1)

		// Also verify CREATE TABLE with constraint
		createSql := "CREATE TABLE foo (id INT, " + constraint + ")"
		d.VerifiedStmt(t, createSql)
	}
}

// TestParseAlterTableDropColumn verifies ALTER TABLE DROP COLUMN parsing.
// Reference: tests/sqlparser_common.rs:5172
func TestParseAlterTableDropColumn(t *testing.T) {
	d := utils.NewTestedDialects()

	// Test various DROP COLUMN variations
	variations := []string{
		"DROP COLUMN IF EXISTS is_active",
		"DROP COLUMN IF EXISTS is_active CASCADE",
		"DROP COLUMN IF EXISTS is_active RESTRICT",
	}

	for _, variation := range variations {
		sql := "ALTER TABLE tab " + variation
		stmt := d.VerifiedStmt(t, sql)

		alterTable, ok := stmt.(*statement.AlterTable)
		require.True(t, ok, "Expected AlterTable statement for: %s", variation)
		require.Len(t, alterTable.Operations, 1)
	}

	// Test without COLUMN keyword with CASCADE
	d.OneStatementParsesTo(t,
		"ALTER TABLE tab DROP is_active CASCADE",
		"ALTER TABLE tab DROP is_active CASCADE")

	// Test comma-separated drop columns (dialects that support it)
	filteredDialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsCommaSeparatedDropColumnList()
	})
	if len(filteredDialects.Dialects) > 0 {
		filteredDialects.VerifiedStmt(t, "ALTER TABLE tbl DROP COLUMN c1, c2, c3")
	}
}

// TestParseAlterTableAlterColumn verifies ALTER TABLE ALTER COLUMN operations.
// Reference: tests/sqlparser_common.rs:5210
func TestParseAlterTableAlterColumn(t *testing.T) {
	d := utils.NewTestedDialects()

	// SET NOT NULL
	sql1 := "ALTER TABLE tab ALTER COLUMN is_active SET NOT NULL"
	stmt := d.VerifiedStmt(t, sql1)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	require.Len(t, alterTable.Operations, 1)

	// DROP NOT NULL (reserializes with COLUMN keyword)
	d.OneStatementParsesTo(t,
		"ALTER TABLE tab ALTER is_active DROP NOT NULL",
		"ALTER TABLE tab ALTER COLUMN is_active DROP NOT NULL")

	// SET DEFAULT
	sql3 := "ALTER TABLE tab ALTER COLUMN is_active SET DEFAULT 0"
	_ = d.VerifiedStmt(t, sql3)

	// DROP DEFAULT
	sql4 := "ALTER TABLE tab ALTER COLUMN is_active DROP DEFAULT"
	_ = d.VerifiedStmt(t, sql4)
}

// TestParseAlterTableAlterColumnType verifies ALTER TABLE ALTER COLUMN SET DATA TYPE.
// Reference: tests/sqlparser_common.rs:5254
func TestParseAlterTableAlterColumnType(t *testing.T) {
	d := utils.NewTestedDialects()

	// SET DATA TYPE
	sql := "ALTER TABLE tab ALTER COLUMN is_active SET DATA TYPE TEXT"
	stmt := d.VerifiedStmt(t, sql)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	require.Len(t, alterTable.Operations, 1)

	// TYPE (without SET)
	sql2 := "ALTER TABLE tab ALTER COLUMN is_active TYPE TEXT"
	d.VerifiedStmt(t, sql2)

	// SET DATA TYPE with USING (dialects that support it)
	filteredDialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsAlterColumnTypeUsing()
	})
	if len(filteredDialects.Dialects) > 0 {
		filteredDialects.VerifiedStmt(t,
			"ALTER TABLE tab ALTER COLUMN is_active SET DATA TYPE TEXT USING 'text'")
	}

	// Test that dialects without USING support reject it
	exceptDialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return !d.SupportsAlterColumnTypeUsing()
	})
	if len(exceptDialects.Dialects) > 0 {
		_, err := parser.ParseSQL(exceptDialects.Dialects[0],
			"ALTER TABLE tab ALTER COLUMN is_active SET DATA TYPE TEXT USING 'text'")
		require.Error(t, err)
	}
}

// alterTableOp extracts the first AlterTableOperation from a statement
func alterTableOp(stmt ast.Statement) *expr.AlterTableOperation {
	if alterTable, ok := stmt.(*statement.AlterTable); ok && len(alterTable.Operations) > 0 {
		return alterTable.Operations[0]
	}
	return nil
}
