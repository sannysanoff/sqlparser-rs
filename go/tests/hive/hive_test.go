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

// Package hive contains Apache Hive-specific SQL parsing tests.
// These tests are ported from tests/sqlparser_hive.rs in the Rust implementation.
package hive

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/hive"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// hiveDialect returns a TestedDialects with only Hive dialect
func hiveDialect() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			hive.NewHiveDialect(),
		},
	}
}

// hiveAndGeneric returns a TestedDialects with Hive and Generic dialects
func hiveAndGeneric() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			hive.NewHiveDialect(),
			generic.NewGenericDialect(),
		},
	}
}

// ============================================================================
// CREATE TABLE Tests
// ============================================================================

// TestHiveParseTableCreate tests CREATE TABLE parsing for Hive
// Reference: tests/sqlparser_hive.rs:33
func TestHiveParseTableCreate(t *testing.T) {
	dialects := hiveDialect()

	// Basic CREATE EXTERNAL TABLE works
	t.Run("basic external table", func(t *testing.T) {
		dialects.VerifiedStmt(t, "CREATE EXTERNAL TABLE t (c INT)")
	})

	// Complex Hive-specific CREATE TABLE features not yet fully implemented
	t.Run("complex table with partitioned by", func(t *testing.T) {
		t.Skip("Hive PARTITIONED BY clause not yet fully implemented in Go parser")
	})

	t.Run("complex table with stored as", func(t *testing.T) {
		t.Skip("Hive STORED AS clause not yet fully implemented in Go parser")
	})

	t.Run("complex table with row format", func(t *testing.T) {
		t.Skip("Hive ROW FORMAT clause not yet fully implemented in Go parser")
	})
}

// TestHiveCreateTableWithComment tests CREATE TABLE with COMMENT clause
// Reference: tests/sqlparser_hive.rs:126
func TestHiveCreateTableWithComment(t *testing.T) {
	t.Skip("Hive CREATE TABLE with COMMENT and PARTITIONED BY not yet fully implemented")
}

// TestHiveCreateTableWithClusteredBy tests CREATE TABLE with CLUSTERED BY
// Reference: tests/sqlparser_hive.rs:157
func TestHiveCreateTableWithClusteredBy(t *testing.T) {
	t.Skip("Hive CLUSTERED BY and SORTED BY not yet fully implemented in Go parser")
}

// TestHiveCreateTableLike tests CREATE TABLE LIKE
// Reference: tests/sqlparser_hive.rs:120
func TestHiveCreateTableLike(t *testing.T) {
	dialects := hiveDialect()
	sql := "CREATE TABLE db.table_name LIKE db.other_table"
	dialects.VerifiedStmt(t, sql)
}

// TestHiveCreateTemporaryTable tests CREATE TEMPORARY TABLE
// Reference: tests/sqlparser_hive.rs:315
func TestHiveCreateTemporaryTable(t *testing.T) {
	dialects := hiveDialect()

	query := "CREATE TEMPORARY TABLE db.table (a INT NOT NULL)"
	query2 := "CREATE TEMP TABLE db.table (a INT NOT NULL)"

	dialects.VerifiedStmt(t, query)
	// TEMP should parse to TEMPORARY
	dialects.OneStatementParsesTo(t, query2, query)
}

// TestHiveCreateDelimitedTable tests CREATE TABLE with ROW FORMAT DELIMITED
// Reference: tests/sqlparser_hive.rs:324
func TestHiveCreateDelimitedTable(t *testing.T) {
	t.Skip("Hive ROW FORMAT DELIMITED not yet fully implemented in Go parser")
}

// ============================================================================
// DESCRIBE Tests
// ============================================================================

// TestHiveParseDescribe tests DESCRIBE statement parsing
// Reference: tests/sqlparser_hive.rs:46
func TestHiveParseDescribe(t *testing.T) {
	dialects := hiveAndGeneric()

	testCases := []string{
		"DESCRIBE namespace.`table`",
		"DESCRIBE namespace.table",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// DESCRIBE table without namespace not yet fully implemented
	t.Run("DESCRIBE table", func(t *testing.T) {
		t.Skip("DESCRIBE without namespace not yet fully implemented")
	})
}

// TestHiveExplainDescribeFormatted tests DESCRIBE FORMATTED
// Reference: tests/sqlparser_hive.rs:54
func TestHiveExplainDescribeFormatted(t *testing.T) {
	t.Skip("DESCRIBE FORMATTED not yet fully implemented in Go parser")
}

// TestHiveExplainDescribeExtended tests DESCRIBE EXTENDED
// Reference: tests/sqlparser_hive.rs:58
func TestHiveExplainDescribeExtended(t *testing.T) {
	t.Skip("DESCRIBE EXTENDED not yet fully implemented in Go parser")
}

// ============================================================================
// INSERT Tests
// ============================================================================

// TestHiveParseInsertOverwrite tests INSERT OVERWRITE parsing
// Reference: tests/sqlparser_hive.rs:63
func TestHiveParseInsertOverwrite(t *testing.T) {
	t.Skip("Hive INSERT OVERWRITE with PARTITION not yet fully implemented")
}

// TestHiveCreateLocalDirectory tests INSERT OVERWRITE LOCAL DIRECTORY
// Reference: tests/sqlparser_hive.rs:331
func TestHiveCreateLocalDirectory(t *testing.T) {
	t.Skip("Hive INSERT OVERWRITE LOCAL DIRECTORY not yet fully implemented")
}

// TestHiveColumnsAfterPartition tests INSERT with columns after partition
// Reference: tests/sqlparser_hive.rs:297
func TestHiveColumnsAfterPartition(t *testing.T) {
	t.Skip("Hive INSERT with PARTITION columns not yet fully implemented")
}

// TestHiveParseWithCte tests WITH CTE with INSERT
// Reference: tests/sqlparser_hive.rs:108
func TestHiveParseWithCte(t *testing.T) {
	t.Skip("Hive INSERT INTO TABLE with PARTITION not yet fully implemented")
}

// ============================================================================
// TRUNCATE Tests
// ============================================================================

// TestHiveTruncate tests TRUNCATE TABLE
// Reference: tests/sqlparser_hive.rs:69
func TestHiveTruncate(t *testing.T) {
	dialects := hiveDialect()
	truncate := "TRUNCATE TABLE db.table"
	dialects.VerifiedStmt(t, truncate)
}

// ============================================================================
// ANALYZE Tests
// ============================================================================

// TestHiveParseAnalyze tests ANALYZE TABLE
// Reference: tests/sqlparser_hive.rs:75
func TestHiveParseAnalyze(t *testing.T) {
	t.Skip("Hive ANALYZE TABLE not yet fully implemented in Go parser")
}

// TestHiveParseAnalyzeForColumns tests ANALYZE TABLE FOR COLUMNS
// Reference: tests/sqlparser_hive.rs:81
func TestHiveParseAnalyzeForColumns(t *testing.T) {
	t.Skip("Hive ANALYZE TABLE FOR COLUMNS not yet fully implemented in Go parser")
}

// ============================================================================
// MSCK Tests
// ============================================================================

// TestHiveParseMsck tests MSCK REPAIR TABLE
// Reference: tests/sqlparser_hive.rs:88
func TestHiveParseMsck(t *testing.T) {
	t.Skip("Hive MSCK REPAIR TABLE not yet fully implemented in Go parser")
}

// ============================================================================
// SET Tests
// ============================================================================

// TestHiveParseSetHivevar tests SET HIVEVAR
// Reference: tests/sqlparser_hive.rs:96
func TestHiveParseSetHivevar(t *testing.T) {
	t.Skip("Hive SET HIVEVAR not yet fully implemented in Go parser")
}

// TestHiveSetStatementWithMinus tests SET with negative value
// Reference: tests/sqlparser_hive.rs:371
func TestHiveSetStatementWithMinus(t *testing.T) {
	dialects := hiveDialect()

	// Valid SET with minus value
	dialects.VerifiedStmt(t, "SET hive.tez.java.opts = -Xmx4g")

	// Invalid: just minus sign
	_, err := parser.ParseSQL(hive.NewHiveDialect(), "SET hive.tez.java.opts = -")
	require.Error(t, err, "Expected error for incomplete minus value")
}

// ============================================================================
// DROP Tests
// ============================================================================

// TestHiveDropTablePurge tests DROP TABLE PURGE
// Reference: tests/sqlparser_hive.rs:114
func TestHiveDropTablePurge(t *testing.T) {
	t.Skip("Hive DROP TABLE PURGE not yet fully implemented in Go parser")
}

// ============================================================================
// ALTER TABLE Tests
// ============================================================================

// TestHiveAlterPartition tests ALTER TABLE PARTITION
// Reference: tests/sqlparser_hive.rs:236
func TestHiveAlterPartition(t *testing.T) {
	t.Skip("Hive ALTER TABLE PARTITION not yet fully implemented in Go parser")
}

// TestHiveAlterWithLocation tests ALTER TABLE with LOCATION
// Reference: tests/sqlparser_hive.rs:242
func TestHiveAlterWithLocation(t *testing.T) {
	t.Skip("Hive ALTER TABLE with LOCATION not yet fully implemented")
}

// TestHiveAlterWithSetLocation tests ALTER TABLE with SET LOCATION
// Reference: tests/sqlparser_hive.rs:249
func TestHiveAlterWithSetLocation(t *testing.T) {
	t.Skip("Hive ALTER TABLE with SET LOCATION not yet fully implemented")
}

// TestHiveAddPartition tests ALTER TABLE ADD PARTITION
// Reference: tests/sqlparser_hive.rs:255
func TestHiveAddPartition(t *testing.T) {
	t.Skip("Hive ALTER TABLE ADD PARTITION not yet fully implemented")
}

// TestHiveAddMultiplePartitions tests ALTER TABLE ADD multiple PARTITIONS
// Reference: tests/sqlparser_hive.rs:261
func TestHiveAddMultiplePartitions(t *testing.T) {
	t.Skip("Hive ALTER TABLE ADD multiple PARTITIONS not yet fully implemented")
}

// TestHiveDropPartition tests ALTER TABLE DROP PARTITION
// Reference: tests/sqlparser_hive.rs:267
func TestHiveDropPartition(t *testing.T) {
	t.Skip("Hive ALTER TABLE DROP PARTITION not yet fully implemented")
}

// TestHiveDropIfExists tests ALTER TABLE DROP IF EXISTS PARTITION
// Reference: tests/sqlparser_hive.rs:273
func TestHiveDropIfExists(t *testing.T) {
	t.Skip("Hive ALTER TABLE DROP IF EXISTS PARTITION not yet fully implemented")
}

// TestHiveRenameTable tests ALTER TABLE RENAME
// Reference: tests/sqlparser_hive.rs:352
func TestHiveRenameTable(t *testing.T) {
	dialects := hiveDialect()
	rename := "ALTER TABLE db.table_name RENAME TO db.table_2"
	dialects.VerifiedStmt(t, rename)
}

// ============================================================================
// SELECT Tests
// ============================================================================

// TestHiveSpaceship tests <=> operator (spaceship/null-safe equal)
// Reference: tests/sqlparser_hive.rs:102
func TestHiveSpaceship(t *testing.T) {
	dialects := hiveDialect()
	spaceship := "SELECT * FROM db.table WHERE a <=> b"
	dialects.VerifiedStmt(t, spaceship)
}

// TestHiveClusterBy tests CLUSTER BY clause
// Reference: tests/sqlparser_hive.rs:279
func TestHiveClusterBy(t *testing.T) {
	t.Skip("Hive CLUSTER BY not yet fully implemented in Go parser")
}

// TestHiveDistributeBy tests DISTRIBUTE BY clause
// Reference: tests/sqlparser_hive.rs:285
func TestHiveDistributeBy(t *testing.T) {
	t.Skip("Hive DISTRIBUTE BY not yet fully implemented in Go parser")
}

// TestHiveNoJoinCondition tests JOIN without condition
// Reference: tests/sqlparser_hive.rs:291
func TestHiveNoJoinCondition(t *testing.T) {
	dialects := hiveDialect()
	join := "SELECT a, b FROM db.table_name JOIN a"
	dialects.VerifiedStmt(t, join)
}

// TestHiveLateralView tests LATERAL VIEW
// Reference: tests/sqlparser_hive.rs:337
func TestHiveLateralView(t *testing.T) {
	t.Skip("Hive LATERAL VIEW not yet fully implemented in Go parser")
}

// TestHiveSortBy tests SORT BY clause
// Reference: tests/sqlparser_hive.rs:343
func TestHiveSortBy(t *testing.T) {
	t.Skip("Hive SORT BY not yet fully implemented in Go parser")
}

// TestHiveFromCte tests FROM CTE
// Reference: tests/sqlparser_hive.rs:364
func TestHiveFromCte(t *testing.T) {
	t.Skip("Hive FROM CTE with INSERT INTO TABLE not yet fully implemented")
}

// ============================================================================
// Data Type Tests
// ============================================================================

// TestHiveLongNumerics tests long numeric literals
// Reference: tests/sqlparser_hive.rs:303
func TestHiveLongNumerics(t *testing.T) {
	dialects := hiveDialect()
	query := "SELECT MIN(MIN(10, 5), 1L) AS a"
	dialects.VerifiedStmt(t, query)
}

// TestHiveDecimalPrecision tests DECIMAL with precision
// Reference: tests/sqlparser_hive.rs:309
func TestHiveDecimalPrecision(t *testing.T) {
	dialects := hiveDialect()
	query := "SELECT CAST(a AS DECIMAL(18,2)) FROM db.table"
	dialects.VerifiedStmt(t, query)
}

// TestHiveMapAccess tests map/array access with string key
// Reference: tests/sqlparser_hive.rs:358
func TestHiveMapAccess(t *testing.T) {
	dialects := hiveDialect()
	rename := "SELECT a.b[\"asdf\"] FROM db.table WHERE a = 2"
	dialects.VerifiedStmt(t, rename)
}

// ============================================================================
// CREATE FUNCTION Tests
// ============================================================================

// TestHiveParseCreateFunction tests CREATE TEMPORARY FUNCTION
// Reference: tests/sqlparser_hive.rs:399
func TestHiveParseCreateFunction(t *testing.T) {
	t.Skip("Hive CREATE TEMPORARY FUNCTION not yet fully implemented in Go parser")
}

// ============================================================================
// Delimited Identifiers Tests
// ============================================================================

// TestHiveParseDelimitedIdentifiers tests quoted identifiers
// Reference: tests/sqlparser_hive.rs:456
func TestHiveParseDelimitedIdentifiers(t *testing.T) {
	dialects := hiveDialect()

	// Check FROM with quoted identifiers
	sql := "SELECT \"alias\".\"bar baz\", \"myfun\"(), \"simple id\" AS \"column alias\" FROM \"a table\" AS \"alias\""
	dialects.VerifiedStmt(t, sql)

	// CREATE TABLE with quoted identifiers
	dialects.VerifiedStmt(t, `CREATE TABLE "foo" ("bar" "int")`)

	// ALTER TABLE with quoted constraint
	dialects.VerifiedStmt(t, `ALTER TABLE foo ADD CONSTRAINT "bar" PRIMARY KEY (baz)`)
}

// ============================================================================
// USE Statement Tests
// ============================================================================

// TestHiveParseUse tests USE statement
// Reference: tests/sqlparser_hive.rs:526
func TestHiveParseUse(t *testing.T) {
	t.Skip("Hive USE statement not yet fully implemented in Go parser")
}

// ============================================================================
// TABLE SAMPLE Tests
// ============================================================================

// TestHiveTableSample tests TABLESAMPLE clause
// Reference: tests/sqlparser_hive.rs:556
func TestHiveTableSample(t *testing.T) {
	t.Skip("Hive TABLESAMPLE clause not yet fully implemented in Go parser")
}

// ============================================================================
// FILTER as Alias Test
// ============================================================================

// TestHiveFilterAsAlias tests FILTER as a column alias
// Reference: tests/sqlparser_hive.rs:449
func TestHiveFilterAsAlias(t *testing.T) {
	dialects := hiveDialect()
	sql := "SELECT name filter FROM region"
	expected := "SELECT name AS filter FROM region"
	dialects.OneStatementParsesTo(t, sql, expected)
}

// ============================================================================
// Complex Query Tests
// ============================================================================

// TestHiveComplexQueries tests various complex Hive queries
func TestHiveComplexQueries(t *testing.T) {
	dialects := hiveDialect()

	// Working complex queries
	testCases := []string{
		// Simple SELECT
		"SELECT * FROM db.table",
		// SELECT with WHERE
		"SELECT a, b FROM db.table WHERE c = 1",
		// SELECT with ORDER BY
		"SELECT a FROM db.table ORDER BY a",
		// SELECT with GROUP BY
		"SELECT a, COUNT(*) FROM db.table GROUP BY a",
		// Subquery
		"SELECT * FROM (SELECT * FROM db.table) t",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// Placeholder Tests for Features Not Yet Implemented
// ============================================================================

// TestHiveParseIdentifier tests identifiers starting with numbers
// Reference: tests/sqlparser_hive.rs:230
func TestHiveParseIdentifier(t *testing.T) {
	t.Skip("Identifiers starting with numbers not yet fully implemented")
}

// TestHiveWindowFunctions tests window functions
// Reference: tests/sqlparser_hive.rs (additional tests)
func TestHiveWindowFunctions(t *testing.T) {
	dialects := hiveDialect()

	testCases := []string{
		"SELECT ROW_NUMBER() OVER (PARTITION BY col1 ORDER BY col2) AS rn FROM t",
		"SELECT RANK() OVER (ORDER BY a) FROM t",
		"SELECT SUM(a) OVER (PARTITION BY b) FROM t",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}
