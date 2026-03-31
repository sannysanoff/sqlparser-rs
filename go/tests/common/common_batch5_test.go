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
// This file contains tests 61-80 from the Rust test suite.
package common

import (
	"fmt"
	"testing"

	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseCompoundExpr1 verifies compound expression with precedence.
// Reference: tests/sqlparser_common.rs:1579
func TestParseCompoundExpr1(t *testing.T) {
	sql := "a + b * c"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseCompoundExpr2 verifies compound expression with precedence.
// Reference: tests/sqlparser_common.rs:1598
func TestParseCompoundExpr2(t *testing.T) {
	sql := "a * b + c"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseUnaryMathWithPlus verifies unary minus with plus operator.
// Reference: tests/sqlparser_common.rs:1617
func TestParseUnaryMathWithPlus(t *testing.T) {
	sql := "-a + -b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseUnaryMathWithMultiply verifies unary minus with multiply operator.
// Reference: tests/sqlparser_common.rs:1637
func TestParseUnaryMathWithMultiply(t *testing.T) {
	sql := "-a * -b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseMod verifies modulo operator parsing.
// Reference: tests/sqlparser_common.rs:1657
func TestParseMod(t *testing.T) {
	sql := "a % b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseJsonOpsWithoutColon verifies JSON operators parsing.
// Reference: tests/sqlparser_common.rs:1682
func TestParseJsonOpsWithoutColon(t *testing.T) {
	// Test arrow operator (->) - PostgreSQL and Generic
	pgGeneric := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
			generic.NewGenericDialect(),
		},
	}
	pgGeneric.VerifiedOnlySelect(t, "SELECT a -> b")

	// Test long arrow operator (->>) - all dialects
	allDialects := utils.NewTestedDialects()
	allDialects.VerifiedOnlySelect(t, "SELECT a ->> b")

	// Test hash arrow operator (#>) - PostgreSQL and Generic
	pgGeneric.VerifiedOnlySelect(t, "SELECT a #> b")

	// Test hash long arrow operator (#>>) - PostgreSQL and Generic
	pgGeneric.VerifiedOnlySelect(t, "SELECT a #>> b")

	// Test at arrow operator (@>) - all dialects
	allDialects.VerifiedOnlySelect(t, "SELECT a @> b")

	// Test arrow at operator (<@) - all dialects
	allDialects.VerifiedOnlySelect(t, "SELECT a <@ b")

	// Test hash minus operator (#-) - PostgreSQL and Generic
	pgGeneric.VerifiedOnlySelect(t, "SELECT a #- b")

	// Test at question operator (@?) - all dialects
	allDialects.VerifiedOnlySelect(t, "SELECT a @? b")

	// Test at at operator (@@) - all dialects
	allDialects.VerifiedOnlySelect(t, "SELECT a @@ b")
}

// TestParseJsonObject verifies JSON_OBJECT function parsing.
// Reference: tests/sqlparser_common.rs:1715
func TestParseJsonObject(t *testing.T) {
	dialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			mssql.NewMsSqlDialect(),
			postgresql.NewPostgreSqlDialect(),
		},
	}

	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT('name' : 'value', 'type' : 1)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT('name' : 'value', 'type' : NULL ABSENT ON NULL)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT(NULL ON NULL)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT(ABSENT ON NULL)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT('name' : 'value', 'type' : JSON_ARRAY(1, 2) ABSENT ON NULL)")
	dialects.VerifiedOnlySelect(t, "SELECT JSON_OBJECT('name' : 'value', 'type' : JSON_OBJECT('type_id' : 1, 'name' : 'a') NULL ON NULL)")
}

// TestParseModNoSpaces verifies modulo operator without spaces.
// Reference: tests/sqlparser_common.rs:1892
func TestParseModNoSpaces(t *testing.T) {
	canonical := "a1 % b1"
	sqls := []string{"a1 % b1", "a1% b1", "a1 %b1", "a1%b1"}

	pgGeneric := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
			generic.NewGenericDialect(),
		},
	}

	for _, sql := range sqls {
		fullSql := fmt.Sprintf("SELECT %s FROM t", sql)
		fullCanonical := fmt.Sprintf("SELECT %s FROM t", canonical)
		pgGeneric.OneStatementParsesTo(t, fullSql, fullCanonical)
	}
}

// TestParseIsNull verifies IS NULL expression parsing.
// Reference: tests/sqlparser_common.rs:1910
func TestParseIsNull(t *testing.T) {
	sql := "a IS NULL"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseIsNotNull verifies IS NOT NULL expression parsing.
// Reference: tests/sqlparser_common.rs:1920
func TestParseIsNotNull(t *testing.T) {
	sql := "a IS NOT NULL"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseIsDistinctFrom verifies IS DISTINCT FROM expression parsing.
// Reference: tests/sqlparser_common.rs:1930
func TestParseIsDistinctFrom(t *testing.T) {
	sql := "a IS DISTINCT FROM b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseIsNotDistinctFrom verifies IS NOT DISTINCT FROM expression parsing.
// Reference: tests/sqlparser_common.rs:1943
func TestParseIsNotDistinctFrom(t *testing.T) {
	sql := "a IS NOT DISTINCT FROM b"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, fmt.Sprintf("SELECT %s FROM t", sql))
}

// TestParseNotPrecedence verifies NOT operator precedence.
// Reference: tests/sqlparser_common.rs:1956
func TestParseNotPrecedence(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// NOT has higher precedence than OR/AND
	dialects.VerifiedOnlySelect(t, "SELECT NOT 1 OR 1 FROM t")

	// NOT has lower precedence than comparison operators
	dialects.VerifiedOnlySelect(t, "SELECT NOT a IS NULL FROM t")

	// NOT has lower precedence than BETWEEN
	dialects.VerifiedOnlySelect(t, "SELECT NOT 1 NOT BETWEEN 1 AND 2 FROM t")

	// NOT has lower precedence than LIKE
	dialects.VerifiedOnlySelect(t, "SELECT NOT 'a' NOT LIKE 'b' FROM t")

	// NOT has lower precedence than IN
	dialects.VerifiedOnlySelect(t, "SELECT NOT a NOT IN ('a') FROM t")
}

// TestParseNullLike verifies LIKE with NULL operands.
// Reference: tests/sqlparser_common.rs:2030
func TestParseNullLike(t *testing.T) {
	sql := "SELECT column1 LIKE NULL AS col_null, NULL LIKE column1 AS null_col FROM customers"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseIlike verifies ILIKE expression parsing.
// Reference: tests/sqlparser_common.rs:2073
func TestParseIlike(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test ILIKE without NOT
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name ILIKE '%a'")

	// Test ILIKE with escape char
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name ILIKE '%a' ESCAPE '^'")

	// Test NOT ILIKE
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT ILIKE '%a'")

	// Test NOT ILIKE with escape char
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT ILIKE '%a' ESCAPE '^'")

	// Test ILIKE precedence with IS NULL
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name ILIKE '%a' IS NULL")

	// Test NOT ILIKE precedence with IS NULL
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT ILIKE '%a' IS NULL")
}

// TestParseLike verifies LIKE expression parsing.
// Reference: tests/sqlparser_common.rs:2137
func TestParseLike(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test LIKE without NOT
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name LIKE '%a'")

	// Test LIKE with escape char
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name LIKE '%a' ESCAPE '^'")

	// Test NOT LIKE
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT LIKE '%a'")

	// Test NOT LIKE with escape char
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT LIKE '%a' ESCAPE '^'")

	// Test LIKE precedence with IS NULL
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name LIKE '%a' IS NULL")

	// Test NOT LIKE precedence with IS NULL
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT LIKE '%a' IS NULL")
}

// TestParseSimilarTo verifies SIMILAR TO expression parsing.
// Reference: tests/sqlparser_common.rs:2201
func TestParseSimilarTo(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test SIMILAR TO without NOT
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name SIMILAR TO '%a'")

	// Test SIMILAR TO with escape char
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name SIMILAR TO '%a' ESCAPE '^'")

	// Test SIMILAR TO with NULL escape
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name SIMILAR TO '%a' ESCAPE NULL")

	// Test NOT SIMILAR TO
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT SIMILAR TO '%a'")

	// Test NOT SIMILAR TO with escape char
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT SIMILAR TO '%a' ESCAPE '^'")

	// Test SIMILAR TO precedence with IS NULL
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name SIMILAR TO '%a' ESCAPE '^' IS NULL")

	// Test NOT SIMILAR TO precedence with IS NULL
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE name NOT SIMILAR TO '%a' ESCAPE '^' IS NULL")
}

// TestParseInList verifies IN list expression parsing.
// Reference: tests/sqlparser_common.rs:2278
func TestParseInList(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test IN
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE segment IN ('HIGH', 'MED')")

	// Test NOT IN
	dialects.VerifiedOnlySelect(t, "SELECT * FROM customers WHERE segment NOT IN ('HIGH', 'MED')")
}

// TestParseInSubquery verifies IN with subquery parsing.
// Reference: tests/sqlparser_common.rs:2302
func TestParseInSubquery(t *testing.T) {
	sql := "SELECT * FROM customers WHERE segment IN (SELECT segm FROM bar)"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, sql)
}

// TestParseInUnion verifies IN with UNION subquery parsing.
// Reference: tests/sqlparser_common.rs:2316
func TestParseInUnion(t *testing.T) {
	sql := "SELECT * FROM customers WHERE segment IN ((SELECT segm FROM bar) UNION (SELECT segm FROM bar2))"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, sql)
}
