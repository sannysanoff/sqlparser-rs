SELECT * FROM users
SELECT count() FROM events WHERE event_date > '2024-01-01'
SELECT * FROM events FINAL
SELECT column1, column2 FROM table1 WHERE column1 > 0
SELECT arrayJoin([1, 2, 3]) AS num
SELECT a.*, b.name FROM table_a AS a INNER JOIN table_b AS b ON a.id = b.id
SELECT id, nested.col1, nested.col2 FROM table_with_nested
CREATE TABLE test_table (id UInt64, name String, timestamp DateTime) ENGINE = MergeTree()
SELECT id, count() AS cnt FROM users WHERE active = 1 GROUP BY id HAVING cnt > 5 ORDER BY cnt DESC LIMIT 20
SELECT toYear(timestamp) AS year, sum(revenue) FROM sales GROUP BY year ORDER BY year
SELECT * FROM users ORDER BY name ASC
SELECT * FROM users ORDER BY name DESC NULLS LAST
SELECT DISTINCT city FROM users
SELECT name, age FROM users WHERE age BETWEEN 18 AND 65
SELECT name, age FROM users WHERE name LIKE '%john%'
SELECT id, name FROM users LIMIT 10 OFFSET 5
SELECT count(*), min(price), max(price), avg(price) FROM products
SELECT * FROM t1 UNION ALL SELECT * FROM t2
SELECT a, b, c FROM (SELECT x AS a, y AS b, z AS c FROM inner_table) AS sub
