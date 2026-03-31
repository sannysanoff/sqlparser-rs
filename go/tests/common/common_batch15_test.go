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
// This file contains tests 261-280 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseRollback verifies ROLLBACK statement parsing.
// Reference: tests/sqlparser_common.rs:9388
func TestParseRollback(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic ROLLBACK
	stmt := dialects.VerifiedStmt(t, "ROLLBACK")
	rollback, ok := stmt.(*statement.Rollback)
	require.True(t, ok, "Expected Rollback statement, got %T", stmt)
	assert.False(t, rollback.Chain, "Expected chain to be false")
	assert.Nil(t, rollback.Savepoint, "Expected savepoint to be nil")

	// ROLLBACK AND CHAIN
	stmt = dialects.VerifiedStmt(t, "ROLLBACK AND CHAIN")
	rollback, ok = stmt.(*statement.Rollback)
	require.True(t, ok, "Expected Rollback statement, got %T", stmt)
	assert.True(t, rollback.Chain, "Expected chain to be true")
	assert.Nil(t, rollback.Savepoint, "Expected savepoint to be nil")

	// ROLLBACK TO SAVEPOINT
	stmt = dialects.VerifiedStmt(t, "ROLLBACK TO SAVEPOINT test1")
	rollback, ok = stmt.(*statement.Rollback)
	require.True(t, ok, "Expected Rollback statement, got %T", stmt)
	assert.False(t, rollback.Chain, "Expected chain to be false")
	require.NotNil(t, rollback.Savepoint, "Expected savepoint to be present")
	assert.Equal(t, "test1", rollback.Savepoint.String())

	// ROLLBACK AND CHAIN TO SAVEPOINT
	stmt = dialects.VerifiedStmt(t, "ROLLBACK AND CHAIN TO SAVEPOINT test1")
	rollback, ok = stmt.(*statement.Rollback)
	require.True(t, ok, "Expected Rollback statement, got %T", stmt)
	assert.True(t, rollback.Chain, "Expected chain to be true")
	require.NotNil(t, rollback.Savepoint, "Expected savepoint to be present")
	assert.Equal(t, "test1", rollback.Savepoint.String())

	// Test one_statement_parses_to equivalents
	dialects.OneStatementParsesTo(t, "ROLLBACK AND NO CHAIN", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK WORK AND NO CHAIN", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK TRANSACTION AND NO CHAIN", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK WORK AND CHAIN", "ROLLBACK AND CHAIN")
	dialects.OneStatementParsesTo(t, "ROLLBACK TRANSACTION AND CHAIN", "ROLLBACK AND CHAIN")
	dialects.OneStatementParsesTo(t, "ROLLBACK WORK", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK TRANSACTION", "ROLLBACK")
	dialects.OneStatementParsesTo(t, "ROLLBACK TO test1", "ROLLBACK TO SAVEPOINT test1")
	dialects.OneStatementParsesTo(t, "ROLLBACK AND CHAIN TO test1", "ROLLBACK AND CHAIN TO SAVEPOINT test1")
}

// TestParseCreateIndex verifies CREATE INDEX statement parsing.
// Reference: tests/sqlparser_common.rs:9449
func TestParseCreateIndex(t *testing.T) {
	sql := "CREATE UNIQUE INDEX IF NOT EXISTS idx_name ON test(name, age DESC)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createIndex, ok := stmts[0].(*statement.CreateIndex)
	require.True(t, ok, "Expected CreateIndex statement, got %T", stmts[0])

	require.NotNil(t, createIndex.Name, "Expected index name to be present")
	assert.Equal(t, "idx_name", createIndex.Name.String())
	assert.Equal(t, "test", createIndex.TableName.String())
	assert.True(t, createIndex.Unique, "Expected unique to be true")
	assert.True(t, createIndex.IfNotExists, "Expected if_not_exists to be true")
	assert.Equal(t, 2, len(createIndex.Columns), "Expected 2 indexed columns")

	// First column: name (no order specified)
	require.NotNil(t, createIndex.Columns[0].Expr, "Expected name for first column")
	assert.Equal(t, "name", createIndex.Columns[0].Expr.String())

	// Second column: age DESC
	require.NotNil(t, createIndex.Columns[1].Expr, "Expected name for second column")
	assert.Equal(t, "age", createIndex.Columns[1].Expr.String())
}

// TestCreateIndexWithUsingFunction verifies CREATE INDEX with USING clause.
// Reference: tests/sqlparser_common.rs:9495
func TestCreateIndexWithUsingFunction(t *testing.T) {
	sql := "CREATE UNIQUE INDEX IF NOT EXISTS idx_name ON test USING BTREE (name, age DESC)"
	dialects := utils.NewTestedDialects()
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createIndex, ok := stmts[0].(*statement.CreateIndex)
	require.True(t, ok, "Expected CreateIndex statement, got %T", stmts[0])

	require.NotNil(t, createIndex.Name, "Expected index name to be present")
	assert.Equal(t, "idx_name", createIndex.Name.String())
	assert.Equal(t, "test", createIndex.TableName.String())
	assert.True(t, createIndex.Unique, "Expected unique to be true")
	assert.True(t, createIndex.IfNotExists, "Expected if_not_exists to be true")
	assert.False(t, createIndex.Concurrently, "Expected concurrently to be false")
	require.NotNil(t, createIndex.Using, "Expected using to be present")
	assert.Equal(t, "BTREE", createIndex.Using.String())
	assert.Equal(t, 2, len(createIndex.Columns), "Expected 2 indexed columns")
}

// TestCreateIndexWithWithClause verifies CREATE INDEX with WITH clause.
// Reference: tests/sqlparser_common.rs:9554
func TestCreateIndexWithWithClause(t *testing.T) {
	sql := "CREATE UNIQUE INDEX title_idx ON films(title) WITH (fillfactor = 70, single_param)"
	dialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		// Only test dialects that support CREATE INDEX with clause
		return true
	})
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createIndex, ok := stmts[0].(*statement.CreateIndex)
	require.True(t, ok, "Expected CreateIndex statement, got %T", stmts[0])

	require.NotNil(t, createIndex.Name, "Expected index name to be present")
	assert.Equal(t, "title_idx", createIndex.Name.String())
	assert.Equal(t, "films", createIndex.TableName.String())
	assert.True(t, createIndex.Unique, "Expected unique to be true")
	assert.False(t, createIndex.Concurrently, "Expected concurrently to be false")
	assert.False(t, createIndex.IfNotExists, "Expected if_not_exists to be false")
	require.NotNil(t, createIndex.With, "Expected with clause to be present")
	assert.Equal(t, 2, len(createIndex.With), "Expected 2 with parameters")
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

// TestParseCreateRole verifies CREATE ROLE statement parsing.
// Reference: tests/sqlparser_common.rs:9625
func TestParseCreateRole(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Single role
	sql := "CREATE ROLE consultant"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createRole, ok := stmts[0].(*statement.CreateRole)
	require.True(t, ok, "Expected CreateRole statement, got %T", stmts[0])
	assert.Equal(t, 1, len(createRole.Names), "Expected 1 role name")
	assert.Equal(t, "consultant", createRole.Names[0].String())
	assert.False(t, createRole.IfNotExists, "Expected if_not_exists to be false")

	// Multiple roles with IF NOT EXISTS
	sql = "CREATE ROLE IF NOT EXISTS mysql_a, mysql_b"
	stmts = dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	createRole, ok = stmts[0].(*statement.CreateRole)
	require.True(t, ok, "Expected CreateRole statement, got %T", stmts[0])
	assert.Equal(t, 2, len(createRole.Names), "Expected 2 role names")
	assert.Equal(t, "mysql_a", createRole.Names[0].String())
	assert.Equal(t, "mysql_b", createRole.Names[1].String())
	assert.True(t, createRole.IfNotExists, "Expected if_not_exists to be true")
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

// TestParseMerge verifies MERGE statement parsing.
// Reference: tests/sqlparser_common.rs:9943
func TestParseMerge(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "MERGE INTO s.bar AS dest USING (SELECT * FROM s.foo) AS stg ON dest.D = stg.D AND dest.E = stg.E WHEN NOT MATCHED THEN INSERT (A, B, C) VALUES (stg.A, stg.B, stg.C) WHEN MATCHED AND dest.A = 'a' THEN UPDATE SET dest.F = stg.F, dest.G = stg.G WHEN MATCHED THEN DELETE"
	sqlNoInto := "MERGE s.bar AS dest USING (SELECT * FROM s.foo) AS stg ON dest.D = stg.D AND dest.E = stg.E WHEN NOT MATCHED THEN INSERT (A, B, C) VALUES (stg.A, stg.B, stg.C) WHEN MATCHED AND dest.A = 'a' THEN UPDATE SET dest.F = stg.F, dest.G = stg.G WHEN MATCHED THEN DELETE"

	// Test both versions parse successfully
	dialects.VerifiedStmt(t, sql)
	dialects.VerifiedStmt(t, sqlNoInto)

	// MERGE with VALUES only
	sql2 := "MERGE INTO s.bar AS dest USING newArrivals AS S ON (1 > 1) WHEN NOT MATCHED THEN INSERT VALUES (stg.A, stg.B, stg.C)"
	dialects.VerifiedStmt(t, sql2)

	// MERGE with predicates
	sql3 := "MERGE INTO FOO USING FOO_IMPORT ON (FOO.ID = FOO_IMPORT.ID) WHEN MATCHED THEN UPDATE SET FOO.NAME = FOO_IMPORT.NAME WHERE 1 = 1 DELETE WHERE FOO.NAME LIKE '%.DELETE' WHEN NOT MATCHED THEN INSERT (ID, NAME) VALUES (FOO_IMPORT.ID, UPPER(FOO_IMPORT.NAME)) WHERE NOT FOO_IMPORT.NAME LIKE '%.DO_NOT_INSERT'"
	dialects.VerifiedStmt(t, sql3)

	// MERGE with simple insert columns
	sql4 := "MERGE INTO FOO USING FOO_IMPORT ON (FOO.ID = FOO_IMPORT.ID) WHEN NOT MATCHED THEN INSERT (ID, NAME) VALUES (1, 'abc')"
	dialects.VerifiedStmt(t, sql4)

	// MERGE with qualified insert columns
	sql5 := "MERGE INTO FOO USING FOO_IMPORT ON (FOO.ID = FOO_IMPORT.ID) WHEN NOT MATCHED THEN INSERT (FOO.ID, FOO.NAME) VALUES (1, 'abc')"
	dialects.VerifiedStmt(t, sql5)

	// MERGE with schema qualified insert columns
	sql6 := "MERGE INTO PLAYGROUND.FOO USING FOO_IMPORT ON (PLAYGROUND.FOO.ID = FOO_IMPORT.ID) WHEN NOT MATCHED THEN INSERT (PLAYGROUND.FOO.ID, PLAYGROUND.FOO.NAME) VALUES (1, 'abc')"
	dialects.VerifiedStmt(t, sql6)
}

// TestMergeInCte verifies MERGE in CTE.
// Reference: tests/sqlparser_common.rs:10206
func TestMergeInCte(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "WITH x AS (MERGE INTO t USING (VALUES (1)) ON 1 = 1 WHEN MATCHED THEN DELETE RETURNING *) SELECT * FROM x"
	dialects.VerifiedOnlySelect(t, sql)
}

// TestMergeWithReturning verifies MERGE with RETURNING clause.
// Reference: tests/sqlparser_common.rs:10217
func TestMergeWithReturning(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "MERGE INTO wines AS w USING wine_stock_changes AS s ON s.winename = w.winename WHEN NOT MATCHED AND s.stock_delta > 0 THEN INSERT VALUES (s.winename, s.stock_delta) WHEN MATCHED AND w.stock + s.stock_delta > 0 THEN UPDATE SET stock = w.stock + s.stock_delta WHEN MATCHED THEN DELETE RETURNING merge_action(), w.*"
	dialects.VerifiedStmt(t, sql)
}

// TestMergeWithOutput verifies MERGE with OUTPUT clause.
// Reference: tests/sqlparser_common.rs:10229
func TestMergeWithOutput(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "MERGE INTO target_table USING source_table ON target_table.id = source_table.oooid WHEN MATCHED THEN UPDATE SET target_table.description = source_table.description WHEN NOT MATCHED THEN INSERT (ID, description) VALUES (source_table.id, source_table.description) OUTPUT inserted.* INTO log_target"
	dialects.VerifiedStmt(t, sql)
}

// TestMergeWithOutputWithoutInto verifies MERGE with OUTPUT without INTO.
// Reference: tests/sqlparser_common.rs:10242
func TestMergeWithOutputWithoutInto(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "MERGE INTO a USING b ON a.id = b.id WHEN MATCHED THEN DELETE OUTPUT inserted.*"
	dialects.VerifiedStmt(t, sql)
}

// TestMergeIntoUsingTable verifies MERGE with simple table source.
// Reference: tests/sqlparser_common.rs:10250
func TestMergeIntoUsingTable(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "MERGE INTO target_table USING source_table ON target_table.id = source_table.oooid WHEN MATCHED THEN UPDATE SET target_table.description = source_table.description WHEN NOT MATCHED THEN INSERT (ID, description) VALUES (source_table.id, source_table.description)"
	dialects.VerifiedStmt(t, sql)
}

// TestMergeWithDelimiter verifies MERGE with trailing semicolon.
// Reference: tests/sqlparser_common.rs:10262
func TestMergeWithDelimiter(t *testing.T) {
	dialects := utils.NewTestedDialects()
	sql := "MERGE INTO target_table USING source_table ON target_table.id = source_table.oooid WHEN MATCHED THEN UPDATE SET target_table.description = source_table.description WHEN NOT MATCHED THEN INSERT (ID, description) VALUES (source_table.id, source_table.description);"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)
}

// TestMergeInvalidStatements verifies MERGE with invalid clauses produces errors.
// Reference: tests/sqlparser_common.rs:10277
func TestMergeInvalidStatements(t *testing.T) {
	dialects := utils.NewTestedDialects()

	testCases := []struct {
		sql    string
		errMsg string
	}{
		{
			sql:    "MERGE INTO T USING U ON TRUE WHEN NOT MATCHED THEN UPDATE SET a = b",
			errMsg: "UPDATE is not allowed in a NOT MATCHED merge clause",
		},
		{
			sql:    "MERGE INTO T USING U ON TRUE WHEN NOT MATCHED THEN DELETE",
			errMsg: "DELETE is not allowed in a NOT MATCHED merge clause",
		},
		{
			sql:    "MERGE INTO T USING U ON TRUE WHEN MATCHED THEN INSERT(a) VALUES(b)",
			errMsg: "INSERT is not allowed in a MATCHED merge clause",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.sql, func(t *testing.T) {
			_, err := utils.ParseSQLWithDialects(dialects.Dialects, tc.sql)
			require.Error(t, err, "Expected parse error for: %s", tc.sql)
			assert.Contains(t, err.Error(), tc.errMsg, "Expected error message to contain: %s", tc.errMsg)
		})
	}
}
