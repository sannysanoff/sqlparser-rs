#!/bin/bash
# Comparison test runner for Go vs Rust ast_to_json
set -e

GO_BIN="/home/hermes-work/Work/solidus/sqlparser-rs/go/ast_to_json_bin"
RUST_BIN="/home/hermes-work/Work/solidus/sqlparser-rs/target/release/examples/ast_to_json"
QUERIES_FILE="/home/hermes-work/Work/solidus/sqlparser-rs/go/test_working_queries.sql"
OUTPUT_DIR="/home/hermes-work/Work/solidus/sqlparser-rs/go/test_output"
mkdir -p "$OUTPUT_DIR"

echo "=========================================="
echo "  Go vs Rust ast_to_json Comparison"
echo "=========================================="
echo ""

n=0
pass=0
fail=0

while IFS= read -r query; do
  [ -z "$query" ] && continue
  n=$((n + 1))
  
  echo "--- Test $n: $query"
  
  # Run Go
  go_out=$(echo "$query" | PATH="/home/hermes-work/local/go/bin:$PATH" "$GO_BIN" clickhouse 2>&1)
  go_exit=$?
  
  # Run Rust
  rust_out=$(echo "$query" | "$RUST_BIN" clickhouse 2>&1)
  rust_exit=$?
  
  echo "$go_out" > "$OUTPUT_DIR/go_test${n}.json"
  echo "$rust_out" > "$OUTPUT_DIR/rust_test${n}.json"
  
  if [ $go_exit -eq 0 ]; then
    echo "   Go: OK  (output saved)"
  else
    echo "   Go: FAIL - $(echo "$go_out" | head -1)"
  fi
  
  if [ $rust_exit -eq 0 ]; then
    echo "   Rust: OK  (output saved)"
  else
    echo "   Rust: FAIL - $(echo "$rust_out" | head -1)"
  fi
  
  if [ $go_exit -eq 0 ] && [ $rust_exit -eq 0 ]; then
    pass=$((pass + 1))
    # Show a summary of what was captured
    echo "   Key fields in Go output:"
    echo "$go_out" | grep -o '"[a-z_]*"' | tr -d '"' | sort -u | head -10 | sed 's/^/     - /'
    echo ""
  else
    fail=$((fail + 1))
    echo ""
  fi
done < "$QUERIES_FILE"

echo "=========================================="
echo "Results: $pass parsed by both, $n total"
echo "=========================================="
