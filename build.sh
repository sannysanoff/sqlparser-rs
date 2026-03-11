cargo build --release --example ast_to_json --features serde_json,serde
mv target/release/examples/ast_to_json ./sql_ast_to_json
