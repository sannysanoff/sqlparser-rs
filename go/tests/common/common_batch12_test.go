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
// This file contains tests 201-220 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseTableFunction verifies TABLE() function in FROM clause.
// Reference: tests/sqlparser_common.rs:6933
func TestParseTableFunction(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic TABLE function with alias
	dialects.VerifiedOnlySelect(t, "SELECT * FROM TABLE(FUN('1')) AS a")
}

// TestParseSelectWithAliasAndColumnDefs verifies SELECT with table alias and column definitions.
// Reference: tests/sqlparser_common.rs:6966
func TestParseSelectWithAliasAndColumnDefs(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM jsonb_to_record('{}'::JSONB) AS x (a TEXT, b INT)"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseUnnest verifies UNNEST function parsing.
// Reference: tests/sqlparser_common.rs:7000
func TestParseUnnest(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT UNNEST(make_array(1, 2, 3))"
	dialects.VerifiedStmt(t, sql)

	sql2 := "SELECT UNNEST(make_array(1, 2, 3), make_array(4, 5))"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseUnnestInFromClause verifies UNNEST in FROM clause with various options.
// Reference: tests/sqlparser_common.rs:7008
func TestParseUnnestInFromClause(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			bigquery.NewBigQueryDialect(),
			generic.NewGenericDialect(),
		},
	}

	// Test various UNNEST configurations
	testCases := []string{
		"SELECT * FROM UNNEST(expr) AS numbers WITH OFFSET",
		"SELECT * FROM UNNEST(expr)",
		"SELECT * FROM UNNEST(expr) WITH OFFSET",
		"SELECT * FROM UNNEST(make_array(1, 2, 3))",
		"SELECT * FROM UNNEST(make_array(1, 2, 3), make_array(5, 6))",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedOnlySelect(t, sql)
		})
	}
}

// TestParseParens verifies parenthesized expression parsing.
// Reference: tests/sqlparser_common.rs:7166
func TestParseParens(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT (a + b) - (c + d)"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSearchedCaseExpr verifies searched CASE expression parsing.
// Reference: tests/sqlparser_common.rs:7189
func TestParseSearchedCaseExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT CASE WHEN bar IS NULL THEN 'null' WHEN bar = 0 THEN '=0' WHEN bar >= 0 THEN '>=0' ELSE '<0' END FROM foo"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseSimpleCaseExpr verifies simple CASE expression with operand parsing.
// Reference: tests/sqlparser_common.rs:7230
func TestParseSimpleCaseExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT CASE foo WHEN 1 THEN 'Y' ELSE 'N' END"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseFromAdvanced verifies complex FROM clause with table-valued function and hints.
// Reference: tests/sqlparser_common.rs:7253
func TestParseFromAdvanced(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM fn(1, 2) AS foo, schema.bar AS bar WITH (NOLOCK)"
	dialects.VerifiedStmt(t, sql)
}

// TestParseNullaryTableValuedFunction verifies nullary (no-argument) table-valued function.
// Reference: tests/sqlparser_common.rs:7259
func TestParseNullaryTableValuedFunction(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM fn()"
	dialects.VerifiedStmt(t, sql)
}

// TestParseImplicitJoin verifies implicit join (comma-separated tables).
// Reference: tests/sqlparser_common.rs:7265
func TestParseImplicitJoin(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM t1, t2"
	dialects.VerifiedOnlySelect(t, sql)

	sql2 := "SELECT * FROM t1a NATURAL JOIN t1b, t2a NATURAL JOIN t2b"
	dialects.VerifiedOnlySelect(t, sql2)
}

// TestParseCrossJoin verifies CROSS JOIN parsing.
// Reference: tests/sqlparser_common.rs:7308
func TestParseCrossJoin(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT * FROM t1 CROSS JOIN t2"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseCrossJoinConstraint verifies CROSS JOIN with ON and USING constraints.
// Reference: tests/sqlparser_common.rs:7322
func TestParseCrossJoinConstraint(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testCases := []string{
		"SELECT * FROM t1 CROSS JOIN t2 ON a = b",
		"SELECT * FROM t1 CROSS JOIN t2 USING(a)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedOnlySelect(t, sql)
		})
	}
}

// TestParseJoinsOn verifies various JOIN types with ON clause.
// Reference: tests/sqlparser_common.rs:7355
func TestParseJoinsOn(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testCases := []string{
		"SELECT * FROM t1 JOIN t2 AS foo ON c1 = c2",
		"SELECT * FROM t1 JOIN t2 foo ON c1 = c2",
		"SELECT * FROM t1 JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 INNER JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 LEFT JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 LEFT OUTER JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 RIGHT JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 RIGHT OUTER JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 SEMI JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 LEFT SEMI JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 RIGHT SEMI JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 ANTI JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 LEFT ANTI JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 RIGHT ANTI JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 FULL JOIN t2 ON c1 = c2",
		"SELECT * FROM t1 GLOBAL FULL JOIN t2 ON c1 = c2",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestParseJoinsUsing verifies various JOIN types with USING clause.
// Reference: tests/sqlparser_common.rs:7498
func TestParseJoinsUsing(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testCases := []string{
		"SELECT * FROM t1 JOIN t2 AS foo USING(c1)",
		"SELECT * FROM t1 JOIN t2 foo USING(c1)",
		"SELECT * FROM t1 JOIN t2 USING(c1)",
		"SELECT * FROM t1 INNER JOIN t2 USING(c1)",
		"SELECT * FROM t1 LEFT JOIN t2 USING(c1)",
		"SELECT * FROM t1 LEFT OUTER JOIN t2 USING(c1)",
		"SELECT * FROM t1 RIGHT JOIN t2 USING(c1)",
		"SELECT * FROM t1 RIGHT OUTER JOIN t2 USING(c1)",
		"SELECT * FROM t1 SEMI JOIN t2 USING(c1)",
		"SELECT * FROM t1 LEFT SEMI JOIN t2 USING(c1)",
		"SELECT * FROM t1 RIGHT SEMI JOIN t2 USING(c1)",
		"SELECT * FROM t1 ANTI JOIN t2 USING(c1)",
		"SELECT * FROM t1 LEFT ANTI JOIN t2 USING(c1)",
		"SELECT * FROM t1 RIGHT ANTI JOIN t2 USING(c1)",
		"SELECT * FROM t1 FULL JOIN t2 USING(c1)",
		"SELECT * FROM tbl1 AS t1 JOIN tbl2 AS t2 USING(t2.col1)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestParseNaturalJoin verifies NATURAL JOIN parsing.
// Reference: tests/sqlparser_common.rs:7597
func TestParseNaturalJoin(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testCases := []string{
		"SELECT * FROM t1 NATURAL JOIN t2",
		"SELECT * FROM t1 NATURAL INNER JOIN t2",
		"SELECT * FROM t1 NATURAL LEFT JOIN t2",
		"SELECT * FROM t1 NATURAL LEFT OUTER JOIN t2",
		"SELECT * FROM t1 NATURAL RIGHT JOIN t2",
		"SELECT * FROM t1 NATURAL RIGHT OUTER JOIN t2",
		"SELECT * FROM t1 NATURAL FULL JOIN t2",
		"SELECT * FROM t1 NATURAL JOIN t2 AS t3",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestParseComplexJoin verifies complex join with multiple tables and conditions.
// Reference: tests/sqlparser_common.rs:7673
func TestParseComplexJoin(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT c1, c2 FROM t1, t4 JOIN t2 ON t2.c = t1.c LEFT JOIN t3 USING(q, c) WHERE t4.c = t1.c"
	dialects.VerifiedStmt(t, sql)
}

// TestParseJoinNesting verifies nested join parsing with parentheses.
// Reference: tests/sqlparser_common.rs:7679
func TestParseJoinNesting(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testCases := []string{
		"SELECT * FROM a NATURAL JOIN (b NATURAL JOIN (c NATURAL JOIN d NATURAL JOIN e)) NATURAL JOIN (f NATURAL JOIN (g NATURAL JOIN h))",
		"SELECT * FROM (a NATURAL JOIN b) NATURAL JOIN c",
		"SELECT * FROM (((a NATURAL JOIN b)))",
		"SELECT * FROM (a NATURAL JOIN b) AS c",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}
}

// TestParseJoinSyntaxVariants verifies various JOIN syntax variations.
// Reference: tests/sqlparser_common.rs:7728
func TestParseJoinSyntaxVariants(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testCases := []string{
		"SELECT c1 FROM t1 JOIN t2 USING(c1)",
		"SELECT c1 FROM t1 INNER JOIN t2 USING(c1)",
		"SELECT c1 FROM t1 LEFT JOIN t2 USING(c1)",
		"SELECT c1 FROM t1 LEFT OUTER JOIN t2 USING(c1)",
		"SELECT c1 FROM t1 RIGHT JOIN t2 USING(c1)",
		"SELECT c1 FROM t1 RIGHT OUTER JOIN t2 USING(c1)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
		})
	}

	// FULL OUTER JOIN normalizes to FULL JOIN
	dialects.OneStatementParsesTo(t, "SELECT c1 FROM t1 FULL OUTER JOIN t2 USING(c1)", "SELECT c1 FROM t1 FULL JOIN t2 USING(c1)")
}

// TestParseCTEs verifies Common Table Expression (CTE) parsing.
// Reference: tests/sqlparser_common.rs:7749
func TestParseCTEs(t *testing.T) {
	dialects := utils.NewTestedDialects()

	cteSqls := []string{
		"SELECT 1 AS foo",
		"SELECT 2 AS bar",
	}

	// Top-level CTE
	with := "WITH a AS (" + cteSqls[0] + "), b AS (" + cteSqls[1] + ") SELECT foo + bar FROM a, b"
	dialects.VerifiedQuery(t, with)

	// CTE in subquery
	sql := "SELECT (" + with + ")"
	dialects.VerifiedStmt(t, sql)

	// CTE in derived table
	sql = "SELECT * FROM (" + with + ")"
	dialects.VerifiedStmt(t, sql)

	// CTE in CREATE VIEW
	sql = "CREATE VIEW v AS " + with
	dialects.VerifiedStmt(t, sql)

	// Nested CTE
	sql = "WITH outer_cte AS (" + with + ") SELECT * FROM outer_cte"
	dialects.VerifiedQuery(t, sql)
}

// TestParseCTERenamedColumns verifies CTE with renamed columns.
// Reference: tests/sqlparser_common.rs:7806
func TestParseCTERenamedColumns(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "WITH cte (col1, col2) AS (SELECT foo, bar FROM baz) SELECT * FROM cte"

	// Verify the SQL parses correctly
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Verify re-serialization works
	assert.Equal(t, sql, stmts[0].String())
}
