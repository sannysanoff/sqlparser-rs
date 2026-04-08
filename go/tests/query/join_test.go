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

// Package query contains query-related SQL parsing tests.
// These tests are ported from tests/sqlparser_common.rs in the Rust implementation.
package query

import (
	"testing"

	"github.com/user/sqlparser/tests/utils"
)

// TestParseJoin verifies JOIN parsing.
// Reference: tests/sqlparser_common.rs:221
func TestParseJoin(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM customer JOIN orders ON customer.id = orders.customer_id"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM customer LEFT JOIN orders ON customer.id = orders.customer_id"
	dialects.VerifiedStmt(t, sql2)

	sql3 := "SELECT * FROM customer INNER JOIN orders ON customer.id = orders.customer_id"
	dialects.VerifiedStmt(t, sql3)
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
		"SELECT c1 FROM t1 FULL JOIN t2 USING(c1)",
		"SELECT c1 FROM t1 FULL OUTER JOIN t2 USING(c1)",
	}

	for _, sql := range testCases {
		t.Run(sql, func(t *testing.T) {
			dialects.VerifiedStmt(t, sql)
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

// TestParseImplicitJoin verifies implicit join (comma-separated tables).
// Reference: tests/sqlparser_common.rs:7265
func TestParseImplicitJoin(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM t1, t2"
	dialects.VerifiedOnlySelect(t, sql)

	sql2 := "SELECT * FROM t1a NATURAL JOIN t1b, t2a NATURAL JOIN t2b"
	dialects.VerifiedOnlySelect(t, sql2)
}

// TestJoinPrecedence verifies JOIN precedence parsing.
// Reference: tests/sqlparser_common.rs:17228
func TestJoinPrecedence(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// Test that joins are parsed with correct precedence
	_ = dialects.VerifiedStmt(t, "SELECT * FROM t1 NATURAL JOIN t5 INNER JOIN t0 ON (t0.v1 + t5.v0) > 0 WHERE t0.v1 = t1.v0")
}

// TestParseOuterJoin verifies OUTER JOIN parsing.
// Reference: Placeholder for outer join specific tests
func TestParseOuterJoin(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// LEFT OUTER JOIN
	dialects.VerifiedStmt(t, "SELECT * FROM t1 LEFT OUTER JOIN t2 ON t1.id = t2.id")

	// RIGHT OUTER JOIN
	dialects.VerifiedStmt(t, "SELECT * FROM t1 RIGHT OUTER JOIN t2 ON t1.id = t2.id")

	// FULL OUTER JOIN
	dialects.VerifiedStmt(t, "SELECT * FROM t1 FULL OUTER JOIN t2 ON t1.id = t2.id")
}
