#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests
go test ./snowflake/... -v -run TestSnowflakeMultiTableInsertUnconditional 2>&1
