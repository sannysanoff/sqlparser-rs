#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests

echo "=== Testing all postgres CREATE tests ==="
go test ./postgres -v -run "TestPostgresCreate" 2>&1 | grep -E "^(=== RUN|--- (PASS|FAIL))" | head -50
