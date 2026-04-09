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

package redshift

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/tests/utils"
)

func redshiftDialects() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			redshift.NewRedshiftSqlDialect(),
		},
	}
}

func redshiftAndGeneric() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			redshift.NewRedshiftSqlDialect(),
			generic.NewGenericDialect(),
		},
	}
}

func TestSquareBracketsOverDbSchemaTableName(t *testing.T) {
	redshiftDialects().VerifiedOnlySelect(t, "SELECT [col1] FROM [test_schema].[test_table]")
}

func TestBracketsOverDbSchemaTableNameWithWhitesPaces(t *testing.T) {
	stmts := redshiftDialects().ParseSQL(t, "SELECT [   col1  ] FROM [  test_schema].[ test_table]")
	assert.Len(t, stmts, 1)
}

func TestDoubleQuotesOverDbSchemaTableName(t *testing.T) {
	redshiftDialects().VerifiedOnlySelect(t, `SELECT "col1" FROM "test_schema"."test_table"`)
}

func TestParseDelimitedIdentifiers(t *testing.T) {
	// check that quoted identifiers in any position remain quoted after serialization
	redshiftDialects().VerifiedOnlySelect(t, `SELECT "alias"."bar baz", "myfun"(), "simple id" AS "column alias" FROM "a table" AS "alias"`)

	redshiftDialects().VerifiedStmt(t, `CREATE TABLE "foo" ("bar" "int")`)
	redshiftDialects().VerifiedStmt(t, `CREATE TABLE "foo" ("1" INT)`)
	redshiftDialects().VerifiedStmt(t, `ALTER TABLE foo ADD CONSTRAINT "bar" PRIMARY KEY (baz)`)
}

func TestSharp(t *testing.T) {
	redshiftDialects().VerifiedOnlySelect(t, "SELECT #_of_values")
}

func TestCreateViewWithNoSchemaBinding(t *testing.T) {
	redshiftAndGeneric().VerifiedStmt(t, "CREATE VIEW myevent AS SELECT eventname FROM event WITH NO SCHEMA BINDING")
}

func TestRedshiftJsonPath(t *testing.T) {
	sql := "SELECT cust.c_orders[0].o_orderkey FROM customer_orders_lineitem"
	redshiftDialects().VerifiedOnlySelect(t, sql)

	sql2 := "SELECT cust.c_orders[0]['id'] FROM customer_orders_lineitem"
	redshiftDialects().VerifiedOnlySelect(t, sql2)

	sql3 := "SELECT db1.sc1.tbl1.col1[0]['id'] FROM customer_orders_lineitem"
	redshiftDialects().VerifiedOnlySelect(t, sql3)

	sql4 := `SELECT db1.sc1.tbl1.col1[0]."id" FROM customer_orders_lineitem`
	redshiftDialects().VerifiedOnlySelect(t, sql4)
}

func TestParseJsonPathFrom(t *testing.T) {
	redshiftDialects().VerifiedOnlySelect(t, "SELECT * FROM src[0].a AS a")
	redshiftDialects().VerifiedOnlySelect(t, "SELECT * FROM src[0].a[1].b AS a")
	redshiftDialects().VerifiedOnlySelect(t, "SELECT * FROM src.a.b")
}

func TestParseSelectNumberedColumns(t *testing.T) {
	redshiftAndGeneric().VerifiedStmt(t, `SELECT 1 AS "1" FROM a`)
	redshiftAndGeneric().VerifiedStmt(t, `SELECT 1 AS "1abc" FROM a`)
}

func TestParseNestedQuotedIdentifier(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, `SELECT 1 AS ["1"] FROM a`)
	redshiftDialects().VerifiedStmt(t, `SELECT 1 AS ["[="] FROM a`)
	redshiftDialects().VerifiedStmt(t, `SELECT 1 AS ["=]"] FROM a`)
	redshiftDialects().VerifiedStmt(t, `SELECT 1 AS ["a[b]"] FROM a`)
	// trim spaces
	redshiftDialects().OneStatementParsesTo(t, `SELECT 1 AS [ " 1 " ]`, `SELECT 1 AS [" 1 "]`)
	// invalid query
	err := redshiftDialects().ParseSQL(t, `SELECT 1 AS ["1]`)
	// Should fail to parse
	_ = err
}

func TestParseExtractSingleQuotes(t *testing.T) {
	sql := "SELECT EXTRACT('month' FROM my_timestamp) FROM my_table"
	redshiftDialects().VerifiedStmt(t, sql)
}

func TestParseStringLiteralBackslashEscape(t *testing.T) {
	redshiftDialects().OneStatementParsesTo(t, `SELECT 'l\'auto'`, "SELECT 'l''auto'")
}

func TestParseUtf8MultibyteIdents(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "SELECT 🚀.city AS 🎸 FROM customers AS 🚀")
}

func TestParseVacuum(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "VACUUM FULL")
	redshiftDialects().VerifiedStmt(t, "VACUUM tbl")
	redshiftDialects().VerifiedStmt(t, "VACUUM FULL SORT ONLY DELETE ONLY REINDEX RECLUSTER db1.sc1.tbl1 TO 20 PERCENT BOOST")
}

func TestCreateTableDiststyleDistkey(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "CREATE TEMPORARY TABLE tmp_sbk_summary_pp DISTSTYLE KEY DISTKEY(bet_id) AS SELECT 1 AS bet_id")
}

func TestCreateTableDiststyle(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE t1 (c1 INT) DISTSTYLE AUTO")
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE t1 (c1 INT) DISTSTYLE EVEN")
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE t1 (c1 INT) DISTSTYLE KEY DISTKEY(c1)")
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE t1 (c1 INT) DISTSTYLE ALL")
}

func TestCopyCredentials(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "COPY t1 FROM 's3://bucket/file.csv' CREDENTIALS 'aws_access_key_id=AK;aws_secret_access_key=SK' CSV")
}

func TestCreateTableSortkey(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE t1 (c1 INT, c2 INT, c3 TIMESTAMP) SORTKEY(c3)")
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE t1 (c1 INT, c2 INT) SORTKEY(c1, c2)")
}

func TestCreateTableDistkeySortkeyWithCtas(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE t1 DISTKEY(1) SORTKEY(1, 3) AS SELECT eventid, venueid, dateid, eventname FROM event")
}

func TestCreateTableDiststyleDistkeySortkey(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE t1 (c1 INT, c2 INT) DISTSTYLE KEY DISTKEY(c1) SORTKEY(c1, c2)")
}

func TestAlterTableAlterSortkey(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "ALTER TABLE users ALTER SORTKEY(created_at)")
	redshiftDialects().VerifiedStmt(t, "ALTER TABLE users ALTER SORTKEY(c1, c2)")
}

func TestCreateTableBackup(t *testing.T) {
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE public.users (id INT, name VARCHAR(255)) BACKUP YES")
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE staging.events (event_id INT) BACKUP NO")
	redshiftDialects().VerifiedStmt(t, "CREATE TABLE public.users_backup_test BACKUP YES DISTSTYLE AUTO AS SELECT id, name, email FROM public.users")
}
