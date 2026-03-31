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

package tokenizer

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/user/sqlparser/errors"
	"github.com/user/sqlparser/span"
)

// Dialect defines the interface for SQL dialects.
// The tokenizer uses this interface to handle dialect-specific behavior.
type Dialect interface {
	// is_identifier_start returns true if the character can start an identifier
	IsIdentifierStart(ch rune) bool
	// is_identifier_part returns true if the character can be part of an identifier
	IsIdentifierPart(ch rune) bool
	// IsDelimitedIdentifierStart returns true if the character starts a delimited identifier
	IsDelimitedIdentifierStart(ch rune) bool
	// IsNestedDelimitedIdentifierStart returns true for nested quote support
	IsNestedDelimitedIdentifierStart(ch rune) bool
	// PeekNestedDelimitedIdentifierQuotes checks for nested delimited identifiers
	PeekNestedDelimitedIdentifierQuotes(chars string) (startQuote byte, nestedQuote byte, ok bool)
	// SupportsStringLiteralBackslashEscape returns true if backslash escape in strings is supported
	SupportsStringLiteralBackslashEscape() bool
	// SupportsTripleQuotedString returns true if triple-quoted strings are supported
	SupportsTripleQuotedString() bool
	// SupportsDollarQuotedString returns true if dollar-quoted strings are supported
	SupportsDollarQuotedString() bool
	// SupportsDollarPlaceholder returns true if dollar placeholders are supported
	SupportsDollarPlaceholder() bool
	// SupportsNumericLiteralUnderscores returns true if underscores in numbers are supported
	SupportsNumericLiteralUnderscores() bool
	// SupportsNumericPrefix returns true if identifiers can start with numbers
	SupportsNumericPrefix() bool
	// SupportsNestedComments returns true if nested comments are supported
	SupportsNestedComments() bool
	// SupportsMultilineCommentHints returns true if comment hints like /*!...*/ are supported
	SupportsMultilineCommentHints() bool
	// SupportsQuoteDelimitedString returns true if quote-delimited strings (Oracle-style) are supported
	SupportsQuoteDelimitedString() bool
	// SupportsStringEscapeConstant returns true if E'...' escape strings are supported
	SupportsStringEscapeConstant() bool
	// SupportsUnicodeStringLiteral returns true if U&'...' Unicode strings are supported
	SupportsUnicodeStringLiteral() bool
	// SupportsGeometricTypes returns true if geometric operators are supported
	SupportsGeometricTypes() bool
	// SupportsPipeOperator returns true if the |> pipe operator is supported
	SupportsPipeOperator() bool
	// IgnoresWildcardEscapes returns true if LIKE wildcard escapes should be ignored
	IgnoresWildcardEscapes() bool
	// RequiresSingleLineCommentWhitespace returns true if -- requires whitespace
	RequiresSingleLineCommentWhitespace() bool
	// IsCustomOperatorPart returns true if the character can be part of a custom operator
	IsCustomOperatorPart(ch rune) bool
	// IsBigQueryDialect returns true if this is a BigQuery dialect
	IsBigQueryDialect() bool
	// IsPostgreSqlDialect returns true if this is a PostgreSQL dialect
	IsPostgreSqlDialect() bool
	// IsMySqlDialect returns true if this is a MySQL dialect
	IsMySqlDialect() bool
	// IsSnowflakeDialect returns true if this is a Snowflake dialect
	IsSnowflakeDialect() bool
	// IsDuckDbDialect returns true if this is a DuckDB dialect
	IsDuckDbDialect() bool
	// IsGenericDialect returns true if this is the generic dialect
	IsGenericDialect() bool
	// IsHiveDialect returns true if this is a Hive dialect
	IsHiveDialect() bool
}

// GenericDialect provides a default implementation of Dialect
type GenericDialect struct{}

func (d *GenericDialect) IsIdentifierStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_'
}

func (d *GenericDialect) IsIdentifierPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$'
}

func (d *GenericDialect) IsDelimitedIdentifierStart(ch rune) bool {
	return ch == '"'
}

func (d *GenericDialect) IsNestedDelimitedIdentifierStart(ch rune) bool {
	return false
}

func (d *GenericDialect) PeekNestedDelimitedIdentifierQuotes(chars string) (byte, byte, bool) {
	return 0, 0, false
}

func (d *GenericDialect) SupportsStringLiteralBackslashEscape() bool {
	return false
}

func (d *GenericDialect) SupportsTripleQuotedString() bool {
	return false
}

func (d *GenericDialect) SupportsDollarQuotedString() bool {
	return true
}

func (d *GenericDialect) SupportsDollarPlaceholder() bool {
	return false
}

func (d *GenericDialect) SupportsNumericLiteralUnderscores() bool {
	return false
}

func (d *GenericDialect) SupportsNumericPrefix() bool {
	return false
}

func (d *GenericDialect) SupportsNestedComments() bool {
	return false
}

func (d *GenericDialect) SupportsMultilineCommentHints() bool {
	return false
}

func (d *GenericDialect) SupportsQuoteDelimitedString() bool {
	return false
}

func (d *GenericDialect) SupportsStringEscapeConstant() bool {
	return false
}

func (d *GenericDialect) SupportsUnicodeStringLiteral() bool {
	return false
}

func (d *GenericDialect) SupportsGeometricTypes() bool {
	return false
}

func (d *GenericDialect) SupportsPipeOperator() bool {
	return false
}

func (d *GenericDialect) IgnoresWildcardEscapes() bool {
	return false
}

func (d *GenericDialect) RequiresSingleLineCommentWhitespace() bool {
	return false
}

func (d *GenericDialect) IsCustomOperatorPart(ch rune) bool {
	return false
}

func (d *GenericDialect) IsBigQueryDialect() bool   { return false }
func (d *GenericDialect) IsPostgreSqlDialect() bool { return false }
func (d *GenericDialect) IsMySqlDialect() bool      { return false }
func (d *GenericDialect) IsSnowflakeDialect() bool  { return false }
func (d *GenericDialect) IsDuckDbDialect() bool     { return false }
func (d *GenericDialect) IsGenericDialect() bool    { return true }
func (d *GenericDialect) IsHiveDialect() bool       { return false }

// Tokenizer converts SQL text into a sequence of tokens
type Tokenizer struct {
	dialect  Dialect
	query    string
	unescape bool // whether to unescape quoted strings
}

// NewTokenizer creates a new SQL tokenizer for the specified SQL statement
func NewTokenizer(dialect Dialect, query string) *Tokenizer {
	return &Tokenizer{
		dialect:  dialect,
		query:    query,
		unescape: true,
	}
}

// WithUnescape sets the unescape mode
func (t *Tokenizer) WithUnescape(unescape bool) *Tokenizer {
	t.unescape = unescape
	return t
}

// Tokenize converts the SQL query into a slice of tokens
func (t *Tokenizer) Tokenize() ([]Token, error) {
	tokens, err := t.TokenizeWithSpan()
	if err != nil {
		return nil, err
	}
	result := make([]Token, len(tokens))
	for i, tws := range tokens {
		result[i] = tws.Token
	}
	return result, nil
}

// TokenizeWithSpan converts the SQL query into a slice of tokens with location information
func (t *Tokenizer) TokenizeWithSpan() ([]TokenWithSpan, error) {
	var tokens []TokenWithSpan
	err := t.TokenizeIntoBuf(&tokens)
	return tokens, err
}

// TokenizeIntoBuf tokenizes the query and appends tokens to the provided buffer
func (t *Tokenizer) TokenizeIntoBuf(buf *[]TokenWithSpan) error {
	return t.TokenizeIntoBufWithMapper(buf, func(tws TokenWithSpan) TokenWithSpan { return tws })
}

// TokenizeIntoBufWithMapper tokenizes with a custom mapper function
func (t *Tokenizer) TokenizeIntoBufWithMapper(buf *[]TokenWithSpan, mapper func(TokenWithSpan) TokenWithSpan) error {
	state := NewState(t.query)

	location := state.Location()
	var prevToken Token

	for !state.IsEOF() {
		token, err := t.NextToken(state, prevToken)
		if err != nil {
			return err
		}
		if token == nil {
			break
		}

		spanVal := span.NewSpan(location, state.Location())

		// Handle multiline comment hints
		if ws, ok := token.(TokenWhitespace); ok && ws.Whitespace.Type == MultiLineComment {
			if t.dialect.SupportsMultilineCommentHints() && strings.HasPrefix(ws.Whitespace.Content, "!") {
				err := t.tokenizeCommentHints(ws.Whitespace.Content, spanVal, buf, mapper)
				if err != nil {
					return err
				}
			} else {
				*buf = append(*buf, mapper(TokenWithSpan{Token: token, Span: spanVal}))
			}
		} else {
			*buf = append(*buf, mapper(TokenWithSpan{Token: token, Span: spanVal}))
		}

		prevToken = token
		location = state.Location()
	}

	return nil
}

// tokenizeCommentHints re-tokenizes optimizer hints from a multiline comment
func (t *Tokenizer) tokenizeCommentHints(comment string, spanVal span.Span, buf *[]TokenWithSpan, mapper func(TokenWithSpan) TokenWithSpan) error {
	// Strip the leading '!' and any version digits (e.g., "50110")
	hintContent := strings.TrimLeftFunc(comment[1:], func(r rune) bool {
		return r >= '0' && r <= '9'
	})

	if hintContent == "" {
		return nil
	}

	// Create a new tokenizer for the hint content
	innerTokenizer := NewTokenizer(t.dialect, hintContent).WithUnescape(t.unescape)
	innerState := NewState(hintContent)

	location := spanVal.Start
	var prevToken Token

	for !innerState.IsEOF() {
		tok, err := innerTokenizer.NextToken(innerState, prevToken)
		if err != nil {
			return err
		}
		if tok == nil {
			break
		}

		tokenSpan := span.NewSpan(location, innerState.Location())
		// Adjust span to match original position
		tokenSpan.Start.Line += spanVal.Start.Line - 1
		tokenSpan.End.Line += spanVal.Start.Line - 1

		*buf = append(*buf, mapper(TokenWithSpan{Token: tok, Span: tokenSpan}))
		prevToken = tok
		location = innerState.Location()
	}

	return nil
}

// NextToken reads the next token from the input
func (t *Tokenizer) NextToken(state *State, prevToken Token) (Token, error) {
	ch, ok := state.Peek()
	if !ok {
		return nil, nil
	}

	switch ch {
	case ' ':
		state.Next()
		return TokenWhitespace{Whitespace{Type: Space}}, nil
	case '\t':
		state.Next()
		return TokenWhitespace{Whitespace{Type: Tab}}, nil
	case '\n':
		state.Next()
		return TokenWhitespace{Whitespace{Type: Newline}}, nil
	case '\r':
		state.Next()
		if next, ok := state.Peek(); ok && next == '\n' {
			state.Next()
		}
		return TokenWhitespace{Whitespace{Type: Newline}}, nil
	case 'B', 'b':
		return t.tokenizeByteStringLiteral(state)
	case 'R', 'r':
		return t.tokenizeRawStringLiteral(state)
	case 'N', 'n':
		return t.tokenizeNationalStringLiteral(state)
	case 'Q', 'q':
		return t.tokenizeQuoteDelimitedLiteral(state)
	case 'E', 'e':
		return t.tokenizeEscapedStringLiteral(state)
	case 'U', 'u':
		return t.tokenizeUnicodeStringLiteral(state)
	case 'X', 'x':
		return t.tokenizeHexStringLiteral(state)
	case '\'':
		return t.tokenizeSingleQuotedString(state)
	case '"':
		if !t.dialect.IsDelimitedIdentifierStart(ch) {
			return t.tokenizeDoubleQuotedString(state)
		}
		return t.tokenizeQuotedIdentifier(state)
	case '`', '[':
		return t.tokenizeQuotedIdentifier(state)
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '.':
		return t.tokenizeNumberOrPeriod(state, prevToken)
	case '(':
		state.Next()
		return TokenLParen{}, nil
	case ')':
		state.Next()
		return TokenRParen{}, nil
	case ',':
		state.Next()
		return TokenComma{}, nil
	case '-':
		return t.tokenizeMinus(state)
	case '/':
		return t.tokenizeSlash(state)
	case '+':
		state.Next()
		return TokenPlus{}, nil
	case '*':
		state.Next()
		return TokenMul{}, nil
	case '%':
		return t.tokenizePercent(state)
	case '|':
		return t.tokenizePipe(state)
	case '=':
		return t.tokenizeEquals(state)
	case '!':
		return t.tokenizeExclamation(state)
	case '<':
		return t.tokenizeLessThan(state)
	case '>':
		return t.tokenizeGreaterThan(state)
	case ':':
		return t.tokenizeColon(state)
	case ';':
		state.Next()
		return TokenSemiColon{}, nil
	case '\\':
		state.Next()
		return TokenBackslash{}, nil
	case ']':
		state.Next()
		return TokenRBracket{}, nil
	case '&':
		return t.tokenizeAmpersand(state)
	case '^':
		return t.tokenizeCaret(state)
	case '{':
		state.Next()
		return TokenLBrace{}, nil
	case '}':
		state.Next()
		return TokenRBrace{}, nil
	case '#':
		return t.tokenizeHash(state)
	case '~':
		return t.tokenizeTilde(state)
	case '@':
		return t.tokenizeAtSign(state)
	case '?':
		return t.tokenizeQuestion(state)
	case '$':
		return t.tokenizeDollar(state)
	default:
		if unicode.IsSpace(ch) {
			state.Next()
			return TokenWhitespace{Whitespace{Type: Space}}, nil
		}
		if t.dialect.IsIdentifierStart(ch) {
			return t.tokenizeIdentifier(state)
		}
		state.Next()
		return TokenChar{Char: ch}, nil
	}
}

func (t *Tokenizer) tokenizeByteStringLiteral(state *State) (Token, error) {
	ch, _ := state.Peek()
	if !((t.dialect.IsBigQueryDialect() || t.dialect.IsPostgreSqlDialect() || t.dialect.IsMySqlDialect() || t.dialect.IsGenericDialect()) && (ch == 'B' || ch == 'b')) {
		return t.tokenizeIdentifier(state)
	}

	state.Next() // consume B/b

	next, ok := state.Peek()
	if !ok {
		return MakeWord(string(ch), nil), nil
	}

	switch next {
	case '\'':
		if t.dialect.SupportsTripleQuotedString() {
			return t.tokenizeSingleOrTripleQuotedString(state, '\'', false,
				func(s string) Token { return TokenSingleQuotedByteStringLiteral{Value: s} },
				func(s string) Token { return TokenTripleSingleQuotedByteStringLiteral{Value: s} })
		}
		s, err := t.tokenizeSingleQuotedStringLiteral(state, '\'', false)
		if err != nil {
			return nil, err
		}
		return TokenSingleQuotedByteStringLiteral{Value: s}, nil
	case '"':
		if t.dialect.SupportsTripleQuotedString() {
			return t.tokenizeSingleOrTripleQuotedString(state, '"', false,
				func(s string) Token { return TokenDoubleQuotedByteStringLiteral{Value: s} },
				func(s string) Token { return TokenTripleDoubleQuotedByteStringLiteral{Value: s} })
		}
		s, err := t.tokenizeSingleQuotedStringLiteral(state, '"', false)
		if err != nil {
			return nil, err
		}
		return TokenDoubleQuotedByteStringLiteral{Value: s}, nil
	default:
		word := t.tokenizeWord(state, string(ch))
		return MakeWord(word, nil), nil
	}
}

func (t *Tokenizer) tokenizeRawStringLiteral(state *State) (Token, error) {
	ch, _ := state.Peek()
	if !(t.dialect.IsBigQueryDialect() || t.dialect.IsGenericDialect()) && (ch == 'R' || ch == 'r') {
		return t.tokenizeIdentifier(state)
	}

	state.Next() // consume R/r

	next, ok := state.Peek()
	if !ok {
		return MakeWord(string(ch), nil), nil
	}

	switch next {
	case '\'':
		return t.tokenizeSingleOrTripleQuotedString(state, '\'', false,
			func(s string) Token { return TokenSingleQuotedRawStringLiteral{Value: s} },
			func(s string) Token { return TokenTripleSingleQuotedRawStringLiteral{Value: s} })
	case '"':
		return t.tokenizeSingleOrTripleQuotedString(state, '"', false,
			func(s string) Token { return TokenDoubleQuotedRawStringLiteral{Value: s} },
			func(s string) Token { return TokenTripleDoubleQuotedRawStringLiteral{Value: s} })
	default:
		word := t.tokenizeWord(state, string(ch))
		return MakeWord(word, nil), nil
	}
}

func (t *Tokenizer) tokenizeNationalStringLiteral(state *State) (Token, error) {
	ch, _ := state.Peek()
	state.Next() // consume N/n
	charStr := string(ch)

	next, ok := state.Peek()
	if !ok {
		return MakeWord(charStr, nil), nil
	}

	if next == '\'' {
		s, err := t.tokenizeSingleQuotedStringLiteral(state, '\'', t.dialect.SupportsStringLiteralBackslashEscape())
		if err != nil {
			return nil, err
		}
		return TokenNationalStringLiteral{Value: s}, nil
	}

	if (next == 'Q' || next == 'q') && t.dialect.SupportsQuoteDelimitedString() {
		state.Next() // consume Q/q
		if n, ok := state.Peek(); ok && n == '\'' {
			qs, err := t.tokenizeQuoteDelimitedString(state, []rune{ch, next})
			if err != nil {
				return nil, err
			}
			return TokenNationalQuoteDelimitedStringLiteral{QuoteDelimitedString: qs}, nil
		}
		word := t.tokenizeWord(state, charStr+string(next))
		return MakeWord(word, nil), nil
	}

	word := t.tokenizeWord(state, charStr)
	return MakeWord(word, nil), nil
}

func (t *Tokenizer) tokenizeQuoteDelimitedLiteral(state *State) (Token, error) {
	if !t.dialect.SupportsQuoteDelimitedString() {
		return t.tokenizeIdentifier(state)
	}

	state.Next() // consume Q/q

	next, ok := state.Peek()
	if !ok || next != '\'' {
		word := t.tokenizeWord(state, "Q")
		return MakeWord(word, nil), nil
	}

	qs, err := t.tokenizeQuoteDelimitedString(state, []rune{'Q'})
	if err != nil {
		return nil, err
	}
	return TokenQuoteDelimitedStringLiteral{QuoteDelimitedString: qs}, nil
}

func (t *Tokenizer) tokenizeEscapedStringLiteral(state *State) (Token, error) {
	if !t.dialect.SupportsStringEscapeConstant() {
		return t.tokenizeIdentifier(state)
	}

	state.Next() // consume E/e

	next, ok := state.Peek()
	if !ok || next != '\'' {
		word := t.tokenizeWord(state, "E")
		return MakeWord(word, nil), nil
	}

	s, err := t.tokenizeEscapedSingleQuotedString(state)
	if err != nil {
		return nil, err
	}
	return TokenEscapedStringLiteral{Value: s}, nil
}

func (t *Tokenizer) tokenizeUnicodeStringLiteral(state *State) (Token, error) {
	if !t.dialect.SupportsUnicodeStringLiteral() {
		return t.tokenizeIdentifier(state)
	}

	state.Next() // consume U/u

	next, ok := state.Peek()
	if !ok || next != '&' {
		word := t.tokenizeWord(state, "U")
		return MakeWord(word, nil), nil
	}

	// Look ahead for '&'
	if n, ok := state.PeekN(1); ok && n == '\'' {
		state.Next() // consume '&'
		s, err := t.unescapeUnicodeSingleQuotedString(state)
		if err != nil {
			return nil, err
		}
		return TokenUnicodeStringLiteral{Value: s}, nil
	}

	word := t.tokenizeWord(state, "U")
	return MakeWord(word, nil), nil
}

func (t *Tokenizer) tokenizeHexStringLiteral(state *State) (Token, error) {
	state.Next() // consume X/x

	next, ok := state.Peek()
	if !ok || next != '\'' {
		word := t.tokenizeWord(state, "X")
		return MakeWord(word, nil), nil
	}

	s, err := t.tokenizeSingleQuotedStringLiteral(state, '\'', true)
	if err != nil {
		return nil, err
	}
	return TokenHexStringLiteral{Value: s}, nil
}

func (t *Tokenizer) tokenizeSingleQuotedString(state *State) (Token, error) {
	if t.dialect.SupportsTripleQuotedString() {
		return t.tokenizeSingleOrTripleQuotedString(state, '\'', t.dialect.SupportsStringLiteralBackslashEscape(),
			func(s string) Token { return TokenSingleQuotedString{Value: s} },
			func(s string) Token { return TokenTripleSingleQuotedString{Value: s} })
	}
	s, err := t.tokenizeSingleQuotedStringLiteral(state, '\'', t.dialect.SupportsStringLiteralBackslashEscape())
	if err != nil {
		return nil, err
	}
	return TokenSingleQuotedString{Value: s}, nil
}

func (t *Tokenizer) tokenizeDoubleQuotedString(state *State) (Token, error) {
	if t.dialect.SupportsTripleQuotedString() {
		return t.tokenizeSingleOrTripleQuotedString(state, '"', t.dialect.SupportsStringLiteralBackslashEscape(),
			func(s string) Token { return TokenDoubleQuotedString{Value: s} },
			func(s string) Token { return TokenTripleDoubleQuotedString{Value: s} })
	}
	s, err := t.tokenizeSingleQuotedStringLiteral(state, '"', t.dialect.SupportsStringLiteralBackslashEscape())
	if err != nil {
		return nil, err
	}
	return TokenDoubleQuotedString{Value: s}, nil
}

func (t *Tokenizer) tokenizeQuotedIdentifier(state *State) (Token, error) {
	ch, _ := state.Peek()

	if t.dialect.IsNestedDelimitedIdentifierStart(ch) {
		if start, nested, ok := t.dialect.PeekNestedDelimitedIdentifierQuotes(state.Remaining()); ok {
			return t.tokenizeNestedQuotedIdentifier(state, start, nested)
		}
	}

	word, err := t.tokenizeQuotedIdentifierLiteral(state, byte(ch))
	if err != nil {
		return nil, err
	}

	quoteStyle := byte(ch)
	return MakeWord(word, &quoteStyle), nil
}

func (t *Tokenizer) tokenizeNestedQuotedIdentifier(state *State, startQuote byte, nestedQuote byte) (Token, error) {
	state.Next() // skip opening quote
	state.SkipWhile(unicode.IsSpace)

	if ch, ok := state.Peek(); !ok || byte(ch) != nestedQuote {
		return nil, errors.NewTokenizerError(fmt.Sprintf("Expected nested delimiter '%c' before EOF.", nestedQuote), state.Location())
	}

	var parts []string
	parts = append(parts, string(nestedQuote))

	nestedEnd := matchingEndQuote(nestedQuote)
	nestedContent, err := t.tokenizeQuotedIdentifierLiteral(state, nestedEnd)
	if err != nil {
		return nil, err
	}

	parts = append(parts, nestedContent, string(nestedEnd))
	state.SkipWhile(unicode.IsSpace)

	endQuote := matchingEndQuote(startQuote)
	if ch, ok := state.Peek(); !ok || byte(ch) != endQuote {
		return nil, errors.NewTokenizerError(fmt.Sprintf("Expected close delimiter '%c' before EOF.", endQuote), state.Location())
	}
	state.Next() // consume closing quote

	return MakeWord(strings.Join(parts, ""), &startQuote), nil
}

func (t *Tokenizer) tokenizeNumberOrPeriod(state *State, prevToken Token) (Token, error) {
	ch, _ := state.Peek()

	// Handle ._ case - if previous token was a word, this is a period followed by identifier
	if ch == '.' {
		if next, ok := state.PeekN(1); ok && next == '_' {
			if _, wasWord := prevToken.(TokenWord); wasWord {
				state.Next()
				return TokenPeriod{}, nil
			}
			return nil, errors.NewTokenizerError("Unexpected character '_'", state.Location())
		}
	}

	// Check for number separator support
	isNumberSeparator := func(current, next rune) bool {
		return t.dialect.SupportsNumericLiteralUnderscores() && current == '_' &&
			unicode.IsDigit(next)
	}

	// Consume integer part
	var s strings.Builder
	for {
		ch, ok := state.Peek()
		if !ok {
			break
		}

		nextCh, _ := state.PeekN(1)
		if unicode.IsDigit(ch) || isNumberSeparator(ch, nextCh) {
			state.Next()
			s.WriteRune(ch)
		} else {
			break
		}
	}

	// Check for hex literal 0x
	if s.String() == "0" {
		if next, ok := state.Peek(); ok && next == 'x' {
			state.Next()
			hexStr := state.TakeWhile(func(r rune) bool {
				return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
			})
			return TokenHexStringLiteral{Value: hexStr}, nil
		}
	}

	// Handle decimal point
	if next, ok := state.Peek(); ok && next == '.' {
		s.WriteRune(next)
		state.Next()

		// If the dialect supports numeric prefix and we just have ".", check if it's a period token
		if s.String() == "." && t.dialect.SupportsNumericPrefix() {
			if _, wasWord := prevToken.(TokenWord); wasWord {
				return TokenPeriod{}, nil
			}
		}

		// Consume fractional part
		for {
			ch, ok := state.Peek()
			if !ok {
				break
			}
			nextCh, _ := state.PeekN(1)
			if unicode.IsDigit(ch) || isNumberSeparator(ch, nextCh) {
				state.Next()
				s.WriteRune(ch)
			} else {
				break
			}
		}
	}

	// No fraction -> just a period
	if s.String() == "." {
		return TokenPeriod{}, nil
	}

	// Handle exponent
	exponent := ""
	if next, ok := state.Peek(); ok && (next == 'e' || next == 'E') {
		clone := state.Clone()
		clone.Next()

		// Optional sign
		if sign, ok := clone.Peek(); ok && (sign == '+' || sign == '-') {
			clone.Next()
		}

		// Must have digits
		if digit, ok := clone.Peek(); ok && unicode.IsDigit(digit) {
			state.Next() // consume 'e' or 'E'
			if sign, ok := state.Peek(); ok && (sign == '+' || sign == '-') {
				state.Next()
				exponent += string(sign)
			}
			expDigits := state.TakeWhile(unicode.IsDigit)
			exponent = "e" + exponent + expDigits
			s.WriteString(exponent)
		}
	}

	// Check for numeric prefix identifiers
	if t.dialect.SupportsNumericPrefix() {
		if exponent == "" {
			word := state.TakeWhile(t.dialect.IsIdentifierPart)
			if word != "" {
				return MakeWord(s.String()+word, nil), nil
			}
		} else if prevToken != nil {
			if _, wasPeriod := prevToken.(TokenPeriod); wasPeriod {
				return MakeWord(s.String(), nil), nil
			}
		}
	}

	// Check for 'L' suffix (long)
	isLong := false
	if next, ok := state.Peek(); ok && next == 'L' {
		state.Next()
		isLong = true
	}

	return TokenNumber{Value: s.String(), Long: isLong}, nil
}

func (t *Tokenizer) tokenizeMinus(state *State) (Token, error) {
	state.Next() // consume '-'

	next, ok := state.Peek()
	if !ok {
		return TokenMinus{}, nil
	}

	if next == '-' {
		isComment := true
		if t.dialect.RequiresSingleLineCommentWhitespace() {
			if n, ok := state.PeekN(1); ok {
				isComment = unicode.IsSpace(n)
			} else {
				isComment = false
			}
		}

		if isComment {
			state.Next() // consume second '-'
			comment := t.tokenizeSingleLineComment(state)
			return TokenWhitespace{Whitespace{Type: SingleLineComment, Content: comment, Prefix: "--"}}, nil
		}
	}

	if next == '>' {
		state.Next()
		if n, ok := state.Peek(); ok && n == '>' {
			state.Next()
			return TokenLongArrow{}, nil
		}
		return TokenArrow{}, nil
	}

	return TokenMinus{}, nil
}

func (t *Tokenizer) tokenizeSlash(state *State) (Token, error) {
	state.Next() // consume '/'

	next, ok := state.Peek()
	if !ok {
		return TokenDiv{}, nil
	}

	if next == '*' {
		state.Next() // consume '*'
		return t.tokenizeMultilineComment(state)
	}

	if next == '/' {
		if t.dialect.IsSnowflakeDialect() {
			state.Next()
			comment := t.tokenizeSingleLineComment(state)
			return TokenWhitespace{Whitespace{Type: SingleLineComment, Content: comment, Prefix: "//"}}, nil
		}
		if t.dialect.IsDuckDbDialect() || t.dialect.IsGenericDialect() {
			return TokenDuckIntDiv{}, nil
		}
	}

	return TokenDiv{}, nil
}

func (t *Tokenizer) tokenizePercent(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok || unicode.IsSpace(next) {
		return TokenMod{}, nil
	}

	if t.dialect.IsIdentifierStart('%') {
		return t.tokenizeIdentifierOrKeyword(state, string('%')+string(next))
	}

	return TokenMod{}, nil
}

func (t *Tokenizer) tokenizePipe(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenPipe{}, nil
	}

	if next == '/' {
		state.Next()
		return TokenPGSquareRoot{}, nil
	}

	if next == '|' {
		state.Next()
		if n, ok := state.Peek(); ok && n == '/' {
			state.Next()
			return TokenPGCubeRoot{}, nil
		}
		return TokenStringConcat{}, nil
	}

	if next == '&' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		if n, ok := state.Peek(); ok && n == '>' {
			state.Next()
			return TokenVerticalBarAmpersandRightAngleBracket{}, nil
		}
		return TokenPipe{}, nil // fall back to custom operator handling
	}

	if next == '>' {
		if t.dialect.SupportsGeometricTypes() {
			state.Next()
			if n, ok := state.Peek(); ok && n == '>' {
				state.Next()
				return TokenVerticalBarShiftRight{}, nil
			}
			return TokenPipe{}, nil
		}
		if t.dialect.SupportsPipeOperator() {
			state.Next()
			return TokenVerticalBarRightAngleBracket{}, nil
		}
	}

	return TokenPipe{}, nil
}

func (t *Tokenizer) tokenizeEquals(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenEq{}, nil
	}

	if next == '>' {
		state.Next()
		return TokenRArrow{}, nil
	}

	if next == '=' {
		state.Next()
		return TokenDoubleEq{}, nil
	}

	return TokenEq{}, nil
}

func (t *Tokenizer) tokenizeExclamation(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenExclamationMark{}, nil
	}

	if next == '=' {
		state.Next()
		return TokenNeq{}, nil
	}

	if next == '!' {
		state.Next()
		return TokenDoubleExclamationMark{}, nil
	}

	if next == '~' {
		state.Next()
		if n, ok := state.Peek(); ok && n == '*' {
			state.Next()
			return TokenExclamationMarkTildeAsterisk{}, nil
		}
		if n, ok := state.Peek(); ok && n == '~' {
			state.Next()
			if m, ok := state.Peek(); ok && m == '*' {
				state.Next()
				return TokenExclamationMarkDoubleTildeAsterisk{}, nil
			}
			return TokenExclamationMarkDoubleTilde{}, nil
		}
		return TokenExclamationMarkTilde{}, nil
	}

	return TokenExclamationMark{}, nil
}

func (t *Tokenizer) tokenizeLessThan(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenLt{}, nil
	}

	if next == '=' {
		state.Next()
		if n, ok := state.Peek(); ok && n == '>' {
			state.Next()
			return TokenSpaceship{}, nil
		}
		if n, ok := state.Peek(); ok && (n == '+' || n == '-') {
			// <=+ or <=- are not valid combined operators
			return TokenLtEq{}, nil
		}
		return TokenLtEq{}, nil
	}

	if next == '>' {
		state.Next()
		return TokenNeq{}, nil
	}

	if next == '<' {
		state.Next()
		if t.dialect.SupportsGeometricTypes() {
			if n, ok := state.Peek(); ok && n == '|' {
				state.Next()
				return TokenShiftLeftVerticalBar{}, nil
			}
		}
		return TokenShiftLeft{}, nil
	}

	if next == '+' {
		return TokenLt{}, nil
	}

	if next == '-' && t.dialect.SupportsGeometricTypes() {
		if n, ok := state.PeekN(1); ok && n == '>' {
			state.Next()
			state.Next()
			return TokenTwoWayArrow{}, nil
		}
	}

	if next == '^' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		return TokenLeftAngleBracketCaret{}, nil
	}

	if next == '@' {
		state.Next()
		return TokenArrowAt{}, nil
	}

	return TokenLt{}, nil
}

func (t *Tokenizer) tokenizeGreaterThan(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenGt{}, nil
	}

	if next == '=' {
		state.Next()
		return TokenGtEq{}, nil
	}

	if next == '>' {
		state.Next()
		return TokenShiftRight{}, nil
	}

	if next == '^' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		return TokenRightAngleBracketCaret{}, nil
	}

	return TokenGt{}, nil
}

func (t *Tokenizer) tokenizeColon(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenColon{}, nil
	}

	if next == ':' {
		state.Next()
		return TokenDoubleColon{}, nil
	}

	if next == '=' {
		state.Next()
		return TokenAssignment{}, nil
	}

	return TokenColon{}, nil
}

func (t *Tokenizer) tokenizeHash(state *State) (Token, error) {
	// Check for comment starts in some dialects
	if t.dialect.IsSnowflakeDialect() || t.dialect.IsBigQueryDialect() ||
		t.dialect.IsMySqlDialect() || t.dialect.IsHiveDialect() {
		state.Next()
		comment := t.tokenizeSingleLineComment(state)
		return TokenWhitespace{Whitespace{Type: SingleLineComment, Content: comment, Prefix: "#"}}, nil
	}

	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenSharp{}, nil
	}

	if next == '-' {
		state.Next()
		return TokenHashMinus{}, nil
	}

	if next == '>' {
		state.Next()
		if n, ok := state.Peek(); ok && n == '>' {
			state.Next()
			return TokenHashLongArrow{}, nil
		}
		return TokenHashArrow{}, nil
	}

	if next == ' ' {
		return TokenSharp{}, nil
	}

	if next == '#' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		return TokenDoubleSharp{}, nil
	}

	if t.dialect.IsIdentifierStart('#') {
		return t.tokenizeIdentifierOrKeyword(state, "#"+string(next))
	}

	return TokenSharp{}, nil
}

func (t *Tokenizer) tokenizeAmpersand(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenAmpersand{}, nil
	}

	if next == '>' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		return TokenAmpersandRightAngleBracket{}, nil
	}

	if next == '<' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		n, ok := state.Peek()
		if ok && n == '|' {
			state.Next()
			return TokenAmpersandLeftAngleBracketVerticalBar{}, nil
		}
		return TokenAmpersandLeftAngleBracket{}, nil
	}

	if next == '&' {
		state.Next()
		return TokenOverlap{}, nil
	}

	return TokenAmpersand{}, nil
}

func (t *Tokenizer) tokenizeCaret(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenCaret{}, nil
	}

	if next == '@' {
		state.Next()
		return TokenCaretAt{}, nil
	}

	return TokenCaret{}, nil
}

func (t *Tokenizer) tokenizeTilde(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenTilde{}, nil
	}

	if next == '*' {
		state.Next()
		return TokenTildeAsterisk{}, nil
	}

	if next == '=' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		return TokenTildeEqual{}, nil
	}

	if next == '~' {
		state.Next()
		if n, ok := state.Peek(); ok && n == '*' {
			state.Next()
			return TokenDoubleTildeAsterisk{}, nil
		}
		return TokenDoubleTilde{}, nil
	}

	return TokenTilde{}, nil
}

func (t *Tokenizer) tokenizeAtSign(state *State) (Token, error) {
	state.Next()

	next, ok := state.Peek()
	if !ok {
		return TokenAtSign{}, nil
	}

	if next == '@' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		return TokenAtAt{}, nil
	}

	if next == '-' && t.dialect.SupportsGeometricTypes() {
		state.Next()
		if n, ok := state.Peek(); ok && n == '@' {
			state.Next()
			return TokenAtDashAt{}, nil
		}
		return TokenAtSign{}, nil
	}

	if next == '>' {
		state.Next()
		return TokenAtArrow{}, nil
	}

	if next == '?' {
		state.Next()
		return TokenAtQuestion{}, nil
	}

	if next == ' ' || next == '\'' || next == '"' || next == '`' {
		return TokenAtSign{}, nil
	}

	if t.dialect.IsIdentifierStart('@') {
		return t.tokenizeIdentifierOrKeyword(state, "@"+string(next))
	}

	return TokenAtSign{}, nil
}

func (t *Tokenizer) tokenizeQuestion(state *State) (Token, error) {
	if t.dialect.SupportsGeometricTypes() {
		state.Next()

		next, ok := state.Peek()
		if !ok {
			return TokenQuestion{}, nil
		}

		if next == '|' {
			state.Next()
			if n, ok := state.Peek(); ok && n == '|' {
				state.Next()
				return TokenQuestionMarkDoubleVerticalBar{}, nil
			}
			return TokenQuestionPipe{}, nil
		}

		if next == '&' {
			state.Next()
			return TokenQuestionAnd{}, nil
		}

		if next == '-' {
			state.Next()
			if n, ok := state.Peek(); ok && n == '|' {
				state.Next()
				return TokenQuestionMarkDashVerticalBar{}, nil
			}
			return TokenQuestionMarkDash{}, nil
		}

		if next == '#' {
			state.Next()
			return TokenQuestionMarkSharp{}, nil
		}

		return TokenQuestion{}, nil
	}

	state.Next()
	s := state.TakeWhile(unicode.IsDigit)
	return TokenPlaceholder{Value: "?" + s}, nil
}

func (t *Tokenizer) tokenizeDollar(state *State) (Token, error) {
	state.Next() // consume '$'

	// Check for dollar-quoted string or placeholder
	if next, ok := state.Peek(); ok && next == '$' && !t.dialect.SupportsDollarPlaceholder() {
		state.Next() // consume second '$'

		var s strings.Builder
		var prev rune
		terminated := false

		for {
			ch, ok := state.Peek()
			if !ok {
				break
			}

			if prev == '$' {
				if ch == '$' {
					state.Next()
					terminated = true
					break
				}
				s.WriteRune('$')
				s.WriteRune(ch)
			} else if ch != '$' {
				s.WriteRune(ch)
			}

			prev = ch
			state.Next()
		}

		if !terminated {
			return nil, errors.NewTokenizerError("Unterminated dollar-quoted string", state.Location())
		}

		return TokenDollarQuotedString{DollarQuotedString{Value: s.String(), Tag: nil}}, nil
	}

	// Parse the tag/value
	var value strings.Builder
	for {
		ch, ok := state.Peek()
		if !ok {
			break
		}

		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' ||
			(ch == '$' && t.dialect.SupportsDollarPlaceholder()) {
			state.Next()
			value.WriteRune(ch)
		} else {
			break
		}
	}

	val := value.String()

	// Check for tagged dollar-quoted string
	if next, ok := state.Peek(); ok && next == '$' && !t.dialect.SupportsDollarPlaceholder() {
		state.Next() // consume '$'

		var s strings.Builder
		endDelimiter := "$" + val + "$"

		for {
			ch, ok := state.Next()
			if !ok {
				return nil, errors.NewTokenizerError("Unterminated dollar-quoted, expected $", state.Location())
			}

			s.WriteRune(ch)

			if strings.HasSuffix(s.String(), endDelimiter) {
				content := strings.TrimSuffix(s.String(), endDelimiter)
				tag := val
				return TokenDollarQuotedString{DollarQuotedString{Value: content, Tag: &tag}}, nil
			}
		}
	}

	return TokenPlaceholder{Value: "$" + val}, nil
}

func (t *Tokenizer) tokenizeIdentifier(state *State) (Token, error) {
	ch, _ := state.Peek()
	state.Next()
	word := t.tokenizeWord(state, string(ch))
	return MakeWord(word, nil), nil
}

func (t *Tokenizer) tokenizeIdentifierOrKeyword(state *State, prefix string) (Token, error) {
	state.Next() // consume first char of prefix
	state.Next() // consume second char of prefix
	word := t.tokenizeWord(state, prefix)
	return MakeWord(word, nil), nil
}

func (t *Tokenizer) tokenizeWord(state *State, prefix string) string {
	var s strings.Builder
	s.WriteString(prefix)
	s.WriteString(state.TakeWhile(t.dialect.IsIdentifierPart))
	return s.String()
}

func (t *Tokenizer) tokenizeSingleLineComment(state *State) string {
	var comment strings.Builder
	for {
		ch, ok := state.Peek()
		if !ok {
			break
		}

		if ch == '\n' {
			break
		}

		// PostgreSQL stops at \r
		if ch == '\r' && t.dialect.IsPostgreSqlDialect() {
			break
		}

		state.Next()
		comment.WriteRune(ch)
	}

	// Include the newline/carriage return in the comment
	if ch, ok := state.Peek(); ok && (ch == '\n' || ch == '\r') {
		state.Next()
		comment.WriteRune(ch)
	}

	return comment.String()
}

func (t *Tokenizer) tokenizeMultilineComment(state *State) (Token, error) {
	var s strings.Builder
	nested := 1

	for {
		ch, ok := state.Next()
		if !ok {
			return nil, errors.NewTokenizerError("Unexpected EOF while in a multi-line comment", state.Location())
		}

		if ch == '/' {
			if next, ok := state.Peek(); ok && next == '*' && t.dialect.SupportsNestedComments() {
				state.Next()
				s.WriteString("/*")
				nested++
				continue
			}
		}

		if ch == '*' {
			if next, ok := state.Peek(); ok && next == '/' {
				state.Next()
				nested--
				if nested == 0 {
					return TokenWhitespace{Whitespace{Type: MultiLineComment, Content: s.String()}}, nil
				}
				s.WriteString("*/")
				continue
			}
		}

		s.WriteRune(ch)
	}
}

func (t *Tokenizer) tokenizeSingleQuotedStringLiteral(state *State, quoteChar rune, backslashEscape bool) (string, error) {
	// Consume opening quote
	state.Next()

	var s strings.Builder

	for {
		ch, ok := state.Peek()
		if !ok {
			return "", errors.NewTokenizerError("Unterminated string literal", state.Location())
		}

		state.Next()

		if ch == quoteChar {
			if state.IsEOF() {
				return s.String(), nil
			}

			next, ok := state.Peek()
			if !ok {
				s.WriteRune(ch)
				return s.String(), nil
			}

			if next == quoteChar {
				// Escaped quote
				s.WriteRune(ch)
				if !t.unescape {
					s.WriteRune(ch)
				}
				state.Next()
			} else {
				// End of string
				return s.String(), nil
			}
		} else if ch == '\\' && backslashEscape {
			if t.unescape && !t.dialect.IgnoresWildcardEscapes() {
				next, ok := state.Peek()
				if !ok {
					return "", errors.NewTokenizerError("Unterminated string literal", state.Location())
				}
				state.Next()

				escaped := t.unescapeChar(next)
				s.WriteRune(escaped)
			} else {
				s.WriteRune(ch)
				if next, ok := state.Peek(); ok {
					s.WriteRune(next)
					state.Next()
				}
			}
		} else {
			s.WriteRune(ch)
		}
	}
}

func (t *Tokenizer) tokenizeQuotedIdentifierLiteral(state *State, quoteEnd byte) (string, error) {
	state.Next() // consume opening quote

	var s strings.Builder
	lastChar := rune(0)

	for {
		ch, ok := state.Next()
		if !ok {
			return "", errors.NewTokenizerError(fmt.Sprintf("Expected close delimiter '%c' before EOF.", quoteEnd), state.Location())
		}

		if byte(ch) == quoteEnd {
			if next, ok := state.Peek(); ok && byte(next) == quoteEnd {
				// Escaped quote
				state.Next()
				s.WriteRune(ch)
				if !t.unescape {
					s.WriteRune(ch)
				}
			} else {
				lastChar = ch
				break
			}
		} else {
			s.WriteRune(ch)
		}
	}

	if lastChar == 0 {
		return "", errors.NewTokenizerError(fmt.Sprintf("Expected close delimiter '%c' before EOF.", quoteEnd), state.Location())
	}

	return s.String(), nil
}

func (t *Tokenizer) tokenizeSingleOrTripleQuotedString(state *State, quoteStyle rune, backslashEscape bool, singleQuoteFn func(string) Token, tripleQuoteFn func(string) Token) (Token, error) {
	// Count opening quotes
	numOpeningQuotes := 0
	for i := 0; i < 3; i++ {
		if ch, ok := state.Peek(); ok && ch == quoteStyle {
			state.Next()
			numOpeningQuotes++
		} else {
			break
		}
	}

	switch numOpeningQuotes {
	case 1:
		s, err := t.tokenizeSingleQuotedStringLiteralWithQuotes(state, quoteStyle, backslashEscape, 0)
		if err != nil {
			return nil, err
		}
		return singleQuoteFn(s), nil
	case 2:
		// Empty string
		return singleQuoteFn(""), nil
	case 3:
		s, err := t.tokenizeSingleQuotedStringLiteralWithQuotes(state, quoteStyle, backslashEscape, 3)
		if err != nil {
			return nil, err
		}
		// Strip trailing quotes for triple-quoted strings
		if len(s) >= 2 {
			s = s[:len(s)-2]
		}
		return tripleQuoteFn(s), nil
	default:
		return nil, errors.NewTokenizerError("invalid string literal opening", state.Location())
	}
}

func (t *Tokenizer) tokenizeSingleQuotedStringLiteralWithQuotes(state *State, quoteStyle rune, backslashEscape bool, numOpeningQuotes int) (string, error) {
	// Consume any remaining opening quotes
	for i := 0; i < numOpeningQuotes; i++ {
		if ch, ok := state.Next(); !ok || ch != quoteStyle {
			return "", errors.NewTokenizerError("invalid string literal opening", state.Location())
		}
	}

	var s strings.Builder
	numConsecutiveQuotes := 0

	for {
		ch, ok := state.Peek()
		if !ok {
			return "", errors.NewTokenizerError("Unterminated string literal", state.Location())
		}

		state.Next()

		if ch == quoteStyle {
			numConsecutiveQuotes++

			if numConsecutiveQuotes == 3 {
				// End of triple-quoted string
				// s now has the content plus two trailing quotes
				return s.String(), nil
			}

			if numOpeningQuotes <= 1 {
				// For single-quoted (numOpeningQuotes=0) or when we need to consume 1 more quote (numOpeningQuotes=1),
				// check for escaped quote or end of string
				if next, ok := state.Peek(); ok && next == quoteStyle {
					s.WriteRune(ch)
					if !t.unescape {
						s.WriteRune(ch)
					}
					state.Next()
				} else {
					// End of string
					return s.String(), nil
				}
			} else {
				s.WriteRune(ch)
			}
		} else {
			numConsecutiveQuotes = 0

			if ch == '\\' && backslashEscape {
				if t.unescape && !t.dialect.IgnoresWildcardEscapes() {
					next, ok := state.Peek()
					if !ok {
						return "", errors.NewTokenizerError("Unterminated string literal", state.Location())
					}
					state.Next()

					escaped := t.unescapeChar(next)
					s.WriteRune(escaped)
				} else {
					s.WriteRune(ch)
					if next, ok := state.Peek(); ok {
						s.WriteRune(next)
						state.Next()
					}
				}
			} else {
				s.WriteRune(ch)
			}
		}
	}
}

func (t *Tokenizer) tokenizeQuoteDelimitedString(state *State, prefix []rune) (QuoteDelimitedString, error) {
	startLoc := state.Location()
	state.Next() // consume opening quote

	startQuote, ok := state.Next()
	if !ok || startQuote == ' ' || startQuote == '\t' || startQuote == '\r' || startQuote == '\n' {
		return QuoteDelimitedString{}, errors.NewTokenizerError(fmt.Sprintf("Invalid space, tab, newline, or EOF after '%s''", string(prefix)), startLoc)
	}

	var endQuote byte
	switch startQuote {
	case '[':
		endQuote = ']'
	case '{':
		endQuote = '}'
	case '<':
		endQuote = '>'
	case '(':
		endQuote = ')'
	default:
		endQuote = byte(startQuote)
	}

	var value strings.Builder
	for {
		ch, ok := state.Next()
		if !ok {
			return QuoteDelimitedString{}, errors.NewTokenizerError("Unterminated string literal", startLoc)
		}

		if byte(ch) == endQuote {
			if next, ok := state.Peek(); ok && next == '\'' {
				state.Next()
				return QuoteDelimitedString{
					StartQuote: byte(startQuote),
					Value:      value.String(),
					EndQuote:   endQuote,
				}, nil
			}
		}

		value.WriteRune(ch)
	}
}

func (t *Tokenizer) tokenizeEscapedSingleQuotedString(state *State) (string, error) {
	state.Next() // consume opening quote

	var unescaped strings.Builder
	for {
		ch, ok := state.Next()
		if !ok {
			return "", errors.NewTokenizerError("Unterminated encoded string literal", state.Location())
		}

		if ch == '\'' {
			if next, ok := state.Peek(); ok && next == '\'' {
				state.Next()
				unescaped.WriteRune('\'')
				continue
			}
			return unescaped.String(), nil
		}

		if ch != '\\' {
			unescaped.WriteRune(ch)
			continue
		}

		next, ok := state.Next()
		if !ok {
			return "", errors.NewTokenizerError("Unterminated encoded string literal", state.Location())
		}

		switch next {
		case 'b':
			unescaped.WriteRune('\u0008')
		case 'f':
			unescaped.WriteRune('\u000c')
		case 'n':
			unescaped.WriteRune('\n')
		case 'r':
			unescaped.WriteRune('\r')
		case 't':
			unescaped.WriteRune('\t')
		case 'u':
			r, err := t.unescapeUnicode(state, 4)
			if err != nil {
				return "", err
			}
			unescaped.WriteRune(r)
		case 'U':
			r, err := t.unescapeUnicode(state, 8)
			if err != nil {
				return "", err
			}
			unescaped.WriteRune(r)
		case 'x':
			r, err := t.unescapeHex(state)
			if err != nil {
				return "", err
			}
			unescaped.WriteRune(r)
		default:
			if next >= '0' && next <= '7' {
				r, err := t.unescapeOctal(state, next)
				if err != nil {
					return "", err
				}
				unescaped.WriteRune(r)
			} else {
				unescaped.WriteRune(next)
			}
		}
	}
}

func (t *Tokenizer) unescapeUnicode(state *State, numDigits int) (rune, error) {
	var s strings.Builder
	for i := 0; i < numDigits; i++ {
		ch, ok := state.Next()
		if !ok {
			return 0, errors.NewTokenizerError(fmt.Sprintf("Unexpected EOF while parsing unicode escape sequence"), state.Location())
		}
		s.WriteRune(ch)
	}

	n, err := strconv.ParseInt(s.String(), 16, 32)
	if err != nil {
		return 0, errors.NewTokenizerError(fmt.Sprintf("Invalid unicode escape sequence: %s", s.String()), state.Location())
	}

	if n == 0 {
		return 0, errors.NewTokenizerError("Invalid null character in escape sequence", state.Location())
	}

	r := rune(n)
	if !utf8.ValidRune(r) {
		return 0, errors.NewTokenizerError(fmt.Sprintf("Invalid unicode character: %x", n), state.Location())
	}

	return r, nil
}

func (t *Tokenizer) unescapeHex(state *State) (rune, error) {
	var s strings.Builder
	for i := 0; i < 2; i++ {
		ch, ok := state.Peek()
		if !ok || !((ch >= '0' && ch <= '9') || (ch >= 'a' && ch <= 'f') || (ch >= 'A' && ch <= 'F')) {
			break
		}
		state.Next()
		s.WriteRune(ch)
	}

	if s.Len() == 0 {
		return 'x', nil
	}

	n, err := strconv.ParseInt(s.String(), 16, 32)
	if err != nil {
		return 0, err
	}

	n = n & 0xFF
	if n > 127 {
		return 0, errors.NewTokenizerError(fmt.Sprintf("Invalid hex escape: %s", s.String()), state.Location())
	}

	if n == 0 {
		return 0, errors.NewTokenizerError("Invalid null character in escape sequence", state.Location())
	}

	return rune(n), nil
}

func (t *Tokenizer) unescapeOctal(state *State, first rune) (rune, error) {
	var s strings.Builder
	s.WriteRune(first)

	for i := 0; i < 2; i++ {
		ch, ok := state.Peek()
		if !ok || ch < '0' || ch > '7' {
			break
		}
		state.Next()
		s.WriteRune(ch)
	}

	n, err := strconv.ParseInt(s.String(), 8, 32)
	if err != nil {
		return 0, err
	}

	n = n & 0xFF
	if n > 127 {
		return 0, errors.NewTokenizerError(fmt.Sprintf("Invalid octal escape: %s", s.String()), state.Location())
	}

	if n == 0 {
		return 0, errors.NewTokenizerError("Invalid null character in escape sequence", state.Location())
	}

	return rune(n), nil
}

func (t *Tokenizer) unescapeChar(ch rune) rune {
	switch ch {
	case '0':
		return '\u0000'
	case 'a':
		return '\u0007'
	case 'b':
		return '\u0008'
	case 'f':
		return '\u000c'
	case 'n':
		return '\n'
	case 'r':
		return '\r'
	case 't':
		return '\t'
	case 'Z':
		return '\u001a'
	default:
		return ch
	}
}

func (t *Tokenizer) unescapeUnicodeSingleQuotedString(state *State) (string, error) {
	state.Next() // consume opening quote

	var unescaped strings.Builder
	for {
		ch, ok := state.Next()
		if !ok {
			return "", errors.NewTokenizerError("Unterminated unicode encoded string literal", state.Location())
		}

		if ch == '\'' {
			if next, ok := state.Peek(); ok && next == '\'' {
				state.Next()
				unescaped.WriteRune('\'')
				continue
			}
			return unescaped.String(), nil
		}

		if ch == '\\' {
			next, ok := state.Peek()
			if !ok {
				return "", errors.NewTokenizerError("Unterminated unicode encoded string literal", state.Location())
			}

			if next == '\\' {
				state.Next()
				unescaped.WriteRune('\\')
			} else if next == '+' {
				state.Next()
				r, err := t.takeCharFromHexDigits(state, 6)
				if err != nil {
					return "", err
				}
				unescaped.WriteRune(r)
			} else {
				r, err := t.takeCharFromHexDigits(state, 4)
				if err != nil {
					return "", err
				}
				unescaped.WriteRune(r)
			}
		} else {
			unescaped.WriteRune(ch)
		}
	}
}

func (t *Tokenizer) takeCharFromHexDigits(state *State, maxDigits int) (rune, error) {
	var result uint32
	for i := 0; i < maxDigits; i++ {
		ch, ok := state.Next()
		if !ok {
			return 0, errors.NewTokenizerError("Unexpected EOF while parsing hex digit in escaped unicode string", state.Location())
		}

		var digit uint32
		switch {
		case ch >= '0' && ch <= '9':
			digit = uint32(ch - '0')
		case ch >= 'a' && ch <= 'f':
			digit = uint32(ch - 'a' + 10)
		case ch >= 'A' && ch <= 'F':
			digit = uint32(ch - 'A' + 10)
		default:
			return 0, errors.NewTokenizerError(fmt.Sprintf("Invalid hex digit in escaped unicode string: %c", ch), state.Location())
		}

		result = result*16 + digit
	}

	r := rune(result)
	if !utf8.ValidRune(r) {
		return 0, errors.NewTokenizerError(fmt.Sprintf("Invalid unicode character: %x", result), state.Location())
	}

	return r, nil
}
