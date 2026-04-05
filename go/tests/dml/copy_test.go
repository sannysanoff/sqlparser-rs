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
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/tests/utils"
)

// TestParseCopyOptions verifies COPY statement options parsing.
// Reference: tests/sqlparser_common.rs:17890
func TestParseCopyOptions(t *testing.T) {
	sql1 := "COPY dst (c1, c2, c3) FROM 's3://bucket/file.txt' IAM_ROLE 'arn:aws:iam::123:role/r' CSV IGNOREHEADER 1"
	stmt1 := utils.NewTestedDialects().VerifiedStmt(t, sql1)

	copy1, ok := stmt1.(*statement.Copy)
	require.True(t, ok, "Expected Copy statement, got %T", stmt1)
	require.NotNil(t, copy1.Source, "Expected source")
	assert.Equal(t, "dst", copy1.Source.String())

	// Verify IAM_ROLE and other options are parsed - LegacyOptions is where COPY options are stored
	assert.True(t, len(copy1.LegacyOptions) > 0 || len(copy1.Options) > 0, "Expected copy options")
}

// TestParseCopyOptionsRedshift verifies Redshift COPY options specifically.
// Reference: tests/sqlparser_common.rs:17890 (additional tests)
func TestParseCopyOptionsRedshift(t *testing.T) {
	dialects := utils.NewTestedDialects()

	// Test various COPY options
	tests := []string{
		"COPY dst FROM 's3://bucket/file.txt' IAM_ROLE DEFAULT CSV",
		"COPY dst FROM 's3://bucket/file.txt' CSV IGNOREHEADER 1",
		"COPY dst FROM 's3://bucket/file.txt' JSON 'auto'",
		"COPY dst FROM 's3://bucket/file.txt' GZIP",
	}

	for _, sql := range tests {
		t.Run(sql, func(t *testing.T) {
			stmts := dialects.ParseSQL(t, sql)
			require.Len(t, stmts, 1)
		})
	}
}
