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
// This file contains tests 221-240 from the Rust test file.
package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/ast/statement"
	sqlparserDialects "github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestDictionarySyntax verifies dictionary literal syntax parsing.
// Reference: tests/sqlparser_common.rs:13596
func TestDictionarySyntax(t *testing.T) {
	// Test empty dictionary
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsDictionarySyntax()
	})
	dialects.VerifiedExpr(t, "{}")

	// Test dictionary with string values
	dialects.VerifiedExpr(t, "{'Alberta': 'Edmonton', 'Manitoba': 'Winnipeg'}")

	// Test dictionary with CAST values
	dialects.VerifiedExpr(t, "{'start': CAST('2023-04-01' AS TIMESTAMP), 'end': CAST('2023-04-05' AS TIMESTAMP)}")
}

// TestMapSyntax verifies MAP literal syntax parsing.
// Reference: tests/sqlparser_common.rs:13656
func TestMapSyntax(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsMapLiteralSyntax()
	})

	// Test MAP with string key-value pairs
	dialects.VerifiedExpr(t, "MAP {'Alberta': 'Edmonton', 'Manitoba': 'Winnipeg'}")

	// Test MAP with numeric key-value pairs
	dialects.VerifiedExpr(t, "MAP {1: 10.0, 2: 20.0}")

	// Test MAP with array keys
	dialects.VerifiedExpr(t, "MAP {[1, 2, 3]: 10.0, [4, 5, 6]: 20.0}")

	// Test MAP with subscript access
	dialects.VerifiedExpr(t, "MAP {'a': 10, 'b': 20}['a']")

	// Test empty MAP
	dialects.VerifiedExpr(t, "MAP {}")

	// Test MAP with NULL values
	dialects.VerifiedExpr(t, "MAP {'a': 1, 'b': NULL}")

	// Test MAP with array values
	dialects.VerifiedExpr(t, "MAP {1: [1, NULL, 3], 2: [4, NULL, 6], 3: [7, 8, 9]}")
}

// TestGroupByNothing verifies GROUP BY () syntax parsing.
// Reference: tests/sqlparser_common.rs:13899
func TestGroupByNothing(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsGroupByExpr()
	})

	// Test GROUP BY () alone
	stmts := dialects.ParseSQL(t, "SELECT count(1) FROM t GROUP BY ()")
	require.Len(t, stmts, 1)

	// Test GROUP BY with other expressions and ()
	stmts = dialects.ParseSQL(t, "SELECT name, count(1) FROM t GROUP BY name, ()")
	require.Len(t, stmts, 1)
}

// TestExtractSecondsOk verifies EXTRACT with SECONDS from INTERVAL.
// Reference: tests/sqlparser_common.rs:13926
func TestExtractSecondsOk(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.AllowExtractCustom()
	})

	// Test EXTRACT with SECONDS
	dialects.VerifiedExpr(t, "EXTRACT(SECONDS FROM '2 seconds'::INTERVAL)")

	// Test in SELECT context
	dialects.VerifiedStmt(t, "SELECT EXTRACT(seconds FROM '2 seconds'::INTERVAL)")
}

// TestExtractSecondsSingleQuoteOk verifies EXTRACT with quoted 'seconds' field.
// Reference: tests/sqlparser_common.rs:14011
func TestExtractSecondsSingleQuoteOk(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.AllowExtractCustom()
	})

	// Test EXTRACT with single-quoted 'seconds'
	dialects.VerifiedExpr(t, "EXTRACT('seconds' FROM '2 seconds'::INTERVAL)")
}

// TestExtractSecondsSingleQuoteErr verifies EXTRACT with quoted field errors on unsupported dialects.
// Reference: tests/sqlparser_common.rs:14041
func TestExtractSecondsSingleQuoteErr(t *testing.T) {
	filteredDialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.AllowExtractSingleQuotes()
	})

	sql := "SELECT EXTRACT('seconds' FROM '2 seconds'::INTERVAL)"
	_, err := parser.ParseSQL(filteredDialects.Dialects[0], sql)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: date/time field")
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

// TestCreatePolicy verifies CREATE POLICY statement parsing.
// Reference: tests/sqlparser_common.rs:14210
func TestCreatePolicy(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE POLICY my_policy ON my_table AS PERMISSIVE FOR SELECT TO my_role, CURRENT_USER USING (c0 = 1) WITH CHECK (1 = 1)"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	policy, ok := stmts[0].(*statement.CreatePolicy)
	require.True(t, ok, "Expected CreatePolicy statement, got %T", stmts[0])
	assert.Equal(t, "my_policy", policy.Name.String())
	assert.Equal(t, "my_table", policy.TableName.String())

	// Test with SELECT subquery in USING
	dialects.VerifiedStmt(t, "CREATE POLICY my_policy ON my_table AS PERMISSIVE FOR SELECT TO my_role, CURRENT_USER USING (c0 IN (SELECT column FROM t0)) WITH CHECK (1 = 1)")

	// Test minimal CREATE POLICY
	dialects.VerifiedStmt(t, "CREATE POLICY my_policy ON my_table")

	// Test error - missing table name
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "CREATE POLICY my_policy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: ON")
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
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "DROP POLICY my_policy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: ON")
}

// TestAlterPolicy verifies ALTER POLICY statement parsing.
// Reference: tests/sqlparser_common.rs:14364
func TestAlterPolicy(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test ALTER POLICY RENAME
	stmts := dialects.ParseSQL(t, "ALTER POLICY old_policy ON my_table RENAME TO new_policy")
	require.Len(t, stmts, 1)

	alterPolicy, ok := stmts[0].(*statement.AlterPolicy)
	require.True(t, ok, "Expected AlterPolicy statement, got %T", stmts[0])
	assert.Equal(t, "old_policy", alterPolicy.Name.String())
	assert.Equal(t, "my_table", alterPolicy.TableName.String())

	// Test ALTER POLICY with TO, USING, WITH CHECK
	dialects.VerifiedStmt(t, "ALTER POLICY my_policy ON my_table TO CURRENT_USER USING ((SELECT c0)) WITH CHECK (c0 > 0)")

	// Test minimal ALTER POLICY
	dialects.VerifiedStmt(t, "ALTER POLICY my_policy ON my_table")

	// Test error - mixing RENAME with other clauses
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "ALTER POLICY old_policy ON my_table TO public RENAME TO new_policy")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: end of statement")
}

// TestCreateConnector verifies CREATE CONNECTOR statement parsing.
// Reference: tests/sqlparser_common.rs:14445
func TestCreateConnector(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql := "CREATE CONNECTOR my_connector TYPE 'jdbc' URL 'jdbc:mysql://localhost:3306/mydb' WITH DCPROPERTIES('user' = 'root', 'password' = 'password')"
	stmts := dialects.ParseSQL(t, sql)
	require.Len(t, stmts, 1)

	connector, ok := stmts[0].(*statement.CreateConnector)
	require.True(t, ok, "Expected CreateConnector statement, got %T", stmts[0])
	assert.Equal(t, "my_connector", connector.Name.String())

	// Test minimal CREATE CONNECTOR
	dialects.VerifiedStmt(t, "CREATE CONNECTOR my_connector")

	// Test error - missing connector name
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "CREATE CONNECTOR")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: identifier")
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
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "DROP CONNECTOR")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: identifier")
}

// TestAlterConnector verifies ALTER CONNECTOR statement parsing.
// Reference: tests/sqlparser_common.rs:14521
func TestAlterConnector(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test ALTER CONNECTOR SET DCPROPERTIES
	stmts := dialects.ParseSQL(t, "ALTER CONNECTOR my_connector SET DCPROPERTIES('user' = 'root', 'password' = 'password')")
	require.Len(t, stmts, 1)

	alterConnector, ok := stmts[0].(*statement.AlterConnector)
	require.True(t, ok, "Expected AlterConnector statement, got %T", stmts[0])
	assert.Equal(t, "my_connector", alterConnector.Name.String())

	// Test ALTER CONNECTOR SET URL
	dialects.VerifiedStmt(t, "ALTER CONNECTOR my_connector SET URL 'jdbc:mysql://localhost:3306/mydb'")

	// Test ALTER CONNECTOR SET OWNER USER
	dialects.VerifiedStmt(t, "ALTER CONNECTOR my_connector SET OWNER USER 'root'")

	// Test ALTER CONNECTOR SET OWNER ROLE
	dialects.VerifiedStmt(t, "ALTER CONNECTOR my_connector SET OWNER ROLE 'admin'")

	// Test error - wrong option name
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "ALTER CONNECTOR my_connector SET WRONG 'jdbc:mysql://localhost:3306/mydb'")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Expected: end of statement")
}

// TestSelectWhereWithLikeOrIlikeAny verifies SELECT with LIKE/ILIKE ANY.
// Reference: tests/sqlparser_common.rs:14622
func TestSelectWhereWithLikeOrIlikeAny(t *testing.T) {
	dialects := utils.NewTestedDialects()

	dialects.VerifiedStmt(t, "SELECT * FROM x WHERE a ILIKE ANY '%abc%'")
	dialects.VerifiedStmt(t, "SELECT * FROM x WHERE a LIKE ANY '%abc%'")
	dialects.VerifiedStmt(t, "SELECT * FROM x WHERE a ILIKE ANY ('%Jo%oe%', 'T%e')")
	dialects.VerifiedStmt(t, "SELECT * FROM x WHERE a LIKE ANY ('%Jo%oe%', 'T%e')")
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

// TestAliasEqualExpr verifies alias assignment with = syntax.
// Reference: tests/sqlparser_common.rs:14638
func TestAliasEqualExpr(t *testing.T) {
	// Test with dialects that support = alias assignment
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsEqAliasAssignment()
	})

	dialects.OneStatementParsesTo(t, "SELECT some_alias = some_column FROM some_table", "SELECT some_column AS some_alias FROM some_table")
	dialects.OneStatementParsesTo(t, "SELECT some_alias = (a*b) FROM some_table", "SELECT (a * b) AS some_alias FROM some_table")

	// Test with dialects that don't support = alias assignment
	dialectsNoSupport := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsEqAliasAssignment()
	})

	dialectsNoSupport.OneStatementParsesTo(t, "SELECT x = (a * b) FROM some_table", "SELECT x = (a * b) FROM some_table")
}

// TestTryConvert verifies TRY_CONVERT function parsing.
// Reference: tests/sqlparser_common.rs:14655
func TestTryConvert(t *testing.T) {
	// Dialects with type before value
	dialects1 := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTryConvert() && d.ConvertTypeBeforeValue()
	})
	dialects1.VerifiedExpr(t, "TRY_CONVERT(VARCHAR(MAX), 'foo')")

	// Dialects with value before type
	dialects2 := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTryConvert() && !d.ConvertTypeBeforeValue()
	})
	dialects2.VerifiedExpr(t, "TRY_CONVERT('foo', VARCHAR(MAX))")
}

// TestShowDbsSchemasTablesViews verifies SHOW statements.
// Reference: tests/sqlparser_common.rs:14744
func TestShowDbsSchemasTablesViews(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic SHOW statements
	stmts := []string{
		"SHOW DATABASES",
		"SHOW SCHEMAS",
		"SHOW TABLES",
		"SHOW VIEWS",
		"SHOW TABLES IN db1",
		"SHOW VIEWS FROM db1",
		"SHOW MATERIALIZED VIEWS",
		"SHOW MATERIALIZED VIEWS IN db1",
		"SHOW MATERIALIZED VIEWS FROM db1",
	}
	for _, sql := range stmts {
		dialects.VerifiedStmt(t, sql)
	}

	// SHOW with LIKE (dialect-dependent)
	likeStmts := []string{
		"SHOW DATABASES LIKE '%abc'",
		"SHOW SCHEMAS LIKE '%abc'",
	}
	for _, sql := range likeStmts {
		// Dialects that support LIKE before IN
		dialectsLikeBeforeIn := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
			return d.SupportsShowLikeBeforeIn()
		})
		dialectsLikeBeforeIn.VerifiedStmt(t, sql)

		// Dialects that don't support LIKE before IN
		dialectsNoLikeBeforeIn := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
			return !d.SupportsShowLikeBeforeIn()
		})
		dialectsNoLikeBeforeIn.VerifiedStmt(t, sql)
	}

	// SHOW with LIKE in suffix (only for dialects that support it)
	suffixLikeStmts := []string{
		"SHOW TABLES IN db1 'abc'",
		"SHOW VIEWS IN db1 'abc'",
		"SHOW VIEWS FROM db1 'abc'",
		"SHOW MATERIALIZED VIEWS IN db1 'abc'",
		"SHOW MATERIALIZED VIEWS FROM db1 'abc'",
	}
	for _, sql := range suffixLikeStmts {
		dialectsNoLikeBeforeIn := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
			return !d.SupportsShowLikeBeforeIn()
		})
		dialectsNoLikeBeforeIn.VerifiedStmt(t, sql)
	}
}

// TestLoadExtension verifies LOAD extension statement parsing.
// Reference: tests/sqlparser_common.rs:15101
func TestLoadExtension(t *testing.T) {
	// Dialects that support LOAD extension
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsLoadExtension()
	})

	stmts := dialects.ParseSQL(t, "LOAD my_extension")
	require.Len(t, stmts, 1)

	load, ok := stmts[0].(*statement.Load)
	require.True(t, ok, "Expected Load statement, got %T", stmts[0])
	assert.Equal(t, "my_extension", load.ExtensionName.String())

	// Test with quoted extension name
	stmts = dialects.ParseSQL(t, "LOAD 'filename'")
	require.Len(t, stmts, 1)

	load, ok = stmts[0].(*statement.Load)
	require.True(t, ok, "Expected Load statement, got %T", stmts[0])
	assert.Equal(t, "filename", load.ExtensionName.String())

	// Dialects that don't support LOAD extension should error
	dialectsNoSupport := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return !d.SupportsLoadExtension()
	})

	if len(dialectsNoSupport.Dialects) > 0 {
		_, err := parser.ParseSQL(dialectsNoSupport.Dialects[0], "LOAD my_extension")
		require.Error(t, err)
	}
}

// TestSelectTop verifies SELECT TOP clause parsing.
// Reference: tests/sqlparser_common.rs:15140
func TestSelectTop(t *testing.T) {
	dialects := utils.NewTestedDialectsWithFilter(func(d sqlparserDialects.Dialect) bool {
		return d.SupportsTopBeforeDistinct()
	})

	dialects.VerifiedStmt(t, "SELECT ALL * FROM tbl")
	dialects.VerifiedStmt(t, "SELECT TOP 3 * FROM tbl")
	dialects.VerifiedStmt(t, "SELECT TOP 3 ALL * FROM tbl")
	dialects.VerifiedStmt(t, "SELECT TOP 3 DISTINCT * FROM tbl")
	dialects.VerifiedStmt(t, "SELECT TOP 3 DISTINCT a, b, c FROM tbl")
}
