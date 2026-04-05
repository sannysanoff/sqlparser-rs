#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go
echo "=== Test Summary ==="
go test ./tests/... 2>&1 | grep -E "(^ok|^FAIL)"
