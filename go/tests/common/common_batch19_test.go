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
// This file contains tests 341-360 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/statement"
	dialectspkg "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseSelectWildcardWithExcept verifies SELECT * EXCEPT parsing.
// Reference: tests/sqlparser_common.rs:13864
func TestParseSelectWildcardWithExcept(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialectspkg.Dialect) bool {
		return d.SupportsSelectWildcardExcept()
	})

	// Test that SELECT * EXCEPT parses correctly
	dialects.VerifiedOnlySelect(t, "SELECT * EXCEPT (col_a) FROM data")
	dialects.VerifiedOnlySelect(t, "SELECT * EXCEPT (department_id, employee_id) FROM employee_table")

	// Test error case: empty EXCEPT
	_, err := utils.ParseSQLWithDialects(dialects.Dialects, "SELECT * EXCEPT () FROM employee_table")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: identifier, found: )")
}

// runExplainAnalyze is a helper function for EXPLAIN ANALYZE tests
func runExplainAnalyzeWithOptions(t *testing.T, dialects *utils.TestedDialects, sql string, expectedOptions []*expr.UtilityOption) {
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	explain, ok := stmts[0].(*statement.Explain)
	require.True(t, ok, "Expected Explain statement, got %T", stmts[0])

	if expectedOptions != nil {
		assert.Equal(t, len(expectedOptions), len(explain.Options))
	}
}

// TestParseExplainWithOptionList verifies EXPLAIN with option list parsing.
// Reference: tests/sqlparser_common.rs:14072
func TestParseExplainWithOptionList(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialectspkg.Dialect) bool {
		return d.SupportsExplainWithUtilityOptions()
	})

	// Test various EXPLAIN options
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (ANALYZE false, VERBOSE true) SELECT sqrt(id) FROM foo", nil)
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (ANALYZE ON, VERBOSE OFF) SELECT sqrt(id) FROM foo", nil)
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (FORMAT1 TEXT, FORMAT2 'JSON', FORMAT3 \"XML\", FORMAT4 YAML) SELECT sqrt(id) FROM foo", nil)
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (NUM1 10, NUM2 +10.1, NUM3 -10.2) SELECT sqrt(id) FROM foo", nil)
	runExplainAnalyzeWithOptions(t, dialects, "EXPLAIN (ANALYZE, VERBOSE true, WAL OFF, FORMAT YAML, USER_DEF_NUM -100.1) SELECT sqrt(id) FROM foo", nil)
}

// TestParseMethodSelect verifies method call parsing in SELECT.
// Reference: tests/sqlparser_common.rs:14665
func TestParseMethodSelect(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedOnlySelect(t, "SELECT LEFT('abc', 1).value('.', 'NVARCHAR(MAX)').value('.', 'NVARCHAR(MAX)') AS T")
	dialects.VerifiedOnlySelect(t, "SELECT STUFF((SELECT ',' + name FROM sys.objects FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)'), 1, 1, '') AS T")
	dialects.VerifiedOnlySelect(t, "SELECT CAST(column AS XML).value('.', 'NVARCHAR(MAX)') AS T")

	// CONVERT support
	dialects2 := utils.NewTestedDialectsWithFilter(func(d dialectspkg.Dialect) bool {
		return d.SupportsTryConvert() && d.ConvertTypeBeforeValue()
	})
	dialects2.VerifiedOnlySelect(t, "SELECT CONVERT(XML, '<Book>abc</Book>').value('.', 'NVARCHAR(MAX)').value('.', 'NVARCHAR(MAX)') AS T")
}

// TestParseMethodExpr verifies method call expression parsing.
// Reference: tests/sqlparser_common.rs:14679
func TestParseMethodExpr(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test method expressions parse correctly
	dialects.VerifiedStmt(t, "SELECT LEFT('abc', 1).value('.', 'NVARCHAR(MAX)').value('.', 'NVARCHAR(MAX)')")
	dialects.VerifiedStmt(t, "SELECT (SELECT ',' + name FROM sys.objects FOR XML PATH(''), TYPE).value('.', 'NVARCHAR(MAX)')")
	dialects.VerifiedStmt(t, "SELECT CAST(column AS XML).value('.', 'NVARCHAR(MAX)')")

	// CONVERT support
	dialects2 := utils.NewTestedDialectsWithFilter(func(d dialectspkg.Dialect) bool {
		return d.SupportsTryConvert() && d.ConvertTypeBeforeValue()
	})
	dialects2.VerifiedStmt(t, "SELECT CONVERT(XML, '<Book>abc</Book>').value('.', 'NVARCHAR(MAX)').value('.', 'NVARCHAR(MAX)')")
}
