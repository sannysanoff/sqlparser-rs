package main

import (
	"fmt"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects/generic"
	"github.com/user/sqlparser/parser"
)

func main() {
	testCases := []string{
		"ALTER TABLE customer ADD COLUMN email VARCHAR(255)",
		"ALTER TABLE tab ADD COLUMN foo TEXT",
		"ALTER TABLE tab DROP COLUMN is_active CASCADE",
		"ALTER TABLE tab RENAME TO new_tab",
		"ALTER TABLE tab RENAME COLUMN foo TO new_foo",
		"ALTER TABLE tab ALTER COLUMN is_active SET NOT NULL",
		"ALTER TABLE tab ALTER COLUMN is_active SET DEFAULT 0",
		"ALTER TABLE tab ALTER COLUMN is_active DROP DEFAULT",
		"ALTER TABLE tab ALTER COLUMN is_active SET DATA TYPE TEXT",
		"ALTER TABLE tab ADD CONSTRAINT address_pkey PRIMARY KEY (address_id)",
		"ALTER TABLE tab DROP CONSTRAINT IF EXISTS constraint_name RESTRICT",
		"ALTER TABLE tab ENABLE TRIGGER ALL",
		"ALTER TABLE tab DISABLE ROW LEVEL SECURITY",
	}

	dialect := generic.NewGenericDialect()
	for _, sql := range testCases {
		stmts, err := parser.ParseSQL(dialect, sql)
		if err != nil {
			fmt.Printf("FAIL: %s\n  Error: %v\n", sql, err)
		} else {
			alterTable, ok := stmts[0].(*statement.AlterTable)
			if !ok {
				fmt.Printf("FAIL: %s\n  Not an AlterTable statement\n", sql)
			} else {
				fmt.Printf("PASS: %s\n  Table: %s, Operations: %d\n", sql, alterTable.Name.String(), len(alterTable.Operations))
			}
		}
	}
}
