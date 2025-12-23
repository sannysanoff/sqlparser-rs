// Licensed to the Apache Software Foundation (ASF) under one
// or more contributor license agreements.  See the NOTICE file
// distributed with this work for additional information
// regarding copyright ownership.  The ASF licenses this file
// to you under the Apache License, Version 2.0 (the
// "License"); you may not use this file except in compliance
// with the License.  You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

//! Parse SQL from stdin and output the AST as JSON
//!
//! Usage:
//!   echo "SELECT * FROM users" | cargo run --features serde,serde_json --example ast_to_json -- generic
//!   cargo run --features serde,serde_json --example ast_to_json -- postgres < query.sql
//!
//! Supported dialects:
//!   ansi, bigquery, clickhouse, databricks, duckdb, generic, hive, mysql, mssql, oracle, postgres, redshift, snowflake, sqlite

use std::io::{self, Read};
use std::process;

use sqlparser::dialect::*;
use sqlparser::parser::Parser;

fn main() {
    // Get dialect from command line argument
    let dialect_arg = std::env::args()
        .nth(1)
        .unwrap_or_else(|| {
            eprintln!("Error: Dialect argument required");
            eprintln!();
            eprintln!("Usage: ast_to_json <dialect>");
            eprintln!();
            eprintln!("Supported dialects:");
            eprintln!("  ansi, bigquery, clickhouse, databricks, duckdb, generic,");
            eprintln!("  hive, mysql, mssql, oracle, postgres, redshift, snowflake, sqlite");
            eprintln!();
            eprintln!("Example:");
            eprintln!("  echo \"SELECT * FROM users\" | cargo run --features serde,serde_json --example ast_to_json -- generic");
            process::exit(1);
        })
        .to_lowercase();

    // Create dialect based on argument
    let dialect: Box<dyn Dialect> = match dialect_arg.as_str() {
        "ansi" => Box::new(AnsiDialect {}),
        "bigquery" => Box::new(BigQueryDialect {}),
        "clickhouse" => Box::new(ClickHouseDialect {}),
        "databricks" => Box::new(DatabricksDialect {}),
        "duckdb" => Box::new(DuckDbDialect {}),
        "generic" => Box::new(GenericDialect {}),
        "hive" => Box::new(HiveDialect {}),
        "mysql" => Box::new(MySqlDialect {}),
        "mssql" => Box::new(MsSqlDialect {}),
        "oracle" => Box::new(OracleDialect {}),
        "postgres" | "postgresql" => Box::new(PostgreSqlDialect {}),
        "redshift" => Box::new(RedshiftSqlDialect {}),
        "snowflake" => Box::new(SnowflakeDialect {}),
        "sqlite" => Box::new(SQLiteDialect {}),
        _ => {
            eprintln!("Error: Unknown dialect '{}'", dialect_arg);
            eprintln!();
            eprintln!("Supported dialects:");
            eprintln!("  ansi, bigquery, clickhouse, databricks, duckdb, generic,");
            eprintln!("  hive, mysql, mssql, oracle, postgres, redshift, snowflake, sqlite");
            process::exit(1);
        }
    };

    // Read SQL from stdin
    let mut sql = String::new();
    io::stdin()
        .read_to_string(&mut sql)
        .unwrap_or_else(|e| {
            eprintln!("Error reading from stdin: {}", e);
            process::exit(1);
        });

    // Remove BOM if present
    let sql = if sql.starts_with('\u{feff}') {
        &sql[3..]
    } else {
        &sql
    };

    // Parse SQL
    let parse_result = Parser::parse_sql(&*dialect, sql);

    match parse_result {
        Ok(statements) => {
            // Serialize to JSON with pretty printing
            match serde_json::to_string_pretty(&statements) {
                Ok(json) => {
                    println!("{}", json);
                    process::exit(0);
                }
                Err(e) => {
                    eprintln!("Error serializing to JSON: {}", e);
                    process::exit(1);
                }
            }
        }
        Err(e) => {
            eprintln!("Parse error: {}", e);
            process::exit(1);
        }
    }
}
