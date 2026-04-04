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
	"github.com/user/sqlparser/dialects/postgresql"
	"github.com/user/sqlparser/span"
	"github.com/user/sqlparser/token"
	"github.com/user/sqlparser/tokenizer"
)

// SelectStatement wraps query.Select to implement ast.Statement
type SelectStatement struct {
	ast.BaseStatement
	query.Select
	LimitClause   query.LimitClause
	PipeOperators []query.PipeOperator // Pipe operators for |> syntax
}

// Span returns the source span
func (s *SelectStatement) Span() span.Span {
	return s.Select.Span()
}

// String returns the SQL representation including LIMIT clause
func (s *SelectStatement) String() string {
	str := s.Select.String()
	if s.LimitClause != nil {
		str = str + " " + s.LimitClause.String()
	}
	return str
}

// ValuesStatement wraps query.Query (for VALUES) to implement ast.Statement
type ValuesStatement struct {
	ast.BaseStatement
	Query *query.Query
}

// Span returns the source span
func (v *ValuesStatement) Span() span.Span {
	if v.Query != nil {
		return v.Query.Span()
	}
	return span.Span{}
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
func (q *QueryStatement) Span() span.Span {
	if q.Query != nil {
		return q.Query.Span()
	}
	return span.Span{}
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
		return nil, p.ExpectedRef("SELECT or VALUES after WITH", p.PeekTokenRef())
	}

	if err != nil {
		return nil, err
	}

	// Check if we need to create a Query wrapper (for WITH clause or pipe operators)
	needsQueryWrapper := withClause != nil

	// Get pipe operators from SelectStatement if present
	var pipeOperators []query.PipeOperator
	if selStmt, ok := body.(*SelectStatement); ok {
		pipeOperators = selStmt.PipeOperators
		if len(pipeOperators) > 0 {
			needsQueryWrapper = true
		}
	}

	// Create a Query statement if needed
	if needsQueryWrapper {
		if selStmt, ok := body.(*SelectStatement); ok {
			return &QueryStatement{
				Query: &query.Query{
					With: withClause,
					Body: &query.SelectSetExpr{
						Select: &selStmt.Select,
					},
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
	if p.ConsumeToken(tokenizer.TokenLParen{}) {
		// Parse the inner query
		innerQuery, err := parseQuery(p)
		if err != nil {
			return nil, err
		}
		// Expect closing parenthesis
		if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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

// parseSelect parses a SELECT statement
func parseSelect(p *Parser) (ast.Statement, error) {
	// Expect SELECT keyword
	_, err := p.ExpectKeyword("SELECT")
	if err != nil {
		return nil, err
	}

	// Parse DISTINCT / ALL (optional)
	p.ParseKeyword("DISTINCT")
	p.ParseKeyword("ALL")

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
		groupExprs, err := parseCommaSeparatedQueryExprs(p)
		if err != nil {
			return nil, err
		}
		groupBy = &query.GroupByExpressions{
			Expressions: groupExprs,
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
		if p.ConsumeToken(tokenizer.TokenComma{}) {
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

	// Parse pipe operators (BigQuery/DuckDB |> syntax)
	pipeOperators, err := parsePipeOperators(p)
	if err != nil {
		return nil, err
	}

	return &SelectStatement{
		Select: query.Select{
			Projection:          projection,
			From:                from,
			Selection:           selection,
			GroupBy:             groupBy,
			Having:              having,
			NamedWindow:         namedWindow,
			Qualify:             qualify,
			WindowBeforeQualify: windowBeforeQualify,
			Flavor:              query.SelectFlavorStandard,
		},
		LimitClause:   limitClause,
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}

		// Check if next token is a reserved keyword (trailing comma check)
		// This prevents eating a comma before a clause keyword like FROM, WHERE, etc.
		nextTok := p.PeekToken()
		if isWordToken(nextTok.Token) {
			word := getWordValue(nextTok.Token)
			if word != "" && isReservedForColumnAlias(word) {
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
	if p.ConsumeToken(tokenizer.TokenMul{}) {
		return &query.Wildcard{}, nil
	}

	// Check for qualified wildcard (table.*)
	// Look ahead: identifier followed by . *
	if isQualifiedWildcard(p) {
		// Parse table name
		tableName, err := p.ParseObjectName()
		if err != nil {
			return nil, err
		}
		p.ConsumeToken(tokenizer.TokenPeriod{})
		p.ConsumeToken(tokenizer.TokenMul{})

		// Convert ast.ObjectName to query.ObjectName
		queryName := astObjectNameToQuery(tableName)
		return &query.QualifiedWildcard{
			Kind: &query.ObjectNameWildcard{Name: queryName},
		}, nil
	}

	// Parse expression
	ep := NewExpressionParser(p)
	expr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Check for alias (AS name or just name)
	if p.ParseKeyword("AS") {
		aliasIdent, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}
		return &query.AliasedExpr{
			Expr:  &queryExprWrapper{expr: expr},
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
					Expr:  &queryExprWrapper{expr: expr},
					Alias: astIdentToQuery(aliasIdent),
				}, nil
			}
		}
	}

	return &query.UnnamedExpr{Expr: &queryExprWrapper{expr: expr}}, nil
}

// isQualifiedWildcard checks if next tokens form table.* pattern
func isQualifiedWildcard(p *Parser) bool {
	// Need at least 3 tokens ahead
	tok0 := p.PeekToken()
	if !isWordToken(tok0.Token) {
		return false
	}

	tok1 := p.PeekNthToken(1)
	if _, ok := tok1.Token.(tokenizer.TokenPeriod); !ok {
		return false
	}

	tok2 := p.PeekNthToken(2)
	if _, ok := tok2.Token.(tokenizer.TokenMul); !ok {
		return false
	}

	return true
}

// isWordToken checks if token is a word/identifier
func isWordToken(tok tokenizer.Token) bool {
	_, ok := tok.(tokenizer.TokenWord)
	return ok
}

// getWordValue extracts the keyword value from a word token
func getWordValue(tok tokenizer.Token) string {
	if word, ok := tok.(tokenizer.TokenWord); ok {
		return word.Word.Value
	}
	return ""
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
		if p.ConsumeToken(tokenizer.TokenComma{}) {
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
func isJoinKeyword(tok tokenizer.TokenWithSpan) bool {
	if word, ok := tok.Token.(tokenizer.TokenWord); ok {
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
		if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
			return query.Join{}, err
		}
		cols, err := parseCommaSeparatedQueryIdents(p)
		if err != nil {
			return query.Join{}, err
		}
		if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
			return query.Join{}, err
		}
		// Convert []query.Ident to []query.ObjectName for UsingJoinConstraint
		attrs := make([]query.ObjectName, len(cols))
		for i, col := range cols {
			attrs[i] = query.ObjectName{Parts: []query.Ident{col}}
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

	// Otherwise, it's a table name
	return parseTableName(p)
}

// isParenthesizedStart checks if next token is a left paren
func isParenthesizedStart(p *Parser) bool {
	tok := p.PeekToken()
	_, ok := tok.Token.(tokenizer.TokenLParen)
	return ok
}

// parseParenthesizedTableFactor parses (subquery) or (nested_join)
// Reference: src/parser/mod.rs:15497-15609
func parseParenthesizedTableFactor(p *Parser) (query.TableFactor, error) {
	// Consume the opening paren
	if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
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
	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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
	// Check if it's SELECT or WITH after the paren
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(tokenizer.TokenWord); ok {
		kw := strings.ToUpper(string(word.Word.Keyword))
		return kw == "SELECT" || kw == "WITH"
	}
	return false
}

// parseDerivedTableAfterParen parses (SELECT ...) or (WITH ... SELECT ...) when we've already consumed '('
func parseDerivedTableAfterParen(p *Parser) (query.TableFactor, error) {
	// Parse the subquery
	subquery, err := parseQuery(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return nil, err
	}

	// Check for alias (required for subqueries)
	alias, err := parseTableAlias(p)
	if err != nil {
		return nil, err
	}

	return &query.DerivedTableFactor{
		Subquery: wrapQueryAsQuery(subquery),
		Alias:    alias,
	}, nil
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

	for {
		if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
			return nil, err
		}

		// Check for empty row () - some dialects allow this
		if _, isRParen := p.PeekToken().Token.(tokenizer.TokenRParen); isRParen {
			p.AdvanceToken()
			rows = append(rows, []query.Expr{})
		} else {
			row, err := parseCommaSeparatedQueryExprs(p)
			if err != nil {
				return nil, err
			}

			if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
				return nil, err
			}

			rows = append(rows, row)
		}

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	// Create the VALUES expression
	values := &query.Values{
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
	if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	ep := NewExpressionParser(p)
	arrayExprs, err := parseCommaSeparatedExprs(ep)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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

		if !ep.parser.ConsumeToken(tokenizer.TokenComma{}) {
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return exprs, nil
}

// isSubqueryStart checks if next token starts a subquery
func isSubqueryStart(p *Parser) bool {
	tok := p.PeekToken()
	if _, ok := tok.Token.(tokenizer.TokenLParen); !ok {
		return false
	}

	// Check if it's SELECT or WITH after the paren
	nextTok := p.PeekNthToken(1)
	if word, ok := nextTok.Token.(tokenizer.TokenWord); ok {
		kw := strings.ToUpper(string(word.Word.Keyword))
		return kw == "SELECT" || kw == "WITH"
	}
	return false
}

// parseTableName parses a simple table name
func parseTableName(p *Parser) (query.TableFactor, error) {
	name, err := p.ParseObjectName()
	if err != nil {
		return nil, err
	}

	// Check for alias
	alias, _ := tryParseTableAlias(p)

	return &query.TableTableFactor{
		Name:  astObjectNameToQuery(name),
		Alias: alias,
	}, nil
}

// parseDerivedTable parses a subquery in parentheses
func parseDerivedTable(p *Parser) (query.TableFactor, error) {
	if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	// Subquery
	subquery, err := parseQuery(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return nil, err
	}

	// Check for alias (required for subqueries)
	alias, err := parseTableAlias(p)
	if err != nil {
		return nil, err
	}

	return &query.DerivedTableFactor{
		Subquery: wrapQueryAsQuery(subquery),
		Alias:    alias,
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

// parseLateralTable parses LATERAL (subquery)
func parseLateralTable(p *Parser) (query.TableFactor, error) {
	if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	_, err := parseQuery(p)
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return nil, err
	}

	alias, err := parseTableAlias(p)
	if err != nil {
		return nil, err
	}

	return &query.DerivedTableFactor{
		Lateral: true,
		Alias:   alias,
	}, nil
}

// parseTableFunction parses TABLE(<expr>) [AS <alias>]
// Reference: src/parser/mod.rs:15490-15496
func parseTableFunction(p *Parser) (query.TableFactor, error) {
	if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse the expression inside TABLE(...)
	ep := NewExpressionParser(p)
	exprVal, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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
		if word, ok := tok.Token.(tokenizer.TokenWord); ok {
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
	if !p.ConsumeToken(tokenizer.TokenLParen{}) {
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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

	for {
		if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
			return nil, err
		}

		// Check for empty row () - MySQL allows this
		// If the dialect supports empty projections and we see RParen immediately, it's an empty row
		if _, isRParen := p.PeekToken().Token.(tokenizer.TokenRParen); isRParen {
			// Empty row - consume the RParen
			p.AdvanceToken()
			rows = append(rows, []query.Expr{})
		} else {
			// Non-empty row - parse expressions
			row, err := parseCommaSeparatedQueryExprs(p)
			if err != nil {
				return nil, err
			}

			if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
				return nil, err
			}

			rows = append(rows, row)
		}

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	// Create the Values expression
	values := &query.Values{
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
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

		if p.ConsumeToken(tokenizer.TokenLParen{}) {
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
		if !p.ConsumeToken(tokenizer.TokenComma{}) {
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
		if word, ok := tok.Token.(tokenizer.TokenWord); ok && word.Word.Keyword == token.NoKeyword {
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
	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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
	// DEBUG
	curTok := p.GetCurrentToken()
	fmt.Printf("DEBUG parseWindowFrameBound: curTok=%T=%v\n", curTok.Token, curTok.Token)

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
	ep := NewExpressionParser(p)
	exprBound, err := ep.ParseExpr()
	if err != nil {
		return nil, err
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
	return nil, fmt.Errorf("UPDATE in query parsing not yet fully implemented")
}

func parseDeleteInQuery(p *Parser, with interface{}) (ast.Statement, error) {
	return nil, fmt.Errorf("DELETE in query parsing not yet fully implemented")
}

func parseMergeInQuery(p *Parser, with interface{}) (ast.Statement, error) {
	return nil, fmt.Errorf("MERGE in query parsing not yet fully implemented")
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
	if p.PeekToken().Token.Equals(tokenizer.TokenLParen{}) {
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
	if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); !ok {
		return query.CTE{}, p.Expected("( before CTE subquery", p.PeekToken())
	}
	p.AdvanceToken() // consume (

	// Parse the inner query
	innerQuery, err := parseQuery(p)
	if err != nil {
		return query.CTE{}, err
	}

	// Expect closing parenthesis
	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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
		if !p.ConsumeToken(tokenizer.TokenComma{}) {
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

func parseFetch(p *Parser) (interface{}, error) {
	return nil, fmt.Errorf("FETCH parsing not yet fully implemented")
}

func parseOffsetClause(p *Parser) (interface{}, error) {
	return nil, nil
}

func parseForClause(p *Parser) (interface{}, error) {
	return nil, nil
}

func parseOrderByExpr(p *Parser) (interface{}, error) {
	return nil, fmt.Errorf("ORDER BY expression parsing not yet fully implemented")
}

func parseLimitClause(p *Parser) (interface{}, error) {
	return nil, fmt.Errorf("LIMIT clause parsing not yet fully implemented")
}

func maybeParseOptimizerHints(p *Parser) ([]interface{}, error) {
	return nil, nil
}

func parseSetExpr(p *Parser) (interface{}, error) {
	return nil, fmt.Errorf("set expression parsing not yet fully implemented")
}

func parseSubquery(p *Parser) (interface{}, error) {
	return nil, fmt.Errorf("subquery parsing not yet fully implemented")
}

func parseOutputClause(p *Parser, tok tokenizer.TokenWithSpan) (interface{}, error) {
	return nil, nil
}

// parsePipeOperators parses BigQuery/DuckDB pipe operators (|> SELECT, |> WHERE, etc.)
// Reference: src/parser/mod.rs:13726 parse_pipe_operators
func parsePipeOperators(p *Parser) ([]query.PipeOperator, error) {
	var pipeOperators []query.PipeOperator

	for p.ConsumeToken(tokenizer.TokenVerticalBarRightAngleBracket{}) {
		// Parse the keyword after |>
		tok := p.PeekToken()
		word, ok := tok.Token.(tokenizer.TokenWord)
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
			// Simplified AGGREGATE parsing - full implementation would handle GROUP BY
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeAggregate{},
			})

		case "RENAME":
			mappings, err := parseQueryIdentWithAliasList(p)
			if err != nil {
				return nil, err
			}
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeRename{Mappings: mappings},
			})

		case "UNION", "INTERSECT", "EXCEPT":
			// These are more complex and would need full implementation
			// For now, add a placeholder
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeUnion{},
			})

		case "TABLESAMPLE":
			// Simplified TABLESAMPLE parsing
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeTableSample{},
			})

		case "CALL":
			// Simplified CALL parsing
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeCall{},
			})

		case "PIVOT":
			// Simplified PIVOT parsing
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipePivot{},
			})

		case "UNPIVOT":
			// Simplified UNPIVOT parsing
			pipeOperators = append(pipeOperators, query.PipeOperator{
				Type: &query.PipeUnpivot{},
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

// parseQueryAssignments parses a comma-separated list of assignments (col = val) for pipe operators
func parseQueryAssignments(p *Parser) ([]query.Assignment, error) {
	var assignments []query.Assignment

	for {
		col, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		if _, err := p.ExpectToken(tokenizer.TokenEq{}); err != nil {
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
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

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return ids, nil
}

// parseQueryIdentWithAliasList parses a comma-separated list of identifiers with optional aliases
func parseQueryIdentWithAliasList(p *Parser) ([]query.IdentWithAlias, error) {
	var mappings []query.IdentWithAlias

	for {
		id, err := p.ParseIdentifier()
		if err != nil {
			return nil, err
		}

		mapping := query.IdentWithAlias{
			Ident: query.Ident{Value: id.Value},
		}

		// Check for optional AS alias
		if p.ParseKeyword("AS") {
			alias, err := p.ParseIdentifier()
			if err != nil {
				return nil, err
			}
			mapping.Alias = query.Ident{Value: alias.Value}
		}

		mappings = append(mappings, mapping)

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return mappings, nil
}
