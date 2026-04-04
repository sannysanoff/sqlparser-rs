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
	"github.com/user/sqlparser/span"
	"github.com/user/sqlparser/tokenizer"
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
	case tokenizer.TokenWord:
		return ep.parsePrefixFromWord(tok, nextTok.Span)

	case tokenizer.TokenLBracket:
		// Array literal [1, 2, 3]
		return ep.parseArrayExpr(false)

	case tokenizer.TokenMinus:
		// Unary minus
		prec := ep.getPrecedence(dialects.PrecedenceMulDivModOp)
		innerExpr, err := ep.ParseExprWithPrecedence(prec)
		if err != nil {
			return nil, err
		}
		return &expr.UnaryOp{
			Op:      operator.UOpMinus,
			Expr:    innerExpr,
			SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
		}, nil

	case tokenizer.TokenPlus:
		// Unary plus
		prec := ep.getPrecedence(dialects.PrecedenceMulDivModOp)
		innerExpr, err := ep.ParseExprWithPrecedence(prec)
		if err != nil {
			return nil, err
		}
		return &expr.UnaryOp{
			Op:      operator.UOpPlus,
			Expr:    innerExpr,
			SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
		}, nil

	case tokenizer.TokenExclamationMark:
		if dialect.SupportsBangNotOperator() {
			// PostgreSQL-style bang not operator (!expr)
			prec := ep.getPrecedence(dialects.PrecedenceUnaryNot)
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

	case tokenizer.TokenDoubleExclamationMark:
		if dialect.Dialect() == "postgresql" {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

	case tokenizer.TokenPGSquareRoot:
		if dialect.Dialect() == "postgresql" {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

	case tokenizer.TokenPGCubeRoot:
		if dialect.Dialect() == "postgresql" {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

	case tokenizer.TokenAtSign:
		if dialect.Dialect() == "postgresql" {
			// PostgreSQL absolute value |x|
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

		if dialect.SupportsGeometricTypes() {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

	case tokenizer.TokenTilde:
		// Bitwise NOT
		prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
		innerExpr, err := ep.ParseExprWithPrecedence(prec)
		if err != nil {
			return nil, err
		}
		return &expr.UnaryOp{
			Op:      operator.UOpBitwiseNot,
			Expr:    innerExpr,
			SpanVal: mergeSpans(nextTok.Span, innerExpr.Span()),
		}, nil

	case tokenizer.TokenSharp:
		if dialect.SupportsGeometricTypes() {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

	case tokenizer.TokenAtDashAt:
		if dialect.SupportsGeometricTypes() {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

	case tokenizer.TokenQuestionMarkDash:
		if dialect.SupportsGeometricTypes() {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

	case tokenizer.TokenQuestionPipe:
		if dialect.SupportsGeometricTypes() {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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

	case tokenizer.TokenEscapedStringLiteral:
		if dialect.Dialect() == "postgresql" {
			ep.parser.PrevToken()
			return ep.parseValue()
		}
		ep.parser.PrevToken()
		return nil, ep.parser.ExpectedRef("an expression", ep.parser.PeekTokenRef())

	case tokenizer.TokenUnicodeStringLiteral:
		ep.parser.PrevToken()
		return ep.parseValue()

	case tokenizer.TokenNumber, tokenizer.TokenSingleQuotedString,
		tokenizer.TokenDoubleQuotedString, tokenizer.TokenTripleSingleQuotedString,
		tokenizer.TokenTripleDoubleQuotedString, tokenizer.TokenDollarQuotedString,
		tokenizer.TokenSingleQuotedByteStringLiteral, tokenizer.TokenDoubleQuotedByteStringLiteral,
		tokenizer.TokenTripleSingleQuotedByteStringLiteral, tokenizer.TokenTripleDoubleQuotedByteStringLiteral,
		tokenizer.TokenSingleQuotedRawStringLiteral, tokenizer.TokenDoubleQuotedRawStringLiteral,
		tokenizer.TokenTripleSingleQuotedRawStringLiteral, tokenizer.TokenTripleDoubleQuotedRawStringLiteral,
		tokenizer.TokenNationalStringLiteral, tokenizer.TokenQuoteDelimitedStringLiteral,
		tokenizer.TokenNationalQuoteDelimitedStringLiteral, tokenizer.TokenHexStringLiteral:
		ep.parser.PrevToken()
		return ep.parseValue()

	case tokenizer.TokenLParen:
		return ep.parseParenthesizedPrefix()

	case tokenizer.TokenPlaceholder, tokenizer.TokenColon:
		ep.parser.PrevToken()
		return ep.parseValue()

	case tokenizer.TokenLBrace:
		ep.parser.PrevToken()
		return ep.parseDictionaryExpr()

	default:
		return nil, ep.parser.ExpectedAt("an expression", tokIndex)
	}
}

// parsePrefixFromWord parses a prefix expression starting with a word token
func (ep *ExpressionParser) parsePrefixFromWord(word tokenizer.TokenWord, spanVal span.Span) (expr.Expr, error) {
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
func (ep *ExpressionParser) tryParseReservedWordPrefix(word *tokenizer.TokenWord, span span.Span) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	switch word.Word.Keyword {
	case "TRUE", "FALSE":
		if dialect.SupportsBooleanLiterals() {
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
		if dialect.SupportsTryConvert() {
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
		if _, ok := nextTok.Token.(tokenizer.TokenLParen); ok {
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
		if _, ok := nextTok.Token.(tokenizer.TokenLBracket); ok {
			// ARRAY[...] syntax
			if _, err := ep.parser.ExpectToken(tokenizer.TokenLBracket{}); err != nil {
				return nil, err
			}
			return ep.parseArrayExpr(true)
		}
		// ARRAY(...) subquery
		if _, ok := nextTok.Token.(tokenizer.TokenLParen); ok {
			if _, err := ep.parser.ExpectToken(tokenizer.TokenLParen{}); err != nil {
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
		if dialect.SupportsMatchAgainst() {
			return ep.parseMatchAgainstExpr()
		}

	case "STRUCT":
		if dialect.SupportsStructLiteral() {
			return ep.parseStructLiteral()
		}

	case "PRIOR":
		if ep.parser.GetState() == dialects.StateConnectBy {
			prec := ep.getPrecedence(dialects.PrecedencePlusMinus)
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
		if _, ok := nextTok.Token.(tokenizer.TokenLBrace); ok && dialect.SupportsMapLiteralSyntax() {
			return ep.parseMapLiteral()
		}

	case "LAMBDA":
		if dialect.SupportsLambdaFunctions() {
			return ep.parseLambdaExpr()
		}

	case "CIRCLE", "BOX", "PATH", "LINE", "LSEG", "POINT", "POLYGON":
		if dialect.SupportsGeometricTypes() {
			return ep.parseGeometricType(string(word.Word.Keyword))
		}
	}

	return nil, nil
}

// parseUnreservedWordPrefix parses an unreserved word as identifier or function
func (ep *ExpressionParser) parseUnreservedWordPrefix(word *tokenizer.TokenWord, span span.Span) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()
	ident := ep.wordToIdent(word, span)

	// Check for outer join operator (Oracle style)
	nextTok := ep.parser.PeekTokenRef()
	isOuterJoin := ep.peekOuterJoinOperator()

	if !isOuterJoin {
		if _, ok := nextTok.Token.(tokenizer.TokenLParen); ok {
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
		case tokenizer.TokenSingleQuotedString, tokenizer.TokenDoubleQuotedString,
			tokenizer.TokenHexStringLiteral:
			return &expr.Prefixed{
				SpanVal: mergeSpans(span, next.Span),
				Prefix:  ident,
				Value:   ep.parseIntroducedString(),
			}, nil
		}
	}

	// Check for lambda expression (single parameter)
	if dialect.SupportsLambdaFunctions() {
		next := ep.parser.PeekTokenRef()
		if _, ok := next.Token.(tokenizer.TokenArrow); ok {
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
	word, ok := tok.Token.(tokenizer.TokenWord)
	if !ok {
		return nil, ep.parser.Expected("an identifier", tok)
	}

	return ep.wordToIdent(&word, tok.Span), nil
}

// parseObjectName parses an object name (potentially qualified)
func (ep *ExpressionParser) parseObjectName() (*expr.ObjectName, error) {
	idents, err := ep.parseCommaSeparatedIdents()
	if err != nil {
		return nil, err
	}

	parts := make([]*expr.ObjectNamePart, len(idents))
	for i, ident := range idents {
		parts[i] = &expr.ObjectNamePart{
			SpanVal: ident.Span(),
			Ident:   ident,
		}
	}

	return &expr.ObjectName{
		SpanVal: idents[0].Span(),
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

		if !ep.parser.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	return idents, nil
}

// wordToIdent converts a TokenWord to an Ident
func (ep *ExpressionParser) wordToIdent(word *tokenizer.TokenWord, spanVal span.Span) *expr.Ident {
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
	case tokenizer.TokenNumber:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenSingleQuotedString:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenDoubleQuotedString:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenNationalStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenHexStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenEscapedStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenUnicodeStringLiteral:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenDollarQuotedString:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenPlaceholder:
		return &expr.ValueExpr{
			SpanVal: tok.Span,
			Value:   t,
		}, nil

	case tokenizer.TokenWord:
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

	case tokenizer.TokenColon, tokenizer.TokenAtSign:
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
	case tokenizer.TokenSingleQuotedString:
		return t.Value, nil
	case tokenizer.TokenDoubleQuotedString:
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

	if _, err := ep.parser.ExpectToken(tokenizer.TokenRParen{}); err != nil {
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
	word, ok := peekTok.Token.(tokenizer.TokenWord)
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
	_, isString := nextTok.Token.(tokenizer.TokenSingleQuotedString)
	if !isString {
		restore()
		return nil, false
	}

	// Parse the string value
	ep.parser.AdvanceToken() // consume the string
	stringTok := ep.parser.GetCurrentToken()
	strVal, _ := stringTok.Token.(tokenizer.TokenSingleQuotedString)

	// Special handling for INTERVAL - it has special syntax: INTERVAL 'value' unit [(precision)]
	if dataTypeName == "INTERVAL" {
		// Check for temporal unit (DAY, MONTH, YEAR, etc.)
		nextTok := ep.parser.PeekTokenRef()
		if word, ok := nextTok.Token.(tokenizer.TokenWord); ok {
			unit := word.Word.Keyword
			switch unit {
			case "YEAR", "YEARS", "MONTH", "MONTHS", "WEEK", "WEEKS",
				"DAY", "DAYS", "HOUR", "HOURS", "MINUTE", "MINUTES",
				"SECOND", "SECONDS":
				ep.parser.AdvanceToken() // consume the unit

				// Check for optional precision (n)
				nextTok2 := ep.parser.PeekTokenRef()
				if _, ok := nextTok2.Token.(tokenizer.TokenLParen); ok {
					ep.parser.AdvanceToken() // consume (
					ep.parser.AdvanceToken() // consume precision number
					// Consume ) if present
					ep.parser.ConsumeToken(tokenizer.TokenRParen{})
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
	if !dialect.SupportsOuterJoinOperator() {
		return false
	}

	toks := []tokenizer.TokenWithSpan{
		ep.parser.PeekNthToken(0),
		ep.parser.PeekNthToken(1),
		ep.parser.PeekNthToken(2),
	}

	_, isLParen := toks[0].Token.(tokenizer.TokenLParen)
	_, isPlus := toks[1].Token.(tokenizer.TokenPlus)
	_, isRParen := toks[2].Token.(tokenizer.TokenRParen)

	return isLParen && isPlus && isRParen
}

// peekSubquery checks if the next tokens indicate a subquery
func (ep *ExpressionParser) peekSubquery() bool {
	next := ep.parser.PeekTokenRef()
	if word, ok := next.Token.(tokenizer.TokenWord); ok {
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
