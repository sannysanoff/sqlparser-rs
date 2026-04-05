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

// Package tests contains the SQL parsing tests.
// These tests are ported from tests/sqlparser_common.rs in the Rust implementation.
// This file contains GRANT/DENY/REVOKE tests.
package tests

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseGrant verifies GRANT statement parsing.
// Reference: tests/sqlparser_common.rs:9678
func TestParseGrant(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Comprehensive GRANT statement
	sql := "GRANT SELECT, INSERT, UPDATE (shape, size), USAGE, DELETE, TRUNCATE, REFERENCES, TRIGGER, CONNECT, CREATE, EXECUTE, TEMPORARY, DROP ON abc, def TO xyz, m WITH GRANT OPTION GRANTED BY jj"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	grant, ok := stmts[0].(*statement.Grant)
	require.True(t, ok, "Expected Grant statement, got %T", stmts[0])

	require.NotNil(t, grant.Privileges, "Expected privileges to be present")
	require.NotNil(t, grant.Objects, "Expected objects to be present")
	assert.Equal(t, 2, len(grant.Grantees), "Expected 2 grantees")
	assert.True(t, grant.WithGrantOption, "Expected with_grant_option to be true")
	require.NotNil(t, grant.GrantedBy, "Expected granted_by to be present")
	assert.Equal(t, "jj", grant.GrantedBy.String())

	// GRANT on all tables in schema
	sql2 := "GRANT INSERT ON ALL TABLES IN SCHEMA public TO browser"
	stmts = dialects.ParseSQL(t, sql2)
	require.Len(t, stmts, 1)

	grant, ok = stmts[0].(*statement.Grant)
	require.True(t, ok, "Expected Grant statement, got %T", stmts[0])
	require.NotNil(t, grant.Privileges, "Expected privileges to be present")
	require.NotNil(t, grant.Objects, "Expected objects to be present")
	assert.Equal(t, 1, len(grant.Grantees), "Expected 1 grantee")
	assert.Equal(t, "browser", grant.Grantees[0].String())
	assert.False(t, grant.WithGrantOption, "Expected with_grant_option to be false")

	// GRANT on sequences
	sql3 := "GRANT USAGE, SELECT ON SEQUENCE p TO u"
	stmts = dialects.ParseSQL(t, sql3)
	require.Len(t, stmts, 1)

	grant, ok = stmts[0].(*statement.Grant)
	require.True(t, ok, "Expected Grant statement, got %T", stmts[0])
	require.NotNil(t, grant.Privileges, "Expected privileges to be present")
	require.NotNil(t, grant.Objects, "Expected objects to be present")
	assert.Equal(t, 1, len(grant.Grantees), "Expected 1 grantee")
	assert.Nil(t, grant.GrantedBy, "Expected granted_by to be nil")

	// GRANT ALL PRIVILEGES
	sql4 := "GRANT ALL PRIVILEGES ON aa, b TO z"
	stmts = dialects.ParseSQL(t, sql4)
	require.Len(t, stmts, 1)

	grant, ok = stmts[0].(*statement.Grant)
	require.True(t, ok, "Expected Grant statement, got %T", stmts[0])
	require.NotNil(t, grant.Privileges, "Expected privileges to be present")

	// GRANT ALL on schemas
	sql5 := "GRANT ALL ON SCHEMA aa, b TO z"
	dialects.VerifiedStmt(t, sql5)

	// GRANT on all sequences in schema
	sql6 := "GRANT USAGE ON ALL SEQUENCES IN SCHEMA bus TO a, beta WITH GRANT OPTION"
	dialects.VerifiedStmt(t, sql6)

	// Additional GRANT statements from the test
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL TABLES IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL TABLES IN SCHEMA db1.sc1 TO ROLE role1 WITH GRANT OPTION")
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL TABLES IN SCHEMA db1.sc1 TO DATABASE ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL TABLES IN SCHEMA db1.sc1 TO APPLICATION role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL TABLES IN SCHEMA db1.sc1 TO APPLICATION ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL TABLES IN SCHEMA db1.sc1 TO SHARE share1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL VIEWS IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL MATERIALIZED VIEWS IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON ALL EXTERNAL TABLES IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT USAGE ON ALL FUNCTIONS IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT USAGE ON SCHEMA sc1 TO a:b")
	dialects.VerifiedStmt(t, "GRANT USAGE ON SCHEMA sc1 TO GROUP group1")
	dialects.VerifiedStmt(t, "GRANT OWNERSHIP ON ALL TABLES IN SCHEMA DEV_STAS_ROGOZHIN TO ROLE ANALYST")
	dialects.VerifiedStmt(t, "GRANT OWNERSHIP ON ALL TABLES IN SCHEMA DEV_STAS_ROGOZHIN TO ROLE ANALYST COPY CURRENT GRANTS")
	dialects.VerifiedStmt(t, "GRANT OWNERSHIP ON ALL TABLES IN SCHEMA DEV_STAS_ROGOZHIN TO ROLE ANALYST REVOKE CURRENT GRANTS")
	dialects.VerifiedStmt(t, "GRANT USAGE ON DATABASE db1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT USAGE ON WAREHOUSE wh1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT OWNERSHIP ON INTEGRATION int1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON VIEW view1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT EXEC ON my_sp TO runner")
	dialects.VerifiedStmt(t, "GRANT UPDATE ON my_table TO updater_role AS dbo")
	dialects.VerifiedStmt(t, "GRANT SELECT ON FUTURE SCHEMAS IN DATABASE db1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON FUTURE TABLES IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON FUTURE EXTERNAL TABLES IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON FUTURE VIEWS IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON FUTURE MATERIALIZED VIEWS IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT SELECT ON FUTURE SEQUENCES IN SCHEMA db1.sc1 TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT USAGE ON PROCEDURE db1.sc1.foo(INT) TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT USAGE ON FUNCTION db1.sc1.foo(INT) TO ROLE role1")
	dialects.VerifiedStmt(t, "GRANT ROLE role1 TO ROLE role2")
	dialects.VerifiedStmt(t, "GRANT ROLE role1 TO USER user")
	dialects.VerifiedStmt(t, "GRANT CREATE SCHEMA ON DATABASE db1 TO ROLE role1")
}

// TestParseDeny verifies DENY statement parsing (MSSQL).
// Reference: tests/sqlparser_common.rs:9863
func TestParseDeny(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "DENY INSERT, DELETE ON users TO analyst CASCADE AS admin"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	deny, ok := stmts[0].(*statement.DenyStatement)
	require.True(t, ok, "Expected DenyStatement, got %T", stmts[0])

	require.NotNil(t, deny.Privileges, "Expected privileges to be present")
	require.NotNil(t, deny.Objects, "Expected objects to be present")
	assert.Equal(t, 1, len(deny.Grantees), "Expected 1 grantee")
	assert.Equal(t, "analyst", deny.Grantees[0].String())

	// Additional DENY statements
	dialects.VerifiedStmt(t, "DENY SELECT, INSERT, UPDATE, DELETE ON db1.sc1 TO role1, role2")
	dialects.VerifiedStmt(t, "DENY ALL ON db1.sc1 TO role1")
	dialects.VerifiedStmt(t, "DENY EXEC ON my_sp TO runner")
}

// TestRevoke verifies REVOKE statement parsing.
// Reference: tests/sqlparser_common.rs:9891
func TestRevoke(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "REVOKE ALL PRIVILEGES ON users, auth FROM analyst"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	revoke, ok := stmts[0].(*statement.Revoke)
	require.True(t, ok, "Expected Revoke statement, got %T", stmts[0])

	require.NotNil(t, revoke.Privileges, "Expected privileges to be present")
	require.NotNil(t, revoke.Objects, "Expected objects to be present")
	assert.Equal(t, 1, len(revoke.Grantees), "Expected 1 grantee")
	assert.Equal(t, "analyst", revoke.Grantees[0].String())
	assert.False(t, revoke.Cascade, "Expected cascade to be false")
	assert.Nil(t, revoke.GrantedBy, "Expected granted_by to be nil")
}

// TestRevokeWithCascade verifies REVOKE with CASCADE option.
// Reference: tests/sqlparser_common.rs:9917
func TestRevokeWithCascade(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		// Exclude MySQL dialect which doesn't support CASCADE in REVOKE
		_, isMySQL := d.(*mysql.MySqlDialect)
		return !isMySQL
	})

	sql := "REVOKE ALL PRIVILEGES ON users, auth FROM analyst CASCADE"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	revoke, ok := stmts[0].(*statement.Revoke)
	require.True(t, ok, "Expected Revoke statement, got %T", stmts[0])

	require.NotNil(t, revoke.Privileges, "Expected privileges to be present")
	require.NotNil(t, revoke.Objects, "Expected objects to be present")
	assert.Equal(t, 1, len(revoke.Grantees), "Expected 1 grantee")
	assert.Equal(t, "analyst", revoke.Grantees[0].String())
	assert.True(t, revoke.Cascade, "Expected cascade to be true")
	assert.Nil(t, revoke.GrantedBy, "Expected granted_by to be nil")
}
