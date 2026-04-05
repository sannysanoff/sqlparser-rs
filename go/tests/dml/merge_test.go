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

// Package dml contains DML (Data Manipulation Language) SQL parsing tests.
// These tests cover INSERT, UPDATE, DELETE, MERGE, and COPY statements.
package dml

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/tests/utils"
)

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
