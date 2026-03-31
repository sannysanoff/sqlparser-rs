package main

import (
	"fmt"
	"log"

	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
)

func main() {
	// Example 1: Simple SELECT
	sql1 := "SELECT * FROM users WHERE active = true"
	dialect := generic.NewGenericDialect()

	statements, err := parser.ParseSQL(dialect, sql1)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Example 1: Simple SELECT")
	fmt.Println("Input:", sql1)
	fmt.Println("Parsed:", statements[0].String())
	fmt.Println()

	// Example 2: Complex query with JOIN
	sql2 := `
		SELECT u.id, u.name, o.order_date
		FROM users u
		JOIN orders o ON u.id = o.user_id
		WHERE u.active = true
		ORDER BY o.order_date DESC
		LIMIT 10
	`

	statements, err = parser.ParseSQL(dialect, sql2)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Example 2: Complex query with JOIN")
	fmt.Println("Input:", sql2)
	fmt.Println("Parsed:", statements[0].String())
	fmt.Println()

	// Example 3: CREATE TABLE
	sql3 := `
		CREATE TABLE products (
			id INT PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			price DECIMAL(10, 2),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`

	statements, err = parser.ParseSQL(dialect, sql3)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Example 3: CREATE TABLE")
	fmt.Println("Input:", sql3)
	fmt.Println("Parsed:", statements[0].String())
	fmt.Println()

	// Example 4: INSERT
	sql4 := "INSERT INTO users (name, email) VALUES ('John Doe', 'john@example.com')"

	statements, err = parser.ParseSQL(dialect, sql4)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Example 4: INSERT")
	fmt.Println("Input:", sql4)
	fmt.Println("Parsed:", statements[0].String())
	fmt.Println()

	// Example 5: Multiple statements
	sql5 := "SELECT * FROM t1; SELECT * FROM t2; SELECT * FROM t3"

	statements, err = parser.ParseSQL(dialect, sql5)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Example 5: Multiple statements")
	fmt.Printf("Input: %s\n", sql5)
	fmt.Printf("Parsed %d statements:\n", len(statements))
	for i, stmt := range statements {
		fmt.Printf("  %d: %s\n", i+1, stmt.String())
	}
}
