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
	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/tokenizer"
)

// ParseInsert parses an INSERT statement
func ParseInsert(p *Parser, tok tokenizer.TokenWithSpan) (ast.Statement, error) {
	return parseInsertInternal(p, tok)
}

// parseInsertInternal parses an INSERT statement
func parseInsertInternal(p *Parser, insertToken tokenizer.TokenWithSpan) (ast.Statement, error) {
	// Parse optimizer hints if present (MySQL style: INSERT /*+ hint */)
	hintsInterface, err := maybeParseOptimizerHints(p)
	if err != nil {
		return nil, err
	}

	// Convert optimizer hints to proper type
	var optimizerHints []*expr.OptimizerHint
	for _, h := range hintsInterface {
		if hint, ok := h.(*expr.OptimizerHint); ok {
			optimizerHints = append(optimizerHints, hint)
		}
	}

	// Parse optional OVERWRITE keyword (used in some dialects like Hive)
	overwrite := p.ParseKeyword("OVERWRITE")

	// Parse optional INTO keyword
	into := p.ParseKeyword("INTO")

	// Parse optional TABLE keyword (Hive allows TABLE here)
	hasTableKeyword := p.ParseKeyword("TABLE")

	// Parse the table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional table alias (MySQL allows this)
	var tableAlias *ast.Ident
	if p.dialect.SupportsInsertTableAlias() {
		// Try to parse AS alias or just implicit alias
		if p.ParseKeyword("AS") {
			alias, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			tableAlias = alias
		} else {
			// Try implicit alias (not followed by certain keywords)
			if !isInsertReservedKeyword(p.PeekToken()) {
				alias, err := p.ParseIdentifier()
				if err == nil {
					tableAlias = alias
				}
			}
		}
	}

	// Check for DEFAULT VALUES
	if p.ParseKeywords([]string{"DEFAULT", "VALUES"}) {
		// Parse optional ON CONFLICT clause (PostgreSQL style)
		var onConflict *expr.OnInsert
		if p.ParseKeyword("ON") {
			if p.ParseKeyword("CONFLICT") {
				// Parse conflict target (optional)
				var conflictTarget *expr.ConflictTarget
				if p.ParseKeywords([]string{"ON", "CONSTRAINT"}) {
					constraintName, err := p.ParseObjectName()
					if err != nil {
						return nil, err
					}
					conflictTarget = &expr.ConflictTarget{
						OnConstraint: constraintName,
					}
				} else if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
					columns, err := p.ParseParenthesizedColumnList()
					if err != nil {
						return nil, err
					}
					conflictTarget = &expr.ConflictTarget{
						Columns: columns,
					}
				}

				if !p.ParseKeyword("DO") {
					return nil, p.Expected("DO", p.PeekToken())
				}

				var action expr.OnConflictAction
				if p.ParseKeyword("NOTHING") {
					action = expr.OnConflictAction{DoNothing: true}
				} else if p.ParseKeyword("UPDATE") {
					if !p.ParseKeyword("SET") {
						return nil, p.Expected("SET", p.PeekToken())
					}
					assignments, err := parseAssignments(p)
					if err != nil {
						return nil, err
					}
					var selection expr.Expr
					if p.ParseKeyword("WHERE") {
						ep := NewExpressionParser(p)
						selection, err = ep.ParseExpr()
						if err != nil {
							return nil, err
						}
					}
					action = expr.OnConflictAction{
						DoUpdate: &expr.DoUpdate{
							Assignments: assignments,
							Selection:   selection,
						},
					}
				} else {
					return nil, p.Expected("NOTHING or UPDATE", p.PeekToken())
				}

				onConflict = &expr.OnInsert{
					OnConflict: &expr.OnConflict{
						ConflictTarget: conflictTarget,
						Action:         action,
					},
				}
			}
		}

		// Parse optional RETURNING clause
		var returning []*query.SelectItem
		if p.ParseKeyword("RETURNING") {
			items, err := parseProjection(p)
			if err != nil {
				return nil, err
			}
			for i := range items {
				item := items[i]
				returning = append(returning, &item)
			}
		}

		insert := &statement.Insert{
			OptimizerHints:  optimizerHints,
			Into:            into,
			Overwrite:       overwrite,
			Table:           tableName,
			TableAlias:      tableAlias,
			HasTableKeyword: hasTableKeyword,
			Columns:         []*ast.Ident{},
			DefaultValues:   true,
			Assignments:     []*expr.Assignment{},
			Returning:       returning,
			On:              onConflict,
		}
		insert.SetSpan(insertToken.Span)
		return insert, nil
	}

	// Parse optional column list (col1, col2, ...)
	var columns []*ast.Ident
	if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
		columns, err = p.ParseParenthesizedColumnList()
		if err != nil {
			return nil, err
		}
	}

	// Parse the source: either VALUES or SELECT
	var source *query.Query

	if p.PeekKeyword("VALUES") || p.PeekKeyword("VALUE") {
		// Parse VALUES clause
		valuesStmt, err := parseValues(p)
		if err != nil {
			return nil, err
		}
		source = valuesStmt.Query
	} else if p.PeekKeyword("SELECT") || p.PeekKeyword("WITH") {
		// Parse SELECT query (with optional CTE)
		selectStmt, err := parseQuery(p)
		if err != nil {
			return nil, err
		}
		if selectStmt != nil {
			source = &query.Query{
				Body: selectStmt,
			}
		}
	} else {
		return nil, p.Expected("VALUES or SELECT", p.PeekToken())
	}

	// Parse optional ON CONFLICT clause (PostgreSQL style)
	var onConflict *expr.OnInsert
	if p.ParseKeyword("ON") {
		if p.ParseKeyword("CONFLICT") {
			// Parse conflict target (optional)
			var conflictTarget *expr.ConflictTarget
			if p.ParseKeywords([]string{"ON", "CONSTRAINT"}) {
				// ON CONSTRAINT constraint_name
				constraintName, err := p.ParseObjectName()
				if err != nil {
					return nil, err
				}
				conflictTarget = &expr.ConflictTarget{
					OnConstraint: constraintName,
				}
			} else if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
				// (column1, column2, ...)
				columns, err := p.ParseParenthesizedColumnList()
				if err != nil {
					return nil, err
				}
				conflictTarget = &expr.ConflictTarget{
					Columns: columns,
				}
			}

			// Expect DO
			if !p.ParseKeyword("DO") {
				return nil, p.Expected("DO", p.PeekToken())
			}

			// Parse action: NOTHING or UPDATE SET ...
			var action expr.OnConflictAction
			if p.ParseKeyword("NOTHING") {
				action = expr.OnConflictAction{DoNothing: true}
			} else if p.ParseKeyword("UPDATE") {
				if !p.ParseKeyword("SET") {
					return nil, p.Expected("SET", p.PeekToken())
				}

				// Parse assignments
				assignments, err := parseAssignments(p)
				if err != nil {
					return nil, err
				}

				// Optional WHERE clause
				var selection expr.Expr
				if p.ParseKeyword("WHERE") {
					ep := NewExpressionParser(p)
					selection, err = ep.ParseExpr()
					if err != nil {
						return nil, err
					}
				}

				action = expr.OnConflictAction{
					DoUpdate: &expr.DoUpdate{
						Assignments: assignments,
						Selection:   selection,
					},
				}
			} else {
				return nil, p.Expected("NOTHING or UPDATE", p.PeekToken())
			}

			onConflict = &expr.OnInsert{
				OnConflict: &expr.OnConflict{
					ConflictTarget: conflictTarget,
					Action:         action,
				},
			}
		} else {
			// Not ON CONFLICT, might be something else - skip for now
			// This should handle ON DUPLICATE KEY UPDATE (MySQL)
		}
	}

	// Parse optional RETURNING clause (PostgreSQL style)
	var returning []*query.SelectItem
	if p.ParseKeyword("RETURNING") {
		items, err := parseProjection(p)
		if err != nil {
			return nil, err
		}
		// Convert []query.SelectItem to []*query.SelectItem
		for i := range items {
			item := items[i]
			returning = append(returning, &item)
		}
	}

	insert := &statement.Insert{
		OptimizerHints:  optimizerHints,
		Into:            into,
		Overwrite:       overwrite,
		Table:           tableName,
		TableAlias:      tableAlias,
		HasTableKeyword: hasTableKeyword,
		Columns:         columns,
		Source:          source,
		Assignments:     []*expr.Assignment{},
		Returning:       returning,
		On:              onConflict,
	}
	insert.SetSpan(insertToken.Span)
	return insert, nil
}

// isInsertReservedKeyword checks if a token is a reserved keyword that cannot be a table alias
func isInsertReservedKeyword(tok tokenizer.TokenWithSpan) bool {
	if word, ok := tok.Token.(tokenizer.TokenWord); ok {
		kw := word.Word.Keyword
		// Keywords that signal the end of table name/alias
		reserved := map[string]bool{
			"VALUES":    true,
			"VALUE":     true,
			"SELECT":    true,
			"DEFAULT":   true,
			"SET":       true,
			"RETURNING": true,
		}
		return reserved[string(kw)]
	}
	return false
}

// ParseUpdate parses UPDATE statements
func ParseUpdate(p *Parser, tok tokenizer.TokenWithSpan) (ast.Statement, error) {
	return parseUpdateInternal(p, tok)
}

// parseUpdateInternal parses an UPDATE statement
// Basic syntax: UPDATE table SET col = val [, col2 = val2] [WHERE condition]
func parseUpdateInternal(p *Parser, updateToken tokenizer.TokenWithSpan) (ast.Statement, error) {
	// Parse optimizer hints if present
	hintsInterface, err := maybeParseOptimizerHints(p)
	if err != nil {
		return nil, err
	}

	// Convert optimizer hints to proper type
	var optimizerHints []*expr.OptimizerHint
	for _, h := range hintsInterface {
		if hint, ok := h.(*expr.OptimizerHint); ok {
			optimizerHints = append(optimizerHints, hint)
		}
	}

	// Parse table name
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Expect SET keyword
	if _, err := p.ExpectKeyword("SET"); err != nil {
		return nil, err
	}

	// Create expression parser for parsing assignment values
	ep := NewExpressionParser(p)

	// Parse comma-separated assignments
	var assignments []*expr.Assignment
	err = p.ParseCommaSeparated(func() error {
		// Parse column identifier
		col, err := p.ParseIdentifier()
		if err != nil {
			return err
		}

		// Expect = token
		if _, err := p.ExpectToken(tokenizer.TokenEq{}); err != nil {
			return err
		}

		// Parse expression using ExpressionParser to get expr.Expr
		val, err := ep.ParseExpr()
		if err != nil {
			return err
		}

		assignment := &expr.Assignment{
			Column: col,
			Value:  val,
		}
		assignments = append(assignments, assignment)
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Parse optional WHERE clause
	var selection expr.Expr
	if p.ParseKeyword("WHERE") {
		selection, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional RETURNING clause (PostgreSQL style)
	var returning []*query.SelectItem
	if p.ParseKeyword("RETURNING") {
		items, err := parseProjection(p)
		if err != nil {
			return nil, err
		}
		// Convert []query.SelectItem to []*query.SelectItem
		for i := range items {
			item := items[i]
			returning = append(returning, &item)
		}
	}

	update := &statement.Update{
		Table:       tableName,
		Assignments: assignments,
		Selection:   selection,
		Returning:   returning,
	}
	update.SetSpan(updateToken.Span)

	// Store optimizer hints if the field exists (it might not yet)
	_ = optimizerHints // Avoid unused variable error if field doesn't exist yet

	return update, nil
}

// ParseDelete parses DELETE statements
func ParseDelete(p *Parser, tok tokenizer.TokenWithSpan) (ast.Statement, error) {
	return parseDeleteInternal(p, tok)
}

// parseDeleteInternal parses a DELETE statement
// Basic syntax: DELETE FROM table [WHERE condition]
func parseDeleteInternal(p *Parser, deleteToken tokenizer.TokenWithSpan) (ast.Statement, error) {
	// Parse optimizer hints if present
	hintsInterface, err := maybeParseOptimizerHints(p)
	if err != nil {
		return nil, err
	}

	// Convert optimizer hints to proper type
	var optimizerHints []*expr.OptimizerHint
	for _, h := range hintsInterface {
		if hint, ok := h.(*expr.OptimizerHint); ok {
			optimizerHints = append(optimizerHints, hint)
		}
	}

	// Optional FROM keyword (required by most dialects but optional in BigQuery)
	hasFromKeyword := p.ParseKeyword("FROM")

	// Parse table name(s)
	var tables []*ast.ObjectName
	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}
	tables = append(tables, tableName)

	// Check for additional tables (multi-table DELETE - MySQL style)
	// Example: DELETE t1, t2 FROM t1, t2 WHERE ...
	if p.ConsumeToken(tokenizer.TokenComma{}) {
		err = p.ParseCommaSeparated(func() error {
			t, err := p.ParseObjectName()
			if err != nil {
				return err
			}
			tables = append(tables, t)
			return nil
		})
		if err != nil {
			return nil, err
		}

		// For multi-table delete, expect FROM keyword after table list
		if !hasFromKeyword {
			if _, err := p.ExpectKeyword("FROM"); err != nil {
				return nil, err
			}
			hasFromKeyword = true
		}
	}

	// Create expression parser for parsing WHERE conditions
	ep := NewExpressionParser(p)

	// Parse optional WHERE clause
	var selection expr.Expr
	if p.ParseKeyword("WHERE") {
		selection, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	// Parse optional RETURNING clause (PostgreSQL style)
	var returning []*query.SelectItem
	if p.ParseKeyword("RETURNING") {
		items, err := parseProjection(p)
		if err != nil {
			return nil, err
		}
		// Convert []query.SelectItem to []*query.SelectItem
		for i := range items {
			item := items[i]
			returning = append(returning, &item)
		}
	}

	delete := &statement.Delete{
		Tables:    tables,
		Selection: selection,
		Returning: returning,
	}
	delete.SetSpan(deleteToken.Span)

	// Store optimizer hints and from keyword info if fields exist (they might not yet)
	_ = optimizerHints // Avoid unused variable error if field doesn't exist yet
	_ = hasFromKeyword

	return delete, nil
}

// parseAssignments parses a comma-separated list of assignments (col = val)
// Used for ON CONFLICT DO UPDATE SET ...
func parseAssignments(p *Parser) ([]*expr.Assignment, error) {
	var assignments []*expr.Assignment

	err := p.ParseCommaSeparated(func() error {
		// Parse column identifier
		col, err := p.ParseIdentifier()
		if err != nil {
			return err
		}

		// Expect = token
		if _, err := p.ExpectToken(tokenizer.TokenEq{}); err != nil {
			return err
		}

		// Parse expression using ExpressionParser to get expr.Expr
		ep := NewExpressionParser(p)
		val, err := ep.ParseExpr()
		if err != nil {
			return err
		}

		assignment := &expr.Assignment{
			Column: col,
			Value:  val,
		}
		assignments = append(assignments, assignment)
		return nil
	})

	if err != nil {
		return nil, err
	}

	return assignments, nil
}
