#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests

OUTPUT=$(go test ./... 2>&1)
echo "$OUTPUT" | grep -E "^(ok|FAIL)"
