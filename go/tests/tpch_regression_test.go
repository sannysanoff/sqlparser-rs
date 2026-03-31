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

package regression

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
)

func TestTPCHQueries(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"tpch_1", "1.sql"},
		{"tpch_2", "2.sql"},
		{"tpch_3", "3.sql"},
		{"tpch_4", "4.sql"},
		{"tpch_5", "5.sql"},
		{"tpch_6", "6.sql"},
		{"tpch_7", "7.sql"},
		{"tpch_8", "8.sql"},
		{"tpch_9", "9.sql"},
		{"tpch_10", "10.sql"},
		{"tpch_11", "11.sql"},
		{"tpch_12", "12.sql"},
		{"tpch_13", "13.sql"},
		{"tpch_14", "14.sql"},
		{"tpch_15", "15.sql"},
		{"tpch_16", "16.sql"},
		{"tpch_17", "17.sql"},
		{"tpch_18", "18.sql"},
		{"tpch_19", "19.sql"},
		{"tpch_20", "20.sql"},
		{"tpch_21", "21.sql"},
		{"tpch_22", "22.sql"},
	}

	dialect := generic.NewGenericDialect()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := os.ReadFile(filepath.Join("fixtures", "tpch", tt.filename))
			require.NoError(t, err, "Failed to read %s", tt.filename)

			stmts, err := parser.ParseSQL(dialect, string(sql))
			assert.NoError(t, err, "Failed to parse %s", tt.filename)
			assert.NotEmpty(t, stmts, "Expected at least one statement from %s", tt.filename)
		})
	}
}

func TestTPCHQueriesRoundtrip(t *testing.T) {
	tests := []struct {
		name     string
		filename string
	}{
		{"tpch_1", "1.sql"},
		{"tpch_2", "2.sql"},
		{"tpch_3", "3.sql"},
		{"tpch_4", "4.sql"},
		{"tpch_5", "5.sql"},
		{"tpch_6", "6.sql"},
		{"tpch_7", "7.sql"},
		{"tpch_8", "8.sql"},
		{"tpch_9", "9.sql"},
		{"tpch_10", "10.sql"},
		{"tpch_11", "11.sql"},
		{"tpch_12", "12.sql"},
		{"tpch_13", "13.sql"},
		{"tpch_14", "14.sql"},
		{"tpch_15", "15.sql"},
		{"tpch_16", "16.sql"},
		{"tpch_17", "17.sql"},
		{"tpch_18", "18.sql"},
		{"tpch_19", "19.sql"},
		{"tpch_20", "20.sql"},
		{"tpch_21", "21.sql"},
		{"tpch_22", "22.sql"},
	}

	dialect := generic.NewGenericDialect()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, err := os.ReadFile(filepath.Join("fixtures", "tpch", tt.filename))
			require.NoError(t, err, "Failed to read %s", tt.filename)

			originalSQL := string(sql)
			stmts, err := parser.ParseSQL(dialect, originalSQL)
			require.NoError(t, err, "Failed to parse %s", tt.filename)
			require.NotEmpty(t, stmts, "Expected at least one statement from %s", tt.filename)

			// Serialize each statement back to SQL
			for i, stmt := range stmts {
				serialized := stmt.String()
				assert.NotEmpty(t, serialized, "Statement %d from %s serialized to empty string", i, tt.filename)

				// Verify the serialized SQL can be parsed again
				reparsed, err := parser.ParseSQL(dialect, serialized)
				assert.NoError(t, err, "Failed to re-parse serialized statement %d from %s", i, tt.filename)
				assert.NotEmpty(t, reparsed, "Re-parsed statement %d from %s is empty", i, tt.filename)
			}
		})
	}
}
