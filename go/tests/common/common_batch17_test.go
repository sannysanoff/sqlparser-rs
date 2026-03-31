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
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mssql"
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

// TestParseEscapedStringWithUnescape verifies string unescaping.
// Reference: tests/sqlparser_common.rs:11161
func TestParseEscapedStringWithUnescape(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test with escape enabled (default)
	sql := "SELECT '\\n\\t\\\\'"
	stmts, err := parser.ParseSQL(dialects.Dialects[0], sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParseEscapedStringWithoutUnescape verifies string parsing without unescape.
// Reference: tests/sqlparser_common.rs:11214
func TestParseEscapedStringWithoutUnescape(t *testing.T) {
	dialect := generic.NewGenericDialect()

	// Test with escape disabled
	sql := "SELECT '\\\\n'"
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)
}

// TestParsePivotTable verifies PIVOT table expression parsing.
// Reference: tests/sqlparser_common.rs:11256
func TestParsePivotTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM monthly_sales PIVOT (SUM(amount) FOR month IN ('JAN', 'FEB'))"
	dialects.VerifiedStmt(t, sql)
}

// TestParseUnpivotTable verifies UNPIVOT table expression parsing.
// Reference: tests/sqlparser_common.rs:11434
func TestParseUnpivotTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM quarterly_sales UNPIVOT (amount FOR quarter IN (q1, q2, q3, q4))"
	dialects.VerifiedStmt(t, sql)
}

// TestParseSelectTableWithIndexHints verifies MySQL index hints parsing.
// Reference: tests/sqlparser_common.rs:11649
func TestParseSelectTableWithIndexHints(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsIndexHints()
	})

	sql := "SELECT * FROM t USE INDEX (idx1)"
	dialects.VerifiedStmt(t, sql)
}

// TestParsePivotUnpivotTable verifies combined PIVOT after UNPIVOT.
// Reference: tests/sqlparser_common.rs:11713
func TestParsePivotUnpivotTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "SELECT * FROM (SELECT * FROM sales UNPIVOT (amount FOR month IN (jan, feb))) PIVOT (SUM(amount) FOR month IN ('JAN', 'FEB'))"
	dialects.VerifiedStmt(t, sql)
}

// TestParseNonLatinIdentifiers verifies Unicode/non-Latin identifier parsing.
// Reference: tests/sqlparser_common.rs:11794
func TestParseNonLatinIdentifiers(t *testing.T) {
	// Test with dialects that support non-Latin identifiers
	sql := "SELECT 説明 FROM table1"
	dialect := bigquery.NewBigQueryDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Test multiple non-Latin identifiers
	sql = "SELECT 説明, hühnervögel, garçon, Москва, 東京 FROM inter01"
	stmts, err = parser.ParseSQL(dialect, sql)
	require.NoError(t, err)
	require.Len(t, stmts, 1)

	// Dialects that don't support emoji
	dialectsWithEmojiRestriction := []dialects.Dialect{
		generic.NewGenericDialect(),
		mssql.NewMsSqlDialect(),
	}

	for _, dialect := range dialectsWithEmojiRestriction {
		sql := "SELECT 💝 FROM table1"
		_, err := parser.ParseSQL(dialect, sql)
		assert.Error(t, err, "Expected error for emoji identifier with %T", dialect)
	}
}
