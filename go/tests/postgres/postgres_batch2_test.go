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

// Package postgres contains additional PostgreSQL-specific SQL parsing tests.
// These tests are ported from tests/sqlparser_postgres.rs in the Rust implementation.
package postgres

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
)

// TestPostgresPgBinaryOps tests PostgreSQL-specific binary operators
// Reference: tests/sqlparser_postgres.rs:2214
func TestPostgresPgBinaryOps(t *testing.T) {
	pg := pg()
	pg.VerifiedExpr(t, "a # b")
	pg.VerifiedExpr(t, "a ^ b")
}

// TestPostgresAmpersandArobase tests &@ operator parsing
// Reference: tests/sqlparser_postgres.rs:2284
func TestPostgresAmpersandArobase(t *testing.T) {
	pg := pg()
	pg.OneStatementParsesTo(t, "a&@b", "a &@ b")
}

// TestPostgresPgUnaryOps tests PostgreSQL-specific unary operators
// Reference: tests/sqlparser_postgres.rs:2290
func TestPostgresPgUnaryOps(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "SELECT |/a")
	pg.VerifiedStmt(t, "SELECT ||/a")
	pg.VerifiedStmt(t, "SELECT !!a")
	pg.VerifiedStmt(t, "SELECT @a")
}

// TestPostgresPgPostfixFactorial tests postfix factorial operator
// Reference: tests/sqlparser_postgres.rs:2310
func TestPostgresPgPostfixFactorial(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "SELECT a!")
}

// TestPostgresPgRegexMatchOps tests PostgreSQL regex match operators
// Reference: tests/sqlparser_postgres.rs:2326
func TestPostgresPgRegexMatchOps(t *testing.T) {
	pg := pg()
	pg.VerifiedExpr(t, "'abc' ~ '^a'")
	pg.VerifiedExpr(t, "'abc' ~* '^a'")
	pg.VerifiedExpr(t, "'abc' !~ '^a'")
	pg.VerifiedExpr(t, "'abc' !~* '^a'")
}

// TestPostgresPgLikeMatchOps tests PostgreSQL LIKE match operators
// Reference: tests/sqlparser_postgres.rs:2370
func TestPostgresPgLikeMatchOps(t *testing.T) {
	pg := pg()
	pg.VerifiedExpr(t, "'abc' ~~ 'a_c%'")
	pg.VerifiedExpr(t, "'abc' ~~* 'a_c%'")
	pg.VerifiedExpr(t, "'abc' !~~ 'a_c%'")
	pg.VerifiedExpr(t, "'abc' !~~* 'a_c%'")
}

// TestPostgresPgBitwiseXor tests PostgreSQL bitwise XOR operator
// Reference: tests/sqlparser_postgres.rs:2218
func TestPostgresPgBitwiseXor(t *testing.T) {
	pg := pg()
	pg.VerifiedExpr(t, "a # b")
}

// TestPostgresPgExp tests PostgreSQL exponentiation operator
// Reference: tests/sqlparser_postgres.rs:2219
func TestPostgresPgExp(t *testing.T) {
	pg := pg()
	pg.VerifiedExpr(t, "a ^ b")
}

// TestPostgresPgShiftOps tests PostgreSQL bitwise shift operators
// Reference: tests/sqlparser_postgres.rs:2220
func TestPostgresPgShiftOps(t *testing.T) {
	pgAndGeneric := pgAndGeneric()
	pgAndGeneric.VerifiedExpr(t, "a >> b")
	pgAndGeneric.VerifiedExpr(t, "a << b")
}

// TestPostgresPgOverlap tests PostgreSQL overlap operator
// Reference: tests/sqlparser_postgres.rs:2222
func TestPostgresPgOverlap(t *testing.T) {
	pg := pg()
	pg.VerifiedExpr(t, "a && b")
}

// TestPostgresPgStartsWith tests PostgreSQL starts with operator
// Reference: tests/sqlparser_postgres.rs:2223
func TestPostgresPgStartsWith(t *testing.T) {
	pg := pg()
	pg.VerifiedExpr(t, "a ^@ b")
}

// TestPostgresArraySubscript tests array subscript expression
// Reference: tests/sqlparser_postgres.rs:2410
func TestPostgresArraySubscript(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "SELECT a[1]")
	pg.VerifiedStmt(t, "SELECT a[1 + 2]")
	pg.VerifiedStmt(t, "SELECT a[1][2]")
}

// TestPostgresArrayMultiSubscript tests multi-dimensional array subscript
// Reference: tests/sqlparser_postgres.rs:2519
func TestPostgresArrayMultiSubscript(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "SELECT a[1][2][3]")
	pg.VerifiedStmt(t, "SELECT a[1 + 2][3 + 4]")
}

// TestPostgresArrayIndexExpr tests array index expressions
// Reference: tests/sqlparser_postgres.rs:2605
func TestPostgresArrayIndexExpr(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "SELECT (ARRAY['a', 'b'])[1]")
	pg.VerifiedStmt(t, "SELECT (ARRAY['a', 'b'])[1 + 2]")
}

// TestPostgresCreateAnonymousIndex tests CREATE INDEX without name
// Reference: tests/sqlparser_postgres.rs:2668
func TestPostgresCreateAnonymousIndex(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "CREATE INDEX ON t(a)")
}

// TestPostgresCreateBloom tests CREATE INDEX USING bloom
// Reference: tests/sqlparser_postgres.rs:2703
func TestPostgresCreateBloom(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "CREATE INDEX ON t USING bloom(a)")
}

// TestPostgresCreateBrin tests CREATE INDEX USING brin
// Reference: tests/sqlparser_postgres.rs:2866
func TestPostgresCreateBrin(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "CREATE INDEX ON t USING brin(a)")
}

// TestPostgresCopyFromStdin tests COPY FROM STDIN
// Reference: tests/sqlparser_postgres.rs:1101
func TestPostgresCopyFromStdin(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "COPY users FROM STDIN")
	pg.VerifiedStmt(t, "COPY users (a, b) FROM STDIN")
}

// TestPostgresCopyFromError tests COPY FROM error handling
// Reference: tests/sqlparser_postgres.rs:1356
func TestPostgresCopyFromError(t *testing.T) {
	dialect := postgresql.NewPostgreSqlDialect()
	_, err := parser.ParseSQL(dialect, "COPY (SELECT 42 AS a, 'hello' AS b) FROM 'query.csv'")
	require.Error(t, err)
}

// TestPostgresCopyFromBeforeV90 tests COPY FROM with legacy syntax
// Reference: tests/sqlparser_postgres.rs:1491
func TestPostgresCopyFromBeforeV90(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "COPY users FROM 'data.csv' BINARY DELIMITER ',' NULL 'null' CSV HEADER QUOTE '\"' ESCAPE '\\' FORCE NOT NULL column")
}

// TestPostgresCopyToBeforeV90 tests COPY TO with legacy syntax
// Reference: tests/sqlparser_postgres.rs:1548
func TestPostgresCopyToBeforeV90(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "COPY users TO 'data.csv' BINARY DELIMITER ',' NULL 'null' CSV HEADER QUOTE '\"' ESCAPE '\\' FORCE QUOTE column")
}

// TestPostgresSetTransaction tests SET TRANSACTION statements
// Reference: tests/sqlparser_postgres.rs:3283
func TestPostgresSetTransaction(t *testing.T) {
	pg := pg()
	pg.VerifiedStmt(t, "SET TRANSACTION SNAPSHOT '000003A1-1'")
	pg.VerifiedStmt(t, "SET SESSION CHARACTERISTICS AS TRANSACTION READ ONLY, READ WRITE, ISOLATION LEVEL SERIALIZABLE")
}

// TestPostgresDoublePrecision tests DOUBLE PRECISION data type
// Reference: tests/sqlparser_postgres.rs:3633
func TestPostgresDoublePrecision(t *testing.T) {
	pgAndGeneric := pgAndGeneric()
	pgAndGeneric.VerifiedStmt(t, "CREATE TABLE t (c DOUBLE PRECISION)")
}

// TestPostgresAnalyze tests ANALYZE statements
// Reference: tests/sqlparser_postgres.rs:8709
func TestPostgresAnalyze(t *testing.T) {
	pgAndGeneric := pgAndGeneric()
	pgAndGeneric.VerifiedStmt(t, "ANALYZE")
	pgAndGeneric.VerifiedStmt(t, "ANALYZE t")
	pgAndGeneric.VerifiedStmt(t, "ANALYZE t (col1, col2)")
}

// TestPostgresLockTable tests LOCK TABLE statements
// Reference: tests/sqlparser_postgres.rs:8734
func TestPostgresLockTable(t *testing.T) {
	pgAndGeneric := pgAndGeneric()
	pgAndGeneric.OneStatementParsesTo(t, "LOCK public.widgets IN EXCLUSIVE MODE", "LOCK TABLE public.widgets IN EXCLUSIVE MODE")
	pgAndGeneric.VerifiedStmt(t, "LOCK TABLE public.widgets NOWAIT")
	pgAndGeneric.VerifiedStmt(t, "LOCK TABLE ONLY public.widgets, analytics.events * IN SHARE ROW EXCLUSIVE MODE NOWAIT")
}
