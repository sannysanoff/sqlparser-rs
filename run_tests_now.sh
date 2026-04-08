#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests
go test ./... -v 2>&1 | tee /tmp/test_output.txt
echo "=== Test Summary ==="
grep -c "^--- PASS:" /tmp/test_output.txt
grep -c "^--- FAIL:" /tmp/test_output.txt
