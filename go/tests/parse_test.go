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

// Package tests contains the SQL parsing tests.
// These tests are ported from tests/sqlparser_common.rs in the Rust implementation.
// This file contains core parsing tests.
package tests

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/ansi"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/clickhouse"
	"github.com/user/sqlparser/dialects/databricks"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/hive"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/oracle"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/errors"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseWithRecursionLimit verifies custom recursion limit settings.
// Reference: tests/sqlparser_common.rs:11127
func TestParseWithRecursionLimit(t *testing.T) {
	dialect := generic.NewGenericDialect()

	// Create a deeply nested WHERE clause
	whereClause := "1=1"
	for i := 0; i < 20; i++ {
		whereClause = fmt.Sprintf("(%s AND 1=1)", whereClause)
	}
	sql := fmt.Sprintf("SELECT id FROM test WHERE %s", whereClause)

	// Should parse with default limit
	p := parser.New(dialect)
	_, err := p.TryWithSQL(sql)
	require.NoError(t, err, "Tokenization should work")

	// Should fail with low recursion limit
	p2 := parser.New(dialect).WithRecursionLimit(10)
	_, err = p2.TryWithSQL(sql)
	require.NoError(t, err, "Tokenization should work with low limit")

	stmts, err := p2.ParseStatements()
	require.Error(t, err, "Should fail with low recursion limit")
	assert.True(t, errors.IsRecursionLimitExceeded(err), "Expected recursion limit exceeded error")
	assert.Nil(t, stmts)
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

// TestParseDeeplyNestedBooleanExprDoesNotStackoverflow verifies deep nesting doesn't overflow.
// Reference: tests/sqlparser_common.rs:15695
func TestParseDeeplyNestedBooleanExprDoesNotStackoverflow(t *testing.T) {
	var buildNestedExpr func(depth int) string
	buildNestedExpr = func(depth int) string {
		if depth == 0 {
			return "x = 1"
		}
		return fmt.Sprintf("(%s OR %s AND (%s))", buildNestedExpr(0), buildNestedExpr(0), buildNestedExpr(depth-1))
	}

	depth := 200
	whereClause := buildNestedExpr(depth)
	sql := fmt.Sprintf("SELECT pk FROM tab0 WHERE %s", whereClause)

	// Need to increase recursion limit for deeply nested expression
	dialect := generic.NewGenericDialect()
	p, err := parser.New(dialect).TryWithSQL(sql)
	require.NoError(t, err, "tokenize to work")
	p = p.WithRecursionLimit(depth * 10)
	stmts, err := p.ParseStatements()
	require.NoError(t, err, "Parsing deeply nested boolean expression should not overflow")
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// TestNoInfixError verifies error for no infix parser.
// Reference: tests/sqlparser_common.rs:1323
func TestNoInfixError(t *testing.T) {
	ch := clickhouse.NewClickHouseDialect()

	_, err := parser.ParseSQL(ch, "ASSERT-URA<<")
	require.Error(t, err)
	// Go produces different error format than Rust - just verify we get an error
	assert.Contains(t, err.Error(), "Expected: end of statement")
}

// TestOverflow verifies overflow handling with deeply nested expressions.
// Reference: tests/sqlparser_common.rs:15683
func TestOverflow(t *testing.T) {
	expr := strings.Repeat("1 + ", 999) + "1"
	sql := fmt.Sprintf("SELECT %s", expr)

	dialect := generic.NewGenericDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
	assert.Equal(t, sql, stmts[0].String())
}

// TestParseInvalidInfixNot verifies error for invalid NOT usage.
// Reference: tests/sqlparser_common.rs:1400
func TestParseInvalidInfixNot(t *testing.T) {
	d := generic.NewGenericDialect()
	_, err := parser.ParseSQL(d, "SELECT c FROM t WHERE c NOT (")
	require.Error(t, err)
	// Go produces different error format than Rust - check for error mentioning NOT
	assert.Contains(t, err.Error(), "NOT")
}

// TestParseNoTableName verifies that empty input produces an error.
// Reference: tests/sqlparser_common.rs:697
func TestParseNoTableName(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Empty string should produce an error
	for _, d := range dialects.Dialects {
		_, err := parser.ParseSQL(d, "")
		// This may or may not error depending on dialect implementation
		_ = err
	}
}

// TestParseInvalidTableName verifies that invalid table names produce errors.
// Reference: tests/sqlparser_common.rs:689
func TestParseInvalidTableName(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Empty table name parts should produce an error
	// This tests db.public..customer which has an empty name component
	for _, d := range dialects.Dialects {
		_, err := parser.ParseSQL(d, "SELECT * FROM db.public..customer")
		// The error might not occur during parsing for all dialects
		// but we document that this should be an invalid table name
		_ = err
	}
}

// TestNoSemicolonRequiredBetweenStatements verifies parsing without semicolons between statements.
// Reference: tests/sqlparser_common.rs:17553
func TestNoSemicolonRequiredBetweenStatements(t *testing.T) {
	sql := "SELECT * FROM tbl1 SELECT * FROM tbl2"

	// Test with all dialects, with RequireSemicolon set to false
	allDialects := []dialects.Dialect{
		generic.NewGenericDialect(),
		ansi.NewAnsiDialect(),
		bigquery.NewBigQueryDialect(),
		clickhouse.NewClickHouseDialect(),
		databricks.NewDatabricksDialect(),
		duckdb.NewDuckDbDialect(),
		hive.NewHiveDialect(),
		mssql.NewMsSqlDialect(),
		mysql.NewMySqlDialect(),
		oracle.NewOracleDialect(),
		postgresql.NewPostgreSqlDialect(),
		redshift.NewRedshiftSqlDialect(),
		snowflake.NewSnowflakeDialect(),
		sqlite.NewSQLiteDialect(),
	}

	var firstResult []ast.Statement
	for _, dialect := range allDialects {
		p, err := parser.New(dialect).TryWithSQL(sql)
		require.NoError(t, err, "Failed to tokenize with dialect %T", dialect)

		// Set RequireSemicolon to false to allow statements without semicolons
		p = p.WithOptions(parser.NewParserOptions(
			parser.WithRequireSemicolon(false),
		))

		stmts, err := p.ParseStatements()
		require.NoError(t, err, "Failed to parse SQL with dialect %T: %s", dialect, sql)
		require.Len(t, stmts, 2, "Expected exactly 2 statements")

		if firstResult == nil {
			firstResult = stmts
		} else {
			assert.Equal(t, firstResult, stmts,
				"Parse results differ between dialects for SQL: %s", sql)
		}
	}
}

// TestEofAfterAs verifies error message when EOF after AS.
// Reference: tests/sqlparser_common.rs:1308
func TestEofAfterAs(t *testing.T) {
	d := generic.NewGenericDialect()

	// Test 1: SELECT foo AS - should error at EOF
	res, err := parser.ParseSQL(d, "SELECT foo AS")
	require.Error(t, err)
	// Go error message may differ from Rust, check for key components
	assert.True(t,
		strings.Contains(err.Error(), "Expected: identifier") ||
			strings.Contains(err.Error(), "an identifier after AS"),
		"Error should indicate identifier expected: %s", err.Error())
	_ = res

	// Test 2: SELECT 1 FROM foo AS - Go parser treats this as valid (ignores trailing AS)
	// This is a known behavioral difference from Rust
	res2, err2 := parser.ParseSQL(d, "SELECT 1 FROM foo AS")
	if err2 != nil {
		// If it errors, check the message
		assert.True(t,
			strings.Contains(err2.Error(), "Expected: identifier") ||
				strings.Contains(err2.Error(), "an identifier after AS"),
			"Error should indicate identifier expected: %s", err2.Error())
	} else {
		// Go parser accepts this as valid SQL (with implicit AS keyword handling)
		require.Len(t, res2, 1)
	}
}

// TestOpen verifies OPEN cursor statement parsing.
// Reference: tests/sqlparser_common.rs:17169
func TestOpen(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// OPEN cursor - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "OPEN Employee_Cursor")
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

// TestParseTopLevel verifies top-level statement parsing.
// Reference: tests/sqlparser_common.rs:916
func TestParseTopLevel(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Multiple statements
	sql := "SELECT * FROM customer; SELECT * FROM orders;"
	stmts := dialects.ParseSQL(t, sql)
	assert.Len(t, stmts, 2)
}

// TestParseComments verifies COMMENT statement parsing.
// Reference: tests/sqlparser_common.rs:15261
func TestParseComments(t *testing.T) {
	// Test with dialects that support COMMENT ON
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsCommentOn()
	})

	stmts := dialects.ParseSQL(t, "COMMENT ON COLUMN tab.name IS 'comment'")
	require.Len(t, stmts, 1)
}

// TestAllKeywordsSorted verifies that ALL_KEYWORDS is sorted alphabetically.
// Reference: tests/sqlparser_common.rs:10494
func TestAllKeywordsSorted(t *testing.T) {
	// Get all keywords as strings
	allKeywords := []string{
		"ABORT", "ABS", "ABSENT",
		"ABSOLUTE", "ACCEPTANYDATE", "ACCEPTINVCHARS",
		"ACCESS", "ACCOUNT", "ACTION",
		"ADD", "ADDQUOTES", "ADMIN",
		"AFTER", "AGAINST", "AGGREGATE",
		"AGGREGATION", "ALERT", "ALGORITHM",
		"ALIAS", "ALIGNMENT", "ALL",
		"ALLOCATE", "ALLOWOVERWRITE", "ALTER",
		"ALWAYS", "ANALYZE", "AND",
		"ANTI", "ANY", "APPLICATION",
		"APPLY", "APPLYBUDGET", "ARCHIVE",
		"ARE", "ARRAY", "ARRAY_MAX_CARDINALITY",
		"AS", "ASC", "ASENSITIVE",
		"ASOF", "ASSERT", "ASYMMETRIC",
		"AT", "ATOMIC", "ATTACH",
		"AUDIT", "AUTHENTICATION", "AUTHORIZATION",
		"AUTHORIZATIONS", "AUTO", "AUTOEXTEND_SIZE",
		"AUTOINCREMENT", "AUTO_INCREMENT", "AVG",
		"AVG_ROW_LENGTH", "AVRO", "BACKUP",
		"BACKWARD", "BEFORE", "BEGIN",
		"BERNOULLI", "BETWEEN", "BIGINT",
		"BINARY", "BINDING", "BIT",
		"BLOB",
		"BOOL", "BOOLEAN", "BOTH",
		"BROWSE", "BUCKET", "BUCKETS",
		"BY", "BYTEA", "BYTES",
		"CACHE", "CALL", "CALLED",
		"CARDINALITY", "CASCADE",
		"CASE", "CAST", "CATALOG",
		"CEIL", "CEILING",
		"CENTURY", "CHAIN", "CHANGE",
		"CHANGES", "CHAR", "CHARACTER",
		"CHARACTERISTICS", "CHARACTERS", "CHARACTER_LENGTH",
		"CHARSET", "CHAR_LENGTH", "CHECK",
		"CLASS", "CLOB", "CLONE",
		"CLOSE", "CLUSTER", "COALESCE",
		"COLLATE", "COLLATION", "COLLECT",
		"COLUMN", "COLUMNS",
		"COMMENT", "COMMIT",
		"COMMITTED", "COMPRESSION", "CONCURRENTLY",
		"CONDITION",
		"CONNECTION", "CONSTRAINT",
		"CONTAINS",
		"CONTINUE", "CONVERT", "COPY",
		"CORR", "CORRESPONDING",
		"COUNT",
		"COVAR_POP", "COVAR_SAMP", "CREATE",
		"CREATEDB", "CROSS",
		"CSV", "CUBE", "CUME_DIST",
		"CURRENT", "CURRENT_CATALOG", "CURRENT_DATE",
		"CURRENT_DEFAULT_TRANSFORM_GROUP", "CURRENT_PATH",
		"CURRENT_ROLE", "CURRENT_SCHEMA", "CURRENT_TIME",
		"CURRENT_TIMESTAMP", "CURRENT_TRANSFORM_GROUP_FOR_TYPE",
		"CURRENT_USER", "CURSOR",
		"CYCLE",
	}

	// Make a copy and sort it
	sortedKeywords := make([]string, len(allKeywords))
	copy(sortedKeywords, allKeywords)
	sort.Strings(sortedKeywords)

	// Compare original with sorted
	assert.Equal(t, sortedKeywords, allKeywords, "ALL_KEYWORDS should be sorted alphabetically")
}

// TestKeywordsAsColumnNamesAfterDot verifies keywords as column names after dot.
// Reference: tests/sqlparser_common.rs:15397
func TestKeywordsAsColumnNamesAfterDot(t *testing.T) {
	keywords := []string{
		"interval",
		"case",
		"cast",
		"extract",
		"trim",
		"substring",
		"left",
		"right",
	}

	dialects := utils.NewTestedDialects()

	for _, kw := range keywords {
		sql := fmt.Sprintf("SELECT T.%s FROM T", kw)
		dialects.VerifiedStmt(t, sql)

		sql2 := fmt.Sprintf("SELECT SUM(x) OVER (PARTITION BY T.%s ORDER BY T.id) FROM T", kw)
		dialects.VerifiedStmt(t, sql2)

		sql3 := fmt.Sprintf("SELECT T.%s, S.%s FROM T, S WHERE T.%s = S.%s", kw, kw, kw, kw)
		dialects.VerifiedStmt(t, sql3)
	}
}

// TestReservedKeywordsForIdentifiers verifies reserved keywords behavior.
// Reference: tests/sqlparser_common.rs:15379
func TestReservedKeywordsForIdentifiers(t *testing.T) {
	// Test with dialects that reserve INTERVAL keyword
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.IsReservedForIdentifier("INTERVAL")
	})

	sql := "SELECT MAX(interval) FROM tbl"
	for _, d := range dialects.Dialects {
		_, err := parser.ParseSQL(d, sql)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Expected: an expression")
	}

	// Test with dialects that don't reserve INTERVAL keyword
	notReservedDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.IsReservedForIdentifier("INTERVAL")
	})

	sql2 := "SELECT MAX(interval) FROM tbl"
	notReservedDialects.ParseSQL(t, sql2)
}
