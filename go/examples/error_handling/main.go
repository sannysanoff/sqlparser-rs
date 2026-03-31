package main

import (
	"fmt"
	"log"

	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
)

func main() {
	fmt.Println("Error Handling Example")
	fmt.Println("======================\n")

	dialect := generic.NewGenericDialect()

	// Example 1: Syntax error with location
	fmt.Println("Example 1: Syntax error")
	sql1 := "SELECT * FRO users" // Typo: FRO instead of FROM

	_, err := parser.ParseSQL(dialect, sql1)
	if err != nil {
		fmt.Printf("SQL: %s\n", sql1)
		fmt.Printf("Error: %v\n\n", err)
	}

	// Example 2: Missing comma in SELECT list
	fmt.Println("Example 2: Missing comma")
	sql2 := "SELECT col1 col2 FROM users" // Missing comma between col1 and col2

	_, err = parser.ParseSQL(dialect, sql2)
	if err != nil {
		fmt.Printf("SQL: %s\n", sql2)
		fmt.Printf("Error: %v\n\n", err)
	}

	// Example 3: Unclosed parenthesis
	fmt.Println("Example 3: Unclosed parenthesis")
	sql3 := "SELECT * FROM users WHERE id IN (1, 2, 3" // Missing closing paren

	_, err = parser.ParseSQL(dialect, sql3)
	if err != nil {
		fmt.Printf("SQL: %s\n", sql3)
		fmt.Printf("Error: %v\n\n", err)
	}

	// Example 4: Invalid identifier
	fmt.Println("Example 4: Invalid identifier")
	sql4 := "SELECT 123col FROM users" // Identifier starting with number

	_, err = parser.ParseSQL(dialect, sql4)
	if err != nil {
		fmt.Printf("SQL: %s\n", sql4)
		fmt.Printf("Error: %v\n\n", err)
	}

	// Example 5: Valid SQL (no error)
	fmt.Println("Example 5: Valid SQL (success)")
	sql5 := "SELECT id, name FROM users WHERE active = true"

	statements, err := parser.ParseSQL(dialect, sql5)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("SQL: %s\n", sql5)
	fmt.Printf("Successfully parsed %d statement(s)\n", len(statements))
	fmt.Printf("Regenerated: %s\n\n", statements[0].String())

	fmt.Println("Error handling examples completed!")
}
