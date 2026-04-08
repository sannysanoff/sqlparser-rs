package main

import (
	"fmt"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
)

func main() {
	sql := "ALTER USER u1 SET DEFAULT_MFA_METHOD='PASSKEY'"
	dialect := generic.NewGenericDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Parsed: %d statements\n", len(stmts))
		for i, stmt := range stmts {
			fmt.Printf("%d: %T - %s\n", i, stmt, stmt.String())
		}
	}
}
