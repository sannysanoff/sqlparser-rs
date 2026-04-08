#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests
go test ./... 2>&1 | grep -E "^(ok|FAIL|---)" | head -100
