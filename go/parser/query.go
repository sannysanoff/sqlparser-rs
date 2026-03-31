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
	"github.com/user/sqlparser/ast/query"
	"github.com/user/sqlparser/span"
	"github.com/user/sqlparser/token"
	"github.com/user/sqlparser/tokenizer"
)

// SelectStatement wraps query.Select to implement ast.Statement
type SelectStatement struct {
	ast.BaseStatement
	query.Select
}

// Span returns the source span
func (s *SelectStatement) Span() span.Span {
	return s.Select.Span()
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

// parseQuery parses a SELECT or other query statement
func parseQuery(p *Parser) (ast.Statement, error) {
	// Check for WITH clause (Common Table Expressions)
	if p.ParseKeyword("WITH") {
		_ = p.ParseKeyword("RECURSIVE")

		// For now, we just skip the CTE definitions
		// A proper implementation would store and use them
		tok := p.PeekToken()
		if _, ok := tok.Token.(tokenizer.TokenWord); ok {
			// Try to parse at least one CTE to skip past it
			if _, err := p.ParseIdentifier(); err == nil {
				// Check for column list
				if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
					p.ParseParenthesizedColumnList()
				}

				if p.ParseKeyword("AS") {
					_ = p.ParseKeyword("MATERIALIZED")
					_ = p.ParseKeyword("NOT")
					_ = p.ParseKeyword("MATERIALIZED")

					if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
						p.AdvanceToken() // consume (
						// Recursively skip the inner query
						parseQuery(p)
						p.ExpectToken(tokenizer.TokenRParen{})
					}
				}

				// Skip additional CTEs separated by comma
				for p.ConsumeToken(tokenizer.TokenComma{}) {
					if _, err := p.ParseIdentifier(); err == nil {
						if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
							p.ParseParenthesizedColumnList()
						}
						if p.ParseKeyword("AS") {
							_ = p.ParseKeyword("MATERIALIZED")
							_ = p.ParseKeyword("NOT")
							_ = p.ParseKeyword("MATERIALIZED")
							if _, ok := p.PeekToken().Token.(tokenizer.TokenLParen); ok {
								p.AdvanceToken()
								parseQuery(p)
								p.ExpectToken(tokenizer.TokenRParen{})
							}
						}
					}
				}
			}
		}
	}

	// Parse the actual query body
	return parseQueryBody(p)
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
	if p.ParseKeyword("LIMIT") {
		ep := NewExpressionParser(p)
		limitExpr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		// Check for OFFSET (PostgreSQL style: LIMIT x OFFSET y)
		if p.ParseKeyword("OFFSET") {
			offsetExpr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			// Store in the query somehow - for now just ignore
			_ = limitExpr
			_ = offsetExpr
		} else {
			// Just LIMIT without OFFSET
			_ = limitExpr
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
			_ = offsetExpr
			_ = limitExpr
		} else {
			// Just OFFSET without LIMIT
			_ = offsetExpr
		}
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
func parseTableFactor(p *Parser) (query.TableFactor, error) {
	// Check for subquery: (SELECT ...)
	if isSubqueryStart(p) {
		return parseDerivedTable(p)
	}

	// Check for LATERAL (if supported)
	if p.ParseKeyword("LATERAL") {
		return parseLateralTable(p)
	}

	// Otherwise, it's a table name
	return parseTableName(p)
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
		"WINDOW": true, "QUALIFY": true,
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

		row, err := parseCommaSeparatedQueryExprs(p)
		if err != nil {
			return nil, err
		}

		if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
			return nil, err
		}

		rows = append(rows, row)

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
		asc := true
		if p.ParseKeyword("DESC") {
			asc = false
		} else if p.ParseKeyword("ASC") {
			asc = true
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
				Asc:        &asc,
				NullsFirst: nullsFirst,
			},
		})

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return exprs, nil
}

// parseCommaSeparatedQueryExprs parses a comma-separated list of query.Expr
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
	return query.CTE{}, fmt.Errorf("CTE parsing not yet fully implemented")
}

func parseCTEList(p *Parser) ([]query.CTE, error) {
	return nil, fmt.Errorf("CTE list parsing not yet fully implemented")
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
