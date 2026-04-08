#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests

echo "=== Testing specific CREATE TYPE and DOMAIN tests ==="
go test ./postgres -v -run "TestPostgresCreateTypeAsEnum" 2>&1 | head -30
