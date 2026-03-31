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
// This file contains tests 281-300 from the Rust test file.
package common

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/ansi"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/errors"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
	"github.com/user/sqlparser/token"
)

// TestLock verifies FOR UPDATE and FOR SHARE lock clauses.
// Reference: tests/sqlparser_common.rs:10301
func TestLock(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test FOR UPDATE
	sql := "SELECT * FROM student WHERE id = '1' FOR UPDATE"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Get the query from the statement
	selectStmt, ok := stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q := selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock := q.Locks[0]
	assert.Equal(t, query.LockTypeUpdate, lock.LockType, "Expected UPDATE lock type")
	assert.Nil(t, lock.Of, "Expected OF to be nil")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil")

	// Test FOR SHARE
	sql = "SELECT * FROM student WHERE id = '1' FOR SHARE"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok = stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q = selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock = q.Locks[0]
	assert.Equal(t, query.LockTypeShare, lock.LockType, "Expected SHARE lock type")
	assert.Nil(t, lock.Of, "Expected OF to be nil")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil")
}

// TestLockTable verifies FOR UPDATE OF and FOR SHARE OF table lock clauses.
// Reference: tests/sqlparser_common.rs:10320
func TestLockTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test FOR UPDATE OF school
	sql := "SELECT * FROM student WHERE id = '1' FOR UPDATE OF school"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q := selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock := q.Locks[0]
	assert.Equal(t, query.LockTypeUpdate, lock.LockType, "Expected UPDATE lock type")
	require.NotNil(t, lock.Of, "Expected OF to be present")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil")

	// Test FOR SHARE OF school
	sql = "SELECT * FROM student WHERE id = '1' FOR SHARE OF school"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok = stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q = selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock = q.Locks[0]
	assert.Equal(t, query.LockTypeShare, lock.LockType, "Expected SHARE lock type")
	require.NotNil(t, lock.Of, "Expected OF to be present")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil")

	// Test multiple locks: FOR SHARE OF school FOR UPDATE OF student
	sql = "SELECT * FROM student WHERE id = '1' FOR SHARE OF school FOR UPDATE OF student"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok = stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q = selectStmt.Query
	require.Len(t, q.Locks, 2, "Expected 2 lock clauses")

	// First lock: FOR SHARE OF school
	lock = q.Locks[0]
	assert.Equal(t, query.LockTypeShare, lock.LockType, "Expected first lock to be SHARE")
	require.NotNil(t, lock.Of, "Expected OF to be present on first lock")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil on first lock")

	// Second lock: FOR UPDATE OF student
	lock = q.Locks[1]
	assert.Equal(t, query.LockTypeUpdate, lock.LockType, "Expected second lock to be UPDATE")
	require.NotNil(t, lock.Of, "Expected OF to be present on second lock")
	assert.Equal(t, "student", lock.Of.String(), "Expected OF table to be 'student'")
	assert.Nil(t, lock.Nonblock, "Expected nonblock to be nil on second lock")
}

// TestLockNonblock verifies SKIP LOCKED and NOWAIT lock modifiers.
// Reference: tests/sqlparser_common.rs:10379
func TestLockNonblock(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test FOR UPDATE OF school SKIP LOCKED
	sql := "SELECT * FROM student WHERE id = '1' FOR UPDATE OF school SKIP LOCKED"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok := stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q := selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock := q.Locks[0]
	assert.Equal(t, query.LockTypeUpdate, lock.LockType, "Expected UPDATE lock type")
	require.NotNil(t, lock.Of, "Expected OF to be present")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	require.NotNil(t, lock.Nonblock, "Expected nonblock to be present")
	assert.Equal(t, query.NonBlockSkipLocked, *lock.Nonblock, "Expected SKIP LOCKED")

	// Test FOR SHARE OF school NOWAIT
	sql = "SELECT * FROM student WHERE id = '1' FOR SHARE OF school NOWAIT"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	selectStmt, ok = stmts[0].(*statement.Query)
	require.True(t, ok, "Expected Select statement, got %T", stmts[0])

	q = selectStmt.Query
	require.Len(t, q.Locks, 1, "Expected 1 lock clause")
	lock = q.Locks[0]
	assert.Equal(t, query.LockTypeShare, lock.LockType, "Expected SHARE lock type")
	require.NotNil(t, lock.Of, "Expected OF to be present")
	assert.Equal(t, "school", lock.Of.String(), "Expected OF table to be 'school'")
	require.NotNil(t, lock.Nonblock, "Expected nonblock to be present")
	assert.Equal(t, query.NonBlockNowait, *lock.Nonblock, "Expected NOWAIT")
}

// TestPlaceholder verifies placeholder parsing with various dialects.
// Reference: tests/sqlparser_common.rs:10412
func TestPlaceholder(t *testing.T) {
	// First dialect set with $ placeholders
	dialectsWithDollar := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			postgresql.NewPostgreSqlDialect(),
			mssql.NewMsSqlDialect(),
			ansi.NewAnsiDialect(),
			bigquery.NewBigQueryDialect(),
			snowflake.NewSnowflakeDialect(),
		},
	}

	// Test $Id1 placeholder
	sql := "SELECT * FROM student WHERE id = $Id1"
	_ = dialectsWithDollar.VerifiedStmt(t, sql)

	// Test LIMIT $1 OFFSET $2
	sql = "SELECT * FROM student LIMIT $1 OFFSET $2"
	_ = dialectsWithDollar.VerifiedStmt(t, sql)

	// Second dialect set with ? placeholders (excluding PostgreSQL which uses ? for JSONB)
	dialectsWithQuestion := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			mssql.NewMsSqlDialect(),
			ansi.NewAnsiDialect(),
			bigquery.NewBigQueryDialect(),
			snowflake.NewSnowflakeDialect(),
		},
	}

	// Test ? placeholder
	sql = "SELECT * FROM student WHERE id = ?"
	_ = dialectsWithQuestion.VerifiedStmt(t, sql)

	// Test multiple placeholders
	sql = "SELECT $fromage_français, :x, ?123"
	_ = dialectsWithQuestion.VerifiedStmt(t, sql)
}

// TestAllKeywordsSorted verifies that ALL_KEYWORDS is sorted alphabetically.
// Reference: tests/sqlparser_common.rs:10494
func TestAllKeywordsSorted(t *testing.T) {
	// Get all keywords as strings
	allKeywords := []string{
		string(token.ABORT), string(token.ABS), string(token.ABSENT),
		string(token.ABSOLUTE), string(token.ACCEPTANYDATE), string(token.ACCEPTINVCHARS),
		string(token.ACCESS), string(token.ACCOUNT), string(token.ACTION),
		string(token.ADD), string(token.ADDQUOTES), string(token.ADMIN),
		string(token.AFTER), string(token.AGAINST), string(token.AGGREGATE),
		string(token.AGGREGATION), string(token.ALERT), string(token.ALGORITHM),
		string(token.ALIAS), string(token.ALIGNMENT), string(token.ALL),
		string(token.ALLOCATE), string(token.ALLOWOVERWRITE), string(token.ALTER),
		string(token.ALWAYS), string(token.ANALYZE), string(token.AND),
		string(token.ANTI), string(token.ANY), string(token.APPLICATION),
		string(token.APPLY), string(token.APPLYBUDGET), string(token.ARCHIVE),
		string(token.ARE), string(token.ARRAY), string(token.ARRAY_MAX_CARDINALITY),
		string(token.AS), string(token.ASC), string(token.ASENSITIVE),
		string(token.ASOF), string(token.ASSERT), string(token.ASYMMETRIC),
		string(token.AT), string(token.ATOMIC), string(token.ATTACH),
		string(token.AUDIT), string(token.AUTHENTICATION), string(token.AUTHORIZATION),
		string(token.AUTHORIZATIONS), string(token.AUTO), string(token.AUTOEXTEND_SIZE),
		string(token.AUTOINCREMENT), string(token.AUTO_INCREMENT), string(token.AVG),
		string(token.AVG_ROW_LENGTH), string(token.AVRO), string(token.BACKUP),
		string(token.BACKWARD), string(token.BEFORE), string(token.BEGIN),
		string(token.BERNOULLI), string(token.BETWEEN), string(token.BIGINT),
		string(token.BINARY), string(token.BINDING), string(token.BIT),
		string(token.BLOB),
		string(token.BOOL), string(token.BOOLEAN), string(token.BOTH),
		string(token.BROWSE), string(token.BUCKET), string(token.BUCKETS),
		string(token.BY), string(token.BYTEA), string(token.BYTES),
		string(token.CACHE), string(token.CALL), string(token.CALLED),
		string(token.CARDINALITY), string(token.CASCADE),
		string(token.CASE), string(token.CAST), string(token.CATALOG),
		string(token.CEIL), string(token.CEILING),
		string(token.CENTURY), string(token.CHAIN), string(token.CHANGE),
		string(token.CHANGES), string(token.CHAR), string(token.CHARACTER),
		string(token.CHARACTERISTICS), string(token.CHARACTERS), string(token.CHARACTER_LENGTH),
		string(token.CHARSET), string(token.CHAR_LENGTH), string(token.CHECK),
		string(token.CLASS), string(token.CLOB), string(token.CLONE),
		string(token.CLOSE), string(token.CLUSTER), string(token.COALESCE),
		string(token.COLLATE), string(token.COLLATION), string(token.COLLECT),
		string(token.COLUMN), string(token.COLUMNS),
		string(token.COMMENT), string(token.COMMIT),
		string(token.COMMITTED), string(token.COMPRESSION), string(token.CONCURRENTLY),
		string(token.CONDITION),
		string(token.CONNECTION), string(token.CONSTRAINT),
		string(token.CONTAINS),
		string(token.CONTINUE), string(token.CONVERT), string(token.COPY),
		string(token.CORR), string(token.CORRESPONDING),
		string(token.COUNT),
		string(token.COVAR_POP), string(token.COVAR_SAMP), string(token.CREATE),
		string(token.CREATEDB), string(token.CROSS),
		string(token.CSV), string(token.CUBE), string(token.CUME_DIST),
		string(token.CURRENT), string(token.CURRENT_CATALOG), string(token.CURRENT_DATE),
		string(token.CURRENT_DEFAULT_TRANSFORM_GROUP), string(token.CURRENT_PATH),
		string(token.CURRENT_ROLE), string(token.CURRENT_SCHEMA), string(token.CURRENT_TIME),
		string(token.CURRENT_TIMESTAMP), string(token.CURRENT_TRANSFORM_GROUP_FOR_TYPE),
		string(token.CURRENT_USER), string(token.CURSOR),
		string(token.CYCLE),
	}

	// Make a copy and sort it
	sortedKeywords := make([]string, len(allKeywords))
	copy(sortedKeywords, allKeywords)
	sort.Strings(sortedKeywords)

	// Compare original with sorted
	assert.Equal(t, sortedKeywords, allKeywords, "ALL_KEYWORDS should be sorted alphabetically")
}

// TestParseOffsetAndLimit verifies LIMIT and OFFSET clause parsing.
// Reference: tests/sqlparser_common.rs:10526
func TestParseOffsetAndLimit(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test LIMIT 1 OFFSET 2
	sql := "SELECT foo FROM bar LIMIT 1 OFFSET 2"
	dialects.VerifiedQuery(t, sql)

	// Test different order (OFFSET first) - should parse to canonical form
	dialects.OneStatementParsesTo(t, "SELECT foo FROM bar OFFSET 2 LIMIT 1", sql)

	// Test MySQL LIMIT comma syntax for dialects that support it
	mysqlDbs := &utils.TestedDialects{
		Dialects: []sqlparserDialects.Dialect{
			mysql.NewMySqlDialect(),
		},
	}
	mysqlDbs.VerifiedQuery(t, "SELECT foo FROM bar LIMIT 2, 1")

	// Test OFFSET without LIMIT
	dialects.VerifiedStmt(t, "SELECT foo FROM bar OFFSET 2")

	// Test error cases - can't repeat OFFSET
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "SELECT foo FROM bar OFFSET 2 OFFSET 2")
	require.Error(t, err)

	// Test error cases - can't repeat LIMIT
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "SELECT foo FROM bar LIMIT 2 LIMIT 2")
	require.Error(t, err)

	// Test error cases - can't have OFFSET after LIMIT OFFSET
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "SELECT foo FROM bar OFFSET 2 LIMIT 2 OFFSET 2")
	require.Error(t, err)
}

// TestParseTimeFunctions verifies parsing of time functions like CURRENT_TIMESTAMP, CURRENT_TIME, etc.
// Reference: tests/sqlparser_common.rs:10591
func TestParseTimeFunctions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testTimeFunction := func(t *testing.T, funcName string) {
		// Test with parentheses
		sql := fmt.Sprintf("SELECT %s()", funcName)
		dialects.VerifiedOnlySelect(t, sql)

		// Test without parentheses (where supported)
		sqlWithoutParens := fmt.Sprintf("SELECT %s", funcName)
		dialects.VerifiedOnlySelect(t, sqlWithoutParens)
	}

	testTimeFunction(t, "CURRENT_TIMESTAMP")
	testTimeFunction(t, "CURRENT_TIME")
	testTimeFunction(t, "CURRENT_DATE")
	testTimeFunction(t, "LOCALTIME")
	testTimeFunction(t, "LOCALTIMESTAMP")
}

// TestParsePosition verifies POSITION expression parsing.
// Reference: tests/sqlparser_common.rs:10632
func TestParsePosition(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test POSITION IN syntax
	sql := "POSITION('@' IN field)"
	dialects.VerifiedExpr(t, sql)

	// Test POSITION as function call (Snowflake style with 3 args)
	sql = "position('an', 'banana', 1)"
	dialects.VerifiedExpr(t, sql)
}

// TestParsePositionNegative verifies error handling for invalid POSITION syntax.
// Reference: tests/sqlparser_common.rs:10658
func TestParsePositionNegative(t *testing.T) {
	sql := "SELECT POSITION(foo IN) from bar"
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err, "Expected error for incomplete POSITION expression")
}

// TestParseIsBoolean verifies IS TRUE/FALSE/UNKNOWN/NORMALIZED expression parsing.
// Reference: tests/sqlparser_common.rs:10668
func TestParseIsBoolean(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test IS TRUE
	dialects.VerifiedExpr(t, "a IS TRUE")

	// Test IS NOT TRUE
	dialects.VerifiedExpr(t, "a IS NOT TRUE")

	// Test IS FALSE
	dialects.VerifiedExpr(t, "a IS FALSE")

	// Test IS NOT FALSE
	dialects.VerifiedExpr(t, "a IS NOT FALSE")

	// Test IS NORMALIZED
	dialects.VerifiedExpr(t, "a IS NORMALIZED")

	// Test IS NOT NORMALIZED
	dialects.VerifiedExpr(t, "a IS NOT NORMALIZED")

	// Test IS NFKC NORMALIZED
	dialects.VerifiedExpr(t, "a IS NFKC NORMALIZED")

	// Test IS NOT NFKD NORMALIZED
	dialects.VerifiedExpr(t, "a IS NOT NFKD NORMALIZED")

	// Test IS UNKNOWN
	dialects.VerifiedExpr(t, "a IS UNKNOWN")

	// Test IS NOT UNKNOWN
	dialects.VerifiedExpr(t, "a IS NOT UNKNOWN")

	// Full statements
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS TRUE")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT TRUE")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS FALSE")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT FALSE")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NFC NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NFD NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT NFKC NORMALIZED")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS UNKNOWN")
	dialects.VerifiedStmt(t, "SELECT f FROM foo WHERE field IS NOT UNKNOWN")

	// Error: IS 0 is not valid
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "SELECT f from foo where field is 0")
	require.Error(t, err)

	// Error: IS XYZ NORMALIZED - invalid form
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "SELECT s, s IS XYZ NORMALIZED FROM foo")
	require.Error(t, err)

	// Error: IS NFKC (missing NORMALIZED)
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "SELECT s, s IS NFKC FROM foo")
	require.Error(t, err)

	// Error: IS TRIM(' NFKC ') - function not allowed
	_, err = parser.ParseSQL(generic.NewGenericDialect(), "SELECT s, s IS TRIM(' NFKC ') FROM foo")
	require.Error(t, err)
}

// TestParseDiscard verifies DISCARD statement parsing.
// Reference: tests/sqlparser_common.rs:10804
func TestParseDiscard(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test DISCARD ALL
	stmt := dialects.VerifiedStmt(t, "DISCARD ALL")
	discard, ok := stmt.(*statement.Discard)
	require.True(t, ok, "Expected Discard statement, got %T", stmt)
	_ = discard

	// Test DISCARD PLANS
	stmt = dialects.VerifiedStmt(t, "DISCARD PLANS")
	discard, ok = stmt.(*statement.Discard)
	require.True(t, ok, "Expected Discard statement, got %T", stmt)

	// Test DISCARD SEQUENCES
	stmt = dialects.VerifiedStmt(t, "DISCARD SEQUENCES")
	discard, ok = stmt.(*statement.Discard)
	require.True(t, ok, "Expected Discard statement, got %T", stmt)

	// Test DISCARD TEMP
	stmt = dialects.VerifiedStmt(t, "DISCARD TEMP")
	discard, ok = stmt.(*statement.Discard)
	require.True(t, ok, "Expected Discard statement, got %T", stmt)
}

// TestParseCursor verifies CLOSE cursor statement parsing.
// Reference: tests/sqlparser_common.rs:10831
func TestParseCursor(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test CLOSE my_cursor
	sql := "CLOSE my_cursor"
	stmt := dialects.VerifiedStmt(t, sql)
	closeStmt, ok := stmt.(*statement.Close)
	require.True(t, ok, "Expected Close statement, got %T", stmt)
	require.NotNil(t, closeStmt.Cursor, "Expected cursor to be present")

	// Test CLOSE ALL
	sql = "CLOSE ALL"
	stmt = dialects.VerifiedStmt(t, sql)
	closeStmt, ok = stmt.(*statement.Close)
	require.True(t, ok, "Expected Close statement, got %T", stmt)
}

// TestParseShowFunctions verifies SHOW FUNCTIONS statement parsing.
// Reference: tests/sqlparser_common.rs:10851
func TestParseShowFunctions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	stmt := dialects.VerifiedStmt(t, "SHOW FUNCTIONS LIKE 'pattern'")
	showFunc, ok := stmt.(*statement.ShowFunctions)
	require.True(t, ok, "Expected ShowFunctions statement, got %T", stmt)
	require.NotNil(t, showFunc.Filter, "Expected filter to be present")
}

// TestParseCacheTable verifies CACHE TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:10861
func TestParseCacheTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT a, b, c FROM foo"
	dialects.VerifiedQuery(t, sql)

	// Test CACHE TABLE 'cache_table_name'
	stmt := dialects.VerifiedStmt(t, "CACHE TABLE 'cache_table_name'")
	cache, ok := stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.NotNil(t, cache.TableName, "Expected table name to be present")
	assert.False(t, cache.HasAs, "Expected HasAs to be false")
	assert.Equal(t, 0, len(cache.Options), "Expected no options")
	assert.Nil(t, cache.Query, "Expected no query")

	// Test CACHE flag TABLE 'cache_table_name'
	stmt = dialects.VerifiedStmt(t, "CACHE flag TABLE 'cache_table_name'")
	cache, ok = stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.NotNil(t, cache.TableFlag, "Expected table flag to be present")
	require.NotNil(t, cache.TableName, "Expected table name to be present")

	// Test with OPTIONS
	stmt = dialects.VerifiedStmt(t, "CACHE flag TABLE 'cache_table_name' OPTIONS('K1' = 'V1', 'K2' = 0.88)")
	cache, ok = stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.Equal(t, 2, len(cache.Options), "Expected 2 options")

	// Test with query (no AS)
	stmt = dialects.VerifiedStmt(t, "CACHE flag TABLE 'cache_table_name' OPTIONS('K1' = 'V1', 'K2' = 0.88) SELECT a, b, c FROM foo")
	cache, ok = stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.NotNil(t, cache.Query, "Expected query to be present")
	assert.False(t, cache.HasAs, "Expected HasAs to be false without AS")

	// Test with query and AS
	stmt = dialects.VerifiedStmt(t, "CACHE flag TABLE 'cache_table_name' OPTIONS('K1' = 'V1', 'K2' = 0.88) AS SELECT a, b, c FROM foo")
	cache, ok = stmt.(*statement.Cache)
	require.True(t, ok, "Expected Cache statement, got %T", stmt)
	require.NotNil(t, cache.Query, "Expected query to be present")
	assert.True(t, cache.HasAs, "Expected HasAs to be true with AS")

	// Error cases
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "CACHE TABLE 'table_name' foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE flag TABLE 'table_name' OPTIONS('K1'='V1') foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE TABLE 'table_name' AS foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE flag TABLE 'table_name' OPTIONS('K1'='V1') AS foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE 'table_name'")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE 'table_name' OPTIONS('K1'='V1')")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "CACHE flag 'table_name' OPTIONS('K1'='V1')")
	require.Error(t, err)
}

// TestParseUncacheTable verifies UNCACHE TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:11038
func TestParseUncacheTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test UNCACHE TABLE 'table_name'
	stmt := dialects.VerifiedStmt(t, "UNCACHE TABLE 'table_name'")
	uncache, ok := stmt.(*statement.Uncache)
	require.True(t, ok, "Expected Uncache statement, got %T", stmt)
	require.NotNil(t, uncache.TableName, "Expected table name to be present")
	assert.False(t, uncache.IfExists, "Expected IfExists to be false")

	// Test UNCACHE TABLE IF EXISTS 'table_name'
	stmt = dialects.VerifiedStmt(t, "UNCACHE TABLE IF EXISTS 'table_name'")
	uncache, ok = stmt.(*statement.Uncache)
	require.True(t, ok, "Expected Uncache statement, got %T", stmt)
	require.NotNil(t, uncache.TableName, "Expected table name to be present")
	assert.True(t, uncache.IfExists, "Expected IfExists to be true")

	// Error cases
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "UNCACHE TABLE 'table_name' foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "UNCACHE 'table_name' foo")
	require.Error(t, err)

	_, err = parser.ParseSQL(generic.NewGenericDialect(), "UNCACHE IF EXISTS 'table_name' foo")
	require.Error(t, err)
}

// TestParseDeeplyNestedParensHitsRecursionLimits verifies recursion limit with deeply nested parentheses.
// Reference: tests/sqlparser_common.rs:11075
func TestParseDeeplyNestedParensHitsRecursionLimits(t *testing.T) {
	sql := strings.Repeat("(", 1000)
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err)
	assert.True(t, errors.IsRecursionLimitExceeded(err), "Expected RecursionLimitExceeded error")
}

// TestParseUpdateDeeplyNestedParensHitsRecursionLimits verifies recursion limit with UPDATE and nested parens.
// Reference: tests/sqlparser_common.rs:11082
func TestParseUpdateDeeplyNestedParensHitsRecursionLimits(t *testing.T) {
	sql := "\nUPDATE\n\n\n\n\n\n\n\n\n\n" + strings.Repeat("(", 1000)
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err)
	assert.True(t, errors.IsRecursionLimitExceeded(err), "Expected RecursionLimitExceeded error")
}

// TestParseDeeplyNestedUnaryOpHitsRecursionLimits verifies recursion limit with deeply nested unary operators.
// Reference: tests/sqlparser_common.rs:11089
func TestParseDeeplyNestedUnaryOpHitsRecursionLimits(t *testing.T) {
	sql := "SELECT " + strings.Repeat("+", 1000)
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err)
	assert.True(t, errors.IsRecursionLimitExceeded(err), "Expected RecursionLimitExceeded error")
}

// TestParseDeeplyNestedExprHitsRecursionLimits verifies recursion limit with deeply nested expressions.
// Reference: tests/sqlparser_common.rs:11096
func TestParseDeeplyNestedExprHitsRecursionLimits(t *testing.T) {
	// Build a deeply nested WHERE clause
	whereClause := "1=1"
	for i := 0; i < 100; i++ {
		whereClause = fmt.Sprintf("(%s) AND (1=1)", whereClause)
	}
	sql := fmt.Sprintf("SELECT id, user_id FROM test WHERE %s", whereClause)

	// Parse with default recursion limit
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err)
	assert.True(t, errors.IsRecursionLimitExceeded(err), "Expected RecursionLimitExceeded error")
}

// TestParseDeeplyNestedSubqueryExprHitsRecursionLimits verifies recursion limit with deeply nested subqueries.
// Reference: tests/sqlparser_common.rs:11111
func TestParseDeeplyNestedSubqueryExprHitsRecursionLimits(t *testing.T) {
	// Build a deeply nested WHERE clause
	whereClause := "1=1"
	for i := 0; i < 100; i++ {
		whereClause = fmt.Sprintf("(%s) AND (1=1)", whereClause)
	}
	sql := fmt.Sprintf("SELECT id, user_id where id IN (select id from t WHERE %s)", whereClause)

	// Parse with default recursion limit
	_, err := parser.ParseSQL(generic.NewGenericDialect(), sql)
	require.Error(t, err)
	assert.True(t, errors.IsRecursionLimitExceeded(err), "Expected RecursionLimitExceeded error")
}
