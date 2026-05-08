package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/user/sqlparser/astjson"
	"github.com/user/sqlparser/dialects/ansi"
	"github.com/user/sqlparser/dialects/bigquery"
	"github.com/user/sqlparser/dialects/clickhouse"
	"github.com/user/sqlparser/dialects/databricks"
	"github.com/user/sqlparser/dialects/duckdb"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/dialects/hive"
	"github.com/user/sqlparser/dialects/mssql"
	"github.com/user/sqlparser/dialects/mysql"
	"github.com/user/sqlparser/dialects/oracle"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/redshift"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/dialects/sqlite"
	"github.com/user/sqlparser/parser"
	"github.com/user/sqlparser/parseriface"
)

func main() {
	// Get dialect from command line argument
	dialectArg := ""
	if len(os.Args) > 1 {
		dialectArg = strings.ToLower(os.Args[1])
	}

	if dialectArg == "" {
		fmt.Fprintln(os.Stderr, "Error: Dialect argument required")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Usage: ast_to_json <dialect>")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Supported dialects:")
		fmt.Fprintln(os.Stderr, "  ansi, bigquery, clickhouse, databricks, duckdb, generic,")
		fmt.Fprintln(os.Stderr, "  hive, mysql, mssql, oracle, postgres, redshift, snowflake, sqlite")
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Example:")
		fmt.Fprintln(os.Stderr, "  echo \"SELECT * FROM users\" | ast_to_json clickhouse")
		os.Exit(1)
	}

	// Create dialect based on argument
	var dialect parseriface.CompleteDialect
	switch dialectArg {
	case "ansi":
		dialect = ansi.NewAnsiDialect()
	case "bigquery":
		dialect = bigquery.NewBigQueryDialect()
	case "clickhouse":
		dialect = clickhouse.NewClickHouseDialect()
	case "databricks":
		dialect = databricks.NewDatabricksDialect()
	case "duckdb":
		dialect = duckdb.NewDuckDbDialect()
	case "generic":
		dialect = generic.NewGenericDialect()
	case "hive":
		dialect = hive.NewHiveDialect()
	case "mysql":
		dialect = mysql.NewMySqlDialect()
	case "mssql":
		dialect = mssql.NewMsSqlDialect()
	case "oracle":
		dialect = oracle.NewOracleDialect()
	case "postgres", "postgresql":
		dialect = postgresql.NewPostgreSqlDialect()
	case "redshift":
		dialect = redshift.NewRedshiftSqlDialect()
	case "snowflake":
		dialect = snowflake.NewSnowflakeDialect()
	case "sqlite":
		dialect = sqlite.NewSQLiteDialect()
	default:
		fmt.Fprintf(os.Stderr, "Error: Unknown dialect '%s'\n", dialectArg)
		fmt.Fprintln(os.Stderr)
		fmt.Fprintln(os.Stderr, "Supported dialects:")
		fmt.Fprintln(os.Stderr, "  ansi, bigquery, clickhouse, databricks, duckdb, generic,")
		fmt.Fprintln(os.Stderr, "  hive, mysql, mssql, oracle, postgres, redshift, snowflake, sqlite")
		os.Exit(1)
	}

	// Read SQL from stdin
	sqlBytes, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error reading from stdin: %v\n", err)
		os.Exit(1)
	}

	sql := string(sqlBytes)

	// Remove BOM if present
	if len(sql) > 3 && sql[0] == 0xEF && sql[1] == 0xBB && sql[2] == 0xBF {
		sql = sql[3:]
	}

	// Parse SQL
	statements, err := parser.ParseSQL(dialect, sql)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Parse error: %v\n", err)
		os.Exit(1)
	}

	// Serialize to JSON
	jsonBytes, err := astjson.MarshalStatements(statements)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error serializing to JSON: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(string(jsonBytes))
}
