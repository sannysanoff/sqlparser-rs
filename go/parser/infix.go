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

// parseInfix parses an infix expression.
// This is called after parsing the left-hand side and encountering an operator.
func (ep *ExpressionParser) parseInfix(left expr.Expr, precedence uint8) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	// TODO: The dialect system uses ast.Expr, but we use expr.Expr
	// This is an architectural mismatch that needs to be resolved.
	// For now, we skip the dialect hook and use default parsing.
	//
	// parsedExpr, handled, err := dialect.ParseInfix(ep.parser, left, precedence)
	// if err != nil {
	// 	return nil, err
	// }
	// if handled {
	// 	return parsedExpr, nil
	// }

	// Get the operator token
	ep.parser.AdvanceToken()
	tok := ep.parser.GetCurrentToken()
	tokSpan := tok.Span

	// Handle custom binary operators (PostgreSQL)
	if customOp, ok := tok.Token.(token.TokenCustomBinaryOperator); ok {
		// Parse right operand
		right, err := ep.ParseExprWithPrecedence(precedence)
		if err != nil {
			return nil, err
		}
		// Use BOpCustom for simple custom operators like &@, @@, etc.
		// These are output directly without OPERATOR() wrapper
		return &expr.BinaryOp{
			Left:             left,
			Op:               operator.BOpCustom,
			Right:            right,
			SpanVal:          mergeSpans(left.Span(), right.Span()),
			PGCustomOperator: []string{customOp.Value},
		}, nil
	}

	// Try to parse as a regular binary operator
	if op := ep.tokenToBinaryOperator(tok.Token); op != operator.BOpNone {
		return ep.parseBinaryOp(left, op, precedence, tokSpan)
	}

	// Handle word-based operators (AND, OR, IS, etc.)
	if word, ok := tok.Token.(token.TokenWord); ok {
		return ep.parseWordInfix(left, word, precedence, tokSpan)
	}

	// Handle special token operators
	switch tok.Token.(type) {
	case token.TokenDoubleColon:
		// PostgreSQL-style cast ::type
		return ep.parseDoubleColonCast(left)

	case token.TokenLBracket:
		// Array subscript: expr[index]
		return ep.parseArraySubscript(left)

	case token.TokenColon:
		// Semi-structured data access: expr:key or expr:[key]
		// Put back the colon token since parseJsonPath expects to consume it
		ep.parser.PrevToken()
		return ep.parseJsonAccess(left)

	case token.TokenExclamationMark:
		if dialects.SupportsFactorialOperator(dialect) {
			return &expr.UnaryOp{
				Op:      operator.UOpPGPostfixFactorial,
				Expr:    left,
				SpanVal: mergeSpans(left.Span(), tokSpan),
			}, nil
		}
	}

	return nil, fmt.Errorf("no infix parser for token %v", tok.Token)
}

// tokenToBinaryOperator converts a token to its corresponding binary operator
func (ep *ExpressionParser) tokenToBinaryOperator(tok token.Token) operator.BinaryOperator {
	dialect := ep.parser.GetDialect()

	switch tok.(type) {
	case token.TokenSpaceship:
		return operator.BOpSpaceship
	case token.TokenDoubleEq:
		return operator.BOpEq
	case token.TokenAssignment:
		return operator.BOpAssignment
	case token.TokenEq:
		return operator.BOpEq
	case token.TokenNeq:
		return operator.BOpNotEq
	case token.TokenGt:
		return operator.BOpGt
	case token.TokenGtEq:
		return operator.BOpGtEq
	case token.TokenLt:
		return operator.BOpLt
	case token.TokenLtEq:
		return operator.BOpLtEq
	case token.TokenPlus:
		return operator.BOpPlus
	case token.TokenMinus:
		return operator.BOpMinus
	case token.TokenMul:
		return operator.BOpMultiply
	case token.TokenMod:
		return operator.BOpModulo
	case token.TokenStringConcat:
		return operator.BOpStringConcat
	case token.TokenPipe:
		return operator.BOpBitwiseOr
	case token.TokenCaret:
		if dialect.Dialect() == "postgresql" {
			return operator.BOpPGExp
		}
		return operator.BOpBitwiseXor
	case token.TokenAmpersand:
		return operator.BOpBitwiseAnd
	case token.TokenDiv:
		return operator.BOpDivide
	case token.TokenDuckIntDiv:
		return operator.BOpDuckIntegerDivide
	case token.TokenShiftLeft:
		if dialects.SupportsBitwiseShiftOperators(dialect) {
			return operator.BOpPGBitwiseShiftLeft
		}
	case token.TokenShiftRight:
		if dialects.SupportsBitwiseShiftOperators(dialect) {
			return operator.BOpPGBitwiseShiftRight
		}
	case token.TokenSharp:
		if dialect.Dialect() == "postgresql" || dialect.Dialect() == "redshift" {
			return operator.BOpPGBitwiseXor
		}
	case token.TokenOverlap:
		if dialect.Dialect() == "postgresql" || dialect.Dialect() == "redshift" {
			return operator.BOpPGOverlap
		}
		if dialects.SupportsDoubleAmpersandOperator(dialect) {
			return operator.BOpAnd
		}
	case token.TokenCaretAt:
		if dialect.Dialect() == "postgresql" || dialect.Dialect() == "redshift" {
			return operator.BOpPGStartsWith
		}
	case token.TokenTilde:
		return operator.BOpPGRegexMatch
	case token.TokenTildeAsterisk:
		return operator.BOpPGRegexIMatch
	case token.TokenExclamationMarkTilde:
		return operator.BOpPGRegexNotMatch
	case token.TokenExclamationMarkTildeAsterisk:
		return operator.BOpPGRegexNotIMatch
	case token.TokenDoubleTilde:
		return operator.BOpPGLikeMatch
	case token.TokenDoubleTildeAsterisk:
		return operator.BOpPGILikeMatch
	case token.TokenExclamationMarkDoubleTilde:
		return operator.BOpPGNotLikeMatch
	case token.TokenExclamationMarkDoubleTildeAsterisk:
		return operator.BOpPGNotILikeMatch
	case token.TokenArrow:
		return operator.BOpArrow
	case token.TokenLongArrow:
		return operator.BOpLongArrow
	case token.TokenHashArrow:
		return operator.BOpHashArrow
	case token.TokenHashLongArrow:
		return operator.BOpHashLongArrow
	case token.TokenAtArrow:
		return operator.BOpAtArrow
	case token.TokenArrowAt:
		return operator.BOpArrowAt
	case token.TokenHashMinus:
		return operator.BOpHashMinus
	case token.TokenAtQuestion:
		return operator.BOpAtQuestion
	case token.TokenAtAt:
		return operator.BOpAtAt
	case token.TokenQuestion:
		return operator.BOpQuestion
	case token.TokenQuestionAnd:
		return operator.BOpQuestionAnd
	case token.TokenQuestionPipe:
		return operator.BOpQuestionPipe
	case token.TokenCustomBinaryOperator:
		// Custom operators require storing the name, which the current
		// BinaryOperator type doesn't support. For now, return BOpCustom.
		// TODO: Enhance BinaryOperator to support custom operator names.
		return operator.BOpCustom
	case token.TokenDoubleSharp:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpDoubleHash
		}
	case token.TokenAmpersandLeftAngleBracket:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpAndLt
		}
	case token.TokenAmpersandRightAngleBracket:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpAndGt
		}
	case token.TokenQuestionMarkDash:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpQuestionDash
		}
	case token.TokenAmpersandLeftAngleBracketVerticalBar:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpAndLtPipe
		}
	case token.TokenVerticalBarAmpersandRightAngleBracket:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpPipeAndGt
		}
	case token.TokenTwoWayArrow:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpLtDashGt
		}
	case token.TokenLeftAngleBracketCaret:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpLtCaret
		}
	case token.TokenRightAngleBracketCaret:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpGtCaret
		}
	case token.TokenQuestionMarkSharp:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpQuestionHash
		}
	case token.TokenQuestionMarkDoubleVerticalBar:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpQuestionDoublePipe
		}
	case token.TokenQuestionMarkDashVerticalBar:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpQuestionDashPipe
		}
	case token.TokenTildeEqual:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpTildeEq
		}
	case token.TokenShiftLeftVerticalBar:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpLtLtPipe
		}
	case token.TokenVerticalBarShiftRight:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpPipeGtGt
		}
	case token.TokenAtSign:
		if dialects.SupportsGeometricTypes(dialect) {
			return operator.BOpAt
		}
	}

	return operator.BOpNone
}

// parseBinaryOp parses a binary operation, handling ALL/ANY/SOME modifiers
func (ep *ExpressionParser) parseBinaryOp(left expr.Expr, op operator.BinaryOperator, precedence uint8, span token.Span) (expr.Expr, error) {
	// Check for ALL/ANY/SOME modifiers
	kw := ep.parser.ParseOneOfKeywords([]string{"ALL", "ANY", "SOME"})

	if kw != "" {
		// We have ALL/ANY/SOME - need to parse subquery or expression list
		if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}

		// Check if it's a subquery
		var right expr.Expr
		if ep.peekSubquery() {
			// Parse as subquery - the opening paren is already consumed
			// Try standard subquery first, then try with set operations
			sq, err := ep.parseSubqueryExpr()
			if err != nil {
				// Try parsing with set operations (e.g., (SELECT ...) UNION (SELECT ...))
				sq2, err2 := ep.parseSubqueryWithSetOps()
				if err2 != nil {
					return nil, err
				}
				right = sq2
			} else {
				right = sq
			}
		} else {
			// Check for parenthesized subquery with set operations (e.g., ((SELECT ...) UNION ...))
			sq, err := ep.parseSubqueryWithSetOps()
			if err == nil && sq != nil {
				right = sq
			} else {
				// Parse expression and expect closing paren
				r, err := ep.ParseExprWithPrecedence(precedence)
				if err != nil {
					return nil, err
				}
				right = r
				if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
			}
		}

		// Validate that the operator is a comparison operator
		validOps := map[operator.BinaryOperator]bool{
			operator.BOpGt: true, operator.BOpLt: true, operator.BOpGtEq: true,
			operator.BOpLtEq: true, operator.BOpEq: true, operator.BOpNotEq: true,
			operator.BOpPGRegexMatch: true, operator.BOpPGRegexIMatch: true,
			operator.BOpPGRegexNotMatch: true, operator.BOpPGRegexNotIMatch: true,
			operator.BOpPGLikeMatch: true, operator.BOpPGILikeMatch: true,
			operator.BOpPGNotLikeMatch: true, operator.BOpPGNotILikeMatch: true,
		}

		if !validOps[op] {
			return nil, fmt.Errorf("expected comparison operator with ALL/ANY/SOME, got %s", op)
		}

		switch kw {
		case "ALL":
			return &expr.AllOp{
				Left:      left,
				CompareOp: op,
				Right:     right,
				SpanVal:   mergeSpans(left.Span(), right.Span()),
			}, nil
		case "ANY", "SOME":
			return &expr.AnyOp{
				Left:      left,
				CompareOp: op,
				Right:     right,
				IsSome:    kw == "SOME",
				SpanVal:   mergeSpans(left.Span(), right.Span()),
			}, nil
		}
	}

	// Regular binary operation
	right, err := ep.ParseExprWithPrecedence(precedence)
	if err != nil {
		return nil, err
	}

	return &expr.BinaryOp{
		Left:    left,
		Op:      op,
		Right:   right,
		SpanVal: mergeSpans(left.Span(), right.Span()),
	}, nil
}

// parseWordInfix parses an infix operator starting with a word token
func (ep *ExpressionParser) parseWordInfix(left expr.Expr, word token.TokenWord, precedence uint8, span token.Span) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	switch word.Word.Keyword {
	case "AND":
		right, err := ep.ParseExprWithPrecedence(precedence)
		if err != nil {
			return nil, err
		}
		return &expr.BinaryOp{
			Left:    left,
			Op:      operator.BOpAnd,
			Right:   right,
			SpanVal: mergeSpans(left.Span(), right.Span()),
		}, nil

	case "OR":
		right, err := ep.ParseExprWithPrecedence(precedence)
		if err != nil {
			return nil, err
		}
		return &expr.BinaryOp{
			Left:    left,
			Op:      operator.BOpOr,
			Right:   right,
			SpanVal: mergeSpans(left.Span(), right.Span()),
		}, nil

	case "XOR":
		right, err := ep.ParseExprWithPrecedence(precedence)
		if err != nil {
			return nil, err
		}
		return &expr.BinaryOp{
			Left:    left,
			Op:      operator.BOpXor,
			Right:   right,
			SpanVal: mergeSpans(left.Span(), right.Span()),
		}, nil

	case "IS":
		return ep.parseIsExpr(left)

	case "AT":
		// AT TIME ZONE expression
		if ep.parser.ParseKeyword("TIME") && ep.parser.ParseKeyword("ZONE") {
			tz, err := ep.ParseExprWithPrecedence(precedence)
			if err != nil {
				return nil, err
			}
			return &expr.AtTimeZone{
				Timestamp: left,
				TimeZone:  tz,
				SpanVal:   mergeSpans(left.Span(), tz.Span()),
			}, nil
		}
		return nil, fmt.Errorf("expected TIME ZONE after AT")

	case "COLLATE":
		// COLLATE expression: expr COLLATE collation_name
		collation, err := ep.parseObjectName()
		if err != nil {
			return nil, err
		}
		return &expr.Collate{
			Expr:      left,
			Collation: collation,
			SpanVal:   mergeSpans(left.Span(), collation.Span()),
		}, nil

	case "NOT":
		// NOT can prefix IN, BETWEEN, LIKE, etc.
		return ep.parseNotPrefixedInfix(left, precedence)

	case "IN":
		return ep.parseInExpr(left, false)

	case "BETWEEN":
		return ep.parseBetweenExpr(left, false)

	case "LIKE":
		return ep.parseLikeExpr(left, false, false)

	case "ILIKE":
		return ep.parseLikeExpr(left, false, true)

	case "SIMILAR":
		if ep.parser.ParseKeyword("TO") {
			return ep.parseSimilarToExpr(left, false)
		}
		return nil, fmt.Errorf("expected TO after SIMILAR")

	case "REGEXP", "RLIKE":
		return ep.parseRLikeExpr(left, false, word.Word.Keyword == "REGEXP")

	case "DIV":
		// MySQL integer division operator
		right, err := ep.ParseExprWithPrecedence(precedence)
		if err != nil {
			return nil, err
		}
		return &expr.BinaryOp{
			Left:    left,
			Op:      operator.BOpMyIntegerDivide,
			Right:   right,
			SpanVal: mergeSpans(left.Span(), right.Span()),
		}, nil

	case "OVERLAPS":
		right, err := ep.ParseExprWithPrecedence(precedence)
		if err != nil {
			return nil, err
		}
		return &expr.BinaryOp{
			Left:    left,
			Op:      operator.BOpOverlaps,
			Right:   right,
			SpanVal: mergeSpans(left.Span(), right.Span()),
		}, nil

	case "NOTNULL":
		if dialects.SupportsNotnullOperator(dialect) {
			return &expr.IsNotNull{
				Expr:    left,
				SpanVal: mergeSpans(left.Span(), span),
			}, nil
		}

	case "MEMBER":
		if ep.parser.ParseKeyword("OF") {
			if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			arrayExpr, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			return &expr.MemberOfExpr{
				Value:   left,
				Array:   arrayExpr,
				SpanVal: mergeSpans(left.Span(), arrayExpr.Span()),
			}, nil
		}
		return nil, fmt.Errorf("expected OF after MEMBER")

	case "OPERATOR":
		if dialect.Dialect() == "postgresql" {
			// PostgreSQL custom operator: OPERATOR(schema.op)
			return ep.parsePgOperator(left, precedence)
		}
	}

	return nil, fmt.Errorf("unknown infix word operator: %s", word.Word.Keyword)
}

// parseNotPrefixedInfix parses NOT followed by IN, BETWEEN, LIKE, etc.
func (ep *ExpressionParser) parseNotPrefixedInfix(left expr.Expr, precedence uint8) (expr.Expr, error) {
	// Handle NOT NULL (only valid outside column definition context)
	if !ep.parser.InColumnDefinitionState() && ep.parser.ParseKeyword("NULL") {
		return &expr.IsNotNull{
			Expr:    left,
			SpanVal: left.Span(),
		}, nil
	}

	// Handle NOT with other operators
	if ep.parser.ParseKeyword("IN") {
		return ep.parseInExpr(left, true)
	}

	if ep.parser.ParseKeyword("BETWEEN") {
		return ep.parseBetweenExpr(left, true)
	}

	if ep.parser.ParseKeyword("LIKE") {
		return ep.parseLikeExpr(left, true, false)
	}

	if ep.parser.ParseKeyword("ILIKE") {
		return ep.parseLikeExpr(left, true, true)
	}

	if ep.parser.PeekKeyword("SIMILAR") {
		if ep.parser.ParseKeyword("SIMILAR") && ep.parser.ParseKeyword("TO") {
			return ep.parseSimilarToExpr(left, true)
		}
	}

	if kw := ep.parser.ParseOneOfKeywords([]string{"REGEXP", "RLIKE"}); kw != "" {
		return ep.parseRLikeExpr(left, true, kw == "REGEXP")
	}

	// In column definition context, NOT followed by something other than
	// IN, BETWEEN, LIKE, etc. is likely a column constraint (e.g., NOT NULL).
	// Put back the NOT token and return nil to signal that NOT is not an infix operator here.
	if ep.parser.InColumnDefinitionState() {
		ep.parser.PrevToken()
		return nil, nil
	}

	return nil, fmt.Errorf("expected IN, BETWEEN, LIKE, ILIKE, or REGEXP after NOT")
}

// parseIsExpr parses IS [NOT] NULL/TRUE/FALSE/DISTINCT FROM expressions
func (ep *ExpressionParser) parseIsExpr(left expr.Expr) (expr.Expr, error) {
	span := left.Span()

	if ep.parser.ParseKeyword("NULL") {
		return &expr.IsNull{
			Expr:    left,
			SpanVal: span,
		}, nil
	}

	if ep.parser.ParseKeywords([]string{"NOT", "NULL"}) {
		return &expr.IsNotNull{
			Expr:    left,
			SpanVal: span,
		}, nil
	}

	if ep.parser.ParseKeyword("TRUE") {
		return &expr.IsTrue{
			Expr:    left,
			SpanVal: span,
		}, nil
	}

	if ep.parser.ParseKeywords([]string{"NOT", "TRUE"}) {
		return &expr.IsNotTrue{
			Expr:    left,
			SpanVal: span,
		}, nil
	}

	if ep.parser.ParseKeyword("FALSE") {
		return &expr.IsFalse{
			Expr:    left,
			SpanVal: span,
		}, nil
	}

	if ep.parser.ParseKeywords([]string{"NOT", "FALSE"}) {
		return &expr.IsNotFalse{
			Expr:    left,
			SpanVal: span,
		}, nil
	}

	if ep.parser.ParseKeyword("UNKNOWN") {
		return &expr.IsUnknown{
			Expr:    left,
			SpanVal: span,
		}, nil
	}

	if ep.parser.ParseKeywords([]string{"NOT", "UNKNOWN"}) {
		return &expr.IsNotUnknown{
			Expr:    left,
			SpanVal: span,
		}, nil
	}

	if ep.parser.ParseKeywords([]string{"DISTINCT", "FROM"}) {
		right, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		return &expr.IsDistinctFrom{
			Left:    left,
			Right:   right,
			SpanVal: mergeSpans(left.Span(), right.Span()),
		}, nil
	}

	if ep.parser.ParseKeywords([]string{"NOT", "DISTINCT", "FROM"}) {
		right, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		return &expr.IsNotDistinctFrom{
			Left:    left,
			Right:   right,
			SpanVal: mergeSpans(left.Span(), right.Span()),
		}, nil
	}

	// Check for IS [form] NORMALIZED (Unicode normalization)
	normalized, err := ep.tryParseIsNormalized(left)
	if err != nil {
		return nil, err
	}
	if normalized != nil {
		return normalized, nil
	}

	return nil, fmt.Errorf("expected [NOT] NULL, TRUE, FALSE, DISTINCT FROM, or NORMALIZED after IS")
}

// tryParseIsNormalized attempts to parse IS [NOT] [form] NORMALIZED
func (ep *ExpressionParser) tryParseIsNormalized(left expr.Expr) (expr.Expr, error) {
	span := left.Span()
	negated := false

	if ep.parser.ParseKeyword("NOT") {
		negated = true
	}

	// Check for optional normalization form
	var form *expr.NormalizationForm
	if ep.parser.ParseKeyword("NFC") {
		f := expr.FormNFC
		form = &f
	} else if ep.parser.ParseKeyword("NFD") {
		f := expr.FormNFD
		form = &f
	} else if ep.parser.ParseKeyword("NFKC") {
		f := expr.FormNFKC
		form = &f
	} else if ep.parser.ParseKeyword("NFKD") {
		f := expr.FormNFKD
		form = &f
	}

	if ep.parser.ParseKeyword("NORMALIZED") {
		return &expr.IsNormalized{
			Expr:    left,
			Form:    form,
			Negated: negated,
			SpanVal: span,
		}, nil
	}

	// Put back NOT if we consumed it
	if negated {
		ep.parser.PrevToken()
	}

	return nil, nil
}

// parseInExpr parses IN (expr1, expr2, ...) or IN (subquery) or IN UNNEST(...)
func (ep *ExpressionParser) parseInExpr(left expr.Expr, negated bool) (expr.Expr, error) {
	// Check for UNNEST first (BigQuery syntax: IN UNNEST(array) - no parentheses after IN)
	// Reference: src/parser/mod.rs:4262
	if ep.parser.ParseKeyword("UNNEST") {
		if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		arrayExpr, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return &expr.InUnnest{
			Expr:      left,
			ArrayExpr: arrayExpr,
			Negated:   negated,
			SpanVal:   mergeSpans(left.Span(), arrayExpr.Span()),
		}, nil
	}

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Check for empty list (if dialect supports it)
	dialect := ep.parser.GetDialect()
	next := ep.parser.PeekTokenRef()
	if _, ok := next.Token.(token.TokenRParen); ok {
		if dialects.SupportsInEmptyList(dialect) {
			ep.parser.AdvanceToken() // consume )
			return &expr.InList{
				Expr:    left,
				List:    []expr.Expr{},
				Negated: negated,
				SpanVal: left.Span(),
			}, nil
		}
		return nil, fmt.Errorf("empty IN list not allowed")
	}

	// Check for subquery - handle both simple (SELECT ...) and complex ((SELECT ...) UNION (SELECT ...))
	// First check for double-parenthesized set operations
	if _, ok := next.Token.(token.TokenLParen); ok {
		// This might be ((SELECT ...) UNION ...) - a subquery with set operations inside
		// Try to parse it as a subquery with set operations first
		subqExpr, err := ep.parseSubqueryWithSetOps()
		if err == nil && subqExpr != nil {
			if subq, ok := subqExpr.(*expr.Subquery); ok {
				if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
				return &expr.InSubquery{
					Expr: left,
					Subquery: &expr.QueryExpr{
						SpanVal:   subq.Span(),
						Statement: subq.Query.Statement,
					},
					Negated: negated,
					SpanVal: mergeSpans(left.Span(), ep.parser.GetCurrentToken().Span),
				}, nil
			}
		}
		// If that failed, continue to other parsing options
	}

	// Check for simple subquery (SELECT ... or WITH ...)
	if ep.peekSubquery() {
		subquery, err := ep.parser.ParseQuery()
		if err != nil {
			return nil, err
		}
		if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return &expr.InSubquery{
			Expr: left,
			Subquery: &expr.QueryExpr{
				SpanVal:   subquery.Span(),
				Statement: subquery,
			},
			Negated: negated,
			SpanVal: mergeSpans(left.Span(), ep.parser.GetCurrentToken().Span),
		}, nil
	}

	// Check for parenthesized subquery with set operations
	if _, ok := next.Token.(token.TokenLParen); ok {
		// Try to parse as a subquery with set operations
		// First, try to parse using parseSubqueryWithSetOps which handles (SELECT ...) UNION (SELECT ...)
		subqExpr, err := ep.parseSubqueryWithSetOps()
		if err == nil && subqExpr != nil {
			// Successfully parsed a subquery (possibly with set operations)
			if subq, ok := subqExpr.(*expr.Subquery); ok {
				if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
				return &expr.InSubquery{
					Expr: left,
					Subquery: &expr.QueryExpr{
						SpanVal:   subq.Span(),
						Statement: subq.Query.Statement,
					},
					Negated: negated,
					SpanVal: mergeSpans(left.Span(), ep.parser.GetCurrentToken().Span),
				}, nil
			}
		}

		// Fall back to parsing as a regular parenthesized expression
		parsedExpr, err := ep.ParseExpr()
		if err != nil {
			// If that fails, fall back to expression list parsing
		} else {
			// Check if we got a subquery (parenthesized query)
			if subq, ok := parsedExpr.(*expr.Subquery); ok {
				if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
					return nil, err
				}
				return &expr.InSubquery{
					Expr: left,
					Subquery: &expr.QueryExpr{
						SpanVal:   subq.Span(),
						Statement: subq.Query.Statement,
					},
					Negated: negated,
					SpanVal: mergeSpans(left.Span(), ep.parser.GetCurrentToken().Span),
				}, nil
			}
			// If it's not a subquery, it might be a single expression in parentheses
			if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
				return nil, err
			}
			return &expr.InList{
				Expr:    left,
				List:    []expr.Expr{parsedExpr},
				Negated: negated,
				SpanVal: mergeSpans(left.Span(), parsedExpr.Span()),
			}, nil
		}
	}

	// Parse expression list
	list, err := ep.parseCommaSeparatedExprs()
	if err != nil {
		return nil, err
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.InList{
		Expr:    left,
		List:    list,
		Negated: negated,
		SpanVal: mergeSpans(left.Span(), list[len(list)-1].Span()),
	}, nil
}

// parseBetweenExpr parses BETWEEN low AND high
func (ep *ExpressionParser) parseBetweenExpr(left expr.Expr, negated bool) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	// Use BETWEEN precedence to ensure we stop at AND, IS, etc.
	// This matches the Rust implementation: parse_subexpr(self.dialect.prec_value(Precedence::Between))
	betweenPrec := dialect.PrecValue(parseriface.PrecedenceBetween)

	low, err := ep.ParseExprWithPrecedence(betweenPrec)
	if err != nil {
		return nil, err
	}

	if !ep.parser.ParseKeyword("AND") {
		return nil, fmt.Errorf("expected AND in BETWEEN expression")
	}

	high, err := ep.ParseExprWithPrecedence(betweenPrec)
	if err != nil {
		return nil, err
	}

	return &expr.Between{
		Expr:    left,
		Negated: negated,
		Low:     low,
		High:    high,
		SpanVal: mergeSpans(left.Span(), high.Span()),
	}, nil
}

// parseLikeExpr parses [NOT] LIKE pattern [ESCAPE char]
func (ep *ExpressionParser) parseLikeExpr(left expr.Expr, negated bool, caseInsensitive bool) (expr.Expr, error) {
	// Check for Snowflake ANY keyword
	any := ep.parser.ParseKeyword("ANY")

	pattern, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Parse optional ESCAPE clause
	var escapeChar interface{}
	if ep.parser.ParseKeyword("ESCAPE") {
		val, err := ep.parseValue()
		if err != nil {
			return nil, err
		}
		escapeChar = val
	}

	if caseInsensitive {
		return &expr.ILike{
			Negated:    negated,
			Any:        any,
			Expr:       left,
			Pattern:    pattern,
			EscapeChar: escapeChar,
			SpanVal:    mergeSpans(left.Span(), pattern.Span()),
		}, nil
	}

	return &expr.Like{
		Negated:    negated,
		Any:        any,
		Expr:       left,
		Pattern:    pattern,
		EscapeChar: escapeChar,
		SpanVal:    mergeSpans(left.Span(), pattern.Span()),
	}, nil
}

// parseSimilarToExpr parses [NOT] SIMILAR TO pattern [ESCAPE char]
func (ep *ExpressionParser) parseSimilarToExpr(left expr.Expr, negated bool) (expr.Expr, error) {
	pattern, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Parse optional ESCAPE clause
	var escapeChar interface{}
	if ep.parser.ParseKeyword("ESCAPE") {
		val, err := ep.parseValue()
		if err != nil {
			return nil, err
		}
		escapeChar = val
	}

	return &expr.SimilarTo{
		Negated:    negated,
		Expr:       left,
		Pattern:    pattern,
		EscapeChar: escapeChar,
		SpanVal:    mergeSpans(left.Span(), pattern.Span()),
	}, nil
}

// parseRLikeExpr parses [NOT] REGEXP/RLIKE pattern
func (ep *ExpressionParser) parseRLikeExpr(left expr.Expr, negated bool, isRegexp bool) (expr.Expr, error) {
	pattern, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	return &expr.RLike{
		Negated: negated,
		Expr:    left,
		Pattern: pattern,
		Regexp:  isRegexp,
		SpanVal: mergeSpans(left.Span(), pattern.Span()),
	}, nil
}

// parseDoubleColonCast parses PostgreSQL-style ::type cast
func (ep *ExpressionParser) parseDoubleColonCast(left expr.Expr) (expr.Expr, error) {
	// Parse the data type after ::
	dataType, err := ep.parser.ParseDataType()
	if err != nil {
		return nil, err
	}

	return &expr.Cast{
		Kind:     expr.CastDoubleColon,
		Expr:     left,
		DataType: dataType.String(),
		SpanVal:  left.Span(),
	}, nil
}

// parsePgOperator parses PostgreSQL custom OPERATOR(schema.op) syntax
func (ep *ExpressionParser) parsePgOperator(left expr.Expr, precedence uint8) (expr.Expr, error) {
	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse operator name components (e.g., database.pg_catalog.~)
	var opParts []string
	for {
		ep.parser.AdvanceToken()
		tok := ep.parser.GetCurrentToken()
		opParts = append(opParts, tok.Token.String())

		if !ep.parser.ConsumeToken(token.TokenPeriod{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Parse right operand
	right, err := ep.ParseExprWithPrecedence(precedence)
	if err != nil {
		return nil, err
	}

	return &expr.BinaryOp{
		Left:             left,
		Op:               operator.BOpPGCustomBinaryOperator,
		Right:            right,
		SpanVal:          mergeSpans(left.Span(), right.Span()),
		PGCustomOperator: opParts,
	}, nil
}
