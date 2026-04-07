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

package parser

import (
	"fmt"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/token"
)

// parseAlter parses ALTER statements
// Reference: src/parser/mod.rs:10579
func parseAlter(p *Parser) (ast.Statement, error) {
	// Check for various ALTER object types
	if p.ParseKeyword("TABLE") {
		return parseAlterTable(p)
	}
	if p.ParseKeyword("VIEW") {
		return parseAlterView(p)
	}
	if p.ParseKeyword("INDEX") {
		return parseAlterIndex(p)
	}
	if p.ParseKeyword("ROLE") {
		return parseAlterRole(p)
	}
	if p.ParseKeyword("USER") {
		return parseAlterUser(p)
	}
	if p.ParseKeyword("SCHEMA") {
		return parseAlterSchema(p)
	}
	if p.ParseKeyword("TYPE") {
		return parseAlterType(p)
	}
	if p.ParseKeyword("POLICY") {
		return parseAlterPolicy(p)
	}
	if p.ParseKeyword("CONNECTOR") {
		return parseAlterConnector(p)
	}
	if p.ParseKeyword("ICEBERG") {
		// ICEBERG TABLE
		if p.ParseKeyword("TABLE") {
			return parseAlterTable(p) // Iceberg flag would be handled in AST
		}
		return nil, fmt.Errorf("expected TABLE after ALTER ICEBERG")
	}
	if p.ParseKeyword("OPERATOR") {
		if p.ParseKeyword("FAMILY") {
			return parseAlterOperatorFamily(p)
		}
		if p.ParseKeyword("CLASS") {
			return parseAlterOperatorClass(p)
		}
		return parseAlterOperator(p)
	}
	return nil, fmt.Errorf("expected TABLE, VIEW, INDEX, ROLE, USER, SCHEMA, TYPE, POLICY, or CONNECTOR after ALTER")
}

// parseAlterTable parses ALTER TABLE statements
func parseAlterTable(p *Parser) (ast.Statement, error) {
	alterTable := &statement.AlterTable{}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected table name after ALTER TABLE: %w", err)
	}
	alterTable.Name = tableName

	// Parse operations (MySQL supports multiple operations separated by commas)
	for {
		op, err := parseAlterTableOperation(p)
		if err != nil {
			return nil, err
		}
		alterTable.Operations = append(alterTable.Operations, op)

		// Check for comma separator (MySQL allows multiple operations)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return alterTable, nil
}

// parseAlterTableOperation parses a single ALTER TABLE operation
func parseAlterTableOperation(p *Parser) (*expr.AlterTableOperation, error) {
	op := &expr.AlterTableOperation{}

	if p.ParseKeyword("ADD") {
		return parseAlterTableAdd(p, op)
	}

	if p.ParseKeyword("DROP") {
		return parseAlterTableDrop(p, op)
	}

	if p.ParseKeyword("RENAME") {
		return parseAlterTableRename(p, op)
	}

	if p.ParseKeyword("ALTER") {
		return parseAlterTableAlterColumn(p, op)
	}

	if p.ParseKeyword("CHANGE") {
		return parseAlterTableChangeColumn(p, op)
	}

	if p.ParseKeyword("MODIFY") {
		return parseAlterTableModifyColumn(p, op)
	}

	if p.ParseKeyword("SET") {
		return parseAlterTableSet(p, op)
	}

	// MySQL table options: AUTO_INCREMENT, ALGORITHM, LOCK
	if p.PeekKeyword("AUTO_INCREMENT") || p.PeekKeyword("ALGORITHM") || p.PeekKeyword("LOCK") {
		return parseAlterTableMySqlOptions(p, op)
	}

	return nil, fmt.Errorf("unknown ALTER TABLE operation")
}

// parseAlterTableAdd parses ADD COLUMN or ADD CONSTRAINT
// Reference: src/parser/mod.rs:10073-10123
func parseAlterTableAdd(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	// First, check if this is a constraint (ADD CONSTRAINT, ADD PRIMARY KEY, ADD UNIQUE, etc.)
	// Try to peek if this looks like a constraint by checking for constraint keywords
	if looksLikeTableConstraint(p) {
		return parseAlterTableAddConstraint(p, op)
	}

	// Check for COLUMN keyword
	if p.ParseKeyword("COLUMN") {
		op.AddColumnKeyword = true
	}

	// Check for IF NOT EXISTS
	if p.ParseKeywords([]string{"IF", "NOT", "EXISTS"}) {
		op.AddIfNotExists = true
	}

	// Check for parenthesized column list (MySQL style: ADD COLUMN (c1 INT, c2 INT))
	if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
		p.AdvanceToken() // consume (

		op.Op = expr.AlterTableOpAddColumn

		for {
			colDef, err := parseColumnDef(p)
			if err != nil {
				return nil, fmt.Errorf("expected column definition: %w", err)
			}
			op.AddColumnDefs = append(op.AddColumnDefs, colDef)

			// Check for comma separator
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}

		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return op, nil
	}

	// Parse single column definition
	colDef, err := parseColumnDef(p)
	if err != nil {
		return nil, fmt.Errorf("expected column definition: %w", err)
	}
	op.Op = expr.AlterTableOpAddColumn
	op.AddColumnDef = colDef

	// Check for MySQL column position (FIRST or AFTER column)
	if p.ParseKeyword("FIRST") {
		op.AddColumnPosition = &expr.MySQLColumnPosition{IsFirst: true}
	} else if p.ParseKeyword("AFTER") {
		afterCol, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected column name after AFTER: %w", err)
		}
		op.AddColumnPosition = &expr.MySQLColumnPosition{
			AfterColumn: afterCol,
		}
	}

	return op, nil
}

// looksLikeTableConstraint checks if the current position looks like a table constraint
// This follows Rust's pattern of checking for constraint keywords without consuming tokens
// Reference: src/parser/mod.rs:10074-10075
func looksLikeTableConstraint(p *Parser) bool {
	// Check for CONSTRAINT keyword
	if p.PeekKeyword("CONSTRAINT") {
		return true
	}

	// Check for PRIMARY KEY
	if p.PeekKeyword("PRIMARY") {
		return true
	}

	// Check for UNIQUE
	if p.PeekKeyword("UNIQUE") {
		return true
	}

	// Check for FOREIGN KEY
	if p.PeekKeyword("FOREIGN") {
		return true
	}

	// Check for CHECK
	if p.PeekKeyword("CHECK") {
		return true
	}

	// MySQL-specific: INDEX, KEY, FULLTEXT, SPATIAL
	if p.GetDialect().SupportsIndexHints() {
		if p.PeekKeyword("INDEX") || p.PeekKeyword("KEY") ||
			p.PeekKeyword("FULLTEXT") || p.PeekKeyword("SPATIAL") {
			return true
		}
	}

	return false
}

// parseAlterTableAddConstraint parses ADD CONSTRAINT operations
// Reference: src/parser/mod.rs:10074-10081
func parseAlterTableAddConstraint(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpAddConstraint

	// Parse the constraint definition
	constraint, err := parseTableConstraint(p)
	if err != nil {
		return nil, fmt.Errorf("expected constraint definition: %w", err)
	}
	op.Constraint = constraint

	// Check for NOT VALID (PostgreSQL deferrable constraints)
	// Reference: src/parser/mod.rs:10076
	if p.ParseKeywords([]string{"NOT", "VALID"}) {
		op.ConstraintNotValid = true
	}

	return op, nil
}

// parseAlterTableDrop parses DROP COLUMN, DROP CONSTRAINT, DROP PRIMARY KEY, DROP FOREIGN KEY, DROP INDEX
func parseAlterTableDrop(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	// Check for CONSTRAINT
	if p.ParseKeyword("CONSTRAINT") {
		return parseAlterTableDropConstraint(p, op)
	}

	// Check for PRIMARY KEY (MySQL)
	if p.ParseKeyword("PRIMARY") {
		if p.ParseKeyword("KEY") {
			op.Op = expr.AlterTableOpDropPrimaryKey
			// Check for CASCADE/RESTRICT
			if p.ParseKeyword("CASCADE") {
				op.DropBehavior = expr.DropBehaviorCascade
			} else if p.ParseKeyword("RESTRICT") {
				op.DropBehavior = expr.DropBehaviorRestrict
			}
			return op, nil
		}
		return nil, fmt.Errorf("expected KEY after PRIMARY")
	}

	// Check for FOREIGN KEY (MySQL)
	if p.ParseKeyword("FOREIGN") {
		if p.ParseKeyword("KEY") {
			op.Op = expr.AlterTableOpDropForeignKey
			// Parse foreign key name
			name, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected foreign key name: %w", err)
			}
			op.DropColumnNames = append(op.DropColumnNames, name)
			// Check for CASCADE/RESTRICT
			if p.ParseKeyword("CASCADE") {
				op.DropBehavior = expr.DropBehaviorCascade
			} else if p.ParseKeyword("RESTRICT") {
				op.DropBehavior = expr.DropBehaviorRestrict
			}
			return op, nil
		}
		return nil, fmt.Errorf("expected KEY after FOREIGN")
	}

	// Check for INDEX (MySQL)
	if p.ParseKeyword("INDEX") {
		op.Op = expr.AlterTableOpDropIndex
		// Parse index name
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected index name: %w", err)
		}
		op.DropColumnNames = append(op.DropColumnNames, name)
		return op, nil
	}

	// Check for COLUMN keyword
	if p.ParseKeyword("COLUMN") {
		op.DropColumnKeyword = true
	}

	// Check for IF EXISTS
	if p.ParseKeywords([]string{"IF", "EXISTS"}) {
		op.DropIfExists = true
	}

	// Parse column name(s)
	colName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected column name: %w", err)
	}
	op.Op = expr.AlterTableOpDropColumn
	op.DropColumnNames = append(op.DropColumnNames, colName)

	// Check for more column names (comma-separated) - some dialects support this
	for {
		if _, isComma := p.PeekToken().Token.(token.TokenComma); !isComma {
			break
		}
		p.NextToken()
		colName, err = p.ParseIdentifier()
		if err != nil {
			break
		}
		op.DropColumnNames = append(op.DropColumnNames, colName)
	}

	// Check for CASCADE/RESTRICT
	if p.ParseKeyword("CASCADE") {
		op.DropBehavior = expr.DropBehaviorCascade
	} else if p.ParseKeyword("RESTRICT") {
		op.DropBehavior = expr.DropBehaviorRestrict
	}

	return op, nil
}

// parseAlterTableDropConstraint parses DROP CONSTRAINT
func parseAlterTableDropConstraint(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpDropConstraint

	// Check for IF EXISTS
	if p.ParseKeywords([]string{"IF", "EXISTS"}) {
		op.DropConstraintIfExists = true
	}

	// Parse constraint name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected constraint name: %w", err)
	}
	op.DropConstraintName = name

	// Check for CASCADE/RESTRICT
	if p.ParseKeyword("CASCADE") {
		op.DropBehavior = expr.DropBehaviorCascade
	} else if p.ParseKeyword("RESTRICT") {
		op.DropBehavior = expr.DropBehaviorRestrict
	}

	return op, nil
}

// parseAlterTableRename parses RENAME operations
func parseAlterTableRename(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	// RENAME TO <new_table_name>
	if p.ParseKeyword("TO") {
		return parseAlterTableRenameTable(p, op)
	}

	// RENAME AS <new_table_name> (alternative syntax)
	if p.ParseKeyword("AS") {
		return parseAlterTableRenameTable(p, op)
	}

	// RENAME [COLUMN] <old_name> TO <new_name>
	return parseAlterTableRenameColumn(p, op)
}

// parseAlterTableRenameTable parses RENAME TO/AS <new_table_name>
func parseAlterTableRenameTable(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpRenameTable

	newName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected table name: %w", err)
	}
	op.NewTableName = newName

	return op, nil
}

// parseAlterTableRenameColumn parses RENAME [COLUMN] <old_name> TO <new_name>
func parseAlterTableRenameColumn(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpRenameColumn

	// Check for COLUMN keyword (optional)
	p.ParseKeyword("COLUMN")

	// Parse old column name
	oldName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected column name: %w", err)
	}
	op.RenameOldColumn = oldName

	// Expect TO
	if !p.ParseKeyword("TO") {
		return nil, fmt.Errorf("expected TO after column name")
	}

	// Parse new column name
	newName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected new column name: %w", err)
	}
	op.RenameNewColumn = newName

	return op, nil
}

// parseAlterTableAlterColumn parses ALTER [COLUMN] <column_name> <operation>
func parseAlterTableAlterColumn(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpAlterColumn

	// Check for COLUMN keyword (optional)
	if p.ParseKeyword("COLUMN") {
		// COLUMN is optional in many dialects
	}

	// Parse column name
	colName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected column name: %w", err)
	}
	op.AlterColumnName = colName

	// Parse the column operation
	if p.ParseKeywords([]string{"SET", "NOT", "NULL"}) {
		op.AlterColumnOp = expr.AlterColumnOpSetNotNull
	} else if p.ParseKeywords([]string{"DROP", "NOT", "NULL"}) {
		op.AlterColumnOp = expr.AlterColumnOpDropNotNull
	} else if p.ParseKeywords([]string{"SET", "DEFAULT"}) {
		op.AlterColumnOp = expr.AlterColumnOpSetDefault
		// Parse default value
		exprParser := NewExpressionParser(p)
		defaultExpr, err := exprParser.ParseExpr()
		if err == nil {
			op.AlterDefault = defaultExpr
		}
	} else if p.ParseKeywords([]string{"DROP", "DEFAULT"}) {
		op.AlterColumnOp = expr.AlterColumnOpDropDefault
	} else if p.ParseKeywords([]string{"SET", "DATA", "TYPE"}) {
		op.AlterColumnOp = expr.AlterColumnOpSetDataType
		// Parse data type
		dataType, err := p.ParseDataType()
		if err == nil {
			op.AlterDataType = dataType
		}
	} else if p.ParseKeyword("TYPE") {
		op.AlterColumnOp = expr.AlterColumnOpSetDataType
		// Parse data type
		dataType, err := p.ParseDataType()
		if err == nil {
			op.AlterDataType = dataType
		}
	} else {
		return nil, fmt.Errorf("expected SET NOT NULL, DROP NOT NULL, SET DEFAULT, DROP DEFAULT, or SET DATA TYPE after ALTER COLUMN")
	}

	return op, nil
}

// parseAlterTableChangeColumn parses CHANGE [COLUMN] <old_name> <new_name> <data_type> [options] [position]
// MySQL-specific syntax
func parseAlterTableChangeColumn(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpChangeColumn

	// Optional COLUMN keyword
	p.ParseKeyword("COLUMN")

	// Parse old column name
	oldName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected old column name: %w", err)
	}
	op.ChangeOldName = oldName

	// Parse new column name
	newName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected new column name: %w", err)
	}
	op.ChangeNewName = newName

	// Parse data type
	dataType, err := p.ParseDataType()
	if err != nil {
		return nil, fmt.Errorf("expected data type: %w", err)
	}
	op.ChangeDataType = dataType

	// Parse optional column options
	for {
		colOpt, err := parseOptionalColumnOption(p)
		if err != nil || colOpt == nil {
			break
		}
		op.ChangeOptions = append(op.ChangeOptions, colOpt)
	}

	// Parse optional column position (FIRST or AFTER column)
	if p.ParseKeyword("FIRST") {
		op.ChangeColumnPosition = &expr.MySQLColumnPosition{IsFirst: true}
	} else if p.ParseKeyword("AFTER") {
		afterCol, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected column name after AFTER: %w", err)
		}
		op.ChangeColumnPosition = &expr.MySQLColumnPosition{AfterColumn: afterCol}
	}

	return op, nil
}

// parseAlterTableModifyColumn parses MODIFY [COLUMN] <col_name> <data_type> [options] [position]
// MySQL-specific syntax
func parseAlterTableModifyColumn(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpModifyColumn

	// Optional COLUMN keyword
	p.ParseKeyword("COLUMN")

	// Parse column name
	colName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected column name: %w", err)
	}
	op.ModifyColumnName = colName

	// Parse data type
	dataType, err := p.ParseDataType()
	if err != nil {
		return nil, fmt.Errorf("expected data type: %w", err)
	}
	op.ModifyDataType = dataType

	// Parse optional column options
	for {
		colOpt, err := parseOptionalColumnOption(p)
		if err != nil || colOpt == nil {
			break
		}
		op.ModifyOptions = append(op.ModifyOptions, colOpt)
	}

	// Parse optional column position (FIRST or AFTER column)
	if p.ParseKeyword("FIRST") {
		op.ModifyColumnPosition = &expr.MySQLColumnPosition{IsFirst: true}
	} else if p.ParseKeyword("AFTER") {
		afterCol, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected column name after AFTER: %w", err)
		}
		op.ModifyColumnPosition = &expr.MySQLColumnPosition{AfterColumn: afterCol}
	}

	return op, nil
}

// parseOptionalColumnOption attempts to parse a column option (e.g., NOT NULL, DEFAULT, etc.)
// Returns nil if no option is found
func parseOptionalColumnOption(p *Parser) (*expr.ColumnOption, error) {
	// Check for various column options
	if p.ParseKeyword("NOT") {
		if p.ParseKeyword("NULL") {
			return &expr.ColumnOption{Name: "NOT NULL"}, nil
		}
		// Put back NOT if not followed by NULL
		// TODO: need a way to put back keywords
	}

	if p.ParseKeyword("NULL") {
		return &expr.ColumnOption{Name: "NULL"}, nil
	}

	if p.ParseKeyword("DEFAULT") {
		// Parse default expression
		exprParser := NewExpressionParser(p)
		defaultExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, err
		}
		return &expr.ColumnOption{
			Name:  "DEFAULT",
			Value: defaultExpr,
		}, nil
	}

	if p.ParseKeyword("AUTO_INCREMENT") || p.ParseKeyword("AUTOINCREMENT") {
		return &expr.ColumnOption{Name: "AUTO_INCREMENT"}, nil
	}

	if p.ParseKeyword("PRIMARY") && p.ParseKeyword("KEY") {
		return &expr.ColumnOption{Name: "PRIMARY KEY"}, nil
	}

	if p.ParseKeyword("UNIQUE") {
		p.ParseKeyword("KEY") // optional
		return &expr.ColumnOption{Name: "UNIQUE"}, nil
	}

	if p.ParseKeyword("COMMENT") {
		// Parse comment string
		tok := p.NextToken()
		switch t := tok.Token.(type) {
		case token.TokenSingleQuotedString:
			return &expr.ColumnOption{
				Name:  "COMMENT",
				Value: &expr.ValueExpr{Value: t.Value},
			}, nil
		case token.TokenDoubleQuotedString:
			return &expr.ColumnOption{
				Name:  "COMMENT",
				Value: &expr.ValueExpr{Value: t.Value},
			}, nil
		default:
			return nil, fmt.Errorf("expected string after COMMENT")
		}
	}

	// TODO: Add more column options as needed

	return nil, nil
}

// parseAlterTableSet parses SET TBLPROPERTIES or SET options
func parseAlterTableSet(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	// SET TBLPROPERTIES(...)
	if p.ParseKeyword("TBLPROPERTIES") {
		return parseAlterTableSetTblProperties(p, op)
	}

	// SET (...) - parenthesized options - just skip for now
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		op.Op = expr.AlterTableOpSetOptionsParens
		// Consume until matching RParen
		depth := 1
		for depth > 0 {
			tok := p.NextToken()
			switch tok.Token.(type) {
			case token.TokenLParen:
				depth++
			case token.TokenRParen:
				depth--
			case token.EOF:
				return nil, fmt.Errorf("unexpected end of input in SET options")
			}
		}
		return op, nil
	}

	return nil, fmt.Errorf("expected TBLPROPERTIES or parenthesized options after SET")
}

// parseAlterTableSetTblProperties parses SET TBLPROPERTIES('key' = 'value', ...)
func parseAlterTableSetTblProperties(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpSetTblProperties

	// Expect opening parenthesis
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse key-value pairs
	for {
		// Parse key (quoted string or identifier)
		tok := p.NextToken()
		var key string
		switch t := tok.Token.(type) {
		case token.TokenSingleQuotedString:
			key = t.Value
		case token.TokenDoubleQuotedString:
			key = t.Value
		case token.TokenWord:
			key = t.Value
		default:
			return nil, fmt.Errorf("expected property key")
		}

		// Expect =
		nextTok := p.NextToken()
		if _, ok := nextTok.Token.(token.TokenEq); !ok {
			// Put back the token if it's not =
			p.PrevToken()
			// Some dialects allow key without = value, defaulting to true
			keyIdent := &expr.Ident{Value: key}
			op.TblProperties = append(op.TblProperties, &expr.SqlOption{
				Name:  keyIdent,
				Value: &expr.ValueExpr{Value: true},
			})
		} else {
			// Parse value
			tok = p.NextToken()
			var value expr.Expr
			switch t := tok.Token.(type) {
			case token.TokenSingleQuotedString:
				value = &expr.ValueExpr{Value: t.Value}
			case token.TokenDoubleQuotedString:
				value = &expr.ValueExpr{Value: t.Value}
			case token.TokenNumber:
				value = &expr.ValueExpr{Value: t.Value}
			case token.TokenWord:
				if t.Value == "TRUE" || t.Value == "true" {
					value = &expr.ValueExpr{Value: true}
				} else if t.Value == "FALSE" || t.Value == "false" {
					value = &expr.ValueExpr{Value: false}
				} else {
					value = &expr.ValueExpr{Value: t.Value}
				}
			default:
				return nil, fmt.Errorf("expected property value")
			}
			keyIdent := &expr.Ident{Value: key}
			op.TblProperties = append(op.TblProperties, &expr.SqlOption{
				Name:  keyIdent,
				Value: value,
			})
		}

		// Check for more properties
		if _, isComma := p.PeekToken().Token.(token.TokenComma); !isComma {
			break
		}
		p.NextToken()
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return op, nil
}

// parseAlterTableMySqlOptions parses MySQL-specific ALTER TABLE options (AUTO_INCREMENT, ALGORITHM, LOCK)
func parseAlterTableMySqlOptions(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpSetOptions

	if p.ParseKeyword("AUTO_INCREMENT") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			// Allow without equals sign too
		}
		tok := p.NextToken()
		if num, ok := tok.Token.(token.TokenNumber); ok {
			op.AutoIncrementValue = num.Value
		} else {
			return nil, fmt.Errorf("expected number after AUTO_INCREMENT")
		}
	} else if p.ParseKeyword("ALGORITHM") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
		algo, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected algorithm after ALGORITHM=: %w", err)
		}
		op.AlgorithmValue = algo
	} else if p.ParseKeyword("LOCK") {
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
		lock, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected lock type after LOCK=: %w", err)
		}
		op.LockValue = lock
	}

	return op, nil
}

// parseAlterView parses ALTER VIEW statements
// Reference: src/parser/mod.rs:10688
func parseAlterView(p *Parser) (ast.Statement, error) {
	// Parse view name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected view name after ALTER VIEW: %w", err)
	}

	// Parse optional column list
	var columns []*ast.Ident
	tok := p.PeekToken()
	if _, ok := tok.Token.(token.TokenLParen); ok {
		p.AdvanceToken() // consume (
		for {
			nextTok := p.PeekToken()
			if _, isRParen := nextTok.Token.(token.TokenRParen); isRParen {
				break
			}
			col, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected column name: %w", err)
			}
			columns = append(columns, col)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse optional WITH options
	var withOptions []*expr.SqlOption
	if p.ParseKeyword("WITH") {
		// Parse WITH options if present
		// For now, skip complex WITH option parsing
	}

	// Expect AS
	if !p.ParseKeyword("AS") {
		return nil, fmt.Errorf("expected AS after ALTER VIEW")
	}

	// Parse query
	queryStmt, err := p.ParseQuery()
	if err != nil {
		return nil, fmt.Errorf("expected query after AS: %w", err)
	}

	// Extract query from statement
	var q *query.Query
	switch stmt := queryStmt.(type) {
	case *QueryStatement:
		q = stmt.Query
	case *SelectStatement:
		q = &query.Query{}
		// Build a simple Query Body from the Select
		setExpr := &query.SelectSetExpr{
			Select: &stmt.Select,
		}
		q.Body = setExpr
	default:
		return nil, fmt.Errorf("expected SELECT query in ALTER VIEW")
	}

	return &statement.AlterView{
		Name:        name,
		Columns:     columns,
		Query:       q,
		WithOptions: withOptions,
	}, nil
}

// parseAlterIndex parses ALTER INDEX statements
// Reference: src/parser/mod.rs:10606
func parseAlterIndex(p *Parser) (ast.Statement, error) {
	// Parse index name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected index name after ALTER INDEX: %w", err)
	}

	// Currently only supports RENAME TO
	if !p.ParseKeyword("RENAME") {
		return nil, fmt.Errorf("expected RENAME after ALTER INDEX")
	}

	if !p.ParseKeyword("TO") {
		return nil, fmt.Errorf("expected TO after RENAME")
	}

	_, err = p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected new index name: %w", err)
	}

	// For now, use empty operation - the rename info would be in the operation
	return &statement.AlterIndex{
		Name:      name,
		Operation: &expr.AlterIndexOperation{},
	}, nil
}

// parseAlterRole parses ALTER ROLE statements
// Reference: src/parser/alter.rs:34
func parseAlterRole(p *Parser) (ast.Statement, error) {
	// Parse role name
	roleName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected role name after ALTER ROLE: %w", err)
	}

	// Check for various operations
	var operation *expr.AlterRoleOperation

	// RENAME TO
	if p.ParseKeyword("RENAME") {
		if !p.ParseKeyword("TO") {
			return nil, fmt.Errorf("expected TO after RENAME")
		}
		newName, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected new role name: %w", err)
		}
		_ = newName // Use in operation
	}

	// WITH options (PostgreSQL style)
	if p.ParseKeyword("WITH") {
		// Parse role options
		for {
			if !p.ParseKeyword("LOGIN") && !p.ParseKeyword("NOLOGIN") &&
				!p.ParseKeyword("SUPERUSER") && !p.ParseKeyword("NOSUPERUSER") &&
				!p.ParseKeyword("CREATEDB") && !p.ParseKeyword("NOCREATEDB") &&
				!p.ParseKeyword("CREATEROLE") && !p.ParseKeyword("NOCREATEROLE") &&
				!p.ParseKeyword("INHERIT") && !p.ParseKeyword("NOINHERIT") &&
				!p.ParseKeyword("REPLICATION") && !p.ParseKeyword("NOREPLICATION") &&
				!p.ParseKeyword("BYPASSRLS") && !p.ParseKeyword("NOBYPASSRLS") {
				break
			}
		}
	}

	return &statement.AlterRole{
		Name:      roleName,
		Operation: operation,
	}, nil
}

// parseAlterUser parses ALTER USER statements
// Reference: src/parser/alter.rs:151
func parseAlterUser(p *Parser) (ast.Statement, error) {
	// Check for IF EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse user name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected user name after ALTER USER: %w", err)
	}

	// Skip WITH if present
	p.ParseKeyword("WITH")

	var renameTo *ast.Ident
	var resetPassword bool

	// Check for RENAME TO
	if p.ParseKeywords([]string{"RENAME", "TO"}) {
		newName, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected new user name: %w", err)
		}
		renameTo = newName
	}

	// Check for RESET PASSWORD
	if p.ParseKeywords([]string{"RESET", "PASSWORD"}) {
		resetPassword = true
	}

	// Check for PASSWORD
	if p.ParseKeyword("PASSWORD") {
		// Parse password value or NULL
		if !p.ParseKeyword("NULL") {
			// Parse password string
			tok := p.NextToken()
			if _, ok := tok.Token.(token.TokenSingleQuotedString); !ok {
				p.PrevToken()
			}
		}
	}

	return &statement.AlterUser{
		IfExists:      ifNotExists,
		Name:          name,
		RenameTo:      renameTo,
		ResetPassword: resetPassword,
	}, nil
}

// parseAlterSchema parses ALTER SCHEMA statements
// Reference: src/parser/mod.rs:11064
func parseAlterSchema(p *Parser) (ast.Statement, error) {
	// Parse schema name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected schema name after ALTER SCHEMA: %w", err)
	}

	// Check for RENAME TO
	if p.ParseKeyword("RENAME") {
		if !p.ParseKeyword("TO") {
			return nil, fmt.Errorf("expected TO after RENAME")
		}
		newName, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected new schema name: %w", err)
		}
		_ = newName
	}

	// Check for OWNER TO
	if p.ParseKeywords([]string{"OWNER", "TO"}) {
		owner, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected owner name: %w", err)
		}
		_ = owner
	}

	return &statement.AlterSchema{
		Name:      name,
		Operation: &expr.AlterSchemaOperation{},
	}, nil
}

// parseAlterType parses ALTER TYPE statements
// Reference: src/parser/mod.rs:10706
func parseAlterType(p *Parser) (ast.Statement, error) {
	// Parse type name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected type name after ALTER TYPE: %w", err)
	}

	var operations []*expr.AlterTypeOperation

	// RENAME TO
	if p.ParseKeywords([]string{"RENAME", "TO"}) {
		newName, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected new type name: %w", err)
		}
		_ = newName
	}

	// ADD VALUE (for enum types)
	if p.ParseKeywords([]string{"ADD", "VALUE"}) {
		p.ParseKeywords([]string{"IF", "NOT", "EXISTS"})
		value, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected enum value: %w", err)
		}
		_ = value
		// Optional BEFORE/AFTER position
		if p.ParseKeyword("BEFORE") {
			p.ParseIdentifier()
		} else if p.ParseKeyword("AFTER") {
			p.ParseIdentifier()
		}
	}

	return &statement.AlterType{
		Name:       name,
		Operations: operations,
	}, nil
}

// parseAlterPolicy parses ALTER POLICY statements
// Reference: src/parser/alter.rs:57
func parseAlterPolicy(p *Parser) (ast.Statement, error) {
	// Parse policy name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected policy name after ALTER POLICY: %w", err)
	}

	// Expect ON
	if !p.ParseKeyword("ON") {
		return nil, fmt.Errorf("expected ON after policy name")
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected table name: %w", err)
	}

	// Check for RENAME TO
	if p.ParseKeyword("RENAME") {
		if !p.ParseKeyword("TO") {
			return nil, fmt.Errorf("expected TO after RENAME")
		}
		newName, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected new policy name: %w", err)
		}
		_ = newName
	}

	return &statement.AlterPolicy{
		Name:      name,
		TableName: tableName,
		Operation: &expr.AlterPolicyOperation{},
	}, nil
}

// parseAlterConnector parses ALTER CONNECTOR statements
// Reference: src/parser/alter.rs:114
func parseAlterConnector(p *Parser) (ast.Statement, error) {
	// Parse connector name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected connector name after ALTER CONNECTOR: %w", err)
	}

	// Expect SET
	if !p.ParseKeyword("SET") {
		return nil, fmt.Errorf("expected SET after connector name")
	}

	// Parse DCPROPERTIES or other options
	if p.ParseKeyword("DCPROPERTIES") {
		// Expect parenthesized properties
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		// Skip properties for now
		for {
			tok := p.PeekToken()
			if _, isRParen := tok.Token.(token.TokenRParen); isRParen {
				break
			}
			p.NextToken()
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &statement.AlterConnector{
		Name:  name,
		Owner: nil,
	}, nil
}

// parseAlterOperator parses ALTER OPERATOR statements
// Reference: src/parser/mod.rs:10758
func parseAlterOperator(p *Parser) (ast.Statement, error) {
	// ALTER OPERATOR name(left_type, right_type) OWNER TO new_owner | SET SCHEMA new_schema | SET (...)
	if _, err := p.ExpectKeyword("OPERATOR"); err != nil {
		return nil, err
	}

	// Parse parentheses with types
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Skip type parsing
	for !p.ConsumeToken(token.TokenRParen{}) {
		p.AdvanceToken()
	}

	// Simplified: just return a placeholder statement
	return &statement.AlterOperator{}, nil
}

// parseAlterOperatorClass parses ALTER OPERATOR CLASS statements
func parseAlterOperatorClass(p *Parser) (ast.Statement, error) {
	// ALTER OPERATOR CLASS name USING index_method RENAME TO new_name | OWNER TO new_owner | SET SCHEMA new_schema
	if err := p.ExpectKeywords([]string{"OPERATOR", "CLASS"}); err != nil {
		return nil, err
	}

	// Skip name and USING clause
	p.AdvanceToken() // name
	if p.ParseKeyword("USING") {
		p.AdvanceToken() // method
	}

	// Simplified: just return a placeholder
	return &statement.AlterOperatorClass{}, nil
}

// parseAlterOperatorFamily parses ALTER OPERATOR FAMILY statements
func parseAlterOperatorFamily(p *Parser) (ast.Statement, error) {
	// ALTER OPERATOR FAMILY name USING index_method ...
	if err := p.ExpectKeywords([]string{"OPERATOR", "FAMILY"}); err != nil {
		return nil, err
	}

	// Skip name and USING clause
	p.AdvanceToken() // name
	if p.ParseKeyword("USING") {
		p.AdvanceToken() // method
	}

	// Simplified: just return a placeholder
	return &statement.AlterOperatorFamily{}, nil
}
