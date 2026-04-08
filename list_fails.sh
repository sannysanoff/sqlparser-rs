#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests

echo "=== All failing tests ==="
go test ./... -v 2>&1 | grep "^--- FAIL:" | sed 's/--- FAIL: //;s/ (.*//' | sort
