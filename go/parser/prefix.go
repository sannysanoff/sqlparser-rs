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

	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/operator"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/parseriface"
	"github.com/user/sqlparser/token"
)

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
		ep.parser.PrevToken()
		return ep.parseDictionaryExpr()

	default:
		return nil, ep.parser.ExpectedAt("an expression", tokIndex)
	}
}

// parsePrefixFromWord parses a prefix expression starting with a word token
func (ep *ExpressionParser) parsePrefixFromWord(word token.TokenWord, spanVal token.Span) (expr.Expr, error) {
	// Check if it's a reserved word with special meaning
	result, err := ep.tryParseReservedWordPrefix(&word, spanVal)
	if err != nil {
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
		// ARRAY(...) subquery
		if _, ok := nextTok.Token.(token.TokenLParen); ok {
			if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
				return nil, err
			}
			// Parse subquery
			// For now, return a placeholder
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
		if _, ok := next.Token.(token.TokenArrow); ok {
			ep.parser.AdvanceToken() // consume ->
			body, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			return &expr.LambdaExpr{
				SpanVal: mergeSpans(span, body.Span()),
				Params: []expr.LambdaFunctionParameter{{
					SpanVal: span,
					Name:    ident,
				}},
				Body:   body,
				Syntax: expr.LambdaArrow,
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
	word, ok := tok.Token.(token.TokenWord)
	if !ok {
		return nil, ep.parser.Expected("an identifier", tok)
	}

	return ep.wordToIdent(&word, tok.Span), nil
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
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case token.TokenDoubleQuotedString:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
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
		// Named or positional placeholder
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil
	}

	return nil, ep.parser.Expected("a value", tok)
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
	// Check for subquery first
	if ep.peekSubquery() {
		return ep.parseSubqueryExpr()
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
	dataTypeName := word.Word.Keyword
	isDataType := false
	switch dataTypeName {
	case "DATE", "TIME", "TIMESTAMP", "TIMESTAMPTZ", "INTERVAL", "DATETIME",
		"DECIMAL", "NUMERIC", "BIGNUMERIC", "CHAR", "VARCHAR", "NCHAR", "NVARCHAR",
		"CHARACTER", "BINARY", "VARBINARY", "JSON":
		isDataType = true
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
		restore()
		return nil, false
	}

	// Parse the string value
	ep.parser.AdvanceToken() // consume the string
	stringTok := ep.parser.GetCurrentToken()
	strVal, _ := stringTok.Token.(token.TokenSingleQuotedString)

	// Special handling for INTERVAL - it has special syntax: INTERVAL 'value' unit [(precision)]
	if dataTypeName == "INTERVAL" {
		// Check for temporal unit (DAY, MONTH, YEAR, etc.)
		nextTok := ep.parser.PeekTokenRef()
		if word, ok := nextTok.Token.(token.TokenWord); ok {
			unit := word.Word.Keyword
			switch unit {
			case "YEAR", "YEARS", "MONTH", "MONTHS", "WEEK", "WEEKS",
				"DAY", "DAYS", "HOUR", "HOURS", "MINUTE", "MINUTES",
				"SECOND", "SECONDS":
				ep.parser.AdvanceToken() // consume the unit

				// Check for optional precision (n)
				nextTok2 := ep.parser.PeekTokenRef()
				if _, ok := nextTok2.Token.(token.TokenLParen); ok {
					ep.parser.AdvanceToken() // consume (
					ep.parser.AdvanceToken() // consume precision number
					// Consume ) if present
					ep.parser.ConsumeToken(token.TokenRParen{})
				}

				unitStr := string(unit)
				return &expr.IntervalExpr{
					SpanVal: mergeSpans(peekTok.Span, ep.parser.GetCurrentToken().Span),
					Value: &expr.ValueExpr{
						SpanVal: stringTok.Span,
						Value:   strVal,
					},
					LeadingField: &unitStr,
				}, true
			}
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
func (ep *ExpressionParser) tryParseLambda() (expr.Expr, bool) {
	// Lambda expressions are: (x, y) -> expr or x -> expr
	// We need to check for the pattern: ( [ident [, ident]*] ) ->

	// For now, return false - complex lambda detection requires more context
	return nil, false
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
