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
	"fmt"

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
			p.PeekKeyword("AUTO_INCREMENT") {

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
	if p.ParseKeyword("CONSTRAINT") {
		_, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected constraint name after CONSTRAINT: %w", err)
		}
		// Continue to parse the actual constraint
	}

	// NOT NULL
	if p.ParseKeywords([]string{"NOT", "NULL"}) {
		return &expr.ColumnOptionDef{Name: "NOT NULL"}, nil
	}

	// NULL
	if p.ParseKeyword("NULL") {
		return &expr.ColumnOptionDef{Name: "NULL"}, nil
	}

	// DEFAULT expr
	if p.ParseKeyword("DEFAULT") {
		exprParser := NewExpressionParser(p)
		defaultExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("expected expression after DEFAULT: %w", err)
		}
		return &expr.ColumnOptionDef{Name: "DEFAULT", Value: defaultExpr}, nil
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
		return &expr.ColumnOptionDef{Name: "COLLATE", Value: &expr.Identifier{Ident: identExpr}}, nil
	}

	// COMMENT 'text'
	if p.ParseKeyword("COMMENT") {
		tok := p.NextToken()
		if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			return &expr.ColumnOptionDef{Name: "COMMENT", Value: &expr.ValueExpr{Value: str.Value}}, nil
		}
		return nil, fmt.Errorf("expected string literal after COMMENT")
	}

	// AUTO_INCREMENT (MySQL column option)
	if p.ParseKeyword("AUTO_INCREMENT") {
		return &expr.ColumnOptionDef{Name: "AUTO_INCREMENT"}, nil
	}

	// PRIMARY KEY
	if p.ParseKeywords([]string{"PRIMARY", "KEY"}) {
		return &expr.ColumnOptionDef{Name: "PRIMARY KEY"}, nil
	}

	// UNIQUE
	if p.ParseKeyword("UNIQUE") {
		return &expr.ColumnOptionDef{Name: "UNIQUE"}, nil
	}

	// CHECK (expr)
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
		return &expr.ColumnOptionDef{Name: "CHECK", Value: checkExpr}, nil
	}

	// ON UPDATE CURRENT_TIMESTAMP [(fractional_seconds_precision)] - MySQL column option
	if p.ParseKeywords([]string{"ON", "UPDATE"}) {
		exprParser := NewExpressionParser(p)
		onUpdateExpr, err := exprParser.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("expected expression after ON UPDATE: %w", err)
		}
		return &expr.ColumnOptionDef{Name: "ON UPDATE", Value: onUpdateExpr}, nil
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

		_ = refTable
		_ = refCols
		_ = onDelete
		_ = onUpdate

		return &expr.ColumnOptionDef{Name: "REFERENCES"}, nil
	}

	// GENERATED ALWAYS AS expr STORED/VIRTUAL
	// GENERATED {ALWAYS | BY DEFAULT} AS IDENTITY
	if p.ParseKeyword("GENERATED") {
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
			return &expr.ColumnOptionDef{Name: name, Value: genExpr}, nil
		}

		if p.ParseKeywords([]string{"BY", "DEFAULT", "AS", "IDENTITY"}) {
			return &expr.ColumnOptionDef{Name: "GENERATED BY DEFAULT AS IDENTITY"}, nil
		}

		return nil, fmt.Errorf("expected ALWAYS AS or BY DEFAULT AS IDENTITY after GENERATED")
	}

	return nil, fmt.Errorf("unknown column constraint")
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
		cols, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
		_ = cols
		// Parse constraint characteristics
		parseConstraintCharacteristics(p)
		return constraint, nil
	}

	// UNIQUE (columns)
	if p.ParseKeyword("UNIQUE") {
		cols, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
		_ = cols
		parseConstraintCharacteristics(p)
		return constraint, nil
	}

	// FOREIGN KEY (columns) REFERENCES table [(cols)] [ON DELETE action] [ON UPDATE action]
	if p.ParseKeywords([]string{"FOREIGN", "KEY"}) {
		cols, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}

		if !p.ParseKeyword("REFERENCES") {
			return nil, fmt.Errorf("expected REFERENCES after FOREIGN KEY column list")
		}

		refTable, err := p.ParseObjectName()
		if err != nil {
			return nil, fmt.Errorf("expected table name after REFERENCES: %w", err)
		}

		// Parse optional reference columns
		var refCols []*ast.Ident
		if _, isLParen := p.PeekToken().Token.(token.TokenLParen); isLParen {
			refCols, err = p.ParseParenthesizedColumnList()
			if err != nil {
				return nil, err
			}
		}

		// Parse ON DELETE/ON UPDATE actions (in any order)
		for {
			if p.ParseKeywords([]string{"ON", "DELETE"}) {
				_ = parseReferentialAction(p)
			} else if p.ParseKeywords([]string{"ON", "UPDATE"}) {
				_ = parseReferentialAction(p)
			} else {
				break
			}
		}

		// Parse constraint characteristics
		parseConstraintCharacteristics(p)

		_ = cols
		_ = refTable
		_ = refCols
		return constraint, nil
	}

	// CHECK (expr)
	if p.ParseKeyword("CHECK") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		exprParser := NewExpressionParser(p)
		_, err := exprParser.ParseExpr()
		if err != nil {
			return nil, fmt.Errorf("expected expression in CHECK constraint: %w", err)
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		parseConstraintCharacteristics(p)
		return constraint, nil
	}

	// MySQL-specific: INDEX/KEY inline index constraints
	// Reference: src/parser/mod.rs:9732-9756
	if p.GetDialect().SupportsIndexHints() {
		if p.ParseKeyword("INDEX") || p.ParseKeyword("KEY") {
			// Optional index name (skip if USING follows)
			if !p.PeekKeyword("USING") {
				p.ParseIdentifier()
			}

			// Optional USING index_type (e.g., USING BTREE, USING HASH)
			if p.ParseKeyword("USING") {
				p.ParseIdentifier() // consume index type
			}

			// Parse column list: (col1, col2, ...)
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			_, err := parseCommaSeparatedIdents(p)
			if err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			return constraint, nil
		}

		// MySQL-specific: FULLTEXT/SPATIAL index constraints
		// Reference: src/parser/mod.rs:9758-9789
		isFulltext := p.ParseKeyword("FULLTEXT")
		isSpatial := p.ParseKeyword("SPATIAL")
		if isFulltext || isSpatial {
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
			_, err := parseCommaSeparatedIdents(p)
			if err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			return constraint, nil
		}
	}

	return nil, fmt.Errorf("unknown table constraint")
}

// parseConstraintCharacteristics parses optional DEFERRABLE, INITIALLY DEFERRED/IMMEDIATE, ENFORCED/NOT ENFORCED
func parseConstraintCharacteristics(p *Parser) {
	for {
		switch {
		case p.ParseKeywords([]string{"NOT", "DEFERRABLE"}):
			// NOT DEFERRABLE
		case p.ParseKeyword("DEFERRABLE"):
			// Check for INITIALLY DEFERRED/IMMEDIATE
			if p.ParseKeyword("INITIALLY") {
				if p.ParseKeyword("DEFERRED") {
					// INITIALLY DEFERRED
				} else if p.ParseKeyword("IMMEDIATE") {
					// INITIALLY IMMEDIATE
				}
			}
		case p.ParseKeyword("INITIALLY"):
			if p.ParseKeyword("DEFERRED") {
				// INITIALLY DEFERRED
			} else if p.ParseKeyword("IMMEDIATE") {
				// INITIALLY IMMEDIATE
			}
		case p.ParseKeywords([]string{"NOT", "ENFORCED"}):
			// NOT ENFORCED
		case p.ParseKeyword("ENFORCED"):
			// ENFORCED
		default:
			return
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

// parseEnumType parses MySQL ENUM('a', 'b', ...) data type
func parseEnumType(p *Parser, span token.Span) (*datatype.EnumType, error) {
	if !p.ConsumeToken(token.TokenLParen{}) {
		return nil, p.expected("(", p.PeekToken())
	}

	var members []datatype.EnumMember
	for {
		tok := p.PeekToken()
		if strTok, ok := tok.Token.(token.TokenSingleQuotedString); ok {
			p.AdvanceToken()
			members = append(members, datatype.EnumMember{
				Name: strTok.Value,
			})
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
	}, nil
}
