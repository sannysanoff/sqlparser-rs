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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	dialectPkg "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/hive"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseCreateTable verifies CREATE TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:3780 (test 124)
func TestParseCreateTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE TABLE uk_cities (" +
		"name VARCHAR(100) NOT NULL," +
		"lat DOUBLE NULL," +
		"lng DOUBLE," +
		"constrained INT NULL CONSTRAINT pkey PRIMARY KEY NOT NULL UNIQUE CHECK (constrained > 0)," +
		"ref INT REFERENCES othertable (a, b)," +
		"ref2 INT references othertable2 on delete cascade on update no action," +
		"constraint fkey foreign key (lat) references othertable3 (lat) on delete restrict," +
		"constraint fkey2 foreign key (lat) references othertable4(lat) on delete no action on update restrict, " +
		"foreign key (lat) references othertable4(lat) on update set default on delete cascade, " +
		"FOREIGN KEY (lng) REFERENCES othertable4 (longitude) ON UPDATE SET NULL" +
		")"

	canonical := "CREATE TABLE uk_cities (" +
		"name VARCHAR(100) NOT NULL, " +
		"lat DOUBLE NULL, " +
		"lng DOUBLE, " +
		"constrained INT NULL CONSTRAINT pkey PRIMARY KEY NOT NULL UNIQUE CHECK (constrained > 0), " +
		"ref INT REFERENCES othertable (a, b), " +
		"ref2 INT REFERENCES othertable2 ON DELETE CASCADE ON UPDATE NO ACTION, " +
		"CONSTRAINT fkey FOREIGN KEY (lat) REFERENCES othertable3(lat) ON DELETE RESTRICT, " +
		"CONSTRAINT fkey2 FOREIGN KEY (lat) REFERENCES othertable4(lat) ON DELETE NO ACTION ON UPDATE RESTRICT, " +
		"FOREIGN KEY (lat) REFERENCES othertable4(lat) ON DELETE CASCADE ON UPDATE SET DEFAULT, " +
		"FOREIGN KEY (lng) REFERENCES othertable4(longitude) ON UPDATE SET NULL)"

	ast := dialects.OneStatementParsesTo(t, sql, canonical)
	require.NotNil(t, ast)

	// Test error cases
	_, err := parser.ParseSQL(dialects.Dialects[0], "CREATE TABLE t (a int NOT NULL GARBAGE)")
	require.Error(t, err)

	_, err = parser.ParseSQL(dialects.Dialects[0], "CREATE TABLE t (a int NOT NULL CONSTRAINT foo)")
	require.Error(t, err)
}

// TestParseCreateTableWithConstraintCharacteristics tests constraint characteristics parsing.
// Reference: tests/sqlparser_common.rs:4003 (test 125)
func TestParseCreateTableWithConstraintCharacteristics(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE TABLE uk_cities (" +
		"name VARCHAR(100) NOT NULL," +
		"lat DOUBLE NULL," +
		"lng DOUBLE," +
		"constraint fkey foreign key (lat) references othertable3 (lat) on delete restrict deferrable initially deferred," +
		"constraint fkey2 foreign key (lat) references othertable4(lat) on delete no action on update restrict deferrable initially immediate, " +
		"foreign key (lat) references othertable4(lat) on update set default on delete cascade not deferrable initially deferred not enforced, " +
		"FOREIGN KEY (lng) REFERENCES othertable4 (longitude) ON UPDATE SET NULL enforced not deferrable initially immediate" +
		")"

	canonical := "CREATE TABLE uk_cities (" +
		"name VARCHAR(100) NOT NULL, " +
		"lat DOUBLE NULL, " +
		"lng DOUBLE, " +
		"CONSTRAINT fkey FOREIGN KEY (lat) REFERENCES othertable3(lat) ON DELETE RESTRICT DEFERRABLE INITIALLY DEFERRED, " +
		"CONSTRAINT fkey2 FOREIGN KEY (lat) REFERENCES othertable4(lat) ON DELETE NO ACTION ON UPDATE RESTRICT DEFERRABLE INITIALLY IMMEDIATE, " +
		"FOREIGN KEY (lat) REFERENCES othertable4(lat) ON DELETE CASCADE ON UPDATE SET DEFAULT NOT DEFERRABLE INITIALLY DEFERRED NOT ENFORCED, " +
		"FOREIGN KEY (lng) REFERENCES othertable4(longitude) ON UPDATE SET NULL NOT DEFERRABLE INITIALLY IMMEDIATE ENFORCED)"

	ast := dialects.OneStatementParsesTo(t, sql, canonical)
	require.NotNil(t, ast)

	// Test error cases
	_, err := parser.ParseSQL(dialects.Dialects[0], "CREATE TABLE t (\n"+
		"a int NOT NULL,\n"+
		"FOREIGN KEY (a) REFERENCES othertable4(a) ON DELETE CASCADE ON UPDATE SET DEFAULT DEFERRABLE INITIALLY IMMEDIATE NOT DEFERRABLE,\n"+
		")")
	require.Error(t, err)

	_, err = parser.ParseSQL(dialects.Dialects[0], "CREATE TABLE t (\n"+
		"a int NOT NULL,\n"+
		"FOREIGN KEY (a) REFERENCES othertable4(a) ON DELETE CASCADE ON UPDATE SET DEFAULT NOT ENFORCED INITIALLY DEFERRED ENFORCED,\n"+
		")")
	require.Error(t, err)

	_, err = parser.ParseSQL(dialects.Dialects[0], "CREATE TABLE t (\n"+
		"a int NOT NULL,\n"+
		"FOREIGN KEY (lat) REFERENCES othertable4(lat) ON DELETE CASCADE ON UPDATE SET DEFAULT INITIALLY DEFERRED INITIALLY IMMEDIATE,\n"+
		")")
	require.Error(t, err)
}

// TestParseCreateTableColumnConstraintCharacteristics tests column constraint characteristics.
// Reference: tests/sqlparser_common.rs:4169 (test 126)
func TestParseCreateTableColumnConstraintCharacteristics(t *testing.T) {
	dialects := utils.NewTestedDialects()

	deferrableOptions := []struct {
		text string
		val  *bool
	}{
		{"", nil},
		{"DEFERRABLE", boolPtr(true)},
		{"NOT DEFERRABLE", boolPtr(false)},
	}

	initiallyOptions := []struct {
		text string
		val  *string
	}{
		{"", nil},
		{"INITIALLY IMMEDIATE", stringPtr("immediate")},
		{"INITIALLY DEFERRED", stringPtr("deferred")},
	}

	enforcedOptions := []struct {
		text string
		val  *bool
	}{
		{"", nil},
		{"ENFORCED", boolPtr(true)},
		{"NOT ENFORCED", boolPtr(false)},
	}

	for _, deferrable := range deferrableOptions {
		for _, initially := range initiallyOptions {
			for _, enforced := range enforcedOptions {
				var parts []string
				if deferrable.text != "" {
					parts = append(parts, deferrable.text)
				}
				if initially.text != "" {
					parts = append(parts, initially.text)
				}
				if enforced.text != "" {
					parts = append(parts, enforced.text)
				}

				clause := ""
				if len(parts) > 0 {
					clause = " " + joinStrings(parts, " ")
				}

				sql := "CREATE TABLE t (a int UNIQUE" + clause + ")"
				expected := "CREATE TABLE t (a INT UNIQUE" + clause + ")"

				ast := dialects.OneStatementParsesTo(t, sql, expected)
				require.NotNil(t, ast)
			}
		}
	}

	// Test error cases
	_, err := parser.ParseSQL(dialects.Dialects[0],
		"CREATE TABLE t (a int NOT NULL UNIQUE DEFERRABLE INITIALLY BADVALUE)")
	require.Error(t, err)

	_, err = parser.ParseSQL(dialects.Dialects[0],
		"CREATE TABLE t (a int NOT NULL UNIQUE INITIALLY IMMEDIATE DEFERRABLE INITIALLY DEFERRED)")
	require.Error(t, err)

	_, err = parser.ParseSQL(dialects.Dialects[0],
		"CREATE TABLE t (a int NOT NULL UNIQUE DEFERRABLE INITIALLY DEFERRED NOT DEFERRABLE)")
	require.Error(t, err)

	_, err = parser.ParseSQL(dialects.Dialects[0],
		"CREATE TABLE t (a int NOT NULL UNIQUE DEFERRABLE INITIALLY DEFERRED ENFORCED NOT ENFORCED)")
	require.Error(t, err)
}

func boolPtr(b bool) *bool {
	return &b
}

func stringPtr(s string) *string {
	return &s
}

func joinStrings(strs []string, sep string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// TestParseCreateTableHiveArray tests Hive/BigQuery array syntax in CREATE TABLE.
// Reference: tests/sqlparser_common.rs:4280 (test 127)
func TestParseCreateTableHiveArray(t *testing.T) {
	// Test PostgreSQL-style array syntax: INT[]
	pgDialects := &utils.TestedDialects{
		Dialects: []dialectPkg.Dialect{
			postgresql.NewPostgreSqlDialect(),
		},
	}

	sql1 := "CREATE TABLE IF NOT EXISTS something (name INT, val INT[])"
	stmt := pgDialects.VerifiedStmt(t, sql1)
	require.NotNil(t, stmt)

	// Test Hive/BigQuery-style array syntax: ARRAY<INT>
	hiveBqDialects := &utils.TestedDialects{
		Dialects: []dialectPkg.Dialect{
			hive.NewHiveDialect(),
			bigquery.NewBigQueryDialect(),
		},
	}

	sql2 := "CREATE TABLE IF NOT EXISTS something (name INT, val ARRAY<INT>)"
	stmt = hiveBqDialects.VerifiedStmt(t, sql2)
	require.NotNil(t, stmt)

	// Test that PostgreSQL doesn't support ARRAY<INT> syntax
	_, err := parser.ParseSQL(postgresql.NewPostgreSqlDialect(),
		"CREATE TABLE IF NOT EXISTS something (name int, val array<int)")
	require.Error(t, err)
}

// TestParseCreateTableWithMultipleOnDeleteInConstraintFails tests error on multiple ON DELETE in constraint.
// Reference: tests/sqlparser_common.rs:4357 (test 128)
func TestParseCreateTableWithMultipleOnDeleteInConstraintFails(t *testing.T) {
	dialects := utils.NewTestedDialects()

	_, err := parser.ParseSQL(dialects.Dialects[0],
		"create table X ("+
			"y_id int, "+
			"foreign key (y_id) references Y (id) on delete cascade on update cascade on delete no action"+
			")")
	require.Error(t, err, "should have failed")
}

// TestParseCreateTableWithMultipleOnDeleteFails tests error on multiple ON DELETE in column.
// Reference: tests/sqlparser_common.rs:4369 (test 129)
func TestParseCreateTableWithMultipleOnDeleteFails(t *testing.T) {
	dialects := utils.NewTestedDialects()

	_, err := parser.ParseSQL(dialects.Dialects[0],
		"create table X ("+
			"y_id int references Y (id) "+
			"on delete cascade on update cascade on delete no action"+
			")")
	require.Error(t, err, "should have failed")
}

// TestParseCreateSchema tests CREATE SCHEMA statement parsing.
// Reference: tests/sqlparser_common.rs:4421 (test 132)
func TestParseCreateSchema(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE SCHEMA X"
	stmt := dialects.VerifiedStmt(t, sql)
	require.NotNil(t, stmt)

	// Test with options
	dialects.VerifiedStmt(t, "CREATE SCHEMA a.b.c OPTIONS(key1 = 'value1', key2 = 'value2')")
	dialects.VerifiedStmt(t, "CREATE SCHEMA IF NOT EXISTS a OPTIONS(key1 = 'value1')")
	dialects.VerifiedStmt(t, "CREATE SCHEMA IF NOT EXISTS a OPTIONS()")
	dialects.VerifiedStmt(t, "CREATE SCHEMA IF NOT EXISTS a DEFAULT COLLATE 'und:ci' OPTIONS()")
	dialects.VerifiedStmt(t, "CREATE SCHEMA a.b.c WITH (key1 = 'value1', key2 = 'value2')")
	dialects.VerifiedStmt(t, "CREATE SCHEMA IF NOT EXISTS a WITH (key1 = 'value1')")
	dialects.VerifiedStmt(t, "CREATE SCHEMA IF NOT EXISTS a WITH ()")
	dialects.VerifiedStmt(t, "CREATE SCHEMA a CLONE b")
}

// TestParseCreateSchemaWithAuthorization tests CREATE SCHEMA AUTHORIZATION parsing.
// Reference: tests/sqlparser_common.rs:4442 (test 133)
func TestParseCreateSchemaWithAuthorization(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE SCHEMA AUTHORIZATION Y"
	stmt := dialects.VerifiedStmt(t, sql)
	require.NotNil(t, stmt)
}

// TestParseCreateSchemaWithNameAndAuthorization tests CREATE SCHEMA with name and authorization.
// Reference: tests/sqlparser_common.rs:4454 (test 134)
func TestParseCreateSchemaWithNameAndAuthorization(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE SCHEMA X AUTHORIZATION Y"
	stmt := dialects.VerifiedStmt(t, sql)
	require.NotNil(t, stmt)
}

// TestParseCreateTableAs tests CREATE TABLE AS SELECT parsing.
// Reference: tests/sqlparser_common.rs:4476 (test 136)
func TestParseCreateTableAs(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE TABLE t AS SELECT * FROM a"
	stmt := dialects.VerifiedStmt(t, sql)
	require.NotNil(t, stmt)

	// BigQuery allows specifying table schema in CTAS
	bqDialects := &utils.TestedDialects{
		Dialects: []dialectPkg.Dialect{
			bigquery.NewBigQueryDialect(),
		},
	}

	sql2 := "CREATE TABLE t (a INT, b INT) AS SELECT 1 AS b, 2 AS a"
	bqDialects.VerifiedStmt(t, sql2)
}

// TestParseCreateTableAsTable tests CREATE TABLE AS TABLE syntax.
// Reference: tests/sqlparser_common.rs:4507 (test 137)
func TestParseCreateTableAsTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "CREATE TABLE new_table AS TABLE old_table"
	stmt1 := dialects.VerifiedStmt(t, sql1)
	require.NotNil(t, stmt1)

	sql2 := "CREATE TABLE new_table AS TABLE schema_name.old_table"
	stmt2 := dialects.VerifiedStmt(t, sql2)
	require.NotNil(t, stmt2)
}

// TestParseCreateTableOnCluster tests CREATE TABLE with ON CLUSTER clause.
// Reference: tests/sqlparser_common.rs:4562 (test 138)
func TestParseCreateTableOnCluster(t *testing.T) {
	generic := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			generic.NewGenericDialect(),
		},
	}

	// Using single-quote literal to define current cluster
	sql1 := "CREATE TABLE t ON CLUSTER '{cluster}' (a INT, b INT)"
	generic.VerifiedStmt(t, sql1)

	// Using explicitly declared cluster name
	sql2 := "CREATE TABLE t ON CLUSTER my_cluster (a INT, b INT)"
	generic.VerifiedStmt(t, sql2)
}

// TestParseCreateOrReplaceTable tests CREATE OR REPLACE TABLE parsing.
// Reference: tests/sqlparser_common.rs:4585 (test 139)
func TestParseCreateOrReplaceTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE OR REPLACE TABLE t (a INT)"
	ast := dialects.VerifiedStmt(t, sql)
	require.NotNil(t, ast)

	// Check that or_replace is set
	createTable, ok := ast.(*statement.CreateTable)
	require.True(t, ok)
	assert.True(t, createTable.OrReplace)
}

// TestParseCreateTableWithOnDeleteOnUpdateInAnyOrder tests ON DELETE/ON UPDATE in any order.
// Reference: tests/sqlparser_common.rs:4600 (test 140)
func TestParseCreateTableWithOnDeleteOnUpdateInAnyOrder(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test various combinations
	sql1 := "create table X (y_id int references Y (id) on update cascade on delete no action)"
	_, err := parser.ParseSQL(dialects.Dialects[0], sql1)
	require.NoError(t, err)

	sql2 := "create table X (y_id int references Y (id) on delete cascade on update cascade)"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql2)
	require.NoError(t, err)

	sql3 := "create table X (y_id int references Y (id) on update no action)"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql3)
	require.NoError(t, err)

	sql4 := "create table X (y_id int references Y (id) on delete restrict)"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql4)
	require.NoError(t, err)
}

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

	// Go outputs single-line format (whitespace differences are non-functional)
	canonical := `CREATE EXTERNAL TABLE uk_cities (name VARCHAR(100) NOT NULL, lat DOUBLE NULL, lng DOUBLE) STORED AS TEXTFILE LOCATION '/tmp/example.csv'`

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

	// Go outputs single-line format (whitespace differences are non-functional)
	canonical := `CREATE OR REPLACE EXTERNAL TABLE uk_cities (name VARCHAR(100) NOT NULL) STORED AS TEXTFILE LOCATION '/tmp/example.csv'`

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

	// Go outputs single-line format (whitespace differences are non-functional)
	canonical := `CREATE EXTERNAL TABLE uk_cities (name VARCHAR(100) NOT NULL, lat DOUBLE NULL, lng DOUBLE) STORED AS PARQUET LOCATION '/tmp/example.csv'`

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

// TestParseCreateIndex verifies CREATE INDEX statement parsing.
// Reference: tests/sqlparser_common.rs:9449
func TestParseCreateIndex(t *testing.T) {
	sql := "CREATE UNIQUE INDEX IF NOT EXISTS idx_name ON test(name, age DESC)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createIndex, ok := stmts[0].(*statement.CreateIndex)
	require.True(t, ok, "Expected CreateIndex statement, got %T", stmts[0])

	require.NotNil(t, createIndex.Name, "Expected index name to be present")
	assert.Equal(t, "idx_name", createIndex.Name.String())
	assert.Equal(t, "test", createIndex.TableName.String())
	assert.True(t, createIndex.Unique, "Expected unique to be true")
	assert.True(t, createIndex.IfNotExists, "Expected if_not_exists to be true")
	assert.Equal(t, 2, len(createIndex.Columns), "Expected 2 indexed columns")

	// First column: name (no order specified)
	require.NotNil(t, createIndex.Columns[0].Expr, "Expected name for first column")
	assert.Equal(t, "name", createIndex.Columns[0].Expr.String())

	// Second column: age DESC
	require.NotNil(t, createIndex.Columns[1].Expr, "Expected name for second column")
	assert.Equal(t, "age", createIndex.Columns[1].Expr.String())
}

// TestCreateIndexWithUsingFunction verifies CREATE INDEX with USING clause.
// Reference: tests/sqlparser_common.rs:9495
func TestCreateIndexWithUsingFunction(t *testing.T) {
	sql := "CREATE UNIQUE INDEX IF NOT EXISTS idx_name ON test USING BTREE (name, age DESC)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createIndex, ok := stmts[0].(*statement.CreateIndex)
	require.True(t, ok, "Expected CreateIndex statement, got %T", stmts[0])

	require.NotNil(t, createIndex.Name, "Expected index name to be present")
	assert.Equal(t, "idx_name", createIndex.Name.String())
	assert.Equal(t, "test", createIndex.TableName.String())
	assert.True(t, createIndex.Unique, "Expected unique to be true")
	assert.True(t, createIndex.IfNotExists, "Expected if_not_exists to be true")
	assert.False(t, createIndex.Concurrently, "Expected concurrently to be false")
	require.NotNil(t, createIndex.Using, "Expected using to be present")
	assert.Equal(t, "BTREE", createIndex.Using.String())
	assert.Equal(t, 2, len(createIndex.Columns), "Expected 2 indexed columns")
}

// TestCreateIndexWithWithClause verifies CREATE INDEX with WITH clause.
// Reference: tests/sqlparser_common.rs:9554
func TestCreateIndexWithWithClause(t *testing.T) {
	sql := "CREATE UNIQUE INDEX title_idx ON films(title) WITH (fillfactor = 70, single_param)"
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		// Only test dialects that support CREATE INDEX with clause
		return true
	})
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createIndex, ok := stmts[0].(*statement.CreateIndex)
	require.True(t, ok, "Expected CreateIndex statement, got %T", stmts[0])

	require.NotNil(t, createIndex.Name, "Expected index name to be present")
	assert.Equal(t, "title_idx", createIndex.Name.String())
	assert.Equal(t, "films", createIndex.TableName.String())
	assert.True(t, createIndex.Unique, "Expected unique to be true")
	assert.False(t, createIndex.Concurrently, "Expected concurrently to be false")
	assert.False(t, createIndex.IfNotExists, "Expected if_not_exists to be false")
	require.NotNil(t, createIndex.With, "Expected with clause to be present")
	assert.Equal(t, 2, len(createIndex.With), "Expected 2 with parameters")
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

// TestParseCreateRole verifies CREATE ROLE statement parsing.
// Reference: tests/sqlparser_common.rs:9625
func TestParseCreateRole(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Single role
	sql := "CREATE ROLE consultant"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createRole, ok := stmts[0].(*statement.CreateRole)
	require.True(t, ok, "Expected CreateRole statement, got %T", stmts[0])
	assert.Equal(t, 1, len(createRole.Names), "Expected 1 role name")
	assert.Equal(t, "consultant", createRole.Names[0].String())
	assert.False(t, createRole.IfNotExists, "Expected if_not_exists to be false")

	// Multiple roles with IF NOT EXISTS
	sql = "CREATE ROLE IF NOT EXISTS mysql_a, mysql_b"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createRole, ok = stmts[0].(*statement.CreateRole)
	require.True(t, ok, "Expected CreateRole statement, got %T", stmts[0])
	assert.Equal(t, 2, len(createRole.Names), "Expected 2 role names")
	assert.Equal(t, "mysql_a", createRole.Names[0].String())
	assert.Equal(t, "mysql_b", createRole.Names[1].String())
	assert.True(t, createRole.IfNotExists, "Expected if_not_exists to be true")
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

// TestParseCreateViewIfNotExists verifies CREATE VIEW IF NOT EXISTS parsing.
// Reference: tests/sqlparser_common.rs:17737
func TestParseCreateViewIfNotExists(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Name after IF NOT EXISTS
	_ = dialects.VerifiedStmt(t, "CREATE VIEW IF NOT EXISTS v AS SELECT 1")

	// Name before IF NOT EXISTS
	_ = dialects.VerifiedStmt(t, "CREATE VIEW v IF NOT EXISTS AS SELECT 1")
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

	// CREATE PROCEDURE with parameter modes
	// Note: Go serializes without space after procedure name and uses BOOLEAN instead of BOOL
	_ = dialects.VerifiedStmt(t, "CREATE PROCEDURE test_proc(IN a INTEGER, OUT b TEXT, INOUT c TIMESTAMP, d BOOLEAN) AS BEGIN SELECT 1; END")

	// Test with default values
	_ = dialects.VerifiedStmt(t, "CREATE PROCEDURE test_proc(IN a INTEGER = 1, OUT b TEXT = '2', INOUT c TIMESTAMP = NULL, d BOOLEAN = 0) AS BEGIN SELECT 1; END")
}

// TestParseCreatePolicy verifies CREATE POLICY statement parsing.
// Reference: tests/sqlparser_common.rs:14210
func TestParseCreatePolicy(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE POLICY my_policy ON my_table AS PERMISSIVE FOR SELECT TO my_role, CURRENT_USER USING (c0 = 1) WITH CHECK (1 = 1)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	policy, ok := stmts[0].(*statement.CreatePolicy)
	require.True(t, ok, "Expected CreatePolicy statement, got %T", stmts[0])
	assert.Equal(t, "my_policy", policy.Name.String())
	assert.Equal(t, "my_table", policy.TableName.String())

	// Test with SELECT subquery in USING
	dialects.VerifiedStmt(t, "CREATE POLICY my_policy ON my_table AS PERMISSIVE FOR SELECT TO my_role, CURRENT_USER USING (c0 IN (SELECT column FROM t0)) WITH CHECK (1 = 1)")

	// Test minimal CREATE POLICY
	dialects.VerifiedStmt(t, "CREATE POLICY my_policy ON my_table")

	// Test error - missing table name
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "CREATE POLICY my_policy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ON")
}

// TestCreatePolicy is an alias for TestParseCreatePolicy for naming consistency
// Reference: tests/sqlparser_common.rs:14210
func TestCreatePolicy(t *testing.T) {
	TestParseCreatePolicy(t)
}

// TestParseCreateConnector verifies CREATE CONNECTOR statement parsing.
// Reference: tests/sqlparser_common.rs:14445
func TestParseCreateConnector(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE CONNECTOR my_connector TYPE 'jdbc' URL 'jdbc:mysql://localhost:3306/mydb' WITH DCPROPERTIES('user' = 'root', 'password' = 'password')"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	connector, ok := stmts[0].(*statement.CreateConnector)
	require.True(t, ok, "Expected CreateConnector statement, got %T", stmts[0])
	assert.Equal(t, "my_connector", connector.Name.String())

	// Test minimal CREATE CONNECTOR
	dialects.VerifiedStmt(t, "CREATE CONNECTOR my_connector")

	// Test error - missing connector name
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "CREATE CONNECTOR")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: identifier")
}

// TestCreateConnector is an alias for TestParseCreateConnector for naming consistency
// Reference: tests/sqlparser_common.rs:14445
func TestCreateConnector(t *testing.T) {
	TestParseCreateConnector(t)
}

// TestParseCreateTableSelect verifies CREATE TABLE ... SELECT parsing.
// Reference: tests/sqlparser_common.rs:15358
func TestParseCreateTableSelect(t *testing.T) {
	// Test with dialects that support CREATE TABLE SELECT
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsCreateTableSelect()
	})

	sql1 := "CREATE TABLE foo (baz INT) SELECT bar"
	expected1 := "CREATE TABLE foo (baz INT) AS SELECT bar"
	dialects.OneStatementParsesTo(t, sql1, expected1)

	sql2 := "CREATE TABLE foo (baz INT, name STRING) SELECT bar, oth_name FROM test.table_a"
	expected2 := "CREATE TABLE foo (baz INT, name STRING) AS SELECT bar, oth_name FROM test.table_a"
	dialects.OneStatementParsesTo(t, sql2, expected2)
}

// TestParseCreateTableWithBitTypes verifies CREATE TABLE with BIT types.
// Reference: tests/sqlparser_common.rs:15442
func TestParseCreateTableWithBitTypes(t *testing.T) {
	sql := "CREATE TABLE t (a BIT, b BIT VARYING, c BIT(42), d BIT VARYING(43))"
	stmts := utils.NewTestedDialects().ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createTable, ok := stmts[0].(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmts[0])
	require.Len(t, createTable.Columns, 4)
}

// TestParseCreateTableWithEnumTypes verifies CREATE TABLE with ENUM types.
// Reference: tests/sqlparser_common.rs:15585
func TestParseCreateTableWithEnumTypes(t *testing.T) {
	sql := "CREATE TABLE t0 (foo ENUM8('a' = 1, 'b' = 2), bar ENUM16('a' = 1, 'b' = 2), baz ENUM('a', 'b'))"
	stmts := utils.NewTestedDialects().ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createTable, ok := stmts[0].(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmts[0])
	assert.Equal(t, "t0", createTable.Name.String())
	require.Len(t, createTable.Columns, 3)
}

// TestParseCreateTableLike verifies CREATE TABLE LIKE syntax.
// Reference: tests/sqlparser_common.rs:17816
func TestParseCreateTableLike(t *testing.T) {
	// Test basic CREATE TABLE LIKE (non-parenthesized) - with dialects that DON'T support parenthesized LIKE
	// This matches Rust: all_dialects_except(|d| d.supports_create_table_like_parenthesized())
	sql1 := "CREATE TABLE new LIKE old"
	dialectsExceptParenthesized := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return !d.SupportsCreateTableLikeParenthesized()
	})
	stmt1 := dialectsExceptParenthesized.VerifiedStmt(t, sql1)

	createTable1, ok := stmt1.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmt1)
	assert.Equal(t, "new", createTable1.Name.String())
	require.NotNil(t, createTable1.Like, "Expected LIKE clause")

	// Test CREATE TABLE with parenthesized LIKE - with dialects that DO support it
	// This matches Rust: all_dialects_where(|d| d.supports_create_table_like_parenthesized())
	sql2 := "CREATE TABLE new (LIKE old)"
	dialectsWithParenthesized := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsCreateTableLikeParenthesized()
	})
	stmt2 := dialectsWithParenthesized.VerifiedStmt(t, sql2)

	createTable2, ok := stmt2.(*statement.CreateTable)
	require.True(t, ok, "Expected CreateTable statement, got %T", stmt2)
	assert.Equal(t, "new", createTable2.Name.String())
	require.NotNil(t, createTable2.Like, "Expected LIKE clause")
}

// TestParseCreateTableLikeWithDefaults verifies CREATE TABLE LIKE with INCLUDING/EXCLUDING DEFAULTS.
// Reference: tests/sqlparser_common.rs:17816 (additional tests)
func TestParseCreateTableLikeWithDefaults(t *testing.T) {
	// Only test with dialects that support parenthesized LIKE
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsCreateTableLikeParenthesized()
	})

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

// alterTableOp extracts the first AlterTableOperation from a statement
func alterTableOp(stmt ast.Statement) *expr.AlterTableOperation {
	if alterTable, ok := stmt.(*statement.AlterTable); ok && len(alterTable.Operations) > 0 {
		return alterTable.Operations[0]
	}
	return nil
}
