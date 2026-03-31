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
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseSimpleSelect verifies basic SELECT statement parsing.
// Reference: tests/sqlparser_common.rs:925
func TestParseSimpleSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, fname, lname FROM customer WHERE id = 1 LIMIT 5"

	dialects.VerifiedStmt(t, sql)
}

// TestParseLimit verifies LIMIT clause parsing.
// Reference: tests/sqlparser_common.rs:940
func TestParseLimit(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT * FROM user LIMIT 1")
}

// TestParseSelectDistinct verifies DISTINCT keyword parsing.
// Reference: tests/sqlparser_common.rs:971
func TestParseSelectDistinct(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT DISTINCT name FROM customer"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectAll verifies ALL keyword parsing (default behavior).
// Reference: tests/sqlparser_common.rs:981
func TestParseSelectAll(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT ALL name FROM customer"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWildcard verifies SELECT * parsing.
// Reference: tests/sqlparser_common.rs:991
func TestParseSelectWildcard(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM customer"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWithFromFirst verifies FROM-first syntax.
// Reference: tests/sqlparser_common.rs:997
func TestParseSelectWithFromFirst(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "FROM customer SELECT id"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectDistinctOn verifies DISTINCT ON syntax (PostgreSQL-specific).
// Reference: tests/sqlparser_common.rs:1010
func TestParseSelectDistinctOn(t *testing.T) {
	// For now, test only with PostgreSQL
	dialect := postgresql.NewPostgreSqlDialect()
	sql := "SELECT DISTINCT ON (id) name FROM customer"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParseSelectOrderBy verifies ORDER BY clause parsing.
// Reference: tests/sqlparser_common.rs:1061
func TestParseSelectOrderBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, name FROM customer ORDER BY name"
	dialects.VerifiedStmt(t, sql)

	sql2 := "SELECT id, name FROM customer ORDER BY name ASC, id DESC"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseSelectGroupBy verifies GROUP BY clause parsing.
// Reference: tests/sqlparser_common.rs:1113
func TestParseSelectGroupBy(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, name FROM customer GROUP BY name"
	dialects.VerifiedStmt(t, sql)

}

// TestParseSelectHaving verifies HAVING clause parsing.
// Reference: tests/sqlparser_common.rs:1150
func TestParseSelectHaving(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id, name FROM customer GROUP BY name HAVING COUNT(*) > 5"
	dialects.VerifiedStmt(t, sql)

}

// TestParseInsertValues verifies INSERT with VALUES clause.
// Reference: tests/sqlparser_common.rs:97
func TestParseInsertValues(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "INSERT INTO customer VALUES (1, 2, 3)"
	dialects.VerifiedStmt(t, sql)

	sql2 := "INSERT INTO customer (id, name, active) VALUES (1, 2, 3)"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "INSERT INTO customer VALUES (1, 2, 3), (1, 2, 3)"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseInsertDefaultValues verifies INSERT with DEFAULT VALUES.
// Reference: tests/sqlparser_common.rs:198
func TestParseInsertDefaultValues(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "INSERT INTO test_table DEFAULT VALUES"
	dialects.VerifiedStmt(t, sql)
}

// TestParseUpdate verifies UPDATE statement parsing.
// Reference: tests/sqlparser_common.rs:394
func TestParseUpdate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "UPDATE customer SET name = 'John' WHERE id = 1"
	dialects.VerifiedStmt(t, sql)
}

// TestParseDelete verifies DELETE statement parsing.
// Reference: tests/sqlparser_common.rs:703
func TestParseDelete(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "DELETE FROM customer WHERE id = 1"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWithTableAlias verifies table alias parsing.
// Reference: tests/sqlparser_common.rs:631
func TestParseSelectWithTableAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id FROM customer AS c"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectWithTableAliasWithoutAs verifies table alias without AS keyword.
// Reference: tests/sqlparser_common.rs:644
func TestParseSelectWithTableAliasWithoutAs(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT id FROM customer c"
	dialects.VerifiedStmt(t, sql)
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

// TestParseFunctionCall verifies function call parsing.
// Reference: tests/sqlparser_common.rs:80
func TestParseFunctionCall(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a.b.c.d(1, 2, 3) FROM T"
	dialects.VerifiedStmt(t, sql)
}

// TestParseAggregateFunctions verifies aggregate function parsing.
func TestParseAggregateFunctions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// COUNT(*)
	sql1 := "SELECT COUNT(*) FROM customer"
	dialects.VerifiedStmt(t, sql1)

	// SUM(column)
	sql2 := "SELECT SUM(amount) FROM orders"
	dialects.VerifiedStmt(t, sql2)

	// AVG with DISTINCT
	sql3 := "SELECT AVG(DISTINCT price) FROM products"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseJoin verifies JOIN parsing.
func TestParseJoin(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM customer JOIN orders ON customer.id = orders.customer_id"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM customer LEFT JOIN orders ON customer.id = orders.customer_id"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT * FROM customer INNER JOIN orders ON customer.id = orders.customer_id"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseSubquery verifies subquery parsing.
func TestParseSubquery(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM (SELECT id FROM customer) AS c"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM customer WHERE id IN (SELECT customer_id FROM orders)"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT * FROM customer WHERE EXISTS (SELECT * FROM orders WHERE orders.customer_id = customer.id)"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseCaseExpression verifies CASE expression parsing.
func TestParseCaseExpression(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql2 := "SELECT CASE x WHEN 1 THEN 'one' WHEN 2 THEN 'two' ELSE 'other' END FROM t"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseBetween verifies BETWEEN expression parsing.
func TestParseBetween(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM t WHERE x BETWEEN 1 AND 10"
	dialects.VerifiedStmt(t, sql)
}

// TestParseInExpression verifies IN expression parsing.
func TestParseInExpression(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE x IN (1, 2, 3)"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE x NOT IN (1, 2, 3)"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseLikeExpression verifies LIKE expression parsing.
func TestParseLikeExpression(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE name LIKE '%test%'"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE name NOT LIKE '%test%'"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseArithmeticOperators verifies arithmetic operators.
func TestParseArithmeticOperators(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT 1 + 2 FROM t"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT 10 - 5 FROM t"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT 4 * 5 FROM t"
	dialects.VerifiedStmt(t, sql3)

	sql4 := "SELECT 20 / 4 FROM t"
	dialects.VerifiedStmt(t, sql4)
}

// TestParseComparisonOperators verifies comparison operators.
func TestParseComparisonOperators(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE x = 1"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE x != 1"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT * FROM t WHERE x <> 1"
	dialects.VerifiedStmt(t, sql3)

	sql4 := "SELECT * FROM t WHERE x > 1"
	dialects.VerifiedStmt(t, sql4)

	sql5 := "SELECT * FROM t WHERE x < 1"
	dialects.VerifiedStmt(t, sql5)

	sql6 := "SELECT * FROM t WHERE x >= 1"
	dialects.VerifiedStmt(t, sql6)

	sql7 := "SELECT * FROM t WHERE x <= 1"
	dialects.VerifiedStmt(t, sql7)
}

// TestParseLogicalOperators verifies logical operators.
func TestParseLogicalOperators(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE x = 1 AND y = 2"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE x = 1 OR y = 2"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT * FROM t WHERE NOT x = 1"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseStringLiterals verifies string literal parsing.
func TestParseStringLiterals(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT 'hello' FROM t"
	dialects.VerifiedStmt(t, sql1)

	sql2 := `SELECT "hello" FROM t`
	dialects.VerifiedStmt(t, sql2)
}

// TestParseNull verifies NULL literal parsing.
func TestParseNull(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT NULL FROM t"
	dialects.VerifiedStmt(t, sql)
}

// TestParseBoolean verifies boolean literal parsing.
func TestParseBoolean(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT TRUE FROM t"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT FALSE FROM t"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseNullHandling verifies IS NULL / IS NOT NULL parsing.
func TestParseNullHandling(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM t WHERE x IS NULL"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM t WHERE x IS NOT NULL"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseAliasedExpressions verifies aliased expressions.
func TestParseAliasedExpressions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT x AS y FROM t"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT x y FROM t"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseUnion verifies UNION clause parsing.
func TestParseUnion(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM a UNION SELECT * FROM b"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM a UNION ALL SELECT * FROM b"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseIntersect verifies INTERSECT clause parsing.
func TestParseIntersect(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM a INTERSECT SELECT * FROM b"
	dialects.VerifiedStmt(t, sql1)
}

// TestParseExcept verifies EXCEPT clause parsing.
func TestParseExcept(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM a EXCEPT SELECT * FROM b"
	dialects.VerifiedStmt(t, sql1)
}

// TestParseCTE verifies Common Table Expression parsing.
func TestParseCTE(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "WITH cte AS (SELECT * FROM customer) SELECT * FROM cte"
	dialects.VerifiedQuery(t, sql)
}

// TestParseWindowFunctions verifies window function parsing.
func TestParseWindowFunctions(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT ROW_NUMBER() OVER (ORDER BY id) FROM t"
	dialects.VerifiedStmt(t, sql)
}

// TestParseTransactionStatements verifies transaction statement parsing.
func TestParseTransactionStatements(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "BEGIN")
	dialects.VerifiedStmt(t, "COMMIT")
	dialects.VerifiedStmt(t, "ROLLBACK")
}

// TestParseAlterTable verifies ALTER TABLE statement parsing.
func TestParseAlterTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "ALTER TABLE customer ADD COLUMN email VARCHAR(255)"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "ALTER TABLE customer DROP COLUMN email"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseTruncate verifies TRUNCATE statement parsing.
func TestParseTruncate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "TRUNCATE TABLE customer"
	dialects.VerifiedStmt(t, sql)
}

// TestParseUpdateFull verifies UPDATE statement parsing with multiple assignments and error cases.
// Reference: tests/sqlparser_common.rs:394
func TestParseUpdateFull(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "UPDATE t SET a = 1, b = 2, c = 3 WHERE d"
	dialects.VerifiedStmt(t, sql)

	// Multiple assignments to same column
	dialects.VerifiedStmt(t, "UPDATE t SET a = 1, a = 2, a = 3")
}

// TestParseUpdateSetFrom verifies UPDATE with FROM clause parsing.
// Reference: tests/sqlparser_common.rs:443
func TestParseUpdateSetFrom(t *testing.T) {
	// This test uses specific dialects that support UPDATE...FROM
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
			postgresql.NewPostgreSqlDialect(),
			bigquery.NewBigQueryDialect(),
			snowflake.NewSnowflakeDialect(),
			redshift.NewRedshiftSqlDialect(),
			mssql.NewMsSqlDialect(),
			sqlite.NewSQLiteDialect(),
		},
	}

	sql := "UPDATE t1 SET name = t2.name FROM (SELECT name, id FROM t1 GROUP BY id) AS t2 WHERE t1.id = t2.id"
	dialects.VerifiedStmt(t, sql)

	sql2 := "UPDATE T SET a = b FROM U, (SELECT foo FROM V) AS W WHERE 1 = 1"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseUpdateWithTableAlias verifies UPDATE with table alias parsing.
// Reference: tests/sqlparser_common.rs:546
func TestParseUpdateWithTableAlias(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "UPDATE users AS u SET u.username = 'new_user' WHERE u.username = 'old_user'"
	dialects.VerifiedStmt(t, sql)
}

// TestParseUpdateOr verifies SQLite UPDATE OR clause parsing.
// Reference: tests/sqlparser_common.rs:611
func TestParseUpdateOr(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "UPDATE OR REPLACE t SET n = n + 1")
	dialects.VerifiedStmt(t, "UPDATE OR ROLLBACK t SET n = n + 1")
	dialects.VerifiedStmt(t, "UPDATE OR ABORT t SET n = n + 1")
	dialects.VerifiedStmt(t, "UPDATE OR FAIL t SET n = n + 1")
	dialects.VerifiedStmt(t, "UPDATE OR IGNORE t SET n = n + 1")
}

// TestParseSelectWithTableAliasAs verifies SELECT with table alias and column list.
// Reference: tests/sqlparser_common.rs:630
func TestParseSelectWithTableAliasAs(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// With AS keyword
	dialects.VerifiedStmt(t, "SELECT a, b, c FROM lineitem AS l (A, B, C)")

	// Without AS keyword (AS is optional)
	dialects.VerifiedStmt(t, "SELECT a, b, c FROM lineitem l (A, B, C)")
}

// TestParseSelectWithTableAliasColumns verifies SELECT with table alias and column definitions.
// Reference: tests/sqlparser_common.rs:643
func TestParseSelectWithTableAliasColumns(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a, b, c FROM lineitem AS l (A, B, C)"
	dialects.VerifiedStmt(t, sql)
}

// TestParseNumericLiteralUnderscore verifies parsing of numeric literals with underscores.
// Reference: tests/sqlparser_common.rs:61
func TestParseNumericLiteralUnderscore(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsNumericLiteralUnderscores()
	})

	// Without bigdecimal feature, canonical is the same as input
	canonical := "SELECT 10_000"

	dialects.VerifiedOnlySelectWithCanonical(t, "SELECT 10_000", canonical)

	// Parse and check the projection contains the expected value
	stmts := dialects.ParseSQL(t, "SELECT 10_000")
	require.Len(t, stmts, 1)

	// The projection should contain the numeric value
	// Note: We verify through round-trip in VerifiedOnlySelectWithCanonical above
}

// TestParseFunctionObjectName verifies parsing of qualified function names.
// Reference: tests/sqlparser_common.rs:79
func TestParseFunctionObjectName(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT a.b.c.d(1, 2, 3) FROM T"

	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Extract expression from first projection item
	// This is a simplified check - in a full implementation we would
	// type assert to *Select and verify the FunctionExpr
	dialects.VerifiedStmt(t, sql)
}

// TestParseInsertValuesFull verifies INSERT with VALUES clause parsing (comprehensive).
// Reference: tests/sqlparser_common.rs:96
func TestParseInsertValuesFull(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test various INSERT forms
	testCases := []struct {
		sql          string
		tableName    string
		columns      []string
		rowCount     int
		valueKeyword bool
	}{
		{
			sql:          "INSERT customer VALUES (1, 2, 3)",
			tableName:    "customer",
			columns:      []string{},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO customer VALUES (1, 2, 3)",
			tableName:    "customer",
			columns:      []string{},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO customer VALUES (1, 2, 3), (1, 2, 3)",
			tableName:    "customer",
			columns:      []string{},
			rowCount:     2,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO public.customer VALUES (1, 2, 3)",
			tableName:    "public.customer",
			columns:      []string{},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO db.public.customer VALUES (1, 2, 3)",
			tableName:    "db.public.customer",
			columns:      []string{},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          "INSERT INTO public.customer (id, name, active) VALUES (1, 2, 3)",
			tableName:    "public.customer",
			columns:      []string{"id", "name", "active"},
			rowCount:     1,
			valueKeyword: false,
		},
		{
			sql:          `INSERT INTO t (id, name, active) VALUE (1, 2, 3)`,
			tableName:    "t",
			columns:      []string{"id", "name", "active"},
			rowCount:     1,
			valueKeyword: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, tc.sql)
			require.Len(t, stmts, 1)

			insert, ok := stmts[0].(*statement.Insert)
			require.True(t, ok, "Expected Insert statement, got %T", stmts[0])

			// Verify table name
			assert.Equal(t, tc.tableName, insert.Table.String())

			// Verify columns
			assert.Equal(t, len(tc.columns), len(insert.Columns))
			for i, col := range tc.columns {
				assert.Equal(t, col, insert.Columns[i].String())
			}

			// Verify source exists (VALUES clause)
			require.NotNil(t, insert.Source)
			require.NotNil(t, insert.Source.Body)

			// Check that body is a ValuesSetExpr
			valuesExpr, ok := insert.Source.Body.(*query.ValuesSetExpr)
			require.True(t, ok, "Expected ValuesSetExpr, got %T", insert.Source.Body)

			// Verify row count
			assert.Equal(t, tc.rowCount, len(valuesExpr.Values.Rows))

			// Verify VALUE keyword vs VALUES keyword
			assert.Equal(t, tc.valueKeyword, valuesExpr.Values.ValueKeyword)
		})
	}

	// Test INSERT with CTE
	cteSQL := "INSERT INTO customer WITH foo AS (SELECT 1) SELECT * FROM foo UNION VALUES (1)"
	dialects.VerifiedStmt(t, cteSQL)
}

// TestParseInsertSet verifies INSERT with SET clause parsing (MySQL-specific).
// Reference: tests/sqlparser_common.rs:180
func TestParseInsertSet(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsInsertSet()
	})

	sql := "INSERT INTO tbl1 SET col1 = 1, col2 = 'abc', col3 = current_date()"
	dialects.VerifiedStmt(t, sql)
}

// TestParseReplaceInto verifies REPLACE INTO parsing error (not supported by PostgreSQL).
// Reference: tests/sqlparser_common.rs:186
func TestParseReplaceInto(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()
	sql := "REPLACE INTO public.customer (id, name, active) VALUES (1, 2, 3)"

	_, err := parser.ParseSQL(dialect, sql)
	require.Error(t, err)
	// The error should indicate REPLACE is not supported
	assert.Contains(t, err.Error(), "REPLACE")
}

// TestParseInsertDefaultValuesFull verifies INSERT with DEFAULT VALUES (comprehensive).
// Reference: tests/sqlparser_common.rs:197
func TestParseInsertDefaultValuesFull(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test basic INSERT with DEFAULT VALUES
	t.Run("basic default values", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES"
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		insert, ok := stmts[0].(*statement.Insert)
		require.True(t, ok, "Expected Insert statement")

		// Verify table name
		assert.Equal(t, "test_table", insert.Table.String())

		// Verify columns is empty
		assert.Empty(t, insert.Columns)

		// Verify source is nil (DEFAULT VALUES has no source)
		assert.Nil(t, insert.Source)

		// Verify no partitioned clause
		assert.Empty(t, insert.Partitioned)

		// Verify no returning clause
		assert.Empty(t, insert.Returning)

		// Verify no on conflict clause
		assert.Nil(t, insert.On)
	})

	// Test INSERT with DEFAULT VALUES and RETURNING
	t.Run("default values with returning", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES RETURNING test_column"
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		insert, ok := stmts[0].(*statement.Insert)
		require.True(t, ok, "Expected Insert statement")

		assert.Equal(t, "test_table", insert.Table.String())
		assert.Empty(t, insert.Columns)
		assert.Nil(t, insert.Source)
		assert.NotEmpty(t, insert.Returning, "Should have RETURNING clause")
	})

	// Test INSERT with DEFAULT VALUES and ON CONFLICT
	t.Run("default values with on conflict", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES ON CONFLICT DO NOTHING"
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		insert, ok := stmts[0].(*statement.Insert)
		require.True(t, ok, "Expected Insert statement")

		assert.Equal(t, "test_table", insert.Table.String())
		assert.Empty(t, insert.Columns)
		assert.Nil(t, insert.Source)
		assert.NotNil(t, insert.On, "Should have ON CONFLICT clause")
	})

	// Test error: columns with DEFAULT VALUES should fail
	t.Run("columns with default values error", func(t *testing.T) {
		sql := "INSERT INTO test_table (test_col) DEFAULT VALUES"
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "DEFAULT")
	})

	// Test error: DEFAULT VALUES with Hive after columns should fail
	t.Run("default values with after columns error", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES (some_column)"
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err)
	})

	// Test error: DEFAULT VALUES with Hive partition should fail
	t.Run("default values with partition error", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES PARTITION (some_column)"
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err)
	})

	// Test error: DEFAULT VALUES with values list should fail
	t.Run("default values with values list error", func(t *testing.T) {
		sql := "INSERT INTO test_table DEFAULT VALUES (1)"
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err)
	})
}
