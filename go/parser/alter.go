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
	"strings"

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
		return parseAlterTable(p, expr.AlterTableTypeNone)
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
			return parseAlterTable(p, expr.AlterTableTypeIceberg)
		}
		return nil, fmt.Errorf("expected TABLE after ALTER ICEBERG")
	}
	// Check for DYNAMIC TABLE (Snowflake)
	if p.ParseKeywords([]string{"DYNAMIC", "TABLE"}) {
		return parseAlterDynamicTable(p)
	}
	// Check for EXTERNAL TABLE (Snowflake)
	if p.ParseKeywords([]string{"EXTERNAL", "TABLE"}) {
		return parseAlterExternalTable(p)
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
	// Check for SESSION (Snowflake)
	if p.ParseKeyword("SESSION") {
		return parseAlterSession(p)
	}
	return nil, fmt.Errorf("expected TABLE, VIEW, INDEX, ROLE, USER, SCHEMA, TYPE, POLICY, CONNECTOR, or SESSION after ALTER")
}

// parseAlterTable parses ALTER TABLE statements
func parseAlterTable(p *Parser, tableType expr.AlterTableType) (ast.Statement, error) {
	alterTable := &statement.AlterTable{
		TableType: tableType,
	}

	// Parse IF EXISTS (PostgreSQL)
	alterTable.IfExists = p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse ONLY (PostgreSQL) - affects inheritance behavior
	alterTable.Only = p.ParseKeyword("ONLY")

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected table name after ALTER TABLE: %w", err)
	}
	alterTable.Name = tableName

	// Parse optional ON CLUSTER (ClickHouse)
	if p.ParseKeywords([]string{"ON", "CLUSTER"}) {
		tok := p.NextToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			alterTable.OnCluster = &ast.Ident{Value: word.Word.Value}
		} else if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			quote := rune('\'')
			alterTable.OnCluster = &ast.Ident{Value: str.Value, QuoteStyle: &quote}
		} else if str, ok := tok.Token.(token.TokenDoubleQuotedString); ok {
			quote := rune('"')
			alterTable.OnCluster = &ast.Ident{Value: str.Value, QuoteStyle: &quote}
		} else {
			return nil, p.Expected("identifier or string after ON CLUSTER", tok)
		}
	}

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

	// PostgreSQL ENABLE/DISABLE operations
	if p.ParseKeyword("DISABLE") {
		return parseAlterTableDisable(p, op)
	}

	if p.ParseKeyword("ENABLE") {
		return parseAlterTableEnable(p, op)
	}

	// PostgreSQL REPLICA IDENTITY
	if p.ParseKeywords([]string{"REPLICA", "IDENTITY"}) {
		return parseAlterTableReplicaIdentity(p, op)
	}

	// PostgreSQL FORCE ROW LEVEL SECURITY
	if p.ParseKeywords([]string{"FORCE", "ROW", "LEVEL", "SECURITY"}) {
		op.Op = expr.AlterTableOpForceRowLevelSecurity
		return op, nil
	}

	// PostgreSQL NO FORCE ROW LEVEL SECURITY
	if p.ParseKeywords([]string{"NO", "FORCE", "ROW", "LEVEL", "SECURITY"}) {
		op.Op = expr.AlterTableOpNoForceRowLevelSecurity
		return op, nil
	}

	// Snowflake SUSPEND RECLUSTER
	if p.ParseKeywords([]string{"SUSPEND", "RECLUSTER"}) {
		op.Op = expr.AlterTableOpSuspendRecluster
		return op, nil
	}

	// Snowflake RESUME RECLUSTER
	if p.ParseKeywords([]string{"RESUME", "RECLUSTER"}) {
		op.Op = expr.AlterTableOpResumeRecluster
		return op, nil
	}

	// Snowflake SWAP WITH
	if p.ParseKeyword("SWAP") {
		return parseAlterTableSwapWith(p, op)
	}

	// Snowflake CLUSTER BY
	if p.ParseKeywords([]string{"CLUSTER", "BY"}) {
		return parseAlterTableClusterBy(p, op)
	}

	// PostgreSQL OWNER TO
	if p.ParseKeywords([]string{"OWNER", "TO"}) {
		return parseAlterTableOwnerTo(p, op)
	}

	// PostgreSQL VALIDATE CONSTRAINT
	if p.ParseKeywords([]string{"VALIDATE", "CONSTRAINT"}) {
		return parseAlterTableValidateConstraint(p, op)
	}

	return nil, fmt.Errorf("unknown ALTER TABLE operation")
}

// parseAlterTableSwapWith parses ALTER TABLE SWAP WITH operation
// Reference: src/parser/mod.rs:10410-10413
func parseAlterTableSwapWith(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	if !p.ParseKeyword("WITH") {
		return nil, fmt.Errorf("expected WITH after SWAP")
	}
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected table name after SWAP WITH: %w", err)
	}
	op.Op = expr.AlterTableOpSwapWith
	op.SwapWithTableName = tableName
	return op, nil
}

// parseAlterTableClusterBy parses ALTER TABLE CLUSTER BY operation
// Reference: src/parser/mod.rs:10459-10463
func parseAlterTableClusterBy(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse comma-separated expressions
	ep := NewExpressionParser(p)
	for {
		exprVal, err := ep.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("expected expression in CLUSTER BY: %w", err)
		}
		op.ClusterBy = append(op.ClusterBy, exprVal)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	op.Op = expr.AlterTableOpClusterBy
	return op, nil
}

// parseAlterTableOwnerTo parses ALTER TABLE OWNER TO operation
// Reference: src/parser/mod.rs:10414-10418
func parseAlterTableOwnerTo(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpOwnerTo

	// Parse owner - can be identifier or special values like CURRENT_USER, CURRENT_ROLE, SESSION_USER
	tok := p.PeekToken()
	if word, ok := tok.Token.(token.TokenWord); ok {
		switch strings.ToUpper(word.Word.Value) {
		case "CURRENT_USER", "CURRENT_ROLE", "SESSION_USER":
			p.AdvanceToken()
			op.NewOwner = &expr.OwnerToTarget{
				Name:      word.Word.Value,
				IsSpecial: true,
			}
			return op, nil
		}
	}

	// Regular identifier (preserves quote style)
	ownerIdent, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected owner name after OWNER TO: %w", err)
	}
	op.NewOwner = &expr.OwnerToTarget{
		Ident:     ownerIdent,
		IsSpecial: false,
	}
	return op, nil
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

	// Check for CLUSTERING KEY (Snowflake)
	if p.ParseKeyword("CLUSTERING") {
		if p.ParseKeyword("KEY") {
			op.Op = expr.AlterTableOpDropClusteringKey
			return op, nil
		}
		return nil, fmt.Errorf("expected KEY after CLUSTERING")
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
		op.RenameTableAsKind = expr.RenameTableTo
		return parseAlterTableRenameTable(p, op)
	}

	// RENAME AS <new_table_name> (alternative syntax)
	if p.ParseKeyword("AS") {
		op.RenameTableAsKind = expr.RenameTableAs
		return parseAlterTableRenameTable(p, op)
	}

	// RENAME CONSTRAINT (PostgreSQL-specific)
	if p.GetDialect().SupportsRenameConstraint() && p.ParseKeyword("CONSTRAINT") {
		return parseAlterTableRenameConstraint(p, op)
	}

	// RENAME [COLUMN] <old_name> TO <new_name>
	return parseAlterTableRenameColumn(p, op)
}

// parseAlterTableRenameConstraint parses RENAME CONSTRAINT <old_name> TO <new_name> (PostgreSQL)
func parseAlterTableRenameConstraint(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpRenameConstraint

	// Parse old constraint name
	oldName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected constraint name: %w", err)
	}
	op.RenameConstraintOldName = oldName

	// Expect TO
	if !p.ParseKeyword("TO") {
		return nil, fmt.Errorf("expected TO after constraint name")
	}

	// Parse new constraint name
	newName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected new constraint name: %w", err)
	}
	op.RenameConstraintNewName = newName

	return op, nil
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
		op.AlterDataTypeHadSet = true // Track that SET was used
		// Parse data type
		dataType, err := p.ParseDataType()
		if err == nil {
			op.AlterDataType = dataType
		}
		// Parse optional USING clause (PostgreSQL)
		if p.GetDialect().SupportsAlterColumnTypeUsing() && p.ParseKeyword("USING") {
			exprParser := NewExpressionParser(p)
			usingExpr, err := exprParser.ParseExpr()
			if err == nil {
				op.AlterUsing = usingExpr
			}
		}
	} else if p.ParseKeyword("TYPE") {
		op.AlterColumnOp = expr.AlterColumnOpSetDataType
		op.AlterDataTypeHadSet = false // Track that SET was NOT used (just TYPE)
		// Parse data type
		dataType, err := p.ParseDataType()
		if err == nil {
			op.AlterDataType = dataType
		}
		// Parse optional USING clause (PostgreSQL)
		if p.GetDialect().SupportsAlterColumnTypeUsing() && p.ParseKeyword("USING") {
			exprParser := NewExpressionParser(p)
			usingExpr, err := exprParser.ParseExpr()
			if err == nil {
				op.AlterUsing = usingExpr
			}
		}
	} else if p.ParseKeyword("ADD") {
		// PostgreSQL: ADD GENERATED { ALWAYS | BY DEFAULT } AS IDENTITY [ ( sequence_options ) ]
		if p.ParseKeyword("GENERATED") {
			op.AlterColumnOp = expr.AlterColumnOpAddGenerated
			// Parse optional ALWAYS or BY DEFAULT
			if p.ParseKeyword("ALWAYS") {
				op.AlterGeneratedAs = expr.GeneratedAsAlways
			} else if p.ParseKeywords([]string{"BY", "DEFAULT"}) {
				op.AlterGeneratedAs = expr.GeneratedAsByDefault
			}
			// Expect AS IDENTITY
			if !p.ParseKeywords([]string{"AS", "IDENTITY"}) {
				return nil, fmt.Errorf("expected AS IDENTITY after ADD GENERATED")
			}
			// Parse optional sequence options in parentheses
			if p.ConsumeToken(token.TokenLParen{}) {
				op.AlterGeneratedHasParens = true
				if !p.ConsumeToken(token.TokenRParen{}) {
					// Parse sequence options
					opts, err := parseCreateSequenceOptions(p)
					if err != nil {
						return nil, err
					}
					op.AlterGeneratedSequenceOpts = opts
					if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
						return nil, err
					}
				}
			}
		} else {
			return nil, fmt.Errorf("expected GENERATED after ADD")
		}
	} else {
		return nil, fmt.Errorf("expected SET NOT NULL, DROP NOT NULL, SET DEFAULT, DROP DEFAULT, SET DATA TYPE, or ADD GENERATED after ALTER COLUMN")
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

// parseOptions parses a parenthesized list of SQL options: (key=value, ...)
// Reference: src/parser/mod.rs:9829-9838
func parseOptions(p *Parser) ([]*expr.SqlOption, error) {
	// Expect opening parenthesis
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse the options using the existing helper
	options, err := parseSqlOptions(p)
	if err != nil {
		return nil, err
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return options, nil
}

// parseAlterTableSet parses SET TBLPROPERTIES or SET options
// Reference: src/parser/mod.rs:10528-10545
func parseAlterTableSet(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	// SET TBLPROPERTIES(...)
	if p.ParseKeyword("TBLPROPERTIES") {
		return parseAlterTableSetTblProperties(p, op)
	}

	// SET (...) - parenthesized options like SET (key=value, ...)
	// Reference: src/parser/mod.rs:10536-10544
	if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
		op.Op = expr.AlterTableOpSetOptionsParens
		// Parse the parenthesized options
		options, err := parseOptions(p)
		if err != nil {
			return nil, err
		}
		op.SetOptions = options
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
		// Expect opening paren
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, fmt.Errorf("expected '(' after WITH: %w", err)
		}
		// Parse options
		withOptions, err = parseSqlOptions(p)
		if err != nil {
			return nil, err
		}
		// Expect closing paren
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
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

	newName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected new index name: %w", err)
	}

	// Convert ast.ObjectName to expr.ObjectName
	parts := make([]*expr.ObjectNamePart, len(newName.Parts))
	for i, part := range newName.Parts {
		if idPart, ok := part.(*ast.ObjectNamePartIdentifier); ok {
			parts[i] = &expr.ObjectNamePart{Ident: &expr.Ident{Value: idPart.Ident.Value, QuoteStyle: idPart.Ident.QuoteStyle}}
		}
	}

	// Create operation with rename info
	return &statement.AlterIndex{
		Name: name,
		Operation: &expr.AlterIndexOperation{
			RenameTo: &expr.ObjectName{Parts: parts},
		},
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

	// Check for IN DATABASE clause (PostgreSQL)
	var inDatabase *ast.ObjectName
	if p.ParseKeywords([]string{"IN", "DATABASE"}) {
		dbName, err := p.ParseObjectName()
		if err != nil {
			return nil, fmt.Errorf("expected database name after IN DATABASE: %w", err)
		}
		inDatabase = dbName
	}

	// Check for various operations
	var operation expr.AlterRoleOperation

	// RENAME TO
	if p.ParseKeyword("RENAME") {
		if !p.ParseKeyword("TO") {
			return nil, fmt.Errorf("expected TO after RENAME")
		}
		newName, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected new role name: %w", err)
		}
		operation = &expr.AlterRoleOperationRenameRole{RoleName: newName}
	} else if p.ParseKeyword("SET") {
		// SET config_name { TO | = } { value | DEFAULT }
		// SET config_name FROM CURRENT
		configName, err := p.ParseObjectName()
		if err != nil {
			return nil, fmt.Errorf("expected config name after SET: %w", err)
		}

		configValue := &expr.SetConfigValue{}
		useEqual := false

		if p.ParseKeywords([]string{"FROM", "CURRENT"}) {
			configValue.Type = expr.SetConfigValueFromCurrent
		} else if p.ConsumeToken(token.TokenEq{}) {
			useEqual = true
			if p.ParseKeyword("DEFAULT") {
				configValue.Type = expr.SetConfigValueDefault
			} else {
				ep := NewExpressionParser(p)
				val, err := ep.ParseExpr()
				if err != nil {
					return nil, fmt.Errorf("expected value after SET config =: %w", err)
				}
				configValue.Type = expr.SetConfigValueExpr
				configValue.Value = val
			}
		} else if p.ParseKeyword("TO") {
			if p.ParseKeyword("DEFAULT") {
				configValue.Type = expr.SetConfigValueDefault
			} else {
				ep := NewExpressionParser(p)
				val, err := ep.ParseExpr()
				if err != nil {
					return nil, fmt.Errorf("expected value after SET config TO: %w", err)
				}
				configValue.Type = expr.SetConfigValueExpr
				configValue.Value = val
			}
		} else {
			return nil, fmt.Errorf("expected =, TO, or FROM CURRENT after SET config_name")
		}

		operation = &expr.AlterRoleOperationSet{
			InDatabase:  inDatabase,
			ConfigName:  configName,
			ConfigValue: configValue,
			UseEqual:    useEqual,
		}
	} else if p.ParseKeyword("RESET") {
		// RESET config_name | RESET ALL
		resetConfig := &expr.ResetConfig{}

		if p.ParseKeyword("ALL") {
			resetConfig.IsAll = true
		} else {
			configName, err := p.ParseObjectName()
			if err != nil {
				return nil, fmt.Errorf("expected config name or ALL after RESET: %w", err)
			}
			resetConfig.ConfigName = configName
		}

		operation = &expr.AlterRoleOperationReset{
			InDatabase: inDatabase,
			ConfigName: resetConfig,
		}
	} else {
		// WITH options (PostgreSQL style) - optional WITH keyword
		p.ParseKeyword("WITH")

		var options []*expr.RoleOption

		for {
			if p.ParseKeyword("LOGIN") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionLogin})
			} else if p.ParseKeyword("NOLOGIN") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionNoLogin})
			} else if p.ParseKeyword("SUPERUSER") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionSuperUser})
			} else if p.ParseKeyword("NOSUPERUSER") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionNoSuperUser})
			} else if p.ParseKeyword("CREATEDB") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionCreateDB})
			} else if p.ParseKeyword("NOCREATEDB") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionNoCreateDB})
			} else if p.ParseKeyword("CREATEROLE") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionCreateRole})
			} else if p.ParseKeyword("NOCREATEROLE") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionNoCreateRole})
			} else if p.ParseKeyword("INHERIT") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionInherit})
			} else if p.ParseKeyword("NOINHERIT") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionNoInherit})
			} else if p.ParseKeyword("REPLICATION") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionReplication})
			} else if p.ParseKeyword("NOREPLICATION") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionNoReplication})
			} else if p.ParseKeyword("BYPASSRLS") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionBypassRLS})
			} else if p.ParseKeyword("NOBYPASSRLS") {
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionNoBypassRLS})
			} else if p.ParseKeyword("CONNECTION") {
				if !p.ParseKeyword("LIMIT") {
					return nil, fmt.Errorf("expected LIMIT after CONNECTION")
				}
				ep := NewExpressionParser(p)
				val, err := ep.ParseExpr()
				if err != nil {
					return nil, fmt.Errorf("expected expression after CONNECTION LIMIT: %w", err)
				}
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionConnectionLimit, Value: val})
			} else if p.ParseKeyword("PASSWORD") {
				if p.ParseKeyword("NULL") {
					options = append(options, &expr.RoleOption{Type: expr.RoleOptionPassword, Password: &expr.Password{IsNull: true}})
				} else {
					ep := NewExpressionParser(p)
					val, err := ep.ParseExpr()
					if err != nil {
						return nil, fmt.Errorf("expected value after PASSWORD: %w", err)
					}
					options = append(options, &expr.RoleOption{Type: expr.RoleOptionPassword, Password: &expr.Password{Value: val}})
				}
			} else if p.ParseKeyword("VALID") {
				if !p.ParseKeyword("UNTIL") {
					return nil, fmt.Errorf("expected UNTIL after VALID")
				}
				ep := NewExpressionParser(p)
				val, err := ep.ParseExpr()
				if err != nil {
					return nil, fmt.Errorf("expected expression after VALID UNTIL: %w", err)
				}
				options = append(options, &expr.RoleOption{Type: expr.RoleOptionValidUntil, Value: val})
			} else {
				break
			}
		}

		if len(options) > 0 {
			operation = &expr.AlterRoleOperationWithOptions{Options: options}
		}
	}

	return &statement.AlterRole{
		Name:      roleName,
		Operation: operation,
	}, nil
}

// parseAlterUser parses ALTER USER statements
// Reference: src/parser/alter.rs:151-333
func parseAlterUser(p *Parser) (ast.Statement, error) {
	// Check for IF EXISTS
	ifNotExists := p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse user name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected user name after ALTER USER: %w", err)
	}

	// Skip optional WITH keyword
	p.ParseKeyword("WITH")

	alterUser := &statement.AlterUser{
		IfExists: ifNotExists,
		Name:     name,
	}

	// Parse various ALTER USER options in a loop
	for {
		// Check for RENAME TO
		if p.ParseKeywords([]string{"RENAME", "TO"}) {
			newName, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected new user name: %w", err)
			}
			alterUser.RenameTo = newName
			continue
		}

		// Check for RESET PASSWORD
		if p.ParseKeywords([]string{"RESET", "PASSWORD"}) {
			alterUser.ResetPassword = true
			continue
		}

		// Check for ABORT ALL QUERIES
		if p.ParseKeywords([]string{"ABORT", "ALL", "QUERIES"}) {
			alterUser.AbortAllQueries = true
			continue
		}

		// Check for ADD DELEGATED AUTHORIZATION OF ROLE ... TO SECURITY INTEGRATION
		if p.ParseKeywords([]string{"ADD", "DELEGATED", "AUTHORIZATION", "OF", "ROLE"}) {
			role, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected role name: %w", err)
			}
			if err := p.ExpectKeywords([]string{"TO", "SECURITY", "INTEGRATION"}); err != nil {
				return nil, err
			}
			integration, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected integration name: %w", err)
			}
			alterUser.AddRoleDelegation = &statement.AlterUserAddRoleDelegation{
				Role:        role,
				Integration: integration,
			}
			continue
		}

		// Check for REMOVE DELEGATED AUTHORIZATION...
		if p.ParseKeywords([]string{"REMOVE", "DELEGATED"}) {
			var role *ast.Ident
			if p.ParseKeywords([]string{"AUTHORIZATION", "OF", "ROLE"}) {
				r, err := p.ParseIdentifier()
				if err != nil {
					return nil, fmt.Errorf("expected role name: %w", err)
				}
				role = r
			} else if !p.ParseKeyword("AUTHORIZATIONS") {
				return nil, fmt.Errorf("expected AUTHORIZATION OF ROLE or AUTHORIZATIONS after REMOVE DELEGATED")
			}
			if err := p.ExpectKeywords([]string{"FROM", "SECURITY", "INTEGRATION"}); err != nil {
				return nil, err
			}
			integration, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected integration name: %w", err)
			}
			alterUser.RemoveRoleDelegation = &statement.AlterUserRemoveRoleDelegation{
				Role:        role,
				Integration: integration,
			}
			continue
		}

		// Check for ENROLL MFA
		if p.ParseKeywords([]string{"ENROLL", "MFA"}) {
			alterUser.EnrollMfa = true
			continue
		}

		// Check for SET DEFAULT_MFA_METHOD - this handles SET DEFAULT_MFA_METHOD=value
		if p.PeekKeyword("SET") {
			// Look ahead to see if next is DEFAULT_MFA_METHOD
			restore := p.SavePosition()
			p.ParseKeyword("SET") // consume SET
			if p.ParseKeyword("DEFAULT_MFA_METHOD") {
				// MFA method can be a keyword (PASSKEY, TOTP, DUO) or a string literal ('PASSKEY')
				if p.ParseKeyword("PASSKEY") {
					alterUser.SetDefaultMfaMethod = statement.MfaMethodKindPassKey
				} else if p.ParseKeyword("TOTP") {
					alterUser.SetDefaultMfaMethod = statement.MfaMethodKindTotp
				} else if p.ParseKeyword("DUO") {
					alterUser.SetDefaultMfaMethod = statement.MfaMethodKindDuo
				} else {
					// Try parsing as string literal like 'PASSKEY'
					methodStr, err := p.ParseStringLiteral()
					if err == nil {
						switch methodStr {
						case "PASSKEY":
							alterUser.SetDefaultMfaMethod = statement.MfaMethodKindPassKey
						case "TOTP":
							alterUser.SetDefaultMfaMethod = statement.MfaMethodKindTotp
						case "DUO":
							alterUser.SetDefaultMfaMethod = statement.MfaMethodKindDuo
						default:
							return nil, fmt.Errorf("unknown MFA method: %s", methodStr)
						}
					} else {
						return nil, fmt.Errorf("expected PASSKEY, TOTP, DUO, or string literal")
					}
				}
				alterUser.HasSetDefaultMfaMethod = true
				continue
			}
			// Not DEFAULT_MFA_METHOD, restore position and try other SET variants
			restore()
		}

		// Check for REMOVE MFA METHOD
		if p.ParseKeywords([]string{"REMOVE", "MFA", "METHOD"}) {
			method, err := parseMfaMethod(p)
			if err != nil {
				return nil, err
			}
			alterUser.RemoveMfaMethod = method
			alterUser.HasRemoveMfaMethod = true
			continue
		}

		// Check for MODIFY MFA METHOD
		if p.ParseKeywords([]string{"MODIFY", "MFA", "METHOD"}) {
			method, err := parseMfaMethod(p)
			if err != nil {
				return nil, err
			}
			if err := p.ExpectKeywords([]string{"SET", "COMMENT"}); err != nil {
				return nil, err
			}
			comment, err := p.ParseStringLiteral()
			if err != nil {
				return nil, fmt.Errorf("expected comment string: %w", err)
			}
			alterUser.ModifyMfaMethod = &statement.AlterUserModifyMfaMethod{
				Method:  method,
				Comment: comment,
			}
			continue
		}

		// Check for ADD MFA METHOD OTP
		if p.ParseKeywords([]string{"ADD", "MFA", "METHOD", "OTP"}) {
			addOtp := &statement.AlterUserAddMfaMethodOtp{}
			if p.ParseKeyword("COUNT") {
				if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
					return nil, err
				}
				tok := p.NextToken()
				if num, ok := tok.Token.(token.TokenNumber); ok {
					addOtp.Count = &expr.ValueExpr{Value: num.Value}
				} else {
					return nil, fmt.Errorf("expected number after COUNT=")
				}
			}
			alterUser.AddMfaMethodOtp = addOtp
			continue
		}

		// Check for SET {AUTHENTICATION|PASSWORD|SESSION} POLICY
		if p.ParseKeyword("SET") {
			// First check for specific SET operations that don't use = syntax
			// Check for policy types: SET AUTHENTICATION POLICY name
			if p.ParseKeywords([]string{"AUTHENTICATION", "POLICY"}) {
				policy, err := p.ParseIdentifier()
				if err != nil {
					return nil, fmt.Errorf("expected policy name: %w", err)
				}
				alterUser.SetPolicy = &statement.AlterUserSetPolicy{
					PolicyKind: statement.UserPolicyKindAuthentication,
					Policy:     policy,
				}
				continue
			}
			if p.ParseKeywords([]string{"PASSWORD", "POLICY"}) {
				policy, err := p.ParseIdentifier()
				if err != nil {
					return nil, fmt.Errorf("expected policy name: %w", err)
				}
				alterUser.SetPolicy = &statement.AlterUserSetPolicy{
					PolicyKind: statement.UserPolicyKindPassword,
					Policy:     policy,
				}
				continue
			}
			if p.ParseKeywords([]string{"SESSION", "POLICY"}) {
				policy, err := p.ParseIdentifier()
				if err != nil {
					return nil, fmt.Errorf("expected policy name: %w", err)
				}
				alterUser.SetPolicy = &statement.AlterUserSetPolicy{
					PolicyKind: statement.UserPolicyKindSession,
					Policy:     policy,
				}
				continue
			}

			// Check for SET TAG key=value...
			if p.ParseKeyword("TAG") {
				tagOpts, err := parseAlterUserKeyValueOptions(p, false)
				if err != nil {
					return nil, fmt.Errorf("expected tag options: %w", err)
				}
				alterUser.SetTag = tagOpts
				continue
			}

			// Generic SET property=value parsing (handles PASSWORD='secret', DEFAULT_MFA_METHOD='PASSKEY', etc.)
			opts, err := parseAlterUserKeyValueOptions(p, false)
			if err != nil {
				return nil, fmt.Errorf("expected property=value after SET: %w", err)
			}
			alterUser.SetProperties = opts
			continue
		}

		// Check for UNSET {AUTHENTICATION|PASSWORD|SESSION} POLICY
		if p.ParseKeyword("UNSET") {
			if p.ParseKeywords([]string{"AUTHENTICATION", "POLICY"}) {
				kind := statement.UserPolicyKindAuthentication
				alterUser.UnsetPolicy = &kind
				continue
			}
			if p.ParseKeywords([]string{"PASSWORD", "POLICY"}) {
				kind := statement.UserPolicyKindPassword
				alterUser.UnsetPolicy = &kind
				continue
			}
			if p.ParseKeywords([]string{"SESSION", "POLICY"}) {
				kind := statement.UserPolicyKindSession
				alterUser.UnsetPolicy = &kind
				continue
			}

			// Check for UNSET TAG
			if p.ParseKeyword("TAG") {
				tags, err := parseCommaSeparatedIdentNames(p)
				if err != nil {
					return nil, fmt.Errorf("expected tag names: %w", err)
				}
				alterUser.UnsetTag = tags
				continue
			}

			// Generic UNSET property1, property2, ...
			props, err := parseCommaSeparatedIdentNames(p)
			if err != nil {
				return nil, fmt.Errorf("expected property names: %w", err)
			}
			alterUser.UnsetProperties = props
			continue
		}

		// Check for ENCRYPTED PASSWORD or PASSWORD
		if p.ParseKeyword("ENCRYPTED") {
			if !p.ParseKeyword("PASSWORD") {
				return nil, fmt.Errorf("expected PASSWORD after ENCRYPTED")
			}
			alterUser.Password = &statement.AlterUserPassword{
				Encrypted: true,
			}
			if p.ParseKeyword("NULL") {
				alterUser.Password.IsNull = true
			} else {
				password, err := p.ParseStringLiteral()
				if err != nil {
					return nil, fmt.Errorf("expected password string or NULL: %w", err)
				}
				alterUser.Password.Password = password
			}
			continue
		}

		if p.ParseKeyword("PASSWORD") {
			alterUser.Password = &statement.AlterUserPassword{}
			if p.ParseKeyword("NULL") {
				alterUser.Password.IsNull = true
			} else {
				password, err := p.ParseStringLiteral()
				if err != nil {
					return nil, fmt.Errorf("expected password string or NULL: %w", err)
				}
				alterUser.Password.Password = password
			}
			continue
		}

		// If none matched, we're done parsing options
		break
	}

	return alterUser, nil
}

// parseAlterUserKeyValueOptions parses key=value options for ALTER USER (space or comma separated)
func parseAlterUserKeyValueOptions(p *Parser, commaSeparated bool) ([]*expr.SqlOption, error) {
	var options []*expr.SqlOption

	for {
		// Parse key (identifier)
		keyTok := p.PeekToken()
		if _, isWord := keyTok.Token.(token.TokenWord); !isWord {
			break
		}
		key, err := p.ParseIdentifier()
		if err != nil {
			break
		}

		// Expect =
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}

		// Parse value
		valTok := p.NextToken()
		var val expr.Expr
		switch v := valTok.Token.(type) {
		case token.TokenSingleQuotedString:
			val = &expr.ValueExpr{Value: v.Value}
		case token.TokenDoubleQuotedString:
			val = &expr.ValueExpr{Value: v.Value}
		case token.TokenNumber:
			val = &expr.ValueExpr{Value: v.Value}
		case token.TokenWord:
			if v.Value == "TRUE" || v.Value == "true" {
				val = &expr.ValueExpr{Value: true}
			} else if v.Value == "FALSE" || v.Value == "false" {
				val = &expr.ValueExpr{Value: false}
			} else {
				val = &expr.ValueExpr{Value: v.Value}
			}
		case token.TokenLParen:
			// Parenthesized value like ('ALL') - parse as expression
			p.PrevToken() // put back (
			exprParser := NewExpressionParser(p)
			parsedExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, err
			}
			val = parsedExpr
		default:
			return nil, fmt.Errorf("expected value after =")
		}

		options = append(options, &expr.SqlOption{
			Name:  &expr.Ident{Value: key.Value},
			Value: val,
		})

		// Check for comma or space separator
		if commaSeparated {
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		} else {
			// For space-separated, peek if next token looks like a key
			nextTok := p.PeekToken()
			if _, isWord := nextTok.Token.(token.TokenWord); !isWord {
				break
			}
		}
	}

	return options, nil
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
// Reference: src/parser/alter.rs:57-104
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
		return &statement.AlterPolicy{
			Name:      name,
			TableName: tableName,
			Operation: &expr.AlterPolicyOperation{RenameTo: newName},
		}, nil
	}

	// Parse optional TO clause (role names)
	var to []*expr.Owner
	if p.ParseKeyword("TO") {
		for {
			owner, err := parseOwner(p)
			if err != nil {
				return nil, err
			}
			to = append(to, owner)
			if !p.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
	}

	// Parse optional USING (expression)
	var usingExpr expr.Expr
	if p.ParseKeyword("USING") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		ep := NewExpressionParser(p)
		usingExpr, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse optional WITH CHECK (expression)
	var withCheckExpr expr.Expr
	if p.ParseKeywords([]string{"WITH", "CHECK"}) {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		ep := NewExpressionParser(p)
		withCheckExpr, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return &statement.AlterPolicy{
		Name:      name,
		TableName: tableName,
		Operation: &expr.AlterPolicyOperation{
			To:        to,
			Using:     usingExpr,
			WithCheck: withCheckExpr,
		},
	}, nil
}

// parseAlterConnector parses ALTER CONNECTOR statements
// Reference: src/parser/alter.rs:114-145
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
	var properties []*expr.SqlOption
	var url *string
	var owner *expr.AlterConnectorOwner

	if p.ParseKeyword("DCPROPERTIES") {
		// Expect parenthesized properties
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		// Parse properties
		for {
			tok := p.PeekToken()
			if _, isRParen := tok.Token.(token.TokenRParen); isRParen {
				break
			}
			// Parse property name
			propName, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected property name: %w", err)
			}
			// Expect =
			if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
				return nil, err
			}
			// Parse property value
			propValue, err := p.ParseStringLiteral()
			if err != nil {
				return nil, fmt.Errorf("expected property value: %w", err)
			}
			properties = append(properties, &expr.SqlOption{
				Name:  exprFromAstIdent(propName),
				Value: &expr.ValueExpr{Value: propValue},
			})
			// Check for comma
			tok = p.PeekToken()
			if _, isComma := tok.Token.(token.TokenComma); isComma {
				p.NextToken() // consume comma
			} else {
				break
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	} else if p.ParseKeyword("URL") {
		// SET URL 'value'
		urlValue, err := p.ParseStringLiteral()
		if err != nil {
			return nil, fmt.Errorf("expected URL value: %w", err)
		}
		url = &urlValue
	} else if p.ParseKeyword("OWNER") {
		// SET OWNER [USER|ROLE] name
		if p.ParseKeyword("USER") {
			ownerName, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected owner name: %w", err)
			}
			owner = &expr.AlterConnectorOwner{
				Kind: expr.AlterConnectorOwnerKindUser,
				Name: ownerName,
			}
		} else if p.ParseKeyword("ROLE") {
			ownerName, err := p.ParseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected owner name: %w", err)
			}
			owner = &expr.AlterConnectorOwner{
				Kind: expr.AlterConnectorOwnerKindRole,
				Name: ownerName,
			}
		} else {
			return nil, fmt.Errorf("expected USER or ROLE after OWNER")
		}
	} else {
		return nil, fmt.Errorf("expected DCPROPERTIES, URL, or OWNER after SET")
	}

	return &statement.AlterConnector{
		Name:       name,
		Properties: properties,
		URL:        url,
		Owner:      owner,
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

// parseAlterTableDisable parses DISABLE { ROW LEVEL SECURITY | RULE | TRIGGER }
// Reference: src/parser/mod.rs:10150
func parseAlterTableDisable(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	if p.ParseKeywords([]string{"ROW", "LEVEL", "SECURITY"}) {
		op.Op = expr.AlterTableOpDisableRowLevelSecurity
		return op, nil
	}

	if p.ParseKeyword("RULE") {
		op.Op = expr.AlterTableOpDisableRule
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		op.DisableEnableName = name
		return op, nil
	}

	if p.ParseKeyword("TRIGGER") {
		op.Op = expr.AlterTableOpDisableTrigger
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		op.DisableEnableName = name
		return op, nil
	}

	return nil, fmt.Errorf("expected ROW LEVEL SECURITY, RULE, or TRIGGER after DISABLE")
}

// parseAlterTableEnable parses ENABLE { ALWAYS | REPLICA }? { ROW LEVEL SECURITY | RULE | TRIGGER }
// Reference: src/parser/mod.rs:10165
func parseAlterTableEnable(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	// Check for ALWAYS or REPLICA prefix
	isAlways := false
	isReplica := false

	if p.ParseKeyword("ALWAYS") {
		isAlways = true
	} else if p.ParseKeyword("REPLICA") {
		isReplica = true
	}

	if p.ParseKeywords([]string{"ROW", "LEVEL", "SECURITY"}) {
		op.Op = expr.AlterTableOpEnableRowLevelSecurity
		return op, nil
	}

	if p.ParseKeyword("RULE") {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		op.DisableEnableName = name
		if isAlways {
			op.Op = expr.AlterTableOpEnableAlwaysRule
		} else if isReplica {
			op.Op = expr.AlterTableOpEnableReplicaRule
		} else {
			op.Op = expr.AlterTableOpEnableRule
		}
		return op, nil
	}

	if p.ParseKeyword("TRIGGER") {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		op.DisableEnableName = name
		if isAlways {
			op.Op = expr.AlterTableOpEnableAlwaysTrigger
		} else if isReplica {
			op.Op = expr.AlterTableOpEnableReplicaTrigger
		} else {
			op.Op = expr.AlterTableOpEnableTrigger
		}
		return op, nil
	}

	return nil, fmt.Errorf("expected ALWAYS, REPLICA, ROW LEVEL SECURITY, RULE, or TRIGGER after ENABLE")
}

// parseAlterTableReplicaIdentity parses REPLICA IDENTITY { NOTHING | FULL | DEFAULT | USING INDEX index_name }
// Reference: src/parser/mod.rs:10508
func parseAlterTableReplicaIdentity(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpReplicaIdentity

	if p.ParseKeyword("NOTHING") {
		op.ReplicaIdentity = expr.ReplicaIdentityNothing
	} else if p.ParseKeyword("FULL") {
		op.ReplicaIdentity = expr.ReplicaIdentityFull
	} else if p.ParseKeyword("DEFAULT") {
		op.ReplicaIdentity = expr.ReplicaIdentityDefault
	} else if p.ParseKeywords([]string{"USING", "INDEX"}) {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		op.ReplicaIdentity = expr.ReplicaIdentityIndex
		op.ReplicaIdentityIndex = name
	} else {
		return nil, fmt.Errorf("expected NOTHING, FULL, DEFAULT, or USING INDEX after REPLICA IDENTITY")
	}

	return op, nil
}

// parseMfaMethod parses an MFA method type (PASSKEY, TOTP, DUO, SMS)
func parseMfaMethod(p *Parser) (statement.MfaMethodKind, error) {
	if p.ParseKeyword("PASSKEY") {
		return statement.MfaMethodKindPassKey, nil
	}
	if p.ParseKeyword("TOTP") {
		return statement.MfaMethodKindTotp, nil
	}
	if p.ParseKeyword("DUO") {
		return statement.MfaMethodKindDuo, nil
	}
	if p.ParseKeyword("SMS") {
		// SMS is a valid MFA method but not in the enum - return a default for now
		return statement.MfaMethodKindPassKey, fmt.Errorf("SMS MFA method not yet fully supported")
	}
	return statement.MfaMethodKindPassKey, fmt.Errorf("expected MFA method (PASSKEY, TOTP, DUO, SMS)")
}

// parseCommaSeparatedIdentNames parses a comma-separated list of identifiers
func parseCommaSeparatedIdentNames(p *Parser) ([]string, error) {
	var names []string
	for {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		names = append(names, ident.Value)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return names, nil
}

// parseAlterDynamicTable parses ALTER DYNAMIC TABLE statements (Snowflake)
// Reference: src/dialect/snowflake.rs:721
func parseAlterDynamicTable(p *Parser) (ast.Statement, error) {
	alterTable := &statement.AlterTable{
		TableType: expr.AlterTableTypeDynamic,
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected table name after ALTER DYNAMIC TABLE: %w", err)
	}
	alterTable.Name = tableName

	// Parse the operation (REFRESH, SUSPEND, or RESUME)
	op := &expr.AlterTableOperation{}
	if p.ParseKeyword("REFRESH") {
		op.Op = expr.AlterTableOpRefresh
	} else if p.ParseKeyword("SUSPEND") {
		op.Op = expr.AlterTableOpSuspend
	} else if p.ParseKeyword("RESUME") {
		op.Op = expr.AlterTableOpResume
	} else {
		return nil, fmt.Errorf("expected REFRESH, SUSPEND, or RESUME after ALTER DYNAMIC TABLE")
	}

	alterTable.Operations = []*expr.AlterTableOperation{op}
	return alterTable, nil
}

// parseAlterExternalTable parses ALTER EXTERNAL TABLE statements (Snowflake)
// Reference: src/dialect/snowflake.rs:759
func parseAlterExternalTable(p *Parser) (ast.Statement, error) {
	alterTable := &statement.AlterTable{
		TableType: expr.AlterTableTypeExternal,
	}

	// Parse optional IF EXISTS
	alterTable.IfExists = p.ParseKeywords([]string{"IF", "EXISTS"})

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected table name after ALTER EXTERNAL TABLE: %w", err)
	}
	alterTable.Name = tableName

	// Parse the operation (REFRESH for now)
	op := &expr.AlterTableOperation{}
	if p.ParseKeyword("REFRESH") {
		// Optional subpath for refreshing specific partitions
		subpath := ""
		nextTok := p.PeekTokenRef()
		if str, ok := nextTok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			subpath = str.Value
		}
		if subpath != "" {
			op.RefreshSubpath = &subpath
		}
		op.Op = expr.AlterTableOpRefresh
	} else {
		return nil, fmt.Errorf("expected REFRESH after ALTER EXTERNAL TABLE")
	}

	alterTable.Operations = []*expr.AlterTableOperation{op}
	return alterTable, nil
}

// parseAlterSession parses ALTER SESSION statements (Snowflake)
// Reference: src/dialect/snowflake.rs:801-810
// Syntax: ALTER SESSION { SET | UNSET } <session_params>
func parseAlterSession(p *Parser) (ast.Statement, error) {
	// Check for SET or UNSET
	set := false
	if p.ParseKeyword("SET") {
		set = true
	} else if p.ParseKeyword("UNSET") {
		set = false
	} else {
		return nil, fmt.Errorf("expected SET or UNSET after ALTER SESSION")
	}

	// Parse session options
	options, err := parseSnowflakeSessionOptions(p, set)
	if err != nil {
		return nil, err
	}

	return &statement.AlterSession{
		Set:           set,
		SessionParams: options,
	}, nil
}

// parseSnowflakeSessionOptions parses space/comma separated key=value options for ALTER SESSION
// Reference: src/dialect/snowflake.rs:1630-1673
func parseSnowflakeSessionOptions(p *Parser, set bool) (*expr.KeyValueOptions, error) {
	options := &expr.KeyValueOptions{
		Options:   []*expr.KeyValueOption{},
		Delimiter: expr.KeyValueOptionsDelimiterSpace,
	}

	for {
		tok := p.PeekToken()
		if token.IsEOF(tok.Token) {
			break
		}
		if _, isSemi := tok.Token.(token.TokenSemiColon); isSemi {
			break
		}

		switch v := tok.Token.(type) {
		case token.TokenComma:
			p.AdvanceToken()
			options.Delimiter = expr.KeyValueOptionsDelimiterComma
			continue
		case token.TokenWord:
			p.AdvanceToken()
			if set {
				// For SET: parse key=value
				// Expect = after key
				if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
					return nil, fmt.Errorf("expected = after session parameter name %s", v.Word.Value)
				}
				// Parse value
				valTok := p.PeekTokenRef()
				var val interface{}
				quoted := false
				if str, ok := valTok.Token.(token.TokenSingleQuotedString); ok {
					val = str.Value
					quoted = true
					p.AdvanceToken()
				} else if word, ok := valTok.Token.(token.TokenWord); ok {
					val = word.Word.Value
					p.AdvanceToken()
				} else if num, ok := valTok.Token.(token.TokenNumber); ok {
					val = num.Value
					p.AdvanceToken()
				} else {
					return nil, fmt.Errorf("expected value for session parameter %s, got %s", v.Word.Value, valTok.Token.String())
				}
				options.Options = append(options.Options, &expr.KeyValueOption{
					OptionName:  v.Word.Value,
					OptionValue: val,
					Kind:        expr.KeyValueOptionKindSingle,
					Quoted:      quoted,
				})
			} else {
				// For UNSET: just the key name (no value)
				// Store nil as value so String() won't output =
				options.Options = append(options.Options, &expr.KeyValueOption{
					OptionName:  v.Word.Value,
					OptionValue: nil,
					Kind:        expr.KeyValueOptionKindSingle,
				})
			}
		default:
			return nil, fmt.Errorf("expected session option name, got %s", tok.Token.String())
		}
	}

	if len(options.Options) == 0 {
		return nil, fmt.Errorf("expected at least one session option")
	}

	return options, nil
}

// parseAlterTableValidateConstraint parses VALIDATE CONSTRAINT <name> (PostgreSQL)
// Reference: src/parser/mod.rs:10525-10527
func parseAlterTableValidateConstraint(p *Parser, op *expr.AlterTableOperation) (*expr.AlterTableOperation, error) {
	op.Op = expr.AlterTableOpValidateConstraint

	// Parse constraint name
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected constraint name: %w", err)
	}
	op.ValidateConstraintName = name

	return op, nil
}
