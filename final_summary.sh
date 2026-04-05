#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go
echo "=== Final Test Summary ==="
echo "Passing tests:"
go test ./tests/... -v 2>&1 | grep -E "^--- PASS:" | wc -l
echo "Failing tests:"
go test ./tests/... -v 2>&1 | grep -E "^--- FAIL:" | wc -l
