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

// Package parser provides SQL parsing functionality.
// This file contains common DDL utilities shared between CREATE and ALTER operations.

package parser

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/datatype"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/token"
)

// parseColumnDef parses a single column definition
// Format: column_name data_type [constraints]
func parseColumnDef(p *Parser) (*expr.ColumnDef, error) {
	// Parse column name
	colName, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected column name: %w", err)
	}

	// Parse data type
	dataType, err := p.ParseDataType()
	if err != nil {
		return nil, fmt.Errorf("expected data type for column %s: %w", colName.Value, err)
	}

	col := &expr.ColumnDef{
		Name:     colName,
		DataType: dataType,
	}

	// Parse column constraints
	for {
		// Check for constraint keywords
		if p.PeekKeyword("NOT") || p.PeekKeyword("NULL") || p.PeekKeyword("DEFAULT") ||
			p.PeekKeyword("COLLATE") || p.PeekKeyword("COMMENT") || p.PeekKeyword("GENERATED") ||
			p.PeekKeyword("CONSTRAINT") || p.PeekKeyword("PRIMARY") || p.PeekKeyword("UNIQUE") ||
			p.PeekKeyword("CHECK") || p.PeekKeyword("REFERENCES") || p.PeekKeyword("ON") ||
			p.PeekKeyword("AUTO_INCREMENT") || p.PeekKeyword("AUTOINCREMENT") || p.PeekKeyword("IDENTITY") {

			constraint, err := parseColumnConstraint(p)
			if err != nil {
				return nil, err
			}
			col.Options = append(col.Options, constraint)
		} else {
			break
		}
	}

	return col, nil
}

// parseColumnConstraint parses a single column constraint
func parseColumnConstraint(p *Parser) (*expr.ColumnOptionDef, error) {
	// Check for named constraint: CONSTRAINT name
	var constraintName *ast.Ident
	if p.ParseKeyword("CONSTRAINT") {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected constraint name after CONSTRAINT: %w", err)
		}
		constraintName = name
		// Continue to parse the actual constraint
	}

	// NOT NULL
	if p.ParseKeywords([]string{"NOT", "NULL"}) {
		return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "NOT NULL"}, nil
	}

	// NULL
	if p.ParseKeyword("NULL") {
		return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "NULL"}, nil
	}

	// DEFAULT expr
	if p.ParseKeyword("DEFAULT") {
		exprParser := NewExpressionParser(p)
		defaultExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("expected expression after DEFAULT: %w", err)
		}
		return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "DEFAULT", Value: defaultExpr}, nil
	}

	// COLLATE collation
	if p.ParseKeyword("COLLATE") {
		collation, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected collation name after COLLATE: %w", err)
		}
		identExpr := &expr.Ident{
			SpanVal:    collation.Span(),
			Value:      collation.Value,
			QuoteStyle: collation.QuoteStyle,
		}
		return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "COLLATE", Value: &expr.Identifier{Ident: identExpr}}, nil
	}

	// COMMENT 'text'
	if p.ParseKeyword("COMMENT") {
		tok := p.NextToken()
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "COMMENT", Value: &expr.ValueExpr{Value: str.Value}}, nil
		}
		return nil, fmt.Errorf("expected string literal after COMMENT")
	}

	// AUTO_INCREMENT/AUTOINCREMENT (MySQL/Snowflake column option)
	// Supports: AUTO_INCREMENT, AUTOINCREMENT, (seed, increment), START n INCREMENT n [ORDER|NOORDER]
	if p.ParseKeyword("AUTO_INCREMENT") || p.ParseKeyword("AUTOINCREMENT") {
		identProp := parseIdentityProperty(p)
		colIdent := &expr.ColumnIdentity{
			Kind:     expr.IdentityPropertyKindAutoincrement,
			Property: identProp,
		}
		return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "AUTOINCREMENT", Value: colIdent}, nil
	}

	// IDENTITY (MSSQL/Snowflake column option)
	// Supports: IDENTITY, IDENTITY(seed, increment), IDENTITY START n INCREMENT n [ORDER|NOORDER]
	if p.ParseKeyword("IDENTITY") {
		identProp := parseIdentityProperty(p)
		colIdent := &expr.ColumnIdentity{
			Kind:     expr.IdentityPropertyKindIdentity,
			Property: identProp,
		}
		return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "IDENTITY", Value: colIdent}, nil
	}

	// PRIMARY KEY [DEFERRABLE/etc]
	if p.ParseKeywords([]string{"PRIMARY", "KEY"}) {
		characteristics, err := parseConstraintCharacteristics(p)
		if err != nil {
			return nil, err
		}
		return &expr.ColumnOptionDef{
			ConstraintName:  constraintName,
			Name:            "PRIMARY KEY",
			Characteristics: characteristics,
		}, nil
	}

	// UNIQUE [DEFERRABLE/etc]
	if p.ParseKeyword("UNIQUE") {
		characteristics, err := parseConstraintCharacteristics(p)
		if err != nil {
			return nil, err
		}
		return &expr.ColumnOptionDef{
			ConstraintName:  constraintName,
			Name:            "UNIQUE",
			Characteristics: characteristics,
		}, nil
	}

	// CHECK (expr) [ENFORCED/NOT ENFORCED] [DEFERRABLE/etc]
	if p.ParseKeyword("CHECK") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		exprParser := NewExpressionParser(p)
		checkExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("expected expression in CHECK constraint: %w", err)
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		// Parse optional constraint characteristics (ENFORCED, DEFERRABLE, etc.)
		characteristics, err := parseConstraintCharacteristics(p)
		if err != nil {
			return nil, err
		}
		return &expr.ColumnOptionDef{
			ConstraintName:  constraintName,
			Name:            "CHECK",
			Value:           checkExpr,
			Characteristics: characteristics,
		}, nil
	}

	// ON UPDATE CURRENT_TIMESTAMP [(fractional_seconds_precision)] - MySQL column option
	if p.ParseKeywords([]string{"ON", "UPDATE"}) {
		exprParser := NewExpressionParser(p)
		onUpdateExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("expected expression after ON UPDATE: %w", err)
		}
		return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "ON UPDATE", Value: onUpdateExpr}, nil
	}

	// REFERENCES table [(cols)] [ON DELETE action] [ON UPDATE action]
	if p.ParseKeyword("REFERENCES") {
		refTable, err := p.ParseObjectName()
		if err != nil {
			return nil, fmt.Errorf("expected table name after REFERENCES: %w", err)
		}

		// Parse optional column list
		var refCols []*ast.Ident
		if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
			refCols, err = p.ParseParenthesizedColumnList()
			if err != nil {
				return nil, err
			}
		}

		// Parse ON DELETE/ON UPDATE actions
		var onDelete, onUpdate expr.ReferentialAction
		for {
			if p.ParseKeywords([]string{"ON", "DELETE"}) {
				onDelete = parseReferentialAction(p)
			} else if p.ParseKeywords([]string{"ON", "UPDATE"}) {
				onUpdate = parseReferentialAction(p)
			} else {
				break
			}
		}

		// Store all REFERENCES details in the Value field
		refDetails := &expr.ColumnOptionReferences{
			Table:    refTable,
			Columns:  refCols,
			OnDelete: onDelete,
			OnUpdate: onUpdate,
		}

		// Parse optional constraint characteristics (DEFERRABLE, etc.)
		characteristics, err := parseConstraintCharacteristics(p)
		if err != nil {
			return nil, err
		}

		return &expr.ColumnOptionDef{
			ConstraintName:  constraintName,
			Name:            "REFERENCES",
			Value:           refDetails,
			Characteristics: characteristics,
		}, nil
	}

	// GENERATED ALWAYS AS expr STORED/VIRTUAL
	// GENERATED {ALWAYS | BY DEFAULT} AS IDENTITY [(sequence_options)]
	if p.ParseKeyword("GENERATED") {
		// Check for GENERATED ALWAYS AS IDENTITY
		if p.ParseKeywords([]string{"ALWAYS", "AS", "IDENTITY"}) {
			genIdent := &expr.GeneratedIdentity{
				GeneratedAs: expr.GeneratedAsAlways,
			}
			// Check for sequence options in parentheses
			if p.ConsumeToken(token.TokenLParen{}) {
				seqOpts, err := parseSequenceOptionsForIdentity(p)
				if err == nil {
					genIdent.SequenceOptions = seqOpts
				}
				p.ConsumeToken(token.TokenRParen{})
			}
			return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "GENERATED AS IDENTITY", Value: genIdent}, nil
		}

		// Check for GENERATED BY DEFAULT AS IDENTITY
		if p.ParseKeywords([]string{"BY", "DEFAULT", "AS", "IDENTITY"}) {
			genIdent := &expr.GeneratedIdentity{
				GeneratedAs: expr.GeneratedAsByDefault,
			}
			// Check for sequence options in parentheses
			if p.ConsumeToken(token.TokenLParen{}) {
				seqOpts, err := parseSequenceOptionsForIdentity(p)
				if err == nil {
					genIdent.SequenceOptions = seqOpts
				}
				p.ConsumeToken(token.TokenRParen{})
			}
			return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: "GENERATED AS IDENTITY", Value: genIdent}, nil
		}

		// GENERATED ALWAYS AS expr [STORED|VIRTUAL] (not IDENTITY)
		if p.ParseKeywords([]string{"ALWAYS", "AS"}) {
			// GENERATED ALWAYS AS expr [STORED|VIRTUAL]
			exprParser := NewExpressionParser(p)
			genExpr, err := exprParser.ParseExpr()
			if err != nil {
				return nil, fmt.Errorf("expected expression after GENERATED ALWAYS AS: %w", err)
			}

			// Check for STORED or VIRTUAL AFTER parsing expression
			var genType string
			if p.ParseKeyword("STORED") {
				genType = "STORED"
			} else if p.ParseKeyword("VIRTUAL") {
				genType = "VIRTUAL"
			}

			name := "GENERATED ALWAYS AS"
			if genType != "" {
				name += " " + genType
			}
			return &expr.ColumnOptionDef{ConstraintName: constraintName, Name: name, Value: genExpr}, nil
		}

		return nil, fmt.Errorf("expected ALWAYS AS or BY DEFAULT AS IDENTITY after GENERATED")
	}

	return nil, fmt.Errorf("unknown column constraint")
}

// parseIdentityProperty parses optional IDENTITY/AUTOINCREMENT parameters
// Syntax: [(seed, increment)] | [START num INCREMENT num] [ORDER|NOORDER]
func parseIdentityProperty(p *Parser) *expr.IdentityProperty {
	prop := &expr.IdentityProperty{}
	exprParser := NewExpressionParser(p)

	// Check for function-call style: (seed, increment)
	if p.ConsumeToken(token.TokenLParen{}) {
		prop.Format = expr.IdentityFormatFunctionCall

		// Parse seed
		seedExpr, err := exprParser.ParseExpr()
		if err == nil {
			// Parse comma
			if p.ConsumeToken(token.TokenComma{}) {
				// Parse increment
				incExpr, err := exprParser.ParseExpr()
				if err == nil {
					prop.Parameters = &expr.IdentityParameters{
						Seed:      seedExpr,
						Increment: incExpr,
					}
				}
			}
		}
		// Consume closing paren - MUST consume it for this syntax
		if !p.ConsumeToken(token.TokenRParen{}) {
			// No closing paren found, something is wrong
			return prop
		}
		// Check for ORDER/NOORDER after function-call style
		if p.ParseKeyword("ORDER") {
			prop.Order = expr.IdentityOrderOrder
		} else if p.ParseKeyword("NOORDER") {
			prop.Order = expr.IdentityOrderNoOrder
		}
	} else if p.ParseKeyword("START") {
		// START num INCREMENT num style (Snowflake)
		prop.Format = expr.IdentityFormatStartAndIncrement
		startExpr, err := exprParser.ParseExpr()
		if err == nil {
			if p.ParseKeyword("INCREMENT") {
				incExpr, err := exprParser.ParseExpr()
				if err == nil {
					prop.Parameters = &expr.IdentityParameters{
						Seed:      startExpr,
						Increment: incExpr,
					}
				}
			}
		}
		// Check for ORDER/NOORDER after START/INCREMENT style
		if p.ParseKeyword("ORDER") {
			prop.Order = expr.IdentityOrderOrder
		} else if p.ParseKeyword("NOORDER") {
			prop.Order = expr.IdentityOrderNoOrder
		}
	}

	// Check for ORDER/NOORDER (applies to all styles including no parameters)
	if p.ParseKeyword("ORDER") {
		prop.Order = expr.IdentityOrderOrder
	} else if p.ParseKeyword("NOORDER") {
		prop.Order = expr.IdentityOrderNoOrder
	}

	return prop
}

// parseSequenceOptionsForIdentity parses sequence options for GENERATED AS IDENTITY
// Similar to parseCreateSequenceOptions but stops at ) or other delimiters
func parseSequenceOptionsForIdentity(p *Parser) ([]*expr.SequenceOptions, error) {
	var sequenceOptions []*expr.SequenceOptions
	exprParser := NewExpressionParser(p)

	for {
		// Check for closing paren - stop parsing
		if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
			break
		}

		// INCREMENT [BY] increment
		if p.ParseKeyword("INCREMENT") {
			hasBy := p.ParseKeyword("BY")
			incExpr, err := exprParser.ParseExpr()
			if err != nil {
				return sequenceOptions, nil // Return what we have so far
			}
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:        expr.SeqOptIncrementBy,
				Expr:        incExpr,
				HasByOrWith: hasBy,
			})
			continue
		}

		// MINVALUE minvalue | NO MINVALUE
		if p.ParseKeyword("MINVALUE") {
			minExpr, err := exprParser.ParseExpr()
			if err == nil {
				sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
					Type:    expr.SeqOptMinValue,
					Expr:    minExpr,
					NoValue: false,
				})
			}
			continue
		} else if p.ParseKeywords([]string{"NO", "MINVALUE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMinValue,
				NoValue: true,
			})
			continue
		}

		// MAXVALUE maxvalue | NO MAXVALUE
		if p.ParseKeyword("MAXVALUE") {
			maxExpr, err := exprParser.ParseExpr()
			if err == nil {
				sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
					Type:    expr.SeqOptMaxValue,
					Expr:    maxExpr,
					NoValue: false,
				})
			}
			continue
		} else if p.ParseKeywords([]string{"NO", "MAXVALUE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptMaxValue,
				NoValue: true,
			})
			continue
		}

		// START [WITH] start
		if p.ParseKeyword("START") {
			hasWith := p.ParseKeyword("WITH")
			startExpr, err := exprParser.ParseExpr()
			if err == nil {
				sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
					Type:        expr.SeqOptStartWith,
					Expr:        startExpr,
					HasByOrWith: hasWith,
				})
			}
			continue
		}

		// CACHE cache
		if p.ParseKeyword("CACHE") {
			cacheExpr, err := exprParser.ParseExpr()
			if err == nil {
				sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
					Type: expr.SeqOptCache,
					Expr: cacheExpr,
				})
			}
			continue
		}

		// [NO] CYCLE
		if p.ParseKeywords([]string{"NO", "CYCLE"}) {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptCycle,
				NoCycle: true,
			})
			continue
		} else if p.ParseKeyword("CYCLE") {
			sequenceOptions = append(sequenceOptions, &expr.SequenceOptions{
				Type:    expr.SeqOptCycle,
				NoCycle: false,
			})
			continue
		}

		// No more options we recognize
		break
	}

	return sequenceOptions, nil
}

// parseReferentialAction parses ON DELETE/ON UPDATE action
func parseReferentialAction(p *Parser) expr.ReferentialAction {
	switch {
	case p.ParseKeyword("CASCADE"):
		return expr.ReferentialActionCascade
	case p.ParseKeyword("RESTRICT"):
		return expr.ReferentialActionRestrict
	case p.ParseKeywords([]string{"SET", "NULL"}):
		return expr.ReferentialActionSetNull
	case p.ParseKeywords([]string{"SET", "DEFAULT"}):
		return expr.ReferentialActionSetDefault
	case p.ParseKeywords([]string{"NO", "ACTION"}):
		return expr.ReferentialActionNoAction
	default:
		return expr.ReferentialActionNone
	}
}

// parseTableConstraint parses a table-level constraint
// PRIMARY KEY (columns), FOREIGN KEY (columns) REFERENCES ..., UNIQUE (columns), CHECK (expr)
func parseTableConstraint(p *Parser) (*expr.TableConstraint, error) {
	constraint := &expr.TableConstraint{}

	// Check for named constraint: CONSTRAINT name
	if p.ParseKeyword("CONSTRAINT") {
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected constraint name after CONSTRAINT: %w", err)
		}
		constraint.Name = name
	}

	// PRIMARY KEY (columns)
	if p.ParseKeywords([]string{"PRIMARY", "KEY"}) {
		pkConstraint := &expr.PrimaryKeyConstraint{}

		// Optional index name (MySQL)
		if !p.PeekKeyword("USING") {
			if ident, err := p.ParseIdentifier(); err == nil {
				pkConstraint.IndexName = ident
			}
		}

		// Optional USING index_type
		if p.ParseKeyword("USING") {
			if typeIdent, err := p.ParseIdentifier(); err == nil {
				switch strings.ToUpper(typeIdent.Value) {
				case "BTREE":
					btree := expr.IndexTypeBTree
					pkConstraint.IndexType = &btree
				case "HASH":
					hash := expr.IndexTypeHash
					pkConstraint.IndexType = &hash
				}
			}
		}

		// Parse column list
		cols, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
		// Convert []*ast.Ident to []*expr.IndexColumn
		pkConstraint.Columns = make([]*expr.IndexColumn, len(cols))
		for i, col := range cols {
			pkConstraint.Columns[i] = &expr.IndexColumn{
				Expr: &expr.Ident{Value: col.Value},
			}
		}

		// Parse constraint characteristics
		characteristics, err := parseConstraintCharacteristics(p)
		if err != nil {
			return nil, err
		}
		pkConstraint.Characteristics = characteristics
		constraint.Constraint = pkConstraint
		return constraint, nil
	}

	// UNIQUE (columns)
	if p.ParseKeyword("UNIQUE") {
		uniqueConstraint := &expr.UniqueConstraint{}

		// Optional NULLS [NOT] DISTINCT (PostgreSQL)
		if p.ParseKeyword("NULLS") {
			if p.ParseKeyword("NOT") {
				if p.ParseKeyword("DISTINCT") {
					uniqueConstraint.NullsDistinct = expr.NullsDistinctOptionNullsNotDistinct
				}
			} else if p.ParseKeyword("DISTINCT") {
				uniqueConstraint.NullsDistinct = expr.NullsDistinctOptionNullsDistinct
			}
		}

		// Optional index name (MySQL)
		if !p.PeekKeyword("USING") && !p.PeekKeyword("(") {
			if ident, err := p.ParseIdentifier(); err == nil {
				uniqueConstraint.IndexName = ident
			}
		}

		// Optional USING index_type
		if p.ParseKeyword("USING") {
			if typeIdent, err := p.ParseIdentifier(); err == nil {
				switch strings.ToUpper(typeIdent.Value) {
				case "BTREE":
					btree := expr.IndexTypeBTree
					uniqueConstraint.IndexType = &btree
				case "HASH":
					hash := expr.IndexTypeHash
					uniqueConstraint.IndexType = &hash
				}
			}
		}

		// Parse column list
		cols, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
		// Convert []*ast.Ident to []*expr.IndexColumn
		uniqueConstraint.Columns = make([]*expr.IndexColumn, len(cols))
		for i, col := range cols {
			uniqueConstraint.Columns[i] = &expr.IndexColumn{
				Expr: &expr.Ident{Value: col.Value},
			}
		}

		// Parse constraint characteristics
		characteristics, err := parseConstraintCharacteristics(p)
		if err != nil {
			return nil, err
		}
		uniqueConstraint.Characteristics = characteristics
		constraint.Constraint = uniqueConstraint
		return constraint, nil
	}

	// FOREIGN KEY (columns) REFERENCES table [(cols)] [ON DELETE action] [ON UPDATE action]
	if p.ParseKeywords([]string{"FOREIGN", "KEY"}) {
		fkConstraint := &expr.ForeignKeyConstraint{}

		// Parse column list
		cols, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
		fkConstraint.Columns = cols

		if !p.ParseKeyword("REFERENCES") {
			return nil, fmt.Errorf("expected REFERENCES after FOREIGN KEY column list")
		}

		refTable, err := p.ParseObjectName()
		if err != nil {
			return nil, fmt.Errorf("expected table name after REFERENCES: %w", err)
		}
		fkConstraint.ForeignTable = refTable

		// Parse optional reference columns
		if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
			refCols, err := p.ParseParenthesizedColumnList()
			if err != nil {
				return nil, err
			}
			fkConstraint.ReferredColumns = refCols
		}

		// Parse MATCH kind (FULL | PARTIAL | SIMPLE)
		if p.ParseKeyword("MATCH") {
			if p.ParseKeyword("FULL") {
				matchKind := expr.ConstraintReferenceMatchKindFull
				fkConstraint.MatchKind = &matchKind
			} else if p.ParseKeyword("PARTIAL") {
				matchKind := expr.ConstraintReferenceMatchKindPartial
				fkConstraint.MatchKind = &matchKind
			} else if p.ParseKeyword("SIMPLE") {
				matchKind := expr.ConstraintReferenceMatchKindSimple
				fkConstraint.MatchKind = &matchKind
			}
		}

		// Parse ON DELETE/ON UPDATE actions (in any order)
		for {
			if p.ParseKeywords([]string{"ON", "DELETE"}) {
				fkConstraint.OnDelete = parseReferentialAction(p)
			} else if p.ParseKeywords([]string{"ON", "UPDATE"}) {
				fkConstraint.OnUpdate = parseReferentialAction(p)
			} else {
				break
			}
		}

		// Parse constraint characteristics
		characteristics, err := parseConstraintCharacteristics(p)
		if err != nil {
			return nil, err
		}
		fkConstraint.Characteristics = characteristics
		constraint.Constraint = fkConstraint
		return constraint, nil
	}

	// CHECK (expr)
	if p.ParseKeyword("CHECK") {
		checkConstraint := &expr.CheckConstraint{}

		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		exprParser := NewExpressionParser(p)
		checkExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("expected expression in CHECK constraint: %w", err)
		}
		checkConstraint.Expr = checkExpr

		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}

		// Parse constraint characteristics (includes ENFORCED/NOT ENFORCED for MySQL)
		characteristics, err := parseConstraintCharacteristics(p)
		if err != nil {
			return nil, err
		}
		if characteristics != nil && characteristics.Enforced != nil {
			checkConstraint.Enforced = characteristics.Enforced
		}
		constraint.Constraint = checkConstraint
		return constraint, nil
	}

	// MySQL-specific: INDEX/KEY inline index constraints
	// Reference: src/parser/mod.rs:9732-9756
	if p.GetDialect().SupportsIndexHints() {
		if p.ParseKeyword("INDEX") || p.ParseKeyword("KEY") {
			indexConstraint := &expr.IndexConstraint{}

			// Optional index name (skip if USING follows)
			if !p.PeekKeyword("USING") {
				if ident, err := p.ParseIdentifier(); err == nil {
					indexConstraint.Name = ident
				}
			}

			// Optional USING index_type (e.g., USING BTREE, USING HASH)
			if p.ParseKeyword("USING") {
				if typeIdent, err := p.ParseIdentifier(); err == nil {
					switch strings.ToUpper(typeIdent.Value) {
					case "BTREE":
						btree := expr.IndexTypeBTree
						indexConstraint.IndexType = &btree
					case "HASH":
						hash := expr.IndexTypeHash
						indexConstraint.IndexType = &hash
					}
				}
			}

			// Parse column list: (col1, col2, ...)
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			cols, err := parseCommaSeparatedIdents(p)
			if err != nil {
				return nil, err
			}
			// Convert []*ast.Ident to []*expr.IndexColumn
			indexConstraint.Columns = make([]*expr.IndexColumn, len(cols))
			for i, col := range cols {
				indexConstraint.Columns[i] = &expr.IndexColumn{
					Expr: &expr.Ident{Value: col.Value},
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			constraint.Constraint = indexConstraint
			return constraint, nil
		}
	}

	// MySQL-specific: FULLTEXT/SPATIAL index constraints
	// Reference: src/parser/mod.rs:9758-9789
	isFulltext := p.ParseKeyword("FULLTEXT")
	isSpatial := p.ParseKeyword("SPATIAL")
	if isFulltext || isSpatial {
		ftsConstraint := &expr.FullTextOrSpatialConstraint{
			Fulltext: isFulltext,
		}

		// Optional INDEX/KEY keyword
		if !p.ParseKeyword("INDEX") {
			p.ParseKeyword("KEY")
		}

		// Optional index name
		p.ParseIdentifier()

		// Parse column list: (col1, col2, ...)
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		cols, err := parseCommaSeparatedIdents(p)
		if err != nil {
			return nil, err
		}
		// Convert []*ast.Ident to []*expr.IndexColumn
		ftsConstraint.Columns = make([]*expr.IndexColumn, len(cols))
		for i, col := range cols {
			ftsConstraint.Columns[i] = &expr.IndexColumn{
				Expr: &expr.Ident{Value: col.Value},
			}
		}

		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}

		constraint.Constraint = ftsConstraint
		return constraint, nil
	}

	return nil, fmt.Errorf("unknown table constraint")
}

// parseConstraintCharacteristics parses optional DEFERRABLE, INITIALLY DEFERRED/IMMEDIATE, ENFORCED/NOT ENFORCED
func parseConstraintCharacteristics(p *Parser) (*expr.ConstraintCharacteristics, error) {
	characteristics := &expr.ConstraintCharacteristics{}
	hasCharacteristics := false
	// Track what we've already parsed to detect duplicates
	deferrableSet := false
	initiallySet := false
	enforcedSet := false

	for {
		switch {
		case p.ParseKeywords([]string{"NOT", "DEFERRABLE"}):
			if deferrableSet {
				return nil, errors.New("duplicate DEFERRABLE specification")
			}
			deferrableSet = true
			notDeferrable := false
			characteristics.Deferrable = &notDeferrable
			hasCharacteristics = true
		case p.ParseKeyword("DEFERRABLE"):
			if deferrableSet {
				return nil, errors.New("duplicate DEFERRABLE specification")
			}
			deferrableSet = true
			deferrable := true
			characteristics.Deferrable = &deferrable
			hasCharacteristics = true
			// Check for INITIALLY DEFERRED/IMMEDIATE
			if p.ParseKeyword("INITIALLY") {
				if initiallySet {
					return nil, errors.New("duplicate INITIALLY specification")
				}
				initiallySet = true
				if p.ParseKeyword("DEFERRED") {
					initially := expr.ConstraintInitiallyOptionDeferred
					characteristics.Initially = &initially
				} else if p.ParseKeyword("IMMEDIATE") {
					initially := expr.ConstraintInitiallyOptionImmediate
					characteristics.Initially = &initially
				}
			}
		case p.ParseKeyword("INITIALLY"):
			if initiallySet {
				return nil, errors.New("duplicate INITIALLY specification")
			}
			initiallySet = true
			hasCharacteristics = true
			if p.ParseKeyword("DEFERRED") {
				initially := expr.ConstraintInitiallyOptionDeferred
				characteristics.Initially = &initially
			} else if p.ParseKeyword("IMMEDIATE") {
				initially := expr.ConstraintInitiallyOptionImmediate
				characteristics.Initially = &initially
			}
		case p.ParseKeywords([]string{"NOT", "ENFORCED"}):
			if enforcedSet {
				return nil, errors.New("duplicate ENFORCED specification")
			}
			enforcedSet = true
			notEnforced := false
			characteristics.Enforced = &notEnforced
			hasCharacteristics = true
		case p.ParseKeyword("ENFORCED"):
			if enforcedSet {
				return nil, errors.New("duplicate ENFORCED specification")
			}
			enforcedSet = true
			enforced := true
			characteristics.Enforced = &enforced
			hasCharacteristics = true
		case p.PeekKeyword("NOT") && p.PeekNthKeyword(1, "VALID"):
			// Handle NOT VALID (PostgreSQL)
			p.ParseKeyword("NOT")   // consume NOT
			p.ParseKeyword("VALID") // consume VALID
			characteristics.NotValid = true
			hasCharacteristics = true
		default:
			if hasCharacteristics {
				return characteristics, nil
			}
			return nil, nil
		}
	}
}

// parseSetType parses MySQL SET('a', 'b', ...) data type
func parseSetType(p *Parser, span token.Span) (*datatype.SetType, error) {
	if !p.ConsumeToken(token.TokenLParen{}) {
		return nil, p.expected("(", p.PeekToken())
	}

	var values []string
	for {
		tok := p.PeekToken()
		if strTok, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			values = append(values, strTok.Value)
		} else {
			return nil, p.expected("string literal", tok)
		}

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if !p.ConsumeToken(token.TokenRParen{}) {
		return nil, p.expected(")", p.PeekToken())
	}

	return &datatype.SetType{
		SpanVal: span,
		Values:  values,
	}, nil
}

// parseEnumType parses MySQL/ClickHouse ENUM('a', 'b', ...) or ENUM8/ENUM16('a' = 1, 'b' = 2, ...) data type
func parseEnumType(p *Parser, span token.Span, bits *uint8) (*datatype.EnumType, error) {
	if !p.ConsumeToken(token.TokenLParen{}) {
		return nil, p.expected("(", p.PeekToken())
	}

	var members []datatype.EnumMember
	for {
		tok := p.PeekToken()
		if strTok, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			member := datatype.EnumMember{
				Name: strTok.Value,
			}

			// Check for optional value assignment like 'a' = 1 (ClickHouse style)
			if p.ConsumeToken(token.TokenEq{}) {
				valTok := p.PeekToken()
				if numTok, ok := valTok.Token.(token.TokenNumber); ok {
					p.AdvanceToken()
					member.Value = &expr.ValueExpr{Value: numTok.Value}
				}
			}

			members = append(members, member)
		} else {
			return nil, p.expected("string literal", tok)
		}

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if !p.ConsumeToken(token.TokenRParen{}) {
		return nil, p.expected(")", p.PeekToken())
	}

	return &datatype.EnumType{
		SpanVal: span,
		Members: members,
		Bits:    bits,
	}, nil
}

// parseBitType parses BIT[(n)] [VARYING] data type (PostgreSQL, etc.)
// If VARYING is present, returns a BitVaryingType instead
func parseBitType(p *Parser, span token.Span) (datatype.DataType, error) {
	// Check for optional length after BIT like BIT(42)
	var length *uint64
	if p.ConsumeToken(token.TokenLParen{}) {
		tok := p.PeekToken()
		if numTok, ok := tok.Token.(token.TokenNumber); ok {
			p.AdvanceToken()
			if val, err := strconv.ParseUint(numTok.Value, 10, 64); err == nil {
				length = &val
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Check for VARYING keyword after BIT[ (n) ]
	if p.ParseKeyword("VARYING") {
		// Parse optional length after VARYING like BIT VARYING(43)
		if p.ConsumeToken(token.TokenLParen{}) {
			tok := p.PeekToken()
			if numTok, ok := tok.Token.(token.TokenNumber); ok {
				p.AdvanceToken()
				if val, err := strconv.ParseUint(numTok.Value, 10, 64); err == nil {
					length = &val
				}
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		}
		// Return BIT VARYING type
		return &datatype.BitVaryingType{
			SpanVal: span,
			Length:  length,
		}, nil
	}

	// Return regular BIT type
	return &datatype.BitType{
		SpanVal: span,
		Length:  length,
	}, nil
}

// parseBitVaryingType parses VARBIT[(n)] data type (PostgreSQL alias for BIT VARYING)
func parseBitVaryingType(p *Parser, span token.Span) (*datatype.BitVaryingType, error) {
	result := &datatype.BitVaryingType{SpanVal: span}

	// Parse optional length like VARBIT(43)
	if p.ConsumeToken(token.TokenLParen{}) {
		tok := p.PeekToken()
		if numTok, ok := tok.Token.(token.TokenNumber); ok {
			p.AdvanceToken()
			if val, err := strconv.ParseUint(numTok.Value, 10, 64); err == nil {
				result.Length = &val
			}
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	return result, nil
}
