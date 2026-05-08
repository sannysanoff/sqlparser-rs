package main

import (
	"fmt"
	"github.com/sannysanoff/sqlparser-rs/go/dialects/snowflake"
	"github.com/sannysanoff/sqlparser-rs/go/parser"
)

func main() {
	sql := "SELECT * FROM db_name..table_name"
	dialect := snowflake.NewSnowflakeDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Printf("Parsed %d statement(s)\n", len(stmts))
	for i, stmt := range stmts {
		fmt.Printf("Statement %d: %s\n", i, stmt.String())
	}
}
