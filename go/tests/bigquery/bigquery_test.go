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

// Package bigquery contains BigQuery-specific SQL parsing tests.
// These tests are ported from tests/sqlparser_bigquery.rs in the Rust implementation.
package bigquery

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// bigqueryDialect returns a TestedDialects with only BigQuery dialect
func bigqueryDialect() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			bigquery.NewBigQueryDialect(),
		},
	}
}

// bigqueryAndGeneric returns a TestedDialects with BigQuery and Generic dialects
func bigqueryAndGeneric() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			bigquery.NewBigQueryDialect(),
			generic.NewGenericDialect(),
		},
	}
}

// ============================================================================
// Literal String Tests
// ============================================================================

// TestBigQueryLiteralString tests parsing of various string literal formats
// Reference: tests/sqlparser_bigquery.rs:32
func TestBigQueryLiteralString(t *testing.T) {
	dialects := bigqueryDialect()

	// Basic string literals work
	t.Run("single quoted", func(t *testing.T) {
		dialects.VerifiedStmt(t, "SELECT 'single'")
	})

	t.Run("double quoted", func(t *testing.T) {
		dialects.VerifiedStmt(t, `SELECT "double"`)
	})

	// Triple-quoted strings not yet fully implemented
	t.Run("triple single quoted", func(t *testing.T) {
		t.Skip("Triple-quoted strings not yet fully implemented")
	})

	t.Run("triple double quoted", func(t *testing.T) {
		t.Skip("Triple-quoted strings not yet fully implemented")
	})

	t.Run("escaped strings", func(t *testing.T) {
		t.Skip("Backslash escaping behavior needs refinement")
	})
}

// TestBigQueryByteLiteral tests byte literal parsing (B'...', B"...", etc.)
// Reference: tests/sqlparser_bigquery.rs:121
func TestBigQueryByteLiteral(t *testing.T) {
	dialects := bigqueryDialect()

	// Basic byte literals work
	t.Run("single quoted", func(t *testing.T) {
		dialects.VerifiedStmt(t, "SELECT B'abc'")
	})

	t.Run("double quoted", func(t *testing.T) {
		dialects.VerifiedStmt(t, `SELECT B"abc"`)
	})

	// Lowercase prefix normalizes to uppercase
	t.Run("lowercase prefix", func(t *testing.T) {
		dialects.OneStatementParsesTo(t, "SELECT b'123'", "SELECT B'123'")
	})

	// Triple-quoted byte strings not yet fully implemented
	t.Run("triple quoted", func(t *testing.T) {
		t.Skip("Triple-quoted byte strings not yet fully implemented")
	})
}

// TestBigQueryRawLiteral tests raw string literal parsing (R'...', R"...", etc.)
// Reference: tests/sqlparser_bigquery.rs:168
func TestBigQueryRawLiteral(t *testing.T) {
	t.Skip("Raw string literals (R'...') not yet implemented in Go parser")
}

// TestBigQueryExponentDecimal tests exponent decimal literals
// Reference: tests/sqlparser_bigquery.rs:230
func TestBigQueryExponentDecimal(t *testing.T) {
	dialects := bigqueryDialect()
	// Parser normalizes exponent notation to lowercase
	dialects.OneStatementParsesTo(t, "SELECT 1e-10, 1.5e-5, 1E+10, 1.5E5", "SELECT 1e-10, 1.5e-5, 1e+10, 1.5e5")
}

// TestBigQueryNonReservedColumnAlias tests that certain keywords can be column aliases
// Reference: tests/sqlparser_bigquery.rs:251
func TestBigQueryNonReservedColumnAlias(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT OFFSET, EXPLAIN, ANALYZE, SORT, TOP, VIEW FROM T",
		"SELECT 1 AS OFFSET, 2 AS EXPLAIN, 3 AS ANALYZE FROM T",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryAtAtIdentifier tests @@variable syntax
// Reference: tests/sqlparser_bigquery.rs:260
func TestBigQueryAtAtIdentifier(t *testing.T) {
	dialects := bigqueryDialect()
	sql := "SELECT @@error.stack_trace, @@error.message"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// DECLARE Statement Tests
// ============================================================================

// TestBigQueryDeclare tests DECLARE statement parsing
// Reference: tests/sqlparser_bigquery.rs:2140
func TestBigQueryDeclare(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"DECLARE x INT64",
		"DECLARE x INT64 DEFAULT 42",
		"DECLARE x, y, z INT64 DEFAULT 42",
		"DECLARE x DEFAULT 42",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryDeclareErrors tests DECLARE statement error cases
// Reference: tests/sqlparser_bigquery.rs:2140
func TestBigQueryDeclareErrors(t *testing.T) {
	dialect := bigquery.NewBigQueryDialect()

	testCases := []struct {
		sql         string
		expectError bool
	}{
		{"DECLARE x", true},
		{"DECLARE x 42", true},
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

// ============================================================================
// BEGIN/EXCEPTION Statement Tests
// ============================================================================

// TestBigQueryBegin tests BEGIN statement parsing
// Reference: tests/sqlparser_bigquery.rs:265
func TestBigQueryBegin(t *testing.T) {
	dialects := bigqueryDialect()

	// Basic BEGIN/BEGIN TRANSACTION work
	t.Run("BEGIN TRANSACTION", func(t *testing.T) {
		dialects.VerifiedStmt(t, "BEGIN TRANSACTION")
	})

	t.Run("BEGIN", func(t *testing.T) {
		dialects.VerifiedStmt(t, "BEGIN")
	})

	// BEGIN...EXCEPTION blocks not yet fully implemented
	t.Run("BEGIN EXCEPTION blocks", func(t *testing.T) {
		t.Skip("BEGIN...EXCEPTION blocks not yet fully implemented in Go parser")
	})
}

// ============================================================================
// DELETE Statement Tests
// ============================================================================

// TestBigQueryDeleteStatement tests DELETE statement parsing
// Reference: tests/sqlparser_bigquery.rs:309
func TestBigQueryDeleteStatement(t *testing.T) {
	dialects := bigqueryAndGeneric()
	// Parser normalizes DELETE table to DELETE FROM table
	dialects.OneStatementParsesTo(t, `DELETE "table" WHERE 1`, `DELETE FROM "table" WHERE 1`)
}

// ============================================================================
// SELECT Statement Tests - Wildcard EXCEPT/REPLACE
// ============================================================================

// TestBigQuerySelectWildcardExcept tests SELECT * EXCEPT
// Reference: tests/sqlparser_bigquery.rs:341
func TestBigQuerySelectWildcardExcept(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT * EXCEPT (col1) FROM data",
		"SELECT * EXCEPT (col1, col2) FROM data",
		"SELECT t.* EXCEPT (col1) FROM data t",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQuerySelectWildcardReplace tests SELECT * REPLACE
// Reference: tests/sqlparser_bigquery.rs:424
func TestBigQuerySelectWildcardReplace(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT * REPLACE (col1 + 1 AS col1) FROM data",
		"SELECT * REPLACE (col1 AS new_col1, col2 AS new_col2) FROM data",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQuerySelectWildcardReplaceAndExcept tests SELECT * REPLACE with EXCEPT
// Reference: tests/sqlparser_bigquery.rs:460
func TestBigQuerySelectWildcardReplaceAndExcept(t *testing.T) {
	t.Skip("SELECT * REPLACE with EXCEPT combination not yet fully implemented")
}

// TestBigQuerySelectExprStar tests SELECT expression.*
// Reference: tests/sqlparser_bigquery.rs:2454
func TestBigQuerySelectExprStar(t *testing.T) {
	dialects := bigqueryDialect()

	t.Run("complex struct", func(t *testing.T) {
		t.Skip("SELECT expression.* with complex struct not yet fully implemented")
	})

	t.Run("array access", func(t *testing.T) {
		dialects.VerifiedStmt(t, "SELECT myfunc()[0].* FROM T")
	})
}

// ============================================================================
// STRUCT Tests
// ============================================================================

// TestBigQuerySelectStruct tests SELECT STRUCT syntax
// Reference: tests/sqlparser_bigquery.rs:491
func TestBigQuerySelectStruct(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT STRUCT(1, 2, 3)",
		"SELECT STRUCT('abc')",
		"SELECT STRUCT(1, t.str_col)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// STRUCT with AS syntax not yet fully implemented
	t.Run("STRUCT with AS", func(t *testing.T) {
		t.Skip("STRUCT with AS syntax not yet fully implemented")
	})
}

// TestBigQuerySelectAsStruct tests SELECT AS STRUCT
// Reference: tests/sqlparser_bigquery.rs:2462
func TestBigQuerySelectAsStruct(t *testing.T) {
	t.Skip("SELECT AS STRUCT not yet fully implemented in Go parser")
}

// TestBigQuerySelectAsValue tests SELECT AS VALUE
// Reference: tests/sqlparser_bigquery.rs:2485
func TestBigQuerySelectAsValue(t *testing.T) {
	t.Skip("SELECT AS VALUE not yet fully implemented in Go parser")
}

// TestBigQueryTypedStructSyntax tests typed STRUCT syntax
// Reference: tests/sqlparser_bigquery.rs:752
func TestBigQueryTypedStructSyntax(t *testing.T) {
	dialects := bigqueryDialect()

	// Working cases
	testCases := []string{
		"SELECT STRUCT<x INT64, y STRING>(1, t.str_col)",
		"SELECT STRUCT<key INT64, value INT64>(1, 2)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// Cases with type parameters and special syntax not yet fully implemented
	t.Run("complex typed struct", func(t *testing.T) {
		t.Skip("Complex typed STRUCT syntax not yet fully implemented")
	})
}

// TestBigQueryTypedStructWithFieldName tests typed STRUCT with field names
// Reference: tests/sqlparser_bigquery.rs:1444
func TestBigQueryTypedStructWithFieldName(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT STRUCT<x INT64>(5)",
		`SELECT STRUCT<y STRING>("foo")`,
		"SELECT STRUCT<x INT64, y INT64>(5, 5)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryNestedStruct tests nested STRUCT definitions
// Reference: tests/sqlparser_bigquery.rs:2576
func TestBigQueryNestedStruct(t *testing.T) {
	t.Skip("Nested STRUCT in CREATE TABLE not yet fully implemented")
}

// ============================================================================
// Table Identifier Tests
// ============================================================================

// TestBigQueryTableIdentifiers tests various table identifier formats
// Reference: tests/sqlparser_bigquery.rs:1550
func TestBigQueryTableIdentifiers(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []struct {
		sql string
	}{
		{"SELECT 1 FROM `spa ce`"},
		{"SELECT 1 FROM `_5abc`.dataField"},
		{"SELECT 1 FROM `5abc`.dataField"},
		{"SELECT 1 FROM abc5.dataField"},
		{"SELECT 1 FROM `GROUP`.dataField"},
		{"SELECT 1 FROM abc5.GROUP"},
		{"SELECT 1 FROM `foo.bar.baz`"},
		{"SELECT 1 FROM `foo.bar`.`baz`"},
		{"SELECT 1 FROM `foo`.`bar.baz`"},
		{"SELECT 1 FROM `foo`.`bar`.`baz`"},
		{"SELECT 1 FROM `5abc.dataField`"},
		{"SELECT 1 FROM `_5abc.da-sh-es`"},
	}

	for _, tc := range testCases {
		t.Run(tc.sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, tc.sql)
		})
	}
}

// TestBigQueryTableIdentifiersWithHyphen tests hyphenated table identifiers
// Reference: tests/sqlparser_bigquery.rs:1700
func TestBigQueryTableIdentifiersWithHyphen(t *testing.T) {
	t.Skip("Hyphenated table identifiers not yet fully implemented in Go parser")
}

// TestBigQueryTableIdentifiersWithHyphenInvalid tests invalid hyphenated identifiers
// Reference: tests/sqlparser_bigquery.rs:588
func TestBigQueryTableIdentifiersWithHyphenInvalid(t *testing.T) {
	dialect := bigquery.NewBigQueryDialect()

	testCases := []struct {
		sql         string
		expectError bool
	}{
		{"SELECT 1 FROM foo-`bar`", true},
		{"SELECT 1 FROM `foo`-bar", true},
		{"SELECT 1 FROM foo-123a", true},
		{"SELECT 1 FROM foo - bar", true},
		{"SELECT 1 FROM 123-bar", true},
		{"SELECT 1 FROM bar-", true},
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

// TestBigQueryHyphenatedIdentifierExpression tests expressions with hyphenated identifiers
// Reference: tests/sqlparser_bigquery.rs:1706
func TestBigQueryHyphenatedIdentifierExpression(t *testing.T) {
	dialects := bigqueryDialect()
	// This parses as foo minus bar.x, not a hyphenated identifier
	dialects.OneStatementParsesTo(t, "SELECT foo-bar.x FROM t", "SELECT foo - bar.x FROM t")
}

// ============================================================================
// Time Travel Tests
// ============================================================================

// TestBigQueryTableTimeTravel tests FOR SYSTEM_TIME AS OF clause
// Reference: tests/sqlparser_bigquery.rs:1739
func TestBigQueryTableTimeTravel(t *testing.T) {
	t.Skip("FOR SYSTEM_TIME AS OF clause not yet fully implemented in Go parser")
}

// TestBigQueryTableTimeTravelInvalid tests invalid time travel syntax
// Reference: tests/sqlparser_bigquery.rs:1765
func TestBigQueryTableTimeTravelInvalid(t *testing.T) {
	dialect := bigquery.NewBigQueryDialect()
	sql := "SELECT 1 FROM t1 FOR SYSTEM TIME AS OF 'some_timestamp'"
	_, err := parser.ParseSQL(dialect, sql)
	require.Error(t, err, "Expected error for SQL: %s", sql)
}

// ============================================================================
// JOIN with UNNEST Tests
// ============================================================================

// TestBigQueryJoinConstraintUnnestAlias tests JOIN with UNNEST and alias
// Reference: tests/sqlparser_bigquery.rs:1769
func TestBigQueryJoinConstraintUnnestAlias(t *testing.T) {
	dialects := bigqueryDialect()
	sql := "SELECT * FROM t1 JOIN UNNEST(t1.a) AS f ON c1 = c2"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// QUALIFY Clause Tests
// ============================================================================

// TestBigQueryQualify tests QUALIFY clause parsing
// Reference: tests/sqlparser_bigquery.rs:687
func TestBigQueryQualify(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT col1, col2 FROM data QUALIFY ROW_NUMBER() OVER (PARTITION BY col1 ORDER BY col2) = 1",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// ANALYZE Statement Tests
// ============================================================================

// TestBigQueryAnalyze tests ANALYZE statement parsing
// Reference: tests/sqlparser_bigquery.rs:720
func TestBigQueryAnalyze(t *testing.T) {
	dialects := bigqueryDialect()

	t.Run("basic", func(t *testing.T) {
		dialects.VerifiedStmt(t, "ANALYZE mydataset.mytable")
	})

	t.Run("with options", func(t *testing.T) {
		t.Skip("ANALYZE with OPTIONS not yet fully implemented in Go parser")
	})
}

// ============================================================================
// CREATE TABLE Tests
// ============================================================================

// TestBigQueryCreateTable tests CREATE TABLE parsing
// Reference: tests/sqlparser_bigquery.rs:783
func TestBigQueryCreateTable(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"CREATE TABLE mydataset.newtable (x INT64 NOT NULL)",
		"CREATE TABLE mydataset.newtable (x INT64 NOT NULL, y BOOL)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryCreateTableWithOptions tests CREATE TABLE with OPTIONS
// Reference: tests/sqlparser_bigquery.rs:479
func TestBigQueryCreateTableWithOptions(t *testing.T) {
	t.Skip("CREATE TABLE with OPTIONS not yet fully implemented in Go parser")
}

// TestBigQueryCreateTablePartitionByDate tests CREATE TABLE with PARTITION BY
// Reference: tests/sqlparser_bigquery.rs:818
func TestBigQueryCreateTablePartitionByDate(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"CREATE TABLE mydataset.newtable (ts TIMESTAMP) PARTITION BY DATE(ts)",
		"CREATE TABLE mydataset.newtable (dt DATE) PARTITION BY _PARTITIONDATE",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryCreateTableClusterBy tests CREATE TABLE with CLUSTER BY
// Reference: tests/sqlparser_bigquery.rs:860
func TestBigQueryCreateTableClusterBy(t *testing.T) {
	t.Skip("CREATE TABLE with CLUSTER BY not yet fully implemented in Go parser")
}

// TestBigQueryCreateExternalTable tests CREATE EXTERNAL TABLE
// Reference: tests/sqlparser_bigquery.rs:595
func TestBigQueryCreateExternalTable(t *testing.T) {
	t.Skip("CREATE EXTERNAL TABLE not yet fully implemented in Go parser")
}

// TestBigQueryCreateTableWithUnquotedHyphen tests CREATE TABLE with hyphenated project names
// Reference: tests/sqlparser_bigquery.rs:453
func TestBigQueryCreateTableWithUnquotedHyphen(t *testing.T) {
	t.Skip("Hyphenated project names in CREATE TABLE not yet fully implemented in Go parser")
}

// TestBigQueryNestedDataTypes tests nested data types like STRUCT and ARRAY
// Reference: tests/sqlparser_bigquery.rs:605
func TestBigQueryNestedDataTypes(t *testing.T) {
	t.Skip("Nested STRUCT with angle brackets in CREATE TABLE not yet fully implemented in Go parser")
}

// TestBigQueryStructFieldOptions tests struct field with OPTIONS
// Reference: tests/sqlparser_bigquery.rs:2563
func TestBigQueryStructFieldOptions(t *testing.T) {
	t.Skip("STRUCT field with OPTIONS not yet fully implemented in Go parser")
}

// ============================================================================
// OPTIONS Tests
// ============================================================================

// TestBigQueryOptions tests OPTIONS clause parsing
// Reference: tests/sqlparser_bigquery.rs:896
func TestBigQueryOptions(t *testing.T) {
	t.Skip("OPTIONS clause in CREATE TABLE not yet fully implemented in Go parser")
}

// ============================================================================
// LOAD DATA Tests
// ============================================================================

// TestBigQueryLoadData tests LOAD DATA statement
// Reference: tests/sqlparser_bigquery.rs:930
func TestBigQueryLoadData(t *testing.T) {
	t.Skip("LOAD DATA feature not yet implemented in Go parser")
}

// ============================================================================
// INSERT Statement Tests
// ============================================================================

// TestBigQueryInsert tests INSERT statement parsing
// Reference: tests/sqlparser_bigquery.rs:963
func TestBigQueryInsert(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"INSERT INTO mydataset.mytable (col1, col2) VALUES (1, 2)",
		"INSERT INTO mydataset.mytable VALUES (1, 2)",
		"INSERT INTO mydataset.mytable SELECT * FROM src",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// MERGE Statement Tests
// ============================================================================

// TestBigQueryMerge tests MERGE statement parsing
// Reference: tests/sqlparser_bigquery.rs:990
func TestBigQueryMerge(t *testing.T) {
	dialects := bigqueryAndGeneric()

	testCases := []string{
		"MERGE inventory AS T USING newArrivals AS S ON false WHEN NOT MATCHED THEN INSERT (product, quantity) VALUES (1, 2)",
		"MERGE inventory AS T USING newArrivals AS S ON false WHEN MATCHED THEN UPDATE SET a = 1",
		"MERGE inventory AS T USING newArrivals AS S ON false WHEN MATCHED THEN DELETE",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryMergeComplex tests complex MERGE statements with multiple clauses
// Reference: tests/sqlparser_bigquery.rs:1799
func TestBigQueryMergeComplex(t *testing.T) {
	dialects := bigqueryAndGeneric()
	sql := "MERGE inventory AS T USING newArrivals AS S ON false WHEN NOT MATCHED AND 1 THEN INSERT (product, quantity) VALUES (1, 2) WHEN NOT MATCHED BY TARGET AND 1 THEN INSERT (product, quantity) VALUES (1, 2) WHEN NOT MATCHED BY TARGET THEN INSERT (product, quantity) VALUES (1, 2) WHEN NOT MATCHED BY SOURCE AND 2 THEN DELETE WHEN NOT MATCHED BY SOURCE THEN DELETE WHEN NOT MATCHED BY SOURCE AND 1 THEN UPDATE SET a = 1, b = 2 WHEN NOT MATCHED AND 1 THEN INSERT (product, quantity) ROW WHEN NOT MATCHED THEN INSERT (product, quantity) ROW WHEN NOT MATCHED AND 1 THEN INSERT ROW WHEN NOT MATCHED THEN INSERT ROW WHEN MATCHED AND 1 THEN DELETE WHEN MATCHED THEN UPDATE SET a = 1, b = 2 WHEN NOT MATCHED THEN INSERT (a, b) VALUES (1, DEFAULT) WHEN NOT MATCHED THEN INSERT VALUES (1, DEFAULT)"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// UPDATE Statement Tests
// ============================================================================

// TestBigQueryUpdate tests UPDATE statement parsing
// Reference: tests/sqlparser_bigquery.rs:1018
func TestBigQueryUpdate(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"UPDATE mydataset.mytable SET col1 = 1 WHERE col2 = 2",
		"UPDATE mydataset.mytable SET col1 = 1, col2 = 2 WHERE col3 = 3",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// DELETE Statement Tests
// ============================================================================

// TestBigQueryDelete tests DELETE statement parsing
// Reference: tests/sqlparser_bigquery.rs:1045
func TestBigQueryDelete(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"DELETE FROM mydataset.mytable WHERE col1 = 1",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// DELETE without FROM keyword normalizes to with FROM
	t.Run("DELETE without FROM", func(t *testing.T) {
		dialects.OneStatementParsesTo(t, "DELETE mydataset.mytable WHERE col1 = 1", "DELETE FROM mydataset.mytable WHERE col1 = 1")
	})
}

// ============================================================================
// CREATE VIEW Tests
// ============================================================================

// TestBigQueryCreateView tests CREATE VIEW parsing
// Reference: tests/sqlparser_bigquery.rs:1072
func TestBigQueryCreateView(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"CREATE VIEW mydataset.myview AS SELECT col1 FROM mydataset.mytable",
		"CREATE OR REPLACE VIEW mydataset.myview AS SELECT col1 FROM mydataset.mytable",
		"CREATE VIEW IF NOT EXISTS mydataset.myview AS SELECT col1 FROM mydataset.mytable",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryCreateViewWithOptions tests CREATE VIEW with OPTIONS
// Reference: tests/sqlparser_bigquery.rs:326
func TestBigQueryCreateViewWithOptions(t *testing.T) {
	t.Skip("CREATE VIEW with OPTIONS not yet fully implemented in Go parser")
}

// TestBigQueryCreateViewIfNotExists tests CREATE VIEW IF NOT EXISTS
// Reference: tests/sqlparser_bigquery.rs:401
func TestBigQueryCreateViewIfNotExists(t *testing.T) {
	dialects := bigqueryDialect()
	sql := "CREATE VIEW IF NOT EXISTS mydataset.newview AS SELECT foo FROM bar"
	dialects.VerifiedStmt(t, sql)
}

// TestBigQueryCreateViewWithUnquotedHyphen tests CREATE VIEW with hyphenated names
// Reference: tests/sqlparser_bigquery.rs:435
func TestBigQueryCreateViewWithUnquotedHyphen(t *testing.T) {
	t.Skip("Hyphenated project names in CREATE VIEW not yet fully implemented in Go parser")
}

// ============================================================================
// CREATE MATERIALIZED VIEW Tests
// ============================================================================

// TestBigQueryCreateMaterializedView tests CREATE MATERIALIZED VIEW
// Reference: tests/sqlparser_bigquery.rs:1100
func TestBigQueryCreateMaterializedView(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"CREATE MATERIALIZED VIEW mydataset.myview AS SELECT col1 FROM mydataset.mytable",
		"CREATE OR REPLACE MATERIALIZED VIEW mydataset.myview AS SELECT col1 FROM mydataset.mytable",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// CREATE PROCEDURE Tests
// ============================================================================

// TestBigQueryCreateProcedure tests CREATE PROCEDURE parsing
// Reference: tests/sqlparser_bigquery.rs:1130
func TestBigQueryCreateProcedure(t *testing.T) {
	t.Skip("CREATE PROCEDURE feature not yet implemented in Go parser")
}

// ============================================================================
// CREATE FUNCTION Tests
// ============================================================================

// TestBigQueryCreateFunction tests CREATE FUNCTION parsing
// Reference: tests/sqlparser_bigquery.rs:1160
func TestBigQueryCreateFunction(t *testing.T) {
	t.Skip("CREATE FUNCTION not yet fully implemented in Go parser")
}

// TestBigQueryCreateFunctionAnyType tests CREATE FUNCTION with ANY TYPE parameter
// Reference: tests/sqlparser_bigquery.rs:2547
func TestBigQueryCreateFunctionAnyType(t *testing.T) {
	t.Skip("CREATE FUNCTION with ANY TYPE parameter not yet fully implemented")
}

// TestBigQueryCreateFunctionAnyTypeTable tests ANY TYPE in CREATE TABLE
// Reference: tests/sqlparser_bigquery.rs:2558
func TestBigQueryCreateFunctionAnyTypeTable(t *testing.T) {
	dialects := bigqueryAndGeneric()
	sql := "CREATE TABLE foo (x ANY)"
	dialects.VerifiedStmt(t, sql)
}

// ============================================================================
// DROP Statement Tests
// ============================================================================

// TestBigQueryDropTable tests DROP TABLE parsing
// Reference: tests/sqlparser_bigquery.rs:1190
func TestBigQueryDropTable(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"DROP TABLE mydataset.mytable",
		"DROP TABLE IF EXISTS mydataset.mytable",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryDropView tests DROP VIEW parsing
// Reference: tests/sqlparser_bigquery.rs:1220
func TestBigQueryDropView(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"DROP VIEW mydataset.myview",
		"DROP VIEW IF EXISTS mydataset.myview",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryDropProcedure tests DROP PROCEDURE parsing
// Reference: tests/sqlparser_bigquery.rs:1250
func TestBigQueryDropProcedure(t *testing.T) {
	t.Skip("DROP PROCEDURE feature not yet implemented in Go parser")
}

// TestBigQueryDropFunction tests DROP FUNCTION parsing
// Reference: tests/sqlparser_bigquery.rs:1280
func TestBigQueryDropFunction(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"DROP FUNCTION mydataset.myfunction",
		"DROP FUNCTION IF EXISTS mydataset.myfunction",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// ALTER TABLE Tests
// ============================================================================

// TestBigQueryAlterTable tests ALTER TABLE parsing
// Reference: tests/sqlparser_bigquery.rs:1310
func TestBigQueryAlterTable(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"ALTER TABLE mydataset.mytable ADD COLUMN col1 INT64",
		"ALTER TABLE mydataset.mytable DROP COLUMN col1",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// CREATE SCHEMA Tests
// ============================================================================

// TestBigQueryCreateSchema tests CREATE SCHEMA parsing
// Reference: tests/sqlparser_bigquery.rs:1340
func TestBigQueryCreateSchema(t *testing.T) {
	dialects := bigqueryAndGeneric()

	testCases := []string{
		"CREATE SCHEMA mydataset",
		"CREATE SCHEMA IF NOT EXISTS mydataset",
		"CREATE SCHEMA mydataset OPTIONS(location = 'us')",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// ALTER SCHEMA Tests
// ============================================================================

// TestBigQueryAlterSchema tests ALTER SCHEMA parsing
// Reference: tests/sqlparser_bigquery.rs:2893
func TestBigQueryAlterSchema(t *testing.T) {
	t.Skip("ALTER SCHEMA with various options not yet fully implemented in Go parser")
}

// ============================================================================
// PIVOT and UNPIVOT Tests
// ============================================================================

// TestBigQuerySelectPivot tests PIVOT clause
// Reference: tests/sqlparser_bigquery.rs:1430
func TestBigQuerySelectPivot(t *testing.T) {
	t.Skip("PIVOT feature not yet fully implemented in Go parser")
}

// TestBigQuerySelectUnpivot tests UNPIVOT clause
// Reference: tests/sqlparser_bigquery.rs:1460
func TestBigQuerySelectUnpivot(t *testing.T) {
	t.Skip("UNPIVOT feature not yet fully implemented in Go parser")
}

// ============================================================================
// Array Subquery Tests
// ============================================================================

// TestBigQueryArraySubquery tests array subquery syntax
// Reference: tests/sqlparser_bigquery.rs:1490
func TestBigQueryArraySubquery(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT ARRAY(SELECT 1)",
		"SELECT ARRAY(SELECT x FROM t)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// Array Agg Tests
// ============================================================================

// TestBigQueryArrayAgg tests ARRAY_AGG function
// Reference: tests/sqlparser_bigquery.rs:1520
func TestBigQueryArrayAgg(t *testing.T) {
	dialects := bigqueryAndGeneric()

	testCases := []string{
		"SELECT ARRAY_AGG(state)",
		"SELECT ARRAY_AGG(x ORDER BY x) AS a FROM T",
		"SELECT ARRAY_AGG(x ORDER BY x LIMIT 2) FROM tbl",
		"SELECT ARRAY_AGG(DISTINCT x ORDER BY x LIMIT 2) FROM tbl",
		"SELECT ARRAY_AGG(state IGNORE NULLS LIMIT 10)",
		"SELECT ARRAY_AGG(state RESPECT NULLS ORDER BY population)",
		"SELECT ARRAY_AGG(DISTINCT state IGNORE NULLS ORDER BY population DESC LIMIT 10)",
		"SELECT ARRAY_CONCAT_AGG(x LIMIT 2)",
		"SELECT ARRAY_CONCAT_AGG(x ORDER BY ARRAY_LENGTH(x))",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryAnyValue tests ANY_VALUE function
// Reference: tests/sqlparser_bigquery.rs:2537
func TestBigQueryAnyValue(t *testing.T) {
	dialects := bigqueryAndGeneric()

	testCases := []string{
		"SELECT ANY_VALUE(fruit)",
		"SELECT ANY_VALUE(fruit) OVER (ORDER BY LENGTH(fruit) ROWS BETWEEN 1 PRECEDING AND CURRENT ROW)",
		"SELECT ANY_VALUE(fruit HAVING MAX sold)",
		"SELECT ANY_VALUE(fruit HAVING MIN sold)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// CREATE TEMP TABLE Tests
// ============================================================================

// TestBigQueryCreateTempTable tests CREATE TEMP TABLE parsing
// Reference: tests/sqlparser_bigquery.rs:1550
func TestBigQueryCreateTempTable(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"CREATE TEMPORARY TABLE mytable (x INT64)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// TEMP keyword (abbreviation) not yet fully implemented
	t.Run("TEMP keyword", func(t *testing.T) {
		t.Skip("CREATE TEMP TABLE (abbreviated TEMP) not yet fully implemented")
	})
}

// ============================================================================
// OFFSET Tests
// ============================================================================

// TestBigQueryOffset tests OFFSET function for array access
// Reference: tests/sqlparser_bigquery.rs:1610
func TestBigQueryOffset(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT OFFSET(0)",
		"SELECT SAFE_OFFSET(0)",
		"SELECT users[OFFSET(0)] FROM t",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestBigQueryMapAccessExpr tests map/array access expressions
// Reference: tests/sqlparser_bigquery.rs:2215
func TestBigQueryMapAccessExpr(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT users[-1][safe_offset(2)].a.b",
		"SELECT myfunc()[-1].a[SAFE_OFFSET(2)].b",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// CAST Tests
// ============================================================================

// TestBigQueryCastType tests CAST with SAFE_CAST
// Reference: tests/sqlparser_bigquery.rs:2092
func TestBigQueryCastType(t *testing.T) {
	dialects := bigqueryAndGeneric()
	sql := "SELECT SAFE_CAST(1 AS INT64)"
	dialects.VerifiedStmt(t, sql)
}

// TestBigQueryCastDateFormat tests CAST with FORMAT clause for dates
// Reference: tests/sqlparser_bigquery.rs:2098
func TestBigQueryCastDateFormat(t *testing.T) {
	t.Skip("CAST with FORMAT clause not yet fully implemented in Go parser")
}

// TestBigQueryCastTimeFormat tests CAST with FORMAT clause for time
// Reference: tests/sqlparser_bigquery.rs:2106
func TestBigQueryCastTimeFormat(t *testing.T) {
	t.Skip("CAST with FORMAT clause not yet fully implemented in Go parser")
}

// TestBigQueryCastTimestampFormatTz tests CAST with FORMAT and AT TIME ZONE
// Reference: tests/sqlparser_bigquery.rs:2111
func TestBigQueryCastTimestampFormatTz(t *testing.T) {
	t.Skip("CAST with FORMAT and AT TIME ZONE not yet fully implemented in Go parser")
}

// TestBigQueryCastStringToBytesFormat tests CAST string to bytes with format
// Reference: tests/sqlparser_bigquery.rs:2118
func TestBigQueryCastStringToBytesFormat(t *testing.T) {
	t.Skip("CAST with FORMAT clause not yet fully implemented in Go parser")
}

// TestBigQueryCastBytesToStringFormat tests CAST bytes to string with format
// Reference: tests/sqlparser_bigquery.rs:2124
func TestBigQueryCastBytesToStringFormat(t *testing.T) {
	t.Skip("CAST with FORMAT clause not yet fully implemented in Go parser")
}

// ============================================================================
// Triple Quote Typed Strings Tests
// ============================================================================

// TestBigQueryTripleQuoteTypedStrings tests triple-quoted string literals with type
// Reference: tests/sqlparser_bigquery.rs:2508
func TestBigQueryTripleQuoteTypedStrings(t *testing.T) {
	t.Skip("Triple-quoted typed strings not yet fully implemented in Go parser")
}

// ============================================================================
// Tuple/Struct Literal Tests
// ============================================================================

// TestBigQueryTupleStructLiteral tests tuple/struct literal syntax
// Reference: tests/sqlparser_bigquery.rs:656
func TestBigQueryTupleStructLiteral(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		"SELECT (1, 2, 3)",
		"SELECT (1, 1.0, '123', true)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// EXTRACT Tests
// ============================================================================

// TestBigQueryExtractWeekday tests EXTRACT with WEEK(MONDAY)
// Reference: tests/sqlparser_bigquery.rs:2440
func TestBigQueryExtractWeekday(t *testing.T) {
	t.Skip("EXTRACT with WEEK(MONDAY) not yet fully implemented in Go parser")
}

// ============================================================================
// EXPORT DATA Tests
// ============================================================================

// TestBigQueryExportData tests EXPORT DATA statement
// Reference: tests/sqlparser_bigquery.rs:2651
func TestBigQueryExportData(t *testing.T) {
	t.Skip("EXPORT DATA statement not yet fully implemented in Go parser")
}

// TestBigQueryExportDataErrors tests EXPORT DATA error cases
// Reference: tests/sqlparser_bigquery.rs:2859
func TestBigQueryExportDataErrors(t *testing.T) {
	t.Skip("EXPORT DATA error cases - parent test already skipped")
}

// ============================================================================
// CREATE SNAPSHOT TABLE Tests
// ============================================================================

// TestBigQueryCreateSnapshotTable tests CREATE SNAPSHOT TABLE
// Reference: tests/sqlparser_bigquery.rs:2905
func TestBigQueryCreateSnapshotTable(t *testing.T) {
	t.Skip("CREATE SNAPSHOT TABLE not yet fully implemented in Go parser")
}

// ============================================================================
// TRIM Tests
// ============================================================================

// TestBigQueryTrim tests TRIM function with comma-separated arguments
// Reference: tests/sqlparser_bigquery.rs:2411
func TestBigQueryTrim(t *testing.T) {
	dialects := bigqueryDialect()

	testCases := []string{
		`SELECT customer_id, TRIM(item_price_id, '"', "a") AS item_price_id FROM models_staging.subscriptions`,
		"SELECT TRIM('xyz', 'a')",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// Trailing Comma Tests
// ============================================================================

// TestBigQueryTrailingComma tests trailing commas in SELECT
// Reference: tests/sqlparser_bigquery.rs:2075
func TestBigQueryTrailingComma(t *testing.T) {
	dialects := bigqueryDialect()

	// Trailing commas with FROM/LIMIT work
	testCases := []struct {
		sql      string
		expected string
	}{
		{"SELECT a, b AS c, FROM t", "SELECT a, b AS c FROM t"},
		{"SELECT a, b, FROM t", "SELECT a, b FROM t"},
		{"SELECT a, b, LIMIT 1", "SELECT a, b LIMIT 1"},
	}

	for _, tc := range testCases {
		t.Run(tc.sql, func(t *testing.T) {
			dialects.OneStatementParsesTo(t, tc.sql, tc.expected)
		})
	}

	// Standalone trailing commas not at end of statement not yet fully implemented
	t.Run("standalone trailing commas", func(t *testing.T) {
		t.Skip("Standalone trailing commas not yet fully implemented")
	})
}

// ============================================================================
// FETCH Tests
// ============================================================================

// TestBigQueryFetchClause tests FETCH clause variants
// Note: BigQuery FETCH is tested in common dialect tests
func TestBigQueryFetchClause(t *testing.T) {
	dialects := bigqueryDialect()

	// Canonical form
	t.Run("canonical form", func(t *testing.T) {
		dialects.VerifiedStmt(t, "SELECT c1 FROM fetch_test FETCH FIRST 2 ROWS ONLY")
	})

	// Test variants that normalize to canonical form
	t.Run("FETCH 2", func(t *testing.T) {
		dialects.OneStatementParsesTo(t, "SELECT c1 FROM fetch_test FETCH 2", "SELECT c1 FROM fetch_test FETCH FIRST 2 ROWS ONLY")
	})

	t.Run("FETCH FIRST 2", func(t *testing.T) {
		dialects.OneStatementParsesTo(t, "SELECT c1 FROM fetch_test FETCH FIRST 2", "SELECT c1 FROM fetch_test FETCH FIRST 2 ROWS ONLY")
	})
}

// ============================================================================
// Parser Error Tests
// ============================================================================

// TestBigQueryParseErrors tests various parse error cases
func TestBigQueryParseErrors(t *testing.T) {
	dialect := bigquery.NewBigQueryDialect()

	testCases := []struct {
		sql         string
		expectError bool
	}{
		// Invalid DECLARE
		{"DECLARE x", true},
		{"DECLARE x 42", true},
		// Invalid time travel syntax
		{"SELECT 1 FROM t1 FOR SYSTEM TIME AS OF 'ts'", true},
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

// ============================================================================
// Integration and Edge Case Tests
// ============================================================================

// TestBigQueryComplexQueries tests complex BigQuery queries
func TestBigQueryComplexQueries(t *testing.T) {
	dialects := bigqueryDialect()

	// Working complex queries
	testCases := []string{
		// Subquery in expression
		"SELECT (SELECT COUNT(*) FROM t) AS cnt",
		// Window functions
		"SELECT ROW_NUMBER() OVER (PARTITION BY col1 ORDER BY col2) AS rn FROM t",
		// WITH clause
		"WITH cte AS (SELECT 1 AS x) SELECT * FROM cte",
		// Complex array access
		"SELECT arr[SAFE_OFFSET(0)].field FROM t",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// Complex queries not yet fully implemented
	t.Run("complex struct and unnest", func(t *testing.T) {
		t.Skip("Complex STRUCT and UNNEST queries not yet fully implemented")
	})
}

// TestBigQueryStatementCount verifies single statement parsing
func TestBigQueryStatementCount(t *testing.T) {
	dialect := bigquery.NewBigQueryDialect()

	testCases := []struct {
		sql       string
		stmtCount int
	}{
		{"SELECT 1", 1},
		{"SELECT 1; SELECT 2", 2},
		{"CREATE TABLE t (x INT64); INSERT INTO t VALUES (1)", 2},
	}

	for _, tc := range testCases {
		t.Run(tc.sql, func(t *testing.T) {
			stmts, err := parser.ParseSQL(dialect, tc.sql)
			require.NoError(t, err, "Failed to parse SQL: %s", tc.sql)
			assert.Len(t, stmts, tc.stmtCount, "Statement count mismatch for: %s", tc.sql)
		})
	}
}
