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
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseDropSchema tests DROP SCHEMA statement parsing.
// Reference: tests/sqlparser_common.rs:4466 (test 135)
func TestParseDropSchema(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "DROP SCHEMA X"
	ast := dialects.VerifiedStmt(t, sql)
	require.NotNil(t, ast)

	// Check it's a Drop statement
	dropStmt, ok := ast.(*statement.Drop)
	require.True(t, ok)
	assert.Equal(t, "schema", dropStmt.ObjectType)
}

// TestParseDropTable verifies DROP TABLE statement parsing.
// Reference: tests/sqlparser_common.rs:8637
func TestParseDropTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test basic DROP TABLE
	sql := "DROP TABLE foo"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	dropStmt, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])

	// Verify if_exists is false
	assert.False(t, dropStmt.IfExists, "Expected if_exists to be false")

	// Verify object_type is Table
	assert.Equal(t, "TABLE", dropStmt.ObjectType.String())

	// Verify names
	require.Equal(t, 1, len(dropStmt.Names))
	assert.Equal(t, "foo", dropStmt.Names[0].String())

	// Verify cascade is false
	assert.False(t, dropStmt.Cascade, "Expected cascade to be false")

	// Verify temporary is false
	assert.False(t, dropStmt.Temporary, "Expected temporary to be false")

	// Test DROP TABLE IF EXISTS with multiple tables and CASCADE
	sql2 := "DROP TABLE IF EXISTS foo, bar CASCADE"
	stmts2 := dialects.ParseSQL(t, sql2)
	require.Len(t, stmts2, 1)

	dropStmt2, ok := stmts2[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts2[0])

	// Verify if_exists is true
	assert.True(t, dropStmt2.IfExists, "Expected if_exists to be true")

	// Verify object_type is Table
	assert.Equal(t, "TABLE", dropStmt2.ObjectType.String())

	// Verify names
	require.Equal(t, 2, len(dropStmt2.Names))
	assert.Equal(t, "foo", dropStmt2.Names[0].String())
	assert.Equal(t, "bar", dropStmt2.Names[1].String())

	// Verify cascade is true
	assert.True(t, dropStmt2.Cascade, "Expected cascade to be true")

	// Verify temporary is false
	assert.False(t, dropStmt2.Temporary, "Expected temporary to be false")

	// Test DROP TABLE without table name should fail
	sql3 := "DROP TABLE"
	_, err := parser.ParseSQL(dialects.Dialects[0], sql3)
	require.Error(t, err, "Expected error for DROP TABLE without table name")
	assert.Contains(t, err.Error(), "identifier", "Error should mention identifier")

	// Test DROP TABLE with both CASCADE and RESTRICT should fail
	sql4 := "DROP TABLE IF EXISTS foo, bar CASCADE RESTRICT"
	_, err = parser.ParseSQL(dialects.Dialects[0], sql4)
	require.Error(t, err, "Expected error for both CASCADE and RESTRICT")
	assert.Contains(t, err.Error(), "CASCADE", "Error should mention CASCADE")
	assert.Contains(t, err.Error(), "RESTRICT", "Error should mention RESTRICT")
}

// TestParseDropView verifies DROP VIEW statement parsing.
// Reference: tests/sqlparser_common.rs:8697
func TestParseDropView(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test DROP VIEW
	sql := "DROP VIEW myschema.myview"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	dropStmt, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])

	// Verify names
	require.Equal(t, 1, len(dropStmt.Names))
	assert.Equal(t, "myschema.myview", dropStmt.Names[0].String())

	// Verify object_type is View
	assert.Equal(t, "VIEW", dropStmt.ObjectType.String())

	// Test DROP MATERIALIZED VIEW
	sql2 := "DROP MATERIALIZED VIEW a.b.c"
	dialects.VerifiedStmt(t, sql2)

	// Test DROP MATERIALIZED VIEW IF EXISTS
	sql3 := "DROP MATERIALIZED VIEW IF EXISTS a.b.c"
	dialects.VerifiedStmt(t, sql3)
}

// TestParseDropUser verifies DROP USER statement parsing.
// Reference: tests/sqlparser_common.rs:8717
func TestParseDropUser(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test DROP USER
	sql := "DROP USER u1"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	dropStmt, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])

	// Verify names
	require.Equal(t, 1, len(dropStmt.Names))
	assert.Equal(t, "u1", dropStmt.Names[0].String())

	// Verify object_type is User
	assert.Equal(t, "USER", dropStmt.ObjectType.String())

	// Test DROP USER IF EXISTS
	sql2 := "DROP USER IF EXISTS u1"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseDropIndex verifies DROP INDEX statement parsing.
// Reference: tests/sqlparser_common.rs:9608
func TestParseDropIndex(t *testing.T) {
	sql := "DROP INDEX idx_a"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	drop, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])
	assert.Equal(t, "INDEX", drop.ObjectType.String(), "Expected object type to be Index")
	assert.Equal(t, 1, len(drop.Names), "Expected 1 name")
	assert.Equal(t, "idx_a", drop.Names[0].String())
}

// TestParseDropRole verifies DROP ROLE statement parsing.
// Reference: tests/sqlparser_common.rs:9645
func TestParseDropRole(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Single role
	sql := "DROP ROLE abc"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	drop, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])
	assert.Equal(t, "ROLE", drop.ObjectType.String(), "Expected object type to be Role")
	assert.Equal(t, 1, len(drop.Names), "Expected 1 name")
	assert.Equal(t, "abc", drop.Names[0].String())
	assert.False(t, drop.IfExists, "Expected if_exists to be false")

	// Multiple roles with IF EXISTS
	sql = "DROP ROLE IF EXISTS def, magician, quaternion"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	drop, ok = stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])
	assert.Equal(t, "ROLE", drop.ObjectType.String(), "Expected object type to be Role")
	assert.Equal(t, 3, len(drop.Names), "Expected 3 names")
	assert.Equal(t, "def", drop.Names[0].String())
	assert.Equal(t, "magician", drop.Names[1].String())
	assert.Equal(t, "quaternion", drop.Names[2].String())
	assert.True(t, drop.IfExists, "Expected if_exists to be true")
}

// TestDropPolicy verifies DROP POLICY statement parsing.
// Reference: tests/sqlparser_common.rs:14323
func TestDropPolicy(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test DROP POLICY with IF EXISTS and RESTRICT
	stmts := dialects.ParseSQL(t, "DROP POLICY IF EXISTS my_policy ON my_table RESTRICT")
	require.Len(t, stmts, 1)

	dropPolicy, ok := stmts[0].(*statement.DropPolicy)
	require.True(t, ok, "Expected DropPolicy statement, got %T", stmts[0])
	assert.True(t, dropPolicy.IfExists)
	assert.Equal(t, "my_policy", dropPolicy.Name.String())
	assert.Equal(t, "my_table", dropPolicy.TableName.String())

	// Test without IF EXISTS
	dialects.VerifiedStmt(t, "DROP POLICY my_policy ON my_table CASCADE")

	// Test minimal DROP POLICY
	dialects.VerifiedStmt(t, "DROP POLICY my_policy ON my_table")

	// Test error - missing table name
	_, err := parser.ParseSQL(utils.NewTestedDialects().Dialects[0], "DROP POLICY my_policy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: ON")
}

// TestDropConnector verifies DROP CONNECTOR statement parsing.
// Reference: tests/sqlparser_common.rs:14497
func TestDropConnector(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test DROP CONNECTOR with IF EXISTS
	stmts := dialects.ParseSQL(t, "DROP CONNECTOR IF EXISTS my_connector")
	require.Len(t, stmts, 1)

	dropConnector, ok := stmts[0].(*statement.DropConnector)
	require.True(t, ok, "Expected DropConnector statement, got %T", stmts[0])
	assert.True(t, dropConnector.IfExists)
	assert.Equal(t, "my_connector", dropConnector.Name.String())

	// Test without IF EXISTS
	dialects.VerifiedStmt(t, "DROP CONNECTOR my_connector")

	// Test error - missing connector name
	_, err := parser.ParseSQL(utils.NewTestedDialects().Dialects[0], "DROP CONNECTOR")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: identifier")
}

// TestParseDropStream verifies DROP STREAM statement parsing.
// Reference: tests/sqlparser_common.rs:17719
func TestParseDropStream(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic DROP STREAM
	stmts := dialects.ParseSQL(t, "DROP STREAM s1")
	require.Len(t, stmts, 1)
	drop, ok := stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])
	assert.Equal(t, "STREAM", drop.ObjectType.String())
	require.Len(t, drop.Names, 1)
	assert.Equal(t, "s1", drop.Names[0].String())
	assert.False(t, drop.IfExists)

	// IF EXISTS
	stmts = dialects.ParseSQL(t, "DROP STREAM IF EXISTS s1")
	require.Len(t, stmts, 1)
	drop, ok = stmts[0].(*statement.Drop)
	require.True(t, ok, "Expected Drop statement, got %T", stmts[0])
	assert.True(t, drop.IfExists)
}
