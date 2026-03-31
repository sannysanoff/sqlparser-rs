-- Licensed to the Apache Software Foundation (ASF) under one
-- or more contributor license agreements.  See the NOTICE file
-- distributed with this work for additional information
-- regarding copyright ownership.  The ASF licenses this file
-- to you under the Apache License, Version 2.0 (the
-- "License"); you may not use this file except in compliance
-- with the License.  You may obtain a copy of the License at
--
--   http://www.apache.org/licenses/LICENSE-2.0
--
-- Unless required by applicable law or agreed to in writing,
-- software distributed under the License is distributed on an
-- "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
-- KIND, either express or implied.  See the License for the
-- specific language governing permissions and limitations
-- under the License.

-- Basic SELECT statements
SELECT * FROM t;
SELECT a, b, c FROM t WHERE x = 1;
SELECT DISTINCT a FROM t ORDER BY a;
SELECT ALL a FROM t;

-- INSERT statements
INSERT INTO t VALUES (1, 2, 3);
INSERT INTO t (a, b, c) VALUES (1, 2, 3);
INSERT INTO t SELECT * FROM u;
INSERT INTO t VALUES (1) ON CONFLICT DO NOTHING;

-- UPDATE statements
UPDATE t SET x = 1 WHERE y = 2;
UPDATE t SET x = 1, y = 2 WHERE z = 3;
UPDATE t1 SET t1.x = t2.y FROM t2 WHERE t1.id = t2.id;

-- DELETE statements
DELETE FROM t WHERE x = 1;
DELETE FROM t;
DELETE FROM t USING u WHERE t.id = u.id;

-- CREATE TABLE statements
CREATE TABLE t (id INT PRIMARY KEY);
CREATE TABLE t (a INT, b VARCHAR(255));
CREATE TABLE IF NOT EXISTS t (id INT);
CREATE TABLE t (id INT, CONSTRAINT pk PRIMARY KEY (id));

-- CREATE VIEW
CREATE VIEW v AS SELECT * FROM t;
CREATE OR REPLACE VIEW v AS SELECT * FROM t;

-- Complex queries with JOINs
SELECT * FROM t1 JOIN t2 ON t1.id = t2.id;
SELECT * FROM t1 LEFT JOIN t2 ON t1.id = t2.id;
SELECT * FROM t1 RIGHT JOIN t2 ON t1.id = t2.id;
SELECT * FROM t1 FULL OUTER JOIN t2 ON t1.id = t2.id;
SELECT * FROM t1 CROSS JOIN t2;
SELECT * FROM t1 INNER JOIN t2 USING (id);
SELECT * FROM t1 NATURAL JOIN t2;

-- Aggregations
SELECT COUNT(*) FROM t;
SELECT COUNT(DISTINCT x) FROM t;
SELECT SUM(x), AVG(x), MIN(x), MAX(x) FROM t;
SELECT x, COUNT(*) FROM t GROUP BY x;
SELECT x, COUNT(*) FROM t GROUP BY x HAVING COUNT(*) > 1;

-- Subqueries
SELECT * FROM t WHERE x IN (SELECT y FROM u);
SELECT * FROM t WHERE x NOT IN (SELECT y FROM u);
SELECT * FROM t WHERE EXISTS (SELECT * FROM u WHERE u.id = t.id);
SELECT * FROM t WHERE x = (SELECT MAX(y) FROM u);
SELECT * FROM (SELECT * FROM t) AS u;

-- CTEs (Common Table Expressions)
WITH cte AS (SELECT * FROM t) SELECT * FROM cte;
WITH cte1 AS (SELECT * FROM t), cte2 AS (SELECT * FROM cte1) SELECT * FROM cte2;
WITH RECURSIVE cte AS (
    SELECT 1 AS n
    UNION ALL
    SELECT n + 1 FROM cte WHERE n < 10
) SELECT * FROM cte;

-- ORDER BY and LIMIT
SELECT * FROM t ORDER BY x;
SELECT * FROM t ORDER BY x DESC;
SELECT * FROM t ORDER BY x, y DESC;
SELECT * FROM t ORDER BY x LIMIT 10;
SELECT * FROM t ORDER BY x LIMIT 10 OFFSET 5;
SELECT * FROM t ORDER BY x FETCH FIRST 10 ROWS ONLY;

-- Window functions
SELECT ROW_NUMBER() OVER (ORDER BY x) FROM t;
SELECT RANK() OVER (ORDER BY x) FROM t;
SELECT DENSE_RANK() OVER (PARTITION BY y ORDER BY x) FROM t;
SELECT SUM(x) OVER (PARTITION BY y) FROM t;
SELECT SUM(x) OVER (PARTITION BY y ORDER BY z) FROM t;
SELECT SUM(x) OVER w FROM t WINDOW w AS (PARTITION BY y);
SELECT LAG(x, 1) OVER (ORDER BY y) FROM t;
SELECT LEAD(x, 1, 0) OVER (ORDER BY y) FROM t;

-- CASE expressions
SELECT CASE WHEN x = 1 THEN 'one' ELSE 'other' END FROM t;
SELECT CASE x WHEN 1 THEN 'one' WHEN 2 THEN 'two' ELSE 'other' END FROM t;
SELECT CASE WHEN x > 0 THEN 1 WHEN x < 0 THEN -1 ELSE 0 END FROM t;

-- UNION, INTERSECT, EXCEPT
SELECT * FROM t1 UNION SELECT * FROM t2;
SELECT * FROM t1 UNION ALL SELECT * FROM t2;
SELECT * FROM t1 INTERSECT SELECT * FROM t2;
SELECT * FROM t1 EXCEPT SELECT * FROM t2;

-- DROP statements
DROP TABLE t;
DROP TABLE IF EXISTS t;
DROP VIEW v;
DROP INDEX idx;
DROP SCHEMA s;

-- ALTER statements
ALTER TABLE t ADD COLUMN x INT;
ALTER TABLE t DROP COLUMN x;
ALTER TABLE t RENAME COLUMN x TO y;
ALTER TABLE t ALTER COLUMN x SET NOT NULL;
ALTER TABLE t RENAME TO u;

-- Transactions
BEGIN;
BEGIN TRANSACTION;
START TRANSACTION;
COMMIT;
ROLLBACK;
SAVEPOINT sp;
RELEASE SAVEPOINT sp;
ROLLBACK TO SAVEPOINT sp;

-- SET statements
SET x = 1;
SET SESSION x = 1;
SET LOCAL x = 1;
SET TIME ZONE 'UTC';

-- Comments
/* Multi-line
   comment */
SELECT * FROM t;
SELECT * FROM t; -- Single line comment
SELECT /* inline comment */ * FROM t;

-- Boolean expressions
SELECT * FROM t WHERE x AND y;
SELECT * FROM t WHERE x OR y;
SELECT * FROM t WHERE NOT x;
SELECT * FROM t WHERE x IS NULL;
SELECT * FROM t WHERE x IS NOT NULL;
SELECT * FROM t WHERE x BETWEEN 1 AND 10;
SELECT * FROM t WHERE x LIKE 'pattern%';
SELECT * FROM t WHERE x IN (1, 2, 3);

-- Mathematical expressions
SELECT x + y FROM t;
SELECT x - y FROM t;
SELECT x * y FROM t;
SELECT x / y FROM t;
SELECT x % y FROM t;
SELECT ABS(x) FROM t;
SELECT ROUND(x, 2) FROM t;

-- String functions
SELECT CONCAT('a', 'b') FROM t;
SELECT LENGTH(s) FROM t;
SELECT UPPER(s) FROM t;
SELECT LOWER(s) FROM t;
SELECT SUBSTRING(s, 1, 5) FROM t;
SELECT TRIM(s) FROM t;
SELECT REPLACE(s, 'a', 'b') FROM t;

-- Date/Time functions
SELECT CURRENT_DATE;
SELECT CURRENT_TIMESTAMP;
SELECT NOW();
SELECT DATE_ADD(d, INTERVAL 1 DAY) FROM t;
SELECT DATE_SUB(d, INTERVAL 1 MONTH) FROM t;
SELECT EXTRACT(YEAR FROM d) FROM t;

-- Type casting
SELECT CAST(x AS INTEGER) FROM t;
SELECT CAST(x AS VARCHAR) FROM t;
SELECT x::integer FROM t;
SELECT x::text FROM t;

-- NULL handling
SELECT COALESCE(x, 0) FROM t;
SELECT NULLIF(x, y) FROM t;
SELECT IFNULL(x, 0) FROM t;
SELECT ISNULL(x, 0) FROM t;

-- Conditional aggregation
SELECT COUNT(*) FILTER (WHERE x > 0) FROM t;
SELECT SUM(x) FILTER (WHERE y = 'a') FROM t;
