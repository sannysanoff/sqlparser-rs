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
// This file contains tests 321-340 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/tests/utils"
)

// TestSelectFromFirst verifies SELECT FROM first syntax.
// Reference: tests/sqlparser_common.rs:16133
func TestSelectFromFirst(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsFromFirstSelect()
	})

	// Test FROM capitals (no SELECT)
	q1 := "FROM capitals"
	query1 := dialects.VerifiedStmt(t, q1)
	require.NotNil(t, query1)
	require.Equal(t, q1, query1.String())

	// Test FROM capitals SELECT *
	q2 := "FROM capitals SELECT *"
	query2 := dialects.VerifiedStmt(t, q2)
	require.NotNil(t, query2)
	require.Equal(t, q2, query2.String())
}

// TestGeometricUnaryOperators verifies geometric unary operators.
// Reference: tests/sqlparser_common.rs:16197
func TestGeometricUnaryOperators(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsGeometricTypes()
	})

	// Number of points in path or polygon: #
	dialects.VerifiedStmt(t, "SELECT # path '((1,0),(0,1),(-1,0))'")

	// Length or circumference: @-@
	dialects.VerifiedStmt(t, "SELECT @-@ path '((0,0),(1,0))'")

	// Center: @@
	dialects.VerifiedStmt(t, "SELECT @@ circle '((0,0),10)'")

	// Is horizontal?: ?-
	dialects.VerifiedStmt(t, "SELECT ?- lseg '((-1,0),(1,0))'")

	// Is vertical?: ?|
	dialects.VerifiedStmt(t, "SELECT ?| lseg '((-1,0),(1,0))'")
}

// TestGeometryType verifies geometry type casting.
// Reference: tests/sqlparser_common.rs:16249
func TestGeometryType(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsGeometricTypes()
	})

	// point
	dialects.VerifiedStmt(t, "SELECT point '1,2'")

	// line
	dialects.VerifiedStmt(t, "SELECT line '1,2,3,4'")

	// path
	dialects.VerifiedStmt(t, "SELECT path '1,2,3,4'")

	// box
	dialects.VerifiedStmt(t, "SELECT box '1,2,3,4'")

	// circle
	dialects.VerifiedStmt(t, "SELECT circle '1,2,3'")

	// polygon
	dialects.VerifiedStmt(t, "SELECT polygon '1,2,3,4'")

	// lseg (line segment)
	dialects.VerifiedStmt(t, "SELECT lseg '1,2,3,4'")
}

// TestGeometricBinaryOperators verifies geometric binary operators.
// Reference: tests/sqlparser_common.rs:16340
func TestGeometricBinaryOperators(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsGeometricTypes()
	})

	// Translation plus: +
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(1,1))' + point '(2.0,0)'")

	// Translation minus: -
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(1,1))' - point '(2.0,0)'")

	// Scaling multiply: *
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(1,1))' * point '(2.0,0)'")

	// Scaling divide: /
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(1,1))' / point '(2.0,0)'")

	// Intersection: #
	dialects.VerifiedStmt(t, "SELECT '((1,-1),(-1,1))' # '((1,1),(-1,-1))'")

	// Point of closest proximity: ##
	dialects.VerifiedStmt(t, "SELECT point '(0,0)' ## lseg '((2,0),(0,2))'")

	// Overlap: &&
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(1,1))' && box '((0,0),(2,2))'")

	// Overlaps to left?: &<
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(1,1))' &< box '((0,0),(2,2))'")

	// Overlaps to right?: &>
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(3,3))' &> box '((0,0),(2,2))'")

	// Distance between: <->
	dialects.VerifiedStmt(t, "SELECT circle '((0,0),1)' <-> circle '((5,0),1)'")

	// Is left of?: <<
	dialects.VerifiedStmt(t, "SELECT circle '((0,0),1)' << circle '((5,0),1)'")

	// Is right of?: >>
	dialects.VerifiedStmt(t, "SELECT circle '((5,0),1)' >> circle '((0,0),1)'")

	// Is below?: <^
	dialects.VerifiedStmt(t, "SELECT circle '((0,0),1)' <^ circle '((0,5),1)'")

	// Intersects or overlaps: ?#
	dialects.VerifiedStmt(t, "SELECT lseg '((-1,0),(1,0))' ?# box '((-2,-2),(2,2))'")

	// Is horizontal?: ?-
	dialects.VerifiedStmt(t, "SELECT point '(1,0)' ?- point '(0,0)'")

	// Is perpendicular?: ?-|
	dialects.VerifiedStmt(t, "SELECT lseg '((0,0),(0,1))' ?-| lseg '((0,0),(1,0))'")

	// Is vertical?: ?|
	dialects.VerifiedStmt(t, "SELECT point '(0,1)' ?| point '(0,0)'")

	// Are parallel?: ?||
	dialects.VerifiedStmt(t, "SELECT lseg '((-1,0),(1,0))' ?|| lseg '((-1,2),(1,2))'")

	// Contained or on?: @
	dialects.VerifiedStmt(t, "SELECT point '(1,1)' @ circle '((0,0),2)'")

	// Same as?: ~=
	dialects.VerifiedStmt(t, "SELECT polygon '((0,0),(1,1))' ~= polygon '((1,1),(0,0))'")

	// Is strictly below?: <<|
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(3,3))' <<| box '((3,4),(5,5))'")

	// Is strictly above?: |>>
	dialects.VerifiedStmt(t, "SELECT box '((3,4),(5,5))' |>> box '((0,0),(3,3))'")

	// Does not extend above?: &<|
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(1,1))' &<| box '((0,0),(2,2))'")

	// Does not extend below?: |&>
	dialects.VerifiedStmt(t, "SELECT box '((0,0),(3,3))' |&> box '((0,0),(2,2))'")
}

// TestParseArrayTypeDefWithBrackets verifies array type definition with brackets.
// Reference: tests/sqlparser_common.rs:16582
func TestParseArrayTypeDefWithBrackets(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsArrayTypedefWithBrackets()
	})

	dialects.VerifiedStmt(t, "SELECT x::INT[]")
	dialects.VerifiedStmt(t, "SELECT STRING_TO_ARRAY('1,2,3', ',')::INT[3]")
}

// TestParseSetNames verifies SET NAMES statement parsing.
// Reference: tests/sqlparser_common.rs:16589
func TestParseSetNames(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsSetNames()
	})

	dialects.VerifiedStmt(t, "SET NAMES 'UTF8'")
	dialects.VerifiedStmt(t, "SET NAMES 'utf8'")
	dialects.VerifiedStmt(t, "SET NAMES UTF8 COLLATE bogus")
}

// TestParsePipeOperatorAs verifies pipe operator AS syntax.
// Reference: tests/sqlparser_common.rs:16597
func TestParsePipeOperatorAs(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AS new_users")
}

// TestParsePipeOperatorSelect verifies pipe operator SELECT syntax.
// Reference: tests/sqlparser_common.rs:16603
func TestParsePipeOperatorSelect(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> SELECT id")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> SELECT id, name")
	dialects.OneStatementParsesTo(t, "SELECT * FROM tbl |> SELECT id user_id", "SELECT * FROM tbl |> SELECT id AS user_id")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> SELECT id AS user_id")
}

// TestParsePipeOperatorExtend verifies pipe operator EXTEND syntax.
// Reference: tests/sqlparser_common.rs:16615
func TestParsePipeOperatorExtend(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> EXTEND id + 1 AS new_id")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> EXTEND id AS new_id, name AS new_name")
	dialects.OneStatementParsesTo(t, "SELECT * FROM tbl |> EXTEND id user_id", "SELECT * FROM tbl |> EXTEND id AS user_id")
}

// TestParsePipeOperatorSet verifies pipe operator SET syntax.
// Reference: tests/sqlparser_common.rs:16626
func TestParsePipeOperatorSet(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> SET id = id + 1")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> SET id = id + 1, name = name + ' Doe'")
}

// TestParsePipeOperatorDrop verifies pipe operator DROP syntax.
// Reference: tests/sqlparser_common.rs:16633
func TestParsePipeOperatorDrop(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> DROP id")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> DROP id, name")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> DROP c |> RENAME a AS x")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> DROP a, b |> SELECT c")
}

// TestParsePipeOperatorLimit verifies pipe operator LIMIT syntax.
// Reference: tests/sqlparser_common.rs:16642
func TestParsePipeOperatorLimit(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> LIMIT 10")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> LIMIT 10 OFFSET 5")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> LIMIT 10 |> LIMIT 5")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> LIMIT 10 |> WHERE true")
}

// TestParsePipeOperatorWhere verifies pipe operator WHERE syntax.
// Reference: tests/sqlparser_common.rs:16651
func TestParsePipeOperatorWhere(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> WHERE id = 1")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> WHERE id = 1 AND name = 'John'")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> WHERE id = 1 OR name = 'John'")
}

// TestParsePipeOperatorAggregate verifies pipe operator AGGREGATE syntax.
// Reference: tests/sqlparser_common.rs:16659
func TestParsePipeOperatorAggregate(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE COUNT(*)")
	dialects.OneStatementParsesTo(t, "SELECT * FROM tbl |> AGGREGATE COUNT(*) total_users", "SELECT * FROM tbl |> AGGREGATE COUNT(*) AS total_users")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE COUNT(*) AS total_users")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE COUNT(*), MIN(id)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE SUM(o_totalprice) AS price, COUNT(*) AS cnt GROUP BY EXTRACT(YEAR FROM o_orderdate) AS year")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE GROUP BY EXTRACT(YEAR FROM o_orderdate) AS year")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE GROUP BY EXTRACT(YEAR FROM o_orderdate)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE GROUP BY a, b")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE SUM(c) GROUP BY a, b")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> AGGREGATE SUM(c) ASC")
}

// TestParsePipeOperatorOrderBy verifies pipe operator ORDER BY syntax.
// Reference: tests/sqlparser_common.rs:16680
func TestParsePipeOperatorOrderBy(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> ORDER BY id ASC")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> ORDER BY id DESC")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> ORDER BY id DESC, name ASC")
}

// TestParsePipeOperatorTablesample verifies pipe operator TABLESAMPLE syntax.
// Reference: tests/sqlparser_common.rs:16688
func TestParsePipeOperatorTablesample(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> TABLESAMPLE BERNOULLI (50)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> TABLESAMPLE SYSTEM (50 PERCENT)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> TABLESAMPLE SYSTEM (50) REPEATABLE (10)")
}

// TestParsePipeOperatorRename verifies pipe operator RENAME syntax.
// Reference: tests/sqlparser_common.rs:16696
func TestParsePipeOperatorRename(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> RENAME old_name AS new_name")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> RENAME id AS user_id, name AS user_name")
	dialects.OneStatementParsesTo(t, "SELECT * FROM tbl |> RENAME id user_id", "SELECT * FROM tbl |> RENAME id AS user_id")
}

// TestParsePipeOperatorUnion verifies pipe operator UNION syntax.
// Reference: tests/sqlparser_common.rs:16707
func TestParsePipeOperatorUnion(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION ALL (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION DISTINCT (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION ALL (SELECT * FROM admins), (SELECT * FROM guests)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION DISTINCT (SELECT * FROM admins), (SELECT * FROM guests), (SELECT * FROM employees)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION (SELECT * FROM admins), (SELECT * FROM guests)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION BY NAME (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION ALL BY NAME (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION DISTINCT BY NAME (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> UNION BY NAME (SELECT * FROM admins), (SELECT * FROM guests)")
}

// TestParsePipeOperatorIntersect verifies pipe operator INTERSECT syntax.
// Reference: tests/sqlparser_common.rs:16727
func TestParsePipeOperatorIntersect(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> INTERSECT DISTINCT (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> INTERSECT DISTINCT BY NAME (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> INTERSECT DISTINCT (SELECT * FROM admins), (SELECT * FROM guests)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> INTERSECT DISTINCT BY NAME (SELECT * FROM admins), (SELECT * FROM guests)")
}

// TestParsePipeOperatorExcept verifies pipe operator EXCEPT syntax.
// Reference: tests/sqlparser_common.rs:16739
func TestParsePipeOperatorExcept(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsPipeOperator()
	})

	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> EXCEPT DISTINCT (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> EXCEPT DISTINCT BY NAME (SELECT * FROM admins)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> EXCEPT DISTINCT (SELECT * FROM admins), (SELECT * FROM guests)")
	dialects.VerifiedStmt(t, "SELECT * FROM tbl |> EXCEPT DISTINCT BY NAME (SELECT * FROM admins), (SELECT * FROM guests)")
}
