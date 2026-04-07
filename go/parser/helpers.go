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

	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/dialects"
	"github.com/user/sqlparser/parseriface"
	"github.com/user/sqlparser/token"
)

// parseCommaSeparatedExprs parses a comma-separated list of expressions
func (ep *ExpressionParser) parseCommaSeparatedExprs() ([]expr.Expr, error) {
	var exprs []expr.Expr

	for {
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		exprs = append(exprs, e)

		// Check for trailing comma support
		dialect := ep.parser.GetDialect()
		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}

		// Check if trailing comma is allowed and we're at the end
		if dialects.SupportsTrailingCommas(dialect) {
			next := ep.parser.PeekTokenRef()
			if _, isRParen := next.Token.(token.TokenRParen); isRParen {
				// Trailing comma is allowed
				break
			}
		}
	}

	return exprs, nil
}

// parseParenthesizedExpr parses an expression in parentheses
func (ep *ExpressionParser) parseParenthesizedExpr() (expr.Expr, error) {
	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	inner, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Nested{
		SpanVal: mergeSpans(ep.parser.GetCurrentToken().Span, inner.Span()),
		Expr:    inner,
	}, nil
}

// parseOptionalAlias parses an optional alias with or without AS keyword
func (ep *ExpressionParser) parseOptionalAlias(reservedWords []string) (*expr.Ident, error) {
	return ep.parseOptionalAliasWithValidator(func(explicit bool, kw string) bool {
		// Check if the keyword is in the reserved list
		for _, reserved := range reservedWords {
			if kw == reserved {
				return false
			}
		}
		return true
	})
}

// parseOptionalAliasWithValidator parses an optional alias with custom validation
func (ep *ExpressionParser) parseOptionalAliasWithValidator(validator func(explicit bool, kw string) bool) (*expr.Ident, error) {
	explicit := ep.parser.ParseKeyword("AS")

	nextTok := ep.parser.PeekTokenRef()
	word, ok := nextTok.Token.(token.TokenWord)
	if !ok {
		if explicit {
			return nil, fmt.Errorf("expected identifier after AS")
		}
		return nil, nil
	}

	// Validate the word can be used as alias
	if !validator(explicit, string(word.Word.Keyword)) {
		if explicit {
			return nil, fmt.Errorf("'%s' cannot be used as alias", word.Word.Keyword)
		}
		return nil, nil
	}

	ep.parser.AdvanceToken()
	return ep.wordToIdent(&word, nextTok.Span), nil
}

// parseOptionalAliasNoAs parses an optional alias without AS keyword
func (ep *ExpressionParser) parseOptionalAliasNoAs(reservedWords []string) (*expr.Ident, error) {
	return ep.parseOptionalAliasWithValidator(func(explicit bool, kw string) bool {
		if explicit {
			return false // Must not have AS
		}
		for _, reserved := range reservedWords {
			if kw == reserved {
				return false
			}
		}
		return true
	})
}

// parseAliasExpr parses an aliased expression (expr [AS] alias)
func (ep *ExpressionParser) parseAliasExpr() (*expr.NamedExpr, error) {
	e, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	alias, err := ep.parseOptionalAlias([]string{})
	if err != nil {
		return nil, err
	}

	return &expr.NamedExpr{
		SpanVal: mergeSpans(e.Span(), alias.Span()),
		Expr:    e,
		Name:    alias,
	}, nil
}

// parseArrayExpr parses an array expression like [1, 2, 3] or ARRAY[1, 2, 3]
func (ep *ExpressionParser) parseArrayExpr(named bool) (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	// If not named (ARRAY[...]), we're already past the [
	// If named, we need to consume the [
	if named {
		// We already consumed ARRAY and [
	} else {
		// We just consumed [
	}

	// Parse elements
	var elements []expr.Expr
	_, isRBracket := ep.parser.PeekTokenRef().Token.(token.TokenRBracket)
	if !isRBracket {
		exprs, err := ep.parseCommaSeparatedExprs()
		if err != nil {
			return nil, err
		}
		elements = exprs
	}

	if _, err := ep.parser.ExpectToken(token.TokenRBracket{}); err != nil {
		return nil, err
	}

	return &expr.ArrayExpr{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Elems:   elements,
		Named:   named,
	}, nil
}

// parseIntervalExpr parses an INTERVAL expression
func (ep *ExpressionParser) parseIntervalExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span
	dialect := ep.parser.GetDialect()

	// Parse value
	var value expr.Expr
	if dialects.RequireIntervalQualifier(dialect) {
		// Parse as full expression
		v, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		value = v
	} else {
		// Parse as prefix expression
		v, err := ep.parsePrefix()
		if err != nil {
			return nil, err
		}
		value = v
	}

	// Parse temporal unit
	var leadingField *string
	if ep.isTemporalUnit() {
		unit := ep.parseTemporalUnit()
		leadingField = &unit
	} else if dialects.RequireIntervalQualifier(dialect) {
		return nil, fmt.Errorf("INTERVAL requires a unit after the literal value")
	}

	// Parse precision for SECOND
	var leadingPrecision, fracPrecision *uint64
	if leadingField != nil && *leadingField == "SECOND" {
		lp, fp, err := ep.parseIntervalPrecisions()
		if err != nil {
			return nil, err
		}
		leadingPrecision = lp
		fracPrecision = fp
	} else if leadingField != nil {
		lp, err := ep.parseOptionalPrecision()
		if err != nil {
			return nil, err
		}
		leadingPrecision = lp
	}

	// Parse TO clause for range
	var lastField *string
	if leadingField != nil && ep.parser.ParseKeyword("TO") {
		// SECOND TO SECOND is not allowed - SQL mandates special format SECOND (precision, frac_precision)
		if leadingField != nil && *leadingField == "SECOND" {
			return nil, fmt.Errorf("syntax error at word: TO")
		}

		unit := ep.parseTemporalUnit()
		lastField = &unit

		// Check for precision on last field - only allowed for SECOND
		// For other fields, having precision on both sides is an error
		if *lastField == "SECOND" {
			fp, err := ep.parseOptionalPrecision()
			if err != nil {
				return nil, err
			}
			fracPrecision = fp
		} else {
			// For non-SECOND fields, check if there's precision on last field (which is an error)
			nextTok := ep.parser.PeekTokenRef()
			if _, ok := nextTok.Token.(token.TokenLParen); ok {
				return nil, fmt.Errorf("syntax error at word: (")
			}
		}
	}

	return &expr.IntervalExpr{
		SpanVal:                    mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Value:                      value,
		LeadingField:               leadingField,
		LeadingPrecision:           leadingPrecision,
		LastField:                  lastField,
		FractionalSecondsPrecision: fracPrecision,
	}, nil
}

// isTemporalUnit checks if the next token is a temporal unit keyword
func (ep *ExpressionParser) isTemporalUnit() bool {
	units := []string{
		"YEAR", "YEARS", "MONTH", "MONTHS", "WEEK", "WEEKS",
		"DAY", "DAYS", "HOUR", "HOURS", "MINUTE", "MINUTES",
		"SECOND", "SECONDS", "CENTURY", "DECADE", "DOW", "DOY",
		"EPOCH", "ISODOW", "ISOYEAR", "JULIAN", "MICROSECOND",
		"MICROSECONDS", "MILLENIUM", "MILLENNIUM", "MILLISECOND",
		"MILLISECONDS", "NANOSECOND", "NANOSECONDS", "QUARTER",
		"TIMEZONE", "TIMEZONE_HOUR", "TIMEZONE_MINUTE",
	}

	next := ep.parser.PeekTokenRef()
	if word, ok := next.Token.(token.TokenWord); ok {
		for _, unit := range units {
			if word.Word.Keyword == token.Keyword(unit) {
				return true
			}
		}
	}
	return false
}

// parseTemporalUnit parses a temporal unit keyword
func (ep *ExpressionParser) parseTemporalUnit() string {
	tok := ep.parser.NextToken()
	if word, ok := tok.Token.(token.TokenWord); ok {
		return string(word.Word.Keyword)
	}
	return ""
}

// parseOptionalPrecision parses an optional precision like YEAR(2)
func (ep *ExpressionParser) parseOptionalPrecision() (*uint64, error) {
	if !ep.parser.ConsumeToken(token.TokenLParen{}) {
		return nil, nil
	}

	tok := ep.parser.NextToken()
	num, ok := tok.Token.(token.TokenNumber)
	if !ok {
		return nil, fmt.Errorf("expected number in precision")
	}

	n, err := strconv.ParseUint(num.Value, 10, 64)
	if err != nil {
		return nil, err
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &n, nil
}

// parseIntervalPrecisions parses precision for SECOND with optional fractional seconds
func (ep *ExpressionParser) parseIntervalPrecisions() (*uint64, *uint64, error) {
	if !ep.parser.ConsumeToken(token.TokenLParen{}) {
		return nil, nil, nil
	}

	// Parse leading precision
	tok := ep.parser.NextToken()
	num, ok := tok.Token.(token.TokenNumber)
	if !ok {
		return nil, nil, fmt.Errorf("expected number in precision")
	}

	lp, err := strconv.ParseUint(num.Value, 10, 64)
	if err != nil {
		return nil, nil, err
	}
	leadingPrec := &lp

	// Check for comma and fractional seconds precision
	var fracPrec *uint64
	if ep.parser.ConsumeToken(token.TokenComma{}) {
		tok = ep.parser.NextToken()
		num, ok = tok.Token.(token.TokenNumber)
		if !ok {
			return nil, nil, fmt.Errorf("expected number for fractional seconds precision")
		}
		fp, err := strconv.ParseUint(num.Value, 10, 64)
		if err != nil {
			return nil, nil, err
		}
		fracPrec = &fp
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, nil, err
	}

	return leadingPrec, fracPrec, nil
}

// parseCastExpr parses a CAST expression
func (ep *ExpressionParser) parseCastExpr(kind expr.CastKind) (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	castExpr, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if !ep.parser.ParseKeyword("AS") {
		return nil, fmt.Errorf("expected AS in CAST expression")
	}

	// Parse data type using the proper data type parser
	dataType, err := ep.parser.ParseDataType()
	if err != nil {
		return nil, err
	}

	// Check for ARRAY suffix (MySQL)
	isArray := ep.parser.ParseKeyword("ARRAY")

	// Parse optional FORMAT clause
	var format *expr.CastFormat
	if ep.parser.ParseKeyword("FORMAT") {
		// Parse format value
		fmtVal, err := ep.parseValue()
		if err != nil {
			return nil, err
		}
		_ = fmtVal // Use in full implementation
		format = &expr.CastFormat{}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Cast{
		SpanVal:  mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Kind:     kind,
		Expr:     castExpr,
		DataType: dataType.String(),
		Array:    isArray,
		Format:   format,
	}, nil
}

// parseConvertExpr parses a CONVERT expression
func (ep *ExpressionParser) parseConvertExpr(isTry bool) (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span
	dialect := ep.parser.GetDialect()

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	var convertExpr expr.Expr
	var dataType string
	var charset *expr.ObjectName
	var targetBeforeValue bool
	var styles []expr.Expr

	if dialects.ConvertTypeBeforeValue(dialect) {
		// MSSQL syntax: CONVERT(type, expr [, style])
		targetBeforeValue = true
		dt, err := ep.parseIdentifier()
		if err != nil {
			return nil, err
		}
		dataType = dt.Value

		// Check for data type arguments like VARCHAR(MAX) or DECIMAL(10,5)
		if ep.parser.ConsumeToken(token.TokenLParen{}) {
			dataType += "("
			// Parse the content inside parentheses
			for {
				tok := ep.parser.PeekTokenRef()
				if ep.parser.ConsumeToken(token.TokenRParen{}) {
					dataType += ")"
					break
				}
				if ep.parser.ConsumeToken(token.TokenComma{}) {
					dataType += ", "
					continue
				}
				// Add the token string representation
				dataType += tok.Token.String()
				ep.parser.AdvanceToken()
			}
		}

		if _, err := ep.parser.ExpectToken(token.TokenComma{}); err != nil {
			return nil, err
		}

		ce, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		convertExpr = ce

		// Parse optional style(s)
		for ep.parser.ConsumeToken(token.TokenComma{}) {
			style, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			styles = append(styles, style)
		}
	} else {
		// MySQL/standard syntax: CONVERT(expr, type) or CONVERT(expr USING charset)
		ce, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		convertExpr = ce

		if ep.parser.ParseKeyword("USING") {
			// USING charset
			cs, err := ep.parseObjectName()
			if err != nil {
				return nil, err
			}
			charset = cs
		} else {
			// AS type
			if _, err := ep.parser.ExpectToken(token.TokenComma{}); err != nil {
				return nil, err
			}
			dt, err := ep.parseIdentifier()
			if err != nil {
				return nil, err
			}
			dataType = dt.Value

			// Check for data type arguments like VARCHAR(MAX) or DECIMAL(10,5)
			if ep.parser.ConsumeToken(token.TokenLParen{}) {
				dataType += "("
				// Parse the content inside parentheses
				for {
					tok := ep.parser.PeekTokenRef()
					if ep.parser.ConsumeToken(token.TokenRParen{}) {
						dataType += ")"
						break
					}
					if ep.parser.ConsumeToken(token.TokenComma{}) {
						dataType += ", "
						continue
					}
					// Add the token string representation
					dataType += tok.Token.String()
					ep.parser.AdvanceToken()
				}
			}

			// Check for CHARACTER SET
			if ep.parser.ParseKeywords([]string{"CHARACTER", "SET"}) {
				cs, err := ep.parseObjectName()
				if err != nil {
					return nil, err
				}
				charset = cs
			}
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	// Only set DataType if it's not empty
	var dataTypePtr *string
	if dataType != "" {
		dataTypePtr = &dataType
	}

	return &expr.Convert{
		SpanVal:           mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		IsTry:             isTry,
		Expr:              convertExpr,
		DataType:          dataTypePtr,
		Charset:           charset,
		TargetBeforeValue: targetBeforeValue,
		Styles:            styles,
	}, nil
}

// parseExtractExpr parses an EXTRACT expression
func (ep *ExpressionParser) parseExtractExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span
	dialect := ep.parser.GetDialect()

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse field (temporal unit)
	field := ep.parseTemporalUnit()

	// Parse syntax: FROM expr or , expr (for some dialects)
	var syntax expr.ExtractSyntax
	if ep.parser.ParseKeyword("FROM") {
		syntax = expr.ExtractFrom
	} else if dialects.SupportsExtractCommaSyntax(dialect) && ep.parser.ConsumeToken(token.TokenComma{}) {
		syntax = expr.ExtractComma
	} else {
		return nil, fmt.Errorf("expected FROM or comma after field in EXTRACT")
	}

	exprVal, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Extract{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Field:   field,
		Syntax:  syntax,
		Expr:    exprVal,
	}, nil
}

// parseCeilFloorExpr parses CEIL or FLOOR expression
func (ep *ExpressionParser) parseCeilFloorExpr(isCeil bool) (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	exprVal, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	ceilExpr := &expr.CeilExpr{
		Expr: exprVal,
	}

	if ep.parser.ParseKeyword("TO") {
		// CEIL/FLOOR(expr TO DateTimeField)
		ceilExpr.Field.Kind = expr.CeilFloorDateTime
		dtField := ep.parseTemporalUnit()
		ceilExpr.Field.DateTimeField = &dtField
	} else if ep.parser.ConsumeToken(token.TokenComma{}) {
		// CEIL/FLOOR(expr, scale)
		ceilExpr.Field.Kind = expr.CeilFloorScale
		scale, err := ep.parseValue()
		if err != nil {
			return nil, err
		}
		ceilExpr.Field.Scale = scale
	} else {
		// CEIL/FLOOR(expr) - simple case, no DateTimeField or Scale
		ceilExpr.Field.Kind = expr.CeilFloorDateTime
		// Don't set DateTimeField - leave it nil so String() outputs simple form
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	ceilExpr.SpanVal = mergeSpans(spanStart, ep.parser.GetCurrentToken().Span)

	if isCeil {
		return ceilExpr, nil
	}

	return &expr.FloorExpr{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Expr:    exprVal,
		Field:   ceilExpr.Field,
	}, nil
}

// parsePositionExpr parses a POSITION expression
// Following the Rust implementation, if the special POSITION(expr IN expr) syntax
// doesn't parse correctly, it falls back to treating POSITION as a regular function.
func (ep *ExpressionParser) parsePositionExpr(word token.TokenWord, span token.Span) (expr.Expr, error) {
	// First, consume the '(' token - this is required for both syntaxes
	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Save position AFTER consuming '(' for potential backtracking
	startIndex := ep.parser.GetCurrentIndex()

	// Try to parse the special POSITION(substr IN str) syntax
	betweenPrec := ep.getPrecedence(parseriface.PrecedenceBetween)
	substrExpr, err := ep.ParseExprWithPrecedence(betweenPrec)
	if err != nil {
		// Backtrack and try as regular function
		ep.parser.SetCurrentIndex(startIndex)
		return ep.parseFunctionWithName(&expr.ObjectName{
			SpanVal: span,
			Parts:   []*expr.ObjectNamePart{{SpanVal: span, Ident: ep.wordToIdent(&word, span)}},
		})
	}

	// Check for IN keyword - if not found, this might be Snowflake-style function call
	if !ep.parser.ParseKeyword("IN") {
		// Backtrack and try as regular function
		ep.parser.SetCurrentIndex(startIndex)
		return ep.parseFunctionWithName(&expr.ObjectName{
			SpanVal: span,
			Parts:   []*expr.ObjectNamePart{{SpanVal: span, Ident: ep.wordToIdent(&word, span)}},
		})
	}

	// Parse the string to search in
	inExpr, err := ep.ParseExpr()
	if err != nil {
		// Backtrack and parse as regular function
		ep.parser.SetCurrentIndex(startIndex)
		return ep.parseFunctionWithName(&expr.ObjectName{
			SpanVal: span,
			Parts:   []*expr.ObjectNamePart{{SpanVal: span, Ident: ep.wordToIdent(&word, span)}},
		})
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		// Backtrack and parse as regular function
		ep.parser.SetCurrentIndex(startIndex)
		return ep.parseFunctionWithName(&expr.ObjectName{
			SpanVal: span,
			Parts:   []*expr.ObjectNamePart{{SpanVal: span, Ident: ep.wordToIdent(&word, span)}},
		})
	}

	return &expr.PositionExpr{
		SpanVal: mergeSpans(span, ep.parser.GetCurrentToken().Span),
		Expr:    substrExpr,
		In:      inExpr,
	}, nil
}

// parseSubstringExpr parses a SUBSTRING expression
func (ep *ExpressionParser) parseSubstringExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	// Parse SUBSTR or SUBSTRING keyword
	shorthand := ep.parser.ParseKeyword("SUBSTR")
	if !shorthand && !ep.parser.ParseKeyword("SUBSTRING") {
		return nil, fmt.Errorf("expected SUBSTR or SUBSTRING")
	}

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse the string expression
	exprVal, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	var fromExpr, forExpr *expr.Expr
	var special bool

	// Check for comma syntax (special) or FROM syntax
	if ep.parser.ConsumeToken(token.TokenComma{}) {
		special = true
		f, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		fromExpr = &f

		if ep.parser.ConsumeToken(token.TokenComma{}) {
			t, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			forExpr = &t
		}
	} else {
		if ep.parser.ParseKeyword("FROM") {
			f, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			fromExpr = &f
		}

		if ep.parser.ParseKeyword("FOR") {
			t, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			forExpr = &t
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Substring{
		SpanVal:       mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Expr:          exprVal,
		SubstringFrom: fromExpr,
		SubstringFor:  forExpr,
		Special:       special,
		Shorthand:     shorthand,
	}, nil
}

// parseOverlayExpr parses an OVERLAY expression
func (ep *ExpressionParser) parseOverlayExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse the base string
	exprVal, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if !ep.parser.ParseKeyword("PLACING") {
		return nil, fmt.Errorf("expected PLACING in OVERLAY expression")
	}

	// Parse the overlay string
	overlayWhat, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	if !ep.parser.ParseKeyword("FROM") {
		return nil, fmt.Errorf("expected FROM in OVERLAY expression")
	}

	// Parse the starting position
	overlayFrom, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	// Optional FOR length
	var overlayFor *expr.Expr
	if ep.parser.ParseKeyword("FOR") {
		f, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		overlayFor = &f
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.OverlayExpr{
		SpanVal:     mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Expr:        exprVal,
		OverlayWhat: overlayWhat,
		OverlayFrom: overlayFrom,
		OverlayFor:  overlayFor,
	}, nil
}

// parseTrimExpr parses a TRIM expression
func (ep *ExpressionParser) parseTrimExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span
	dialect := ep.parser.GetDialect()

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Check for trim where specification
	var trimWhere *expr.TrimWhere
	if ep.parser.ParseKeyword("BOTH") {
		w := expr.TrimBoth
		trimWhere = &w
	} else if ep.parser.ParseKeyword("LEADING") {
		w := expr.TrimLeading
		trimWhere = &w
	} else if ep.parser.ParseKeyword("TRAILING") {
		w := expr.TrimTrailing
		trimWhere = &w
	}

	// Parse trim what and expr
	var trimWhat *expr.Expr
	var trimExpr expr.Expr

	if trimWhere != nil {
		// [where] [what] FROM expr
		if !ep.parser.PeekKeyword("FROM") {
			w, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			trimWhat = &w
		}
		if !ep.parser.ParseKeyword("FROM") {
			return nil, fmt.Errorf("expected FROM in TRIM expression")
		}
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		trimExpr = e
	} else {
		// expr or what FROM expr
		e, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}

		if ep.parser.ParseKeyword("FROM") {
			trimWhat = &e
			e2, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			trimExpr = e2
		} else {
			trimExpr = e
		}
	}

	// Check for comma-separated trim characters (some dialects)
	var trimChars []expr.Expr
	if dialects.SupportsCommaSeparatedTrim(dialect) && ep.parser.ConsumeToken(token.TokenComma{}) {
		chars, err := ep.parseCommaSeparatedExprs()
		if err != nil {
			return nil, err
		}
		trimChars = chars
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.TrimExpr{
		SpanVal:        mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		TrimWhere:      trimWhere,
		TrimWhat:       trimWhat,
		Expr:           trimExpr,
		TrimCharacters: trimChars,
	}, nil
}

// parseStructLiteral parses a STRUCT literal expression
func (ep *ExpressionParser) parseStructLiteral() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	// Check for typed struct syntax: STRUCT<type>(value1, value2)
	var fields []expr.StructField
	if _, ok := ep.parser.PeekTokenRef().Token.(token.TokenLt); ok {
		// Parse type definition
		if _, err := ep.parser.ExpectToken(token.TokenLt{}); err != nil {
			return nil, err
		}
		// Parse field definitions (simplified)
		for {
			_, isGt := ep.parser.PeekTokenRef().Token.(token.TokenGt)
			if isGt {
				break
			}
			fieldName, err := ep.parseIdentifier()
			if err != nil {
				return nil, err
			}
			fieldType, err := ep.parseIdentifier()
			if err != nil {
				return nil, err
			}
			fields = append(fields, expr.StructField{
				SpanVal:   fieldName.Span(),
				FieldName: fieldName,
				FieldType: fieldType.Value,
			})
			if !ep.parser.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := ep.parser.ExpectToken(token.TokenGt{}); err != nil {
			return nil, err
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse values
	values, err := ep.parseCommaSeparatedExprs()
	if err != nil {
		return nil, err
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.StructExpr{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Values:  values,
		Fields:  fields,
	}, nil
}

// parseMapLiteral parses a MAP literal expression (DuckDB)
func (ep *ExpressionParser) parseMapLiteral() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span
	dialect := ep.parser.GetDialect()

	if !dialects.SupportsMapLiteralSyntax(dialect) {
		return nil, fmt.Errorf("MAP literal not supported in this dialect")
	}

	if _, err := ep.parser.ExpectToken(token.TokenLBrace{}); err != nil {
		return nil, err
	}

	var entries []expr.MapEntry
	for {
		_, isRBrace := ep.parser.PeekTokenRef().Token.(token.TokenRBrace)
		if isRBrace {
			break
		}
		key, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		if _, err := ep.parser.ExpectToken(token.TokenColon{}); err != nil {
			return nil, err
		}
		value, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		entries = append(entries, expr.MapEntry{
			SpanVal: mergeSpans(key.Span(), value.Span()),
			Key:     key,
			Value:   value,
		})
		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRBrace{}); err != nil {
		return nil, err
	}

	return &expr.MapExpr{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Entries: entries,
	}, nil
}

// parseDictionaryExpr parses a dictionary literal expression
func (ep *ExpressionParser) parseDictionaryExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	if _, err := ep.parser.ExpectToken(token.TokenLBrace{}); err != nil {
		return nil, err
	}

	var fields []expr.DictionaryField
	for {
		_, isRBrace := ep.parser.PeekTokenRef().Token.(token.TokenRBrace)
		if isRBrace {
			break
		}
		key, err := ep.parseIdentifier()
		if err != nil {
			return nil, err
		}
		if _, err := ep.parser.ExpectToken(token.TokenColon{}); err != nil {
			return nil, err
		}
		value, err := ep.ParseExpr()
		if err != nil {
			return nil, err
		}
		fields = append(fields, expr.DictionaryField{
			SpanVal: key.Span(),
			Key:     key,
			Value:   value,
		})
		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRBrace{}); err != nil {
		return nil, err
	}

	return &expr.DictionaryExpr{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Fields:  fields,
	}, nil
}

// parseLambdaExpr parses a lambda function expression
func (ep *ExpressionParser) parseLambdaExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span
	dialect := ep.parser.GetDialect()

	if !dialects.SupportsLambdaFunctions(dialect) {
		return nil, fmt.Errorf("lambda functions not supported in this dialect")
	}

	// Parse parameters: either (x, y) or just x
	var params []expr.LambdaFunctionParameter
	next := ep.parser.PeekTokenRef()
	if _, ok := next.Token.(token.TokenLParen); ok {
		// Multiple parameters in parentheses
		if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
			return nil, err
		}
		for {
			_, isRParen := ep.parser.PeekTokenRef().Token.(token.TokenRParen)
			if isRParen {
				break
			}
			ident, err := ep.parseIdentifier()
			if err != nil {
				return nil, err
			}
			param := expr.LambdaFunctionParameter{
				SpanVal: ident.Span(),
				Name:    ident,
			}
			// Check for optional type
			_, isArrow := ep.parser.PeekTokenRef().Token.(token.TokenArrow)
			if !isArrow {
				dt, err := ep.parseIdentifier()
				if err == nil {
					param.DataType = dt.Value
				}
			}
			params = append(params, param)
			if !ep.parser.ConsumeToken(token.TokenComma{}) {
				break
			}
		}
		if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
			return nil, err
		}
	} else {
		// Single parameter without parentheses
		ident, err := ep.parseIdentifier()
		if err != nil {
			return nil, err
		}
		params = append(params, expr.LambdaFunctionParameter{
			SpanVal: ident.Span(),
			Name:    ident,
		})
	}

	// Expect arrow
	if !ep.parser.ConsumeToken(token.TokenArrow{}) {
		return nil, fmt.Errorf("expected -> in lambda expression")
	}

	// Parse body
	body, err := ep.ParseExpr()
	if err != nil {
		return nil, err
	}

	return &expr.LambdaExpr{
		SpanVal: mergeSpans(spanStart, body.Span()),
		Params:  params,
		Body:    body,
		Syntax:  expr.LambdaArrow,
	}, nil
}

// parseTimeFunction parses time functions like CURRENT_TIMESTAMP, NOW(), etc.
func (ep *ExpressionParser) parseTimeFunction(word *token.TokenWord, span token.Span) (expr.Expr, error) {
	// Check for parentheses (some time functions can be called like NOW())
	next := ep.parser.PeekTokenRef()
	if _, ok := next.Token.(token.TokenLParen); ok {
		return ep.parseFunction(&expr.ObjectName{
			SpanVal: span,
			Parts: []*expr.ObjectNamePart{{
				SpanVal: span,
				Ident:   ep.wordToIdent(word, span),
			}},
		})
	}

	// Without parentheses, treat as identifier/expression
	ident := ep.wordToIdent(word, span)
	return &expr.Identifier{
		SpanVal: span,
		Ident:   ident,
	}, nil
}

// parseMatchAgainstExpr parses a MATCH AGAINST expression (MySQL)
func (ep *ExpressionParser) parseMatchAgainstExpr() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse column list
	var columns []*expr.ObjectName
	for {
		col, err := ep.parseObjectName()
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)
		if !ep.parser.ConsumeToken(token.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	if !ep.parser.ParseKeyword("AGAINST") {
		return nil, fmt.Errorf("expected AGAINST in MATCH expression")
	}

	if _, err := ep.parser.ExpectToken(token.TokenLParen{}); err != nil {
		return nil, err
	}

	// Parse match value
	matchValue, err := ep.parseValue()
	if err != nil {
		return nil, err
	}

	// Parse optional search modifier
	var modifier *expr.SearchModifier
	if ep.parser.ParseKeywords([]string{"IN", "NATURAL", "LANGUAGE", "MODE"}) {
		if ep.parser.ParseKeywords([]string{"WITH", "QUERY", "EXPANSION"}) {
			m := expr.SearchNaturalLanguageWithQueryExpansion
			modifier = &m
		} else {
			m := expr.SearchNaturalLanguage
			modifier = &m
		}
	} else if ep.parser.ParseKeywords([]string{"IN", "BOOLEAN", "MODE"}) {
		m := expr.SearchBooleanMode
		modifier = &m
	} else if ep.parser.ParseKeywords([]string{"WITH", "QUERY", "EXPANSION"}) {
		m := expr.SearchWithQueryExpansion
		modifier = &m
	}

	if _, err := ep.parser.ExpectToken(token.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.MatchAgainstExpr{
		SpanVal:           mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Columns:           columns,
		MatchValue:        matchValue,
		OptSearchModifier: modifier,
	}, nil
}

// parseGeometricType parses a PostgreSQL geometric type literal
func (ep *ExpressionParser) parseGeometricType(kind string) (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	val, err := ep.parseValue()
	if err != nil {
		return nil, err
	}

	return &expr.TypedString{
		SpanVal:  mergeSpans(spanStart, val.Span()),
		DataType: kind,
		Value:    val.String(),
	}, nil
}

// parseLBraceExpr parses an expression starting with '{'
// This handles ODBC literals like {d '2025-07-17'}, {t '14:12:01'}, {ts '2025-07-17 14:12:01'}
// and falls back to dictionary literal syntax if not an ODBC literal.
func (ep *ExpressionParser) parseLBraceExpr() (expr.Expr, error) {
	dialect := ep.parser.GetDialect()

	// Try to parse as ODBC literal first
	if odbcExpr, ok := ep.tryParseOdbcLiteral(); ok {
		return odbcExpr, nil
	}

	// Fall back to dictionary expression if supported
	if dialects.SupportsDictionarySyntax(dialect) {
		// Put back the '{' token since parseDictionaryExpr expects to consume it
		// This matches the Rust behavior: self.prev_token(); return self.parse_dictionary();
		ep.parser.PrevToken()
		return ep.parseDictionaryExpr()
	}

	return nil, fmt.Errorf("expected expression, found: %s", ep.parser.GetCurrentToken().Token.String())
}

// tryParseOdbcLiteral tries to parse an ODBC literal (datetime or function)
// Returns (expr, true) if successful, (nil, false) if not an ODBC literal
func (ep *ExpressionParser) tryParseOdbcLiteral() (expr.Expr, bool) {
	// Try ODBC function body first: {fn function_name(args)}
	if fnExpr, ok := ep.tryParseOdbcFnBody(); ok {
		// Expect closing '}'
		if _, err := ep.parser.ExpectToken(token.TokenRBrace{}); err != nil {
			return nil, false
		}
		return fnExpr, true
	}

	// Try ODBC datetime literal: {d '...'}, {t '...'}, {ts '...'}
	if dtExpr, ok := ep.tryParseOdbcDatetime(); ok {
		// Expect closing '}'
		if _, err := ep.parser.ExpectToken(token.TokenRBrace{}); err != nil {
			return nil, false
		}
		return dtExpr, true
	}

	return nil, false
}

// tryParseOdbcFnBody tries to parse {fn function_name(args)}
func (ep *ExpressionParser) tryParseOdbcFnBody() (expr.Expr, bool) {
	// Check for FN keyword (case insensitive)
	if !ep.parser.ParseKeyword("FN") {
		return nil, false
	}

	// Parse function name
	fnName, err := ep.parseObjectName()
	if err != nil {
		return nil, false
	}

	// Parse function arguments
	fnExpr, err := ep.parseFunctionWithName(fnName)
	if err != nil {
		return nil, false
	}

	return fnExpr, true
}

// tryParseOdbcDatetime tries to parse {d '...'}, {t '...'}, or {ts '...'}
func (ep *ExpressionParser) tryParseOdbcDatetime() (expr.Expr, bool) {
	spanStart := ep.parser.GetCurrentToken().Span

	// Peek at the next token to see if it's d, t, or ts
	// Note: These are not keywords, they're literal letters
	nextTok := ep.parser.PeekTokenRef()
	wordTok, ok := nextTok.Token.(token.TokenWord)
	if !ok {
		return nil, false
	}

	// Convert to lowercase string for comparison (like Rust does)
	wordStr := wordTok.Word.Value
	var dataType string
	switch wordStr {
	case "t":
		dataType = "TIME"
	case "d":
		dataType = "DATE"
	case "ts":
		dataType = "TIMESTAMP"
	default:
		return nil, false
	}

	// Consume the word token
	ep.parser.NextToken()

	// Parse the string literal value
	val, err := ep.parseValue()
	if err != nil {
		return nil, false
	}

	// Extract the raw string value from the ValueExpr
	var rawValue string
	if valExpr, ok := val.(*expr.ValueExpr); ok {
		switch v := valExpr.Value.(type) {
		case token.TokenSingleQuotedString:
			rawValue = v.Value
		case token.TokenDoubleQuotedString:
			rawValue = v.Value
		default:
			rawValue = val.String()
		}
	} else {
		rawValue = val.String()
	}

	return &expr.TypedString{
		SpanVal:        mergeSpans(spanStart, val.Span()),
		DataType:       dataType,
		Value:          rawValue,
		UsesOdbcSyntax: true,
	}, true
}
