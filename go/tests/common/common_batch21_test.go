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
// This file contains tests 381-400 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/require"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseOverlaps verifies OVERLAPS expression parsing.
// Reference: tests/sqlparser_common.rs:15741
func TestParseOverlaps(t *testing.T) {
	dialects := utils.NewTestedDialects()
	dialects.VerifiedStmt(t, "SELECT (DATE '2016-01-10', DATE '2016-02-01') OVERLAPS (DATE '2016-01-20', DATE '2016-02-10')")
}

// TestParseColumnDefinitionTrailingCommas verifies column definition trailing comma parsing.
// Reference: tests/sqlparser_common.rs:15746
func TestParseColumnDefinitionTrailingCommas(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsColumnDefinitionTrailingCommas()
	})

	dialects.OneStatementParsesTo(t, "CREATE TABLE T (x INT64,)", "CREATE TABLE T (x INT64)")
	dialects.OneStatementParsesTo(t, "CREATE TABLE T (x INT64, y INT64, )", "CREATE TABLE T (x INT64, y INT64)")
	dialects.OneStatementParsesTo(t, "CREATE VIEW T (x, y, ) AS SELECT 1", "CREATE VIEW T (x, y) AS SELECT 1")

	// Test unsupported dialects get an error
	unsupportedDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsProjectionTrailingCommas() && !d.SupportsTrailingCommas()
	})

	_, err := utils.ParseSQLWithDialects(unsupportedDialects.Dialects, "CREATE TABLE employees (name text, age int,)")
	require.Error(t, err)
}

// TestTrailingCommasInFrom verifies trailing commas in FROM clause parsing.
// Reference: tests/sqlparser_common.rs:15772
func TestTrailingCommasInFrom(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsFromTrailingCommas()
	})

	dialects.VerifiedOnlySelectWithCanonical(t, "SELECT 1, 2 FROM t,", "SELECT 1, 2 FROM t")
	dialects.VerifiedOnlySelectWithCanonical(t, "SELECT 1, 2 FROM t1, t2,", "SELECT 1, 2 FROM t1, t2")

	sql := "SELECT a, FROM b, LIMIT 1"
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	sql = "INSERT INTO a SELECT b FROM c,"
	_, err = utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	sql = "SELECT a FROM b, HAVING COUNT(*) > 1"
	_, err = utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	sql = "SELECT a FROM b, WHERE c = 1"
	_, err = utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	// nested
	sql = "SELECT 1, 2 FROM (SELECT * FROM t,),"
	_, err = utils.ParseSQLWithDialects(dialects.Dialects, sql)
	require.NoError(t, err)

	// multiple_subqueries
	dialects.VerifiedOnlySelectWithCanonical(t, "SELECT 1, 2 FROM (SELECT * FROM t1), (SELECT * FROM t2),", "SELECT 1, 2 FROM (SELECT * FROM t1), (SELECT * FROM t2)")
}

// TestParseCaseStatement verifies CASE statement parsing.
// Reference: tests/sqlparser_common.rs:15828
func TestParseCaseStatement(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Case statement forms - parsing only (struct fields not yet implemented)
	dialects.VerifiedStmt(t, "CASE 1 WHEN 2 THEN SELECT 1; SELECT 2; ELSE SELECT 3; END CASE")
	dialects.VerifiedStmt(t, "CASE 1 WHEN a THEN SELECT 1; SELECT 2; SELECT 3; WHEN b THEN SELECT 4; SELECT 5; ELSE SELECT 7; SELECT 8; END CASE")
	dialects.VerifiedStmt(t, "CASE 1 WHEN a THEN SELECT 1; SELECT 2; SELECT 3; WHEN b THEN SELECT 4; SELECT 5; END CASE")
	dialects.VerifiedStmt(t, "CASE 1 WHEN a THEN SELECT 1; SELECT 2; SELECT 3; END CASE")
	dialects.VerifiedStmt(t, "CASE 1 WHEN a THEN SELECT 1; SELECT 2; SELECT 3; END")

	// Error cases
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, "CASE 1 WHEN a END")
	require.Error(t, err)

	_, err = utils.ParseSQLWithDialects(dialects.Dialects, "CASE 1 ELSE SELECT 1; END")
	require.Error(t, err)
}

// TestParseIfStatement verifies IF statement parsing.
// Reference: tests/sqlparser_common.rs:15894
func TestParseIfStatement(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		_, ok := d.(*mssql.MsSqlDialect)
		return !ok
	})

	// IF statement forms - parsing only (struct fields not yet implemented)
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; ELSEIF 2 THEN SELECT 2; ELSE SELECT 3; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; SELECT 2; SELECT 3; ELSEIF 2 THEN SELECT 4; SELECT 5; ELSEIF 3 THEN SELECT 6; SELECT 7; ELSE SELECT 8; SELECT 9; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; SELECT 2; ELSE SELECT 3; SELECT 4; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; SELECT 2; SELECT 3; ELSEIF 2 THEN SELECT 3; SELECT 4; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; SELECT 2; END IF")
	dialects.VerifiedStmt(t, "IF (1) THEN SELECT 1; SELECT 2; END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN END IF")
	dialects.VerifiedStmt(t, "IF 1 THEN SELECT 1; ELSEIF 1 THEN END IF")

	// Error case
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, "IF 1 THEN SELECT 1; ELSEIF 1 THEN SELECT 2; END")
	require.Error(t, err)
}

// TestParseRaiseStatement verifies RAISE statement parsing.
// Reference: tests/sqlparser_common.rs:16021
func TestParseRaiseStatement(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// RAISE statement forms - parsing only (struct fields not yet implemented)
	dialects.VerifiedStmt(t, "RAISE USING MESSAGE = 42")
	dialects.VerifiedStmt(t, "RAISE USING MESSAGE = 'error'")
	dialects.VerifiedStmt(t, "RAISE myerror")
	dialects.VerifiedStmt(t, "RAISE 42")
	dialects.VerifiedStmt(t, "RAISE using")
	dialects.VerifiedStmt(t, "RAISE")

	// Error case
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, "RAISE USING MESSAGE error")
	require.Error(t, err)
}

// TestLambdas verifies lambda function parsing.
// Reference: tests/sqlparser_common.rs:16043
func TestLambdas(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsLambdaFunctions()
	})

	// Verify various lambda expressions parse
	dialects.VerifiedExpr(t, "map_zip_with(map(1, 'a', 2, 'b'), map(1, 'x', 2, 'y'), (k, v1, v2) -> concat(v1, v2))")
	dialects.VerifiedExpr(t, "transform(array(1, 2, 3), x -> x + 1)")

	// Ensure all lambda variants are parsed correctly
	dialects.VerifiedExpr(t, "a -> a * 2")                // Single parameter without type
	dialects.VerifiedExpr(t, "a INT -> a * 2")            // Single parameter with type
	dialects.VerifiedExpr(t, "(a, b) -> a * b")           // Multiple parameters without types
	dialects.VerifiedExpr(t, "(a INT, b FLOAT) -> a * b") // Multiple parameters with types
}
