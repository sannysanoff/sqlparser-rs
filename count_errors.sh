#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go
go test ./tests/... 2>&1 | grep -E "Expected:.*found|not yet implemented" | sed 's/.*Expected: //' | sed 's/ not yet implemented//' | sed 's/, found:.*//' | sort | uniq -c | sort -rn | head -30
