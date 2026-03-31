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

	"github.com/user/sqlparser/errors"
	"github.com/user/sqlparser/span"
	"github.com/user/sqlparser/tokenizer"
)

// ============================================================
// Token Stream Management
// ============================================================

// PeekToken returns a copy of the first non-whitespace token that has not yet been processed,
// or Token::EOF if at the end of the stream.
func (p *Parser) PeekToken() tokenizer.TokenWithSpan {
	return p.PeekNthToken(0)
}

// PeekTokenRef returns a reference to the first non-whitespace token that has not yet
// been processed, or Token::EOF if at the end of the stream.
// This is more efficient than PeekToken when you don't need a copy.
func (p *Parser) PeekTokenRef() *tokenizer.TokenWithSpan {
	return p.PeekNthTokenRef(0)
}

// PeekNthToken returns the nth non-whitespace token that has not yet been processed.
func (p *Parser) PeekNthToken(n int) tokenizer.TokenWithSpan {
	return *p.PeekNthTokenRef(n)
}

// PeekNthTokenRef returns a reference to the nth non-whitespace token that has not yet been processed.
func (p *Parser) PeekNthTokenRef(n int) *tokenizer.TokenWithSpan {
	index := p.index
	for {
		if index >= len(p.tokens) {
			return &eofToken
		}
		tok := &p.tokens[index]
		if _, isWhitespace := tok.Token.(tokenizer.TokenWhitespace); isWhitespace {
			index++
			continue
		}
		if n == 0 {
			return tok
		}
		n--
		index++
	}
}

// PeekTokenNoSkip returns the first token, possibly whitespace, that has not yet been processed.
func (p *Parser) PeekTokenNoSkip() tokenizer.TokenWithSpan {
	return p.PeekNthTokenNoSkip(0)
}

// PeekNthTokenNoSkip returns the nth token, possibly whitespace, that has not yet been processed.
func (p *Parser) PeekNthTokenNoSkip(n int) tokenizer.TokenWithSpan {
	if p.index+n >= len(p.tokens) {
		return eofToken
	}
	return p.tokens[p.index+n]
}

// NextToken advances to the next non-whitespace token and returns a copy.
func (p *Parser) NextToken() tokenizer.TokenWithSpan {
	p.AdvanceToken()
	return *p.GetCurrentToken()
}

// NextTokenNoSkip advances one token and returns the token, possibly whitespace.
// Returns nil if reached end-of-file.
func (p *Parser) NextTokenNoSkip() *tokenizer.TokenWithSpan {
	p.index++
	if p.index-1 >= len(p.tokens) {
		return nil
	}
	return &p.tokens[p.index-1]
}

// AdvanceToken advances the current token to the next non-whitespace token.
func (p *Parser) AdvanceToken() {
	for {
		p.index++
		if p.index-1 >= len(p.tokens) {
			return
		}
		if _, isWhitespace := p.tokens[p.index-1].Token.(tokenizer.TokenWhitespace); !isWhitespace {
			return
		}
	}
}

// GetCurrentToken returns a reference to the current token.
// The current token is the one at index - 1.
func (p *Parser) GetCurrentToken() *tokenizer.TokenWithSpan {
	return p.tokenAt(p.index - 1)
}

// GetPreviousToken returns a reference to the previous token.
// The previous token is the one at index - 2.
func (p *Parser) GetPreviousToken() *tokenizer.TokenWithSpan {
	return p.tokenAt(p.index - 2)
}

// GetNextToken returns a reference to the next token.
// The next token is the one at index.
func (p *Parser) GetNextToken() *tokenizer.TokenWithSpan {
	return p.tokenAt(p.index)
}

// GetCurrentIndex returns the index of the current token.
func (p *Parser) GetCurrentIndex() int {
	if p.index == 0 {
		return 0
	}
	return p.index - 1
}

// PrevToken seeks back the last one non-whitespace token.
// Must be called after NextToken(), otherwise might panic.
// OK to call after NextToken() indicates an EOF.
func (p *Parser) PrevToken() {
	for {
		if p.index <= 0 {
			panic("cannot go before first token")
		}
		p.index--
		if p.index < len(p.tokens) {
			if _, isWhitespace := p.tokens[p.index].Token.(tokenizer.TokenWhitespace); !isWhitespace {
				return
			}
		}
	}
}

// SavePosition saves the current parser position and returns a restore function.
// Usage:
//
//	restore := p.SavePosition()
//	if err := tryParseSomething(); err != nil {
//	    restore() // backtrack
//	    return nil, err
//	}
func (p *Parser) SavePosition() func() {
	savedIndex := p.index
	return func() {
		p.index = savedIndex
	}
}

// TryParse attempts to run a parse function and restores position if it fails.
// This is the Go equivalent of Rust's try_parse/maybe_parse.
func TryParse[T any](p *Parser, parseFn func() (T, error)) (T, bool) {
	var zero T
	restore := p.SavePosition()
	result, err := parseFn()
	if err != nil {
		restore()
		return zero, false
	}
	return result, true
}

// ConsumeToken consumes the next token if it matches the expected token.
// Returns true if the token was consumed, false otherwise.
func (p *Parser) ConsumeToken(expected tokenizer.Token) bool {
	if p.PeekTokenRef().Token.Equals(expected) {
		p.AdvanceToken()
		return true
	}
	return false
}

// ConsumeTokens consumes multiple tokens in sequence.
// Returns true if all tokens were consumed, false otherwise.
// If any token doesn't match, the parser state is restored.
func (p *Parser) ConsumeTokens(tokens []tokenizer.Token) bool {
	index := p.index
	for _, tok := range tokens {
		if !p.ConsumeToken(tok) {
			p.index = index
			return false
		}
	}
	return true
}

// ExpectToken consumes the next token if it matches the expected token,
// otherwise returns an error.
func (p *Parser) ExpectToken(expected tokenizer.Token) (tokenizer.TokenWithSpan, error) {
	if p.PeekTokenRef().Token.Equals(expected) {
		return p.NextToken(), nil
	}
	return tokenizer.TokenWithSpan{}, p.expectedRef(expected.String(), p.PeekTokenRef())
}

// ============================================================
// Keyword Helpers
// ============================================================

// ParseKeyword checks if the current token is the expected keyword,
// consumes it and returns true. Otherwise, no tokens are consumed and returns false.
func (p *Parser) ParseKeyword(expected string) bool {
	if p.PeekKeyword(expected) {
		p.AdvanceToken()
		return true
	}
	return false
}

// PeekKeyword checks if the current token is the expected keyword without consuming it.
func (p *Parser) PeekKeyword(expected string) bool {
	tok := p.PeekTokenRef()
	if wordTok, ok := tok.Token.(tokenizer.TokenWord); ok {
		return string(wordTok.Word.Keyword) == expected
	}
	return false
}

// ParseKeywords checks if the current and subsequent tokens exactly match the
// keywords sequence, consumes them and returns true.
// Otherwise, no tokens are consumed and returns false.
func (p *Parser) ParseKeywords(keywords []string) bool {
	startIndex := p.index
	for _, keyword := range keywords {
		if !p.ParseKeyword(keyword) {
			p.index = startIndex
			return false
		}
	}
	return true
}

// ExpectKeyword consumes the current token if it is the expected keyword.
// Otherwise, returns an error.
func (p *Parser) ExpectKeyword(expected string) (tokenizer.TokenWithSpan, error) {
	if p.ParseKeyword(expected) {
		return *p.GetCurrentToken(), nil
	}
	return tokenizer.TokenWithSpan{}, p.expectedRef(fmt.Sprintf("%q", expected), p.PeekTokenRef())
}

// ExpectKeywordIs consumes the current token if it is the expected keyword.
// Otherwise, returns an error. This is like ExpectKeyword but doesn't return the token.
func (p *Parser) ExpectKeywordIs(expected string) error {
	if p.ParseKeyword(expected) {
		return nil
	}
	return p.expectedRef(fmt.Sprintf("%q", expected), p.PeekTokenRef())
}

// ExpectKeywords consumes the current and subsequent tokens if they match
// the expected keywords sequence. Otherwise, returns an error.
func (p *Parser) ExpectKeywords(expected []string) error {
	for _, kw := range expected {
		if err := p.ExpectKeywordIs(kw); err != nil {
			return err
		}
	}
	return nil
}

// ParseOneOfKeywords checks if the current token is one of the given keywords,
// consumes it and returns the matched keyword.
// Returns empty string if no match.
func (p *Parser) ParseOneOfKeywords(keywords []string) string {
	tok := p.PeekTokenRef()
	if wordTok, ok := tok.Token.(tokenizer.TokenWord); ok {
		for _, keyword := range keywords {
			if string(wordTok.Word.Keyword) == keyword {
				p.AdvanceToken()
				return keyword
			}
		}
	}
	return ""
}

// PeekOneOfKeywords checks if the current token is one of the given keywords
// without consuming it. Returns the matched keyword or empty string.
func (p *Parser) PeekOneOfKeywords(keywords []string) string {
	tok := p.PeekTokenRef()
	if wordTok, ok := tok.Token.(tokenizer.TokenWord); ok {
		for _, keyword := range keywords {
			if string(wordTok.Word.Keyword) == keyword {
				return keyword
			}
		}
	}
	return ""
}

// ExpectOneOfKeywords checks if the current token is one of the expected keywords,
// consumes it and returns the matched keyword. Otherwise, returns an error.
func (p *Parser) ExpectOneOfKeywords(keywords []string) (string, error) {
	if keyword := p.ParseOneOfKeywords(keywords); keyword != "" {
		return keyword, nil
	}
	return "", p.expectedRef(fmt.Sprintf("one of %v", keywords), p.PeekTokenRef())
}

// ============================================================
// Comma-Separated Lists
// ============================================================

// ParseCommaSeparated parses a comma-separated list of items.
// The parseFn function is called to parse each item.
// Trailing commas are controlled by the parser's options.
func (p *Parser) ParseCommaSeparated(parseFn func() error) error {
	return p.ParseCommaSeparatedWithTrailingCommas(parseFn, p.options.TrailingCommas)
}

// ParseCommaSeparatedWithTrailingCommas parses a comma-separated list of items
// with explicit control over trailing commas.
func (p *Parser) ParseCommaSeparatedWithTrailingCommas(parseFn func() error, trailingCommas bool) error {
	for {
		if err := parseFn(); err != nil {
			return err
		}

		if !p.ConsumeToken(tokenizer.TokenComma{}) {
			return nil
		}

		// Check for trailing comma if enabled
		if trailingCommas {
			nextTok := p.PeekTokenRef()
			switch nextTok.Token.(type) {
			case tokenizer.TokenRParen, tokenizer.EOF:
				return nil
			}
			if wordTok, ok := nextTok.Token.(tokenizer.TokenWord); ok {
				if p.isReservedForColumnAlias(string(wordTok.Word.Keyword)) {
					p.PrevToken() // Put back the comma
					return nil
				}
			}
		}
	}
}

// ParseCommaSeparatedToSlice parses a comma-separated list and collects the results.
func (p *Parser) ParseCommaSeparatedToSlice(parseFn func() (interface{}, error)) ([]interface{}, error) {
	var results []interface{}
	err := p.ParseCommaSeparated(func() error {
		item, err := parseFn()
		if err != nil {
			return err
		}
		results = append(results, item)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// ============================================================
// Parenthesized Expressions
// ============================================================

// ParseParenthesized parses a parenthesized expression.
// It expects a '(' before calling parseFn and expects a ')' after.
func (p *Parser) ParseParenthesized(parseFn func() error) error {
	if _, err := p.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return err
	}
	if err := parseFn(); err != nil {
		return err
	}
	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return err
	}
	return nil
}

// ParseParenthesizedOptional parses an optional parenthesized expression.
// If the next token is '(', it parses the content, otherwise returns nil.
func (p *Parser) ParseParenthesizedOptional(parseFn func() error) error {
	if !p.ConsumeToken(tokenizer.TokenLParen{}) {
		return nil
	}
	if err := parseFn(); err != nil {
		return err
	}
	if _, err := p.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return err
	}
	return nil
}

// ============================================================
// Optional Parsing
// ============================================================

// MaybeParse attempts to parse an optional element.
// If parsing succeeds, returns the result and true.
// If parsing fails, restores the parser state and returns nil and false.
func (p *Parser) MaybeParse(parseFn func() (interface{}, error)) (interface{}, bool) {
	index := p.index
	result, err := parseFn()
	if err != nil {
		p.index = index
		return nil, false
	}
	return result, true
}

// ParseOptional attempts to parse an optional element.
// Returns the result if parsed, nil otherwise.
func (p *Parser) ParseOptional(parseFn func() (interface{}, error)) interface{} {
	result, _ := p.MaybeParse(parseFn)
	return result
}

// ============================================================
// Expression Parsing Infrastructure
// ============================================================

// GetNextPrecedence returns the precedence of the next token.
// This is used for Pratt parsing (top-down operator precedence).
func (p *Parser) GetNextPrecedence() (uint8, error) {
	return p.dialect.GetNextPrecedence(p)
}

// ============================================================
// Reserved Keyword Check
// ============================================================

// isReservedForColumnAlias returns true if the keyword is reserved
// and cannot be used as a column alias without AS.
func (p *Parser) isReservedForColumnAlias(keyword string) bool {
	// Common keywords that indicate the end of a select item
	reserved := []string{
		"FROM", "WHERE", "GROUP", "HAVING", "ORDER", "LIMIT", "OFFSET",
		"UNION", "INTERSECT", "EXCEPT", "INTO", "VALUES", "ON", "JOIN",
		"INNER", "LEFT", "RIGHT", "FULL", "CROSS", "WINDOW", "QUALIFY",
	}
	for _, r := range reserved {
		if keyword == r {
			return true
		}
	}
	return false
}

// ============================================================
// Recursion Protection
// ============================================================

// RecursionCounter tracks remaining recursion depth.
// When it reaches 0, an error is returned to prevent stack overflow.
type RecursionCounter struct {
	remainingDepth int
}

// NewRecursionCounter creates a new RecursionCounter with the specified maximum depth.
func NewRecursionCounter(remainingDepth int) RecursionCounter {
	return RecursionCounter{remainingDepth: remainingDepth}
}

// TryDecrease attempts to decrease the remaining depth by 1.
// Returns an error if the remaining depth falls to 0.
func (rc *RecursionCounter) TryDecrease() error {
	if rc.remainingDepth == 0 {
		return errors.NewRecursionLimitError(span.Span{})
	}
	rc.remainingDepth--
	return nil
}

// Increase increases the remaining depth by 1.
// Should be called (via defer) when exiting a recursive function.
func (rc *RecursionCounter) Increase() {
	rc.remainingDepth++
}

// GetRemaining returns the current remaining depth.
func (rc *RecursionCounter) GetRemaining() int {
	return rc.remainingDepth
}
