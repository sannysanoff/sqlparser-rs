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

// parseMerge parses a MERGE statement
// Reference: src/parser/merge.rs
func parseMerge(p *Parser, tok token.TokenWithSpan) (*statement.Merge, error) {
	return parseMergeInternal(p, tok)
}

// parseMergeInternal parses a MERGE statement
func parseMergeInternal(p *Parser, mergeToken token.TokenWithSpan) (*statement.Merge, error) {
	// Parse optimizer hints if present
	// TODO: optimizerHints := p.maybeParseOptimizerHints()

	// Parse optional INTO keyword
	into := p.ParseKeyword("INTO")

	// Parse target table using table factor
	tableFactor, err := parseTableFactor(p)
	if err != nil {
		return nil, fmt.Errorf("expected table after MERGE: %w", err)
	}

	// Expect USING
	if !p.ParseKeyword("USING") {
		return nil, fmt.Errorf("expected USING after table in MERGE statement")
	}

	// Parse source table/subquery
	sourceFactor, err := parseTableFactor(p)
	if err != nil {
		return nil, fmt.Errorf("expected source after USING in MERGE: %w", err)
	}

	// Expect ON
	if !p.ParseKeyword("ON") {
		return nil, fmt.Errorf("expected ON after source in MERGE statement")
	}

	// Parse join condition using ExpressionParser
	ep := NewExpressionParser(p)
	onExpr, err := ep.ParseExpr()
	if err != nil {
		return nil, fmt.Errorf("expected expression after ON in MERGE: %w", err)
	}

	// Parse merge clauses
	clauses, err := parseMergeClauses(p)
	if err != nil {
		return nil, err
	}

	// Parse optional OUTPUT/RETURNING clause
	var output *expr.OutputClause
	var outputToken token.Token
	var returningToken token.Token

	// Check for OUTPUT or RETURNING without consuming yet
	tok := p.PeekToken()
	if isWordToken(tok.Token) {
		word := getWordValue(tok.Token)
		if word == "OUTPUT" {
			p.AdvanceToken() // consume OUTPUT
			outputToken = tok.Token
			outputClause, err := parseOutputClauseInternal(p, true) // true = is OUTPUT
			if err != nil {
				return nil, err
			}
			output = outputClause
			output.OutputToken = &outputToken
		} else if word == "RETURNING" {
			p.AdvanceToken() // consume RETURNING
			returningToken = tok.Token
			outputClause, err := parseOutputClauseInternal(p, false) // false = is RETURNING
			if err != nil {
				return nil, err
			}
			output = outputClause
			output.ReturningToken = &returningToken
		}
	}

	return &statement.Merge{
		Into:    into,
		Table:   tableFactor,
		Source:  sourceFactor,
		On:      onExpr,
		Clauses: clauses,
		Output:  output,
	}, nil
}

// parseMergeClauses parses WHEN clauses in a MERGE statement
func parseMergeClauses(p *Parser) ([]*expr.MergeClause, error) {
	var clauses []*expr.MergeClause

	for p.ParseKeyword("WHEN") {
		whenToken := p.PeekToken()

		clauseKind := expr.MergeClauseKindMatched
		if p.ParseKeyword("NOT") {
			clauseKind = expr.MergeClauseKindNotMatched
		}

		if !p.ParseKeyword("MATCHED") {
			return nil, fmt.Errorf("expected MATCHED after WHEN in MERGE clause")
		}

		// Check for BY SOURCE / BY TARGET modifiers
		if clauseKind == expr.MergeClauseKindNotMatched && p.ParseKeyword("BY") {
			if p.ParseKeyword("SOURCE") {
				clauseKind = expr.MergeClauseKindNotMatchedBySource
			} else if p.ParseKeyword("TARGET") {
				clauseKind = expr.MergeClauseKindNotMatchedByTarget
			} else {
				return nil, fmt.Errorf("expected SOURCE or TARGET after BY")
			}
		}

		// Parse optional AND predicate
		var predicate expr.Expr
		if p.ParseKeyword("AND") {
			ep := NewExpressionParser(p)
			var err error
			predicate, err = ep.ParseExpr()
			if err != nil {
				return nil, fmt.Errorf("expected expression after AND in MERGE clause: %w", err)
			}
		}

		// Expect THEN
		if !p.ParseKeyword("THEN") {
			return nil, fmt.Errorf("expected THEN after WHEN clause in MERGE")
		}

		// Parse action: UPDATE, DELETE, or INSERT
		action, err := parseMergeAction(p, clauseKind)
		if err != nil {
			return nil, err
		}

		clauses = append(clauses, &expr.MergeClause{
			WhenToken:  &whenToken.Token,
			ClauseKind: clauseKind,
			Predicate:  predicate,
			Action:     action,
		})
	}

	return clauses, nil
}

// parseMergeAction parses UPDATE, DELETE, or INSERT action in a MERGE clause
func parseMergeAction(p *Parser, clauseKind expr.MergeClauseKind) (*expr.MergeAction, error) {
	ep := NewExpressionParser(p)

	if p.ParseKeyword("UPDATE") {
		// UPDATE not allowed in NOT MATCHED or NOT MATCHED BY TARGET clauses
		if clauseKind == expr.MergeClauseKindNotMatched || clauseKind == expr.MergeClauseKindNotMatchedByTarget {
			return nil, fmt.Errorf("UPDATE is not allowed in a NOT MATCHED merge clause")
		}

		updateToken := p.PeekToken()

		if !p.ParseKeyword("SET") {
			return nil, fmt.Errorf("expected SET after UPDATE in MERGE clause")
		}

		// Parse assignments
		assignments, err := parseCommaSeparatedAssignments(p)
		if err != nil {
			return nil, fmt.Errorf("expected assignments after SET in MERGE: %w", err)
		}

		// Parse optional WHERE clause (Oracle specific)
		var updatePredicate expr.Expr
		if p.ParseKeyword("WHERE") {
			updatePredicate, err = ep.ParseExpr()
			if err != nil {
				return nil, fmt.Errorf("expected expression after WHERE in MERGE UPDATE: %w", err)
			}
		}

		// Parse optional DELETE WHERE clause (Oracle specific)
		var deletePredicate expr.Expr
		if p.ParseKeyword("DELETE") {
			if !p.ParseKeyword("WHERE") {
				return nil, fmt.Errorf("expected WHERE after DELETE in MERGE UPDATE")
			}
			deletePredicate, err = ep.ParseExpr()
			if err != nil {
				return nil, fmt.Errorf("expected expression after DELETE WHERE in MERGE: %w", err)
			}
		}

		return &expr.MergeAction{
			Update: &expr.MergeUpdateExpr{
				UpdateToken:     &updateToken.Token,
				Assignments:     assignments,
				UpdatePredicate: updatePredicate,
				DeletePredicate: deletePredicate,
			},
		}, nil

	} else if p.ParseKeyword("DELETE") {
		// DELETE not allowed in NOT MATCHED or NOT MATCHED BY TARGET clauses
		if clauseKind == expr.MergeClauseKindNotMatched || clauseKind == expr.MergeClauseKindNotMatchedByTarget {
			return nil, fmt.Errorf("DELETE is not allowed in a NOT MATCHED merge clause")
		}

		deleteToken := p.PeekToken()

		return &expr.MergeAction{
			Delete: &deleteToken.Token,
		}, nil

	} else if p.ParseKeyword("INSERT") {
		// INSERT only allowed in NOT MATCHED or NOT MATCHED BY TARGET clauses
		if clauseKind != expr.MergeClauseKindNotMatched && clauseKind != expr.MergeClauseKindNotMatchedByTarget {
			return nil, fmt.Errorf("INSERT is not allowed in a MATCHED merge clause")
		}

		insertToken := p.PeekToken()

		// Parse optional column list
		var columns []*ast.ObjectName
		if p.ConsumeToken(token.TokenLParen{}) {
			// Parse parenthesized column list - each column can be a simple or compound identifier
			for {
				// Check for closing paren first
				if p.ConsumeToken(token.TokenRParen{}) {
					break
				}

				// Expect comma before subsequent columns
				if len(columns) > 0 {
					if !p.ConsumeToken(token.TokenComma{}) {
						return nil, fmt.Errorf("expected comma or ) in column list")
					}
				}

				// Parse column identifier (possibly compound like FOO.ID)
				col, err := p.ParseIdentifier()
				if err != nil {
					return nil, fmt.Errorf("expected column name: %w", err)
				}

				// Check for compound identifier (table.column)
				parts := []*ast.Ident{col}
				for p.ConsumeToken(token.TokenPeriod{}) {
					nextIdent, err := p.ParseIdentifier()
					if err != nil {
						return nil, fmt.Errorf("expected identifier after '.': %w", err)
					}
					parts = append(parts, nextIdent)
				}

				// Create ObjectName from the identifier parts
				objNameParts := make([]ast.ObjectNamePart, len(parts))
				for i, part := range parts {
					objNameParts[i] = &ast.ObjectNamePartIdentifier{Ident: part}
				}
				astObjName := &ast.ObjectName{
					Parts: objNameParts,
				}
				columns = append(columns, astObjName)
			}
		}

		// Parse VALUES or ROW
		kind := expr.MergeInsertKindValues
		kindToken := p.PeekToken()
		var values []expr.Expr

		if p.ParseKeyword("ROW") {
			kind = expr.MergeInsertKindRow
		} else if p.ParseKeyword("VALUES") {
			kind = expr.MergeInsertKindValues
			kindToken = p.PeekToken()

			// Parse parenthesized values list
			if p.ConsumeToken(token.TokenLParen{}) {
				for {
					if p.ConsumeToken(token.TokenRParen{}) {
						break
					}
					if len(values) > 0 && !p.ConsumeToken(token.TokenComma{}) {
						return nil, fmt.Errorf("expected comma or ) in VALUES list")
					}
					val, err := ep.ParseExpr()
					if err != nil {
						return nil, fmt.Errorf("expected expression in VALUES: %w", err)
					}
					values = append(values, val)
				}
			}
		} else {
			return nil, fmt.Errorf("expected VALUES or ROW after INSERT in MERGE clause")
		}

		// Parse optional WHERE clause (Oracle specific)
		var insertPredicate expr.Expr
		if p.ParseKeyword("WHERE") {
			var err error
			insertPredicate, err = ep.ParseExpr()
			if err != nil {
				return nil, fmt.Errorf("expected expression after WHERE in MERGE INSERT: %w", err)
			}
		}

		return &expr.MergeAction{
			Insert: &expr.MergeInsertExpr{
				InsertToken:     &insertToken.Token,
				Columns:         columns,
				KindToken:       &kindToken.Token,
				Kind:            kind,
				Values:          values,
				InsertPredicate: insertPredicate,
			},
		}, nil
	}

	return nil, fmt.Errorf("expected UPDATE, DELETE or INSERT in merge clause")
}

// parseCommaSeparatedAssignments parses a comma-separated list of assignments
func parseCommaSeparatedAssignments(p *Parser) ([]*expr.Assignment, error) {
	var assignments []*expr.Assignment

	for {
		assign, err := parseAssignment(p)
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, assign)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return assignments, nil
}

// parseAssignment parses a single assignment (col = expr)
// Handles both simple column names (col) and qualified names (table.col)
func parseAssignment(p *Parser) (*expr.Assignment, error) {
	// Parse column identifier (possibly compound like "dest.F")
	col, err := p.ParseIdentifier()
	if err != nil {
		return nil, fmt.Errorf("expected column name in assignment: %w", err)
	}

	// Check for compound identifier (table.column)
	for {
		if !p.ConsumeToken(token.TokenPeriod{}) {
			break
		}
		nextIdent, err := p.ParseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected identifier after '.': %w", err)
		}
		// Create a compound identifier
		col = &ast.Ident{
			Value: col.Value + "." + nextIdent.Value,
		}
	}

	// Expect =
	if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
		return nil, fmt.Errorf("expected = after column name in assignment")
	}

	// Parse value expression using ExpressionParser
	ep := NewExpressionParser(p)
	val, err := ep.ParseExpr()
	if err != nil {
		return nil, fmt.Errorf("expected expression after = in assignment: %w", err)
	}

	return &expr.Assignment{
		Column: col,
		Value:  val,
	}, nil
}

// parseOutputClauseInternal parses the OUTPUT or RETURNING clause
// Reference: src/parser/merge.rs:235-260
func parseOutputClauseInternal(p *Parser, isOutput bool) (*expr.OutputClause, error) {
	// Parse comma-separated select items (same as projection)
	selectItems, err := parseProjection(p)
	if err != nil {
		return nil, fmt.Errorf("expected select items in OUTPUT/RETURNING clause: %w", err)
	}

	// For OUTPUT clause, optionally parse INTO table
	var intoTable *query.SelectInto
	if isOutput && p.ParseKeyword("INTO") {
		intoTable, err = parseSelectIntoInternal(p)
		if err != nil {
			return nil, fmt.Errorf("expected table name after INTO: %w", err)
		}
	}

	return &expr.OutputClause{
		SelectItems: selectItems,
		IntoTable:   intoTable,
	}, nil
}

// parseSelectIntoInternal parses the INTO clause for OUTPUT
// Reference: src/parser/mod.rs:19011-19025
func parseSelectIntoInternal(p *Parser) (*query.SelectInto, error) {
	temporary := false
	unlogged := false
	table := false

	// Check for TEMP/TEMPORARY keywords
	if p.ParseKeyword("TEMP") || p.ParseKeyword("TEMPORARY") {
		temporary = true
	}

	// Check for UNLOGGED
	if p.ParseKeyword("UNLOGGED") {
		unlogged = true
	}

	// Check for TABLE keyword
	if p.ParseKeyword("TABLE") {
		table = true
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, fmt.Errorf("expected table name after INTO: %w", err)
	}

	return &query.SelectInto{
		Temporary: temporary,
		Unlogged:  unlogged,
		Table:     table,
		Name:      astObjectNameToQuery(tableName),
	}, nil
}
