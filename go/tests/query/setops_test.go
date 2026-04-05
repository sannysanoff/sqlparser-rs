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

// TestParseUnion verifies UNION clause parsing.
// Reference: tests/sqlparser_common.rs:394
func TestParseUnion(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM a UNION SELECT * FROM b"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "SELECT * FROM a UNION ALL SELECT * FROM b"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseIntersect verifies INTERSECT clause parsing.
// Reference: tests/sqlparser_common.rs:406
func TestParseIntersect(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM a INTERSECT SELECT * FROM b"
	dialects.VerifiedStmt(t, sql1)
}

// TestParseExcept verifies EXCEPT clause parsing.
// Reference: tests/sqlparser_common.rs:414
func TestParseExcept(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "SELECT * FROM a EXCEPT SELECT * FROM b"
	dialects.VerifiedStmt(t, sql1)
}

// TestParseInUnion verifies IN with UNION subquery parsing.
// Reference: tests/sqlparser_common.rs:2316
func TestParseInUnion(t *testing.T) {
	sql := "SELECT * FROM customers WHERE segment IN ((SELECT segm FROM bar) UNION (SELECT segm FROM bar2))"
	dialects := utils.NewTestedDialects()
	dialects.VerifiedOnlySelect(t, sql)
}

// TestAnySomeAllComparison verifies ANY/SOME/ALL comparison operators.
// Reference: tests/sqlparser_common.rs:14630
func TestAnySomeAllComparison(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT c1 FROM tbl WHERE c1 = ANY(SELECT c2 FROM tbl)")
	dialects.VerifiedStmt(t, "SELECT c1 FROM tbl WHERE c1 >= ALL(SELECT c2 FROM tbl)")
	dialects.VerifiedStmt(t, "SELECT c1 FROM tbl WHERE c1 <> SOME(SELECT c2 FROM tbl)")
	dialects.VerifiedStmt(t, "SELECT 1 = ANY(WITH x AS (SELECT 1) SELECT * FROM x)")
}
