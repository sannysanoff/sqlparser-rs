SELECT * FROM users
SELECT count() FROM events WHERE event_date > '2024-01-01'
SELECT * FROM events FINAL
SELECT column1, column2 FROM table1 PREWHERE column3 > 100 WHERE column1 > 0
SELECT * FROM table1 LIMIT 10 BY category
SELECT key, count() AS cnt FROM table1 GROUP BY key ORDER BY cnt DESC LIMIT 10 WITH TIES
SELECT key, sum(value) FROM table1 GROUP BY key WITH ROLLUP WITH FILL FROM 0 TO 100 STEP 10
SELECT arrayJoin([1, 2, 3]) AS num
SELECT quantile(0.5)(value) AS med FROM measurements
SELECT a.*, b.name FROM table_a AS a INNER JOIN table_b AS b ON a.id = b.id
SELECT id, nested.col1, nested.col2 FROM table_with_nested
CREATE TABLE test_table (id UInt64, name String, timestamp DateTime) ENGINE = MergeTree()
SELECT id, count() AS cnt FROM users WHERE active = 1 GROUP BY id HAVING cnt > 5 ORDER BY cnt DESC LIMIT 20
SELECT * FROM table1 SETTINGS max_threads = 4
SELECT toYear(timestamp) AS year, sum(revenue) FROM sales GROUP BY year ORDER BY year
