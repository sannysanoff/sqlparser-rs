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

// Package fuzz provides comprehensive fuzz testing for the SQL parser.
// This is the Go port of the sqlparser-rs fuzz testing targets.
package fuzz

import (
	"testing"

	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/parser"
)

// FuzzParser tests the generic dialect parser with arbitrary input.
// The parser should never panic, only return errors for invalid input.
func FuzzParser(f *testing.F) {
	// Seed corpus with valid SQL examples covering various SQL constructs
	seeds := []string{
		// Basic SELECT statements
		"SELECT * FROM t",
		"SELECT a, b, c FROM t WHERE x = 1",
		"SELECT DISTINCT a FROM t",
		"SELECT ALL a FROM t",
		"SELECT * FROM t1, t2",

		// INSERT statements
		"INSERT INTO t VALUES (1, 2, 3)",
		"INSERT INTO t (a, b, c) VALUES (1, 2, 3)",
		"INSERT INTO t SELECT * FROM u",

		// UPDATE statements
		"UPDATE t SET x = 1 WHERE y = 2",
		"UPDATE t SET x = 1, y = 2 WHERE z = 3",

		// DELETE statements
		"DELETE FROM t WHERE x = 1",
		"DELETE FROM t",

		// CREATE TABLE statements
		"CREATE TABLE t (id INT PRIMARY KEY)",
		"CREATE TABLE t (a INT, b VARCHAR(255))",
		"CREATE TABLE IF NOT EXISTS t (id INT)",

		// Complex queries
		"SELECT * FROM t WHERE x IN (SELECT y FROM u)",
		"SELECT * FROM t1 JOIN t2 ON t1.id = t2.id",
		"SELECT * FROM t1 LEFT JOIN t2 ON t1.id = t2.id",
		"SELECT * FROM t1 INNER JOIN t2 USING (id)",
		"SELECT COUNT(*) FROM t GROUP BY x",
		"SELECT * FROM t GROUP BY x HAVING COUNT(*) > 1",

		// CTEs
		"WITH cte AS (SELECT * FROM t) SELECT * FROM cte",
		"WITH cte1 AS (SELECT * FROM t), cte2 AS (SELECT * FROM cte1) SELECT * FROM cte2",

		// ORDER BY and LIMIT
		"SELECT * FROM t ORDER BY x",
		"SELECT * FROM t ORDER BY x DESC",
		"SELECT * FROM t ORDER BY x LIMIT 10 OFFSET 5",
		"SELECT * FROM t LIMIT 10",

		// Subqueries
		"SELECT * FROM (SELECT * FROM t) AS u",
		"SELECT * FROM t WHERE x = (SELECT MAX(y) FROM u)",
		"SELECT * FROM t WHERE EXISTS (SELECT * FROM u WHERE u.id = t.id)",

		// DROP statements
		"DROP TABLE t",
		"DROP TABLE IF EXISTS t",
		"DROP VIEW v",

		// ALTER statements
		"ALTER TABLE t ADD COLUMN x INT",
		"ALTER TABLE t DROP COLUMN x",
		"ALTER TABLE t RENAME TO u",

		// Window functions
		"SELECT ROW_NUMBER() OVER (ORDER BY x) FROM t",
		"SELECT SUM(x) OVER (PARTITION BY y) FROM t",
		"SELECT * FROM (SELECT *, ROW_NUMBER() OVER (ORDER BY x) AS rn FROM t) WHERE rn = 1",

		// CASE expressions
		"SELECT CASE WHEN x = 1 THEN 'one' ELSE 'other' END FROM t",
		"SELECT CASE x WHEN 1 THEN 'one' WHEN 2 THEN 'two' END FROM t",

		// UNION and set operations
		"SELECT * FROM t1 UNION SELECT * FROM t2",
		"SELECT * FROM t1 UNION ALL SELECT * FROM t2",
		"SELECT * FROM t1 INTERSECT SELECT * FROM t2",

		// Comments
		"SELECT /* comment */ * FROM t",
		"SELECT * FROM t -- end comment",
		"/* multi\nline\ncomment */ SELECT * FROM t",

		// Empty and edge cases
		"",
		"   ",
		";",
		"SELECT * FROM",
		"SELEC", // typo intentionally
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	dialect := generic.NewGenericDialect()

	f.Fuzz(func(t *testing.T, data string) {
		// The parser should never panic on any input, only return errors
		_, _ = parser.ParseSQL(dialect, data)
	})
}

// FuzzPostgreSQL tests the PostgreSQL dialect parser with arbitrary input.
func FuzzPostgreSQL(f *testing.F) {
	// PostgreSQL-specific seed corpus
	seeds := []string{
		// Basic statements
		"SELECT * FROM t",
		"SELECT * FROM t WHERE x = 1",

		// PostgreSQL-specific features
		"SELECT * FROM t WHERE x = ANY(ARRAY[1, 2, 3])",
		"SELECT * FROM t WHERE x = ALL(ARRAY[1, 2, 3])",
		"SELECT * FROM t LIMIT 10 OFFSET 5",
		"SELECT * FROM t FETCH FIRST 10 ROWS ONLY",
		"SELECT * FROM t OFFSET 5 ROWS",

		// Dollar-quoted strings
		"SELECT $$hello world$$",
		"SELECT $tag$hello$tag$",

		// Arrays
		"SELECT ARRAY[1, 2, 3]",
		"SELECT t.array_col[1] FROM t",
		"SELECT * FROM t WHERE array_col @> ARRAY[1]",

		// JSON operators
		"SELECT '{\"a\":1}'::json->'a'",
		"SELECT '{\"a\":1}'::json->>'a'",
		"SELECT '{\"a\":{\"b\":1}}'::json#>'{a,b}'",

		// Type casting
		"SELECT '123'::integer",
		"SELECT '123'::int4",
		"SELECT x::text FROM t",
		"SELECT CAST(x AS integer) FROM t",

		// COPY statement
		"COPY t TO STDOUT",
		"COPY t (a, b, c) FROM STDIN",

		// LISTEN/NOTIFY
		"LISTEN channel",
		"NOTIFY channel",
		"NOTIFY channel, 'payload'",
		"UNLISTEN channel",

		// DO statement
		"DO $$ BEGIN PERFORM 1; END $$",

		// RETURNING clause
		"INSERT INTO t VALUES (1) RETURNING id",
		"UPDATE t SET x = 1 RETURNING *",
		"DELETE FROM t RETURNING id",

		// Window functions with PostgreSQL syntax
		"SELECT row_number() OVER () FROM t",
		"SELECT sum(x) OVER (PARTITION BY y ORDER BY z) FROM t",
		"SELECT sum(x) OVER w FROM t WINDOW w AS (PARTITION BY y)",

		// Distinct ON (PostgreSQL-specific)
		"SELECT DISTINCT ON (x) * FROM t ORDER BY x",

		// Custom operators
		"SELECT * FROM t WHERE x OPERATOR(=) y",
		"SELECT x || y FROM t",

		// Quoted identifiers
		`SELECT * FROM "My Table"`,
		`SELECT "Column Name" FROM t`,
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	dialect := postgresql.NewPostgreSqlDialect()

	f.Fuzz(func(t *testing.T, data string) {
		_, _ = parser.ParseSQL(dialect, data)
	})
}

// FuzzMySQL tests the MySQL dialect parser with arbitrary input.
func FuzzMySQL(f *testing.F) {
	// MySQL-specific seed corpus
	seeds := []string{
		// Basic statements
		"SELECT * FROM t",
		"SELECT * FROM t WHERE x = 1",

		// MySQL-specific LIMIT syntax
		"SELECT * FROM t LIMIT 10",
		"SELECT * FROM t LIMIT 10, 20",
		"SELECT * FROM t LIMIT 10 OFFSET 20",

		// Backtick identifiers
		"SELECT * FROM `my table`",
		"SELECT `column name` FROM t",
		"SELECT `t`.`c` FROM `t`",

		// MySQL data types
		"CREATE TABLE t (id INT AUTO_INCREMENT PRIMARY KEY)",
		"CREATE TABLE t (name VARCHAR(255) CHARACTER SET utf8mb4)",
		"CREATE TABLE t (price DECIMAL(10, 2))",

		// ENGINE and character set
		"CREATE TABLE t (id INT) ENGINE=InnoDB",
		"CREATE TABLE t (id INT) DEFAULT CHARSET=utf8mb4",

		// INSERT with ON DUPLICATE KEY UPDATE
		"INSERT INTO t VALUES (1) ON DUPLICATE KEY UPDATE x = 1",

		// REPLACE statement
		"REPLACE INTO t VALUES (1, 2, 3)",
		"REPLACE INTO t (a, b) VALUES (1, 2)",

		// MySQL-specific functions
		"SELECT CONCAT('a', 'b', 'c')",
		"SELECT IFNULL(x, 0) FROM t",
		"SELECT COALESCE(a, b, c) FROM t",
		"SELECT NOW()",

		// REGEXP operator
		"SELECT * FROM t WHERE x REGEXP '^[0-9]+'",

		// STRAIGHT_JOIN
		"SELECT STRAIGHT_JOIN * FROM t1, t2 WHERE t1.id = t2.id",

		// SQL_CALC_FOUND_ROWS and FOUND_ROWS()
		"SELECT SQL_CALC_FOUND_ROWS * FROM t LIMIT 10",

		// FORCE INDEX
		"SELECT * FROM t FORCE INDEX (idx) WHERE x = 1",
		"SELECT * FROM t USE INDEX (idx) WHERE x = 1",
		"SELECT * FROM t IGNORE INDEX (idx) WHERE x = 1",

		// LOCK IN SHARE MODE
		"SELECT * FROM t WHERE x = 1 LOCK IN SHARE MODE",
		"SELECT * FROM t WHERE x = 1 FOR UPDATE",

		// MySQL SHOW statements
		"SHOW TABLES",
		"SHOW DATABASES",
		"SHOW COLUMNS FROM t",
		"SHOW CREATE TABLE t",
		"SHOW INDEX FROM t",
		"SHOW VARIABLES LIKE 'max%'",

		// MySQL ALTER syntax
		"ALTER TABLE t ADD COLUMN x INT FIRST",
		"ALTER TABLE t ADD COLUMN y INT AFTER z",
		"ALTER TABLE t MODIFY COLUMN x VARCHAR(100)",
		"ALTER TABLE t CHANGE COLUMN x y INT",

		// String literals with escape
		"SELECT 'hello\\nworld'",
		`SELECT "hello"`,

		// Boolean values
		"SELECT TRUE, FALSE",
		"SELECT * FROM t WHERE active = TRUE",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	dialect := mysql.NewMySqlDialect()

	f.Fuzz(func(t *testing.T, data string) {
		_, _ = parser.ParseSQL(dialect, data)
	})
}

// FuzzBigQuery tests the BigQuery dialect parser with arbitrary input.
func FuzzBigQuery(f *testing.F) {
	// BigQuery-specific seed corpus
	seeds := []string{
		// Basic statements
		"SELECT * FROM t",
		"SELECT * FROM t WHERE x = 1",

		// BigQuery project.dataset.table notation
		"SELECT * FROM project.dataset.table",
		"SELECT * FROM `project.dataset.table`",
		"SELECT * FROM dataset.table",

		// BigQuery-specific functions
		"SELECT CURRENT_TIMESTAMP()",
		"SELECT TIMESTAMP_ADD(ts, INTERVAL 1 HOUR)",
		"SELECT DATE_TRUNC(d, MONTH)",
		"SELECT FORMAT_TIMESTAMP('%Y-%m-%d', ts)",

		// STRUCT literals
		"SELECT STRUCT(1 AS x, 'a' AS y)",
		"SELECT STRUCT<INT64, STRING>(1, 'a')",

		// ARRAY literals and functions
		"SELECT [1, 2, 3]",
		"SELECT ARRAY<INT64>[1, 2, 3]",
		"SELECT ARRAY_AGG(x) FROM t",

		// UNNEST
		"SELECT * FROM UNNEST([1, 2, 3])",
		"SELECT * FROM t, UNNEST(arr) AS u",

		// QUALIFY clause
		"SELECT * FROM t QUALIFY ROW_NUMBER() OVER (PARTITION BY x) = 1",

		// EXCEPT and REPLACE
		"SELECT * EXCEPT(x) FROM t",
		"SELECT * REPLACE(y AS x) FROM t",
		"SELECT * EXCEPT(a) REPLACE(b AS c) FROM t",

		// PIVOT and UNPIVOT
		"SELECT * FROM t PIVOT (SUM(x) FOR y IN ('a', 'b', 'c'))",
		"SELECT * FROM t UNPIVOT (x FOR y IN (a, b, c))",

		// TABLESAMPLE
		"SELECT * FROM t TABLESAMPLE SYSTEM (10 PERCENT)",
		"SELECT * FROM t TABLESAMPLE BERNOULLI (1000 ROWS)",

		// FOR SYSTEM_TIME AS OF
		"SELECT * FROM t FOR SYSTEM_TIME AS OF TIMESTAMP '2023-01-01'",

		// EXTRACT with different date parts
		"SELECT EXTRACT(YEAR FROM d) FROM t",
		"SELECT EXTRACT(WEEK FROM d) FROM t",
		"SELECT EXTRACT(DAYOFWEEK FROM d) FROM t",

		// SAFE. prefix
		"SELECT SAFE.DIVIDE(x, y) FROM t",

		// Parameterized queries (@param)
		"SELECT * FROM t WHERE x = @param",
		"SELECT * FROM t WHERE y = @other_param",

		// JOIN with USING
		"SELECT * FROM t1 JOIN t2 USING (x)",

		// INTERVAL
		"SELECT DATE_ADD(d, INTERVAL 1 DAY) FROM t",
		"SELECT TIMESTAMP_SUB(ts, INTERVAL 1 HOUR) FROM t",

		// COUNT(DISTINCT)
		"SELECT COUNT(DISTINCT x) FROM t",
		"SELECT COUNT(DISTINCT x, y) FROM t",

		// String aggregation
		"SELECT STRING_AGG(x, ',') FROM t",
		"SELECT STRING_AGG(x, ',' ORDER BY y) FROM t",

		// Window functions with specific BigQuery syntax
		"SELECT RANK() OVER (ORDER BY x) FROM t",
		"SELECT DENSE_RANK() OVER (PARTITION BY y ORDER BY x) FROM t",
		"SELECT PERCENTILE_CONT(x, 0.5) OVER () FROM t",

		// ML functions
		"SELECT ML.PREDICT(MODEL `project.dataset.model`, (SELECT * FROM t))",

		// Approximate functions
		"SELECT APPROX_COUNT_DISTINCT(x) FROM t",
		"SELECT APPROX_QUANTILES(x, 100) FROM t",
	}

	for _, seed := range seeds {
		f.Add(seed)
	}

	dialect := bigquery.NewBigQueryDialect()

	f.Fuzz(func(t *testing.T, data string) {
		_, _ = parser.ParseSQL(dialect, data)
	})
}
