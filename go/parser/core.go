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

// Package expressions provides SQL expression parsing methods for the sqlparser.
package parser

import (
	"fmt"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/datatype"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/span"
	"github.com/user/sqlparser/tokenizer"
)

// ParserInterface defines the methods needed from the parser
// This avoids circular dependencies between packages
type ParserInterface interface {
	// Token access methods
	PeekToken() tokenizer.TokenWithSpan
	PeekTokenRef() *tokenizer.TokenWithSpan
	PeekNthToken(n int) tokenizer.TokenWithSpan
	PeekNthTokenRef(n int) *tokenizer.TokenWithSpan
	NextToken() tokenizer.TokenWithSpan
	NextTokenNoSkip() *tokenizer.TokenWithSpan
	AdvanceToken()
	PrevToken()
	GetCurrentToken() *tokenizer.TokenWithSpan
	GetCurrentIndex() int

	// Token consumption
	ConsumeToken(expected tokenizer.Token) bool
	ExpectToken(expected tokenizer.Token) (tokenizer.TokenWithSpan, error)

	// Keyword helpers
	ParseKeyword(expected string) bool
	PeekKeyword(expected string) bool
	PeekNthKeyword(n int, expected string) bool
	ParseOneOfKeywords(keywords []string) string
	PeekOneOfKeywords(keywords []string) string
	ParseKeywords(keywords []string) bool
	ExpectKeyword(expected string) (tokenizer.TokenWithSpan, error)
	ExpectKeywords(expected []string) error

	// Comma-separated parsing
	ParseCommaSeparated(parseFn func() error) error

	// Expression and statement parsing
	ParseExpression() (ast.Expr, error)
	ParseInsert() (ast.Statement, error)
	ParseQuery() (ast.Statement, error)
	ParseDataType() (datatype.DataType, error)

	// State management
	GetState() dialects.ParserState
	SetState(state dialects.ParserState)
	InColumnDefinitionState() bool

	// Dialect access
	GetDialect() dialects.Dialect
	GetOptions() dialects.ParserOptions

	// Utility methods
	Expected(expected string, found tokenizer.TokenWithSpan) error
	ExpectedRef(expected string, found *tokenizer.TokenWithSpan) error
	ExpectedAt(expected string, index int) error
	SavePosition() func()
}

// mergeSpans combines two spans into one that covers both
func mergeSpans(a, b span.Span) span.Span {
	return a.Merge(b)
}

// ExpressionParser provides methods for parsing SQL expressions
type ExpressionParser struct {
	parser ParserInterface
}

// NewExpressionParser creates a new expression parser
func NewExpressionParser(parser ParserInterface) *ExpressionParser {
	return &ExpressionParser{parser: parser}
}

// ParseExpr parses an expression with the default precedence.
// This is the main entry point for expression parsing.
func (ep *ExpressionParser) ParseExpr() (expr.Expr, error) {
	return ep.ParseExprWithPrecedence(ep.parser.GetDialect().PrecUnknown())
}

// ParseExprWithPrecedence parses an expression with a minimum precedence.
// This implements precedence climbing (Pratt parsing) for handling operator
// precedence correctly.
func (ep *ExpressionParser) ParseExprWithPrecedence(precedence uint8) (expr.Expr, error) {
	// Parse the prefix (left-hand side)
	left, err := ep.parsePrefix()
	if err != nil {
		return nil, err
	}

	// Handle compound expressions (field access chains like a.b.c)
	left, err = ep.parseCompoundExpr(left, nil)
	if err != nil {
		return nil, err
	}

	// Continue parsing infix operators while they have higher precedence
	for {
		nextPrecedence, err := ep.GetNextPrecedence()
		if err != nil {
			return nil, err
		}

		// Stop if the next operator has lower or equal precedence
		if precedence >= nextPrecedence {
			break
		}

		// The period operator is handled exclusively by compound field access parsing
		nextTok := ep.parser.PeekTokenRef()
		if _, ok := nextTok.Token.(tokenizer.TokenPeriod); ok {
			break
		}

		// Parse the infix operator
		left, err = ep.parseInfix(left, nextPrecedence)
		if err != nil {
			return nil, err
		}
	}

	return left, nil
}

// GetNextPrecedence returns the precedence of the next token.
// This is used by the precedence climbing algorithm to determine
// when to stop parsing the current expression.
func (ep *ExpressionParser) GetNextPrecedence() (uint8, error) {
	// First, try the dialect's custom precedence handling
	prec, err := ep.parser.GetDialect().GetNextPrecedence(ep.parser)
	if err != nil {
		return 0, err
	}
	// If the dialect returns 0, fall back to the default implementation
	if prec == 0 {
		return ep.GetNextPrecedenceDefault()
	}
	return prec, nil
}

// GetNextPrecedenceDefault implements the default precedence logic
// that is used when the dialect doesn't provide custom precedence handling.
func (ep *ExpressionParser) GetNextPrecedenceDefault() (uint8, error) {
	dialect := ep.parser.GetDialect()
	nextTok := ep.parser.PeekTokenRef()

	switch tok := nextTok.Token.(type) {
	case tokenizer.TokenWord:
		switch tok.Word.Keyword {
		case "OR":
			return dialect.PrecValue(dialects.PrecedenceOr), nil
		case "AND":
			return dialect.PrecValue(dialects.PrecedenceAnd), nil
		case "NOT":
			return dialect.PrecValue(dialects.PrecedenceUnaryNot), nil
		case "IS":
			return dialect.PrecValue(dialects.PrecedenceIs), nil
		case "IN":
			return dialect.PrecValue(dialects.PrecedenceBetween), nil
		case "BETWEEN":
			return dialect.PrecValue(dialects.PrecedenceBetween), nil
		case "LIKE", "ILIKE":
			return dialect.PrecValue(dialects.PrecedenceLike), nil
		case "SIMILAR":
			return dialect.PrecValue(dialects.PrecedenceLike), nil
		case "REGEXP", "RLIKE":
			return dialect.PrecValue(dialects.PrecedenceLike), nil
		case "AT":
			return dialect.PrecValue(dialects.PrecedenceAtTz), nil
		case "NOTNULL":
			if dialect.SupportsNotnullOperator() {
				return dialect.PrecValue(dialects.PrecedenceIs), nil
			}
		case "MEMBER":
			return dialect.PrecValue(dialects.PrecedenceBetween), nil
		case "XOR":
			return dialect.PrecValue(dialects.PrecedenceXor), nil
		case "OVERLAPS":
			return dialect.PrecValue(dialects.PrecedenceBetween), nil
		case "OPERATOR":
			// PostgreSQL custom operator
			return dialect.PrecValue(dialects.PrecedenceEq), nil
		}

	case tokenizer.TokenPeriod:
		return dialect.PrecValue(dialects.PrecedencePeriod), nil

	case tokenizer.TokenDoubleColon:
		return dialect.PrecValue(dialects.PrecedenceDoubleColon), nil

	case tokenizer.TokenEq, tokenizer.TokenNeq, tokenizer.TokenGt, tokenizer.TokenGtEq,
		tokenizer.TokenLt, tokenizer.TokenLtEq, tokenizer.TokenSpaceship, tokenizer.TokenDoubleEq:
		return dialect.PrecValue(dialects.PrecedenceEq), nil

	case tokenizer.TokenPlus, tokenizer.TokenMinus:
		return dialect.PrecValue(dialects.PrecedencePlusMinus), nil

	case tokenizer.TokenMul, tokenizer.TokenDiv, tokenizer.TokenMod:
		return dialect.PrecValue(dialects.PrecedenceMulDivModOp), nil

	case tokenizer.TokenPipe:
		return dialect.PrecValue(dialects.PrecedencePipe), nil

	case tokenizer.TokenAmpersand:
		return dialect.PrecValue(dialects.PrecedenceAmpersand), nil

	case tokenizer.TokenCaret:
		return dialect.PrecValue(dialects.PrecedenceCaret), nil

	case tokenizer.TokenStringConcat:
		return dialect.PrecValue(dialects.PrecedencePlusMinus), nil

	case tokenizer.TokenShiftLeft, tokenizer.TokenShiftRight:
		if dialect.SupportsBitwiseShiftOperators() {
			return dialect.PrecValue(dialects.PrecedenceMulDivModOp), nil
		}

	case tokenizer.TokenSharp:
		if dialect.SupportsGeometricTypes() {
			return dialect.PrecValue(dialects.PrecedencePgOther), nil
		}

	case tokenizer.TokenOverlap:
		if dialect.Dialect() == "postgresql" {
			return dialect.PrecValue(dialects.PrecedenceEq), nil
		}
		if dialect.SupportsDoubleAmpersandOperator() {
			return dialect.PrecValue(dialects.PrecedenceAnd), nil
		}

	case tokenizer.TokenTilde, tokenizer.TokenTildeAsterisk,
		tokenizer.TokenExclamationMarkTilde, tokenizer.TokenExclamationMarkTildeAsterisk,
		tokenizer.TokenDoubleTilde, tokenizer.TokenDoubleTildeAsterisk,
		tokenizer.TokenExclamationMarkDoubleTilde, tokenizer.TokenExclamationMarkDoubleTildeAsterisk:
		return dialect.PrecValue(dialects.PrecedenceEq), nil

	case tokenizer.TokenArrow, tokenizer.TokenLongArrow,
		tokenizer.TokenHashArrow, tokenizer.TokenHashLongArrow:
		return dialect.PrecValue(dialects.PrecedenceEq), nil

	case tokenizer.TokenAtArrow, tokenizer.TokenArrowAt,
		tokenizer.TokenHashMinus, tokenizer.TokenAtQuestion,
		tokenizer.TokenAtAt, tokenizer.TokenQuestion,
		tokenizer.TokenQuestionAnd, tokenizer.TokenQuestionPipe:
		return dialect.PrecValue(dialects.PrecedenceEq), nil

	case tokenizer.TokenLBracket:
		// Array subscript
		return dialect.PrecValue(dialects.PrecedencePeriod), nil

	case tokenizer.TokenColon:
		if dialect.SupportsPartiQL() {
			return dialect.PrecValue(dialects.PrecedenceColon), nil
		}

	case tokenizer.TokenExclamationMark:
		if dialect.SupportsFactorialOperator() {
			// Postfix factorial operator
			return dialect.PrecValue(dialects.PrecedencePeriod), nil
		}
		// Also used for bang-not operator in some dialects
		if dialect.SupportsBangNotOperator() {
			return dialect.PrecValue(dialects.PrecedenceUnaryNot), nil
		}

	case tokenizer.TokenCustomBinaryOperator:
		// Custom operators (PostgreSQL)
		return dialect.PrecValue(dialects.PrecedenceEq), nil

	case tokenizer.TokenDuckIntDiv:
		// DuckDB integer division
		return dialect.PrecValue(dialects.PrecedenceMulDivModOp), nil
	}

	// Default: no infix operator
	return 0, nil
}

// parseCompoundExpr parses compound expressions like a.b.c or a.b[1].c
func (ep *ExpressionParser) parseCompoundExpr(root expr.Expr, chain []expr.AccessExpr) (expr.Expr, error) {
	return ep.parseCompoundExprWithOptions(root, chain, false)
}

// parseCompoundExprWithOptions parses compound expressions with wildcard support
func (ep *ExpressionParser) parseCompoundExprWithOptions(root expr.Expr, chain []expr.AccessExpr, allowWildcard bool) (expr.Expr, error) {
	dialect := ep.parser.GetDialect()
	var endingWildcard *tokenizer.TokenWithSpan

	for {
		// Check for dot access (e.g., foo.bar)
		if ep.parser.ConsumeToken(tokenizer.TokenPeriod{}) {
			nextTok := ep.parser.PeekTokenRef()

			switch tok := nextTok.Token.(type) {
			case tokenizer.TokenMul:
				// Handle qualified wildcard like foo.*
				if dialect.SupportsSelectWildcardExcept() {
					endingWildcard = &tokenizer.TokenWithSpan{
						Token: tok,
						Span:  nextTok.Span,
					}
					ep.parser.AdvanceToken()
				} else {
					// Put back the period for context parsing
					ep.parser.PrevToken()
				}
				goto done

			case tokenizer.TokenSingleQuotedString:
				// Quoted identifier as field name
				fieldExpr := &expr.ValueExpr{
					SpanVal: nextTok.Span,
					Value:   tok,
				}
				chain = append(chain, &expr.DotAccess{
					SpanVal: nextTok.Span,
					Expr:    fieldExpr,
				})
				ep.parser.AdvanceToken()

			case tokenizer.TokenPlaceholder:
				// Positional column reference (Snowflake $1, $2, etc.)
				ident := &expr.Ident{
					SpanVal: nextTok.Span,
					Value:   tok.Value,
				}
				chain = append(chain, &expr.DotAccess{
					SpanVal: nextTok.Span,
					Expr:    &expr.Identifier{SpanVal: nextTok.Span, Ident: ident},
				})
				ep.parser.AdvanceToken()

			case tokenizer.TokenWord:
				// Try to parse as expression with restricted precedence
				periodPrec := dialect.PrecValue(dialects.PrecedencePeriod)
				subExpr, err := ep.ParseExprWithPrecedence(periodPrec)
				if err != nil {
					// Fall back to identifier
					ident := ep.parseIdentifierFromWord(tok, nextTok.Span)
					chain = append(chain, &expr.DotAccess{
						SpanVal: nextTok.Span,
						Expr:    ident,
					})
				} else {
					// Flatten compound expressions
					switch e := subExpr.(type) {
					case *expr.CompoundFieldAccess:
						chain = append(chain, &expr.DotAccess{SpanVal: e.Root.Span(), Expr: e.Root})
						chain = append(chain, e.AccessChain...)
					case *expr.CompoundIdentifier:
						for _, part := range e.Idents {
							chain = append(chain, &expr.DotAccess{
								SpanVal: part.Span(),
								Expr:    &expr.Identifier{SpanVal: part.Span(), Ident: part},
							})
						}
					default:
						chain = append(chain, &expr.DotAccess{
							SpanVal: subExpr.Span(),
							Expr:    subExpr,
						})
					}
				}

			default:
				// Try to parse as expression
				periodPrec := dialect.PrecValue(dialects.PrecedencePeriod)
				subExpr, err := ep.ParseExprWithPrecedence(periodPrec)
				if err != nil {
					return nil, err
				}
				chain = append(chain, &expr.DotAccess{
					SpanVal: subExpr.Span(),
					Expr:    subExpr,
				})
			}

		} else if !dialect.SupportsPartiQL() {
			// Check for array subscript (e.g., foo[1])
			nextTok := ep.parser.PeekTokenRef()
			if _, ok := nextTok.Token.(tokenizer.TokenLBracket); ok {
				err := ep.parseMultiDimSubscript(&chain)
				if err != nil {
					return nil, err
				}
			} else {
				break
			}
		} else {
			break
		}
	}

done:
	// Handle outer join operator (+)
	if ep.maybeParseOuterJoinOperator() {
		if !ep.isAllIdent(root, chain) {
			return nil, ep.parser.ExpectedRef("column identifier before (+)", ep.parser.PeekTokenRef())
		}

		// Build the expression
		var result expr.Expr
		if len(chain) == 0 {
			result = root
		} else {
			idents, err := ep.exprsToIdents(root, chain)
			if err != nil {
				return nil, err
			}
			result = &expr.CompoundIdentifier{
				SpanVal: root.Span(),
				Idents:  idents,
			}
		}

		return &expr.OuterJoin{
			SpanVal: result.Span(),
			Expr:    result,
		}, nil
	}

	// Handle qualified wildcard
	if endingWildcard != nil {
		if !ep.isAllIdent(root, chain) {
			return nil, ep.parser.ExpectedRef("an identifier or '*' after '.'", ep.parser.PeekTokenRef())
		}

		idents, err := ep.exprsToIdents(root, chain)
		if err != nil {
			return nil, err
		}

		prefix := &expr.ObjectName{
			SpanVal: root.Span(),
			Parts:   make([]*expr.ObjectNamePart, len(idents)),
		}
		for i, ident := range idents {
			prefix.Parts[i] = &expr.ObjectNamePart{
				SpanVal: ident.Span(),
				Ident:   ident,
			}
		}

		return &expr.QualifiedWildcard{
			SpanVal: mergeSpans(root.Span(), endingWildcard.Span),
			Prefix:  prefix,
		}, nil
	}

	return ep.buildCompoundExpr(root, chain)
}

// buildCompoundExpr combines root expression and access chain into final expression
func (ep *ExpressionParser) buildCompoundExpr(root expr.Expr, chain []expr.AccessExpr) (expr.Expr, error) {
	if len(chain) == 0 {
		return root, nil
	}

	// If all parts are identifiers, create CompoundIdentifier
	if ep.isAllIdent(root, chain) {
		idents, err := ep.exprsToIdents(root, chain)
		if err != nil {
			return nil, err
		}
		return &expr.CompoundIdentifier{
			SpanVal: mergeSpans(root.Span(), chain[len(chain)-1].Span()),
			Idents:  idents,
		}, nil
	}

	// Check for qualified function call (e.g., schema.func(...))
	if ident, ok := root.(*expr.Identifier); ok {
		if len(chain) > 0 {
			lastAccess := chain[len(chain)-1]
			if dotAccess, ok := lastAccess.(*expr.DotAccess); ok {
				if fnExpr, ok := dotAccess.Expr.(*expr.FunctionExpr); ok {
					// Build qualified function name
					newParts := []*expr.ObjectNamePart{
						{SpanVal: ident.Span(), Ident: ident.Ident},
					}

					for _, access := range chain[:len(chain)-1] {
						if da, ok := access.(*expr.DotAccess); ok {
							if innerIdent, ok := da.Expr.(*expr.Identifier); ok {
								newParts = append(newParts, &expr.ObjectNamePart{
									SpanVal: innerIdent.Span(),
									Ident:   innerIdent.Ident,
								})
							}
						}
					}

					newParts = append(newParts, fnExpr.Name.Parts...)
					fnExpr.Name = &expr.ObjectName{
						SpanVal: mergeSpans(ident.Span(), fnExpr.Name.Span()),
						Parts:   newParts,
					}

					return fnExpr, nil
				}
			}
		}
	}

	// Return as CompoundFieldAccess
	return &expr.CompoundFieldAccess{
		SpanVal:     mergeSpans(root.Span(), chain[len(chain)-1].Span()),
		Root:        root,
		AccessChain: chain,
	}, nil
}

// isAllIdent checks if root and all fields are identifiers
func (ep *ExpressionParser) isAllIdent(root expr.Expr, fields []expr.AccessExpr) bool {
	if _, ok := root.(*expr.Identifier); !ok {
		return false
	}

	for _, f := range fields {
		dot, ok := f.(*expr.DotAccess)
		if !ok {
			return false
		}
		if _, ok := dot.Expr.(*expr.Identifier); !ok {
			return false
		}
	}

	return true
}

// exprsToIdents converts root and dot-access chain to identifier list
func (ep *ExpressionParser) exprsToIdents(root expr.Expr, fields []expr.AccessExpr) ([]*expr.Ident, error) {
	var idents []*expr.Ident

	rootIdent, ok := root.(*expr.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier, got %T", root)
	}
	idents = append(idents, rootIdent.Ident)

	for _, f := range fields {
		dot, ok := f.(*expr.DotAccess)
		if !ok {
			return nil, fmt.Errorf("expected dot access, got %T", f)
		}
		ident, ok := dot.Expr.(*expr.Identifier)
		if !ok {
			return nil, fmt.Errorf("expected identifier, got %T", dot.Expr)
		}
		idents = append(idents, ident.Ident)
	}

	return idents, nil
}

// maybeParseOuterJoinOperator checks for and consumes Oracle-style outer join operator
func (ep *ExpressionParser) maybeParseOuterJoinOperator() bool {
	dialect := ep.parser.GetDialect()
	if !dialect.SupportsOuterJoinOperator() {
		return false
	}

	toks := []tokenizer.TokenWithSpan{
		ep.parser.PeekNthToken(0),
		ep.parser.PeekNthToken(1),
		ep.parser.PeekNthToken(2),
	}

	if _, ok := toks[0].Token.(tokenizer.TokenLParen); !ok {
		return false
	}
	if _, ok := toks[1].Token.(tokenizer.TokenPlus); !ok {
		return false
	}
	if _, ok := toks[2].Token.(tokenizer.TokenRParen); !ok {
		return false
	}

	// Consume the tokens
	ep.parser.AdvanceToken() // (
	ep.parser.AdvanceToken() // +
	ep.parser.AdvanceToken() // )

	return true
}

// parseIdentifierFromWord creates an identifier from a word token
func (ep *ExpressionParser) parseIdentifierFromWord(word tokenizer.TokenWord, spanVal span.Span) expr.Expr {
	// Preserve the original value - no dialect-specific normalization
	// This matches the Rust reference implementation behavior
	value := word.Word.Value

	ident := &expr.Ident{
		SpanVal:    spanVal,
		Value:      value,
		QuoteStyle: nil,
	}
	if word.Word.QuoteStyle != nil {
		q := rune(*word.Word.QuoteStyle)
		ident.QuoteStyle = &q
	}

	return &expr.Identifier{
		SpanVal: spanVal,
		Ident:   ident,
	}
}

// parseMultiDimSubscript parses multi-dimensional array subscripts like [1][2][3]
func (ep *ExpressionParser) parseMultiDimSubscript(chain *[]expr.AccessExpr) error {
	for {
		nextTok := ep.parser.PeekTokenRef()
		if _, ok := nextTok.Token.(tokenizer.TokenLBracket); !ok {
			break
		}

		ep.parser.AdvanceToken() // consume [
		err := ep.parseSubscript(chain)
		if err != nil {
			return err
		}
	}
	return nil
}

// parseSubscript parses a single subscript expression like [1] or [1:3]
func (ep *ExpressionParser) parseSubscript(chain *[]expr.AccessExpr) error {
	sub, err := ep.parseSubscriptInner()
	if err != nil {
		return err
	}

	*chain = append(*chain, &expr.SubscriptAccess{
		SpanVal:   sub.Span(),
		Subscript: sub,
	})

	return nil
}

// parseSubscriptInner parses the contents of a subscript
func (ep *ExpressionParser) parseSubscriptInner() (*expr.Subscript, error) {
	dialect := ep.parser.GetDialect()
	colonPrec := dialect.PrecValue(dialects.PrecedenceColon)

	// Check for start of slice [:...]
	if ep.parser.ConsumeToken(tokenizer.TokenColon{}) {
		// We have [:...]
		if ep.parser.ConsumeToken(tokenizer.TokenRBracket{}) {
			// [:]
			return &expr.Subscript{
				SpanVal: ep.parser.GetCurrentToken().Span,
			}, nil
		}

		// Parse upper bound
		upperBound, err := ep.ParseExprWithPrecedence(colonPrec)
		if err != nil {
			return nil, err
		}

		if ep.parser.ConsumeToken(tokenizer.TokenRBracket{}) {
			// [:upper]
			return &expr.Subscript{
				SpanVal:    mergeSpans(ep.parser.GetCurrentToken().Span, upperBound.Span()),
				UpperBound: &upperBound,
			}, nil
		}

		// Expect second colon for stride
		if _, err := ep.parser.ExpectToken(tokenizer.TokenColon{}); err != nil {
			return nil, err
		}

		var stride expr.Expr
		if !ep.parser.ConsumeToken(tokenizer.TokenRBracket{}) {
			stride, err = ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			if _, err := ep.parser.ExpectToken(tokenizer.TokenRBracket{}); err != nil {
				return nil, err
			}
		}

		return &expr.Subscript{
			SpanVal:    mergeSpans(ep.parser.GetCurrentToken().Span, upperBound.Span()),
			UpperBound: &upperBound,
			Stride:     &stride,
		}, nil
	}

	// Parse index or lower bound of slice
	index, err := ep.ParseExprWithPrecedence(colonPrec)
	if err != nil {
		return nil, err
	}

	if ep.parser.ConsumeToken(tokenizer.TokenRBracket{}) {
		// Simple index [index]
		return &expr.Subscript{
			SpanVal: index.Span(),
			Index:   index,
		}, nil
	}

	// Must be slice [lower:...]
	if _, err := ep.parser.ExpectToken(tokenizer.TokenColon{}); err != nil {
		return nil, err
	}

	if ep.parser.ConsumeToken(tokenizer.TokenRBracket{}) {
		// [lower:]
		return &expr.Subscript{
			SpanVal:    index.Span(),
			LowerBound: &index,
		}, nil
	}

	// Parse upper bound
	upperBound, err := ep.ParseExprWithPrecedence(colonPrec)
	if err != nil {
		return nil, err
	}

	if ep.parser.ConsumeToken(tokenizer.TokenRBracket{}) {
		// [lower:upper]
		return &expr.Subscript{
			SpanVal:    mergeSpans(index.Span(), upperBound.Span()),
			LowerBound: &index,
			UpperBound: &upperBound,
		}, nil
	}

	// Must have stride [lower:upper:...]
	if _, err := ep.parser.ExpectToken(tokenizer.TokenColon{}); err != nil {
		return nil, err
	}

	var stride expr.Expr
	if !ep.parser.ConsumeToken(tokenizer.TokenRBracket{}) {
		stride, err = ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := ep.parser.ExpectToken(tokenizer.TokenRBracket{}); err != nil {
			return nil, err
		}
	}

	return &expr.Subscript{
		SpanVal:    mergeSpans(index.Span(), upperBound.Span()),
		LowerBound: &index,
		UpperBound: &upperBound,
		Stride:     &stride,
	}, nil
}

// Parser returns the underlying parser interface
func (ep *ExpressionParser) Parser() ParserInterface {
	return ep.parser
}

// Helper method to get operator precedence from dialect
func (ep *ExpressionParser) getPrecedence(prec dialects.Precedence) uint8 {
	return ep.parser.GetDialect().PrecValue(prec)
}
