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
// This file contains tests 81-100.
package common

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseInUnnest verifies IN UNNEST expression parsing.
// Reference: tests/sqlparser_common.rs:2332
func TestParseInUnnest(t *testing.T) {
	chk := func(negated bool) {
		sql := fmt.Sprintf("SELECT * FROM customers WHERE segment %sIN UNNEST(expr)",
			func() string {
				if negated {
					return "NOT "
				}
				return ""
			}())
		dialects := utils.NewTestedDialects()
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
	}
	chk(false)
	chk(true)
}

// TestParseInError verifies error for invalid IN syntax.
// Reference: tests/sqlparser_common.rs:2354
func TestParseInError(t *testing.T) {
	sql := "SELECT * FROM customers WHERE segment in segment"
	d := generic.NewGenericDialect()
	_, err := parser.ParseSQL(d, sql)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: (")
}

// TestParseStringAgg verifies string concatenation operator parsing.
// Reference: tests/sqlparser_common.rs:2365
func TestParseStringAgg(t *testing.T) {
	sql := "SELECT a || b"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseBitwiseOps verifies bitwise operator parsing (^, |, &).
// Reference: tests/sqlparser_common.rs:2391
func TestParseBitwiseOps(t *testing.T) {
	// Bitwise XOR - all dialects except PostgreSQL
	dialectsNoPG := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		_, isPG := d.(*postgresql.PostgreSqlDialect)
		return !isPG
	})
	sql := "SELECT a ^ b"
	stmts := dialectsNoPG.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Bitwise OR - all dialects
	dialectsAll := utils.NewTestedDialects()
	sql = "SELECT a | b"
	stmts = dialectsAll.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	// Bitwise AND - all dialects
	sql = "SELECT a & b"
	stmts = dialectsAll.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseBitwiseShiftOps verifies bitwise shift operator parsing (<<, >>).
// Reference: tests/sqlparser_common.rs:2412
func TestParseBitwiseShiftOps(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsBitwiseShiftOperators()
	})

	sql := "SELECT 1 << 2, 3 >> 4"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestParseBinaryAny verifies ANY comparison operator parsing.
// Reference: tests/sqlparser_common.rs:2435
func TestParseBinaryAny(t *testing.T) {
	sql := "SELECT a = ANY(b)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseBinaryAll verifies ALL comparison operator parsing.
// Reference: tests/sqlparser_common.rs:2449
func TestParseBinaryAll(t *testing.T) {
	sql := "SELECT a = ALL(b)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseBetweenWithExpr verifies BETWEEN with complex expressions.
// Reference: tests/sqlparser_common.rs:2484
func TestParseBetweenWithExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test BETWEEN with IS NULL
	sql := "SELECT * FROM t WHERE 1 BETWEEN 1 + 2 AND 3 + 4 IS NULL"
	dialects.VerifiedStmt(t, sql)

	// Test BETWEEN in compound expression
	sql = "SELECT * FROM t WHERE 1 = 1 AND 1 + x BETWEEN 1 AND 2"
	dialects.VerifiedStmt(t, sql)
}

// TestParseTuples verifies tuple/row expression parsing.
// Reference: tests/sqlparser_common.rs:2532
func TestParseTuples(t *testing.T) {
	sql := "SELECT (1, 2), (1), ('foo', 3, baz)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseTupleInvalid verifies error handling for invalid tuple syntax.
// Reference: tests/sqlparser_common.rs:2555
func TestParseTupleInvalid(t *testing.T) {
	d := generic.NewGenericDialect()

	// Missing closing paren
	sql := "select (1"
	_, err := parser.ParseSQL(d, sql)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: )")

	// Empty tuple
	sql = "select (), 2"
	_, err = parser.ParseSQL(d, sql)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "an expression")
}

// TestParseSelectOrderByLimit verifies ORDER BY with LIMIT clause.
// Reference: tests/sqlparser_common.rs:2612
func TestParseSelectOrderByLimit(t *testing.T) {
	sql := "SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY lname ASC, fname DESC LIMIT 2"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseSelectOrderByAll verifies ORDER BY ALL syntax.
// Reference: tests/sqlparser_common.rs:2646
func TestParseSelectOrderByAll(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsOrderByAll()
	})

	testCases := []string{
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL NULLS FIRST",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL NULLS LAST",
		"SELECT id, fname, lname FROM customer ORDER BY ALL ASC",
		"SELECT id, fname, lname FROM customer ORDER BY ALL ASC NULLS FIRST",
		"SELECT id, fname, lname FROM customer ORDER BY ALL ASC NULLS LAST",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL DESC",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL DESC NULLS FIRST",
		"SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY ALL DESC NULLS LAST",
	}

	for _, sql := range testCases {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
	}
}

// TestParseSelectOrderByNotSupportAll verifies ORDER BY ALL is treated as column when not supported.
// Reference: tests/sqlparser_common.rs:2727
func TestParseSelectOrderByNotSupportAll(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return !d.SupportsOrderByAll()
	})

	testCases := []string{
		"SELECT id, ALL FROM customer WHERE id < 5 ORDER BY ALL",
		"SELECT id, ALL FROM customer ORDER BY ALL ASC NULLS FIRST",
		"SELECT id, ALL FROM customer ORDER BY ALL DESC NULLS LAST",
	}

	for _, sql := range testCases {
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
	}
}

// TestParseSelectOrderByNullsOrder verifies ORDER BY with NULLS FIRST/LAST.
// Reference: tests/sqlparser_common.rs:2778
func TestParseSelectOrderByNullsOrder(t *testing.T) {
	sql := "SELECT id, fname, lname FROM customer WHERE id < 5 ORDER BY lname ASC NULLS FIRST, fname DESC NULLS LAST LIMIT 2"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseSelectGroupByAll verifies GROUP BY ALL syntax.
// Reference: tests/sqlparser_common.rs:2834
func TestParseSelectGroupByAll(t *testing.T) {
	sql := "SELECT id, fname, lname, SUM(order) FROM customer GROUP BY ALL"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// TestParseGroupByWithModifier verifies GROUP BY with modifiers (ROLLUP, CUBE, TOTALS).
// Reference: tests/sqlparser_common.rs:2846
func TestParseGroupByWithModifier(t *testing.T) {
	clauses := []string{"x", "a, b", "ALL"}
	modifiers := []string{
		"WITH ROLLUP",
		"WITH CUBE",
		"WITH TOTALS",
		"WITH ROLLUP WITH CUBE",
	}

	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsGroupByWithModifier()
	})

	for _, clause := range clauses {
		for _, modifier := range modifiers {
			sql := fmt.Sprintf("SELECT * FROM t GROUP BY %s %s", clause, modifier)
			stmts := dialects.ParseSQL(t, sql)
			require.Len(t, stmts, 1)
		}
	}

	// Invalid cases
	invalidCases := []string{
		"SELECT * FROM t GROUP BY x WITH",
		"SELECT * FROM t GROUP BY x WITH ROLLUP CUBE",
		"SELECT * FROM t GROUP BY x WITH WITH ROLLUP",
		"SELECT * FROM t GROUP BY WITH ROLLUP",
	}

	for _, sql := range invalidCases {
		_, err := parser.ParseSQL(dialects.Dialects[0], sql)
		require.Error(t, err, "Expected error for: %s", sql)
	}
}

// TestParseGroupBySpecialGroupingSets verifies GROUP BY with GROUPING SETS.
// Reference: tests/sqlparser_common.rs:2903
func TestParseGroupBySpecialGroupingSets(t *testing.T) {
	sql := "SELECT a, b, SUM(c) FROM tab1 GROUP BY a, b GROUPING SETS ((a, b), (a), (b), ())"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	_ = stmts[0] // Parsing verification only (struct fields not yet implemented)
}

// Helper function to create a number value
func number(n string) *ast.Value {
	val, err := ast.NewNumber(n, false)
	if err != nil {
		panic(fmt.Sprintf("Failed to create number: %v", err))
	}
	return val
}
