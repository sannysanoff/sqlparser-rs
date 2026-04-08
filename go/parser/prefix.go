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

	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/operator"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/parseriface"
	"github.com/user/sqlparser/token"
)

// subqueryAttemptedPositions tracks positions where subquery parsing was attempted
// to prevent infinite recursion when parsing expressions like ((1))
var subqueryAttemptedPositions = make(map[int]bool)

// ResetSubqueryAttemptedPositions resets the tracking map (for testing)
func ResetSubqueryAttemptedPositions() {
	subqueryAttemptedPositions = make(map[int]bool)
}

// parsePrefix parses a prefix expression (the left-hand side of an expression).
// This includes literals, identifiers, function calls, parenthesized expressions,
// and various special expressions like CASE, CAST, etc.
func (ep *ExpressionParser) parsePrefix() (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	// TODO: The dialect system uses ast.Expr, but we use expr.Expr
	// This is an architectural mismatch that needs to be resolved.
	// For now, we skip the dialect hook and use default parsing.
	//
	// parsedExpr, handled, err := dialect.ParsePrefix(ep.parser)
	// if err != nil {
	// 	return nil, err
	// }
	// if handled {
	// 	return parsedExpr, nil
	// }

	// Try to parse as typed string literal (e.g., DATE '2020-01-01')
	// This is a PostgreSQL feature also supported by some other databases
	if typedExpr, ok := ep.tryParseTypedString(); ok {
		return typedExpr, nil
	}

	// Get the next token
	ep.parser.AdvanceToken()
	nextTok := ep.parser.GetCurrentToken()
	tokIndex := ep.parser.GetCurrentIndex()

	switch tok := nextTok.Token.(type) {
	case token.TokenWord:
		return ep.parsePrefixFromWord(tok, nextTok.Span)

	case token.TokenLBracket:
		// Array literal [1, 2, 3]
		return ep.parseArrayExpr(false)

	case token.TokenMinus:
		// Unary minus
		prec := ep.getPrecedence(parseriface.PrecedenceMulDivModOp)
		innerExpr, err := ep.ParseExprWithPrecedence(prec)
		if err != nil {
			return nil, err
		}
		return &expr.UnaryOp{
			Op:      operator.UOpMinus,
			Expr:    innerExpr,
			SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
		}, nil

	case token.TokenPlus:
		// Unary plus
		prec := ep.getPrecedence(parseriface.PrecedenceMulDivModOp)
		innerExpr, err := ep.ParseExprWithPrecedence(prec)
		if err != nil {
			return nil, err
		}
		return &expr.UnaryOp{
			Op:      operator.UOpPlus,
			Expr:    innerExpr,
			SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
		}, nil

	case token.TokenExclamationMark:
		if dialects.SupportsBangNotOperator(dialect) {
			// PostgreSQL-style bang not operator (!expr)
			prec := ep.getPrecedence(parseriface.PrecedenceUnaryNot)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpBangNot,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenDoubleExclamationMark:
		if dialect.Dialect() == "postgresql" {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpPGPrefixFactorial,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenPGSquareRoot:
		if dialect.Dialect() == "postgresql" {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpPGSquareRoot,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenPGCubeRoot:
		if dialect.Dialect() == "postgresql" {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpPGCubeRoot,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenAtSign:
		if dialect.Dialect() == "postgresql" {
			// PostgreSQL absolute value |x|
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpPGAbs,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}

		if dialects.SupportsGeometricTypes(dialect) {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpDoubleAt,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}

		// Also a placeholder prefix
		ep.parser.PrevToken()
		return ep.parseValue()

	case token.TokenAtAt:
		// Check if this should be a geometric center operator (PostgreSQL/Redshift)
		// or a MySQL-style system variable reference
		if dialects.SupportsGeometricTypes(dialect) {
			// For geometric dialects, first try to parse as geometric center operator
			// The geometric operator is followed by a geometric type keyword like 'circle', 'path', etc.
			// We peek at the next token to determine if it's a type keyword
			nextPeek := ep.parser.PeekToken()
			if word, ok := nextPeek.Token.(token.TokenWord); ok {
				kw := string(word.Word.Keyword)
				// Check if it's a geometric type keyword
				if kw == "CIRCLE" || kw == "PATH" || kw == "BOX" || kw == "POINT" ||
					kw == "POLYGON" || kw == "LINE" || kw == "LSEG" {
					prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
					innerExpr, err := ep.ParseExprWithPrecedence(prec)
					if err != nil {
						return nil, err
					}
					return &expr.UnaryOp{
						Op:      operator.UOpDoubleAt,
						Expr:    innerExpr,
						SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
					}, nil
				}
			}
		}

		// MySQL-style system variable reference: @@variable or @@global.variable
		// Parse the identifier that follows @@
		ident, err := ep.parseIdentifier()
		if err != nil {
			return nil, fmt.Errorf("expected identifier after @@: %w", err)
		}

		// Check for .var (@@global.var syntax)
		if ep.parser.ConsumeToken(token.TokenPeriod{}) {
			scope := ident.Value
			varName, err := ep.parseIdentifier()
			if err != nil {
				return nil, fmt.Errorf("expected variable name after @@%s.: %w", scope, err)
			}
			return &expr.SystemVariable{
				SpanVal: mergeSpans(nextTok.Span, varName.Span()),
				Name: &expr.CompoundIdentifier{
					SpanVal: mergeSpans(ident.Span(), varName.Span()),
					Idents:  []*expr.Ident{ident, varName},
				},
			}, nil
		}

		// Simple @@var syntax
		return &expr.SystemVariable{
			SpanVal: mergeSpans(nextTok.Span, ident.Span()),
			Name: &expr.CompoundIdentifier{
				SpanVal: ident.Span(),
				Idents:  []*expr.Ident{ident},
			},
		}, nil

	case token.TokenTilde:
		// Bitwise NOT
		prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
		innerExpr, err := ep.ParseExprWithPrecedence(prec)
		if err != nil {
			return nil, err
		}
		return &expr.UnaryOp{
			Op:      operator.UOpBitwiseNot,
			Expr:    innerExpr,
			SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
		}, nil

	case token.TokenSharp:
		if dialects.SupportsGeometricTypes(dialect) {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpHash,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenAtDashAt:
		if dialects.SupportsGeometricTypes(dialect) {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpAtDashAt,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenQuestionMarkDash:
		if dialects.SupportsGeometricTypes(dialect) {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpQuestionDash,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenQuestionPipe:
		if dialects.SupportsGeometricTypes(dialect) {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.UnaryOp{
				Op:      operator.UOpQuestionPipe,
				Expr:    innerExpr,
				SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
			}, nil
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenEscapedStringLiteral:
		if dialect.Dialect() == "postgresql" {
			ep.parser.PrevToken()
			return ep.parseValue()
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case token.TokenUnicodeStringLiteral:
		ep.parser.PrevToken()
		return ep.parseValue()

	case token.TokenNumber, token.TokenSingleQuotedString,
		token.TokenDoubleQuotedString, token.TokenTripleSingleQuotedString,
		token.TokenTripleDoubleQuotedString, token.TokenDollarQuotedString,
		token.TokenSingleQuotedByteStringLiteral, token.TokenDoubleQuotedByteStringLiteral,
		token.TokenTripleSingleQuotedByteStringLiteral, token.TokenTripleDoubleQuotedByteStringLiteral,
		token.TokenSingleQuotedRawStringLiteral, token.TokenDoubleQuotedRawStringLiteral,
		token.TokenTripleSingleQuotedRawStringLiteral, token.TokenTripleDoubleQuotedRawStringLiteral,
		token.TokenNationalStringLiteral, token.TokenQuoteDelimitedStringLiteral,
		token.TokenNationalQuoteDelimitedStringLiteral, token.TokenHexStringLiteral:
		ep.parser.PrevToken()
		return ep.parseValue()

	case token.TokenLParen:
		return ep.parseParenthesizedPrefix()

	case token.TokenPlaceholder, token.TokenColon:
		ep.parser.PrevToken()
		return ep.parseValue()

	case token.TokenLBrace:
		return ep.parseLBraceExpr()

	default:
		return nil, ep.parser.ExpectedAt("an expression", tokIndex)
	}
}

// parsePrefixFromWord parses a prefix expression starting with a word token
func (ep *ExpressionParser) parsePrefixFromWord(word token.TokenWord, spanVal token.Span) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	// Save position before attempting special expression parsing
	// This is needed for fallback to identifier if special parsing fails
	// Pattern E251: GetCurrentIndex() returns p.index-1, SetCurrentIndex() sets p.index
	// After AdvanceToken(), use SetCurrentIndex(savedIdx+1) to restore correctly
	savedIdx := ep.parser.GetCurrentIndex()

	// Check if it's a reserved word with special meaning
	result, err := ep.tryParseReservedWordPrefix(&word, spanVal)
	if err != nil {
		// If parsing as special expression failed, check if the keyword is reserved
		// If NOT reserved, try to parse it as an identifier instead
		// This handles cases like "SELECT MAX(interval) FROM tbl" in PostgreSQL/Snowflake
		// where INTERVAL is not reserved and can be used as an identifier
		if !dialect.IsReservedForIdentifier(word.Word.Keyword) {
			// Restore position to before special parsing was attempted
			// Use savedIdx+1 because SetCurrentIndex sets p.index directly (Pattern E251)
			ep.parser.SetCurrentIndex(savedIdx + 1)
			// Try parsing as an unreserved word (identifier)
			identResult, identErr := ep.parseUnreservedWordPrefix(&word, spanVal)
			if identErr == nil {
				return identResult, nil
			}
			// Restore position and return original error
			ep.parser.SetCurrentIndex(savedIdx + 1)
		}
		return nil, err
	}
	if result != nil {
		return result, nil
	}

	// Not a reserved word, parse as identifier or function
	return ep.parseUnreservedWordPrefix(&word, spanVal)
}

// tryParseReservedWordPrefix tries to parse a reserved word as a special expression prefix
func (ep *ExpressionParser) tryParseReservedWordPrefix(word *token.TokenWord, span token.Span) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	switch word.Word.Keyword {
	case "TRUE", "FALSE":
		if dialects.SupportsBooleanLiterals(dialect) {
			ep.parser.PrevToken()
			return ep.parseValue()
		}

	case "NULL":
		ep.parser.PrevToken()
		return ep.parseValue()

	case "CURRENT_CATALOG", "CURRENT_USER", "SESSION_USER", "USER":
		if dialect.Dialect() == "postgresql" {
			// These are treated as functions in PostgreSQL
			return &expr.FunctionExpr{
				Name: &expr.ObjectName{
					SpanVal: span,
					Parts:   []*expr.ObjectNamePart{{SpanVal: span, Ident: ep.wordToIdent(word, span)}},
				},
				UsesOdbcSyntax: false,
				Args:           &expr.FunctionArguments{None: true},
				SpanVal:        span,
			}, nil
		}

	case "CURRENT_TIMESTAMP", "CURRENT_TIME", "CURRENT_DATE", "LOCALTIME", "LOCALTIMESTAMP":
		return ep.parseTimeFunction(word, span)

	case "CASE":
		return ep.parseCaseExpr()

	case "CONVERT":
		return ep.parseConvertExpr(false)

	case "TRY_CONVERT":
		if dialects.SupportsTryConvert(dialect) {
			return ep.parseConvertExpr(true)
		}

	case "CAST":
		return ep.parseCastExpr(expr.CastStandard)

	case "TRY_CAST":
		return ep.parseCastExpr(expr.CastTry)

	case "SAFE_CAST":
		return ep.parseCastExpr(expr.CastSafe)

	case "EXISTS":
		return ep.parseExistsExpr(false)

	case "EXTRACT":
		return ep.parseExtractExpr()

	case "CEIL":
		return ep.parseCeilFloorExpr(true)

	case "FLOOR":
		return ep.parseCeilFloorExpr(false)

	case "POSITION":
		nextTok := ep.parser.PeekTokenRef()
		if _, ok := nextTok.Token.(token.TokenLParen); ok {
			return ep.parsePositionExpr(*word, span)
		}

	case "SUBSTR", "SUBSTRING":
		ep.parser.PrevToken()
		return ep.parseSubstringExpr()

	case "OVERLAY":
		return ep.parseOverlayExpr()

	case "TRIM":
		return ep.parseTrimExpr()

	case "INTERVAL":
		return ep.parseIntervalExpr()

	case "ARRAY":
		nextTok := ep.parser.PeekTokenRef()
		if _, ok := nextTok.Token.(token.TokenLBracket); ok {
			// ARRAY[...] syntax
			if _, err := ep.parser.ExpectToken(token.TokenLBracket{}); err != nil {
				return nil, err
			}
			return ep.parseArrayExpr(true)
		}
		// ARRAY(...) subquery or function call
		if _, ok := nextTok.Token.(token.TokenLParen); ok {
			// Save position to backtrack if it's not a subquery
			restore := ep.parser.SavePosition()
			if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}

			// Check if it's a subquery (starts with SELECT, WITH, VALUES, TABLE)
			subqueryKeywords := []string{"SELECT", "WITH", "VALUES", "TABLE"}
			nextTok := ep.parser.PeekTokenRef()
			wordPtr, ok := nextTok.Token.(*token.TokenWord)
			if !ok {
				// Not a word token, try value type assertion
				if wordVal, ok := nextTok.Token.(token.TokenWord); ok {
					wordPtr = &wordVal
				}
			}
			if wordPtr != nil {
				kw := string(wordPtr.Word.Keyword)
				isSubquery := false
				for _, skw := range subqueryKeywords {
					if kw == skw {
						isSubquery = true
						break
					}
				}

				if isSubquery {
					// Parse subquery
					// For now, return a placeholder (TODO: proper subquery parsing)
					return &expr.FunctionExpr{
						Name: &expr.ObjectName{
							SpanVal: span,
							Parts:   []*expr.ObjectNamePart{{SpanVal: span, Ident: ep.wordToIdent(wordPtr, span)}},
						},
						UsesOdbcSyntax: false,
						Args:           &expr.FunctionArguments{None: true},
						SpanVal:        span,
					}, nil
				}
			}

			// Not a subquery - restore and treat as regular function call
			restore()
			return ep.parseFunction(&expr.ObjectName{
				SpanVal: span,
				Parts:   []*expr.ObjectNamePart{{SpanVal: span, Ident: ep.wordToIdent(word, span)}},
			})
		}

	case "NOT":
		return ep.parseNotExpr()

	case "MATCH":
		if dialects.SupportsMatchAgainst(dialect) {
			return ep.parseMatchAgainstExpr()
		}

	case "STRUCT":
		if dialects.SupportsStructLiteral(dialect) {
			return ep.parseStructLiteral()
		}

	case "PRIOR":
		if ep.parser.GetState() == dialects.StateConnectBy {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.PriorExpr{
				SpanVal: mergeSpans(span, innerExpr.Span()),
				Expr:    innerExpr,
			}, nil
		}

	case "CONNECT_BY_ROOT":
		if dialects.SupportsConnectBy(dialect) {
			prec := ep.getPrecedence(parseriface.PrecedencePlusMinus)
			innerExpr, err := ep.ParseExprWithPrecedence(prec)
			if err != nil {
				return nil, err
			}
			return &expr.ConnectByRootExpr{
				SpanVal: mergeSpans(span, innerExpr.Span()),
				Expr:    innerExpr,
			}, nil
		}

	case "MAP":
		nextTok := ep.parser.PeekTokenRef()
		if _, ok := nextTok.Token.(token.TokenLBrace); ok && dialects.SupportsMapLiteralSyntax(dialect) {
			return ep.parseMapLiteral()
		}

	case "LAMBDA":
		if dialects.SupportsLambdaFunctions(dialect) {
			return ep.parseLambdaExpr()
		}

	case "CIRCLE", "BOX", "PATH", "LINE", "LSEG", "POINT", "POLYGON":
		if dialects.SupportsGeometricTypes(dialect) {
			return ep.parseGeometricType(string(word.Word.Keyword))
		}

	case "GROUPING":
		// GROUPING SETS (...) expression for GROUP BY
		nextTok := ep.parser.PeekTokenRef()
		if wordNext, ok := nextTok.Token.(token.TokenWord); ok && wordNext.Word.Keyword == "SETS" {
			ep.parser.AdvanceToken() // consume SETS
			return ep.parseGroupingSetsExpr(span)
		}

	case "CUBE":
		// CUBE (...) expression for GROUP BY
		nextTok := ep.parser.PeekTokenRef()
		if _, ok := nextTok.Token.(token.TokenLParen); ok {
			return ep.parseCubeExpr(span)
		}

	case "ROLLUP":
		// ROLLUP (...) expression for GROUP BY
		nextTok := ep.parser.PeekTokenRef()
		if _, ok := nextTok.Token.(token.TokenLParen); ok {
			return ep.parseRollupExpr(span)
		}
	}

	return nil, nil
}

// parseUnreservedWordPrefix parses an unreserved word as identifier or function
func (ep *ExpressionParser) parseUnreservedWordPrefix(word *token.TokenWord, span token.Span) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()
	ident := ep.wordToIdent(word, span)

	// Check for outer join operator (Oracle style)
	nextTok := ep.parser.PeekTokenRef()
	isOuterJoin := ep.peekOuterJoinOperator()

	if !isOuterJoin {
		if _, ok := nextTok.Token.(token.TokenLParen); ok {
			// Function call
			return ep.parseFunction(&expr.ObjectName{
				SpanVal: span,
				Parts:   []*expr.ObjectNamePart{{SpanVal: span, Ident: ident}},
			})
		}
	}

	// Check for MySQL string introducer (e.g., _utf8'string')
	if strings.HasPrefix(word.Word.Value, "_") {
		next := ep.parser.PeekTokenRef()
		switch next.Token.(type) {
		case token.TokenSingleQuotedString, token.TokenDoubleQuotedString,
			token.TokenHexStringLiteral:
			return &expr.Prefixed{
				SpanVal: mergeSpans(span, next.Span),
				Prefix:  ident,
				Value:   ep.parseIntroducedString(),
			}, nil
		}
	}

	// Check for lambda expression (single parameter)
	if dialects.SupportsLambdaFunctions(dialect) {
		next := ep.parser.PeekTokenRef()
		// Check for optional type annotation (e.g., "a INT -> expr")
		param := expr.LambdaFunctionParameter{
			SpanVal: span,
			Name:    ident,
		}

		// Peek ahead to see if we have [type] -> pattern
		savePos := ep.parser.SavePosition()
		hasType := false
		if word, ok := next.Token.(token.TokenWord); ok {
			nextKw := string(word.Word.Keyword)
			// If next is a word that's not an arrow, it might be a type
			if nextKw != "->" {
				typeIdent, err := ep.parseIdentifier()
				if err == nil {
					// Check if the token after the type is ->
					afterType := ep.parser.PeekTokenRef()
					if _, isArrow := afterType.Token.(token.TokenArrow); isArrow {
						hasType = true
						param.DataType = typeIdent.Value
					}
				}
			}
		}
		if !hasType {
			savePos()
		}

		next = ep.parser.PeekTokenRef()
		if _, ok := next.Token.(token.TokenArrow); ok {
			ep.parser.AdvanceToken() // consume ->
			body, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			return &expr.LambdaExpr{
				SpanVal: mergeSpans(span, body.Span()),
				Params:  []expr.LambdaFunctionParameter{param},
				Body:    body,
				Syntax:  expr.LambdaArrow,
			}, nil
		}
	}

	// Simple identifier
	return &expr.Identifier{
		SpanVal: span,
		Ident:   ident,
	}, nil
}

// parseIdentifier parses a simple identifier
func (ep *ExpressionParser) parseIdentifier() (*expr.Ident, error) {
	tok := ep.parser.NextToken()

	switch t := tok.Token.(type) {
	case token.TokenWord:
		return ep.wordToIdent(&t, tok.Span), nil
	case token.TokenSingleQuotedString:
		// Single-quoted strings can be used as identifiers (e.g., collation names)
		singleQuote := rune('\'')
		return &expr.Ident{
			SpanVal:    tok.Span,
			Value:      t.Value,
			QuoteStyle: &singleQuote,
		}, nil
	case token.TokenDoubleQuotedString:
		// Double-quoted strings can be used as identifiers
		doubleQuote := rune('"')
		return &expr.Ident{
			SpanVal:    tok.Span,
			Value:      t.Value,
			QuoteStyle: &doubleQuote,
		}, nil
	}

	return nil, ep.parser.Expected("an identifier", tok)
}

// parseObjectName parses an object name (potentially qualified), e.g.,
// `foo` or `myschema."table"` or `db.schema.table`
//
// Reference: src/parser/mod.rs:12715 - parse_object_name
func (ep *ExpressionParser) parseObjectName() (*expr.ObjectName, error) {
	var parts []*expr.ObjectNamePart

	for {
		ident, err := ep.parseIdentifier()
		if err != nil {
			return nil, err
		}
		parts = append(parts, &expr.ObjectNamePart{
			SpanVal: ident.Span(),
			Ident:   ident,
		})

		// Check for dot (period) to continue parsing more parts
		if !ep.parser.ConsumeToken(token.TokenPeriod{}) {
			break
		}
	}

	if len(parts) == 0 {
		return nil, fmt.Errorf("expected identifier")
	}

	return &expr.ObjectName{
		SpanVal: parts[0].SpanVal,
		Parts:   parts,
	}, nil
}

// parseCommaSeparatedIdents parses comma-separated identifiers
func (ep *ExpressionParser) parseCommaSeparatedIdents() ([]*expr.Ident, error) {
	var idents []*expr.Ident

	for {
		ident, err := ep.parseIdentifier()
		if err != nil {
			return nil, err
		}
		idents = append(idents, ident)

		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	return idents, nil
}

// wordToIdent converts a TokenWord to an Ident
func (ep *ExpressionParser) wordToIdent(word *token.TokenWord, spanVal token.Span) *expr.Ident {
	// Preserve the original value - no dialect-specific normalization
	// This matches the Rust reference implementation behavior
	value := word.Word.Value

	ident := &expr.Ident{
		SpanVal: spanVal,
		Value:   value,
	}
	if word.Word.QuoteStyle != nil {
		q := rune(*word.Word.QuoteStyle)
		ident.QuoteStyle = &q
	}
	return ident
}

// parseValue parses a literal value
func (ep *ExpressionParser) parseValue() (expr.Expr, error) {
	tok := ep.parser.NextToken()

	switch t := tok.Token.(type) {
	case token.TokenNumber:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenSingleQuotedString:
		concatenatedValue, endSpan := ep.maybeConcatStringLiteral(t.Value, tok.Span)
		return &expr.ValueExpr{
			SpanVal: mergeSpans(tok.Span, endSpan),
			Value:   token.TokenSingleQuotedString{Value: concatenatedValue},
		}, nil

	case token.TokenDoubleQuotedString:
		concatenatedValue, endSpan := ep.maybeConcatStringLiteral(t.Value, tok.Span)
		return &expr.ValueExpr{
			SpanVal: mergeSpans(tok.Span, endSpan),
			Value:   token.TokenDoubleQuotedString{Value: concatenatedValue},
		}, nil

	case token.TokenNationalStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenHexStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenSingleQuotedByteStringLiteral, token.TokenDoubleQuotedByteStringLiteral,
		token.TokenTripleSingleQuotedByteStringLiteral, token.TokenTripleDoubleQuotedByteStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenEscapedStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenUnicodeStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenDollarQuotedString:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenPlaceholder:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenWord:
		switch t.Word.Keyword {
		case "TRUE":
			return &expr.ValueExpr{
				SpanVal: tok.Span,
				Value:   true,
			}, nil
		case "FALSE":
			return &expr.ValueExpr{
				SpanVal: tok.Span,
				Value:   false,
			}, nil
		case "NULL":
			return &expr.ValueExpr{
				SpanVal: tok.Span,
				Value:   nil,
			}, nil
		}

	case token.TokenColon, token.TokenAtSign:
		// Named or positional placeholder (e.g., :1, :name, @1, @name)
		// Reference: src/parser/mod.rs:11660-11676
		// Get the next token without skipping - it should be a word or number
		nextTok := ep.parser.NextToken()
		prefix := ":"
		if _, ok := t.(token.TokenAtSign); ok {
			prefix = "@"
		}
		var placeholderValue string
		switch n := nextTok.Token.(type) {
		case token.TokenWord:
			placeholderValue = prefix + n.Word.Value
		case token.TokenNumber:
			placeholderValue = prefix + n.Value
		default:
			return nil, ep.parser.Expected("placeholder identifier or number", nextTok)
		}
		return &expr.ValueExpr{
			SpanVal: mergeSpans(tok.Span, nextTok.Span),
			Value:   token.TokenPlaceholder{Value: placeholderValue},
		}, nil
	}

	return nil, ep.parser.Expected("a value", tok)
}

// maybeConcatStringLiteral concatenates adjacent string literals if the dialect supports it.
// This handles both simple concatenation (e.g., 'a' 'b' -> 'ab') and newline-based
// concatenation for dialects that support it.
// Returns the concatenated string value and the span of the last token consumed.
func (ep *ExpressionParser) maybeConcatStringLiteral(initial string, initialSpan token.Span) (string, token.Span) {
	dialect := ep.parser.GetDialect()
	result := initial
	endSpan := initialSpan

	if dialects.SupportsStringLiteralConcatenation(dialect) {
		// Simple adjacent string concatenation: 'a' 'b' -> 'ab'
		for {
			nextTok := ep.parser.PeekTokenRef()
			switch t := nextTok.Token.(type) {
			case token.TokenSingleQuotedString:
				result += t.Value
				endSpan = nextTok.Span
				ep.parser.AdvanceToken() // consume the string
			case token.TokenDoubleQuotedString:
				result += t.Value
				endSpan = nextTok.Span
				ep.parser.AdvanceToken() // consume the string
			default:
				return result, endSpan
			}
		}
	} else if dialects.SupportsStringLiteralConcatenationWithNewline(dialect) {
		// Newline-based concatenation (e.g., Redshift):
		// 'a'
		// 'b' -> 'ab'
		afterNewline := false
		for {
			// Use PeekTokenNoSkip to see whitespace tokens (including newlines)
			nextTok := ep.parser.PeekTokenNoSkip()
			switch t := nextTok.Token.(type) {
			case token.TokenWhitespace:
				if t.Whitespace.Type == token.Newline {
					afterNewline = true
				}
				// Use NextTokenNoSkip to consume whitespace
				ep.parser.NextTokenNoSkip()
			case token.TokenSingleQuotedString:
				if afterNewline {
					result += t.Value
					endSpan = nextTok.Span
					ep.parser.NextTokenNoSkip() // consume the string
					afterNewline = false
					// After concatenating, continue the loop to check for more newlines + strings
				} else {
					return result, endSpan
				}
			case token.TokenDoubleQuotedString:
				if afterNewline {
					result += t.Value
					endSpan = nextTok.Span
					ep.parser.NextTokenNoSkip() // consume the string
					afterNewline = false
					// After concatenating, continue the loop to check for more newlines + strings
				} else {
					return result, endSpan
				}
			default:
				return result, endSpan
			}
		}
	}

	return result, endSpan
}

// parseLiteralString parses a string literal
func (ep *ExpressionParser) parseLiteralString() (string, error) {
	tok := ep.parser.NextToken()

	switch t := tok.Token.(type) {
	case token.TokenSingleQuotedString:
		return t.Value, nil
	case token.TokenDoubleQuotedString:
		return t.Value, nil
	}

	return "", ep.parser.Expected("a string literal", tok)
}

// parseParenthesizedPrefix parses a parenthesized expression or tuple
func (ep *ExpressionParser) parseParenthesizedPrefix() (expr.Expr, error) {
	currentIdx := ep.parser.GetCurrentIndex()

	// Check for subquery first
	if ep.peekSubquery() {
		return ep.parseSubqueryExpr()
	}

	// Check for double-parenthesized subquery with set operations
	// e.g., ((SELECT ...) UNION (SELECT ...))
	// But only if we haven't already tried at this position
	if _, ok := ep.parser.PeekToken().Token.(token.TokenLParen); ok {
		if !subqueryAttemptedPositions[currentIdx] {
			subqueryAttemptedPositions[currentIdx] = true
			subq, err := ep.parseSubqueryWithSetOps()
			if err == nil && subq != nil {
				return subq, nil
			}
		}
	}

	// Check for lambda expression
	if lambda, ok := ep.tryParseLambda(); ok {
		return lambda, nil
	}

	// Regular parenthesized expression or tuple
	// Save the current state to handle state switching properly
	oldState := ep.parser.GetState()
	ep.parser.SetState(dialects.StateNormal)

	exprs, err := ep.parseCommaSeparatedExprs()
	if err != nil {
		ep.parser.SetState(oldState)
		return nil, err
	}

	ep.parser.SetState(oldState)

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	if len(exprs) == 0 {
		return nil, fmt.Errorf("empty expression list in parentheses")
	}

	if len(exprs) == 1 {
		// Single expression in parentheses
		return &expr.Nested{
			SpanVal: exprs[0].Span(),
			Expr:    exprs[0],
		}, nil
	}

	// Multiple expressions form a tuple
	return &expr.TupleExpr{
		SpanVal: mergeSpans(exprs[0].Span(), exprs[len(exprs)-1].Span()),
		Exprs:   exprs,
	}, nil
}

// tryParseTypedString tries to parse a typed string literal like DATE '2020-01-01'
// Returns the expression and true if successful, nil and false otherwise.
func (ep *ExpressionParser) tryParseTypedString() (expr.Expr, bool) {
	// Look ahead: we need a data type keyword followed by a string literal
	// Data types that can be used with typed string syntax: DATE, TIME, TIMESTAMP,
	// INTERVAL, and some others

	peekTok := ep.parser.PeekTokenRef()
	word, ok := peekTok.Token.(token.TokenWord)
	if !ok {
		return nil, false
	}

	// Check if it's a data type keyword
	dataTypeKeyword := word.Word.Keyword
	dataTypeName := word.Word.Value // Preserve original case for output
	isDataType := false
	switch dataTypeKeyword {
	case "DATE", "TIME", "TIMESTAMP", "TIMESTAMPTZ", "INTERVAL", "DATETIME",
		"DECIMAL", "NUMERIC", "BIGNUMERIC", "CHAR", "VARCHAR", "NCHAR", "NVARCHAR",
		"CHARACTER", "BINARY", "VARBINARY", "JSON":
		isDataType = true
	case "POINT", "LINE", "LSEG", "BOX", "PATH", "POLYGON", "CIRCLE":
		// Geometric types (PostgreSQL)
		if dialects.SupportsGeometricTypes(ep.parser.GetDialect()) {
			isDataType = true
		}
	}

	if !isDataType {
		return nil, false
	}

	// Try to parse: consume the keyword, check if next is a string literal
	restore := ep.parser.SavePosition()
	ep.parser.AdvanceToken() // consume the keyword

	nextTok := ep.parser.PeekTokenRef()
	_, isString := nextTok.Token.(token.TokenSingleQuotedString)
	if !isString {
		// Check for BINARY as cast: if dialect supports it and data type is BINARY,
		// parse an expression and create a Cast
		if dataTypeName == "BINARY" && dialects.SupportsBinaryKwAsCast(ep.parser.GetDialect()) {
			// Parse the expression after BINARY keyword
			innerExpr, err := ep.ParseExpr()
			if err != nil {
				restore()
				return nil, false
			}
			// Create a Cast expression: CAST(expr AS BINARY)
			return &expr.Cast{
				SpanVal:  mergeSpans(peekTok.Span, innerExpr.Span()),
				Kind:     expr.CastStandard,
				Expr:     innerExpr,
				DataType: "BINARY",
			}, true
		}
		restore()
		return nil, false
	}

	// Parse the string value
	ep.parser.AdvanceToken() // consume the string
	stringTok := ep.parser.GetCurrentToken()
	strVal, _ := stringTok.Token.(token.TokenSingleQuotedString)

	// Special handling for INTERVAL - it has special syntax: INTERVAL 'value' unit [(precision)]
	if strings.EqualFold(dataTypeName, "INTERVAL") {
		// Check for temporal unit (DAY, MONTH, YEAR, etc.)
		nextTok := ep.parser.PeekTokenRef()
		if word, ok := nextTok.Token.(token.TokenWord); ok {
			unit := word.Word.Keyword
			switch unit {
			case "YEAR", "YEARS", "MONTH", "MONTHS", "WEEK", "WEEKS",
				"DAY", "DAYS", "HOUR", "HOURS", "MINUTE", "MINUTES",
				"SECOND", "SECONDS":
				ep.parser.AdvanceToken() // consume the unit
				unitStr := string(unit)

				// Parse optional precision for the leading unit
				var leadingPrecision, fracPrecision *uint64
				nextTok2 := ep.parser.PeekTokenRef()
				if _, ok := nextTok2.Token.(token.TokenLParen); ok {
					ep.parser.AdvanceToken() // consume (

					// Parse the first number (leading precision)
					precTok := ep.parser.NextToken()
					if numTok, ok := precTok.Token.(token.TokenNumber); ok {
						n, _ := strconv.ParseUint(numTok.Value, 10, 64)
						leadingPrecision = &n
					}

					// Check for comma and second number (fractional seconds precision for SECOND)
					nextTok3 := ep.parser.PeekTokenRef()
					if _, ok := nextTok3.Token.(token.TokenComma); ok {
						ep.parser.AdvanceToken() // consume ,
						fracTok := ep.parser.NextToken()
						if numTok, ok := fracTok.Token.(token.TokenNumber); ok {
							n, _ := strconv.ParseUint(numTok.Value, 10, 64)
							fracPrecision = &n
						}
					}

					// Consume ) if present
					ep.parser.ConsumeToken(token.TokenRParen{})
				}

				// Special handling for SECOND: when leading field is SECOND, we don't allow TO clause
				// SQL mandates special format: SECOND [( <leading precision> [ , <fractional seconds precision>] )]
				if unit == "SECOND" || unit == "SECONDS" {
					// Check if next token is TO - this is an error for SECOND
					nextTokCheck := ep.parser.PeekTokenRef()
					if wordCheck, ok := nextTokCheck.Token.(token.TokenWord); ok && wordCheck.Word.Keyword == "TO" {
						// SECOND TO SECOND is not allowed - let normal parsing handle the error
						restore()
						return nil, false
					}

					return &expr.IntervalExpr{
						SpanVal: mergeSpans(peekTok.Span, ep.parser.GetCurrentToken().Span),
						Value: &expr.ValueExpr{
							SpanVal: stringTok.Span,
							Value:   strVal,
						},
						LeadingField:               &unitStr,
						LeadingPrecision:           leadingPrecision,
						FractionalSecondsPrecision: fracPrecision,
					}, true
				}

				// Check for TO clause (e.g., YEAR TO MONTH) - not applicable for SECOND
				nextTok3 := ep.parser.PeekTokenRef()
				if word3, ok := nextTok3.Token.(token.TokenWord); ok && word3.Word.Keyword == "TO" {
					ep.parser.AdvanceToken() // consume TO
					nextTok4 := ep.parser.PeekTokenRef()
					if word4, ok := nextTok4.Token.(token.TokenWord); ok {
						lastUnit := word4.Word.Keyword
						switch lastUnit {
						case "YEAR", "YEARS", "MONTH", "MONTHS", "WEEK", "WEEKS",
							"DAY", "DAYS", "HOUR", "HOURS", "MINUTE", "MINUTES",
							"SECOND", "SECONDS":
							ep.parser.AdvanceToken() // consume last unit
							lastUnitStr := string(lastUnit)

							// Check for optional precision on last unit
							// Only SECOND can have precision on the last field
							// Also, having precision on both leading and last (non-SECOND) field is an error
							nextTok5 := ep.parser.PeekTokenRef()
							hasLastPrecision := false
							if _, ok := nextTok5.Token.(token.TokenLParen); ok {
								// If leading field already has precision and last field is not SECOND, error
								if leadingPrecision != nil && lastUnit != "SECOND" && lastUnit != "SECONDS" {
									// Let normal parsing handle the error
									restore()
									return nil, false
								}

								ep.parser.AdvanceToken() // consume (
								lastPrecTok := ep.parser.NextToken()
								if numTok, ok := lastPrecTok.Token.(token.TokenNumber); ok {
									n, _ := strconv.ParseUint(numTok.Value, 10, 64)
									fracPrecision = &n
									hasLastPrecision = true
								}
								// Consume ) if present
								ep.parser.ConsumeToken(token.TokenRParen{})
							}

							// If we have precision on both sides for non-SECOND, let normal parsing error
							if leadingPrecision != nil && hasLastPrecision &&
								lastUnit != "SECOND" && lastUnit != "SECONDS" {
								restore()
								return nil, false
							}

							return &expr.IntervalExpr{
								SpanVal: mergeSpans(peekTok.Span, ep.parser.GetCurrentToken().Span),
								Value: &expr.ValueExpr{
									SpanVal: stringTok.Span,
									Value:   strVal,
								},
								LeadingField:               &unitStr,
								LeadingPrecision:           leadingPrecision,
								LastField:                  &lastUnitStr,
								FractionalSecondsPrecision: fracPrecision,
							}, true
						}
					}
				}

				return &expr.IntervalExpr{
					SpanVal: mergeSpans(peekTok.Span, ep.parser.GetCurrentToken().Span),
					Value: &expr.ValueExpr{
						SpanVal: stringTok.Span,
						Value:   strVal,
					},
					LeadingField:               &unitStr,
					LeadingPrecision:           leadingPrecision,
					FractionalSecondsPrecision: fracPrecision,
				}, true
			}
		}

		// INTERVAL without unit - for dialects that require qualifiers,
		// let the normal INTERVAL parsing handle this (which will error)
		if dialects.RequireIntervalQualifier(ep.parser.GetDialect()) {
			restore()
			return nil, false
		}

		// INTERVAL without unit - return basic interval
		return &expr.IntervalExpr{
			SpanVal: mergeSpans(peekTok.Span, stringTok.Span),
			Value: &expr.ValueExpr{
				SpanVal: stringTok.Span,
				Value:   strVal,
			},
		}, true
	}

	// Create a TypedString expression
	return &expr.TypedString{
		SpanVal:  mergeSpans(peekTok.Span, stringTok.Span),
		DataType: string(dataTypeName),
		Value:    strVal.Value,
	}, true
}

// peekOuterJoinOperator checks if the next tokens indicate Oracle outer join operator
func (ep *ExpressionParser) peekOuterJoinOperator() bool {
	dialect := ep.parser.GetDialect()
	if !dialects.SupportsOuterJoinOperator(dialect) {
		return false
	}

	toks := []token.TokenWithSpan{
		ep.parser.PeekNthToken(0),
		ep.parser.PeekNthToken(1),
		ep.parser.PeekNthToken(2),
	}

	_, isLParen := toks[0].Token.(token.TokenLParen)
	_, isPlus := toks[1].Token.(token.TokenPlus)
	_, isRParen := toks[2].Token.(token.TokenRParen)

	return isLParen && isPlus && isRParen
}

// peekSubquery checks if the next tokens indicate a subquery
func (ep *ExpressionParser) peekSubquery() bool {
	next := ep.parser.PeekTokenRef()
	if word, ok := next.Token.(token.TokenWord); ok {
		return word.Word.Keyword == "SELECT" || word.Word.Keyword == "WITH"
	}
	return false
}

// tryParseLambda attempts to parse a lambda expression
// Lambda expressions have the form: (param1, param2, ...) -> expr
// This is called when we've already seen the opening '('
func (ep *ExpressionParser) tryParseLambda() (expr.Expr, bool) {
	dialect := ep.parser.GetDialect()
	if !dialects.SupportsLambdaFunctions(dialect) {
		return nil, false
	}

	// We need to check if this looks like a lambda without consuming tokens.
	// Lambda pattern: ( [ident [, ident]*] ) -> expr
	// We peek ahead to see if there's an arrow after the closing paren.

	// Save position for backtrack
	restore := ep.parser.SavePosition()
	spanStart := ep.parser.GetCurrentToken().Span

	// Advance past the '(' we already consumed in the caller
	// (Note: the caller already consumed '(' before calling this function)

	// Try to parse comma-separated identifiers
	foundIdentifiers := false
	for {
		tok := ep.parser.PeekTokenRef()

		// Check for closing paren - end of parameter list
		if _, isRParen := tok.Token.(token.TokenRParen); isRParen {
			if !foundIdentifiers {
				// Empty parens () - not a lambda
				restore()
				return nil, false
			}
			break
		}

		// Expect an identifier
		_, err := ep.parseIdentifier()
		if err != nil {
			// Not an identifier - not a lambda
			restore()
			return nil, false
		}
		foundIdentifiers = true

		// Skip optional type annotation (e.g., "x INT")
		nextTok := ep.parser.PeekTokenRef()
		if word, ok := nextTok.Token.(token.TokenWord); ok {
			nextKw := string(word.Word.Keyword)
			// If next token is a word that's not a reserved keyword, it might be a type
			if nextKw != "->" && nextKw != "," && nextKw != ")" {
				// Try to consume it as a type (but don't fail if it's not)
				_, _ = ep.parseIdentifier()
			}
		}

		// Check for comma or closing paren
		tok = ep.parser.PeekTokenRef()
		if _, isComma := tok.Token.(token.TokenComma); isComma {
			ep.parser.AdvanceToken() // consume ,
			continue
		}
		if _, isRParen := tok.Token.(token.TokenRParen); isRParen {
			break
		}

		// Unexpected token - not a lambda
		restore()
		return nil, false
	}

	// Expect closing paren
	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		restore()
		return nil, false
	}

	// Check for arrow token ->
	nextTok := ep.parser.PeekTokenRef()
	if _, isArrow := nextTok.Token.(token.TokenArrow); !isArrow {
		// No arrow - not a lambda, restore and return false
		restore()
		return nil, false
	}

	// It's a lambda! Now consume the arrow and parse the body
	ep.parser.AdvanceToken() // consume ->

	// Parse the body expression
	body, err := ep.ParseExpr()
	if err != nil {
		restore()
		return nil, false
	}

	// Build lambda expression with collected identifiers
	// Note: We need to re-parse to get the actual identifier details
	// For now, restore and re-parse properly
	restore()

	// Now do the actual parse
	// Note: After restore, we're at the position after '(', so don't consume it again
	// The caller already consumed '(', and restore() puts us back to that position (after '(')

	var params []expr.LambdaFunctionParameter
	for {
		tok := ep.parser.PeekTokenRef()
		if _, isRParen := tok.Token.(token.TokenRParen); isRParen {
			break
		}

		ident, err := ep.parseIdentifier()
		if err != nil {
			return nil, false
		}

		param := expr.LambdaFunctionParameter{
			SpanVal: ident.Span(),
			Name:    ident,
		}

		// Check for optional type
		nextTok := ep.parser.PeekTokenRef()
		if word, ok := nextTok.Token.(token.TokenWord); ok {
			nextKw := string(word.Word.Keyword)
			if nextKw != "->" && nextKw != "," && nextKw != ")" {
				if typeIdent, err := ep.parseIdentifier(); err == nil {
					param.DataType = typeIdent.Value
				}
			}
		}

		params = append(params, param)

		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, false
	}

	if _, err := ep.parser.ExpectToken(token.TokenArrow{}); err != nil {
		return nil, false
	}

	body, err = ep.ParseExpr()
	if err != nil {
		return nil, false
	}

	return &expr.LambdaExpr{
		SpanVal: mergeSpans(spanStart, body.Span()),
		Params:  params,
		Body:    body,
		Syntax:  expr.LambdaArrow,
	}, true
}

// parseIntroducedString parses a string with charset introducer
func (ep *ExpressionParser) parseIntroducedString() expr.Expr {
	// Parse the actual string value
	val, _ := ep.parseValue()
	return val
}

// ExpectedAt creates an error at a specific token index
func (ep *ExpressionParser) ExpectedAt(expected string, index int) error {
	// This is a helper that delegates to the parser
	return fmt.Errorf("expected %s at position %d", expected, index)
}

// parseGroupingSetsExpr parses a GROUPING SETS expression
// Reference: src/parser/mod.rs:2593-2597
func (ep *ExpressionParser) parseGroupingSetsExpr(spanStart token.Span) (expr.Expr, error) {
	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var sets [][]expr.Expr
	for {
		// Parse each set as a tuple (can have 0 or more expressions)
		set, err := ep.parseTuple(false, true)
		if err != nil {
			return nil, err
		}
		sets = append(sets, set)

		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.GroupingSets{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Sets:    sets,
	}, nil
}

// parseCubeExpr parses a CUBE expression
// Reference: src/parser/mod.rs:2598-2602
func (ep *ExpressionParser) parseCubeExpr(spanStart token.Span) (expr.Expr, error) {
	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var sets [][]expr.Expr
	for {
		// Parse each set as a tuple (lift singleton to tuple, allow empty)
		set, err := ep.parseTuple(true, true)
		if err != nil {
			return nil, err
		}
		sets = append(sets, set)

		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Cube{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Sets:    sets,
	}, nil
}

// parseRollupExpr parses a ROLLUP expression
// Reference: src/parser/mod.rs:2603-2607
func (ep *ExpressionParser) parseRollupExpr(spanStart token.Span) (expr.Expr, error) {
	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var sets [][]expr.Expr
	for {
		// Parse each set as a tuple (lift singleton to tuple, allow empty)
		set, err := ep.parseTuple(true, true)
		if err != nil {
			return nil, err
		}
		sets = append(sets, set)

		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Rollup{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Sets:    sets,
	}, nil
}

// parseTuple parses a tuple with optional parentheses handling
// If liftSingleton is true, a single expression without parens is treated as a 1-element tuple
// If allowEmpty is true, empty parentheses are allowed
// Reference: src/parser/mod.rs:2625-2654
func (ep *ExpressionParser) parseTuple(liftSingleton, allowEmpty bool) ([]expr.Expr, error) {
	tok := ep.parser.PeekTokenRef()
	if _, ok := tok.Token.(token.TokenLParen); ok {
		// Has parentheses - consume them and parse contents
		ep.parser.AdvanceToken() // consume (

		// Check for empty tuple
		nextTok := ep.parser.PeekTokenRef()
		if _, isRParen := nextTok.Token.(token.TokenRParen); isRParen && allowEmpty {
			ep.parser.AdvanceToken() // consume )
			return []expr.Expr{}, nil
		}

		// Parse comma-separated expressions
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

		if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
		return exprs, nil
	}

	// No parentheses
	if liftSingleton {
		// Single expression becomes a 1-element tuple
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		return []expr.Expr{e}, nil
	}

	// Not a tuple - this is an error condition for non-lift mode
	return nil, fmt.Errorf("expected tuple")
}
