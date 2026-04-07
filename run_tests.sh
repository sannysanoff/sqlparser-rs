#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests
echo "=== Running Tests ==="
go test -v -count=1 . 2>&1
