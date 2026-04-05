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

// Package ddl contains the DDL (Data Definition Language) SQL parsing tests.
// These tests are ported from tests/sqlparser_common.rs in the Rust implementation.
package ddl

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseTruncate verifies TRUNCATE statement parsing.
// Reference: tests/sqlparser_common.rs
func TestParseTruncate(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "TRUNCATE TABLE customer"
	dialects.VerifiedStmt(t, sql)
}

// TestParseTruncateOnly verifies TRUNCATE TABLE with ONLY clause parsing.
// Reference: tests/sqlparser_common.rs:17181
func TestParseTruncateOnly(t *testing.T) {
	dialects := utils.NewTestedDialects()
	// TRUNCATE with ONLY - parsing only (struct fields not yet implemented)
	_ = dialects.VerifiedStmt(t, "TRUNCATE TABLE employee, ONLY dept")
}

// TestTruncateTableWithOnCluster verifies TRUNCATE TABLE with ON CLUSTER clause.
// Reference: tests/sqlparser_common.rs:14052
func TestTruncateTableWithOnCluster(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test TRUNCATE with ON CLUSTER
	stmts := dialects.ParseSQL(t, "TRUNCATE TABLE t ON CLUSTER cluster_name")
	require.Len(t, stmts, 1)

	truncate, ok := stmts[0].(*statement.Truncate)
	require.True(t, ok, "Expected Truncate statement, got %T", stmts[0])
	require.NotNil(t, truncate.OnCluster)
	assert.Equal(t, "cluster_name", truncate.OnCluster.String())

	// Test without ON CLUSTER
	dialects.VerifiedStmt(t, "TRUNCATE TABLE t")

	// Test error case - missing cluster name
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "TRUNCATE TABLE t ON CLUSTER")
	require.Error(t, err)
}
