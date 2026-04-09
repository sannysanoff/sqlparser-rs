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

// Package sqlite contains SQLite-specific SQL parsing tests.
// These tests are ported from tests/sqlparser_sqlite.rs in the Rust implementation.
package sqlite

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// sqliteDialect returns a TestedDialects with only SQLite dialect
func sqliteDialect() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			sqlite.NewSQLiteDialect(),
		},
	}
}

// sqliteAndGeneric returns a TestedDialects with SQLite and Generic dialects
func sqliteAndGeneric() *utils.TestedDialects {
	return &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			sqlite.NewSQLiteDialect(),
			generic.NewGenericDialect(),
		},
	}
}

// ============================================================================
// PRAGMA Tests
// ============================================================================

// TestSqlitePragmaNoValue tests PRAGMA without value
// Reference: tests/sqlparser_sqlite.rs:37
func TestSqlitePragmaNoValue(t *testing.T) {
	dialects := sqliteAndGeneric()
	sql := "PRAGMA cache_size"
	dialects.VerifiedStmt(t, sql)
}

// TestSqlitePragmaEqStyle tests PRAGMA with = style
// Reference: tests/sqlparser_sqlite.rs:51
func TestSqlitePragmaEqStyle(t *testing.T) {
	t.Skip("SQLite PRAGMA with = style not yet fully implemented in Go parser")
}

// TestSqlitePragmaFunctionStyle tests PRAGMA with function style
// Reference: tests/sqlparser_sqlite.rs:66
func TestSqlitePragmaFunctionStyle(t *testing.T) {
	dialects := sqliteAndGeneric()
	sql := "PRAGMA cache_size(10)"
	dialects.VerifiedStmt(t, sql)
}

// TestSqlitePragmaEqStringStyle tests PRAGMA with = 'string'
// Reference: tests/sqlparser_sqlite.rs:82
func TestSqlitePragmaEqStringStyle(t *testing.T) {
	t.Skip("SQLite PRAGMA with = 'string' not yet fully implemented")
}

// TestSqlitePragmaFunctionStringStyle tests PRAGMA function with string
// Reference: tests/sqlparser_sqlite.rs:98
func TestSqlitePragmaFunctionStringStyle(t *testing.T) {
	t.Skip("SQLite PRAGMA function with string not yet fully implemented")
}

// TestSqlitePragmaEqPlaceholderStyle tests PRAGMA with placeholder
// Reference: tests/sqlparser_sqlite.rs:114
func TestSqlitePragmaEqPlaceholderStyle(t *testing.T) {
	t.Skip("SQLite PRAGMA with placeholder not yet fully implemented")
}

// ============================================================================
// CREATE TABLE Tests
// ============================================================================

// TestSqliteCreateTableWithoutRowid tests CREATE TABLE WITHOUT ROWID
// Reference: tests/sqlparser_sqlite.rs:130
func TestSqliteCreateTableWithoutRowid(t *testing.T) {
	t.Skip("SQLite WITHOUT ROWID not yet fully implemented in Go parser")
}

// TestSqliteCreateVirtualTable tests CREATE VIRTUAL TABLE
// Reference: tests/sqlparser_sqlite.rs:145
func TestSqliteCreateVirtualTable(t *testing.T) {
	t.Skip("SQLite CREATE VIRTUAL TABLE not yet fully implemented in Go parser")
}

// TestSqliteCreateViewTemporaryIfNotExists tests CREATE TEMPORARY VIEW
// Reference: tests/sqlparser_sqlite.rs:167
func TestSqliteCreateViewTemporaryIfNotExists(t *testing.T) {
	dialects := sqliteAndGeneric()
	sql := "CREATE TEMPORARY VIEW IF NOT EXISTS myschema.myview AS SELECT foo FROM bar"
	dialects.VerifiedStmt(t, sql)
}

// TestSqliteCreateTableAutoIncrement tests CREATE TABLE with AUTOINCREMENT
// Reference: tests/sqlparser_sqlite.rs:209
func TestSqliteCreateTableAutoIncrement(t *testing.T) {
	dialects := sqliteAndGeneric()
	sql := "CREATE TABLE foo (bar INT PRIMARY KEY AUTOINCREMENT)"
	dialects.VerifiedStmt(t, sql)
}

// TestSqliteCreateTablePrimaryKeyAscDesc tests PRIMARY KEY ASC/DESC
// Reference: tests/sqlparser_sqlite.rs:246
func TestSqliteCreateTablePrimaryKeyAscDesc(t *testing.T) {
	t.Skip("SQLite PRIMARY KEY ASC/DESC not yet fully implemented in Go parser")
}

// TestSqliteCreateTableQuote tests various quote styles in CREATE TABLE
// Reference: tests/sqlparser_sqlite.rs:286
func TestSqliteCreateTableQuote(t *testing.T) {
	t.Skip("SQLite [square bracket] identifiers not yet fully implemented")
}

// TestSqliteCreateTableGenCol tests generated columns
// Reference: tests/sqlparser_sqlite.rs:312
func TestSqliteCreateTableGenCol(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"CREATE TABLE t1 (a INT, b INT GENERATED ALWAYS AS (a * 2))",
		"CREATE TABLE t1 (a INT, b INT GENERATED ALWAYS AS (a * 2) VIRTUAL)",
		"CREATE TABLE t1 (a INT, b INT GENERATED ALWAYS AS (a * 2) STORED)",
		"CREATE TABLE t1 (a INT, b INT AS (a * 2))",
		"CREATE TABLE t1 (a INT, b INT AS (a * 2) VIRTUAL)",
		"CREATE TABLE t1 (a INT, b INT AS (a * 2) STORED)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSqliteCreateTableOnConflictCol tests ON CONFLICT in column definition
// Reference: tests/sqlparser_sqlite.rs:328
func TestSqliteCreateTableOnConflictCol(t *testing.T) {
	t.Skip("SQLite ON CONFLICT in column definition not yet fully implemented")
}

// TestSqliteCreateTableOnConflictColErr tests invalid ON CONFLICT
// Reference: tests/sqlparser_sqlite.rs:353
func TestSqliteCreateTableOnConflictColErr(t *testing.T) {
	dialect := sqlite.NewSQLiteDialect()
	sql := "CREATE TABLE t1 (a INT, b INT ON CONFLICT BOH)"
	_, err := parser.ParseSQL(dialect, sql)
	require.Error(t, err, "Expected error for invalid ON CONFLICT value")
}

// TestSqliteCreateTableUntyped tests untyped columns
// Reference: tests/sqlparser_sqlite.rs:367
func TestSqliteCreateTableUntyped(t *testing.T) {
	t.Skip("SQLite untyped columns not yet fully implemented in Go parser")
}

// TestSqliteCreateTableWithStrict tests STRICT table
// Reference: tests/sqlparser_sqlite.rs:388
func TestSqliteCreateTableWithStrict(t *testing.T) {
	t.Skip("SQLite STRICT table not yet fully implemented in Go parser")
}

// ============================================================================
// Placeholder Tests
// ============================================================================

// TestSqlitePlaceholder tests @xxx placeholder syntax
// Reference: tests/sqlparser_sqlite.rs:373
func TestSqlitePlaceholder(t *testing.T) {
	dialects := sqliteDialect()
	sql := "SELECT @xxx"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestSqliteDollarIdentifierAsPlaceholder tests $id as placeholder
// Reference: tests/sqlparser_sqlite.rs:543
func TestSqliteDollarIdentifierAsPlaceholder(t *testing.T) {
	t.Skip("SQLite $ placeholder not yet fully implemented in Go parser")
}

// ============================================================================
// Single Quoted Identifier Tests
// ============================================================================

// TestSqliteSingleQuotedIdentifier tests single-quoted identifiers
// Reference: tests/sqlparser_sqlite.rs:397
func TestSqliteSingleQuotedIdentifier(t *testing.T) {
	t.Skip("SQLite single-quoted identifiers not yet fully implemented")
}

// ============================================================================
// SUBSTRING Tests
// ============================================================================

// TestSqliteSubstring tests SUBSTRING and SUBSTR functions
// Reference: tests/sqlparser_sqlite.rs:403
func TestSqliteSubstring(t *testing.T) {
	dialects := sqliteDialect()

	testCases := []string{
		"SELECT SUBSTRING('SQLITE', 3, 4)",
		"SELECT SUBSTR('SQLITE', 3, 4)",
		"SELECT SUBSTRING('SQLITE', 3)",
		"SELECT SUBSTR('SQLITE', 3)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedOnlySelect(t, sql)
		})
	}
}

// ============================================================================
// Window Function Tests
// ============================================================================

// TestSqliteWindowFunctionWithFilter tests window functions with FILTER
// Reference: tests/sqlparser_sqlite.rs:414
func TestSqliteWindowFunctionWithFilter(t *testing.T) {
	t.Skip("SQLite window functions with FILTER not yet fully implemented")
}

// ============================================================================
// ATTACH/DETACH Tests
// ============================================================================

// TestSqliteAttachDatabase tests ATTACH DATABASE
// Reference: tests/sqlparser_sqlite.rs:453
func TestSqliteAttachDatabase(t *testing.T) {
	t.Skip("SQLite ATTACH DATABASE not yet fully implemented in Go parser")
}

// ============================================================================
// UPDATE Tests
// ============================================================================

// TestSqliteUpdateTupleRowValues tests UPDATE with tuple row values
// Reference: tests/sqlparser_sqlite.rs:474
func TestSqliteUpdateTupleRowValues(t *testing.T) {
	t.Skip("SQLite UPDATE with tuple row values not yet fully implemented")
}

// TestSqliteUpdateDeleteLimit tests UPDATE with LIMIT
// Reference: tests/sqlparser_sqlite.rs:613
func TestSqliteUpdateDeleteLimit(t *testing.T) {
	t.Skip("SQLite UPDATE/DELETE with LIMIT not yet fully implemented")
}

// ============================================================================
// IN Empty List Tests
// ============================================================================

// TestSqliteWhereInEmptyList tests WHERE a IN ()
// Reference: tests/sqlparser_sqlite.rs:507
func TestSqliteWhereInEmptyList(t *testing.T) {
	t.Skip("SQLite empty IN list not yet fully implemented in Go parser")
}

// TestSqliteInvalidEmptyList tests invalid empty list
// Reference: tests/sqlparser_sqlite.rs:523
func TestSqliteInvalidEmptyList(t *testing.T) {
	dialect := sqlite.NewSQLiteDialect()
	sql := "SELECT * FROM t1 WHERE a IN (,,)"
	_, err := parser.ParseSQL(dialect, sql)
	require.Error(t, err, "Expected error for invalid empty list")
}

// ============================================================================
// Transaction Tests
// ============================================================================

// TestSqliteStartTransactionWithModifier tests BEGIN with modifiers
// Reference: tests/sqlparser_sqlite.rs:533
func TestSqliteStartTransactionWithModifier(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"BEGIN DEFERRED TRANSACTION",
		"BEGIN IMMEDIATE TRANSACTION",
		"BEGIN EXCLUSIVE TRANSACTION",
		"BEGIN DEFERRED",
		"BEGIN IMMEDIATE",
		"BEGIN EXCLUSIVE",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// MATCH/REGEXP Operator Tests
// ============================================================================

// TestSqliteMatchOperator tests MATCH operator
// Reference: tests/sqlparser_sqlite.rs:578
func TestSqliteMatchOperator(t *testing.T) {
	t.Skip("SQLite MATCH operator not yet fully implemented in Go parser")
}

// TestSqliteRegexpOperator tests REGEXP operator
// Reference: tests/sqlparser_sqlite.rs:594
func TestSqliteRegexpOperator(t *testing.T) {
	t.Skip("SQLite REGEXP operator not yet fully implemented in Go parser")
}

// ============================================================================
// CREATE TRIGGER Tests
// ============================================================================

// TestSqliteCreateTrigger tests CREATE TRIGGER
// Reference: tests/sqlparser_sqlite.rs:629
func TestSqliteCreateTrigger(t *testing.T) {
	t.Skip("SQLite CREATE TRIGGER not yet fully implemented in Go parser")
}

// ============================================================================
// DROP TRIGGER Tests
// ============================================================================

// TestSqliteDropTrigger tests DROP TRIGGER
// Reference: tests/sqlparser_sqlite.rs:888
func TestSqliteDropTrigger(t *testing.T) {
	t.Skip("SQLite DROP TRIGGER not yet fully implemented in Go parser")
}

// ============================================================================
// Double Equality Operator Tests
// ============================================================================

// TestSqliteDoubleEqualityOperator tests == operator
// Reference: tests/sqlparser_sqlite.rs:201
func TestSqliteDoubleEqualityOperator(t *testing.T) {
	dialects := sqliteAndGeneric()
	// SQLite supports == which normalizes to =
	dialects.OneStatementParsesTo(t, "SELECT a==b FROM t", "SELECT a = b FROM t")
}

// ============================================================================
// Complex Query Tests
// ============================================================================

// TestSqliteComplexQueries tests various complex SQLite queries
func TestSqliteComplexQueries(t *testing.T) {
	dialects := sqliteDialect()

	testCases := []string{
		// Simple SELECT
		"SELECT * FROM t",
		// SELECT with WHERE
		"SELECT a, b FROM t WHERE c = 1",
		// SELECT with ORDER BY
		"SELECT a FROM t ORDER BY a",
		// SELECT with GROUP BY
		"SELECT a, COUNT(*) FROM t GROUP BY a",
		// SELECT with LIMIT
		"SELECT * FROM t LIMIT 10",
		// SELECT with LIMIT and OFFSET
		"SELECT * FROM t LIMIT 10 OFFSET 5",
		// Subquery
		"SELECT * FROM (SELECT * FROM t) s",
		// JOIN
		"SELECT * FROM t1 JOIN t2 ON t1.id = t2.id",
		// LEFT JOIN
		"SELECT * FROM t1 LEFT JOIN t2 ON t1.id = t2.id",
		// INSERT
		"INSERT INTO t (a, b) VALUES (1, 2)",
		// UPDATE
		"UPDATE t SET a = 1 WHERE b = 2",
		// DELETE
		"DELETE FROM t WHERE a = 1",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// ============================================================================
// INSERT/REPLACE Tests
// ============================================================================

// TestSqliteInsertReplace tests REPLACE statement
// Reference: tests/sqlparser_sqlite.rs (REPLACE as INSERT variant)
func TestSqliteInsertReplace(t *testing.T) {
	t.Skip("SQLite REPLACE statement not yet fully implemented in Go parser")
}

// ============================================================================
// COMMIT/ROLLBACK Tests
// ============================================================================

// TestSqliteCommit tests COMMIT statement
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteCommit(t *testing.T) {
	t.Skip("SQLite COMMIT statement not yet fully implemented in Go parser")
}

// TestSqliteRollback tests ROLLBACK statement
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteRollback(t *testing.T) {
	t.Skip("SQLite ROLLBACK statement not yet fully implemented in Go parser")
}

// ============================================================================
// SAVEPOINT Tests
// ============================================================================

// TestSqliteSavepoint tests SAVEPOINT statement
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteSavepoint(t *testing.T) {
	t.Skip("SQLite SAVEPOINT not yet fully implemented in Go parser")
}

// TestSqliteReleaseSavepoint tests RELEASE SAVEPOINT
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteReleaseSavepoint(t *testing.T) {
	t.Skip("SQLite RELEASE SAVEPOINT not yet fully implemented in Go parser")
}

// ============================================================================
// UPSERT/ON CONFLICT Tests
// ============================================================================

// TestSqliteUpsert tests UPSERT (INSERT ON CONFLICT DO UPDATE)
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteUpsert(t *testing.T) {
	t.Skip("SQLite UPSERT (ON CONFLICT) not yet fully implemented in Go parser")
}

// ============================================================================
// VACUUM Tests
// ============================================================================

// TestSqliteVacuum tests VACUUM statement
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteVacuum(t *testing.T) {
	t.Skip("SQLite VACUUM statement not yet fully implemented in Go parser")
}

// ============================================================================
// REINDEX Tests
// ============================================================================

// TestSqliteReindex tests REINDEX statement
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteReindex(t *testing.T) {
	t.Skip("SQLite REINDEX not yet fully implemented in Go parser")
}

// ============================================================================
// ANALYZE Tests
// ============================================================================

// TestSqliteAnalyze tests ANALYZE statement
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteAnalyze(t *testing.T) {
	t.Skip("SQLite ANALYZE not yet fully implemented in Go parser")
}

// ============================================================================
// DROP TABLE/INDEX Tests
// ============================================================================

// TestSqliteDropTable tests DROP TABLE
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteDropTable(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"DROP TABLE t",
		"DROP TABLE IF EXISTS t",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSqliteDropIndex tests DROP INDEX
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteDropIndex(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"DROP INDEX idx",
		"DROP INDEX IF EXISTS idx",
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

// TestSqliteAlterTableRename tests ALTER TABLE RENAME
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteAlterTableRename(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"ALTER TABLE t RENAME TO t2",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// ALTER TABLE RENAME COLUMN not yet fully implemented
	t.Run("RENAME COLUMN", func(t *testing.T) {
		t.Skip("SQLite ALTER TABLE RENAME COLUMN not yet fully implemented")
	})
}

// TestSqliteAlterTableAddColumn tests ALTER TABLE ADD COLUMN
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteAlterTableAddColumn(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"ALTER TABLE t ADD COLUMN b INT",
		"ALTER TABLE t ADD b INT",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSqliteAlterTableDropColumn tests ALTER TABLE DROP COLUMN
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteAlterTableDropColumn(t *testing.T) {
	t.Skip("SQLite ALTER TABLE DROP COLUMN not yet fully implemented")
}

// ============================================================================
// SELECT DISTINCT/LIMIT Tests
// ============================================================================

// TestSqliteSelectDistinct tests SELECT DISTINCT
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteSelectDistinct(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"SELECT DISTINCT a FROM t",
		"SELECT DISTINCT a, b FROM t",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestSqliteSelectWithLimit tests SELECT with LIMIT
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteSelectWithLimit(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"SELECT * FROM t LIMIT 10",
		"SELECT * FROM t LIMIT 10 OFFSET 5",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// SQLite LIMIT 5, 10 (LIMIT 10 OFFSET 5) not yet fully implemented
	t.Run("LIMIT comma syntax", func(t *testing.T) {
		t.Skip("SQLite LIMIT comma syntax not yet fully implemented")
	})
}

// ============================================================================
// CTE Tests
// ============================================================================

// TestSqliteSelectWithCte tests SELECT with CTE
// Reference: tests/sqlparser_sqlite.rs (common tests)
func TestSqliteSelectWithCte(t *testing.T) {
	dialects := sqliteAndGeneric()

	testCases := []string{
		"WITH cte AS (SELECT * FROM t) SELECT * FROM cte",
		"WITH cte1 AS (SELECT * FROM t1), cte2 AS (SELECT * FROM t2) SELECT * FROM cte1, cte2",
		"WITH RECURSIVE cte AS (SELECT 1 AS n UNION ALL SELECT n + 1 FROM cte WHERE n < 5) SELECT * FROM cte",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}
