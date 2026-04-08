#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests
go test ./snowflake/... -v -run TestSnowflakeSemiStructuredDataTraversal 2>&1 | tail -60
