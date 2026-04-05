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

	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/operator"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/parseriface"
	"github.com/user/sqlparser/token"
)

// parseFunction parses a function call expression
func (ep *ExpressionParser) parseFunction(name *expr.ObjectName) (expr.Expr, error) {
	return ep.parseFunctionWithName(name)
}

// parseFunctionWithName parses a function call with the given name
func (ep *ExpressionParser) parseFunctionWithName(name *expr.ObjectName) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Check for empty args
	if ep.parser.ConsumeToken(token.TokenRParen{}) {
		fnExpr := &expr.FunctionExpr{
			Name:           name,
			UsesOdbcSyntax: false,
			Args:           &expr.FunctionArguments{None: true},
			SpanVal:        name.Span(),
		}

		// Parse OVER clause for window functions (even with empty args)
		if ep.parser.ParseKeyword("OVER") {
			windowSpec, err := ep.parseWindowSpec()
			if err != nil {
				return nil, err
			}
			fnExpr.Over = windowSpec
		}

		return fnExpr, nil
	}

	// Check for DISTINCT/ALL
	duplicateTreatment := expr.DuplicateNone
	if ep.parser.ParseKeyword("DISTINCT") {
		duplicateTreatment = expr.DuplicateDistinct
	} else if ep.parser.ParseKeyword("ALL") {
		duplicateTreatment = expr.DuplicateAll
	}

	// Parse arguments
	args, clauses, err := ep.parseFunctionArgs()
	if err != nil {
		return nil, err
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	fnExpr := &expr.FunctionExpr{
		Name: name,
		Args: &expr.FunctionArguments{
			List: &expr.FunctionArgumentList{
				DuplicateTreatment: duplicateTreatment,
				Args:               args,
				Clauses:            clauses,
			},
		},
		SpanVal: mergeSpans(name.Span(), ep.parser.GetCurrentToken().Span),
	}

	// Parse OVER clause for window functions
	if ep.parser.ParseKeyword("OVER") {
		windowSpec, err := ep.parseWindowSpec()
		if err != nil {
			return nil, err
		}
		fnExpr.Over = windowSpec
	}

	// Parse FILTER clause for aggregate functions (if supported)
	if dialects.SupportsFilterDuringAggregation(dialect) && ep.parser.ParseKeyword("FILTER") {
		if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		if !ep.parser.ParseKeyword("WHERE") {
			return nil, fmt.Errorf("expected WHERE after FILTER")
		}
		filterExpr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		fnExpr.Filter = filterExpr
		if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	}

	// Parse null treatment (IGNORE NULLS / RESPECT NULLS)
	if ep.parser.ParseKeyword("IGNORE") {
		if ep.parser.ParseKeyword("NULLS") {
			fnExpr.NullTreatment = expr.NullTreatmentIgnore
		}
	} else if ep.parser.ParseKeyword("RESPECT") {
		if ep.parser.ParseKeyword("NULLS") {
			fnExpr.NullTreatment = expr.NullTreatmentRespect
		}
	}

	// Parse WITHIN GROUP for ordered set aggregates
	if dialects.SupportsWithinAfterArrayAggregation(dialect) && ep.parser.ParseKeywords([]string{"WITHIN", "GROUP"}) {
		if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		if !ep.parser.ParseKeyword("ORDER") || !ep.parser.ParseKeyword("BY") {
			return nil, fmt.Errorf("expected ORDER BY after WITHIN GROUP")
		}
		orderBy, err := ep.parseCommaSeparatedOrderByExprs()
		if err != nil {
			return nil, err
		}
		if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		// TODO: WithinGroup expects []Expr but OrderByExpr doesn't implement Expr
		// This is a design issue that needs to be fixed in the AST types
		// For now, we skip setting WithinGroup
		_ = orderBy
		// fnExpr.WithinGroup = orderBy
	}

	return fnExpr, nil
}

// parseFunctionArgs parses function arguments and clauses
func (ep *ExpressionParser) parseFunctionArgs() ([]expr.FunctionArg, []expr.FunctionArgumentClause, error) {
	var args []expr.FunctionArg
	var clauses []expr.FunctionArgumentClause

	for {
		// Check for clause keywords
		if ep.parser.PeekKeyword("ORDER") {
			break // ORDER BY is a clause, not an argument
		}
		if ep.parser.PeekKeyword("LIMIT") {
			break
		}
		if ep.parser.PeekKeyword("SEPARATOR") {
			break
		}
		if ep.parser.PeekKeyword("ON") {
			break // ON OVERFLOW
		}
		if ep.parser.PeekKeyword("HAVING") {
			break
		}
		if ep.parser.PeekKeyword("NULL") {
			next := ep.parser.PeekNthToken(1)
			if word, ok := next.Token.(token.TokenWord); ok && word.Word.Keyword == "ON" {
				break // NULL ON NULL
			}
		}
		if ep.parser.PeekKeyword("ABSENT") {
			break // ABSENT ON NULL
		}

		// Parse argument
		arg, err := ep.parseFunctionArg()
		if err != nil {
			return nil, nil, err
		}
		args = append(args, arg)

		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	// Parse any clauses
	for {
		if ep.parser.PeekKeyword("ORDER") {
			clause, err := ep.parseOrderByClause()
			if err != nil {
				return nil, nil, err
			}
			clauses = append(clauses, clause)
		} else if ep.parser.PeekKeyword("LIMIT") {
			clause, err := ep.parseLimitClause()
			if err != nil {
				return nil, nil, err
			}
			clauses = append(clauses, clause)
		} else if ep.parser.PeekKeyword("SEPARATOR") {
			clause, err := ep.parseSeparatorClause()
			if err != nil {
				return nil, nil, err
			}
			clauses = append(clauses, clause)
		} else if ep.parser.PeekKeyword("ON") {
			clause, err := ep.parseOnOverflowClause()
			if err != nil {
				return nil, nil, err
			}
			clauses = append(clauses, clause)
		} else if ep.parser.PeekKeyword("HAVING") {
			clause, err := ep.parseHavingClause()
			if err != nil {
				return nil, nil, err
			}
			clauses = append(clauses, clause)
		} else if ep.parser.ParseKeywords([]string{"NULL", "ON", "NULL"}) {
			clauses = append(clauses, &expr.JsonNullOnNullClause{
				Clause: expr.JsonNullNull,
			})
		} else if ep.parser.ParseKeywords([]string{"ABSENT", "ON", "NULL"}) {
			clauses = append(clauses, &expr.JsonNullOnNullClause{
				Clause: expr.JsonNullAbsent,
			})
		} else if ep.parser.PeekKeyword("RETURNING") {
			clause, err := ep.parseJsonReturningClause()
			if err != nil {
				return nil, nil, err
			}
			clauses = append(clauses, clause)
		} else {
			break
		}
	}

	return args, clauses, nil
}

// parseFunctionArg parses a single function argument
// Handles named arguments (name => value, name = value, etc.) and unnamed expressions
// Reference: src/parser/mod.rs:17788-17836 parse_function_args
func (ep *ExpressionParser) parseFunctionArg() (expr.FunctionArg, error) {
	// Check for wildcard * (used in COUNT(*) and similar)
	if ep.parser.ConsumeToken(token.TokenMul{}) {
		return &expr.FunctionArgExpr{Expr: &expr.Wildcard{}}, nil
	}

	dialect := ep.parser.GetDialect()

	// Try to parse as named argument if dialect supports it
	// Check for: name => value, name = value, name := value, name : value
	if dialects.SupportsNamedFnArgsWithRArrowOperator(dialect) ||
		dialects.SupportsNamedFnArgsWithEqOperator(dialect) ||
		dialects.SupportsNamedFnArgsWithAssignmentOperator(dialect) ||
		dialects.SupportsNamedFnArgsWithColonOperator(dialect) {

		// Try to parse named argument using "maybe parse" pattern
		// First, attempt to parse the name part
		var arg expr.FunctionArg
		var err error

		if dialects.SupportsNamedFnArgsWithExprName(dialect) {
			// Name can be an arbitrary expression
			arg, err = ep.tryParseNamedArgWithExprName()
		} else {
			// Name must be a simple identifier
			arg, err = ep.tryParseNamedArgWithIdentName()
		}

		if err == nil && arg != nil {
			return arg, nil
		}
		// If named arg parsing failed, fall through to regular expression parsing
	}

	// Parse as regular (unnamed) argument
	argExpr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	return &expr.FunctionArgExpr{Expr: argExpr}, nil
}

// tryParseNamedArgWithIdentName tries to parse named argument with identifier name
// Pattern: identifier => value, identifier = value, etc.
func (ep *ExpressionParser) tryParseNamedArgWithIdentName() (expr.FunctionArg, error) {
	// Save position for backtracking
	restore := ep.parser.SavePosition()

	// Try to parse an identifier
	name, err := ep.parseIdentifier()
	if err != nil {
		restore()
		return nil, err
	}

	// Check for named argument operator
	operator := ep.parseNamedArgOperator()
	if operator == "" {
		// No operator found, backtrack
		restore()
		return nil, fmt.Errorf("not a named argument")
	}

	// Parse the argument value (supports wildcard expressions)
	valueExpr, err := ep.parseWildcardExpr()
	if err != nil {
		restore()
		return nil, err
	}

	return &expr.FunctionArgNamed{
		Name:  name,
		Value: valueExpr,
	}, nil
}

// tryParseNamedArgWithExprName tries to parse named argument with expression name
// Pattern: expression => value (used by some dialects like BigQuery)
func (ep *ExpressionParser) tryParseNamedArgWithExprName() (expr.FunctionArg, error) {
	// Save position for backtracking
	restore := ep.parser.SavePosition()

	// Try to parse an expression as the name
	nameExpr, err := ep.ParseExpr()
	if err != nil {
		restore()
		return nil, err
	}

	// Check for named argument operator
	operator := ep.parseNamedArgOperator()
	if operator == "" {
		// No operator found, backtrack
		restore()
		return nil, fmt.Errorf("not a named argument")
	}

	// Parse the argument value
	valueExpr, err := ep.parseWildcardExpr()
	if err != nil {
		restore()
		return nil, err
	}

	// For expression-named args, we still use FunctionArgNamed but with expression as name
	// Convert expression to identifier if possible, otherwise use the expression directly
	return &expr.FunctionArgNamed{
		Name:  &expr.Ident{Value: nameExpr.String()},
		Value: valueExpr,
	}, nil
}

// parseNamedArgOperator tries to parse a named argument operator (=>, =, :=, :)
// Returns the operator string if found, empty string otherwise
func (ep *ExpressionParser) parseNamedArgOperator() string {
	dialect := ep.parser.GetDialect()

	tok := ep.parser.PeekToken()

	// Check for => operator (RArrow)
	if _, ok := tok.Token.(token.TokenRArrow); ok {
		if dialects.SupportsNamedFnArgsWithRArrowOperator(dialect) {
			ep.parser.NextToken() // consume
			return "=>"
		}
	}

	// Check for = operator (Eq)
	if _, ok := tok.Token.(token.TokenEq); ok {
		if dialects.SupportsNamedFnArgsWithEqOperator(dialect) {
			ep.parser.NextToken() // consume
			return "="
		}
	}

	// Check for := operator (Assignment)
	// Note: Need to check if Assignment token exists in tokenizer
	// For now, check for Colon followed by Eq
	if _, ok := tok.Token.(token.TokenColon); ok {
		nextTok := ep.parser.PeekNthToken(1)
		if _, ok := nextTok.Token.(token.TokenEq); ok {
			if dialects.SupportsNamedFnArgsWithAssignmentOperator(dialect) {
				ep.parser.NextToken() // consume :
				ep.parser.NextToken() // consume =
				return ":="
			}
		}
	}

	// Check for : operator (Colon) - used by some dialects
	if _, ok := tok.Token.(token.TokenColon); ok {
		if dialects.SupportsNamedFnArgsWithColonOperator(dialect) {
			ep.parser.NextToken() // consume
			return ":"
		}
	}

	return ""
}

// parseWildcardExpr parses an expression that can include wildcards (like *)
// Used for function argument values like COUNT(*)
func (ep *ExpressionParser) parseWildcardExpr() (expr.Expr, error) {
	// Check for wildcard
	if ep.parser.ConsumeToken(token.TokenMul{}) {
		return &expr.Wildcard{}, nil
	}
	// Otherwise parse as regular expression
	return ep.ParseExpr()
}

// parseOrderByClause parses ORDER BY clause in function arguments
func (ep *ExpressionParser) parseOrderByClause() (expr.FunctionArgumentClause, error) {
	if !ep.parser.ParseKeyword("ORDER") || !ep.parser.ParseKeyword("BY") {
		return nil, fmt.Errorf("expected ORDER BY")
	}

	orderBy, err := ep.parseCommaSeparatedOrderByExprs()
	if err != nil {
		return nil, err
	}

	// TODO: orderBy is []*OrderByExpr but OrderByClause.OrderBy expects []Expr
	// This is a design issue that needs to be fixed in the AST types
	// For now, we create an empty slice
	_ = orderBy
	return &expr.OrderByClause{OrderBy: []expr.Expr{}}, nil
}

// parseCommaSeparatedOrderByExprs parses a comma-separated list of ORDER BY expressions
func (ep *ExpressionParser) parseCommaSeparatedOrderByExprs() ([]*expr.OrderByExpr, error) {
	var items []*expr.OrderByExpr

	for {
		item, err := ep.parseOrderByExpr()
		if err != nil {
			return nil, err
		}
		items = append(items, item)

		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return items, nil
}

// parseOrderByExpr parses a single ORDER BY expression
func (ep *ExpressionParser) parseOrderByExpr() (*expr.OrderByExpr, error) {
	exprVal, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Parse ASC/DESC
	var asc *bool
	if ep.parser.ParseKeyword("ASC") {
		b := true
		asc = &b
	} else if ep.parser.ParseKeyword("DESC") {
		b := false
		asc = &b
	}

	// Parse NULLS FIRST/LAST
	var nullsFirst *bool
	if ep.parser.ParseKeywords([]string{"NULLS", "FIRST"}) {
		b := true
		nullsFirst = &b
	} else if ep.parser.ParseKeywords([]string{"NULLS", "LAST"}) {
		b := false
		nullsFirst = &b
	}

	return &expr.OrderByExpr{
		Expr:       exprVal,
		Asc:        asc,
		NullsFirst: nullsFirst,
	}, nil
}

// parseLimitClause parses LIMIT clause in function arguments
func (ep *ExpressionParser) parseLimitClause() (expr.FunctionArgumentClause, error) {
	if !ep.parser.ParseKeyword("LIMIT") {
		return nil, fmt.Errorf("expected LIMIT")
	}

	limitExpr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	return &expr.LimitClause{Limit: limitExpr}, nil
}

// parseSeparatorClause parses SEPARATOR clause for LISTAGG
func (ep *ExpressionParser) parseSeparatorClause() (expr.FunctionArgumentClause, error) {
	if !ep.parser.ParseKeyword("SEPARATOR") {
		return nil, fmt.Errorf("expected SEPARATOR")
	}

	val, err := ep.parseValue()
	if err != nil {
		return nil, err
	}

	return &expr.SeparatorClause{Value: val}, nil
}

// parseOnOverflowClause parses ON OVERFLOW clause for LISTAGG
func (ep *ExpressionParser) parseOnOverflowClause() (expr.FunctionArgumentClause, error) {
	if !ep.parser.ParseKeyword("ON") || !ep.parser.ParseKeyword("OVERFLOW") {
		return nil, fmt.Errorf("expected ON OVERFLOW")
	}

	if ep.parser.ParseKeyword("ERROR") {
		return &expr.OnOverflowClause{OnOverflow: expr.ListAggError}, nil
	}

	if ep.parser.ParseKeyword("TRUNCATE") {
		return &expr.OnOverflowClause{OnOverflow: expr.ListAggTruncate}, nil
	}

	return nil, fmt.Errorf("expected ERROR or TRUNCATE after ON OVERFLOW")
}

// parseHavingClause parses HAVING clause for ANY_VALUE
func (ep *ExpressionParser) parseHavingClause() (expr.FunctionArgumentClause, error) {
	if !ep.parser.ParseKeyword("HAVING") {
		return nil, fmt.Errorf("expected HAVING")
	}

	isMax := false
	if ep.parser.ParseKeyword("MAX") {
		isMax = true
	} else if !ep.parser.ParseKeyword("MIN") {
		return nil, fmt.Errorf("expected MAX or MIN after HAVING")
	}

	boundExpr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	return &expr.HavingClause{
		Bound: &expr.HavingBound{
			IsMax: isMax,
			Expr:  boundExpr,
		},
	}, nil
}

// parseJsonReturningClause parses RETURNING clause for JSON functions
func (ep *ExpressionParser) parseJsonReturningClause() (expr.FunctionArgumentClause, error) {
	if !ep.parser.ParseKeyword("RETURNING") {
		return nil, fmt.Errorf("expected RETURNING")
	}

	// For now, just consume the data type as a string
	// Full implementation would parse the actual data type
	dtype, err := ep.parseIdentifier()
	if err != nil {
		return nil, err
	}

	return &expr.JsonReturning{
		Clause: &expr.JsonReturningClause{DataType: dtype.Value},
	}, nil
}

// parseWindowSpec parses a window specification for OVER clause
// Handles: OVER window_name, OVER (window_spec), OVER ()
func (ep *ExpressionParser) parseWindowSpec() (*expr.WindowType, error) {
	// Check for empty window spec: OVER ()
	// We need to check if next is ( and if so, if the following is )
	tok := ep.parser.PeekToken()
	if _, ok := tok.Token.(token.TokenLParen); ok {
		// Check if it's an empty ()
		nextTok := ep.parser.PeekNthToken(1)
		if _, ok := nextTok.Token.(token.TokenRParen); ok {
			// Empty window specification OVER ()
			ep.parser.AdvanceToken() // consume (
			ep.parser.AdvanceToken() // consume )
			return &expr.WindowType{Spec: &expr.WindowSpec{}}, nil
		}
	}

	// Check for named window reference (just an identifier, no parentheses)
	tok = ep.parser.PeekToken()
	if word, ok := tok.Token.(token.TokenWord); ok {
		// If it's not a keyword that starts window spec, treat as named window
		if word.Word.Keyword != "PARTITION" && word.Word.Keyword != "ORDER" &&
			word.Word.Keyword != "ROWS" && word.Word.Keyword != "RANGE" &&
			word.Word.Keyword != "GROUPS" && word.Word.Keyword != "CURRENT" &&
			word.Word.Keyword != "UNBOUNDED" {
			// It's a named window reference
			name, err := ep.parseIdentifier()
			if err != nil {
				return nil, err
			}
			return &expr.WindowType{Named: name}, nil
		}
	}

	// Parse window specification in parentheses
	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}
	spec, err := ep.parseWindowSpecDetails()
	if err != nil {
		return nil, err
	}
	return &expr.WindowType{Spec: spec}, nil
}

// parseWindowSpecDetails parses the details of a window specification
// Expects: PARTITION BY ..., ORDER BY ..., ROWS/RANGE/GROUPS ...
// Note: Opening parenthesis was already consumed by caller
func (ep *ExpressionParser) parseWindowSpecDetails() (*expr.WindowSpec, error) {
	spec := &expr.WindowSpec{}

	// Opening paren should have been consumed by caller, so we start parsing the content
	// But actually parseWindowSpec handles empty (), so we need to check for RParen
	if ep.parser.ConsumeToken(token.TokenRParen{}) {
		return spec, nil
	}

	// Parse optional window name reference at the beginning
	// e.g., OVER (w PARTITION BY ...) where w is a named window
	if !ep.parser.PeekKeyword("PARTITION") && !ep.parser.PeekKeyword("ORDER") &&
		!ep.parser.PeekKeyword("ROWS") && !ep.parser.PeekKeyword("RANGE") &&
		!ep.parser.PeekKeyword("GROUPS") && !ep.parser.PeekKeyword("CURRENT") &&
		!ep.parser.PeekKeyword("UNBOUNDED") {
		// Might be a named window reference
		name, err := ep.parseIdentifier()
		if err == nil {
			spec.WindowName = name
		}
	}

	// Parse PARTITION BY
	if ep.parser.ParseKeyword("PARTITION") {
		if !ep.parser.ParseKeyword("BY") {
			return nil, fmt.Errorf("expected BY after PARTITION")
		}
		partitionBy, err := ep.parseCommaSeparatedExprs()
		if err != nil {
			return nil, err
		}
		spec.PartitionBy = partitionBy
	}

	// Parse ORDER BY
	if ep.parser.ParseKeywords([]string{"ORDER", "BY"}) {
		orderBy, err := ep.parseCommaSeparatedOrderByExprs()
		if err != nil {
			return nil, err
		}
		// Convert []*expr.OrderByExpr to []expr.Expr
		for _, ob := range orderBy {
			spec.OrderBy = append(spec.OrderBy, ob)
		}
	}

	// Parse window frame
	if ep.parser.PeekKeyword("ROWS") || ep.parser.PeekKeyword("RANGE") ||
		ep.parser.PeekKeyword("GROUPS") {
		frame, err := ep.parseWindowFrame()
		if err != nil {
			return nil, err
		}
		spec.WindowFrame = frame
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return spec, nil
}

// parseWindowFrame parses a window frame specification
func (ep *ExpressionParser) parseWindowFrame() (*expr.WindowFrame, error) {
	var units expr.WindowFrameUnits

	if ep.parser.ParseKeyword("ROWS") {
		units = expr.WindowRows
	} else if ep.parser.ParseKeyword("RANGE") {
		units = expr.WindowRange
	} else if ep.parser.ParseKeyword("GROUPS") {
		units = expr.WindowGroups
	} else {
		return nil, fmt.Errorf("expected ROWS, RANGE, or GROUPS")
	}

	frame := &expr.WindowFrame{Units: units}

	// Check for BETWEEN
	if ep.parser.ParseKeyword("BETWEEN") {
		startBound, err := ep.parseWindowFrameBound()
		if err != nil {
			return nil, err
		}
		frame.StartBound = startBound

		if !ep.parser.ParseKeyword("AND") {
			return nil, fmt.Errorf("expected AND in window frame BETWEEN")
		}

		endBound, err := ep.parseWindowFrameBound()
		if err != nil {
			return nil, err
		}
		frame.EndBound = endBound
	} else {
		bound, err := ep.parseWindowFrameBound()
		if err != nil {
			return nil, err
		}
		frame.StartBound = bound
	}

	return frame, nil
}

// parseWindowFrameBound parses a window frame bound
// Handles: CURRENT ROW, UNBOUNDED PRECEDING/FOLLOWING, <expr> PRECEDING/FOLLOWING
// Also handles INTERVAL expressions for RANGE frames
func (ep *ExpressionParser) parseWindowFrameBound() (*expr.WindowFrameBound, error) {
	if ep.parser.ParseKeyword("CURRENT") {
		if !ep.parser.ParseKeyword("ROW") {
			return nil, fmt.Errorf("expected ROW after CURRENT")
		}
		return &expr.WindowFrameBound{BoundType: expr.BoundTypeCurrentRow}, nil
	}

	if ep.parser.ParseKeyword("UNBOUNDED") {
		if ep.parser.ParseKeyword("PRECEDING") {
			return &expr.WindowFrameBound{BoundType: expr.BoundTypeUnboundedPreceding}, nil
		}
		if ep.parser.ParseKeyword("FOLLOWING") {
			return &expr.WindowFrameBound{BoundType: expr.BoundTypeUnboundedFollowing}, nil
		}
		return nil, fmt.Errorf("expected PRECEDING or FOLLOWING after UNBOUNDED")
	}

	// Check for INTERVAL expression (used in RANGE frames)
	// When the next token is a string literal, we need to parse it as an INTERVAL
	// Reference: src/parser/mod.rs:2575-2578
	nextTok := ep.parser.PeekTokenRef()
	if _, isString := nextTok.Token.(token.TokenSingleQuotedString); isString {
		// The expression is a string literal, parse as INTERVAL
		intervalExpr, err := ep.parseIntervalExpr()
		if err != nil {
			return nil, err
		}
		if ep.parser.ParseKeyword("PRECEDING") {
			return &expr.WindowFrameBound{
				BoundType: expr.BoundTypePreceding,
				Expr:      &intervalExpr,
			}, nil
		}
		if ep.parser.ParseKeyword("FOLLOWING") {
			return &expr.WindowFrameBound{
				BoundType: expr.BoundTypeFollowing,
				Expr:      &intervalExpr,
			}, nil
		}
		return nil, fmt.Errorf("expected PRECEDING or FOLLOWING after INTERVAL expression")
	}

	// Expression bound (number or other expression)
	exprBound, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if ep.parser.ParseKeyword("PRECEDING") {
		return &expr.WindowFrameBound{
			BoundType: expr.BoundTypePreceding,
			Expr:      &exprBound,
		}, nil
	}

	if ep.parser.ParseKeyword("FOLLOWING") {
		return &expr.WindowFrameBound{
			BoundType: expr.BoundTypeFollowing,
			Expr:      &exprBound,
		}, nil
	}

	return nil, fmt.Errorf("expected PRECEDING or FOLLOWING after expression")
}

// parseCaseExpr parses a CASE expression
func (ep *ExpressionParser) parseCaseExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	// Check for CASE expr WHEN... syntax vs CASE WHEN... syntax
	var operand expr.Expr
	if !ep.parser.PeekKeyword("WHEN") {
		// CASE expr WHEN...
		op, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		operand = op
	}

	// Parse WHEN clauses
	var conditions []expr.CaseWhen
	for ep.parser.ParseKeyword("WHEN") {
		cond, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if !ep.parser.ParseKeyword("THEN") {
			return nil, fmt.Errorf("expected THEN after WHEN condition")
		}
		result, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		conditions = append(conditions, expr.CaseWhen{
			Condition: cond,
			Result:    result,
		})
	}

	// Parse optional ELSE clause
	var elseResult expr.Expr
	if ep.parser.ParseKeyword("ELSE") {
		r, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		elseResult = r
	}

	if !ep.parser.ParseKeyword("END") {
		return nil, fmt.Errorf("expected END to close CASE expression")
	}

	spanEnd := ep.parser.GetCurrentToken().Span

	return &expr.CaseExpr{
		SpanVal:    mergeSpans(spanStart, spanEnd),
		Operand:    operand,
		Conditions: conditions,
		ElseResult: elseResult,
	}, nil
}

// parseSubqueryExpr parses a subquery expression assuming LParen was already consumed
func (ep *ExpressionParser) parseSubqueryExpr() (expr.Expr, error) {
	// Note: The opening parenthesis has already been consumed by parseParenthesizedPrefix
	// We need to parse the subquery content and then expect the closing parenthesis

	// Parse the actual query
	query, err := ep.parser.ParseQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to parse subquery: %w", err)
	}

	// Expect the closing parenthesis
	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Subquery{
		SpanVal: mergeSpans(query.Span(), ep.parser.GetCurrentToken().Span),
		Query: &expr.QueryExpr{
			SpanVal:   query.Span(),
			Statement: query,
		},
	}, nil
}

// parseExistsExpr parses an EXISTS expression
func (ep *ExpressionParser) parseExistsExpr(negated bool) (expr.Expr, error) {
	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse the actual query
	query, err := ep.parser.ParseQuery()
	if err != nil {
		return nil, fmt.Errorf("failed to parse EXISTS subquery: %w", err)
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Exists{
		SpanVal: mergeSpans(ep.parser.GetCurrentToken().Span, query.Span()),
		Subquery: &expr.QueryExpr{
			SpanVal:   query.Span(),
			Statement: query,
		},
		Negated: negated,
	}, nil
}

// parseNotExpr parses a NOT expression
func (ep *ExpressionParser) parseNotExpr() (expr.Expr, error) {
	// NOT can be a prefix or part of IN/EXISTS/BETWEEN
	// This handles the prefix case: NOT expr
	innerExpr, err := ep.ParseExprWithPrecedence(ep.getPrecedence(parseriface.PrecedenceUnaryNot))
	if err != nil {
		return nil, err
	}

	return &expr.UnaryOp{
		Op:      operator.UOpNot,
		Expr:    innerExpr,
		SpanVal: mergeSpans(ep.parser.GetCurrentToken().Span, innerExpr.Span()),
	}, nil
}
