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

// Package snowflake contains Snowflake-specific SQL parsing tests.
// These tests are ported from tests/sqlparser_snowflake.rs in the Rust implementation.
package snowflake

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// snowflakeDialect returns a TestedDialects with only Snowflake dialect
func snowflakeDialect() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			snowflake.NewSnowflakeDialect(),
		},
	}
}

// snowflakeAndGeneric returns a TestedDialects with Snowflake and Generic dialects
func snowflakeAndGeneric() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			snowflake.NewSnowflakeDialect(),
			generic.NewGenericDialect(),
		},
	}
}

// TestSnowflakeCreateTable tests basic CREATE TABLE parsing
// Reference: tests/sqlparser_snowflake.rs:37
func TestSnowflakeCreateTable(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "CREATE TABLE _my_$table (am00unt number)"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateSecureView tests SECURE VIEW creation
// Reference: tests/sqlparser_snowflake.rs:48
func TestSnowflakeCreateSecureView(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"CREATE SECURE VIEW v AS SELECT 1",
		"CREATE SECURE MATERIALIZED VIEW v AS SELECT 1",
		"CREATE OR REPLACE SECURE VIEW v AS SELECT 1",
		"CREATE OR REPLACE SECURE MATERIALIZED VIEW v AS SELECT 1",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateOrReplaceTable tests CREATE OR REPLACE TABLE
// Reference: tests/sqlparser_snowflake.rs:75
func TestSnowflakeCreateOrReplaceTable(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE OR REPLACE TABLE my_table (a number)"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableCopyGrants tests COPY GRANTS option
// Reference: tests/sqlparser_snowflake.rs:89
func TestSnowflakeCreateTableCopyGrants(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE OR REPLACE TABLE my_table (a number) COPY GRANTS"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableEnableSchemaEvolution tests ENABLE_SCHEMA_EVOLUTION option
// Reference: tests/sqlparser_snowflake.rs:144
func TestSnowflakeCreateTableEnableSchemaEvolution(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE TABLE my_table (a number) ENABLE_SCHEMA_EVOLUTION=TRUE"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableChangeTracking tests CHANGE_TRACKING option
// Reference: tests/sqlparser_snowflake.rs:160
func TestSnowflakeCreateTableChangeTracking(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE TABLE my_table (a number) CHANGE_TRACKING=TRUE"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableDataRetentionTime tests DATA_RETENTION_TIME_IN_DAYS option
// Reference: tests/sqlparser_snowflake.rs:176
func TestSnowflakeCreateTableDataRetentionTime(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE TABLE my_table (a number) DATA_RETENTION_TIME_IN_DAYS=5"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTransientTable tests CREATE TRANSIENT TABLE
// Reference: tests/sqlparser_snowflake.rs:374
func TestSnowflakeCreateTransientTable(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "CREATE TRANSIENT TABLE CUSTOMER (id INT, name VARCHAR(255))"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableColumnComment tests column COMMENT option
// Reference: tests/sqlparser_snowflake.rs:388
func TestSnowflakeCreateTableColumnComment(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE TABLE my_table (a STRING COMMENT 'some comment')"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableOnCommit tests ON COMMIT options
// Reference: tests/sqlparser_snowflake.rs:410
func TestSnowflakeCreateTableOnCommit(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		`CREATE LOCAL TEMPORARY TABLE "AAA"."foo" ("bar" INTEGER) ON COMMIT PRESERVE ROWS`,
		`CREATE TABLE "AAA"."foo" ("bar" INTEGER) ON COMMIT DELETE ROWS`,
		`CREATE TABLE "AAA"."foo" ("bar" INTEGER) ON COMMIT DROP`,
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateLocalTable tests CREATE LOCAL TABLE
// Reference: tests/sqlparser_snowflake.rs:419
func TestSnowflakeCreateLocalTable(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE LOCAL TABLE my_table (a INT)"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateGlobalTable tests CREATE GLOBAL TABLE
// Reference: tests/sqlparser_snowflake.rs:438
func TestSnowflakeCreateGlobalTable(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE GLOBAL TABLE my_table (a INT)"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableIfNotExists tests IF NOT EXISTS
// Reference: tests/sqlparser_snowflake.rs:490
func TestSnowflakeCreateTableIfNotExists(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE TABLE IF NOT EXISTS my_table (a INT)"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableClusterBy tests CLUSTER BY option
// Reference: tests/sqlparser_snowflake.rs:526
func TestSnowflakeCreateTableClusterBy(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE TABLE my_table (a INT) CLUSTER BY (a, b, my_func(c))"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableComment tests COMMENT option
// Reference: tests/sqlparser_snowflake.rs:561
func TestSnowflakeCreateTableComment(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE TABLE my_table (a INT) COMMENT = 'some comment'"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableAutoincrement tests AUTOINCREMENT/IDENTITY columns
// Reference: tests/sqlparser_snowflake.rs:622
func TestSnowflakeCreateTableAutoincrement(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE TABLE my_table (" +
		"a INT AUTOINCREMENT ORDER, " +
		"b INT AUTOINCREMENT(100, 1) NOORDER, " +
		"c INT IDENTITY, " +
		"d INT IDENTITY START 100 INCREMENT 1 ORDER" +
		")"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateTableCollatedColumn tests COLLATE option
// Reference: tests/sqlparser_snowflake.rs:713
func TestSnowflakeCreateTableCollatedColumn(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "CREATE TABLE my_table (a TEXT COLLATE 'de_DE')"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateIcebergTable tests CREATE ICEBERG TABLE
// Reference: tests/sqlparser_snowflake.rs:1009
func TestSnowflakeCreateIcebergTable(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE ICEBERG TABLE my_table (a INT) BASE_LOCATION='relative_path'"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCreateDynamicTable tests CREATE DYNAMIC TABLE
// Reference: tests/sqlparser_snowflake.rs:1123
func TestSnowflakeCreateDynamicTable(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"CREATE OR REPLACE DYNAMIC TABLE my_dynamic_table TARGET_LAG='20 minutes' WAREHOUSE=mywh AS SELECT product_id, product_name FROM staging_table",
		"CREATE DYNAMIC ICEBERG TABLE my_dynamic_table (date TIMESTAMP_NTZ, id NUMBER, content STRING) EXTERNAL_VOLUME='my_external_volume' CATALOG='SNOWFLAKE' BASE_LOCATION='my_iceberg_table' TARGET_LAG='20 minutes' WAREHOUSE=mywh AS SELECT product_id, product_name FROM staging_table",
		"CREATE DYNAMIC TABLE my_dynamic_table (date TIMESTAMP_NTZ, id NUMBER, content VARIANT) CLUSTER BY (date, id) TARGET_LAG='20 minutes' WAREHOUSE=mywh AS SELECT product_id, product_name FROM staging_table",
		"CREATE DYNAMIC TABLE my_dynamic_table TARGET_LAG='DOWNSTREAM' WAREHOUSE=mywh INITIALIZE=ON_SCHEDULE REQUIRE USER AS SELECT product_id, product_name FROM staging_table",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateViewWithComment tests CREATE VIEW with COMMENT
// Reference: tests/sqlparser_snowflake.rs:1071
func TestSnowflakeCreateViewWithComment(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE OR REPLACE VIEW v COMMENT = 'hello, world' AS SELECT 1"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeDerivedTableInParenthesis tests parenthesized derived tables
// Reference: tests/sqlparser_snowflake.rs:1196
func TestSnowflakeDerivedTableInParenthesis(t *testing.T) {
	dialects := snowflakeAndGeneric()
	dialects.OneStatementParsesTo(t, "SELECT * FROM ((SELECT 1) AS t)", "SELECT * FROM (SELECT 1) AS t")
	dialects.OneStatementParsesTo(t, "SELECT * FROM (((SELECT 1) AS t))", "SELECT * FROM (SELECT 1) AS t")
}

// TestSnowflakeSingleTableInParenthesis tests parenthesized table names
// Reference: tests/sqlparser_snowflake.rs:1210
func TestSnowflakeSingleTableInParenthesis(t *testing.T) {
	dialects := snowflakeAndGeneric()
	dialects.OneStatementParsesTo(t, "SELECT * FROM (a NATURAL JOIN (b))", "SELECT * FROM (a NATURAL JOIN b)")
	dialects.OneStatementParsesTo(t, "SELECT * FROM (a NATURAL JOIN ((b)))", "SELECT * FROM (a NATURAL JOIN b)")
}

// TestSnowflakeParseArray tests ARRAY type parsing
// Reference: tests/sqlparser_snowflake.rs:1271
func TestSnowflakeParseArray(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "SELECT CAST(a AS ARRAY) FROM customer"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeLateralFlatten tests LATERAL FLATTEN parsing
// Reference: tests/sqlparser_snowflake.rs:1287
func TestSnowflakeLateralFlatten(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		`SELECT * FROM TABLE(FLATTEN(input => parse_json('{"a":1, "b":[77,88]}'), outer => true)) AS f`,
		`SELECT emp.employee_ID, emp.last_name, index, value AS project_name FROM employees AS emp, LATERAL FLATTEN(INPUT => emp.project_names) AS proj_names`,
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeSemiStructuredDataTraversal tests semi-structured data access
// Reference: tests/sqlparser_snowflake.rs:1294
func TestSnowflakeSemiStructuredDataTraversal(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		`SELECT a[2 + 2] FROM t`,
		`SELECT a[0].foo.bar`,
		`SELECT a:b::ARRAY[1]`,
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeArrayAggFunc tests ARRAY_AGG function
// Reference: tests/sqlparser_snowflake.rs:1468
func TestSnowflakeArrayAggFunc(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT ARRAY_AGG(x) WITHIN GROUP (ORDER BY x) AS a FROM T",
		"SELECT ARRAY_AGG(DISTINCT x) WITHIN GROUP (ORDER BY x ASC) FROM tbl",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeSelectWildcardWithExclude tests SELECT * EXCLUDE
// Reference: tests/sqlparser_snowflake.rs:1500
func TestSnowflakeSelectWildcardWithExclude(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"SELECT * EXCLUDE (col_a) FROM data",
		"SELECT name.* EXCLUDE department_id FROM employee_table",
		"SELECT * EXCLUDE (department_id, employee_id) FROM employee_table",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeSelectWildcardWithRename tests SELECT * RENAME
// Reference: tests/sqlparser_snowflake.rs:1536
func TestSnowflakeSelectWildcardWithRename(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"SELECT * RENAME col_a AS col_b FROM data",
		"SELECT name.* RENAME (department_id AS new_dep, employee_id AS new_emp) FROM employee_table",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeSelectWildcardWithReplace tests SELECT * REPLACE
// Reference: tests/sqlparser_snowflake.rs:1571
func TestSnowflakeSelectWildcardWithReplace(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "SELECT * REPLACE (col_z || col_z AS col_z) RENAME (col_z AS col_zz) FROM data"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeSelectWildcardWithExcludeAndRename tests SELECT * EXCLUDE ... RENAME
// Reference: tests/sqlparser_snowflake.rs:1609
func TestSnowflakeSelectWildcardWithExcludeAndRename(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "SELECT * EXCLUDE col_z RENAME col_a AS col_b FROM data"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeAlterTableSwapWith tests ALTER TABLE SWAP WITH
// Reference: tests/sqlparser_snowflake.rs:1636
func TestSnowflakeAlterTableSwapWith(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "ALTER TABLE tab1 SWAP WITH tab2"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeAlterTableClustering tests ALTER TABLE CLUSTERING
// Reference: tests/sqlparser_snowflake.rs:1647
func TestSnowflakeAlterTableClustering(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		`ALTER TABLE tab CLUSTER BY (c1, "c2", TO_DATE(c3))`,
		"ALTER TABLE tbl DROP CLUSTERING KEY",
		"ALTER TABLE tbl SUSPEND RECLUSTER",
		"ALTER TABLE tbl RESUME RECLUSTER",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeDropStage tests DROP STAGE
// Reference: tests/sqlparser_snowflake.rs:1691
func TestSnowflakeDropStage(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"DROP STAGE s1",
		"DROP STAGE IF EXISTS s1",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeDeclareCursor tests DECLARE CURSOR
// Reference: tests/sqlparser_snowflake.rs:1718
func TestSnowflakeDeclareCursor(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"DECLARE c1 CURSOR FOR SELECT id, price FROM invoices",
		"DECLARE c1 CURSOR FOR res",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeDeclareResultSet tests DECLARE RESULTSET
// Reference: tests/sqlparser_snowflake.rs:1787
func TestSnowflakeDeclareResultSet(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"DECLARE res RESULTSET DEFAULT 42",
		"DECLARE res RESULTSET := 42",
		"DECLARE res RESULTSET",
		"DECLARE res RESULTSET DEFAULT (SELECT price FROM invoices)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeDeclareException tests DECLARE EXCEPTION
// Reference: tests/sqlparser_snowflake.rs:1841
func TestSnowflakeDeclareException(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"DECLARE ex EXCEPTION (42, 'ERROR')",
		"DECLARE ex EXCEPTION",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeDeclareVariable tests DECLARE variable
// Reference: tests/sqlparser_snowflake.rs:1879
func TestSnowflakeDeclareVariable(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"DECLARE profit TEXT DEFAULT 42",
		"DECLARE profit DEFAULT 42",
		"DECLARE profit TEXT",
		"DECLARE profit",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateStage tests CREATE STAGE
// Reference: tests/sqlparser_snowflake.rs:1974
func TestSnowflakeCreateStage(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"CREATE STAGE s1.s2",
		"CREATE OR REPLACE TEMPORARY STAGE IF NOT EXISTS s1.s2 COMMENT='some-comment'",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateStageWithParams tests CREATE STAGE with stage parameters
// Reference: tests/sqlparser_snowflake.rs:2026
func TestSnowflakeCreateStageWithParams(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE OR REPLACE STAGE my_ext_stage " +
		"URL='s3://load/files/' " +
		"STORAGE_INTEGRATION=myint " +
		"ENDPOINT='<s3_api_compatible_endpoint>' " +
		"CREDENTIALS=(AWS_KEY_ID='1a2b3c' AWS_SECRET_KEY='4x5y6z') " +
		"ENCRYPTION=(MASTER_KEY='key' TYPE='AWS_SSE_KMS')"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCopyInto tests COPY INTO
// Reference: tests/sqlparser_snowflake.rs:2171
func TestSnowflakeCopyInto(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"COPY INTO my_company.emp_basic FROM 'gcs://mybucket/./../a.csv'",
		"COPY INTO 's3://a/b/c/data.parquet' FROM db.sc.tbl PARTITION BY ('date=' || to_varchar(dt, 'YYYY-MM-DD') || '/hour=' || to_varchar(date_part(hour, ts)))",
		"COPY INTO 's3://a/b/c/data.parquet' FROM (SELECT * FROM tbl)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCopyIntoWithStageParams tests COPY INTO with stage parameters
// Reference: tests/sqlparser_snowflake.rs:2262
func TestSnowflakeCopyIntoWithStageParams(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "COPY INTO my_company.emp_basic FROM 's3://load/files/' " +
		"STORAGE_INTEGRATION=myint " +
		"ENDPOINT='<s3_api_compatible_endpoint>' " +
		"CREDENTIALS=(AWS_KEY_ID='1a2b3c' AWS_SECRET_KEY='4x5y6z') " +
		"ENCRYPTION=(MASTER_KEY='key' TYPE='AWS_SSE_KMS')"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCopyIntoWithFilesAndPattern tests COPY INTO with FILES and PATTERN
// Reference: tests/sqlparser_snowflake.rs:2347
func TestSnowflakeCopyIntoWithFilesAndPattern(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "COPY INTO my_company.emp_basic FROM 'gcs://mybucket/./../a.csv' AS some_alias " +
		"FILES = ('file1.json', 'file2.json') " +
		"PATTERN = '.*employees0[1-5].csv.gz' " +
		"VALIDATION_MODE = RETURN_7_ROWS"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCopyIntoWithTransformations tests COPY INTO with transformations
// Reference: tests/sqlparser_snowflake.rs:2375
func TestSnowflakeCopyIntoWithTransformations(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "COPY INTO my_company.emp_basic FROM " +
		"(SELECT t1.$1:st AS st, $1:index, t2.$1, 4, '5' AS const_str FROM @schema.general_finished AS T) " +
		"FILES = ('file1.json', 'file2.json') " +
		"PATTERN = '.*employees0[1-5].csv.gz' " +
		"VALIDATION_MODE = RETURN_7_ROWS"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeStageObjectNames tests stage object names
// Reference: tests/sqlparser_snowflake.rs:2557
func TestSnowflakeStageObjectNames(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"COPY INTO @namespace.%table_name FROM 'gcs://mybucket/./../a.csv'",
		"COPY INTO @namespace.%table_name/path FROM 'gcs://mybucket/./../a.csv'",
		"COPY INTO @namespace.stage_name/path FROM 'gcs://mybucket/./../a.csv'",
		"COPY INTO @~/path FROM 'gcs://mybucket/./../a.csv'",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeTrim tests TRIM function
// Reference: tests/sqlparser_snowflake.rs:2697
func TestSnowflakeTrim(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		`SELECT customer_id, TRIM(sub_items.value:item_price_id, '"', "a") AS item_price_id FROM models_staging.subscriptions`,
		"SELECT TRIM('xyz', 'a')",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeNumberPlaceholder tests numeric placeholders
// Reference: tests/sqlparser_snowflake.rs:2726
func TestSnowflakeNumberPlaceholder(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "SELECT :1"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeParsePositionNotFunctionColumns tests POSITION as identifier
// Reference: tests/sqlparser_snowflake.rs:2740
func TestSnowflakeParsePositionNotFunctionColumns(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "SELECT position FROM tbl1 WHERE position NOT IN ('first', 'last')"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeSubqueryFunctionArgument tests subquery as function argument
// Reference: tests/sqlparser_snowflake.rs:2746
func TestSnowflakeSubqueryFunctionArgument(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT parse_json(SELECT '{}')",
		"SELECT parse_json(WITH q AS (SELECT '{}' AS foo) SELECT foo FROM q)",
		"SELECT func(SELECT 1, 2)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeParseDivision tests division parsing
// Reference: tests/sqlparser_snowflake.rs:2761
func TestSnowflakeParseDivision(t *testing.T) {
	dialects := snowflakeAndGeneric()
	dialects.OneStatementParsesTo(t, "SELECT field/1000 FROM tbl1", "SELECT field / 1000 FROM tbl1")
}

// TestSnowflakeParseTop tests TOP clause
// Reference: tests/sqlparser_snowflake.rs:2781
func TestSnowflakeParseTop(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "SELECT TOP 4 c1 FROM testtable"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeExtractCustomPart tests EXTRACT with custom date part
// Reference: tests/sqlparser_snowflake.rs:2789
func TestSnowflakeExtractCustomPart(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "SELECT EXTRACT(eod FROM d)"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeExtractComma tests EXTRACT with comma syntax
// Reference: tests/sqlparser_snowflake.rs:2802
func TestSnowflakeExtractComma(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := "SELECT EXTRACT(HOUR, d)"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeCommaOuterJoin tests outer join with (+) syntax
// Reference: tests/sqlparser_snowflake.rs:2831
func TestSnowflakeCommaOuterJoin(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT t1.c1, t2.c2 FROM t1, t2 WHERE t1.c1 = t2.c2 (+)",
		"SELECT t1.c1, t2.c2 FROM t1, t2 WHERE c1 = c2 (+)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeTrailingCommas tests trailing commas in SELECT
// Reference: tests/sqlparser_snowflake.rs:2890
func TestSnowflakeTrailingCommas(t *testing.T) {
	dialects := snowflakeDialect()
	dialects.OneStatementParsesTo(t, "SELECT 1, 2, FROM t", "SELECT 1, 2 FROM t")
}

// TestSnowflakeSelectWildcardWithIlike tests SELECT * ILIKE
// Reference: tests/sqlparser_snowflake.rs:2895
func TestSnowflakeSelectWildcardWithIlike(t *testing.T) {
	dialects := snowflakeAndGeneric()
	sql := `SELECT * ILIKE '%id%' FROM tbl`
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeFirstValueIgnoreNulls tests FIRST_VALUE with IGNORE NULLS
// Reference: tests/sqlparser_snowflake.rs:2934
func TestSnowflakeFirstValueIgnoreNulls(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "SELECT FIRST_VALUE(column2 IGNORE NULLS) OVER (PARTITION BY column1 ORDER BY column2) FROM some_table"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakePivot tests PIVOT clause
// Reference: tests/sqlparser_snowflake.rs:2943
func TestSnowflakePivot(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT * FROM quarterly_sales PIVOT(SUM(amount) FOR quarter IN ('2023_Q1', '2023_Q2', '2023_Q3', '2023_Q4', '2024_Q1') DEFAULT ON NULL (0)) ORDER BY empid",
		"SELECT * FROM quarterly_sales PIVOT(SUM(amount) FOR quarter IN (SELECT DISTINCT quarter FROM ad_campaign_types_by_quarter WHERE television = true ORDER BY quarter)) ORDER BY empid",
		"SELECT * FROM quarterly_sales PIVOT(SUM(amount) FOR quarter IN (ANY ORDER BY quarter)) ORDER BY empid",
		"SELECT * FROM sales_data PIVOT(SUM(total_sales) FOR fis_quarter IN (ANY)) WHERE fis_year IN (2023) ORDER BY region",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeAsofJoins tests ASOF JOIN
// Reference: tests/sqlparser_snowflake.rs:2996
func TestSnowflakeAsofJoins(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"SELECT * FROM trades_unixtime AS tu ASOF JOIN quotes_unixtime AS qu MATCH_CONDITION (tu.trade_time >= qu.quote_time)",
		"SELECT t.stock_symbol, t.trade_time, t.quantity, q.quote_time, q.price FROM trades AS t ASOF JOIN quotes AS q MATCH_CONDITION (t.trade_time >= quote_time) ON t.stock_symbol = q.stock_symbol ORDER BY t.stock_symbol",
		"SELECT * FROM snowtime AS s ASOF JOIN raintime AS r MATCH_CONDITION (s.observed >= r.observed) ON s.state = r.state ASOF JOIN preciptime AS p MATCH_CONDITION (s.observed >= p.observed) ON s.state = p.state ORDER BY s.observed",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeExplainDescribe tests EXPLAIN/DESCRIBE
// Reference: tests/sqlparser_snowflake.rs:3084
func TestSnowflakeExplainDescribe(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"DESCRIBE test.table",
		"DESCRIBE TABLE test.table",
		"DESC test.table",
		"DESC TABLE test.table",
		"EXPLAIN TABLE test_identifier",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeUse tests USE statements
// Reference: tests/sqlparser_snowflake.rs:3114
func TestSnowflakeUse(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"USE mydb",
		"USE 'mydb'",
		"USE mydb.my_schema",
		"USE DATABASE 'my_database'",
		"USE SCHEMA 'my_schema'",
		"USE ROLE 'my_role'",
		"USE WAREHOUSE 'my_wh'",
		"USE SECONDARY ROLES ALL",
		"USE SECONDARY ROLES NONE",
		"USE SECONDARY ROLES r1, r2, r3",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeViewCommentOption tests CREATE VIEW COMMENT option
// Reference: tests/sqlparser_snowflake.rs:3222
func TestSnowflakeViewCommentOption(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"CREATE OR REPLACE VIEW v (a) COMMENT = 'Comment' AS SELECT a FROM t",
		"CREATE OR REPLACE VIEW v (a COMMENT 'a comment', b, c COMMENT 'c comment') COMMENT = 'Comment' AS SELECT a FROM t",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeShowDatabases tests SHOW DATABASES
// Reference: tests/sqlparser_snowflake.rs:3289
func TestSnowflakeShowDatabases(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SHOW DATABASES",
		"SHOW TERSE DATABASES",
		"SHOW DATABASES HISTORY",
		"SHOW DATABASES LIKE '%abc%'",
		"SHOW DATABASES STARTS WITH 'demo_db'",
		"SHOW DATABASES LIMIT 12",
		"SHOW DATABASES HISTORY LIKE '%aa' STARTS WITH 'demo' LIMIT 20 FROM 'abc'",
		"SHOW DATABASES IN ACCOUNT abc",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeShowSchemas tests SHOW SCHEMAS
// Reference: tests/sqlparser_snowflake.rs:3302
func TestSnowflakeShowSchemas(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SHOW SCHEMAS",
		"SHOW TERSE SCHEMAS",
		"SHOW SCHEMAS IN ACCOUNT",
		"SHOW SCHEMAS IN ACCOUNT abc",
		"SHOW SCHEMAS IN DATABASE",
		"SHOW SCHEMAS IN DATABASE xyz",
		"SHOW SCHEMAS HISTORY LIKE '%xa%'",
		"SHOW SCHEMAS STARTS WITH 'abc' LIMIT 20",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeShowObjects tests SHOW OBJECTS
// Reference: tests/sqlparser_snowflake.rs:3315
func TestSnowflakeShowObjects(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SHOW OBJECTS",
		"SHOW OBJECTS IN abc",
		"SHOW OBJECTS LIKE '%test%' IN abc",
		"SHOW OBJECTS IN ACCOUNT",
		"SHOW OBJECTS IN DATABASE",
		"SHOW OBJECTS IN DATABASE abc",
		"SHOW OBJECTS IN SCHEMA",
		"SHOW OBJECTS IN SCHEMA abc",
		"SHOW TERSE OBJECTS",
		"SHOW TERSE OBJECTS LIKE '%test%' IN abc STARTS WITH 'b' LIMIT 10 FROM 'x'",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeShowTables tests SHOW TABLES
// Reference: tests/sqlparser_snowflake.rs:3357
func TestSnowflakeShowTables(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SHOW TABLES",
		"SHOW TERSE TABLES",
		"SHOW TABLES IN ACCOUNT",
		"SHOW TABLES IN DATABASE",
		"SHOW TABLES IN DATABASE xyz",
		"SHOW TABLES IN SCHEMA",
		"SHOW TABLES IN SCHEMA xyz",
		"SHOW TABLES HISTORY LIKE '%xa%'",
		"SHOW TABLES STARTS WITH 'abc' LIMIT 20",
		"SHOW EXTERNAL TABLES",
		"SHOW EXTERNAL TABLES IN ACCOUNT",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeShowViews tests SHOW VIEWS
// Reference: tests/sqlparser_snowflake.rs:3379
func TestSnowflakeShowViews(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SHOW VIEWS",
		"SHOW TERSE VIEWS",
		"SHOW VIEWS IN ACCOUNT",
		"SHOW VIEWS IN DATABASE",
		"SHOW VIEWS IN DATABASE xyz",
		"SHOW VIEWS IN SCHEMA",
		"SHOW VIEWS IN SCHEMA xyz",
		"SHOW VIEWS STARTS WITH 'abc' LIMIT 20",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeDoubleDotNotation tests double dot notation
// Reference: tests/sqlparser_snowflake.rs:3416
func TestSnowflakeDoubleDotNotation(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT * FROM db_name..table_name",
		"SELECT * FROM x, y..z JOIN a..b AS b ON x.id = b.id",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeInsertOverwrite tests INSERT OVERWRITE
// Reference: tests/sqlparser_snowflake.rs:3448
func TestSnowflakeInsertOverwrite(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "INSERT OVERWRITE INTO schema.table SELECT a FROM b"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeTableSample tests TABLESAMPLE
// Reference: tests/sqlparser_snowflake.rs:3454
func TestSnowflakeTableSample(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"SELECT * FROM testtable SAMPLE (10)",
		"SELECT * FROM testtable TABLESAMPLE (10)",
		"SELECT * FROM testtable AS t TABLESAMPLE BERNOULLI (10)",
		"SELECT * FROM testtable AS t TABLESAMPLE ROW (10)",
		"SELECT * FROM testtable AS t TABLESAMPLE ROW (10 ROWS)",
		"SELECT * FROM testtable TABLESAMPLE BLOCK (3) SEED (82)",
		"SELECT * FROM testtable TABLESAMPLE SYSTEM (3) REPEATABLE (82)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeMultiTableInsertUnconditional tests unconditional multi-table INSERT
// Reference: tests/sqlparser_snowflake.rs:3487
func TestSnowflakeMultiTableInsertUnconditional(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"INSERT ALL INTO t1 SELECT n1, n2, n3 FROM src",
		"INSERT ALL INTO t1 INTO t2 SELECT n1, n2, n3 FROM src",
		"INSERT ALL INTO t1 (c1, c2, c3) SELECT n1, n2, n3 FROM src",
		"INSERT ALL INTO t1 (c1, c2, c3) VALUES (n2, n1, DEFAULT) SELECT n1, n2, n3 FROM src",
		"INSERT ALL INTO t1 INTO t1 (c1, c2, c3) VALUES (n2, n1, DEFAULT) INTO t2 (c1, c2, c3) INTO t2 VALUES (n3, n2, n1) SELECT n1, n2, n3 FROM src",
		"INSERT OVERWRITE ALL INTO t1 INTO t2 SELECT n1, n2, n3 FROM src",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeMultiTableInsertConditional tests conditional multi-table INSERT
// Reference: tests/sqlparser_snowflake.rs:3513
func TestSnowflakeMultiTableInsertConditional(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"INSERT ALL WHEN n1 > 100 THEN INTO t1 SELECT n1 FROM src",
		"INSERT ALL WHEN n1 > 100 THEN INTO t1 WHEN n1 > 10 THEN INTO t2 SELECT n1 FROM src",
		"INSERT ALL WHEN n1 > 10 THEN INTO t1 INTO t2 SELECT n1 FROM src",
		"INSERT ALL WHEN n1 > 100 THEN INTO t1 ELSE INTO t2 SELECT n1 FROM src",
		"INSERT ALL WHEN n1 > 100 THEN INTO t1 WHEN n1 > 10 THEN INTO t1 INTO t2 ELSE INTO t2 SELECT n1 FROM src",
		"INSERT FIRST WHEN n1 > 100 THEN INTO t1 WHEN n1 > 10 THEN INTO t1 INTO t2 ELSE INTO t2 SELECT n1 FROM src",
		"INSERT OVERWRITE ALL WHEN n1 > 100 THEN INTO t1 ELSE INTO t2 SELECT n1 FROM src",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeMultiTableInsertWithValues tests multi-table INSERT with VALUES
// Reference: tests/sqlparser_snowflake.rs:3550
func TestSnowflakeMultiTableInsertWithValues(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"INSERT ALL INTO t1 VALUES (n1, n2) SELECT n1, n2 FROM src",
		"INSERT ALL INTO t1 (c1, c2, c3) VALUES (n1, n2, DEFAULT) SELECT n1, n2 FROM src",
		"INSERT ALL INTO t1 (c1, c2, c3) VALUES (n1, NULL, n2) SELECT n1, n2 FROM src",
		"INSERT ALL INTO t1 VALUES ($1, $2) SELECT 1, 50 AS an_alias",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeLsAndRm tests LIST and REMOVE statements
// Reference: tests/sqlparser_snowflake.rs:3808
func TestSnowflakeLsAndRm(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"LIST @~",
		"LIST @SNOWFLAKE_KAFKA_CONNECTOR_externalDataLakeSnowflakeConnector_STAGE_call_tracker_stream/",
		`LIST @"STAGE_WITH_QUOTES"`,
		"REMOVE @my_csv_stage/analysis/ PATTERN='.*data_0.*'",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeChangesClause tests CHANGES clause
// Reference: tests/sqlparser_snowflake.rs:4023
func TestSnowflakeChangesClause(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		`SELECT a FROM "PCH_ODS_FIDELIO"."SRC_VW_SYS_ACC_MASTER" CHANGES(INFORMATION => DEFAULT) AT(TIMESTAMP => TO_TIMESTAMP_TZ('2026-02-18 11:23:19.660000000')) END(TIMESTAMP => TO_TIMESTAMP_TZ('2026-02-18 11:38:30.211000000'))`,
		"SELECT a FROM t CHANGES(INFORMATION => DEFAULT) AT(TIMESTAMP => TO_TIMESTAMP_TZ('2026-02-18 11:23:19.660000000'))",
		"SELECT a FROM t CHANGES(INFORMATION => APPEND_ONLY) AT(TIMESTAMP => TO_TIMESTAMP_TZ('2026-01-01 00:00:00'))",
		"SELECT a FROM t CHANGES(INFORMATION => DEFAULT) AT(OFFSET => -60)",
		"SELECT a FROM t CHANGES(INFORMATION => DEFAULT) AT(STATEMENT => '8e5d0ca9-005e-44e6-b858-a8f5b37c5726')",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeGrantAccountGlobalPrivileges tests GRANT account global privileges
// Reference: tests/sqlparser_snowflake.rs:4049
func TestSnowflakeGrantAccountGlobalPrivileges(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"GRANT ALL ON ACCOUNT TO ROLE role1",
		"GRANT ALL PRIVILEGES ON ACCOUNT TO ROLE role1",
		"GRANT CREATE DATABASE ON ACCOUNT TO ROLE role1",
		"GRANT CREATE WAREHOUSE ON ACCOUNT TO ROLE role1 WITH GRANT OPTION",
		"GRANT APPLY MASKING POLICY ON ACCOUNT TO ROLE role1",
		"GRANT EXECUTE TASK ON ACCOUNT TO ROLE role1",
		"GRANT MANAGE GRANTS ON ACCOUNT TO ROLE role1",
		"GRANT MONITOR USAGE ON ACCOUNT TO ROLE role1",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeGrantRoleTo tests GRANT ROLE TO
// Reference: tests/sqlparser_snowflake.rs:4193
func TestSnowflakeGrantRoleTo(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"GRANT ROLE r1 TO ROLE r2",
		"GRANT ROLE r1 TO USER u1",
		"GRANT DATABASE ROLE r1 TO ROLE r2",
		"GRANT DATABASE ROLE db1.sc1.r1 TO ROLE db1.sc1.r2",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeAlterSession tests ALTER SESSION
// Reference: tests/sqlparser_snowflake.rs:4204
func TestSnowflakeAlterSession(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"ALTER SESSION SET AUTOCOMMIT=TRUE",
		"ALTER SESSION SET AUTOCOMMIT=false QUERY_TAG='tag'",
		"ALTER SESSION UNSET AUTOCOMMIT",
		"ALTER SESSION UNSET AUTOCOMMIT, QUERY_TAG",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeFetchClause tests FETCH clause
// Reference: tests/sqlparser_snowflake.rs:4717
func TestSnowflakeFetchClause(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT c1 FROM fetch_test FETCH FIRST 2 ROWS ONLY",
		"SELECT c1 FROM fetch_test FETCH 2",
		"SELECT c1 FROM fetch_test FETCH FIRST 2",
		"SELECT c1 FROM fetch_test FETCH NEXT 2",
		"SELECT c1 FROM fetch_test FETCH 2 ROW",
		"SELECT c1 FROM fetch_test FETCH FIRST 2 ROWS",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateViewCopyGrants tests CREATE VIEW COPY GRANTS
// Reference: tests/sqlparser_snowflake.rs:4757
func TestSnowflakeCreateViewCopyGrants(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"CREATE OR REPLACE VIEW bla COPY GRANTS AS (SELECT * FROM source)",
		"CREATE OR REPLACE SECURE VIEW bla COPY GRANTS AS (SELECT * FROM source)",
		"CREATE OR REPLACE VIEW bla COPY GRANTS (a, b) AS (SELECT a, b FROM source)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateDatabase tests CREATE DATABASE
// Reference: tests/sqlparser_snowflake.rs:4903
func TestSnowflakeCreateDatabase(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"CREATE DATABASE my_db",
		"CREATE OR REPLACE DATABASE my_db",
		"CREATE TRANSIENT DATABASE IF NOT EXISTS my_db",
		"CREATE DATABASE my_db CLONE src_db",
		"CREATE OR REPLACE DATABASE my_db CLONE src_db DATA_RETENTION_TIME_IN_DAYS = 1",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeTimestampNtzWithPrecision tests TIMESTAMP_NTZ with precision
// Reference: tests/sqlparser_snowflake.rs:4954
func TestSnowflakeTimestampNtzWithPrecision(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT CAST('2024-01-01 01:00:00' AS TIMESTAMP_NTZ(1))",
		"SELECT CAST('2024-01-01 01:00:00' AS TIMESTAMP_NTZ(9))",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeDropConstraints tests DROP constraints
// Reference: tests/sqlparser_snowflake.rs:4969
func TestSnowflakeDropConstraints(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"ALTER TABLE tbl DROP PRIMARY KEY",
		"ALTER TABLE tbl DROP FOREIGN KEY k1",
		"ALTER TABLE tbl DROP CONSTRAINT c1",
		"ALTER TABLE tbl DROP PRIMARY KEY CASCADE",
		"ALTER TABLE tbl DROP FOREIGN KEY k1 RESTRICT",
		"ALTER TABLE tbl DROP CONSTRAINT c1 CASCADE",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeAlterDynamicTable tests ALTER DYNAMIC TABLE
// Reference: tests/sqlparser_snowflake.rs:4979
func TestSnowflakeAlterDynamicTable(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"ALTER DYNAMIC TABLE MY_DYNAMIC_TABLE REFRESH",
		"ALTER DYNAMIC TABLE my_database.my_schema.my_dynamic_table REFRESH",
		"ALTER DYNAMIC TABLE my_dyn_table SUSPEND",
		"ALTER DYNAMIC TABLE my_dyn_table RESUME",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeAlterExternalTable tests ALTER EXTERNAL TABLE
// Reference: tests/sqlparser_snowflake.rs:4987
func TestSnowflakeAlterExternalTable(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"ALTER EXTERNAL TABLE some_table REFRESH",
		"ALTER EXTERNAL TABLE some_table REFRESH 'year=2025/month=12/'",
		"ALTER EXTERNAL TABLE IF EXISTS some_table REFRESH",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeTruncateTableIfExists tests TRUNCATE with IF EXISTS
// Reference: tests/sqlparser_snowflake.rs:4996
func TestSnowflakeTruncateTableIfExists(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"TRUNCATE TABLE IF EXISTS my_table",
		"TRUNCATE TABLE my_table",
		"TRUNCATE IF EXISTS my_table",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeSelectDollarColumnFromStage tests selecting $N columns from stage
// Reference: tests/sqlparser_snowflake.rs:5003
func TestSnowflakeSelectDollarColumnFromStage(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT t.$1, t.$2 FROM @mystage1(file_format => 'myformat', pattern => '.*data.*[.]csv.gz') t",
		"SELECT t.$1, t.$2 FROM @mystage1 t",
		"SELECT $1, $2 FROM @mystage1",
		"SELECT $1, $2 FROM @mystage1(file_format => 'myformat')",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeParseErrors tests various parse error cases
// Reference: various tests in sqlparser_snowflake.rs
func TestSnowflakeParseErrors(t *testing.T) {
	dialect := snowflake.NewSnowflakeDialect()

	testCases := []struct {
		sql         string
		expectError bool
	}{
		// Invalid: LOCAL and GLOBAL together
		{"CREATE LOCAL GLOBAL TABLE my_table (a INT)", true},
		{"CREATE GLOBAL LOCAL TABLE my_table (a INT)", true},
		// Invalid: duplicate temporal keywords
		{"CREATE TEMP TEMPORARY TABLE my_table (a INT)", true},
		{"CREATE TEMP VOLATILE TABLE my_table (a INT)", true},
		// Invalid: missing BASE_LOCATION for ICEBERG table
		{"CREATE ICEBERG TABLE my_table (a INT)", true},
		// Invalid: VIEW COMMENT without equals
		{"CREATE OR REPLACE VIEW v COMMENT 'hello, world' AS SELECT 1", true},
	}

	for _, tc := range testCases {
		t.Run(tc.sql, func(t *testing.T) {
			_, err := parser.ParseSQL(dialect, tc.sql)
			if tc.expectError {
				require.Error(t, err, "Expected error for SQL: %s", tc.sql)
			} else {
				require.NoError(t, err, "Unexpected error for SQL: %s", tc.sql)
			}
		})
	}
}

// TestSnowflakeConnectByRoot tests CONNECT_BY_ROOT operator
// Reference: tests/sqlparser_snowflake.rs:4590
func TestSnowflakeConnectByRoot(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT CONNECT_BY_ROOT name AS root_name FROM Tbl1",
		"SELECT CONNECT_BY_ROOT name FROM Tbl2",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeBeginTransaction tests BEGIN TRANSACTION
// Reference: tests/sqlparser_snowflake.rs:4696
func TestSnowflakeBeginTransaction(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"BEGIN TRANSACTION",
		"BEGIN WORK",
		"BEGIN",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateViewWithTags tests CREATE VIEW with TAG
// Reference: tests/sqlparser_snowflake.rs:4736
func TestSnowflakeCreateViewWithTags(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"CREATE VIEW X (COL WITH TAG (pii='email') COMMENT 'foobar') AS SELECT * FROM Y",
		"CREATE VIEW X (COL WITH TAG (foo.bar.baz.pii='email')) AS SELECT * FROM Y",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeCreateViewWithPolicy tests CREATE VIEW with MASKING POLICY
// Reference: tests/sqlparser_snowflake.rs:4750
func TestSnowflakeCreateViewWithPolicy(t *testing.T) {
	dialects := snowflakeDialect()
	sql := "CREATE VIEW X (COL WITH MASKING POLICY foo.bar.baz) AS SELECT * FROM Y"
	dialects.VerifiedStmt(t, sql)
}

// TestSnowflakeIdentifierFunction tests IDENTIFIER() function
// Reference: tests/sqlparser_snowflake.rs:4768
func TestSnowflakeIdentifierFunction(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT identifier('email') FROM customers",
		`SELECT identifier('"Email"') FROM customers`,
		"SELECT identifier('alias1').* FROM tbl AS alias1",
		"CREATE DATABASE IDENTIFIER('tbl')",
		"CREATE SCHEMA IDENTIFIER('db1.sc1')",
		"CREATE TABLE IDENTIFIER('tbl') (id INT)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeTimeTravel tests AT/BEFORE time travel
// Reference: tests/sqlparser_snowflake.rs:4017
func TestSnowflakeTimeTravel(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT * FROM tbl AT(TIMESTAMP => '2024-12-15 00:00:00')",
		"SELECT * FROM tbl BEFORE(TIMESTAMP => '2024-12-15 00:00:00')",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeSubquerySample tests SAMPLE on subqueries
// Reference: tests/sqlparser_snowflake.rs:3470
func TestSnowflakeSubquerySample(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"SELECT * FROM (SELECT * FROM mytable) SAMPLE (10)",
		"SELECT * FROM (SELECT * FROM mytable) SAMPLE (10000 ROWS)",
		"SELECT * FROM (SELECT * FROM mytable) AS t SAMPLE (50 PERCENT)",
		"SELECT * FROM (SELECT * FROM (SELECT report_from FROM mytable) SAMPLE (10000 ROWS)) AS anon_1",
		"SELECT * FROM (SELECT * FROM mytable) SAMPLE (10) SEED (42)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeShowColumns tests SHOW COLUMNS
// Reference: tests/sqlparser_snowflake.rs:3392
func TestSnowflakeShowColumns(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SHOW COLUMNS IN TABLE",
		"SHOW COLUMNS IN TABLE abc",
		"SHOW COLUMNS LIKE '%xyz%' IN TABLE abc",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeAlterIcebergTable tests ALTER ICEBERG TABLE
// Reference: tests/sqlparser_snowflake.rs:1684
func TestSnowflakeAlterIcebergTable(t *testing.T) {
	dialects := snowflakeAndGeneric()

	testCases := []string{
		"ALTER ICEBERG TABLE tbl DROP CLUSTERING KEY",
		"ALTER ICEBERG TABLE tbl SUSPEND RECLUSTER",
		"ALTER ICEBERG TABLE tbl RESUME RECLUSTER",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSnowflakeStageNameWithSpecialChars tests stage names with special characters
// Reference: tests/sqlparser_snowflake.rs:2682
func TestSnowflakeStageNameWithSpecialChars(t *testing.T) {
	dialects := snowflakeDialect()

	testCases := []string{
		"SELECT * FROM @stage/day=18/23.parquet",
		"SELECT * FROM @stage/0:18:23/23.parquet",
		"COPY INTO my_table FROM @stage/day=18/file.parquet",
		"COPY INTO my_table FROM @stage/0:18:23/file.parquet",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}
