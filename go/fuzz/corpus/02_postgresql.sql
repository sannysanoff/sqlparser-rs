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

-- PostgreSQL-specific examples

-- Dollar-quoted strings
SELECT $$hello world$$;
SELECT $tag$hello world$tag$;
SELECT $func$
CREATE OR REPLACE FUNCTION test()
RETURNS INTEGER AS $$
BEGIN
    RETURN 1;
END;
$$ LANGUAGE plpgsql;
$func$;

-- Arrays
SELECT ARRAY[1, 2, 3];
SELECT ARRAY['a', 'b', 'c'];
SELECT t.array_col[1] FROM t;
SELECT t.array_col[1:3] FROM t;
SELECT * FROM t WHERE array_col @> ARRAY[1];
SELECT * FROM t WHERE array_col <@ ARRAY[1, 2, 3];
SELECT * FROM t WHERE array_col && ARRAY[1];

-- JSON and JSONB
SELECT '{"a": 1}'::json;
SELECT '{"a": 1}'::jsonb;
SELECT '{"a": {"b": 1}}'::json->'a';
SELECT '{"a": {"b": 1}}'::json->>'a';
SELECT '{"a": {"b": 1}}'::json#>'{a,b}';
SELECT '{"a": {"b": 1}}'::json#>>'{a,b}';
SELECT * FROM t WHERE json_col @> '{"a": 1}'::jsonb;
SELECT * FROM t WHERE json_col ? 'a';
SELECT * FROM t WHERE json_col ?| ARRAY['a', 'b'];
SELECT jsonb_pretty(json_col) FROM t;
SELECT jsonb_build_object('a', 1, 'b', 2);
SELECT jsonb_agg(x) FROM t;

-- Type casting syntax
SELECT '123'::integer;
SELECT '123'::int4;
SELECT '123'::bigint;
SELECT 'text'::varchar(100);
SELECT x::text FROM t;
SELECT x::integer FROM t;
SELECT CAST(x AS integer) FROM t;
SELECT x::regclass;

-- COPY statements
COPY t TO STDOUT;
COPY t TO '/tmp/data.csv' CSV;
COPY t (a, b, c) FROM STDIN;
COPY t FROM '/tmp/data.csv' CSV HEADER;

-- LISTEN/NOTIFY/UNLISTEN
LISTEN channel;
LISTEN "my channel";
NOTIFY channel;
NOTIFY channel, 'payload message';
UNLISTEN channel;
UNLISTEN *;

-- RETURNING clause
INSERT INTO t VALUES (1) RETURNING id;
INSERT INTO t VALUES (1) RETURNING *;
UPDATE t SET x = 1 RETURNING id;
UPDATE t SET x = 1 RETURNING *;
DELETE FROM t RETURNING id;
DELETE FROM t RETURNING *;

-- Distinct ON
SELECT DISTINCT ON (x) * FROM t ORDER BY x;
SELECT DISTINCT ON (x, y) * FROM t ORDER BY x, y;

-- Window functions with PostgreSQL syntax
SELECT row_number() OVER () FROM t;
SELECT row_number() OVER (ORDER BY x) FROM t;
SELECT sum(x) OVER (PARTITION BY y) FROM t;
SELECT sum(x) OVER (PARTITION BY y ORDER BY z) FROM t;
SELECT sum(x) OVER w FROM t WINDOW w AS (PARTITION BY y);
SELECT sum(x) OVER w FROM t WINDOW w AS (PARTITION BY y ORDER BY z);

-- LATERAL join
SELECT * FROM t, LATERAL (SELECT * FROM u WHERE u.id = t.id) AS sub;
SELECT * FROM t JOIN LATERAL (SELECT * FROM u WHERE u.id = t.id) AS sub ON true;

-- Custom operators
SELECT x OPERATOR(pg_catalog.+) y FROM t;
SELECT * FROM t WHERE x OPERATOR(=) y;

-- POSIX regular expressions
SELECT * FROM t WHERE x ~ '^[0-9]+';
SELECT * FROM t WHERE x ~* '^[a-z]+';
SELECT * FROM t WHERE x !~ '^[0-9]+';
SELECT * FROM t WHERE x !~* '^[a-z]+';

-- Geometric types
SELECT point(0, 0);
SELECT box(point(0,0), point(1,1));
SELECT circle(point(0,0), 1);
SELECT line(point(0,0), point(1,1));
SELECT lseg(point(0,0), point(1,1));
SELECT polygon(point(0,0), point(1,0), point(1,1), point(0,1));

-- Range types
SELECT int4range(1, 10);
SELECT int4range(1, 10, '[)');
SELECT '[1,10)'::int4range;
SELECT * FROM t WHERE range_col @> 5;
SELECT * FROM t WHERE range_col && '[5,6)'::int4range;

-- Full text search
SELECT * FROM t WHERE to_tsvector('english', text_col) @@ to_tsquery('english', 'search');
SELECT plainto_tsquery('search term');
SELECT phraseto_tsquery('search term');
SELECT to_tsvector('english', 'text to search');

-- Do statement
DO $$
BEGIN
    PERFORM some_function();
END;
$$;

-- Prepare/Execute
PREPARE stmt AS SELECT * FROM t WHERE x = $1;
EXECUTE stmt(1);
DEALLOCATE stmt;

-- Savepoints and transaction control
BEGIN;
SAVEPOINT my_savepoint;
ROLLBACK TO SAVEPOINT my_savepoint;
RELEASE SAVEPOINT my_savepoint;
COMMIT;

-- Table inheritance (PostgreSQL-specific)
CREATE TABLE parent (id int);
CREATE TABLE child () INHERITS (parent);

-- Advisory locks
SELECT pg_advisory_lock(1);
SELECT pg_advisory_unlock(1);
SELECT pg_try_advisory_lock(1);

-- EXPLAIN options
EXPLAIN SELECT * FROM t;
EXPLAIN ANALYZE SELECT * FROM t;
EXPLAIN (ANALYZE, BUFFERS) SELECT * FROM t;
EXPLAIN (FORMAT JSON) SELECT * FROM t;

-- TRUNCATE options
TRUNCATE t;
TRUNCATE t RESTART IDENTITY;
TRUNCATE t CONTINUE IDENTITY;
TRUNCATE t CASCADE;
TRUNCATE t RESTRICT;
