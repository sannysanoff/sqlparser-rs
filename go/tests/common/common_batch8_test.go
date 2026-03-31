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
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	dialectPkg "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/hive"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestCompoundExpr tests compound expression parsing with array access and field access.
// Reference: tests/sqlparser_common.rs:3636 (test 121)
func TestCompoundExpr(t *testing.T) {
	supportedDialects := &utils.TestedDialects{
		Dialects: []dialectPkg.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			bigquery.NewBigQueryDialect(),
		},
	}

	sqls := []string{
		"SELECT abc[1].f1 FROM t",
		"SELECT abc[1].f1.f2 FROM t",
		"SELECT f1.abc[1] FROM t",
		"SELECT f1.f2.abc[1] FROM t",
		"SELECT f1.abc[1].f2 FROM t",
		"SELECT named_struct('a', 1, 'b', 2).a",
		"SELECT named_struct('a', 1, 'b', 2).a",
		"SELECT make_array(1, 2, 3)[1]",
		"SELECT make_array(named_struct('a', 1))[1].a",
		"SELECT abc[1][-1].a.b FROM t",
		"SELECT abc[1][-1].a.b[1] FROM t",
	}

	for _, sql := range sqls {
		supportedDialects.VerifiedStmt(t, sql)
	}
}

// TestDoubleValue tests parsing of various double/floating-point number formats.
// Reference: tests/sqlparser_common.rs:3661 (test 122)
func TestDoubleValue(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test cases for various double formats
	testCases := []string{
		"0.",
		"0.0",
		"0000.",
		"0000.00",
		".0",
		".00",
		"0e0",
		"0e+0",
		"0e-0",
		"0.e-0",
		"0.e+0",
		".0e-0",
		".0e+0",
		"00.0e+0",
		"00.0e-0",
	}

	for _, num := range testCases {
		// Test positive and negative versions
		for _, sign := range []string{"", "+", "-"} {
			signedNum := sign + num
			sql := "SELECT " + signedNum
			_, err := parser.ParseSQL(dialects.Dialects[0], sql)
			require.NoError(t, err, "Failed to parse: %s", sql)
		}
	}
}

// TestParseNegativeValue tests parsing of negative values in SELECT and CREATE SEQUENCE.
// Reference: tests/sqlparser_common.rs:3768 (test 123)
func TestParseNegativeValue(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT -1"
	dialects.OneStatementParsesTo(t, sql1, "SELECT -1")

	sql2 := "CREATE SEQUENCE name INCREMENT -10 MINVALUE -1000 MAXVALUE 15 START -100;"
	dialects.OneStatementParsesTo(t, sql2,
		"CREATE SEQUENCE name INCREMENT -10 MINVALUE -1000 MAXVALUE 15 START -100")
}

// TestParseCreateTable tests CREATE TABLE statement parsing.
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

// TestParseAssert tests ASSERT statement parsing.
// Reference: tests/sqlparser_common.rs:4381 (test 130)
func TestParseAssert(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "ASSERT (SELECT COUNT(*) FROM my_table) > 0"
	ast := dialects.OneStatementParsesTo(t, sql, "ASSERT (SELECT COUNT(*) FROM my_table) > 0")
	require.NotNil(t, ast)

	// Check it's an Assert statement without message
	assertStmt, ok := ast.(*statement.Assert)
	require.True(t, ok)
	require.Nil(t, assertStmt.Message)
}

// TestParseAssertMessage tests ASSERT statement with AS message parsing.
// Reference: tests/sqlparser_common.rs:4396 (test 131)
func TestParseAssertMessage(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "ASSERT (SELECT COUNT(*) FROM my_table) > 0 AS 'No rows in my_table'"
	ast := dialects.OneStatementParsesTo(t, sql,
		"ASSERT (SELECT COUNT(*) FROM my_table) > 0 AS 'No rows in my_table'")
	require.NotNil(t, ast)

	// Check it's an Assert statement with message
	assertStmt, ok := ast.(*statement.Assert)
	require.True(t, ok)
	require.NotNil(t, assertStmt.Message)
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

// TestParseDropSchema tests DROP SCHEMA statement parsing.
// Reference: tests/sqlparser_common.rs:4466 (test 135)
func TestParseDropSchema(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "DROP SCHEMA X"
	ast := dialects.VerifiedStmt(t, sql)
	require.NotNil(t, ast)

	// Check it's a Drop statement
	dropStmt, ok := ast.(*statement.Drop)
	require.True(t, ok)
	assert.Equal(t, "schema", dropStmt.ObjectType)
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
		Dialects: []dialectPkg.Dialect{
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
