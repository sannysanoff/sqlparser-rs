package main

import (
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/sannysanoff/sqlparser-rs/go/ast"
	"github.com/sannysanoff/sqlparser-rs/go/ast/statement"
	"github.com/sannysanoff/sqlparser-rs/go/dialects/clickhouse"
	"github.com/sannysanoff/sqlparser-rs/go/parser"
)

func main() {
	dialect := clickhouse.NewClickHouseDialect()
	sql := "SELECT * FROM users"
	stmts, err := parser.ParseSQL(dialect, sql)
	if err != nil {
		panic(err)
	}
	for _, stmt := range stmts {
		fmt.Printf("Type: %T\n", stmt)
		v := reflect.ValueOf(stmt)
		if v.Kind() == reflect.Ptr {
			v = v.Elem()
		}
		t := v.Type()
		for i := 0; i < t.NumField(); i++ {
			f := t.Field(i)
			fmt.Printf("  %s (%s)\n", f.Name, f.Type)
		}

		// Try json.Marshal
		b, err := json.MarshalIndent(stmt, "", "  ")
		if err != nil {
			fmt.Printf("  JSON error: %v\n", err)
		} else {
			fmt.Printf("  JSON: %s\n", string(b))
		}
	}
}
