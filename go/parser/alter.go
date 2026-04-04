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
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/tokenizer"
)

// parseAlter parses ALTER statements
func parseAlter(p *Parser) (ast.Statement, error) {
	if p.ParseKeyword("TABLE") {
		return parseAlterTable(p)
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
		if !p.ConsumeToken(tokenizer.TokenComma{}) {
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
func parseAlterTableAdd(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	// Check for COLUMN keyword
	if p.ParseKeyword("COLUMN") {
		op.AddColumnKeyword = true
	}

	// Check for IF NOT EXISTS
	if p.ParseKeywords([]string{"IF", "NOT", "EXISTS"}) {
		op.AddIfNotExists = true
	}

	// Check for parenthesized column list (MySQL style: ADD COLUMN (c1 INT, c2 INT))
	if _, isLParen := p.PeekToken().Token.(tokenizer.TokenLParen); isLParen {
		p.AdvanceToken() // consume (
		op.Op = expr.AlterTableOpAddColumn

		for {
			colDef, err := parseColumnDef(p)
			if err != nil {
				return nil, fmt.Errorf("expected column definition: %w", err)
			}
			op.AddColumnDefs = append(op.AddColumnDefs, colDef)

			// Check for comma separator
			if !p.ConsumeToken(tokenizer.TokenComma{}) {
				break
			}
		}

		if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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
		if _, isComma := p.PeekToken().Token.(tokenizer.TokenComma); !isComma {
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
		case tokenizer.TokenSingleQuotedString:
			return &expr.ColumnOption{
				Name:  "COMMENT",
				Value: &expr.ValueExpr{Value: t.Value},
			}, nil
		case tokenizer.TokenDoubleQuotedString:
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
	if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
		op.Op = expr.AlterTableOpSetOptionsParens
		// Consume until matching RParen
		depth := 1
		for depth > 0 {
			tok := p.NextToken()
			switch tok.Token.(type) {
			case tokenizer.TokenLParen:
				depth++
			case tokenizer.TokenRParen:
				depth--
			case tokenizer.EOF:
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
	if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse key-value pairs
	for {
		// Parse key (quoted string or identifier)
		tok := p.NextToken()
		var key string
		switch t := tok.Token.(type) {
		case tokenizer.TokenSingleQuotedString:
			key = t.Value
		case tokenizer.TokenDoubleQuotedString:
			key = t.Value
		case tokenizer.TokenWord:
			key = t.Value
		default:
			return nil, fmt.Errorf("expected property key")
		}

		// Expect =
		nextTok := p.NextToken()
		if _, ok := nextTok.Token.(tokenizer.TokenEq); !ok {
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
			case tokenizer.TokenSingleQuotedString:
				value = &expr.ValueExpr{Value: t.Value}
			case tokenizer.TokenDoubleQuotedString:
				value = &expr.ValueExpr{Value: t.Value}
			case tokenizer.TokenNumber:
				value = &expr.ValueExpr{Value: t.Value}
			case tokenizer.TokenWord:
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
		if _, isComma := p.PeekToken().Token.(tokenizer.TokenComma); !isComma {
			break
		}
		p.NextToken()
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return nil, err
	}

	return op, nil
}

// parseAlterTableMySqlOptions parses MySQL-specific ALTER TABLE options (AUTO_INCREMENT, ALGORITHM, LOCK)
func parseAlterTableMySqlOptions(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpSetOptions

	if p.ParseKeyword("AUTO_INCREMENT") {
		if _, err := p.ExpectToken(tokenizer.TokenEq{}); err != nil {
			// Allow without equals sign too
		}
		tok := p.NextToken()
		if num, ok := tok.Token.(tokenizer.TokenNumber); ok {
			op.AutoIncrementValue = num.Value
		} else {
			return nil, fmt.Errorf("expected number after AUTO_INCREMENT")
		}
	} else if p.ParseKeyword("ALGORITHM") {
		if _, err := p.ExpectToken(tokenizer.TokenEq{}); err != nil {
			return nil, err
		}
		algo, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected algorithm after ALGORITHM=: %w", err)
		}
		op.AlgorithmValue = algo
	} else if p.ParseKeyword("LOCK") {
		if _, err := p.ExpectToken(tokenizer.TokenEq{}); err != nil {
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

func parseAlterView(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER VIEW parsing not yet implemented")
}

func parseAlterIndex(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER INDEX parsing not yet implemented")
}

func parseAlterRole(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER ROLE parsing not yet implemented")
}

func parseAlterUser(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER USER parsing not yet implemented")
}

func parseAlterSchema(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER SCHEMA parsing not yet implemented")
}

func parseAlterType(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER TYPE parsing not yet implemented")
}

func parseAlterPolicy(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER POLICY parsing not yet implemented")
}

func parseAlterConnector(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER CONNECTOR parsing not yet implemented")
}

func parseAlterOperator(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER OPERATOR parsing not yet implemented")
}

func parseAlterOperatorClass(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER OPERATOR OPERATOR CLASS parsing not yet implemented")
}

func parseAlterOperatorFamily(p *Parser) (ast.Statement, error) {
	return nil, fmt.Errorf("ALTER OPERATOR FAMILY parsing not yet implemented")
}

// ParseAlter parses ALTER statements (exported version)
func ParseAlter(p *Parser) (ast.Statement, error) {
	return parseAlter(p)
}
