#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests

echo "=== Top failing test patterns ==="
go test ./... -v 2>&1 | grep "^--- FAIL:" | sed 's/--- FAIL: Test//;s/ (.*/ /' | sort | uniq -c | sort -rn | head -30
