#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests
go test ./snowflake/... -v 2>&1 | grep -E "^(--- FAIL:)" | head -50
