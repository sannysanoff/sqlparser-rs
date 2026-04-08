#!/bin/bash
cd /Users/san/Fun/sqlparser-rs/go/tests

OUTPUT=$(go test ./... -v 2>&1)
PASS_COUNT=$(echo "$OUTPUT" | grep -c "^--- PASS:")
FAIL_COUNT=$(echo "$OUTPUT" | grep -c "^--- FAIL:")
SKIP_COUNT=$(echo "$OUTPUT" | grep -c "^--- SKIP:")
TOTAL=$((PASS_COUNT + FAIL_COUNT))

if [ "$TOTAL" -gt 0 ]; then
    PASS_RATE=$(echo "scale=1; $PASS_COUNT * 100 / $TOTAL" | bc 2>/dev/null || echo "N/A")
else
    PASS_RATE="0.0"
fi

echo ""
echo "=== Test Summary ==="
echo "Passing:  $PASS_COUNT"
echo "Failing:  $FAIL_COUNT"
echo "Skipped:  $SKIP_COUNT"
echo "Total:    $TOTAL"
echo "Pass Rate: ${PASS_RATE}%"
