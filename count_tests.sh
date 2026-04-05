#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go
echo "=== Counting Passing Tests ==="
go test ./tests/... -v 2>&1 | grep -E "^--- PASS:" | wc -l
echo "=== Counting Failing Tests ==="
go test ./tests/... -v 2>&1 | grep -E "^--- FAIL:" | wc -l
