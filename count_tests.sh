#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go

OUTPUT=$(go test ./tests/... -v 2>&1)
PASS=$(echo "$OUTPUT" | grep -c "^--- PASS:")
FAIL=$(echo "$OUTPUT" | grep -c "^--- FAIL:")
TOTAL=$((PASS + FAIL))

if [ "$TOTAL" -gt 0 ]; then
    RATE=$(echo "scale=1; $PASS * 100 / $TOTAL" | bc)
else
    RATE="0.0"
fi

echo "=== Test Summary ==="
echo "Passing:  $PASS"
echo "Failing:  $FAIL"
echo "Total:    $TOTAL"
echo "Pass Rate: ${RATE}%"
