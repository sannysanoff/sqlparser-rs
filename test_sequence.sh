#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests

echo "=== Testing CREATE SEQUENCE ==="
go test ./postgres -v -run "TestPostgresCreateSequence" 2>&1
