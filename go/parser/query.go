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
	"strconv"
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/operator"
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/ast/statement"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/dialects/snowflake"
	"github.com/user/sqlparser/parseriface"
	"github.com/user/sqlparser/token"
)

// SelectStatement wraps query.Select to implement ast.Statement
type SelectStatement struct {
	ast.BaseStatement
	query.Select
	LimitClause   query.LimitClause
	ForClause     *query.ForClause     // MSSQL FOR XML/FOR JSON clause
	PipeOperators []query.PipeOperator // Pipe operators for |> syntax
}

// Span returns the source span
func (s *SelectStatement) Span() token.Span {
	return s.Select.Span()
}

// String returns the SQL representation including LIMIT clause
func (s *SelectStatement) String() string {
	str := s.Select.String()
	if s.LimitClause != nil {
		str = str + " " + s.LimitClause.String()
	}
	if s.ForClause != nil {
		str = str + " " + s.ForClause.String()
	}
	for _, pipe := range s.PipeOperators {
		str = str + " |> " + pipe.String()
	}
	return str
}

// ValuesStatement wraps query.Query (for VALUES) to implement ast.Statement
type ValuesStatement struct {
	ast.BaseStatement
	Query *query.Query
}

// Span returns the source span
func (v *ValuesStatement) Span() token.Span {
	if v.Query != nil {
		return v.Query.Span()
	}
	return token.Span{}
}

// String returns the SQL representation
func (v *ValuesStatement) String() string {
	if v.Query != nil {
		return v.Query.String()
	}
	return ""
}

// QueryStatement wraps query.Query (for SELECT WITH CTE) to implement ast.Statement
type QueryStatement struct {
	ast.BaseStatement
	Query *query.Query
}

// Span returns the source span
func (q *QueryStatement) Span() token.Span {
	if q.Query != nil {
		return q.Query.Span()
	}
	return token.Span{}
}

// String returns the SQL representation
func (q *QueryStatement) String() string {
	if q.Query != nil {
		return q.Query.String()
	}
	return ""
}

// parseQuery parses a SELECT or other query statement
// Reference: src/parser/mod.rs:13599 parse_query
func parseQuery(p *Parser) (ast.Statement, error) {
	// Check recursion limit
	if err := p.recursionCounter.TryDecrease(); err != nil {
		return nil, err
	}
	defer p.recursionCounter.Increase()

	// Check for WITH clause (Common Table Expressions)
	var withClause *query.With
	if p.ParseKeyword("WITH") {
		recursive := p.ParseKeyword("RECURSIVE")

		// Parse comma-separated list of CTEs
		ctes, err := parseCTEList(p)
		if err != nil {
			return nil, err
		}

		withClause = &query.With{
			Recursive: recursive,
			CteTables: ctes,
		}
	}

	// Parse the actual query body (SELECT, INSERT, UPDATE, DELETE, etc.)
	var body ast.Statement
	var err error

	if p.PeekKeyword("SELECT") {
		body, err = parseSelect(p)
	} else if p.PeekKeyword("VALUES") {
		body, err = parseValues(p)
	} else {
		if withClause != nil {
			return nil, p.ExpectedRef("SELECT or VALUES after WITH", p.PeekTokenRef())
		}
		return nil, p.ExpectedRef("SELECT, VALUES, or WITH", p.PeekTokenRef())
	}

	if err != nil {
		return nil, err
	}

	// Check if we need to create a Query wrapper (for WITH clause, pipe operators, FOR clause, or locks)
	needsQueryWrapper := withClause != nil

	// Get pipe operators, FOR clause, and locks from SelectStatement if present
	var pipeOperators []query.PipeOperator
	var forClause *query.ForClause
	var locks []query.LockClause
	if selStmt, ok := body.(*SelectStatement); ok {
		pipeOperators = selStmt.PipeOperators
		forClause = selStmt.ForClause
		locks = selStmt.Select.Locks
		if len(pipeOperators) > 0 || forClause != nil || len(locks) > 0 {
			needsQueryWrapper = true
		}
	}

	// Create a Query statement if needed
	if needsQueryWrapper {
		if selStmt, ok := body.(*SelectStatement); ok {
			return &statement.Query{
				Query: &query.Query{
					With: withClause,
					Body: &query.SelectSetExpr{
						Select: &selStmt.Select,
					},
					Locks:         locks,
					ForClause:     forClause,
					PipeOperators: pipeOperators,
				},
			}, nil
		}
	}

	return body, nil
}

// parseQueryBody parses the body of a query (SELECT, VALUES, etc.)
func parseQueryBody(p *Parser) (ast.Statement, error) {
	// Check for parenthesized subquery: (SELECT ...) or (VALUES ...)
	if p.ConsumeToken(token.TokenLParen{}) {
		// Parse the inner query
		innerQuery, err := parseQuery(p)
		if err != nil {
			return nil, err
		}
		// Expect closing parenthesis
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return innerQuery, nil
	}

	if p.PeekKeyword("SELECT") {
		return parseSelect(p)
	} else if p.PeekKeyword("VALUES") {
		return parseValues(p)
	}

	return nil, p.ExpectedRef("SELECT or VALUES", p.PeekTokenRef())
}

// parseSelectModifiers parses MySQL SELECT modifiers in any order.
// Also handles DISTINCT, DISTINCTROW, and ALL which can appear intermixed with modifiers.
// Reference: src/parser/mod.rs:14538 parse_select_modifiers
// Returns: (modifiers, distinct)
// Modifiers: HIGH_PRIORITY, STRAIGHT_JOIN, SQL_SMALL_RESULT, SQL_BIG_RESULT,
// SQL_BUFFER_RESULT, SQL_NO_CACHE, SQL_CALC_FOUND_ROWS
func parseSelectModifiers(p *Parser) (*query.SelectModifiers, *query.Distinct, error) {
	modifiers := &query.SelectModifiers{}
	var distinct *query.Distinct

	modifierKeywords := map[string]func() bool{
		"ALL": func() bool {
			if distinct == nil {
				dVal := query.DistinctAll
				distinct = &dVal
				return true
			}
			return false
		},
		"DISTINCT": func() bool {
			if distinct == nil {
				dVal := query.DistinctDistinct
				distinct = &dVal
				return true
			}
			return false
		},
		"DISTINCTROW": func() bool {
			if distinct == nil {
				dVal := query.DistinctDistinct // DISTINCTROW is alias for DISTINCT
				distinct = &dVal
				return true
			}
			return false
		},
		"HIGH_PRIORITY":       func() bool { modifiers.HighPriority = true; return true },
		"STRAIGHT_JOIN":       func() bool { modifiers.StraightJoin = true; return true },
		"SQL_SMALL_RESULT":    func() bool { modifiers.SqlSmallResult = true; return true },
		"SQL_BIG_RESULT":      func() bool { modifiers.SqlBigResult = true; return true },
		"SQL_BUFFER_RESULT":   func() bool { modifiers.SqlBufferResult = true; return true },
		"SQL_NO_CACHE":        func() bool { modifiers.SqlNoCache = true; return true },
		"SQL_CALC_FOUND_ROWS": func() bool { modifiers.SqlCalcFoundRows = true; return true },
	}

	for {
		found := false
		for kw, setFn := range modifierKeywords {
			if p.ParseKeyword(kw) {
				setFn()
				found = true
				break
			}
		}
		if !found {
			break
		}
	}

	// Return nil if no modifiers were actually set
	if !modifiers.IsAnySet() {
		modifiers = nil
	}

	return modifiers, distinct, nil
}

// parseSelect parses a SELECT statement
func parseSelect(p *Parser) (ast.Statement, error) {
	// Expect SELECT keyword
	_, err := p.ExpectKeyword("SELECT")
	if err != nil {
		return nil, err
	}

	// Parse optimizer hints (MySQL /*+ ... */ style)
	// TODO: parse optimizer hints

	// Parse MySQL SELECT modifiers (HIGH_PRIORITY, STRAIGHT_JOIN, etc.)
	// Also handles DISTINCT/DISTINCTROW/ALL which can appear intermixed with modifiers.
	// Reference: src/parser/mod.rs:14285-14290
	var selectModifiers *query.SelectModifiers
	var distinct *query.Distinct
	if p.GetDialect().SupportsSelectModifiers() {
		selectModifiers, distinct, err = parseSelectModifiers(p)
		if err != nil {
			return nil, err
		}
	}

	// Parse DISTINCT / ALL if not already parsed by parseSelectModifiers
	if distinct == nil {
		if p.ParseKeyword("DISTINCT") {
			distinctVal := query.DistinctDistinct
			distinct = &distinctVal
		} else if p.ParseKeyword("ALL") {
			distinctVal := query.DistinctAll
			distinct = &distinctVal
		}
	}

	// Parse TOP clause (MSSQL) - can appear before or after DISTINCT depending on dialect
	var top *query.Top
	var topBeforeDistinct bool
	if p.GetDialect().SupportsTopBeforeDistinct() && p.ParseKeyword("TOP") {
		top = parseTop(p)
		topBeforeDistinct = true
	}

	// Parse DISTINCT / ALL if not already parsed by parseSelectModifiers
	if distinct == nil {
		if p.ParseKeyword("DISTINCT") {
			distinctVal := query.DistinctDistinct
			distinct = &distinctVal
		} else if p.ParseKeyword("ALL") {
			distinctVal := query.DistinctAll
			distinct = &distinctVal
		}
	}

	// Parse TOP after DISTINCT for dialects that don't support TOP before DISTINCT
	if top == nil && !p.GetDialect().SupportsTopBeforeDistinct() && p.ParseKeyword("TOP") {
		top = parseTop(p)
		topBeforeDistinct = false
	}

	// Parse projection (select list)
	projection, err := parseProjection(p)
	if err != nil {
		return nil, err
	}

	// Parse FROM clause
	var from []query.TableWithJoins
	if p.ParseKeyword("FROM") {
		from, err = parseTableWithJoinsList(p)
		if err != nil {
			return nil, err
		}
	}

	// Parse WHERE clause
	var selection query.Expr
	if p.ParseKeyword("WHERE") {
		ep := NewExpressionParser(p)
		expr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		// Convert expr.Expr to query.Expr
		selection = &queryExprWrapper{expr: expr}
	}

	// Parse GROUP BY clause
	var groupBy query.GroupByExpr
	if p.ParseKeyword("GROUP") {
		if !p.ParseKeyword("BY") {
			return nil, p.Expected("BY after GROUP", p.PeekToken())
		}

		// Check for GROUP BY ALL (Snowflake/DuckDB/ClickHouse)
		if p.ParseKeyword("ALL") {
			modifiers, err := parseGroupByModifiers(p)
			if err != nil {
				return nil, err
			}
			// Check for GROUPING SETS after modifiers
			if groupingSets := parseGroupingSetsModifier(p); groupingSets != nil {
				modifiers = append(modifiers, groupingSets)
			}
			groupBy = &query.GroupByAll{
				Modifiers: modifiers,
			}
		} else {
			// Parse expressions with support for empty tuple
			groupExprs, err := parseGroupByExpressions(p)
			if err != nil {
				return nil, err
			}
			modifiers, err := parseGroupByModifiers(p)
			if err != nil {
				return nil, err
			}
			// Check for GROUPING SETS after modifiers
			if groupingSets := parseGroupingSetsModifier(p); groupingSets != nil {
				modifiers = append(modifiers, groupingSets)
			}
			groupBy = &query.GroupByExpressions{
				Expressions: groupExprs,
				Modifiers:   modifiers,
			}
		}
	}

	// Parse HAVING clause
	var having query.Expr
	if p.ParseKeyword("HAVING") {
		ep := NewExpressionParser(p)
		expr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		having = &queryExprWrapper{expr: expr}
	}

	// Parse WINDOW and QUALIFY clauses (order depends on dialect)
	var namedWindow []query.NamedWindowDefinition
	var qualify query.Expr
	windowBeforeQualify := false

	// Check for QUALIFY first (BigQuery style)
	if p.ParseKeyword("QUALIFY") {
		ep := NewExpressionParser(p)
		expr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		qualify = &queryExprWrapper{expr: expr}

		// Check for WINDOW after QUALIFY
		if p.ParseKeyword("WINDOW") {
			namedWindow, err = parseNamedWindows(p)
			if err != nil {
				return nil, err
			}
		}
	} else if p.ParseKeyword("WINDOW") {
		// WINDOW before QUALIFY (DuckDB style)
		windowBeforeQualify = true
		namedWindow, err = parseNamedWindows(p)
		if err != nil {
			return nil, err
		}

		// Check for QUALIFY after WINDOW
		if p.ParseKeyword("QUALIFY") {
			ep := NewExpressionParser(p)
			expr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			qualify = &queryExprWrapper{expr: expr}
		}
	}

	// Parse ORDER BY clause
	if p.ParseKeyword("ORDER") {
		p.ParseKeyword("BY")
		_, err = parseOrderByExpressions(p)
		if err != nil {
			return nil, err
		}
		// TODO: Store orderBy in result - needs to be added to SelectStatement
	}

	// Parse LIMIT clause
	var limitClause query.LimitClause
	if p.ParseKeyword("LIMIT") {
		ep := NewExpressionParser(p)
		firstExpr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		// Check for MySQL LIMIT offset,limit syntax (LIMIT 10, 5)
		if p.ConsumeToken(token.TokenComma{}) {
			// MySQL style: LIMIT offset, limit
			secondExpr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			limitClause = &query.OffsetCommaLimit{
				Offset: &queryExprWrapper{expr: firstExpr},
				Limit:  &queryExprWrapper{expr: secondExpr},
			}
		} else if p.ParseKeyword("OFFSET") {
			// PostgreSQL style: LIMIT x OFFSET y
			offsetExpr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			limitClause = &query.LimitOffset{
				Limit:  &queryExprWrapper{expr: firstExpr},
				Offset: &query.Offset{Value: &queryExprWrapper{expr: offsetExpr}},
			}
		} else {
			// Standard LIMIT expr
			limitClause = &query.LimitOffset{
				Limit: &queryExprWrapper{expr: firstExpr},
			}
		}
	} else if p.ParseKeyword("OFFSET") {
		// OFFSET without LIMIT (OFFSET first style)
		ep := NewExpressionParser(p)
		offsetExpr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		// Check for LIMIT after OFFSET (alternative syntax)
		if p.ParseKeyword("LIMIT") {
			limitExpr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			limitClause = &query.LimitOffset{
				Limit:  &queryExprWrapper{expr: limitExpr},
				Offset: &query.Offset{Value: &queryExprWrapper{expr: offsetExpr}},
			}
		} else {
			// Just OFFSET without LIMIT
			limitClause = &query.Offset{
				Value: &queryExprWrapper{expr: offsetExpr},
			}
		}
	}

	// Parse FOR clause (MSSQL FOR XML/FOR JSON/FOR BROWSE or lock clauses FOR UPDATE/FOR SHARE)
	// Reference: src/parser/mod.rs:13682-13691, 18493-18519
	var forClause *query.ForClause
	var locks []query.LockClause

	for {
		if p.ParseKeyword("FOR") {
			// Try to parse as FOR XML/JSON/BROWSE
			if p.ParseKeyword("XML") {
				fc, err := parseForXml(p)
				if err != nil {
					return nil, err
				}
				forClause = fc
				break
			} else if p.ParseKeyword("JSON") {
				fc, err := parseForJson(p)
				if err != nil {
					return nil, err
				}
				forClause = fc
				break
			} else if p.ParseKeyword("BROWSE") {
				forClause = &query.ForClause{Type: &query.ForBrowseClause{}}
				break
			} else if p.ParseKeyword("UPDATE") {
				// FOR UPDATE lock clause
				lock := query.LockClause{
					LockType: query.LockTypeUpdate,
				}
				// Optional OF table_name
				if p.ParseKeyword("OF") {
					objName, err := p.ParseObjectName()
					if err != nil {
						return nil, err
					}
					queryName := astObjectNameToQuery(objName)
					lock.Of = &queryName
				}
				// Optional NOWAIT or SKIP LOCKED
				if p.ParseKeyword("NOWAIT") {
					nb := query.NonBlockNowait
					lock.Nonblock = &nb
				} else if p.ParseKeyword("SKIP") {
					if p.ParseKeyword("LOCKED") {
						nb := query.NonBlockSkipLocked
						lock.Nonblock = &nb
					}
				}
				locks = append(locks, lock)
			} else if p.ParseKeyword("SHARE") {
				// FOR SHARE lock clause
				lock := query.LockClause{
					LockType: query.LockTypeShare,
				}
				// Optional OF table_name
				if p.ParseKeyword("OF") {
					objName, err := p.ParseObjectName()
					if err != nil {
						return nil, err
					}
					queryName := astObjectNameToQuery(objName)
					lock.Of = &queryName
				}
				// Optional NOWAIT or SKIP LOCKED
				if p.ParseKeyword("NOWAIT") {
					nb := query.NonBlockNowait
					lock.Nonblock = &nb
				} else if p.ParseKeyword("SKIP") {
					if p.ParseKeyword("LOCKED") {
						nb := query.NonBlockSkipLocked
						lock.Nonblock = &nb
					}
				}
				locks = append(locks, lock)
			} else {
				// Unknown FOR clause
				break
			}
		} else {
			break
		}
	}

	// Parse pipe operators (BigQuery/DuckDB |> syntax)
	pipeOperators, err := parsePipeOperators(p)
	if err != nil {
		return nil, err
	}

	return &SelectStatement{
		Select: query.Select{
			SelectModifiers:     selectModifiers,
			Distinct:            distinct,
			Top:                 top,
			TopBeforeDistinct:   topBeforeDistinct,
			Projection:          projection,
			From:                from,
			Selection:           selection,
			GroupBy:             groupBy,
			Having:              having,
			NamedWindow:         namedWindow,
			Qualify:             qualify,
			WindowBeforeQualify: windowBeforeQualify,
			Flavor:              query.SelectFlavorStandard,
			Locks:               locks,
		},
		LimitClause:   limitClause,
		ForClause:     forClause,
		PipeOperators: pipeOperators,
	}, nil
}

// queryExprWrapper wraps expr.Expr to implement query.Expr
type queryExprWrapper struct {
	expr interface{} // Actually expr.Expr
}

func (w *queryExprWrapper) String() string {
	if s, ok := w.expr.(fmt.Stringer); ok {
		return s.String()
	}
	return ""
}

// parseProjection parses the SELECT list
func parseProjection(p *Parser) ([]query.SelectItem, error) {
	var items []query.SelectItem

	for {
		item, err := parseSelectItem(p)
		if err != nil {
			return nil, err
		}
		items = append(items, item)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}

		// Check if next token is a clause-starting keyword (trailing comma check)
		// This prevents eating a comma before a clause keyword like FROM, WHERE, etc.
		// We only check for clause keywords, not expression keywords like NULL.
		nextTok := p.PeekToken()
		if isWordToken(nextTok.Token) {
			word := getWordValue(nextTok.Token)
			if word != "" && isClauseKeyword(word) {
				// Put back the comma - it's not a separator but part of the next clause
				p.PrevToken()
				break
			}
		}
	}

	return items, nil
}

// parseSelectItem parses a single SELECT item
func parseSelectItem(p *Parser) (query.SelectItem, error) {
	// Check for wildcard *
	if p.ConsumeToken(token.TokenMul{}) {
		opts, err := parseWildcardAdditionalOptions(p)
		if err != nil {
			return nil, err
		}
		return &query.Wildcard{AdditionalOptions: opts}, nil
	}

	// Check for qualified wildcard (table.*)
	// Look ahead: identifier followed by . *
	if isQualifiedWildcard(p) {
		// Parse just the table name (single identifier, not full object name)
		// Don't use ParseObjectName() because it would consume the . and try to parse more
		tableIdent, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		p.ConsumeToken(token.TokenPeriod{}) // consume the .
		p.ConsumeToken(token.TokenMul{})    // consume the *

		// Create query.ObjectName from the single identifier
		queryName := query.ObjectName{
			Parts: []query.Ident{{Value: tableIdent.Value}},
		}

		// Parse wildcard options (EXCLUDE, REPLACE, etc.)
		opts, err := parseWildcardAdditionalOptions(p)
		if err != nil {
			return nil, err
		}

		return &query.QualifiedWildcard{
			Kind:              &query.ObjectNameWildcard{Name: queryName},
			AdditionalOptions: opts,
		}, nil
	}

	// Parse expression
	ep := NewExpressionParser(p)
	parsedExpr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Check for = alias assignment (e.g., SELECT alias = expr FROM t)
	// This is supported by MSSQL and some other dialects
	if binOp, ok := parsedExpr.(*expr.BinaryOp); ok {
		if binOp.Op == operator.BOpEq {
			if leftIdent, ok := binOp.Left.(*expr.Identifier); ok {
				if dialects.SupportsEqAliasAssignment(p.GetDialect()) {
					// Convert expr.Ident to query.Ident
					aliasIdent := query.Ident{Value: leftIdent.Ident.Value}
					if leftIdent.Ident.QuoteStyle != nil {
						q := byte(*leftIdent.Ident.QuoteStyle)
						aliasIdent.QuoteStyle = &q
					}
					return &query.AliasedExpr{
						Expr:  &queryExprWrapper{expr: binOp.Right},
						Alias: aliasIdent,
					}, nil
				}
			}
		}
	}

	// Check for alias (AS name or just name)
	if p.ParseKeyword("AS") {
		aliasIdent, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &query.AliasedExpr{
			Expr:  &queryExprWrapper{expr: parsedExpr},
			Alias: astIdentToQuery(aliasIdent),
		}, nil
	}

	// Try implicit alias
	if tok := p.PeekToken(); isWordToken(tok.Token) {
		word := getWordValue(tok.Token)
		if word != "" && !isReservedForColumnAlias(word) {
			aliasIdent, err := p.ParseIdentifier()
			if err == nil {
				return &query.AliasedExpr{
					Expr:  &queryExprWrapper{expr: parsedExpr},
					Alias: astIdentToQuery(aliasIdent),
				}, nil
			}
		}
	}

	return &query.UnnamedExpr{Expr: &queryExprWrapper{expr: parsedExpr}}, nil
}

// parseWildcardAdditionalOptions parses optional EXCLUDE, EXCEPT, REPLACE, RENAME, ILIKE for wildcards
func parseWildcardAdditionalOptions(p *Parser) (query.WildcardAdditionalOptions, error) {
	var opts query.WildcardAdditionalOptions
	dialect := p.GetDialect()

	// Parse ILIKE (Snowflake)
	if dialects.SupportsSelectWildcardIlike(dialect) {
		if p.ParseKeyword("ILIKE") {
			tok := p.PeekToken()
			if str, ok := tok.Token.(token.TokenSingleQuotedString); ok {
				p.AdvanceToken()
				opts.OptIlike = &query.IlikeSelectItem{Pattern: str.Value}
			}
		}
	}

	// Parse EXCLUDE (Snowflake) - only if ILIKE wasn't present
	if opts.OptIlike == nil && dialects.SupportsSelectWildcardExclude(dialect) {
		if p.ParseKeyword("EXCLUDE") {
			// Check for parenthesized list or single column
			if p.ConsumeToken(token.TokenLParen{}) {
				cols, err := parseQueryObjectNames(p)
				if err != nil {
					return opts, err
				}
				if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
					return opts, err
				}
				opts.OptExclude = &query.ExcludeSelectItemMultiple{Columns: cols}
			} else {
				// Single column without parens
				col, err := p.ParseObjectName()
				if err != nil {
					return opts, err
				}
				queryName := astObjectNameToQuery(col)
				opts.OptExclude = &query.ExcludeSelectItemSingle{Column: queryName}
			}
		}
	}

	// Parse EXCEPT (BigQuery/ClickHouse)
	if dialects.SupportsSelectWildcardExcept(dialect) {
		if p.ParseKeyword("EXCEPT") {
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return opts, err
			}
			first, err := p.ParseIdentifier()
			if err != nil {
				return opts, err
			}
			var additional []ast.Ident
			for p.ConsumeToken(token.TokenComma{}) {
				ident, err := p.ParseIdentifier()
				if err != nil {
					return opts, err
				}
				additional = append(additional, *ident)
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return opts, err
			}
			opts.OptExcept = &query.ExceptSelectItem{
				FirstElement:       astIdentToQuery(first),
				AdditionalElements: astIdentsToQuery(additional),
			}
		}
	}

	// Parse REPLACE (BigQuery/ClickHouse/Snowflake)
	if dialects.SupportsSelectWildcardReplace(dialect) {
		if p.ParseKeyword("REPLACE") {
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return opts, err
			}
			items, err := parseCommaSeparatedReplaceElements(p)
			if err != nil {
				return opts, err
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return opts, err
			}
			opts.OptReplace = &query.ReplaceSelectItem{Items: items}
		}
	}

	// Parse RENAME (Snowflake)
	if dialects.SupportsSelectWildcardRename(dialect) {
		if p.ParseKeyword("RENAME") {
			if p.ConsumeToken(token.TokenLParen{}) {
				idents, err := parseCommaSeparatedIdentsWithAlias(p)
				if err != nil {
					return opts, err
				}
				if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
					return opts, err
				}
				opts.OptRename = &query.RenameSelectItemMultiple{
					Columns: idents,
				}
			} else {
				// Single rename without parens
				oldName, err := p.ParseIdentifier()
				if err != nil {
					return opts, err
				}
				if !p.ParseKeyword("AS") {
					return opts, fmt.Errorf("expected AS after RENAME identifier")
				}
				newName, err := p.ParseIdentifier()
				if err != nil {
					return opts, err
				}
				opts.OptRename = &query.RenameSelectItemSingle{
					Column: query.IdentWithAlias{
						Ident: astIdentToQuery(oldName), Alias: astIdentToQuery(newName),
					},
				}
			}
		}
	}

	return opts, nil
}

// parseQueryObjectNames parses a comma-separated list of object names for query package
func parseQueryObjectNames(p *Parser) ([]query.ObjectName, error) {
	var names []query.ObjectName
	for {
		name, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		// Convert ast.ObjectName to query.ObjectName
		queryName := astObjectNameToQuery(name)
		names = append(names, queryName)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return names, nil
}

// parseCommaSeparatedReplaceElements parses comma-separated REPLACE elements
func parseCommaSeparatedReplaceElements(p *Parser) ([]*query.ReplaceSelectElement, error) {
	var items []*query.ReplaceSelectElement
	ep := NewExpressionParser(p)

	for {
		expr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		asKeyword := p.ParseKeyword("AS")
		colName, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		items = append(items, &query.ReplaceSelectElement{
			Expr:       &queryExprWrapper{expr: expr},
			ColumnName: astIdentToQuery(colName),
			AsKeyword:  asKeyword,
		})
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return items, nil
}

// parseCommaSeparatedIdentsWithAlias parses comma-separated identifiers with aliases
func parseCommaSeparatedIdentsWithAlias(p *Parser) ([]query.IdentWithAlias, error) {
	var result []query.IdentWithAlias
	for {
		oldName, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		if !p.ParseKeyword("AS") {
			return nil, fmt.Errorf("expected AS in RENAME clause")
		}
		newName, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		result = append(result, query.IdentWithAlias{
			Ident: astIdentToQuery(oldName),
			Alias: astIdentToQuery(newName),
		})
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return result, nil
}

// astIdentsToQuery converts []ast.Ident to []query.Ident
func astIdentsToQuery(idents []ast.Ident) []query.Ident {
	result := make([]query.Ident, len(idents))
	for i, id := range idents {
		result[i] = astIdentToQuery(&id)
	}
	return result
}

// isQualifiedWildcard checks if next tokens form table.* pattern
func isQualifiedWildcard(p *Parser) bool {
	// Need at least 3 tokens ahead
	tok0 := p.PeekToken()
	if !isWordToken(tok0.Token) {
		return false
	}

	tok1 := p.PeekNthToken(1)
	if _, ok := tok1.Token.(token.TokenPeriod); !ok {
		return false
	}

	tok2 := p.PeekNthToken(2)
	if _, ok := tok2.Token.(token.TokenMul); !ok {
		return false
	}

	return true
}

// isWordToken checks if token is a word/identifier
func isWordToken(tok token.Token) bool {
	_, ok := tok.(token.TokenWord)
	return ok
}

// getWordValue extracts the keyword value from a word token
func getWordValue(tok token.Token) string {
	if word, ok := tok.(token.TokenWord); ok {
		return word.Word.Value
	}
	return ""
}

// isClauseKeyword checks if a keyword starts a new clause (FROM, WHERE, etc.)
// This is used for trailing comma detection in projection parsing.
// Unlike isReservedForColumnAlias, this only includes clause-starting keywords,
// not expression keywords like NULL, TRUE, etc.
func isClauseKeyword(keyword string) bool {
	keyword = strings.ToUpper(keyword)
	clauseKeywords := map[string]bool{
		"FROM": true, "WHERE": true, "GROUP": true, "HAVING": true,
		"ORDER": true, "LIMIT": true, "UNION": true, "INTERSECT": true,
		"EXCEPT": true, "WINDOW": true, "QUALIFY": true, "INTO": true,
		"FOR": true, // FOR XML, FOR JSON, FOR BROWSE, or lock clauses
	}
	return clauseKeywords[keyword]
}

// isReservedForColumnAlias checks if a keyword cannot be used as a column alias
func isReservedForColumnAlias(keyword string) bool {
	keyword = strings.ToUpper(keyword)
	reserved := map[string]bool{
		"FROM": true, "WHERE": true, "GROUP": true, "HAVING": true,
		"ORDER": true, "LIMIT": true, "UNION": true, "INTERSECT": true,
		"EXCEPT": true, "SELECT": true, "INSERT": true, "UPDATE": true,
		"DELETE": true, "CREATE": true, "DROP": true, "ALTER": true,
		"AND": true, "OR": true, "NOT": true, "IN": true,
		"BETWEEN": true, "LIKE": true, "ILIKE": true, "IS": true,
		"NULL": true, "TRUE": true, "FALSE": true,
		"WINDOW": true, "QUALIFY": true, "INTO": true,
		"FOR": true, // FOR XML, FOR JSON, FOR BROWSE, lock clauses
	}
	return reserved[keyword]
}

// astObjectNameToQuery converts ast.ObjectName to query.ObjectName
func astObjectNameToQuery(name *ast.ObjectName) query.ObjectName {
	parts := make([]query.Ident, len(name.Parts))
	for i, part := range name.Parts {
		if identPart, ok := part.(*ast.ObjectNamePartIdentifier); ok {
			parts[i] = query.Ident{Value: identPart.Ident.Value}
		}
	}
	return query.ObjectName{Parts: parts}
}

// astIdentToQuery converts ast.Ident to query.Ident
func astIdentToQuery(ident *ast.Ident) query.Ident {
	return query.Ident{Value: ident.Value}
}

// queryObjectNameToAst converts query.ObjectName to *ast.ObjectName
func queryObjectNameToAst(name query.ObjectName) *ast.ObjectName {
	parts := make([]ast.ObjectNamePart, len(name.Parts))
	for i, part := range name.Parts {
		parts[i] = &ast.ObjectNamePartIdentifier{
			Ident: &ast.Ident{Value: part.Value},
		}
	}
	return &ast.ObjectName{Parts: parts}
}

// queryExprToAstExpr converts query.Expr to expr.Expr
// This unwraps queryExprWrapper if the query.Expr is a wrapped expr.Expr
func queryExprToAstExpr(qExpr query.Expr) expr.Expr {
	if qExpr == nil {
		return nil
	}
	// Check if it's our wrapper type
	if wrapper, ok := qExpr.(*queryExprWrapper); ok {
		if e, ok := wrapper.expr.(expr.Expr); ok {
			return e
		}
	}
	return nil
}

// parseTableWithJoinsList parses a comma-separated list of tables
func parseTableWithJoinsList(p *Parser) ([]query.TableWithJoins, error) {
	var tables []query.TableWithJoins

	for {
		table, err := parseTableWithJoins(p)
		if err != nil {
			return nil, err
		}
		tables = append(tables, table)

		// Check for comma (cross join) or JOIN keywords
		if p.ConsumeToken(token.TokenComma{}) {
			continue
		}

		// Check for join keywords
		nextTok := p.PeekToken()
		if isJoinKeyword(nextTok) {
			continue
		}

		break
	}

	return tables, nil
}

// isJoinKeyword checks if the next token starts a join clause
func isJoinKeyword(tok token.TokenWithSpan) bool {
	if word, ok := tok.Token.(token.TokenWord); ok {
		kw := strings.ToUpper(string(word.Word.Keyword))
		return kw == "JOIN" || kw == "CROSS" || kw == "INNER" ||
			kw == "LEFT" || kw == "RIGHT" || kw == "FULL" ||
			kw == "NATURAL"
	}
	return false
}

// parseTableWithJoins parses a table reference with optional joins
func parseTableWithJoins(p *Parser) (query.TableWithJoins, error) {
	// Parse the base table
	relation, err := parseTableFactor(p)
	if err != nil {
		return query.TableWithJoins{}, err
	}

	result := query.TableWithJoins{
		Relation: relation,
	}

	// Parse any joins
	for isJoinKeyword(p.PeekToken()) {
		join, err := parseJoin(p)
		if err != nil {
			return query.TableWithJoins{}, err
		}
		result.Joins = append(result.Joins, join)
	}

	return result, nil
}

// parseJoin parses a JOIN clause
func parseJoin(p *Parser) (query.Join, error) {
	// Parse join type modifiers
	natural := p.ParseKeyword("NATURAL")

	// Determine join type string
	// Default is just "JOIN" (not "INNER JOIN" - we preserve the original syntax)
	joinTypeStr := "JOIN"
	if p.ParseKeyword("CROSS") {
		joinTypeStr = "CROSS JOIN"
	} else if p.ParseKeyword("INNER") {
		joinTypeStr = "INNER JOIN"
	} else if p.ParseKeyword("LEFT") {
		if p.ParseKeyword("OUTER") {
			joinTypeStr = "LEFT OUTER JOIN"
		} else if p.ParseKeyword("SEMI") {
			joinTypeStr = "LEFT SEMI JOIN"
		} else if p.ParseKeyword("ANTI") {
			joinTypeStr = "LEFT ANTI JOIN"
		} else {
			joinTypeStr = "LEFT JOIN"
		}
	} else if p.ParseKeyword("RIGHT") {
		if p.ParseKeyword("OUTER") {
			joinTypeStr = "RIGHT OUTER JOIN"
		} else if p.ParseKeyword("SEMI") {
			joinTypeStr = "RIGHT SEMI JOIN"
		} else if p.ParseKeyword("ANTI") {
			joinTypeStr = "RIGHT ANTI JOIN"
		} else {
			joinTypeStr = "RIGHT JOIN"
		}
	} else if p.ParseKeyword("FULL") {
		if p.ParseKeyword("OUTER") {
			joinTypeStr = "FULL OUTER JOIN"
		} else {
			joinTypeStr = "FULL JOIN"
		}
	} else if p.ParseKeyword("SEMI") {
		joinTypeStr = "SEMI JOIN"
	} else if p.ParseKeyword("ANTI") {
		joinTypeStr = "ANTI JOIN"
	}

	// Expect JOIN
	if !p.ParseKeyword("JOIN") {
		return query.Join{}, p.Expected("JOIN", p.PeekToken())
	}

	// Parse table
	table, err := parseTableFactor(p)
	if err != nil {
		return query.Join{}, err
	}

	// Parse join constraint (ON or USING)
	var constraint query.JoinConstraint
	if p.ParseKeyword("ON") {
		ep := NewExpressionParser(p)
		cond, err := ep.ParseExpr()
		if err != nil {
			return query.Join{}, err
		}
		constraint = &query.OnJoinConstraint{
			Expr: &queryExprWrapper{expr: cond},
		}
	} else if p.ParseKeyword("USING") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return query.Join{}, err
		}
		// Parse qualified column names like (col1, t2.col1, schema.table.col)
		// Reference: src/parser/mod.rs:16687 - parse_parenthesized_qualified_column_list
		objNames, err := parseCommaSeparatedObjectNames(p)
		if err != nil {
			return query.Join{}, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return query.Join{}, err
		}
		// Convert []*ast.ObjectName to []query.ObjectName
		attrs := make([]query.ObjectName, len(objNames))
		for i, objName := range objNames {
			queryParts := make([]query.Ident, len(objName.Parts))
			for j, part := range objName.Parts {
				if identPart, ok := part.(*ast.ObjectNamePartIdentifier); ok {
					queryParts[j] = query.Ident{Value: identPart.Ident.Value}
				}
			}
			attrs[i] = query.ObjectName{Parts: queryParts}
		}
		constraint = &query.UsingJoinConstraint{
			Attrs: attrs,
		}
	} else if natural {
		constraint = &query.NaturalJoinConstraint{}
	} else {
		constraint = &query.NoneJoinConstraint{}
	}

	// Handle NATURAL prefix in type
	if natural {
		joinTypeStr = "NATURAL " + joinTypeStr
	}

	return query.Join{
		Relation: table,
		JoinOperator: &query.StandardJoinOp{
			Type:       joinTypeStr,
			Constraint: constraint,
		},
	}, nil
}

// parseTableFactor parses a single table reference
// Reference: src/parser/mod.rs:15472 parse_table_factor
func parseTableFactor(p *Parser) (query.TableFactor, error) {
	// Check recursion limit
	if err := p.recursionCounter.TryDecrease(); err != nil {
		return nil, err
	}
	defer p.recursionCounter.Increase()

	// Check for TABLE(<expr>) syntax
	if p.ParseKeyword("TABLE") {
		return parseTableFunction(p)
	}

	// Check for LATERAL (must be followed by subquery or table function)
	if p.ParseKeyword("LATERAL") {
		return parseLateralTable(p)
	}

	// Check for parenthesized expression: could be subquery or nested join
	if isParenthesizedStart(p) {
		return parseParenthesizedTableFactor(p)
	}

	// Check for VALUES as table factor (Snowflake/Databricks)
	if p.GetDialect().SupportsValuesAsTableFactor() {
		if p.PeekKeyword("VALUES") {
			return parseValuesTableFactor(p)
		}
	}

	// Check for UNNEST (BigQuery/PostgreSQL)
	if p.GetDialect().SupportsUnnestTableFactor() {
		if p.ParseKeyword("UNNEST") {
			return parseUnnestTableFactor(p)
		}
	}

	// Check for Snowflake stage reference: @mystage or @namespace.stage/path
	if _, isAtSign := p.PeekToken().Token.(token.TokenAtSign); isAtSign {
		return parseSnowflakeStageTableFactor(p)
	}

	// Otherwise, it's a table name (possibly with PIVOT/UNPIVOT)
	return parseTableNameWithPivot(p)
}

// parseTableNameWithPivot parses a table name followed by optional PIVOT/UNPIVOT
// Reference: src/parser/mod.rs:15522-15531
func parseTableNameWithPivot(p *Parser) (query.TableFactor, error) {
	table, err := parseTableName(p)
	if err != nil {
		return nil, err
	}

	// Check for PIVOT/UNPIVOT operations after table name
	for {
		tok := p.PeekToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			kw := strings.ToUpper(string(word.Word.Keyword))
			if kw == "PIVOT" {
				p.AdvanceToken()
				table, err = parsePivotTableFactor(p, table)
				if err != nil {
					return nil, err
				}
				continue
			} else if kw == "UNPIVOT" {
				p.AdvanceToken()
				table, err = parseUnpivotTableFactor(p, table)
				if err != nil {
					return nil, err
				}
				continue
			}
		}
		break
	}

	return table, nil
}

// parseSnowflakeStageTableFactor parses a Snowflake stage reference as a table factor.
// Handles syntax like: @mystage, @namespace.stage, @~/path, @%table
// Reference: src/parser/mod.rs:15807-15836
func parseSnowflakeStageTableFactor(p *Parser) (query.TableFactor, error) {
	// Use the Snowflake dialect's stage name parsing
	if sfDialect, ok := p.dialect.(*snowflake.SnowflakeDialect); ok {
		stageName, err := sfDialect.ParseSnowflakeStageName(p)
		if err != nil {
			return nil, err
		}

		// Parse optional stage function args: (file_format => 'fmt', pattern => 'pat')
		var args *query.TableFunctionArgs
		if p.ConsumeToken(token.TokenLParen{}) {
			ep := NewExpressionParser(p)
			var funcArgs []query.FunctionArg
			for {
				if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
					p.AdvanceToken()
					break
				}

				arg, err := ep.ParseFunctionArg()
				if err != nil {
					return nil, err
				}
				funcArgs = append(funcArgs, convertFunctionArgToQuery(arg))

				if p.ConsumeToken(token.TokenComma{}) {
					continue
				}
				if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
					p.AdvanceToken()
					break
				}
			}
			if len(funcArgs) > 0 {
				args = &query.TableFunctionArgs{
					Args: funcArgs,
				}
			}
		}

		// Parse optional alias
		alias, _ := tryParseTableAlias(p)

		return &query.TableTableFactor{
			Name:       astObjectNameToQuery(stageName),
			Alias:      alias,
			Args:       args,
			WithHints:  nil,
			Version:    nil,
			Partitions: nil,
			Sample:     nil,
			IndexHints: nil,
		}, nil
	}

	return nil, fmt.Errorf("expected Snowflake dialect for stage reference")
}

// isParenthesizedStart checks if next token is a left paren
func isParenthesizedStart(p *Parser) bool {
	tok := p.PeekToken()
	_, ok := tok.Token.(token.TokenLParen)
	return ok
}

// parseParenthesizedTableFactor parses (subquery) or (nested_join)
// Reference: src/parser/mod.rs:15497-15609
func parseParenthesizedTableFactor(p *Parser) (query.TableFactor, error) {
	// Consume the opening paren
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// First, try to parse a derived table (subquery)
	// This handles cases like (SELECT ...), (WITH ... SELECT ...)
	if isSubqueryStartAfterParen(p) {
		return parseDerivedTableAfterParen(p)
	}

	// Not a subquery - parse as nested join
	// Inside the parentheses we expect a table factor followed by joins
	// Reference: src/parser/mod.rs:15541
	tableAndJoins, err := parseTableAndJoins(p)
	if err != nil {
		return nil, err
	}

	// Expect closing paren
	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Parse optional alias
	alias, _ := tryParseTableAlias(p)

	// Check if it's actually a nested join or just a parenthesized table
	if len(tableAndJoins.Joins) > 0 {
		// It's a nested join: (a JOIN b)
		return &query.NestedJoinTableFactor{
			TableWithJoins: tableAndJoins,
			Alias:          alias,
		}, nil
	}

	// No joins inside - check if it's a nested NestedJoin or dialect-specific parens
	if _, ok := tableAndJoins.Relation.(*query.NestedJoinTableFactor); ok {
		// Case (B): `(foo JOIN bar)` not followed by other joins, but wrapped in extra parens
		return &query.NestedJoinTableFactor{
			TableWithJoins: tableAndJoins,
			Alias:          alias,
		}, nil
	}

	// Dialect-specific: Snowflake allows parens around lone table names
	if p.GetDialect().SupportsParensAroundTableFactor() {
		// Apply outer alias to inner table if present
		if alias != nil {
			applyTableAlias(tableAndJoins.Relation, alias)
		}
		return tableAndJoins.Relation, nil
	}

	// Standard SQL: derived tables and bare tables cannot appear alone in parentheses
	// e.g., FROM (mytable) is not allowed without JOIN
	return &query.NestedJoinTableFactor{
		TableWithJoins: tableAndJoins,
		Alias:          alias,
	}, nil
}

// isSubqueryStartAfterParen checks if we're at a subquery after consuming '('
func isSubqueryStartAfterParen(p *Parser) bool {
	// Check if it's SELECT, WITH, or VALUES after the paren
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(token.TokenWord); ok {
		kw := strings.ToUpper(string(word.Word.Keyword))
		return kw == "SELECT" || kw == "WITH" || kw == "VALUES" || kw == "VALUE"
	}
	return false
}

// parseDerivedTableAfterParen parses (SELECT ...) or (WITH ... SELECT ...) when we've already consumed '('
// Reference: src/parser/mod.rs:15515-15533
func parseDerivedTableAfterParen(p *Parser) (query.TableFactor, error) {
	// Parse the subquery
	subquery, err := parseQuery(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Check for alias (required for subqueries)
	alias, err := parseTableAlias(p)
	if err != nil {
		return nil, err
	}

	// Parse optional SAMPLE clause after alias
	// Reference: src/parser/mod.rs:16514-16527
	var sample *query.TableSampleKind
	if parsedSample := maybeParseTableSample(p); parsedSample != nil {
		sample = &query.TableSampleKind{AfterTableAlias: parsedSample}
	}

	var table query.TableFactor = &query.DerivedTableFactor{
		Subquery: wrapQueryAsQuery(subquery),
		Alias:    alias,
		Sample:   sample,
	}

	// Check for PIVOT/UNPIVOT operations after derived table
	// Reference: src/parser/mod.rs:15522-15531
	for {
		tok := p.PeekToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			kw := strings.ToUpper(string(word.Word.Keyword))
			if kw == "PIVOT" {
				p.AdvanceToken()
				table, err = parsePivotTableFactor(p, table)
				if err != nil {
					return nil, err
				}
				continue
			} else if kw == "UNPIVOT" {
				p.AdvanceToken()
				table, err = parseUnpivotTableFactor(p, table)
				if err != nil {
					return nil, err
				}
				continue
			}
		}
		break
	}

	return table, nil
}

// parseTableAndJoins parses a table factor followed by optional joins
// Reference: src/parser/mod.rs:15278 parse_table_and_joins
func parseTableAndJoins(p *Parser) (*query.TableWithJoins, error) {
	relation, err := parseTableFactor(p)
	if err != nil {
		return nil, err
	}

	joins, err := parseJoins(p)
	if err != nil {
		return nil, err
	}

	return &query.TableWithJoins{
		Relation: relation,
		Joins:    joins,
	}, nil
}

// parsePivotTableFactor parses a PIVOT table factor (ClickHouse/Oracle style)
// Reference: src/parser/mod.rs:16590-16644
func parsePivotTableFactor(p *Parser, table query.TableFactor) (query.TableFactor, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse comma-separated aggregate functions
	aggregateFunctions, err := parseCommaSeparatedPivotAggregates(p)
	if err != nil {
		return nil, err
	}

	if err := p.ExpectKeywordIs("FOR"); err != nil {
		return nil, err
	}

	// Parse value column(s) - can be single expr or parenthesized list
	var valueColumns []query.Expr
	tok := p.PeekToken()
	if _, ok := tok.Token.(token.TokenLParen); ok {
		// Parenthesized list of value columns
		// Use precedence that stops before IN/BETWEEN to avoid consuming IN keyword
		// Reference: src/parser/mod.rs:16599-16603
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		dialect := p.GetDialect()
		betweenPrec := dialect.PrecValue(parseriface.PrecedenceBetween)
		valueColumns, err = parseCommaSeparatedQueryExprsWithPrecedence(p, betweenPrec)
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	} else {
		// Single value column
		// Use precedence that stops before IN/BETWEEN to avoid consuming IN keyword
		// Reference: src/parser/mod.rs:16600-16603
		ep := NewExpressionParser(p)
		dialect := p.GetDialect()
		betweenPrec := dialect.PrecValue(parseriface.PrecedenceBetween)
		expr, err := ep.ParseExprWithPrecedence(betweenPrec)
		if err != nil {
			return nil, err
		}
		valueColumns = []query.Expr{exprToQueryExpr(expr)}
	}

	if err := p.ExpectKeywordIs("IN"); err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse value source: ANY, subquery, or list of expressions
	valueSource, err := parsePivotValueSource(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Check for DEFAULT ON NULL clause
	var defaultOnNull query.Expr
	if p.ParseKeyword("DEFAULT") {
		if err := p.ExpectKeywordIs("ON"); err != nil {
			return nil, err
		}
		if err := p.ExpectKeywordIs("NULL"); err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		ep := NewExpressionParser(p)
		expr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		defaultOnNull = exprToQueryExpr(expr)
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Parse optional alias
	alias, _ := tryParseTableAlias(p)

	return &query.PivotTableFactor{
		Table:              table,
		AggregateFunctions: aggregateFunctions,
		ValueColumn:        valueColumns,
		ValueSource:        valueSource,
		DefaultOnNull:      defaultOnNull,
		Alias:              alias,
	}, nil
}

// parseUnpivotTableFactor parses an UNPIVOT table factor
// Reference: src/parser/mod.rs:16647-16678
func parseUnpivotTableFactor(p *Parser, table query.TableFactor) (query.TableFactor, error) {
	// Parse optional INCLUDE/EXCLUDE NULLS
	var nullInclusion *query.NullInclusion
	if p.ParseKeyword("INCLUDE") {
		if err := p.ExpectKeywordIs("NULLS"); err != nil {
			return nil, err
		}
		inc := query.IncludeNulls
		nullInclusion = &inc
	} else if p.ParseKeyword("EXCLUDE") {
		if err := p.ExpectKeywordIs("NULLS"); err != nil {
			return nil, err
		}
		exc := query.ExcludeNulls
		nullInclusion = &exc
	}

	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse value expression
	ep := NewExpressionParser(p)
	value, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if err := p.ExpectKeywordIs("FOR"); err != nil {
		return nil, err
	}

	// Parse name identifier
	name, err := p.ParseIdentifier()
	if err != nil {
		return nil, err
	}

	if err := p.ExpectKeywordIs("IN"); err != nil {
		return nil, err
	}

	// Parse parenthesized column list with optional aliases
	columns, err := parseParenthesizedExprWithAliasList(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Parse optional alias
	alias, _ := tryParseTableAlias(p)

	return &query.UnpivotTableFactor{
		Table:         table,
		Value:         exprToQueryExpr(value),
		Name:          query.Ident{Value: name.Value},
		Columns:       columns,
		NullInclusion: nullInclusion,
		Alias:         alias,
	}, nil
}

// parseCommaSeparatedPivotAggregates parses comma-separated aggregate functions with optional aliases
func parseCommaSeparatedPivotAggregates(p *Parser) ([]query.ExprWithAlias, error) {
	var aggregates []query.ExprWithAlias
	for {
		agg, err := parsePivotAggregateFunction(p)
		if err != nil {
			return nil, err
		}
		aggregates = append(aggregates, agg)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return aggregates, nil
}

// parsePivotAggregateFunction parses a single aggregate function with optional alias for PIVOT
// Reference: src/parser/mod.rs:16573-16587
func parsePivotAggregateFunction(p *Parser) (query.ExprWithAlias, error) {
	// Use ExpressionParser to parse the aggregate function call
	// The expression parser already knows how to handle function calls
	ep := NewExpressionParser(p)
	exprVal, err := ep.ParseExpr()
	if err != nil {
		return query.ExprWithAlias{}, err
	}

	// Parse optional alias
	// For PIVOT, the alias must not be "FOR" keyword
	var alias *query.Ident
	if p.ParseKeyword("AS") {
		ident, err := p.ParseIdentifier()
		if err == nil && strings.ToUpper(ident.Value) != "FOR" {
			alias = &query.Ident{Value: ident.Value}
		}
	} else {
		// Try implicit alias (not FOR keyword)
		tok := p.PeekToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			kw := strings.ToUpper(string(word.Word.Keyword))
			if kw != "FOR" && !isReservedForTableAlias(kw) {
				p.AdvanceToken()
				alias = &query.Ident{Value: word.Word.Value}
			}
		}
	}

	return query.ExprWithAlias{
		Expr:  exprToQueryExpr(exprVal),
		Alias: alias,
	}, nil
}

// parsePivotValueSource parses the IN clause value source for PIVOT
func parsePivotValueSource(p *Parser) (query.PivotValueSource, error) {
	// Check for ANY keyword (Snowflake)
	if p.ParseKeyword("ANY") {
		var orderBy []query.OrderByExpr
		if p.ParseKeyword("ORDER") {
			if err := p.ExpectKeywordIs("BY"); err != nil {
				return nil, err
			}
			// Parse comma-separated ORDER BY expressions
			for {
				ep := NewExpressionParser(p)
				expr, err := ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				orderBy = append(orderBy, query.OrderByExpr{
					Expr: exprToQueryExpr(expr),
				})
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
		}
		return &query.PivotValueAny{OrderBy: orderBy}, nil
	}

	// Check for subquery
	tok := p.PeekToken()
	if _, ok := tok.Token.(token.TokenLParen); ok {
		// Could be subquery or expression list - peek ahead
		nextTok := p.PeekNthToken(1)
		if word, ok := nextTok.Token.(token.TokenWord); ok {
			kw := strings.ToUpper(string(word.Word.Keyword))
			if kw == "SELECT" || kw == "WITH" {
				// It's a subquery
				subquery, err := parseQuery(p)
				if err != nil {
					return nil, err
				}
				return &query.PivotValueSubquery{Query: wrapQueryAsQuery(subquery)}, nil
			}
		}
	}

	// Parse as expression list with optional aliases
	values, err := parseCommaSeparatedExprWithAlias(p)
	if err != nil {
		return nil, err
	}

	return &query.PivotValueList{Values: values}, nil
}

// parseCommaSeparatedExprWithAlias parses comma-separated expressions with optional aliases
func parseCommaSeparatedExprWithAlias(p *Parser) ([]query.ExprWithAlias, error) {
	var values []query.ExprWithAlias
	for {
		ep := NewExpressionParser(p)
		expr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		// Check for optional alias
		var alias *query.Ident
		if p.ParseKeyword("AS") {
			ident, err := p.ParseIdentifier()
			if err == nil {
				alias = &query.Ident{Value: ident.Value}
			}
		} else {
			// Try implicit alias
			tok := p.PeekToken()
			if word, ok := tok.Token.(token.TokenWord); ok {
				if !isReservedForTableAlias(string(word.Word.Keyword)) {
					p.AdvanceToken()
					alias = &query.Ident{Value: word.Word.Value}
				}
			}
		}

		values = append(values, query.ExprWithAlias{
			Expr:  exprToQueryExpr(expr),
			Alias: alias,
		})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}
	return values, nil
}

// parseParenthesizedExprWithAliasList parses (expr [AS alias], ...)
func parseParenthesizedExprWithAliasList(p *Parser) ([]query.ExprWithAlias, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	values, err := parseCommaSeparatedExprWithAlias(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return values, nil
}

// parseJoins parses zero or more JOIN clauses
func parseJoins(p *Parser) ([]query.Join, error) {
	var joins []query.Join
	for isJoinKeyword(p.PeekToken()) {
		join, err := parseJoin(p)
		if err != nil {
			return nil, err
		}
		joins = append(joins, join)
	}
	return joins, nil
}

// applyTableAlias applies an alias to a table factor
func applyTableAlias(table query.TableFactor, alias *query.TableAlias) {
	// This function sets the alias on the table factor if it doesn't already have one
	switch t := table.(type) {
	case *query.TableTableFactor:
		if t.Alias == nil {
			t.Alias = alias
		}
	case *query.DerivedTableFactor:
		if t.Alias == nil {
			t.Alias = alias
		}
	case *query.TableFunctionTableFactor:
		if t.Alias == nil {
			t.Alias = alias
		}
	case *query.NestedJoinTableFactor:
		if t.Alias == nil {
			t.Alias = alias
		}
	}
}

// parseValuesTableFactor parses VALUES (...) as a table factor
// Reference: src/parser/mod.rs:15610-15646
func parseValuesTableFactor(p *Parser) (query.TableFactor, error) {
	// Accept either VALUES or VALUE keyword
	isValueKeyword := p.ParseKeyword("VALUE")
	if !isValueKeyword && !p.ParseKeyword("VALUES") {
		return nil, p.Expected("VALUES", p.PeekToken())
	}

	var rows [][]query.Expr
	var explicitRow bool

	for {
		// PostgreSQL/MySQL allow explicit ROW(...) syntax: VALUES ROW(1, 2, 'foo')
		if p.ParseKeyword("ROW") {
			explicitRow = true
		}

		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}

		// Check for empty row () - some dialects allow this
		if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
			p.AdvanceToken()
			rows = append(rows, []query.Expr{})
		} else {
			row, err := parseCommaSeparatedQueryExprs(p)
			if err != nil {
				return nil, err
			}

			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			rows = append(rows, row)
		}

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Create the VALUES expression
	values := &query.Values{
		ExplicitRow:  explicitRow,
		ValueKeyword: isValueKeyword,
		Rows:         rows,
	}

	valuesSetExpr := &query.ValuesSetExpr{
		Values: values,
	}

	q := &query.Query{
		Body: valuesSetExpr,
	}

	// Parse optional alias
	alias, _ := tryParseTableAlias(p)

	return &query.DerivedTableFactor{
		Subquery: q,
		Alias:    alias,
	}, nil
}

// parseUnnestTableFactor parses UNNEST(array_expr) [WITH OFFSET]
// Reference: src/parser/mod.rs:15647-15681
func parseUnnestTableFactor(p *Parser) (query.TableFactor, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	ep := NewExpressionParser(p)
	arrayExprs, err := parseCommaSeparatedExprs(ep)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// WITH ORDINALITY (PostgreSQL)
	withOrdinality := p.ParseKeywords([]string{"WITH", "ORDINALITY"})

	// Parse optional alias
	alias, _ := tryParseTableAlias(p)

	// WITH OFFSET (BigQuery)
	withOffset := p.ParseKeywords([]string{"WITH", "OFFSET"})
	var withOffsetAlias *query.Ident
	if withOffset {
		if p.ParseKeyword("AS") {
			ident, err := p.ParseIdentifier()
			if err == nil {
				withOffsetAlias = &query.Ident{Value: ident.Value}
			}
		}
	}

	// Convert []expr.Expr to []query.Expr
	var queryExprs []query.Expr
	for _, e := range arrayExprs {
		queryExprs = append(queryExprs, &queryExprWrapper{expr: e})
	}

	return &query.UnnestTableFactor{
		ArrayExprs:      queryExprs,
		WithOffset:      withOffset,
		WithOffsetAlias: withOffsetAlias,
		WithOrdinality:  withOrdinality,
		Alias:           alias,
	}, nil
}

// parseCommaSeparatedExprs parses a comma-separated list of expressions
func parseCommaSeparatedExprs(ep *ExpressionParser) ([]expr.Expr, error) {
	var exprs []expr.Expr

	for {
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, e)

		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return exprs, nil
}

// parseCommaSeparatedQueryExprs parses a comma-separated list of query expressions
func parseCommaSeparatedQueryExprs(p *Parser) ([]query.Expr, error) {
	var exprs []query.Expr

	ep := NewExpressionParser(p)
	for {
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, &queryExprWrapper{expr: e})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return exprs, nil
}

// parseCommaSeparatedQueryExprsWithPrecedence parses a comma-separated list of query expressions
// with a minimum precedence. Used for PIVOT value columns to avoid consuming IN keyword.
func parseCommaSeparatedQueryExprsWithPrecedence(p *Parser, precedence uint8) ([]query.Expr, error) {
	var exprs []query.Expr

	ep := NewExpressionParser(p)
	for {
		e, err := ep.ParseExprWithPrecedence(precedence)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, &queryExprWrapper{expr: e})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return exprs, nil
}

// parseGroupByExpressions parses GROUP BY expressions with support for empty tuple (PostgreSQL)
// Reference: src/parser/mod.rs:2591-2620
func parseGroupByExpressions(p *Parser) ([]query.Expr, error) {
	var exprs []query.Expr

	ep := NewExpressionParser(p)

	for {
		// Check for empty tuple () - PostgreSQL allows GROUP BY (), name
		tok := p.PeekTokenRef()
		if _, ok := tok.Token.(token.TokenLParen); ok {
			nextTok := p.PeekNthToken(1)
			if _, isRParen := nextTok.Token.(token.TokenRParen); isRParen {
				// Empty tuple
				p.AdvanceToken() // consume (
				p.AdvanceToken() // consume )
				exprs = append(exprs, &queryExprWrapper{expr: &expr.TupleExpr{
					SpanVal: tok.Span,
					Exprs:   []expr.Expr{},
				}})
				// Check for comma and continue
				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
				continue
			}
		}

		// Parse regular expression
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, &queryExprWrapper{expr: e})

		// Check for comma to continue
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return exprs, nil
}

// parseGroupByModifiers parses optional WITH ROLLUP, WITH CUBE, WITH TOTALS modifiers
// Reference: src/ast/query.rs:3666-3694
func parseGroupByModifiers(p *Parser) ([]query.GroupByWithModifier, error) {
	var modifiers []query.GroupByWithModifier

	for {
		if p.ParseKeyword("WITH") {
			if p.ParseKeyword("ROLLUP") {
				modifiers = append(modifiers, query.SimpleGroupByModifierRollup)
			} else if p.ParseKeyword("CUBE") {
				modifiers = append(modifiers, query.SimpleGroupByModifierCube)
			} else if p.ParseKeyword("TOTALS") {
				modifiers = append(modifiers, query.SimpleGroupByModifierTotals)
			} else {
				// WITH not followed by a valid modifier - this is an error
				return nil, p.Expected("ROLLUP, CUBE, or TOTALS after WITH", p.PeekToken())
			}
		} else {
			break
		}
	}

	return modifiers, nil
}

// parseGroupingSetsModifier parses optional GROUPING SETS modifier
// Reference: src/parser/mod.rs:12590-12603
func parseGroupingSetsModifier(p *Parser) query.GroupByWithModifier {
	if !p.ParseKeyword("GROUPING") {
		return nil
	}
	if !p.ParseKeyword("SETS") {
		// Backtrack - put back GROUPING if SETS doesn't follow
		return nil
	}

	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil
	}

	ep := NewExpressionParser(p)
	var sets [][]expr.Expr

	for {
		// Parse each set - either a tuple or a single expression
		tok := p.PeekTokenRef()
		if _, ok := tok.Token.(token.TokenLParen); ok {
			// Tuple: (a, b, c) or empty tuple ()
			p.AdvanceToken() // consume (
			// Check for empty tuple
			nextTok := p.PeekTokenRef()
			if _, isRParen := nextTok.Token.(token.TokenRParen); isRParen {
				p.AdvanceToken() // consume )
				// Empty tuple
				sets = append(sets, []expr.Expr{})
			} else {
				// Non-empty tuple - parse expressions
				var tupleExprs []expr.Expr
				for {
					e, err := ep.ParseExpr()
					if err != nil {
						return nil
					}
					tupleExprs = append(tupleExprs, e)
					if !p.ConsumeToken(token.TokenComma{}) {
						break
					}
				}
				if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
					return nil
				}
				sets = append(sets, tupleExprs)
			}
		} else {
			// Single expression: a - wrap in a single-element slice
			e, err := ep.ParseExpr()
			if err != nil {
				return nil
			}
			sets = append(sets, []expr.Expr{e})
		}

		// Check for comma to continue to next set
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil
	}

	// Create the GROUPING SETS expression
	groupingSetsExpr := &expr.GroupingSets{
		Sets:    sets,
		SpanVal: p.GetCurrentToken().Span,
	}

	return &query.GroupingSetsModifier{
		Expr: &queryExprWrapper{expr: groupingSetsExpr},
	}
}

// isSubqueryStart checks if next token starts a subquery
func isSubqueryStart(p *Parser) bool {
	tok := p.PeekToken()
	if _, ok := tok.Token.(token.TokenLParen); !ok {
		return false
	}

	// Check if it's SELECT or WITH after the paren
	nextTok := p.PeekNthToken(1)
	if word, ok := nextTok.Token.(token.TokenWord); ok {
		kw := strings.ToUpper(string(word.Word.Keyword))
		return kw == "SELECT" || kw == "WITH"
	}
	return false
}

// parseTableName parses a simple table name or table-valued function
// Reference: src/parser/mod.rs:15712-15805
func parseTableName(p *Parser) (query.TableFactor, error) {
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Check for table-valued function: fn() or schema.fn()
	// Reference: src/parser/mod.rs:15731-15735
	if p.ConsumeToken(token.TokenLParen{}) {
		// This is a table-valued function
		args, err := parseTableFunctionArgs(p)
		if err != nil {
			return nil, err
		}

		// Check for alias
		alias, _ := tryParseTableAlias(p)

		return &query.FunctionTableFactor{
			Lateral: false,
			Name:    astObjectNameToQuery(name),
			Args:    args,
			Alias:   alias,
		}, nil
	}

	// Parse potential version qualifier (AT, BEFORE, CHANGES for time travel)
	// Reference: src/parser/mod.rs:15728
	var version *query.TableVersionWithInfo
	if p.GetDialect().SupportsTableVersioning() {
		version = maybeParseTableVersion(p)
	}

	// Check for TABLESAMPLE before alias (some dialects)
	var sampleKind *query.TableSampleKind
	if p.GetDialect().SupportsTableSampleBeforeAlias() {
		if sample := maybeParseTableSample(p); sample != nil {
			sampleKind = &query.TableSampleKind{BeforeTableAlias: sample}
		}
	}

	// Check for alias
	alias, _ := tryParseTableAlias(p)

	// Check for TABLESAMPLE after alias (other dialects)
	if sampleKind == nil {
		if sample := maybeParseTableSample(p); sample != nil {
			sampleKind = &query.TableSampleKind{AfterTableAlias: sample}
		}
	}

	// Check for MYSQL-specific table hints: USE INDEX, IGNORE INDEX, FORCE INDEX
	var indexHints []query.TableIndexHints
	if p.GetDialect().SupportsIndexHints() {
		indexHints = parseTableIndexHints(p)
	}

	// Check for MSSQL-specific table hints: WITH (NOLOCK, etc.)
	// Reference: src/parser/mod.rs:15756-15766
	// Note: Don't check dialect here - WITH is already reserved for table aliases,
	// so if we see it here, it's either a table hint or needs to be put back for CTE parsing.
	var withHints []query.Expr
	if p.ParseKeyword("WITH") {
		if p.ConsumeToken(token.TokenLParen{}) {
			ep := NewExpressionParser(p)
			hints, err := parseCommaSeparatedExprs(ep)
			if err == nil {
				for _, h := range hints {
					withHints = append(withHints, &queryExprWrapper{expr: h})
				}
			}
			p.ExpectToken(token.TokenRParen{}) // consume ) even if expr parsing failed
		} else {
			// Not a table hint, put back WITH for CTE parsing
			p.PrevToken()
		}
	}

	return &query.TableTableFactor{
		Name:       astObjectNameToQuery(name),
		Alias:      alias,
		Sample:     sampleKind,
		IndexHints: indexHints,
		WithHints:  withHints,
		Version:    version,
	}, nil
}

// maybeParseTableVersion parses optional table version clauses like AT/BEFORE/CHANGES
// Reference: src/parser/mod.rs:16370-16415
func maybeParseTableVersion(p *Parser) *query.TableVersionWithInfo {
	if !p.GetDialect().SupportsTableVersioning() {
		return nil
	}

	ep := NewExpressionParser(p)

	// Check for AT(...) or BEFORE(...) - Snowflake time travel
	if p.PeekKeyword("AT") || p.PeekKeyword("BEFORE") {
		// Parse the function name (AT or BEFORE)
		funcName, err := p.ParseObjectName()
		if err != nil {
			return nil
		}

		// Expect opening parenthesis - if not present, this isn't time travel syntax
		tok := p.PeekTokenRef()
		if _, ok := tok.Token.(token.TokenLParen); !ok {
			// Not a function call, put back the object name tokens
			// TODO: Implement proper token backup
			return nil
		}
		p.AdvanceToken() // consume (

		// Collect the arguments using function arg parser (handles named arguments like TIMESTAMP => expr)
		var funcArgs []query.FunctionArg
		if _, ok := p.PeekTokenRef().Token.(token.TokenRParen); !ok {
			for {
				// Parse function argument (handles named arguments with => operator)
				arg, err := ep.ParseFunctionArg()
				if err != nil {
					return nil
				}
				funcArgs = append(funcArgs, convertFunctionArgToQuery(arg))

				if p.ConsumeToken(token.TokenComma{}) {
					continue
				}
				break
			}
		}

		// Expect closing parenthesis
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil
		}

		// Create a simple expression wrapper that stores the function-like call
		funcExpr := &timeTravelExpr{
			name: funcName.String(),
			args: funcArgs,
		}

		return &query.TableVersionWithInfo{
			Type: query.TableVersionFunction,
			Expr: funcExpr,
		}
	}

	// Check for CHANGES(...) - Snowflake change tracking
	if p.PeekKeyword("CHANGES") {
		// Parse CHANGES function-like call
		changesFunc, err := parseTimeTravelFunction(p, ep)
		if err != nil {
			return nil
		}

		// Parse AT function-like call
		atFunc, err := parseTimeTravelFunction(p, ep)
		if err != nil {
			return nil
		}

		// Optional END function-like call
		var endFunc query.Expr
		if p.PeekKeyword("END") {
			endFunc, err = parseTimeTravelFunction(p, ep)
			if err != nil {
				return nil
			}
		}

		return &query.TableVersionWithInfo{
			Type: query.TableVersionChanges,
			ChangesInfo: &query.TableVersionChangesInfo{
				Changes: changesFunc,
				At:      atFunc,
				End:     endFunc,
			},
		}
	}

	return nil
}

// timeTravelExpr represents a time travel function-like expression (AT, BEFORE, CHANGES)
type timeTravelExpr struct {
	name string
	args []query.FunctionArg
}

func (t *timeTravelExpr) String() string {
	var argStrs []string
	for _, a := range t.args {
		argStrs = append(argStrs, a.String())
	}
	return fmt.Sprintf("%s(%s)", t.name, strings.Join(argStrs, ", "))
}

// parseTimeTravelFunction parses a function-like call like AT(TIMESTAMP => expr) or CHANGES(INFO => DEFAULT)
func parseTimeTravelFunction(p *Parser, ep *ExpressionParser) (query.Expr, error) {
	funcName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Expect opening parenthesis
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse arguments using function arg parser (handles named arguments with => operator)
	var funcArgs []query.FunctionArg
	if _, ok := p.PeekTokenRef().Token.(token.TokenRParen); !ok {
		for {
			arg, err := ep.ParseFunctionArg()
			if err != nil {
				return nil, err
			}
			funcArgs = append(funcArgs, convertFunctionArgToQuery(arg))

			if p.ConsumeToken(token.TokenComma{}) {
				continue
			}
			break
		}
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &timeTravelExpr{
		name: funcName.String(),
		args: funcArgs,
	}, nil
}

// parseTableIndexHints parses MySQL-style index hints: USE INDEX, IGNORE INDEX, FORCE INDEX
// Reference: src/parser/mod.rs:12440-12497
func parseTableIndexHints(p *Parser) []query.TableIndexHints {
	var hints []query.TableIndexHints

	for {
		var hintType query.TableIndexHintType
		if p.ParseKeyword("USE") {
			hintType = query.TableIndexHintTypeUse
		} else if p.ParseKeyword("IGNORE") {
			hintType = query.TableIndexHintTypeIgnore
		} else if p.ParseKeyword("FORCE") {
			hintType = query.TableIndexHintTypeForce
		} else {
			break
		}

		// Expect INDEX or KEY
		var indexType query.TableIndexType
		if p.ParseKeyword("INDEX") {
			indexType = query.TableIndexTypeIndex
		} else if p.ParseKeyword("KEY") {
			indexType = query.TableIndexTypeKey
		} else {
			// Invalid syntax, but we'll just break
			break
		}

		// Optional FOR clause: FOR JOIN, FOR ORDER BY, FOR GROUP BY
		var forClause *query.TableIndexHintForClause
		if p.ParseKeyword("FOR") {
			var clause query.TableIndexHintForClause
			if p.ParseKeyword("JOIN") {
				clause = query.TableIndexHintForClauseJoin
			} else if p.ParseKeyword("ORDER") && p.ParseKeyword("BY") {
				clause = query.TableIndexHintForClauseOrderBy
			} else if p.ParseKeyword("GROUP") && p.ParseKeyword("BY") {
				clause = query.TableIndexHintForClauseGroupBy
			} else {
				// Invalid, but continue
				clause = query.TableIndexHintForClauseJoin
			}
			forClause = &clause
		}

		// Expect parenthesized index name list
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			break
		}

		var indexNames []query.Ident
		// Check for empty list: )
		if _, isRParen := p.PeekToken().Token.(token.TokenRParen); !isRParen {
			// Parse comma-separated identifier list
			for {
				ident, err := p.ParseIdentifier()
				if err != nil {
					break
				}
				indexNames = append(indexNames, query.Ident{Value: ident.Value})

				if !p.ConsumeToken(token.TokenComma{}) {
					break
				}
			}
		}

		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			break
		}

		hints = append(hints, query.TableIndexHints{
			HintType:   hintType,
			IndexType:  indexType,
			ForClause:  forClause,
			IndexNames: indexNames,
		})
	}

	return hints
}

// parseTableFunctionArgs parses the arguments for a table-valued function
// Handles both empty args () and non-empty args (arg1, arg2, ...)
// Supports named arguments like FLATTEN(input => expr, outer => true)
// Reference: src/parser/mod.rs:17878-17897 parse_table_function_args
func parseTableFunctionArgs(p *Parser) ([]query.FunctionArg, error) {
	var args []query.FunctionArg

	// Check for empty args: )
	if p.ConsumeToken(token.TokenRParen{}) {
		return args, nil
	}

	// Parse non-empty argument list using function arg parser
	// This supports named arguments (name => value) unlike regular expressions
	ep := NewExpressionParser(p)
	for {
		funcArg, err := ep.ParseFunctionArg()
		if err != nil {
			return nil, err
		}

		// Convert expr.FunctionArg to query.FunctionArg
		queryArg := convertFunctionArgToQuery(funcArg)
		args = append(args, queryArg)

		// Check for comma or closing paren
		if p.ConsumeToken(token.TokenComma{}) {
			continue
		}
		if p.ConsumeToken(token.TokenRParen{}) {
			break
		}
		return nil, fmt.Errorf("expected comma or closing parenthesis in function argument list")
	}

	return args, nil
}

// convertFunctionArgToQuery converts an expr.FunctionArg to a query.FunctionArg
// Handles both regular expression arguments and named arguments
func convertFunctionArgToQuery(arg expr.FunctionArg) query.FunctionArg {
	switch a := arg.(type) {
	case *expr.FunctionArgNamed:
		// Named argument: name => value, name = value, etc.
		// For now, we represent it as a query expression that stringifies correctly
		return query.FunctionArg{Expr: &queryExprWrapper{expr: a}}
	case *expr.FunctionArgExpr:
		// Regular expression argument
		return query.FunctionArg{Expr: &queryExprWrapper{expr: a.Expr}}
	default:
		// Fallback: wrap the argument as an expression
		return query.FunctionArg{Expr: &queryExprWrapper{expr: arg.(expr.Expr)}}
	}
}

// exprToQueryExpr converts an expr.Expr to a query.Expr
// This is a helper to bridge the two expression types
func exprToQueryExpr(e expr.Expr) query.Expr {
	return &queryExprWrapper{expr: e}
}

// parseDerivedTable parses a subquery in parentheses
func parseDerivedTable(p *Parser) (query.TableFactor, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Subquery
	subquery, err := parseQuery(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Check for alias (required for subqueries)
	alias, err := parseTableAlias(p)
	if err != nil {
		return nil, err
	}

	// Parse optional SAMPLE clause after alias
	var sample *query.TableSampleKind
	if parsedSample := maybeParseTableSample(p); parsedSample != nil {
		sample = &query.TableSampleKind{AfterTableAlias: parsedSample}
	}

	return &query.DerivedTableFactor{
		Subquery: wrapQueryAsQuery(subquery),
		Alias:    alias,
		Sample:   sample,
	}, nil
}

// wrapQueryAsQuery wraps a Statement as a *query.Query
func wrapQueryAsQuery(stmt ast.Statement) *query.Query {
	if stmt == nil {
		return nil
	}
	// The query types need to be properly connected
	// For now, return a placeholder
	return &query.Query{
		Body: stmt,
	}
}

// parseLateralTable parses LATERAL followed by either:
// - A subquery: LATERAL (SELECT ...)
// - A table function: LATERAL function_name(...)
// Reference: src/parser/mod.rs:15474-15489
func parseLateralTable(p *Parser) (query.TableFactor, error) {
	// LATERAL must be followed by either a subquery or a table function
	if p.ConsumeToken(token.TokenLParen{}) {
		// It's a subquery: LATERAL (SELECT ...)
		subquery, err := parseQuery(p)
		if err != nil {
			return nil, err
		}

		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}

		alias, err := parseTableAlias(p)
		if err != nil {
			return nil, err
		}

		// Parse optional SAMPLE clause after alias
		var sample *query.TableSampleKind
		if parsedSample := maybeParseTableSample(p); parsedSample != nil {
			sample = &query.TableSampleKind{AfterTableAlias: parsedSample}
		}

		return &query.DerivedTableFactor{
			Lateral:  true,
			Subquery: wrapQueryAsQuery(subquery),
			Alias:    alias,
			Sample:   sample,
		}, nil
	}

	// It's a table function: LATERAL function_name(...)
	// Parse the function name
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Expect opening paren for function args
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse function arguments
	args, err := parseTableFunctionArgs(p)
	if err != nil {
		return nil, err
	}

	// Parse optional alias
	alias, _ := tryParseTableAlias(p)

	return &query.FunctionTableFactor{
		Lateral: true,
		Name:    astObjectNameToQuery(name),
		Args:    args,
		Alias:   alias,
	}, nil
}

// parseTableFunction parses TABLE(<expr>) [AS <alias>]
// Reference: src/parser/mod.rs:15490-15496
func parseTableFunction(p *Parser) (query.TableFactor, error) {
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse the expression inside TABLE(...)
	ep := NewExpressionParser(p)
	exprVal, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Parse optional alias
	alias, _ := tryParseTableAlias(p)

	return &query.TableFunctionTableFactor{
		Expr:  &queryExprWrapper{expr: exprVal},
		Alias: alias,
	}, nil
}

// parseTableAlias parses a table alias (AS name or just name)
// Optionally followed by a column list: AS alias (col1, col2, ...)
func parseTableAlias(p *Parser) (*query.TableAlias, error) {
	explicit := false
	var name *query.Ident

	// Try to parse AS name or just name
	if p.ParseKeyword("AS") {
		explicit = true
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		name = &query.Ident{Value: ident.Value}
	} else {
		// Try implicit alias
		tok := p.PeekToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			if !isReservedForTableAlias(strings.ToUpper(string(word.Word.Keyword))) {
				p.AdvanceToken()
				name = &query.Ident{Value: word.Word.Value}
			}
		}
	}

	if name == nil {
		return nil, nil
	}

	// Check for optional column list: (col1, col2, ...)
	columns, err := parseTableAliasColumnDefs(p)
	if err != nil {
		return nil, err
	}

	return &query.TableAlias{
		Name:     *name,
		Explicit: explicit,
		Columns:  columns,
	}, nil
}

// parseTableAliasColumnDefs parses an optional (col1, col2, ...) after table alias
func parseTableAliasColumnDefs(p *Parser) ([]query.TableAliasColumnDef, error) {
	if !p.ConsumeToken(token.TokenLParen{}) {
		return nil, nil
	}

	var columns []query.TableAliasColumnDef
	for {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		columns = append(columns, query.TableAliasColumnDef{
			Name: query.Ident{Value: ident.Value},
		})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return columns, nil
}

// tryParseTableAlias attempts to parse a table alias, returning empty if none
func tryParseTableAlias(p *Parser) (*query.TableAlias, error) {
	return parseTableAlias(p)
}

// isReservedForTableAlias checks if a keyword cannot be used as a table alias
func isReservedForTableAlias(keyword string) bool {
	reserved := map[string]bool{
		"WHERE": true, "GROUP": true, "HAVING": true, "ORDER": true,
		"LIMIT": true, "UNION": true, "INTERSECT": true, "EXCEPT": true,
		"JOIN": true, "CROSS": true, "INNER": true, "LEFT": true,
		"RIGHT": true, "FULL": true, "ON": true, "USING": true,
		"SELECT": true, "INSERT": true, "UPDATE": true, "DELETE": true,
		"WINDOW": true, "QUALIFY": true, "SET": true,
		"PIVOT": true, "UNPIVOT": true, "MATCH_RECOGNIZE": true, "SEMANTIC_VIEW": true,
		"FOR": true, // FOR XML, FOR JSON, FOR BROWSE, lock clauses
		// MySQL index hints - these are not aliases
		"USE": true, "IGNORE": true, "FORCE": true,
		// Other clause-starting keywords
		"PARTITION": true, "TABLESAMPLE": true, "SAMPLE": true,
		// Reserved as both a table and a column alias (from Rust RESERVED_FOR_TABLE_ALIAS)
		"WITH": true, "EXPLAIN": true, "ANALYZE": true,
		"SORT": true, "LATERAL": true, "VIEW": true,
		"OFFSET": true, "FETCH": true, "MINUS": true,
		"NATURAL": true, "CLUSTER": true, "DISTRIBUTE": true, "GLOBAL": true, "ANTI": true,
		// Snowflake time travel keywords - these start time travel clauses, not aliases
		"AT": true, "BEFORE": true, "CHANGES": true,
	}
	return reserved[keyword]
}

// parseValues parses a VALUES expression and returns it as a *ValuesStatement
func parseValues(p *Parser) (*ValuesStatement, error) {
	// Accept either VALUES or VALUE keyword
	isValueKeyword := p.ParseKeyword("VALUE")
	if !isValueKeyword && !p.ParseKeyword("VALUES") {
		return nil, p.Expected("VALUES", p.PeekToken())
	}

	var rows [][]query.Expr
	var explicitRow bool

	for {
		// PostgreSQL/MySQL allow explicit ROW(...) syntax: VALUES ROW(1, 2, 'foo')
		if p.ParseKeyword("ROW") {
			explicitRow = true
		}

		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}

		// Check for empty row () - MySQL allows this
		// If the dialect supports empty projections and we see RParen immediately, it's an empty row
		if _, isRParen := p.PeekToken().Token.(token.TokenRParen); isRParen {
			// Empty row - consume the RParen
			p.AdvanceToken()
			rows = append(rows, []query.Expr{})
		} else {
			// Non-empty row - parse expressions
			row, err := parseCommaSeparatedQueryExprs(p)
			if err != nil {
				return nil, err
			}

			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			rows = append(rows, row)
		}

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Create the Values expression
	values := &query.Values{
		ExplicitRow:  explicitRow,
		ValueKeyword: isValueKeyword,
		Rows:         rows,
	}

	// Create a ValuesSetExpr as the query body
	valuesSetExpr := &query.ValuesSetExpr{
		Values: values,
	}

	// Create the Query with VALUES as the body
	q := &query.Query{
		Body: valuesSetExpr,
	}

	return &ValuesStatement{Query: q}, nil
}

// parseOrderByExpressions parses ORDER BY expressions
func parseOrderByExpressions(p *Parser) ([]query.OrderByExpr, error) {
	var exprs []query.OrderByExpr

	for {
		ep := NewExpressionParser(p)
		expr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		// Check for ASC/DESC
		var asc *bool
		if p.ParseKeyword("DESC") {
			b := false
			asc = &b
		} else if p.ParseKeyword("ASC") {
			b := true
			asc = &b
		}

		// Check for NULLS FIRST/LAST
		var nullsFirst *bool
		if p.ParseKeywords([]string{"NULLS", "FIRST"}) {
			b := true
			nullsFirst = &b
		} else if p.ParseKeywords([]string{"NULLS", "LAST"}) {
			b := false
			nullsFirst = &b
		}

		exprs = append(exprs, query.OrderByExpr{
			Expr: &queryExprWrapper{expr: expr},
			Options: query.OrderByOptions{
				Asc:        asc,
				NullsFirst: nullsFirst,
			},
		})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return exprs, nil
}

// parseCommaSeparatedQueryIdents parses a comma-separated list of query.Ident
func parseCommaSeparatedQueryIdents(p *Parser) ([]query.Ident, error) {
	var idents []query.Ident

	for {
		ident, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		idents = append(idents, query.Ident{Value: ident.Value})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return idents, nil
}

// parseNamedWindows parses a comma-separated list of named window definitions
// Used for: WINDOW window1 AS (...), window2 AS (...)
func parseNamedWindows(p *Parser) ([]query.NamedWindowDefinition, error) {
	var windows []query.NamedWindowDefinition

	for {
		// Parse window name
		name, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		// Expect AS
		if !p.ParseKeyword("AS") {
			return nil, p.Expected("AS after window name", p.PeekToken())
		}

		// Parse window expression
		var windowExpr query.NamedWindowExpr

		if p.ConsumeToken(token.TokenLParen{}) {
			// Window specification in parentheses
			spec, err := parseWindowSpec(p)
			if err != nil {
				return nil, err
			}
			windowExpr = &query.WindowSpecExpr{Spec: *spec}
		} else if p.GetDialect().SupportsWindowClauseNamedWindowReference() {
			// Named window reference (e.g., WINDOW w1 AS w2)
			refName, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			windowExpr = &query.NamedWindowReference{Name: query.Ident{Value: refName.Value}}
		} else {
			return nil, p.Expected("( or window name after AS", p.PeekToken())
		}

		windows = append(windows, query.NamedWindowDefinition{
			Name: query.Ident{Value: name.Value},
			Expr: windowExpr,
		})

		// Check for comma (more windows)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return windows, nil
}

// parseWindowSpec parses a window specification (inside parentheses)
func parseWindowSpec(p *Parser) (*query.WindowSpec, error) {
	spec := &query.WindowSpec{}

	// Check for optional window name reference at the beginning
	// e.g., (window_name ORDER BY ...)
	if !p.PeekKeyword("PARTITION") && !p.PeekKeyword("ORDER") &&
		!p.PeekKeyword("ROWS") && !p.PeekKeyword("RANGE") &&
		!p.PeekKeyword("GROUPS") && !p.PeekKeyword("UNBOUNDED") {
		// Might be a named window reference
		tok := p.PeekToken()
		if word, ok := tok.Token.(token.TokenWord); ok && word.Word.Keyword == token.NoKeyword {
			name, err := p.ParseIdentifier()
			if err == nil {
				spec.WindowName = &query.Ident{Value: name.Value}
			}
		}
	}

	// Parse PARTITION BY
	if p.ParseKeyword("PARTITION") {
		if !p.ParseKeyword("BY") {
			return nil, fmt.Errorf("expected BY after PARTITION")
		}
		partitionExprs, err := parseCommaSeparatedQueryExprs(p)
		if err != nil {
			return nil, err
		}
		for _, expr := range partitionExprs {
			spec.PartitionBy = append(spec.PartitionBy, expr)
		}
	}

	// Parse ORDER BY
	if p.ParseKeywords([]string{"ORDER", "BY"}) {
		orderByExprs, err := parseOrderByExpressions(p)
		if err != nil {
			return nil, err
		}
		spec.OrderBy = orderByExprs
	}

	// Parse window frame (ROWS, RANGE, GROUPS)
	if p.PeekKeyword("ROWS") || p.PeekKeyword("RANGE") || p.PeekKeyword("GROUPS") {
		frame, err := parseWindowFrame(p)
		if err != nil {
			return nil, err
		}
		spec.WindowFrame = frame
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return spec, nil
}

// parseWindowFrame parses a window frame specification
func parseWindowFrame(p *Parser) (*query.WindowFrame, error) {
	var units query.WindowFrameUnits

	if p.ParseKeyword("ROWS") {
		units = query.WindowFrameUnitsRows
	} else if p.ParseKeyword("RANGE") {
		units = query.WindowFrameUnitsRange
	} else if p.ParseKeyword("GROUPS") {
		units = query.WindowFrameUnitsGroups
	} else {
		return nil, fmt.Errorf("expected ROWS, RANGE, or GROUPS")
	}

	frame := &query.WindowFrame{Units: units}

	// Check for BETWEEN
	if p.ParseKeyword("BETWEEN") {
		startBound, err := parseWindowFrameBound(p)
		if err != nil {
			return nil, err
		}
		frame.Start = startBound

		if !p.ParseKeyword("AND") {
			return nil, fmt.Errorf("expected AND in window frame BETWEEN")
		}

		endBound, err := parseWindowFrameBound(p)
		if err != nil {
			return nil, err
		}
		frame.End = endBound
	} else {
		bound, err := parseWindowFrameBound(p)
		if err != nil {
			return nil, err
		}
		frame.Start = bound
	}

	return frame, nil
}

// parseWindowFrameBound parses a window frame bound
func parseWindowFrameBound(p *Parser) (query.WindowFrameBound, error) {
	if p.ParseKeyword("CURRENT") {
		if !p.ParseKeyword("ROW") {
			return nil, fmt.Errorf("expected ROW after CURRENT")
		}
		return &query.CurrentRowBound{}, nil
	}

	if p.ParseKeyword("UNBOUNDED") {
		if p.ParseKeyword("PRECEDING") {
			return &query.UnboundedPrecedingBound{}, nil
		}
		if p.ParseKeyword("FOLLOWING") {
			return &query.UnboundedFollowingBound{}, nil
		}
		return nil, fmt.Errorf("expected PRECEDING or FOLLOWING after UNBOUNDED")
	}

	// Expression bound (could be a number or INTERVAL expression)
	// Reference: src/parser/mod.rs:2575-2578
	// Check for INTERVAL expression starting with a string literal OR the INTERVAL keyword
	curTok := p.PeekToken()
	var exprBound expr.Expr
	var err error

	if strTok, isString := curTok.Token.(token.TokenSingleQuotedString); isString {
		// This is an INTERVAL expression like '1' DAY or '1 DAY' (string contains unit)
		// Parse the value first (already peeked as string)
		ep := NewExpressionParser(p)

		// Check if the string contains just the value or value+unit
		strVal := strTok.Value
		if strings.Contains(strVal, " ") {
			// The string contains the unit, e.g., '1 DAY' - this is the full interval
			// Just parse it as a literal and let the String() method handle it
			p.AdvanceToken() // consume the string
			exprBound = &expr.ValueExpr{Value: strVal}
		} else {
			// The string is just the value, e.g., '1' - need to parse the unit after
			p.AdvanceToken() // consume the string value

			// Parse the temporal unit (DAY, MONTH, etc.)
			if ep.isTemporalUnit() {
				unit := ep.parseTemporalUnit()
				// Create an interval expression with value and unit
				exprBound = &expr.IntervalExpr{
					Value:        &expr.ValueExpr{Value: strVal},
					LeadingField: &unit,
				}
			} else {
				// No unit found, just use the literal
				exprBound = &expr.ValueExpr{Value: strVal}
			}
		}
	} else if word, ok := curTok.Token.(token.TokenWord); ok && word.Word.Keyword == "INTERVAL" {
		// INTERVAL keyword present, parse the full interval expression
		ep := NewExpressionParser(p)
		exprBound, err = ep.parseIntervalExpr()
		if err != nil {
			return nil, err
		}
	} else {
		// Regular expression
		ep := NewExpressionParser(p)
		exprBound, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	if p.ParseKeyword("PRECEDING") {
		return &query.PrecedingBound{Expr: &queryExprWrapper{expr: exprBound}}, nil
	}

	if p.ParseKeyword("FOLLOWING") {
		return &query.FollowingBound{Expr: &queryExprWrapper{expr: exprBound}}, nil
	}

	return nil, fmt.Errorf("expected PRECEDING or FOLLOWING after expression")
}

func parseUpdateInQuery(p *Parser, with interface{}) (ast.Statement, error) {
	// UPDATE table SET assignments WHERE ...
	p.ParseKeyword("UPDATE")

	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectKeyword("SET"); err != nil {
		return nil, err
	}

	// Parse assignments
	var assignments []*expr.Assignment
	for {
		col, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}
		ep := NewExpressionParser(p)
		val, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		assignments = append(assignments, &expr.Assignment{
			Column: col,
			Value:  val,
		})
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Parse optional WHERE
	var selection expr.Expr
	if p.ParseKeyword("WHERE") {
		ep := NewExpressionParser(p)
		selection, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &statement.Update{
		Table:       tableName,
		Assignments: assignments,
		Selection:   selection,
	}, nil
}

func parseDeleteInQuery(p *Parser, with interface{}) (ast.Statement, error) {
	// DELETE FROM table WHERE ...
	p.ParseKeyword("DELETE")
	p.ParseKeyword("FROM")

	tableName, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Parse optional WHERE
	var selection expr.Expr
	if p.ParseKeyword("WHERE") {
		ep := NewExpressionParser(p)
		selection, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return &statement.Delete{
		Tables:    []*ast.ObjectName{tableName},
		Selection: selection,
	}, nil
}

func parseMergeInQuery(p *Parser, with interface{}) (ast.Statement, error) {
	// MERGE INTO target USING source ON condition WHEN ...
	p.ParseKeyword("MERGE")
	p.ParseKeyword("INTO")

	// Simplified: return placeholder
	return &statement.Merge{}, nil
}

func parseCTE(p *Parser) (query.CTE, error) {
	// Parse CTE name
	name, err := p.ParseIdentifier()
	if err != nil {
		return query.CTE{}, err
	}

	cte := query.CTE{}
	alias := query.TableAlias{
		Name: query.Ident{Value: name.Value},
	}

	// Check for optional column list: CTE_name (col1, col2, ...) AS ...
	if p.PeekToken().Token.Equals(token.TokenLParen{}) {
		columns, err := p.ParseParenthesizedColumnList()
		if err != nil {
			return query.CTE{}, err
		}
		// Convert []ast.Ident to []query.TableAliasColumnDef
		for _, col := range columns {
			alias.Columns = append(alias.Columns, query.TableAliasColumnDef{
				Name: query.Ident{Value: col.Value},
			})
		}
	}

	// Expect AS keyword
	if !p.ParseKeyword("AS") {
		return query.CTE{}, p.Expected("AS after CTE name", p.PeekToken())
	}

	// Check for MATERIALIZED / NOT MATERIALIZED (PostgreSQL)
	var materialized *query.CteAsMaterialized
	// Check if this is PostgreSQL dialect
	dialect := p.GetDialect()
	if _, isPostgres := dialect.(*postgresql.PostgreSqlDialect); isPostgres {
		if p.ParseKeyword("MATERIALIZED") {
			m := query.CteMaterializedYes
			materialized = &m
		} else if p.ParseKeywords([]string{"NOT", "MATERIALIZED"}) {
			m := query.CteMaterializedNo
			materialized = &m
		}
	}
	cte.Materialized = materialized

	// Expect opening parenthesis for subquery
	if _, ok := p.PeekToken().Token.(token.TokenLParen); !ok {
		return query.CTE{}, p.Expected("( before CTE subquery", p.PeekToken())
	}
	p.AdvanceToken() // consume (

	// Parse the inner query
	innerQuery, err := parseQuery(p)
	if err != nil {
		return query.CTE{}, err
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return query.CTE{}, err
	}

	// Check for optional FROM keyword (BigQuery/Snowflake extension)
	if p.ParseKeyword("FROM") {
		fromName, err := p.ParseIdentifier()
		if err != nil {
			return query.CTE{}, err
		}
		cte.From = &query.Ident{Value: fromName.Value}
	}

	// Create the Query wrapper for the inner statement
	if selStmt, ok := innerQuery.(*SelectStatement); ok {
		cte.Query = &query.Query{
			Body: &query.SelectSetExpr{
				Select: &selStmt.Select,
			},
		}
	} else if qStmt, ok := innerQuery.(*QueryStatement); ok {
		// Nested CTE (WITH clause inside CTE)
		cte.Query = qStmt.Query
	} else if sQuery, ok := innerQuery.(*statement.Query); ok {
		// Query with WITH clause
		cte.Query = sQuery.Query
	}
	cte.Alias = alias

	return cte, nil
}

func parseCTEList(p *Parser) ([]query.CTE, error) {
	var ctes []query.CTE

	for {
		cte, err := parseCTE(p)
		if err != nil {
			return nil, err
		}
		ctes = append(ctes, cte)

		// Check for comma (more CTEs)
		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return ctes, nil
}

func parseOptionalOrderBy(p *Parser) (interface{}, error) {
	return nil, nil
}

func parseOptionalLimitClause(p *Parser) (interface{}, error) {
	return nil, nil
}

func parseSettings(p *Parser) (interface{}, error) {
	return nil, nil
}

func parseOffsetClause(p *Parser) (interface{}, error) {
	return nil, nil
}

// parseForClause parses MSSQL FOR XML, FOR JSON, or FOR BROWSE clause
// Reference: src/parser/mod.rs:13962 parse_for_clause
func parseForClause(p *Parser) (*query.ForClause, error) {
	if !p.ParseKeyword("FOR") {
		return nil, nil
	}

	// Check for XML
	if p.ParseKeyword("XML") {
		return parseForXml(p)
	}

	// Check for JSON
	if p.ParseKeyword("JSON") {
		return parseForJson(p)
	}

	// Check for BROWSE
	if p.ParseKeyword("BROWSE") {
		return &query.ForClause{Type: &query.ForBrowseClause{}}, nil
	}

	// Not a recognized FOR clause - we consumed FOR but it wasn't valid
	// This is an error, but for now just return nil
	return nil, fmt.Errorf("expected XML, JSON, or BROWSE after FOR")
}

// parseForXml parses FOR XML clause
// Reference: src/parser/mod.rs:13975 parse_for_xml
func parseForXml(p *Parser) (*query.ForClause, error) {
	var forXml query.ForXml
	var elementName *string

	// Parse mode: RAW, AUTO, EXPLICIT, or PATH
	if p.ParseKeyword("RAW") {
		// Check for optional element name: RAW('name')
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.AdvanceToken() // consume (
			str, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			elementName = &str
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		}
		forXml = query.ForXmlRaw
	} else if p.ParseKeyword("AUTO") {
		forXml = query.ForXmlAuto
	} else if p.ParseKeyword("EXPLICIT") {
		forXml = query.ForXmlExplicit
	} else if p.ParseKeyword("PATH") {
		// Check for optional element name: PATH('name')
		if _, ok := p.PeekToken().Token.(token.TokenLParen); ok {
			p.AdvanceToken() // consume (
			str, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			elementName = &str
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		}
		forXml = query.ForXmlPath
	} else {
		return nil, fmt.Errorf("expected FOR XML [RAW | AUTO | EXPLICIT | PATH]")
	}

	// Parse optional options: ELEMENTS, BINARY BASE64, ROOT('...'), TYPE
	elements := false
	binaryBase64 := false
	var root *string
	typeFlag := false

	for p.ConsumeToken(token.TokenComma{}) {
		if p.ParseKeyword("ELEMENTS") {
			elements = true
		} else if p.ParseKeyword("BINARY") {
			if !p.ParseKeyword("BASE64") {
				return nil, fmt.Errorf("expected BASE64 after BINARY")
			}
			binaryBase64 = true
		} else if p.ParseKeyword("ROOT") {
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			str, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			root = &str
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else if p.ParseKeyword("TYPE") {
			typeFlag = true
		} else {
			// Unknown option, break
			break
		}
	}

	return &query.ForClause{
		Type: &query.ForXmlClause{
			ForXml:       forXml,
			ElementName:  elementName,
			Elements:     elements,
			BinaryBase64: binaryBase64,
			Root:         root,
			Type:         typeFlag,
		},
	}, nil
}

// parseForJson parses FOR JSON clause
// Reference: src/parser/mod.rs:14029 parse_for_json
func parseForJson(p *Parser) (*query.ForClause, error) {
	var forJson query.ForJson

	// Parse mode: AUTO or PATH
	if p.ParseKeyword("AUTO") {
		forJson = query.ForJsonAuto
	} else if p.ParseKeyword("PATH") {
		forJson = query.ForJsonPath
	} else {
		return nil, fmt.Errorf("expected FOR JSON [AUTO | PATH]")
	}

	// Parse optional options: ROOT('...'), INCLUDE_NULL_VALUES, WITHOUT_ARRAY_WRAPPER
	var root *string
	includeNullValues := false
	withoutArrayWrapper := false

	for p.ConsumeToken(token.TokenComma{}) {
		if p.ParseKeyword("ROOT") {
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			str, err := p.ParseStringLiteral()
			if err != nil {
				return nil, err
			}
			root = &str
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
		} else if p.ParseKeyword("INCLUDE_NULL_VALUES") {
			includeNullValues = true
		} else if p.ParseKeyword("WITHOUT_ARRAY_WRAPPER") {
			withoutArrayWrapper = true
		} else {
			// Unknown option, break
			break
		}
	}

	return &query.ForClause{
		Type: &query.ForJsonClause{
			ForJson:             forJson,
			Root:                root,
			IncludeNullValues:   includeNullValues,
			WithoutArrayWrapper: withoutArrayWrapper,
		},
	}, nil
}

func maybeParseOptimizerHints(p *Parser) ([]interface{}, error) {
	// Parse MySQL optimizer hints: /*+ hint1 hint2 */
	// This is simplified - full implementation would parse comment tokens
	return nil, nil
}

func parseOrderByExpr(p *Parser) (interface{}, error) {
	// Parse ORDER BY expression
	ep := NewExpressionParser(p)
	exprVal, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Parse optional ASC/DESC
	if p.ParseKeyword("DESC") || p.ParseKeyword("ASC") {
		// Skip for now
	}

	// Skip NULLS FIRST/LAST parsing for now

	return &query.OrderByExpr{
		Expr: exprVal,
	}, nil
}

func parseLimitClause(p *Parser) (interface{}, error) {
	// Parse LIMIT clause
	if !p.ParseKeyword("LIMIT") {
		return nil, nil
	}

	ep := NewExpressionParser(p)
	limitExpr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Parse optional OFFSET
	var offsetExpr expr.Expr
	if p.ParseKeyword("OFFSET") {
		offsetExpr, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
	}

	return map[string]expr.Expr{
		"limit":  limitExpr,
		"offset": offsetExpr,
	}, nil
}

func parseSetExpr(p *Parser) (interface{}, error) {
	// Parse set expression (UNION/INTERSECT/EXCEPT) - simplified
	if p.ParseKeyword("UNION") || p.ParseKeyword("INTERSECT") || p.ParseKeyword("EXCEPT") {
		// Skip full parsing for now
		return nil, nil
	}
	return nil, nil
}

func parseSubquery(p *Parser) (interface{}, error) {
	// Parse subquery: (SELECT ...)
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse the inner query
	innerQuery, err := parseQuery(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Return the query directly
	return innerQuery, nil
}

func parseOutputClause(p *Parser, tok token.TokenWithSpan) (interface{}, error) {
	return nil, nil
}

// parsePipeOperators parses BigQuery/DuckDB pipe operators (|> SELECT, |> WHERE, etc.)
// Reference: src/parser/mod.rs:13726 parse_pipe_operators
func parsePipeOperators(p *Parser) ([]query.PipeOperator, error) {
	var pipeOperators []query.PipeOperator

	for p.ConsumeToken(token.TokenVerticalBarRightAngleBracket{}) {
		// Parse the keyword after |>
		tok := p.PeekToken()
		word, ok := tok.Token.(token.TokenWord)
		if !ok {
			return nil, fmt.Errorf("expected keyword after |>, found %v", tok.Token)
		}

		kw := strings.ToUpper(string(word.Word.Keyword))
		p.AdvanceToken() // consume the keyword

		switch kw {
		case "SELECT":
			exprs, err := parseProjection(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeSelect{Exprs: exprs},
			})

		case "EXTEND":
			exprs, err := parseProjection(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeExtend{Exprs: exprs},
			})

		case "SET":
			assignments, err := parseQueryAssignments(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeSet{Assignments: assignments},
			})

		case "DROP":
			columns, err := parseQueryIdents(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeDrop{Columns: columns},
			})

		case "AS":
			alias, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeAs{Alias: query.Ident{Value: alias.Value}},
			})

		case "WHERE":
			ep := NewExpressionParser(p)
			expr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeWhere{Expr: &queryExprWrapper{expr: expr}},
			})

		case "LIMIT":
			ep := NewExpressionParser(p)
			expr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			var offset query.Expr
			if p.ParseKeyword("OFFSET") {
				offExpr, err := ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				offset = &queryExprWrapper{expr: offExpr}
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeLimit{
					Expr:   &queryExprWrapper{expr: expr},
					Offset: offset,
				},
			})

		case "ORDER":
			if !p.ParseKeyword("BY") {
				return nil, fmt.Errorf("expected BY after ORDER")
			}
			exprs, err := parseOrderByExpressions(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeOrderBy{Exprs: exprs},
			})

		case "AGGREGATE":
			// Parse AGGREGATE operator: |> AGGREGATE [exprs] [GROUP BY exprs]
			var fullTableExprs []query.ExprWithAliasAndOrderBy
			// Only parse expressions if not followed by GROUP BY
			if !p.PeekKeyword("GROUP") {
				exprs, err := parseCommaSeparatedExprWithAliasAndOrderBy(p)
				if err != nil {
					return nil, err
				}
				fullTableExprs = exprs
			}

			var groupByExpr []query.ExprWithAliasAndOrderBy
			if p.ParseKeyword("GROUP") {
				if !p.ParseKeyword("BY") {
					return nil, fmt.Errorf("expected BY after GROUP")
				}
				exprs, err := parseCommaSeparatedExprWithAliasAndOrderBy(p)
				if err != nil {
					return nil, err
				}
				groupByExpr = exprs
			}

			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeAggregate{
					FullTableExprs: fullTableExprs,
					GroupByExpr:    groupByExpr,
				},
			})

		case "RENAME":
			mappings, err := parseQueryIdentWithAliasList(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeRename{Mappings: mappings},
			})

		case "UNION":
			setQuantifier := parseSetQuantifier(p, "UNION")
			queries, err := parsePipeOperatorQueries(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeUnion{
					SetQuantifier: setQuantifier,
					Queries:       queries,
				},
			})

		case "INTERSECT":
			setQuantifier := parseDistinctRequiredSetQuantifier(p, "INTERSECT")
			queries, err := parsePipeOperatorQueries(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeIntersect{
					SetQuantifier: setQuantifier,
					Queries:       queries,
				},
			})

		case "EXCEPT":
			setQuantifier := parseDistinctRequiredSetQuantifier(p, "EXCEPT")
			queries, err := parsePipeOperatorQueries(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeExcept{
					SetQuantifier: setQuantifier,
					Queries:       queries,
				},
			})

		case "TABLESAMPLE":
			sample, err := parseTableSample(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeTableSample{Sample: sample},
			})

		case "CALL":
			// Parse function call: |> CALL function_name(args) [AS alias]
			funcName, err := p.ParseObjectName()
			if err != nil {
				return nil, err
			}
			// Convert ast.ObjectName to expr.ObjectName
			exprFuncName := &expr.ObjectName{
				Parts: make([]*expr.ObjectNamePart, len(funcName.Parts)),
			}
			for i, part := range funcName.Parts {
				if identPart, ok := part.(*ast.ObjectNamePartIdentifier); ok {
					exprFuncName.Parts[i] = &expr.ObjectNamePart{
						Ident: &expr.Ident{Value: identPart.Ident.Value},
					}
				}
			}
			// Parse the function
			ep := NewExpressionParser(p)
			funcExpr, err := ep.parseFunction(exprFuncName)
			if err != nil {
				return nil, err
			}
			if fn, ok := funcExpr.(*expr.FunctionExpr); ok {
				// Parse optional AS alias
				var alias *query.Ident
				if p.ParseKeyword("AS") {
					id, err := p.ParseIdentifier()
					if err != nil {
						return nil, err
					}
					alias = &query.Ident{Value: id.Value}
				}
				pipeOperators = append(pipeOperators, query.PipeOperator{
					Type: &query.PipeCall{
						Function: fn,
						Alias:    alias,
					},
				})
			} else {
				return nil, fmt.Errorf("expected function call after CALL")
			}

		case "PIVOT":
			// Parse |> PIVOT (aggregate FOR col IN (values)) [AS alias]
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}

			// Parse aggregate functions
			aggregateFuncs, err := parseCommaSeparatedPivotAggregates(p)
			if err != nil {
				return nil, err
			}

			if !p.ParseKeyword("FOR") {
				return nil, fmt.Errorf("expected FOR after aggregates in PIVOT")
			}

			// Parse value column (can be period-separated like t.col)
			valueColParts, err := parsePeriodSeparatedIdents(p)
			if err != nil {
				return nil, err
			}

			if !p.ParseKeyword("IN") {
				return nil, fmt.Errorf("expected IN after value column in PIVOT")
			}

			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}

			// Parse value source (ANY, subquery, or value list)
			var valueSource query.PivotValueSource
			if p.ParseKeyword("ANY") {
				// ANY with optional ORDER BY
				var orderBy []query.OrderByExpr
				if p.ParseKeyword("ORDER") && p.ParseKeyword("BY") {
					orderBy, err = parseOrderByExpressions(p)
					if err != nil {
						return nil, err
					}
				}
				valueSource = &query.PivotValueAny{OrderBy: orderBy}
			} else if isSubqueryStart(p) {
				q, err := p.parseQuery()
				if err != nil {
					return nil, err
				}
				queryVal := extractQueryFromStatement(q)
				if queryVal == nil {
					return nil, fmt.Errorf("expected query in PIVOT value source")
				}
				valueSource = &query.PivotValueSubquery{Query: queryVal}
			} else {
				// Value list
				values, err := parseCommaSeparatedExprWithAlias(p)
				if err != nil {
					return nil, err
				}
				valueSource = &query.PivotValueList{Values: values}
			}

			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			// Parse optional AS alias
			var alias *query.Ident
			if id, err := p.ParseIdentifier(); err == nil {
				alias = &query.Ident{Value: id.Value}
			}

			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipePivot{
					AggregateFunctions: aggregateFuncs,
					ValueColumn:        valueColParts,
					ValueSource:        valueSource,
					Alias:              alias,
				},
			})

		case "UNPIVOT":
			// Parse |> UNPIVOT (value_col FOR name_col IN (cols)) [AS alias]
			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}

			valueCol, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}

			if !p.ParseKeyword("FOR") {
				return nil, fmt.Errorf("expected FOR after value column in UNPIVOT")
			}

			nameCol, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}

			if !p.ParseKeyword("IN") {
				return nil, fmt.Errorf("expected IN after name column in UNPIVOT")
			}

			if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}

			unpivotCols, err := parseQueryIdents(p)
			if err != nil {
				return nil, err
			}

			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}

			// Parse optional AS alias
			var alias *query.Ident
			if id, err := p.ParseIdentifier(); err == nil {
				alias = &query.Ident{Value: id.Value}
			}

			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeUnpivot{
					ValueColumn:    query.Ident{Value: valueCol.Value},
					NameColumn:     query.Ident{Value: nameCol.Value},
					UnpivotColumns: unpivotCols,
					Alias:          alias,
				},
			})

		case "JOIN", "INNER", "LEFT", "RIGHT", "FULL", "CROSS":
			// Join pipe operators - simplified
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeJoin{},
			})

		default:
			return nil, fmt.Errorf("unexpected keyword after |>: %s", kw)
		}
	}

	return pipeOperators, nil
}

// parseCommaSeparatedExprWithAliasAndOrderBy parses a comma-separated list of expressions with optional aliases and order by
func parseCommaSeparatedExprWithAliasAndOrderBy(p *Parser) ([]query.ExprWithAliasAndOrderBy, error) {
	var exprs []query.ExprWithAliasAndOrderBy

	for {
		expr, err := parseExprWithAliasAndOrderBy(p)
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, *expr)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return exprs, nil
}

// parseExprWithAliasAndOrderBy parses a single expression with optional alias and order by
// Reference: src/parser/mod.rs parse_expr_with_alias_and_order_by
func parseExprWithAliasAndOrderBy(p *Parser) (*query.ExprWithAliasAndOrderBy, error) {
	ep := NewExpressionParser(p)
	expr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Check for optional alias
	var alias *query.Ident
	if p.ParseKeyword("AS") {
		id, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		alias = &query.Ident{Value: id.Value}
	} else {
		// Try implicit alias (identifier without AS)
		// But don't consume reserved keywords like GROUP, WHERE, etc.
		tok := p.PeekToken()
		if word, ok := tok.Token.(token.TokenWord); ok {
			kw := strings.ToUpper(string(word.Word.Keyword))
			// Don't treat clause keywords or ASC/DESC/NULLS as implicit aliases
			if !isClauseKeyword(kw) && !isReservedForTableAlias(kw) &&
				kw != "ASC" && kw != "DESC" && kw != "NULLS" {
				p.AdvanceToken()
				alias = &query.Ident{Value: word.Word.Value}
			}
		}
	}

	// Check for optional ASC/DESC
	var orderBy query.OrderByOptions
	if p.ParseKeyword("ASC") {
		asc := true
		orderBy.Asc = &asc
	} else if p.ParseKeyword("DESC") {
		asc := false
		orderBy.Asc = &asc
	}

	if p.ParseKeyword("NULLS") {
		if p.ParseKeyword("FIRST") {
			nullsFirst := true
			orderBy.NullsFirst = &nullsFirst
		} else if p.ParseKeyword("LAST") {
			nullsFirst := false
			orderBy.NullsFirst = &nullsFirst
		}
	}

	return &query.ExprWithAliasAndOrderBy{
		Expr: query.ExprWithAlias{
			Expr:  &queryExprWrapper{expr: expr},
			Alias: alias,
		},
		OrderBy: orderBy,
	}, nil
}

func parseQueryAssignments(p *Parser) ([]query.Assignment, error) {
	var assignments []query.Assignment

	for {
		col, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		if _, err := p.ExpectToken(token.TokenEq{}); err != nil {
			return nil, err
		}

		ep := NewExpressionParser(p)
		val, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		assignments = append(assignments, query.Assignment{
			Column: query.Ident{Value: col.Value},
			Value:  &queryExprWrapper{expr: val},
		})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return assignments, nil
}

// parseQueryIdents parses a comma-separated list of identifiers for pipe operators
func parseQueryIdents(p *Parser) ([]query.Ident, error) {
	var ids []query.Ident

	for {
		id, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		ids = append(ids, query.Ident{Value: id.Value})

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return ids, nil
}

// parseQueryIdentWithAliasList parses a comma-separated list of identifiers with optional aliases
// The AS keyword is optional: `id AS alias` or `id alias`
// Reference: src/parser/mod.rs:12370-12375
func parseQueryIdentWithAliasList(p *Parser) ([]query.IdentWithAlias, error) {
	var mappings []query.IdentWithAlias

	for {
		id, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		// AS keyword is optional
		p.ParseKeyword("AS")

		// Parse the alias identifier
		alias, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		mapping := query.IdentWithAlias{
			Ident: query.Ident{Value: id.Value},
			Alias: query.Ident{Value: alias.Value},
		}

		mappings = append(mappings, mapping)

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return mappings, nil
}

// parseSetQuantifier parses ALL/DISTINCT/BY NAME for set operations
// Reference: src/parser/mod.rs:14214-14241
func parseSetQuantifier(p *Parser, op string) query.SetQuantifier {
	// Check for DISTINCT BY NAME
	if p.ParseKeywords([]string{"DISTINCT", "BY", "NAME"}) {
		return query.SetQuantifierDistinctByName
	}
	// Check for BY NAME
	if p.ParseKeywords([]string{"BY", "NAME"}) {
		return query.SetQuantifierByName
	}
	// Check for ALL [BY NAME]
	if p.ParseKeyword("ALL") {
		if p.ParseKeywords([]string{"BY", "NAME"}) {
			return query.SetQuantifierAllByName
		}
		return query.SetQuantifierAll
	}
	if p.ParseKeyword("DISTINCT") {
		return query.SetQuantifierDistinct
	}
	return query.SetQuantifierNone
}

// parseDistinctRequiredSetQuantifier parses DISTINCT/ALL/BY NAME for INTERSECT/EXCEPT
// Reference: src/parser/mod.rs:14214-14241
func parseDistinctRequiredSetQuantifier(p *Parser, op string) query.SetQuantifier {
	// Check for DISTINCT BY NAME
	if p.ParseKeywords([]string{"DISTINCT", "BY", "NAME"}) {
		return query.SetQuantifierDistinctByName
	}
	// Check for BY NAME (implies DISTINCT)
	if p.ParseKeywords([]string{"BY", "NAME"}) {
		return query.SetQuantifierByName
	}
	if p.ParseKeyword("DISTINCT") {
		return query.SetQuantifierDistinct
	}
	if p.ParseKeyword("ALL") {
		if p.ParseKeywords([]string{"BY", "NAME"}) {
			return query.SetQuantifierAllByName
		}
		return query.SetQuantifierAll
	}
	return query.SetQuantifierDistinct // Default to DISTINCT for INTERSECT/EXCEPT
}

// parsePipeOperatorQueries parses a comma-separated list of subqueries for UNION/INTERSECT/EXCEPT
func parsePipeOperatorQueries(p *Parser) ([]*query.Query, error) {
	var queries []*query.Query

	for {
		// Expect parenthesized query
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}

		// Parse the inner query
		stmt, err := p.parseQuery()
		if err != nil {
			return nil, err
		}

		// Extract query from statement
		q := extractQueryFromStatement(stmt)
		if q == nil {
			return nil, fmt.Errorf("expected query in pipe operator")
		}

		queries = append(queries, q)

		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return queries, nil
}

// parsePeriodSeparatedIdents parses period-separated identifiers (like t.col)
func parsePeriodSeparatedIdents(p *Parser) ([]query.Ident, error) {
	var parts []query.Ident

	for {
		id, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		parts = append(parts, query.Ident{Value: id.Value})

		if !p.ConsumeToken(token.TokenPeriod{}) {
			break
		}
	}

	return parts, nil
}

// parseTableSample parses a table sample clause for pipe operators
// Reference: src/parser/mod.rs parse_table_sample
func parseTableSample(p *Parser) (*query.TableSample, error) {
	// Parse sample type: SYSTEM or BERNOULLI
	sampleMethod := query.TableSampleMethodSystem
	if p.ParseKeyword("BERNOULLI") {
		sampleMethod = query.TableSampleMethodBernoulli
	} else if p.ParseKeyword("SYSTEM") {
		sampleMethod = query.TableSampleMethodSystem
	}

	// Parse sample size: (expr PERCENT) or (expr ROWS)
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	ep := NewExpressionParser(p)
	sizeExpr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	var unitPtr *query.TableSampleUnit
	if p.ParseKeyword("PERCENT") {
		unit := query.TableSampleUnitPercent
		unitPtr = &unit
	} else if p.ParseKeyword("ROWS") {
		unit := query.TableSampleUnitRows
		unitPtr = &unit
	}

	if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Parse optional REPEATABLE/SEED clause
	var seed *query.TableSampleSeed
	if p.ParseKeyword("REPEATABLE") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		seedVal, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		seed = &query.TableSampleSeed{
			Modifier: query.TableSampleSeedModifierRepeatable,
			Value:    query.ValueWithSpan{Value: seedVal.String()},
		}
	} else if p.ParseKeyword("SEED") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		seedVal, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		seed = &query.TableSampleSeed{
			Modifier: query.TableSampleSeedModifierSeed,
			Value:    query.ValueWithSpan{Value: seedVal.String()},
		}
	}

	sampleMethodPtr := &sampleMethod

	return &query.TableSample{
		Modifier: query.TableSampleModifierTableSample,
		Name:     sampleMethodPtr,
		Quantity: &query.TableSampleQuantity{
			Parenthesized: true, // Always parenthesized in pipe syntax
			Value:         &queryExprWrapper{expr: sizeExpr},
			Unit:          unitPtr,
		},
		Seed: seed,
	}, nil
}

// maybeParseTableSample attempts to parse a TABLESAMPLE or SAMPLE clause
// Returns nil if no TABLESAMPLE keyword is found
// Reference: src/parser/mod.rs:15838
func maybeParseTableSample(p *Parser) *query.TableSample {
	// Check for TABLESAMPLE or SAMPLE keyword
	var modifier query.TableSampleModifier
	if p.ParseKeyword("TABLESAMPLE") {
		modifier = query.TableSampleModifierTableSample
	} else if p.ParseKeyword("SAMPLE") {
		modifier = query.TableSampleModifierSample
	} else {
		return nil
	}

	// Parse sampling method: BERNOULLI, ROW, SYSTEM, BLOCK
	var method *query.TableSampleMethod
	if p.ParseKeyword("BERNOULLI") {
		m := query.TableSampleMethodBernoulli
		method = &m
	} else if p.ParseKeyword("ROW") {
		m := query.TableSampleMethodRow
		method = &m
	} else if p.ParseKeyword("SYSTEM") {
		m := query.TableSampleMethodSystem
		method = &m
	} else if p.ParseKeyword("BLOCK") {
		m := query.TableSampleMethodBlock
		method = &m
	}

	// Check if parenthesized
	parenthesized := p.ConsumeToken(token.TokenLParen{})

	var quantity *query.TableSampleQuantity
	var bucket *query.TableSampleBucket

	if parenthesized && p.ParseKeyword("BUCKET") {
		// Parse BUCKET syntax: BUCKET n OUT OF m [ON expr]
		ep := NewExpressionParser(p)
		selectedBucket, _ := ep.ParseExpr()
		if p.ParseKeywords([]string{"OUT", "OF"}) {
			total, _ := ep.ParseExpr()
			var on query.Expr
			if p.ParseKeyword("ON") {
				onVal, _ := ep.ParseExpr()
				on = exprToQueryExpr(onVal)
			}
			bucket = &query.TableSampleBucket{
				Bucket: query.ValueWithSpan{Value: selectedBucket.String()},
				Total:  query.ValueWithSpan{Value: total.String()},
				On:     on,
			}
		}
	} else {
		// Parse quantity expression
		ep := NewExpressionParser(p)
		value, err := ep.ParseExpr()
		if err != nil {
			// Try to parse as placeholder/word if expression fails
			tok := p.NextToken()
			if word, ok := tok.Token.(token.TokenWord); ok {
				value = &expr.ValueExpr{Value: word.Word.Value}
			}
		}

		// Parse optional unit: ROWS or PERCENT
		var unit *query.TableSampleUnit
		if p.ParseKeyword("ROWS") {
			u := query.TableSampleUnitRows
			unit = &u
		} else if p.ParseKeyword("PERCENT") {
			u := query.TableSampleUnitPercent
			unit = &u
		}

		quantity = &query.TableSampleQuantity{
			Parenthesized: parenthesized,
			Value:         value,
			Unit:          unit,
		}
	}

	if parenthesized {
		p.ExpectToken(token.TokenRParen{})
	}

	// Parse optional SEED or REPEATABLE
	var seed *query.TableSampleSeed
	if p.ParseKeyword("REPEATABLE") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err == nil {
			ep := NewExpressionParser(p)
			seedVal, _ := ep.ParseExpr()
			p.ExpectToken(token.TokenRParen{})
			seed = &query.TableSampleSeed{
				Modifier: query.TableSampleSeedModifierRepeatable,
				Value:    query.ValueWithSpan{Value: seedVal.String()},
			}
		}
	} else if p.ParseKeyword("SEED") {
		if _, err := p.ExpectToken(token.TokenLParen{}); err == nil {
			ep := NewExpressionParser(p)
			seedVal, _ := ep.ParseExpr()
			p.ExpectToken(token.TokenRParen{})
			seed = &query.TableSampleSeed{
				Modifier: query.TableSampleSeedModifierSeed,
				Value:    query.ValueWithSpan{Value: seedVal.String()},
			}
		}
	}

	return &query.TableSample{
		Modifier: modifier,
		Name:     method,
		Quantity: quantity,
		Bucket:   bucket,
		Seed:     seed,
	}
}

// parseTop parses a TOP clause (MSSQL equivalent of LIMIT)
// Reference: src/parser/mod.rs:parse_top
func parseTop(p *Parser) *query.Top {
	var quantity *query.TopQuantity

	// Check for parenthesized expression
	if p.ConsumeToken(token.TokenLParen{}) {
		ep := NewExpressionParser(p)
		expr, _ := ep.ParseExpr()
		p.ExpectToken(token.TokenRParen{})
		quantity = &query.TopQuantity{Expr: &queryExprWrapper{expr: expr}}
	} else {
		// Parse as constant number
		tok := p.NextToken()
		if numTok, ok := tok.Token.(token.TokenNumber); ok {
			val, _ := strconv.ParseUint(numTok.Value, 10, 64)
			quantity = &query.TopQuantity{Constant: &val}
		}
	}

	// Parse optional PERCENT
	percent := p.ParseKeyword("PERCENT")

	// Parse optional WITH TIES
	withTies := p.ParseKeywords([]string{"WITH", "TIES"})

	return &query.Top{
		Quantity: quantity,
		Percent:  percent,
		WithTies: withTies,
	}
}
