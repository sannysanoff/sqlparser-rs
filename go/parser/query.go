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
	"github.com/user/sqlparser/dialects/postgresql"
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

	// Check if we need to create a Query wrapper (for WITH clause, pipe operators, or FOR clause)
	needsQueryWrapper := withClause != nil

	// Get pipe operators and FOR clause from SelectStatement if present
	var pipeOperators []query.PipeOperator
	var forClause *query.ForClause
	if selStmt, ok := body.(*SelectStatement); ok {
		pipeOperators = selStmt.PipeOperators
		forClause = selStmt.ForClause
		if len(pipeOperators) > 0 || forClause != nil {
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

// parseSelect parses a SELECT statement
func parseSelect(p *Parser) (ast.Statement, error) {
	// Expect SELECT keyword
	_, err := p.ExpectKeyword("SELECT")
	if err != nil {
		return nil, err
	}

	// Parse DISTINCT / ALL (optional)
	var distinct *query.Distinct
	if p.ParseKeyword("DISTINCT") {
		distinctVal := query.DistinctDistinct
		distinct = &distinctVal
	} else if p.ParseKeyword("ALL") {
		distinctVal := query.DistinctAll
		distinct = &distinctVal
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

	// Parse FOR clause (MSSQL FOR XML/FOR JSON/FOR BROWSE)
	// Reference: src/parser/mod.rs:13682-13691
	var forClause *query.ForClause
	for p.ParseKeyword("FOR") {
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
		} else {
			// It's a LOCK clause, not FOR XML/JSON/BROWSE
			// TODO: parse lock clauses
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
			Distinct:            distinct,
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
		return &query.Wildcard{}, nil
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
	// Check if it's SELECT or WITH after the paren
	nextTok := p.PeekToken()
	if word, ok := nextTok.Token.(token.TokenWord); ok {
		kw := strings.ToUpper(string(word.Word.Keyword))
		return kw == "SELECT" || kw == "WITH"
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

	var table query.TableFactor = &query.DerivedTableFactor{
		Subquery: wrapQueryAsQuery(subquery),
		Alias:    alias,
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
		if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		valueColumns, err = parseCommaSeparatedQueryExprs(p)
		if err != nil {
			return nil, err
		}
		if _, err := p.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	} else {
		// Single value column
		ep := NewExpressionParser(p)
		expr, err := ep.ParseExpr()
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

	for {
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

	// Check for alias
	alias, _ := tryParseTableAlias(p)

	return &query.TableTableFactor{
		Name:  astObjectNameToQuery(name),
		Alias: alias,
	}, nil
}

// parseTableFunctionArgs parses the arguments for a table-valued function
// Handles both empty args () and non-empty args (arg1, arg2, ...)
func parseTableFunctionArgs(p *Parser) ([]query.FunctionArg, error) {
	var args []query.FunctionArg

	// Check for empty args: )
	if p.ConsumeToken(token.TokenRParen{}) {
		return args, nil
	}

	// Parse non-empty argument list
	ep := NewExpressionParser(p)
	for {
		argExpr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		args = append(args, query.FunctionArg{Expr: exprToQueryExpr(argExpr)})

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
	if _, err := p.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	_, err := parseQuery(p)
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

	return &query.DerivedTableFactor{
		Lateral: true,
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

		if !p.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return mappings, nil
}

// parseSetQuantifier parses ALL/DISTINCT for set operations
func parseSetQuantifier(p *Parser, op string) query.SetQuantifier {
	if p.ParseKeyword("ALL") {
		return query.SetQuantifierAll
	}
	if p.ParseKeyword("DISTINCT") {
		return query.SetQuantifierDistinct
	}
	return query.SetQuantifierNone
}

// parseDistinctRequiredSetQuantifier parses DISTINCT (required) for INTERSECT/EXCEPT
func parseDistinctRequiredSetQuantifier(p *Parser, op string) query.SetQuantifier {
	if p.ParseKeyword("DISTINCT") {
		return query.SetQuantifierDistinct
	}
	if p.ParseKeyword("ALL") {
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

	var unit query.TableSampleUnit = query.TableSampleUnitPercent
	if p.ParseKeyword("PERCENT") {
		unit = query.TableSampleUnitPercent
	} else if p.ParseKeyword("ROWS") {
		unit = query.TableSampleUnitRows
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
	unitPtr := &unit

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
