#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go
go test ./... -v 2>&1 | grep -E "^(=== RUN|--- (PASS|FAIL))" | head -200
