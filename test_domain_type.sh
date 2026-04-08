#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests

echo "=== Testing CREATE DOMAIN ==="
go test ./postgres -v -run "TestPostgresCreateDomain" 2>&1

echo ""
echo "=== Testing CREATE TYPE ==="
go test ./postgres -v -run "TestPostgresCreateType" 2>&1
