<!--
  Licensed to the Apache Software Foundation (ASF) under one
  or more contributor license agreements.  See the NOTICE file
  distributed with this work for additional information
  regarding copyright ownership.  The ASF licenses this file
  to you under the Apache License, Version 2.0 (the
  "License"); you may not use this file except in compliance
  with the License.  You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

  Unless required by applicable law or agreed to in writing,
  software distributed under the License is distributed on an
  "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
  KIND, either express or implied.  See the License for the
  specific language governing permissions and limitations
  under the License.
-->

# SQL Parser Fuzz Testing

This directory contains comprehensive fuzz tests for the Go SQL parser. These tests are designed to ensure the parser never panics on arbitrary input and handles edge cases gracefully.

## Overview

This fuzz testing suite is the Go port of the [sqlparser-rs](https://github.com/apache/datafusion-sqlparser-rs) fuzz targets. It tests the SQL parser with various dialects using Go's native fuzzing capabilities introduced in Go 1.18.

## Available Fuzz Tests

### FuzzParser
Tests the generic dialect parser with arbitrary input including:
- Basic SELECT, INSERT, UPDATE, DELETE statements
- Complex queries with JOINs, subqueries, and CTEs
- Window functions and aggregations
- Edge cases like empty strings, malformed SQL, and special characters

```bash
go test -fuzz=FuzzParser -fuzztime=1h
```

### FuzzPostgreSQL
Tests PostgreSQL-specific syntax:
- Dollar-quoted strings ($$...$$)
- Array types and operations
- JSON/JSONB operators
- Type casting syntax (::)
- COPY, LISTEN/NOTIFY statements
- RETURNING clause
- Distinct ON clause
- LATERAL joins

```bash
go test -fuzz=FuzzPostgreSQL -fuzztime=1h
```

### FuzzMySQL
Tests MySQL-specific syntax:
- Backtick-quoted identifiers (`identifier`)
- AUTO_INCREMENT and ENGINE options
- ON DUPLICATE KEY UPDATE
- REPLACE statement
- MySQL-specific LIMIT syntax
- REGEXP operators
- STRAIGHT_JOIN hint
- Index hints (USE INDEX, FORCE INDEX, IGNORE INDEX)
- SHOW statements

```bash
go test -fuzz=FuzzMySQL -fuzztime=1h
```

### FuzzBigQuery
Tests BigQuery-specific syntax:
- Project.dataset.table notation
- STRUCT and ARRAY literals
- UNNEST operations
- QUALIFY clause
- EXCEPT and REPLACE in SELECT
- PIVOT and UNPIVOT
- TABLESAMPLE
- FOR SYSTEM_TIME AS OF
- Safe functions (SAFE. prefix)
- Parameterized queries (@param)

```bash
go test -fuzz=FuzzBigQuery -fuzztime=1h
```

## Running Fuzz Tests

### Basic Usage

Run fuzz tests with default settings (10 seconds):

```bash
cd go/fuzz
go test -fuzz=FuzzParser
```

### Extended Fuzzing

Run fuzz tests for a specified duration:

```bash
# Run for 1 hour
go test -fuzz=FuzzParser -fuzztime=1h

# Run for 30 minutes
go test -fuzz=FuzzParser -fuzztime=30m

# Run for 1000 iterations
go test -fuzz=FuzzParser -fuzztime=1000x
```

### Parallel Fuzzing

Run fuzz tests in parallel for faster coverage:

```bash
# Run with 4 parallel workers
go test -fuzz=FuzzParser -parallel=4 -fuzztime=1h
```

### Running All Fuzz Tests

```bash
# Run all fuzz tests sequentially
go test -fuzz=FuzzParser -fuzztime=30m
go test -fuzz=FuzzPostgreSQL -fuzztime=30m
go test -fuzz=FuzzMySQL -fuzztime=30m
go test -fuzz=FuzzBigQuery -fuzztime=30m
```

## Corpus

The `corpus/` directory contains seed inputs that help guide the fuzzer toward interesting code paths:

- `01_basic_sql.sql` - Standard SQL statements across all dialects
- `02_postgresql.sql` - PostgreSQL-specific features and syntax
- `03_mysql.sql` - MySQL-specific features and syntax
- `04_bigquery.sql` - BigQuery-specific features and syntax
- `05_edge_cases.sql` - Edge cases and potentially problematic inputs

The fuzzer automatically loads these files as seed corpus. You can add more seed files to improve fuzzing effectiveness.

## Understanding Results

### Successful Run

A successful fuzz test run will output something like:

```
fuzz: elapsed: 1h0m0s, execs: 14285714 (3968/sec), new interesting: 123 (total: 456)
```

This indicates:
- `execs`: Total number of test cases executed
- `new interesting`: Number of new coverage-increasing inputs found
- `total`: Total unique inputs in the corpus

### Finding Crashes

If the fuzzer finds a crash (panic), it will:
1. Save the crashing input to a file (e.g., `testdata/fuzz/FuzzParser/<hash>`)
2. Print the crash details including the stack trace
3. Exit with a non-zero status

You can reproduce a crash:

```bash
# Run the specific crashing input
go test -run=FuzzParser/<hash>
```

## Corpus Management

### Adding New Seeds

Add new SQL files to the `corpus/` directory to improve fuzzing:

```bash
# Create a new seed file
cat > corpus/06_custom.sql << 'EOF'
SELECT * FROM my_special_table;
INSERT INTO t VALUES (1, 2, 3);
EOF
```

### Minimizing Corpus

Remove redundant inputs from the corpus:

```bash
go test -fuzz=FuzzParser -fuzzminimize -fuzztime=30m
```

## Continuous Fuzzing

For CI/CD integration, run fuzz tests with a time limit:

```bash
# GitHub Actions example
- name: Fuzz Test
  run: go test -fuzz=FuzzParser -fuzztime=5m
```

## Performance Tuning

### Memory Limits

Set memory limits to prevent resource exhaustion:

```bash
# Limit to 4GB of memory
go test -fuzz=FuzzParser -fuzztime=1h -memprofile=mem.out
```

### CPU Profiling

Profile CPU usage during fuzzing:

```bash
go test -fuzz=FuzzParser -fuzztime=1h -cpuprofile=cpu.out
```

## Troubleshooting

### Parser Panics

If the parser panics, the fuzzer will capture the input. Report issues with:
1. The crashing input
2. The dialect being tested
3. The stack trace

### Slow Fuzzing

If fuzzing is slow:
- Reduce the seed corpus size
- Use `-parallel` flag for parallel execution
- Focus on specific fuzz targets

### Build Errors

Ensure all dependencies are available:

```bash
cd go/fuzz
go mod tidy
go mod download
```

## Contributing

When adding new features to the parser:
1. Add relevant seed corpus entries
2. Run extended fuzzing to catch edge cases
3. Update this documentation with new test coverage

## References

- [Go Fuzzing](https://go.dev/security/fuzz/)
- [sqlparser-rs Fuzz Tests](https://github.com/apache/datafusion-sqlparser-rs/tree/main/fuzz)
- [Apache License 2.0](http://www.apache.org/licenses/LICENSE-2.0)
