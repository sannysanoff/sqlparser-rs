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

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/sannysanoff/sqlparser-rs/go/dialects"
	"github.com/sannysanoff/sqlparser-rs/go/dialects/clickhouse"
	"github.com/sannysanoff/sqlparser-rs/go/tests/utils"
)

// TestParseCTE verifies Common Table Expression parsing.
// Reference: tests/sqlparser_common.rs:421
func TestParseCTE(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "WITH cte AS (SELECT * FROM customer) SELECT * FROM cte"
	dialects.VerifiedQuery(t, sql)
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

// TestMergeInCte verifies MERGE in CTE.
// Reference: tests/sqlparser_common.rs:10206
func TestMergeInCte(t *testing.T) {
	// Note: MERGE in CTE is not yet fully implemented for all dialects
	t.Skip("MERGE in CTE not yet fully implemented in Go port")
}

// TestParseClickHouseScalarCTE verifies ClickHouse scalar/expression Common Table
// Expressions (WITH <expression> AS <name>), which can be mixed with ordinary
// subquery CTEs. This feature is ClickHouse-specific, so it is only exercised
// against the ClickHouse dialect.
func TestParseClickHouseScalarCTE(t *testing.T) {
	dial := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		_, isClickHouse := d.(*clickhouse.ClickHouseDialect)
		return isClickHouse
	})
	require.NotEmpty(t, dial.Dialects, "ClickHouse dialect must be present")

	valid := []string{
		"WITH 'SCHWAB' AS tenant SELECT tenant",
		"WITH 1 AS x SELECT x",
		"WITH now() AS n SELECT n",
		"WITH sum(bytes) AS s SELECT s FROM hits",
		// Scalar and subquery CTEs can be mixed in a single WITH clause.
		"WITH a AS (SELECT 1), 'b' AS sc SELECT * FROM a, sc",
		"WITH '2019-08-01 15:23:00' AS ts_upper_bound SELECT * FROM hits WHERE EventDate = toDate(ts_upper_bound)",
	}
	for _, sql := range valid {
		dial.VerifiedStmt(t, sql)
	}

	// Trailing semicolons are dropped during serialization, so the reported
	// case is verified for parseability (no error) rather than exact round-trip.
	stmts := dial.ParseSQL(t, "WITH 'SCHWAB' AS tenant SELECT tenant;")
	require.Len(t, stmts, 1)

	// Ordinary subquery CTEs must still parse identically.
	dial.VerifiedStmt(t, "WITH x AS (SELECT 1) SELECT * FROM x")
	dial.VerifiedStmt(t, "WITH cte AS (SELECT * FROM customer) SELECT * FROM cte")
}

// TestParseSubqueryLimit verifies subquery with LIMIT parsing.
// Reference: tests/sqlparser_common.rs:17164
func TestParseSubqueryLimit(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "SELECT t1_id, t1_name FROM t1 WHERE t1_id IN (SELECT t2_id FROM t2 WHERE t1_name = t2_name LIMIT 10)"
	_ = dialects.VerifiedStmt(t, sql)
}
