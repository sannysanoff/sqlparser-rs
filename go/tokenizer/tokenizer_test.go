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
	"testing"

	"github.com/user/sqlparser/span"
)

// Helper functions for tests
func compareTokens(t *testing.T, expected, actual []Token) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("Token count mismatch: expected %d, got %d", len(expected), len(actual))
	}
	for i := range expected {
		if !expected[i].Equals(actual[i]) {
			t.Errorf("Token %d mismatch:\nexpected: %v\nactual: %v", i, expected[i], actual[i])
		}
	}
}

func compareTokensWithSpan(t *testing.T, expected, actual []TokenWithSpan) {
	t.Helper()
	if len(expected) != len(actual) {
		t.Fatalf("Token count mismatch: expected %d, got %d", len(expected), len(actual))
	}
	for i := range expected {
		if !expected[i].Token.Equals(actual[i].Token) || expected[i].Span != actual[i].Span {
			t.Errorf("Token %d mismatch:\nexpected: %v @ %v\nactual: %v @ %v",
				i, expected[i].Token, expected[i].Span, actual[i].Token, actual[i].Span)
		}
	}
}

func TestTokenizerError(t *testing.T) {
	err := &TokenizerError{
		Message:  "test",
		Location: span.Location{Line: 1, Column: 1},
	}
	expected := "test at Line: 1, Column: 1"
	if err.Error() != expected {
		t.Errorf("Expected %q, got %q", expected, err.Error())
	}
}

func TestTokenizeSelect1(t *testing.T) {
	sql := "SELECT 1"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1", Long: false},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeSelectFloat(t *testing.T) {
	sql := "SELECT .1"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: ".1", Long: false},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeWithMapper(t *testing.T) {
	sql := "SELECT ?"
	dialect := &GenericDialect{}
	paramNum := 1

	var tokens []TokenWithSpan
	mapper := func(tws TokenWithSpan) TokenWithSpan {
		if ph, ok := tws.Token.(TokenPlaceholder); ok {
			if ph.Value == "?" {
				tws.Token = TokenPlaceholder{Value: fmt.Sprintf("$%d", paramNum)}
				paramNum++
			}
		}
		return tws
	}

	err := NewTokenizer(dialect, sql).TokenizeIntoBufWithMapper(&tokens, mapper)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tokens) != 3 {
		t.Fatalf("Expected 3 tokens, got %d", len(tokens))
	}
	if !tokens[2].Token.Equals(TokenPlaceholder{Value: "$1"}) {
		t.Errorf("Expected placeholder $1, got %v", tokens[2].Token)
	}
}

func TestTokenizeNumericLiteralUnderscore(t *testing.T) {
	sql := "SELECT 10_000"
	dialect := &GenericDialect{} // GenericDialect does NOT support numeric literal underscores
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Without numeric literal underscores support, 10_000 should be tokenized as 10 followed by _000
	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "10", Long: false},
		MakeWord("_000", nil),
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeSelectExponent(t *testing.T) {
	sql := "SELECT 1e10, 1e-10, 1e+10, 1ea, 1e-10a, 1e-10-10"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1e10", Long: false},
		TokenComma{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1e-10", Long: false},
		TokenComma{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1e+10", Long: false},
		TokenComma{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1", Long: false},
		MakeWord("ea", nil),
		TokenComma{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1e-10", Long: false},
		MakeWord("a", nil),
		TokenComma{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1e-10", Long: false},
		TokenMinus{},
		TokenNumber{Value: "10", Long: false},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeScalarFunction(t *testing.T) {
	sql := "SELECT sqrt(1)"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("sqrt", nil),
		TokenLParen{},
		TokenNumber{Value: "1", Long: false},
		TokenRParen{},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeStringConcat(t *testing.T) {
	sql := "SELECT 'a' || 'b'"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenSingleQuotedString{Value: "a"},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenStringConcat{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenSingleQuotedString{Value: "b"},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeBitwiseOp(t *testing.T) {
	sql := "SELECT one | two ^ three"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("one", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenPipe{},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("two", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenCaret{},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("three", nil),
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeLogicalXor(t *testing.T) {
	sql := "SELECT true XOR true, false XOR false, true XOR false, false XOR true"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("true", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("XOR"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("true", nil),
		TokenComma{},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("false", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("XOR"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("false", nil),
		TokenComma{},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("true", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("XOR"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("false", nil),
		TokenComma{},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("false", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("XOR"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("true", nil),
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeSimpleSelect(t *testing.T) {
	sql := "SELECT * FROM customer WHERE id = 1 LIMIT 5"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenMul{},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("FROM"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("customer", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("WHERE"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("id", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenEq{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1", Long: false},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("LIMIT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "5", Long: false},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeExplainSelect(t *testing.T) {
	sql := "EXPLAIN SELECT * FROM customer WHERE id = 1"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("EXPLAIN"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenMul{},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("FROM"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("customer", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("WHERE"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("id", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenEq{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1", Long: false},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeStringPredicate(t *testing.T) {
	sql := "SELECT * FROM customer WHERE salary <> 'Not Provided'"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenMul{},
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("FROM"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("customer", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("WHERE"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord("salary", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNeq{},
		TokenWhitespace{Whitespace{Type: Space}},
		TokenSingleQuotedString{Value: "Not Provided"},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeInvalidString(t *testing.T) {
	sql := "\n💝مصطفىh"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		TokenWhitespace{Whitespace{Type: Newline}},
		TokenChar{Char: '💝'},
		MakeWord("مصطفىh", nil),
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeNewlineInStringLiteral(t *testing.T) {
	sql := "'foo\r\nbar\nbaz'"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		TokenSingleQuotedString{Value: "foo\r\nbar\nbaz"},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeUnterminatedStringLiteral(t *testing.T) {
	sql := "select 'foo"
	dialect := &GenericDialect{}
	_, err := NewTokenizer(dialect, sql).Tokenize()
	if err == nil {
		t.Fatal("Expected error for unterminated string literal")
	}
}

func TestTokenizeDollarQuotedString(t *testing.T) {
	testCases := []struct {
		sql      string
		expected []Token
	}{
		{
			sql: "SELECT $tag$dollar '$' quoted strings have $tags like this$ or like this $$$tag$",
			expected: []Token{
				MakeKeyword("SELECT"),
				TokenWhitespace{Whitespace{Type: Space}},
				TokenDollarQuotedString{DollarQuotedString{Value: "dollar '$' quoted strings have $tags like this$ or like this $$", Tag: strPtr("tag")}},
			},
		},
		{
			sql: "SELECT $abc$x$ab$abc$",
			expected: []Token{
				MakeKeyword("SELECT"),
				TokenWhitespace{Whitespace{Type: Space}},
				TokenDollarQuotedString{DollarQuotedString{Value: "x$ab", Tag: strPtr("abc")}},
			},
		},
		{
			sql: "SELECT $abc$$abc$",
			expected: []Token{
				MakeKeyword("SELECT"),
				TokenWhitespace{Whitespace{Type: Space}},
				TokenDollarQuotedString{DollarQuotedString{Value: "", Tag: strPtr("abc")}},
			},
		},
		{
			sql: "0$abc$$abc$1",
			expected: []Token{
				TokenNumber{Value: "0", Long: false},
				TokenDollarQuotedString{DollarQuotedString{Value: "", Tag: strPtr("abc")}},
				TokenNumber{Value: "1", Long: false},
			},
		},
		{
			sql: "$function$abc$q$data$q$$function$",
			expected: []Token{
				TokenDollarQuotedString{DollarQuotedString{Value: "abc$q$data$q$", Tag: strPtr("function")}},
			},
		},
	}

	dialect := &GenericDialect{}
	for _, tc := range testCases {
		tokens, err := NewTokenizer(dialect, tc.sql).Tokenize()
		if err != nil {
			t.Fatalf("Unexpected error for %q: %v", tc.sql, err)
		}
		compareTokens(t, tc.expected, tokens)
	}
}

func TestTokenizeDollarQuotedStringUnterminated(t *testing.T) {
	sql := "SELECT $tag$dollar '$' quoted strings have $tags like this$ or like this $$$different tag$"
	dialect := &GenericDialect{}
	_, err := NewTokenizer(dialect, sql).Tokenize()
	if err == nil {
		t.Fatal("Expected error for unterminated dollar-quoted string")
	}
}

func TestTokenizeRightArrow(t *testing.T) {
	sql := "FUNCTION(key=>value)"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeWord("FUNCTION", nil),
		TokenLParen{},
		MakeWord("key", nil),
		TokenRArrow{},
		MakeWord("value", nil),
		TokenRParen{},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeIsNull(t *testing.T) {
	sql := "a IS NULL"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeWord("a", nil),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("IS"),
		TokenWhitespace{Whitespace{Type: Space}},
		MakeKeyword("NULL"),
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeComment(t *testing.T) {
	testCases := []struct {
		sql      string
		expected []Token
	}{
		{
			sql: "0--this is a comment\n1",
			expected: []Token{
				TokenNumber{Value: "0", Long: false},
				TokenWhitespace{Whitespace{Type: SingleLineComment, Prefix: "--", Content: "this is a comment\n"}},
				TokenNumber{Value: "1", Long: false},
			},
		},
		{
			sql: "0--this is a comment\r1",
			expected: []Token{
				TokenNumber{Value: "0", Long: false},
				TokenWhitespace{Whitespace{Type: SingleLineComment, Prefix: "--", Content: "this is a comment\r1"}},
			},
		},
		{
			sql: "0--this is a comment\r\n1",
			expected: []Token{
				TokenNumber{Value: "0", Long: false},
				TokenWhitespace{Whitespace{Type: SingleLineComment, Prefix: "--", Content: "this is a comment\r\n"}},
				TokenNumber{Value: "1", Long: false},
			},
		},
	}

	dialect := &GenericDialect{}
	for _, tc := range testCases {
		tokens, err := NewTokenizer(dialect, tc.sql).Tokenize()
		if err != nil {
			t.Fatalf("Unexpected error for %q: %v", tc.sql, err)
		}
		compareTokens(t, tc.expected, tokens)
	}
}

func TestTokenizeMultilineComment(t *testing.T) {
	sql := "0/*multi-line\n* /comment*/1"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		TokenNumber{Value: "0", Long: false},
		TokenWhitespace{Whitespace{Type: MultiLineComment, Content: "multi-line\n* /comment"}},
		TokenNumber{Value: "1", Long: false},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeNestedMultilineComment(t *testing.T) {
	// GenericDialect does not support nested comments
	sql := "SELECT 1/*/* nested comment */*/0"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Without nested comment support, /*/* is just a regular comment start
	// followed by " nested comment " then */ */
	expected := []Token{
		MakeKeyword("SELECT"),
		TokenWhitespace{Whitespace{Type: Space}},
		TokenNumber{Value: "1", Long: false},
		TokenWhitespace{Whitespace{Type: MultiLineComment, Content: "/* nested comment "}},
		TokenMul{},
		TokenDiv{},
		TokenNumber{Value: "0", Long: false},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeMismatchedQuotes(t *testing.T) {
	sql := `"foo"`
	dialect := &GenericDialect{}
	// Double quotes start delimited identifiers in generic dialect
	// so this should succeed and tokenize as a word
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if len(tokens) != 1 {
		t.Fatalf("Expected 1 token, got %d", len(tokens))
	}

	word, ok := tokens[0].(TokenWord)
	if !ok {
		t.Fatalf("Expected TokenWord, got %T", tokens[0])
	}
	if word.Value != "foo" {
		t.Errorf("Expected word 'foo', got %q", word.Value)
	}
}

func TestTokenizeNewlines(t *testing.T) {
	sql := "line1\nline2\rline3\r\nline4\r"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		MakeWord("line1", nil),
		TokenWhitespace{Whitespace{Type: Newline}},
		MakeWord("line2", nil),
		TokenWhitespace{Whitespace{Type: Newline}},
		MakeWord("line3", nil),
		TokenWhitespace{Whitespace{Type: Newline}},
		MakeWord("line4", nil),
		TokenWhitespace{Whitespace{Type: Newline}},
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeQuotedIdentifier(t *testing.T) {
	sql := ` "a "" b"`
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).Tokenize()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []Token{
		TokenWhitespace{Whitespace{Type: Space}},
		MakeWord(`a " b`, bytePtr('"')),
	}

	compareTokens(t, expected, tokens)
}

func TestTokenizeWithLocation(t *testing.T) {
	sql := "SELECT a,\n b"
	dialect := &GenericDialect{}
	tokens, err := NewTokenizer(dialect, sql).TokenizeWithSpan()
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	expected := []TokenWithSpan{
		{Token: MakeKeyword("SELECT"), Span: span.Span{Start: span.Location{Line: 1, Column: 1}, End: span.Location{Line: 1, Column: 7}}},
		{Token: TokenWhitespace{Whitespace{Type: Space}}, Span: span.Span{Start: span.Location{Line: 1, Column: 7}, End: span.Location{Line: 1, Column: 8}}},
		{Token: MakeWord("a", nil), Span: span.Span{Start: span.Location{Line: 1, Column: 8}, End: span.Location{Line: 1, Column: 9}}},
		{Token: TokenComma{}, Span: span.Span{Start: span.Location{Line: 1, Column: 9}, End: span.Location{Line: 1, Column: 10}}},
		{Token: TokenWhitespace{Whitespace{Type: Newline}}, Span: span.Span{Start: span.Location{Line: 1, Column: 10}, End: span.Location{Line: 2, Column: 1}}},
		{Token: TokenWhitespace{Whitespace{Type: Space}}, Span: span.Span{Start: span.Location{Line: 2, Column: 1}, End: span.Location{Line: 2, Column: 2}}},
		{Token: MakeWord("b", nil), Span: span.Span{Start: span.Location{Line: 2, Column: 2}, End: span.Location{Line: 2, Column: 3}}},
	}

	compareTokensWithSpan(t, expected, tokens)
}

func TestTokenEquality(t *testing.T) {
	// Test various token types for equality
	tests := []struct {
		a        Token
		b        Token
		expected bool
	}{
		{TokenComma{}, TokenComma{}, true},
		{TokenComma{}, TokenPeriod{}, false},
		{TokenNumber{Value: "1"}, TokenNumber{Value: "1"}, true},
		{TokenNumber{Value: "1"}, TokenNumber{Value: "2"}, false},
		{MakeWord("foo", nil), MakeWord("foo", nil), true},
		{MakeWord("foo", nil), MakeWord("bar", nil), false},
	}

	for _, test := range tests {
		result := test.a.Equals(test.b)
		if result != test.expected {
			t.Errorf("Expected %v.Equals(%v) = %v, got %v", test.a, test.b, test.expected, result)
		}
	}
}

func TestTokenTypes(t *testing.T) {
	tests := []struct {
		token    Token
		typeName string
	}{
		{EOF{}, "EOF"},
		{MakeWord("foo", nil), "Word"},
		{TokenNumber{Value: "1"}, "Number"},
		{TokenComma{}, "Operator/Punctuation"},
		{TokenSingleQuotedString{Value: "foo"}, "SingleQuotedString"},
	}

	for _, test := range tests {
		result := TokenType(test.token)
		if result != test.typeName {
			t.Errorf("Expected TokenType(%T) = %q, got %q", test.token, test.typeName, result)
		}
	}
}

// Helper function
func strPtr(s string) *string {
	return &s
}

// Helper function for byte pointer
func bytePtr(b byte) *byte {
	return &b
}

// Ensure GenericDialect implements Dialect
var _ Dialect = (*GenericDialect)(nil)
