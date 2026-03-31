package main

import (
	"fmt"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
)

func main() {
	sql := "SELECT sum(l_extendedprice * (1 - l_discount)) FROM t"
	dialect := generic.NewGenericDialect()
	stmts, err := parser.ParseSQL(dialect, sql)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}
	fmt.Printf("Parsed %d statements\n", len(stmts))
	for i, stmt := range stmts {
		fmt.Printf("Statement %d: %s\n", i, stmt.String())
	}
}
