package main

import (
	"fmt"
	"log"

	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
)

func main() {
	fmt.Println("AST Traversal Example")
	fmt.Println("=====================\n")

	sql := `SELECT * FROM users WHERE active = true`

	// Create a generic dialect
	dialect := generic.NewGenericDialect()

	// Parse the SQL
	statements, err := parser.ParseSQL(dialect, sql)
	if err != nil {
		log.Printf("Parsing not yet fully implemented: %v", err)
		fmt.Println("Note: Full parsing implementation is in progress")
		return
	}

	fmt.Printf("Successfully parsed %d statement(s)\n", len(statements))
	for i, stmt := range statements {
		fmt.Printf("Statement %d: %s\n", i+1, stmt.String())
	}
}
