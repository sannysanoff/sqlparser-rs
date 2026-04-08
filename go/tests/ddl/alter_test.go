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
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseAlterTable verifies ALTER TABLE statement parsing.
// Reference: tests/sqlparser_common.rs: - this is a basic ALTER TABLE test
func TestParseAlterTable(t *testing.T) {
	dialects := utils.NewTestedDialects()

	sql1 := "ALTER TABLE customer ADD COLUMN email VARCHAR(255)"
	dialects.VerifiedStmt(t, sql1)

	sql2 := "ALTER TABLE customer DROP COLUMN email"
	dialects.VerifiedStmt(t, sql2)
}

// TestParseAlterTableBatch9 verifies ALTER TABLE operations parsing.
// Reference: tests/sqlparser_common.rs:4819
func TestParseAlterTableBatch9(t *testing.T) {
	d := utils.NewTestedDialects()

	// Test ADD COLUMN
	addColumn := "ALTER TABLE tab ADD COLUMN foo TEXT;"
	stmt := d.OneStatementParsesTo(t, addColumn, "ALTER TABLE tab ADD COLUMN foo TEXT")

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	assert.Equal(t, "tab", alterTable.Name.String())
	require.Len(t, alterTable.Operations, 1)

	// Test RENAME TABLE TO
	renameTable := "ALTER TABLE tab RENAME TO new_tab"
	stmt = d.VerifiedStmt(t, renameTable)
	alterTable, ok = stmt.(*statement.AlterTable)
	require.True(t, ok)
	assert.Equal(t, "tab", alterTable.Name.String())

	// Test RENAME TABLE AS
	renameTableAs := "ALTER TABLE tab RENAME AS new_tab"
	d.VerifiedStmt(t, renameTableAs)

	// Test RENAME COLUMN
	renameColumn := "ALTER TABLE tab RENAME COLUMN foo TO new_foo"
	d.VerifiedStmt(t, renameColumn)

	// Test SET TBLPROPERTIES
	setTblProperties := "ALTER TABLE tab SET TBLPROPERTIES('classification' = 'parquet')"
	d.VerifiedStmt(t, setTblProperties)

	// Test SET with parentheses
	setStorageParams := "ALTER TABLE tab SET (autovacuum_vacuum_scale_factor = 0.01, autovacuum_vacuum_threshold = 500)"
	d.VerifiedStmt(t, setStorageParams)
}

// TestAlterTableWithOnCluster verifies ALTER TABLE with ON CLUSTER parsing.
// Reference: tests/sqlparser_common.rs:4983
func TestAlterTableWithOnCluster(t *testing.T) {
	d := utils.NewTestedDialects()

	// Test with quoted cluster name
	sql := "ALTER TABLE t ON CLUSTER 'cluster' ADD CONSTRAINT bar PRIMARY KEY (baz)"
	stmt := d.VerifiedStmt(t, sql)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	assert.Equal(t, "t", alterTable.Name.String())

	// Test with unquoted cluster name
	sql2 := "ALTER TABLE t ON CLUSTER cluster_name ADD CONSTRAINT bar PRIMARY KEY (baz)"
	stmt = d.VerifiedStmt(t, sql2)

	alterTable, ok = stmt.(*statement.AlterTable)
	require.True(t, ok)
	assert.Equal(t, "t", alterTable.Name.String())

	// Test error: numeric cluster name
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "ALTER TABLE t ON CLUSTER 123 ADD CONSTRAINT bar PRIMARY KEY (baz)")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "identifier")
}

// TestParseAlterTableDropConstraint verifies ALTER TABLE DROP CONSTRAINT parsing.
// Reference: tests/sqlparser_common.rs:5291
func TestParseAlterTableDropConstraint(t *testing.T) {
	dialects := utils.NewTestedDialects()

	checkOne := func(t *testing.T, constraintText string) {
		sql := "ALTER TABLE tab " + constraintText
		stmts := dialects.ParseSQL(t, sql)
		require.Len(t, stmts, 1)

		alterTable, ok := stmts[0].(*statement.AlterTable)
		require.True(t, ok, "Expected AlterTable statement, got %T", stmts[0])

		assert.Equal(t, "tab", alterTable.Name.String())
		require.NotEmpty(t, alterTable.Operations, "Expected at least one operation")
	}

	checkOne(t, "DROP CONSTRAINT IF EXISTS constraint_name")
	checkOne(t, "DROP CONSTRAINT IF EXISTS constraint_name RESTRICT")
	checkOne(t, "DROP CONSTRAINT IF EXISTS constraint_name CASCADE")

	// Test parsing error for invalid syntax
	_, err := parser.ParseSQL(generic.NewGenericDialect(), "ALTER TABLE tab DROP CONSTRAINT is_active TEXT")
	require.Error(t, err)
}

// TestParseAlterIndex verifies ALTER INDEX statement parsing.
// Reference: tests/sqlparser_common.rs:5017
func TestParseAlterIndex(t *testing.T) {
	sql := "ALTER INDEX idx RENAME TO new_idx"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	alterIndex, ok := stmt.(*statement.AlterIndex)
	require.True(t, ok, "Expected AlterIndex statement")
	assert.Equal(t, "idx", alterIndex.Name.String())
	require.NotNil(t, alterIndex.Operation)
}

// TestParseAlterView verifies ALTER VIEW statement parsing.
// Reference: tests/sqlparser_common.rs:5032
func TestParseAlterView(t *testing.T) {
	sql := "ALTER VIEW myschema.myview AS SELECT foo FROM bar"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	alterView, ok := stmt.(*statement.AlterView)
	require.True(t, ok, "Expected AlterView statement")
	assert.Equal(t, "myschema.myview", alterView.Name.String())
	assert.Empty(t, alterView.Columns)
	require.NotNil(t, alterView.Query)
	assert.Empty(t, alterView.WithOptions)
}

// TestParseAlterViewWithOptions verifies ALTER VIEW with options parsing.
// Reference: tests/sqlparser_common.rs:5051
func TestParseAlterViewWithOptions(t *testing.T) {
	sql := "ALTER VIEW v WITH (foo = 'bar', a = 123) AS SELECT 1"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	alterView, ok := stmt.(*statement.AlterView)
	require.True(t, ok, "Expected AlterView statement")
	require.Len(t, alterView.WithOptions, 2)

	// Verify first option
	opt1 := alterView.WithOptions[0]
	assert.Equal(t, "foo", opt1.Name.String())

	// Verify second option
	opt2 := alterView.WithOptions[1]
	assert.Equal(t, "a", opt2.Name.String())
}

// TestParseAlterViewWithColumns verifies ALTER VIEW with columns parsing.
// Reference: tests/sqlparser_common.rs:5076
func TestParseAlterViewWithColumns(t *testing.T) {
	sql := "ALTER VIEW v (has, cols) AS SELECT 1, 2"
	stmt := utils.NewTestedDialects().VerifiedStmt(t, sql)

	alterView, ok := stmt.(*statement.AlterView)
	require.True(t, ok, "Expected AlterView statement")
	assert.Equal(t, "v", alterView.Name.String())
	require.Len(t, alterView.Columns, 2)
	assert.Equal(t, "has", alterView.Columns[0].String())
	assert.Equal(t, "cols", alterView.Columns[1].String())
	require.NotNil(t, alterView.Query)
	assert.Empty(t, alterView.WithOptions)
}

// TestParseAlterTableAddColumn verifies ALTER TABLE ADD COLUMN variations.
// Reference: tests/sqlparser_common.rs:5095
func TestParseAlterTableAddColumn(t *testing.T) {
	d := utils.NewTestedDialects()

	// Without COLUMN keyword
	sql1 := "ALTER TABLE tab ADD foo TEXT"
	stmt := d.VerifiedStmt(t, sql1)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	require.Len(t, alterTable.Operations, 1)

	// With COLUMN keyword
	sql2 := "ALTER TABLE tab ADD COLUMN foo TEXT"
	_ = d.VerifiedStmt(t, sql2)
}

// TestParseAlterTableAddColumnIfNotExists verifies ALTER TABLE ADD COLUMN IF NOT EXISTS.
// Reference: tests/sqlparser_common.rs:5112
func TestParseAlterTableAddColumnIfNotExists(t *testing.T) {
	// Test with specific dialects that support this feature
	filteredDialects := &utils.TestedDialects{
		Dialects: []dialects.Dialect{
			postgresql.NewPostgreSqlDialect(),
			generic.NewGenericDialect(),
			duckdb.NewDuckDbDialect(),
		},
	}

	// Without COLUMN keyword
	sql1 := "ALTER TABLE tab ADD IF NOT EXISTS foo TEXT"
	stmt := filteredDialects.VerifiedStmt(t, sql1)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	require.Len(t, alterTable.Operations, 1)

	// With COLUMN keyword
	sql2 := "ALTER TABLE tab ADD COLUMN IF NOT EXISTS foo TEXT"
	stmt = filteredDialects.VerifiedStmt(t, sql2)

	alterTable, ok = stmt.(*statement.AlterTable)
	require.True(t, ok)
	require.Len(t, alterTable.Operations, 1)
}

// TestParseAlterTableConstraints verifies ALTER TABLE ADD CONSTRAINT parsing.
// Reference: tests/sqlparser_common.rs:5143
func TestParseAlterTableConstraints(t *testing.T) {
	d := utils.NewTestedDialects()

	constraints := []string{
		"CONSTRAINT address_pkey PRIMARY KEY (address_id)",
		"CONSTRAINT uk_task UNIQUE (report_date, task_id)",
		"CONSTRAINT customer_address_id_fkey FOREIGN KEY (address_id) REFERENCES public.address(address_id)",
		"CONSTRAINT ck CHECK (rtrim(ltrim(REF_CODE)) <> '')",
		"PRIMARY KEY (foo, bar)",
		"UNIQUE (id)",
		"FOREIGN KEY (foo, bar) REFERENCES AnotherTable(foo, bar)",
		"CHECK (end_date > start_date OR end_date IS NULL)",
		"CONSTRAINT fk FOREIGN KEY (lng) REFERENCES othertable4",
	}

	for _, constraint := range constraints {
		sql := "ALTER TABLE tab ADD " + constraint
		stmt := d.VerifiedStmt(t, sql)

		alterTable, ok := stmt.(*statement.AlterTable)
		require.True(t, ok, "Expected AlterTable statement for: %s", constraint)
		require.Len(t, alterTable.Operations, 1)

		// Also verify CREATE TABLE with constraint
		createSql := "CREATE TABLE foo (id INT, " + constraint + ")"
		d.VerifiedStmt(t, createSql)
	}
}

// TestParseAlterTableDropColumn verifies ALTER TABLE DROP COLUMN parsing.
// Reference: tests/sqlparser_common.rs:5172
func TestParseAlterTableDropColumn(t *testing.T) {
	d := utils.NewTestedDialects()

	// Test various DROP COLUMN variations
	variations := []string{
		"DROP COLUMN IF EXISTS is_active",
		"DROP COLUMN IF EXISTS is_active CASCADE",
		"DROP COLUMN IF EXISTS is_active RESTRICT",
	}

	for _, variation := range variations {
		sql := "ALTER TABLE tab " + variation
		stmt := d.VerifiedStmt(t, sql)

		alterTable, ok := stmt.(*statement.AlterTable)
		require.True(t, ok, "Expected AlterTable statement for: %s", variation)
		require.Len(t, alterTable.Operations, 1)
	}

	// Test without COLUMN keyword with CASCADE
	d.OneStatementParsesTo(t,
		"ALTER TABLE tab DROP is_active CASCADE",
		"ALTER TABLE tab DROP is_active CASCADE")

	// Test comma-separated drop columns (dialects that support it)
	filteredDialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsCommaSeparatedDropColumnList()
	})
	if len(filteredDialects.Dialects) > 0 {
		filteredDialects.VerifiedStmt(t, "ALTER TABLE tbl DROP COLUMN c1, c2, c3")
	}
}

// TestParseAlterTableAlterColumn verifies ALTER TABLE ALTER COLUMN operations.
// Reference: tests/sqlparser_common.rs:5210
func TestParseAlterTableAlterColumn(t *testing.T) {
	d := utils.NewTestedDialects()

	// SET NOT NULL
	sql1 := "ALTER TABLE tab ALTER COLUMN is_active SET NOT NULL"
	stmt := d.VerifiedStmt(t, sql1)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	require.Len(t, alterTable.Operations, 1)

	// DROP NOT NULL (reserializes with COLUMN keyword)
	d.OneStatementParsesTo(t,
		"ALTER TABLE tab ALTER is_active DROP NOT NULL",
		"ALTER TABLE tab ALTER COLUMN is_active DROP NOT NULL")

	// SET DEFAULT
	sql3 := "ALTER TABLE tab ALTER COLUMN is_active SET DEFAULT 0"
	_ = d.VerifiedStmt(t, sql3)

	// DROP DEFAULT
	sql4 := "ALTER TABLE tab ALTER COLUMN is_active DROP DEFAULT"
	_ = d.VerifiedStmt(t, sql4)
}

// TestParseAlterTableAlterColumnType verifies ALTER TABLE ALTER COLUMN SET DATA TYPE.
// Reference: tests/sqlparser_common.rs:5254
func TestParseAlterTableAlterColumnType(t *testing.T) {
	d := utils.NewTestedDialects()

	// SET DATA TYPE
	sql := "ALTER TABLE tab ALTER COLUMN is_active SET DATA TYPE TEXT"
	stmt := d.VerifiedStmt(t, sql)

	alterTable, ok := stmt.(*statement.AlterTable)
	require.True(t, ok, "Expected AlterTable statement")
	require.Len(t, alterTable.Operations, 1)

	// TYPE (without SET)
	sql2 := "ALTER TABLE tab ALTER COLUMN is_active TYPE TEXT"
	d.VerifiedStmt(t, sql2)

	// SET DATA TYPE with USING (dialects that support it)
	filteredDialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return d.SupportsAlterColumnTypeUsing()
	})
	if len(filteredDialects.Dialects) > 0 {
		filteredDialects.VerifiedStmt(t,
			"ALTER TABLE tab ALTER COLUMN is_active SET DATA TYPE TEXT USING 'text'")
	}

	// Test that dialects without USING support reject it
	exceptDialects := utils.NewTestedDialectsWithFilter(func(d dialects.Dialect) bool {
		return !d.SupportsAlterColumnTypeUsing()
	})
	if len(exceptDialects.Dialects) > 0 {
		_, err := parser.ParseSQL(exceptDialects.Dialects[0],
			"ALTER TABLE tab ALTER COLUMN is_active SET DATA TYPE TEXT USING 'text'")
		require.Error(t, err)
	}
}

// TestParseAlterUser verifies ALTER USER statement parsing.
// Reference: tests/sqlparser_common.rs:18245
func TestParseAlterUser(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Basic ALTER USER
	dialects.VerifiedStmt(t, "ALTER USER u1")

	// ALTER USER IF EXISTS
	dialects.VerifiedStmt(t, "ALTER USER IF EXISTS u1")

	// ALTER USER RENAME TO
	stmt1 := dialects.VerifiedStmt(t, "ALTER USER IF EXISTS u1 RENAME TO u2")
	alterUser1, ok := stmt1.(*statement.AlterUser)
	require.True(t, ok, "Expected AlterUser statement, got %T", stmt1)
	assert.True(t, alterUser1.IfExists, "Expected IfExists to be true")
	assert.Equal(t, "u1", alterUser1.Name.String())
	require.NotNil(t, alterUser1.RenameTo, "Expected RenameTo")
	assert.Equal(t, "u2", alterUser1.RenameTo.String())

	// ALTER USER SET PASSWORD
	stmt2 := dialects.VerifiedStmt(t, "ALTER USER u1 PASSWORD 'AAA'")
	alterUser2, ok := stmt2.(*statement.AlterUser)
	require.True(t, ok, "Expected AlterUser statement, got %T", stmt2)
	assert.Equal(t, "u1", alterUser2.Name.String())

	// ALTER USER ENCRYPTED PASSWORD
	dialects.VerifiedStmt(t, "ALTER USER u1 ENCRYPTED PASSWORD 'AAA'")

	// ALTER USER PASSWORD NULL
	dialects.VerifiedStmt(t, "ALTER USER u1 PASSWORD NULL")

	// WITH PASSWORD should parse to canonical form
	dialects.OneStatementParsesTo(t, "ALTER USER u1 WITH PASSWORD 'AAA'", "ALTER USER u1 PASSWORD 'AAA'")
}

// TestParseAlterUserSetOptions verifies ALTER USER SET options.
// Reference: tests/sqlparser_common.rs:18245 (additional tests)
func TestParseAlterUserSetOptions(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// SET options with various types
	tests := []string{
		"ALTER USER u1 SET PASSWORD='secret'",
		"ALTER USER u1 SET DEFAULT_MFA_METHOD='PASSKEY'",
		"ALTER USER u1 SET TAG k1='v1'",
		"ALTER USER u1 SET DEFAULT_SECONDARY_ROLES=('ALL')",
		"ALTER USER u1 UNSET PASSWORD",
	}

	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, sql)
			require.Len(t, stmts, 1)
		})
	}
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
	// Go returns a descriptive error message rather than "Expected: end of statement"
	assert.Contains(t, err.Error(), "DCPROPERTIES")
}
