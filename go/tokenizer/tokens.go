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

// Package tokenizer provides SQL tokenization functionality.
// It converts SQL text into a sequence of tokens that can be
// used by the parser to build an Abstract Syntax Tree (AST).
package tokenizer

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/span"
	"github.com/user/sqlparser/token"
)

// Whitespace represents whitespace in the input: spaces, newlines, tabs and comments.
type Whitespace struct {
	Type    WhitespaceType
	Content string
	Prefix  string // For single line comments
}

// WhitespaceType represents different types of whitespace
type WhitespaceType int

const (
	Space WhitespaceType = iota
	Newline
	Tab
	SingleLineComment
	MultiLineComment
)

func (w Whitespace) String() string {
	switch w.Type {
	case Space:
		return " "
	case Newline:
		return "\n"
	case Tab:
		return "\t"
	case SingleLineComment:
		return w.Prefix + w.Content
	case MultiLineComment:
		return "/*" + w.Content + "*/"
	default:
		return ""
	}
}

// DollarQuotedString represents a dollar-quoted string literal
type DollarQuotedString struct {
	Value string
	Tag   *string // nil for untagged ($$...$$)
}

func (d DollarQuotedString) String() string {
	if d.Tag != nil {
		return "$" + *d.Tag + "$" + d.Value + "$" + *d.Tag + "$"
	}
	return "$$" + d.Value + "$$"
}

// QuoteDelimitedString represents an Oracle-style quote-delimited string
type QuoteDelimitedString struct {
	StartQuote byte
	Value      string
	EndQuote   byte
}

func (q QuoteDelimitedString) String() string {
	return "Q'" + string(q.StartQuote) + q.Value + string(q.EndQuote) + "'"
}

// Word represents a keyword or SQL identifier
type Word struct {
	Value      string
	QuoteStyle *byte // nil for unquoted, otherwise the quote character used
	Keyword    token.Keyword
}

func (w Word) String() string {
	if w.QuoteStyle != nil {
		quote := *w.QuoteStyle
		endQuote := matchingEndQuote(quote)
		return string(quote) + w.Value + string(endQuote)
	}
	return w.Value
}

// matchingEndQuote returns the matching closing quote character
func matchingEndQuote(ch byte) byte {
	switch ch {
	case '"':
		return '"' // ANSI and most dialects
	case '[':
		return ']' // MS SQL
	case '`':
		return '`' // MySQL
	default:
		return ch
	}
}

// Token represents a single SQL token.
// This is implemented as an interface with concrete types for each variant.
type Token interface {
	String() string
	Equals(other Token) bool
}

// TokenType returns a string identifier for the token type
func TokenType(t Token) string {
	switch t.(type) {
	case EOF:
		return "EOF"
	case TokenWord:
		return "Word"
	case TokenNumber:
		return "Number"
	case TokenChar:
		return "Char"
	case TokenSingleQuotedString:
		return "SingleQuotedString"
	case TokenDoubleQuotedString:
		return "DoubleQuotedString"
	case TokenTripleSingleQuotedString:
		return "TripleSingleQuotedString"
	case TokenTripleDoubleQuotedString:
		return "TripleDoubleQuotedString"
	case TokenDollarQuotedString:
		return "DollarQuotedString"
	case TokenSingleQuotedByteStringLiteral:
		return "SingleQuotedByteStringLiteral"
	case TokenDoubleQuotedByteStringLiteral:
		return "DoubleQuotedByteStringLiteral"
	case TokenTripleSingleQuotedByteStringLiteral:
		return "TripleSingleQuotedByteStringLiteral"
	case TokenTripleDoubleQuotedByteStringLiteral:
		return "TripleDoubleQuotedByteStringLiteral"
	case TokenSingleQuotedRawStringLiteral:
		return "SingleQuotedRawStringLiteral"
	case TokenDoubleQuotedRawStringLiteral:
		return "DoubleQuotedRawStringLiteral"
	case TokenTripleSingleQuotedRawStringLiteral:
		return "TripleSingleQuotedRawStringLiteral"
	case TokenTripleDoubleQuotedRawStringLiteral:
		return "TripleDoubleQuotedRawStringLiteral"
	case TokenNationalStringLiteral:
		return "NationalStringLiteral"
	case TokenQuoteDelimitedStringLiteral:
		return "QuoteDelimitedStringLiteral"
	case TokenNationalQuoteDelimitedStringLiteral:
		return "NationalQuoteDelimitedStringLiteral"
	case TokenEscapedStringLiteral:
		return "EscapedStringLiteral"
	case TokenUnicodeStringLiteral:
		return "UnicodeStringLiteral"
	case TokenHexStringLiteral:
		return "HexStringLiteral"
	case TokenWhitespace:
		return "Whitespace"
	case TokenComma, TokenDoubleEq, TokenEq, TokenNeq, TokenLt, TokenGt, TokenLtEq, TokenGtEq,
		TokenSpaceship, TokenPlus, TokenMinus, TokenMul, TokenDiv, TokenDuckIntDiv, TokenMod,
		TokenStringConcat, TokenLParen, TokenRParen, TokenPeriod, TokenColon, TokenDoubleColon,
		TokenAssignment, TokenSemiColon, TokenBackslash, TokenLBracket, TokenRBracket,
		TokenAmpersand, TokenPipe, TokenCaret, TokenLBrace, TokenRBrace, TokenRArrow,
		TokenSharp, TokenDoubleSharp, TokenTilde, TokenTildeAsterisk, TokenExclamationMarkTilde,
		TokenExclamationMarkTildeAsterisk, TokenDoubleTilde, TokenDoubleTildeAsterisk,
		TokenExclamationMarkDoubleTilde, TokenExclamationMarkDoubleTildeAsterisk,
		TokenShiftLeft, TokenShiftRight, TokenOverlap, TokenPGSquareRoot, TokenPGCubeRoot,
		TokenAtDashAt, TokenQuestionMarkDash, TokenAmpersandLeftAngleBracket,
		TokenAmpersandRightAngleBracket, TokenAmpersandLeftAngleBracketVerticalBar,
		TokenVerticalBarAmpersandRightAngleBracket, TokenTwoWayArrow, TokenLeftAngleBracketCaret,
		TokenRightAngleBracketCaret, TokenQuestionMarkSharp, TokenQuestionMarkDashVerticalBar,
		TokenQuestionMarkDoubleVerticalBar, TokenTildeEqual, TokenShiftLeftVerticalBar,
		TokenVerticalBarShiftRight, TokenVerticalBarRightAngleBracket, TokenArrow, TokenLongArrow,
		TokenHashArrow, TokenHashLongArrow, TokenAtArrow, TokenArrowAt, TokenHashMinus,
		TokenAtQuestion, TokenAtAt, TokenQuestion, TokenQuestionAnd, TokenQuestionPipe,
		TokenPlaceholder, TokenCustomBinaryOperator, TokenExclamationMark, TokenDoubleExclamationMark,
		TokenAtSign, TokenCaretAt:
		return "Operator/Punctuation"
	default:
		return "Unknown"
	}
}

// EOF represents end of file token
type EOF struct{}

func (EOF) String() string          { return "EOF" }
func (EOF) Equals(other Token) bool { _, ok := other.(EOF); return ok }

// TokenWord represents a keyword or identifier
type TokenWord struct{ Word }

func (t TokenWord) Equals(other Token) bool {
	if o, ok := other.(TokenWord); ok {
		// Compare Value and Keyword
		if t.Word.Value != o.Word.Value || t.Word.Keyword != o.Word.Keyword {
			return false
		}
		// Compare QuoteStyle (handling nil pointers)
		if t.Word.QuoteStyle == nil && o.Word.QuoteStyle == nil {
			return true
		}
		if t.Word.QuoteStyle == nil || o.Word.QuoteStyle == nil {
			return false
		}
		return *t.Word.QuoteStyle == *o.Word.QuoteStyle
	}
	return false
}

// TokenNumber represents a numeric literal
type TokenNumber struct {
	Value string
	Long  bool // true if suffixed with 'L'
}

func (t TokenNumber) String() string {
	if t.Long {
		return t.Value + "L"
	}
	return t.Value
}
func (t TokenNumber) Equals(other Token) bool {
	if o, ok := other.(TokenNumber); ok {
		return t.Value == o.Value && t.Long == o.Long
	}
	return false
}

// TokenChar represents an unrecognized character
type TokenChar struct {
	Char rune
}

func (t TokenChar) String() string { return string(t.Char) }
func (t TokenChar) Equals(other Token) bool {
	if o, ok := other.(TokenChar); ok {
		return t.Char == o.Char
	}
	return false
}

// String literal tokens
type TokenSingleQuotedString struct{ Value string }
type TokenDoubleQuotedString struct{ Value string }
type TokenTripleSingleQuotedString struct{ Value string }
type TokenTripleDoubleQuotedString struct{ Value string }

func (t TokenSingleQuotedString) String() string       { return "'" + t.Value + "'" }
func (t TokenDoubleQuotedString) String() string       { return "\"" + t.Value + "\"" }
func (t TokenTripleSingleQuotedString) String() string { return "'''" + t.Value + "'''" }
func (t TokenTripleDoubleQuotedString) String() string { return "\"\"\"" + t.Value + "\"\"\"" }

func (t TokenSingleQuotedString) Equals(other Token) bool {
	if o, ok := other.(TokenSingleQuotedString); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenDoubleQuotedString) Equals(other Token) bool {
	if o, ok := other.(TokenDoubleQuotedString); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenTripleSingleQuotedString) Equals(other Token) bool {
	if o, ok := other.(TokenTripleSingleQuotedString); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenTripleDoubleQuotedString) Equals(other Token) bool {
	if o, ok := other.(TokenTripleDoubleQuotedString); ok {
		return t.Value == o.Value
	}
	return false
}

// Dollar quoted string token
type TokenDollarQuotedString struct{ DollarQuotedString }

func (t TokenDollarQuotedString) String() string { return t.DollarQuotedString.String() }
func (t TokenDollarQuotedString) Equals(other Token) bool {
	if o, ok := other.(TokenDollarQuotedString); ok {
		return t.Value == o.Value && ((t.Tag == nil && o.Tag == nil) || (t.Tag != nil && o.Tag != nil && *t.Tag == *o.Tag))
	}
	return false
}

// Byte string literal tokens
type TokenSingleQuotedByteStringLiteral struct{ Value string }
type TokenDoubleQuotedByteStringLiteral struct{ Value string }
type TokenTripleSingleQuotedByteStringLiteral struct{ Value string }
type TokenTripleDoubleQuotedByteStringLiteral struct{ Value string }

func (t TokenSingleQuotedByteStringLiteral) String() string       { return "B'" + t.Value + "'" }
func (t TokenDoubleQuotedByteStringLiteral) String() string       { return "B\"" + t.Value + "\"" }
func (t TokenTripleSingleQuotedByteStringLiteral) String() string { return "B'''" + t.Value + "'''" }
func (t TokenTripleDoubleQuotedByteStringLiteral) String() string {
	return "B\"\"\"" + t.Value + "\"\"\""
}

func (t TokenSingleQuotedByteStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenSingleQuotedByteStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenDoubleQuotedByteStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenDoubleQuotedByteStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenTripleSingleQuotedByteStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenTripleSingleQuotedByteStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenTripleDoubleQuotedByteStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenTripleDoubleQuotedByteStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}

// Raw string literal tokens
type TokenSingleQuotedRawStringLiteral struct{ Value string }
type TokenDoubleQuotedRawStringLiteral struct{ Value string }
type TokenTripleSingleQuotedRawStringLiteral struct{ Value string }
type TokenTripleDoubleQuotedRawStringLiteral struct{ Value string }

func (t TokenSingleQuotedRawStringLiteral) String() string       { return "R'" + t.Value + "'" }
func (t TokenDoubleQuotedRawStringLiteral) String() string       { return "R\"" + t.Value + "\"" }
func (t TokenTripleSingleQuotedRawStringLiteral) String() string { return "R'''" + t.Value + "'''" }
func (t TokenTripleDoubleQuotedRawStringLiteral) String() string {
	return "R\"\"\"" + t.Value + "\"\"\""
}

func (t TokenSingleQuotedRawStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenSingleQuotedRawStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenDoubleQuotedRawStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenDoubleQuotedRawStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenTripleSingleQuotedRawStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenTripleSingleQuotedRawStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenTripleDoubleQuotedRawStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenTripleDoubleQuotedRawStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}

// National string literal
type TokenNationalStringLiteral struct{ Value string }

func (t TokenNationalStringLiteral) String() string { return "N'" + t.Value + "'" }
func (t TokenNationalStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenNationalStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}

// Quote delimited string tokens
type TokenQuoteDelimitedStringLiteral struct{ QuoteDelimitedString }
type TokenNationalQuoteDelimitedStringLiteral struct{ QuoteDelimitedString }

func (t TokenQuoteDelimitedStringLiteral) String() string {
	return "Q" + t.QuoteDelimitedString.String()
}
func (t TokenNationalQuoteDelimitedStringLiteral) String() string {
	return "NQ" + t.QuoteDelimitedString.String()
}

func (t TokenQuoteDelimitedStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenQuoteDelimitedStringLiteral); ok {
		return t.Value == o.Value && t.StartQuote == o.StartQuote && t.EndQuote == o.EndQuote
	}
	return false
}
func (t TokenNationalQuoteDelimitedStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenNationalQuoteDelimitedStringLiteral); ok {
		return t.Value == o.Value && t.StartQuote == o.StartQuote && t.EndQuote == o.EndQuote
	}
	return false
}

// Escaped and Unicode string literals
type TokenEscapedStringLiteral struct{ Value string }
type TokenUnicodeStringLiteral struct{ Value string }
type TokenHexStringLiteral struct{ Value string }

func (t TokenEscapedStringLiteral) String() string { return "E'" + t.Value + "'" }
func (t TokenUnicodeStringLiteral) String() string { return "U&'" + t.Value + "'" }
func (t TokenHexStringLiteral) String() string     { return "X'" + t.Value + "'" }

func (t TokenEscapedStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenEscapedStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenUnicodeStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenUnicodeStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}
func (t TokenHexStringLiteral) Equals(other Token) bool {
	if o, ok := other.(TokenHexStringLiteral); ok {
		return t.Value == o.Value
	}
	return false
}

// Whitespace token
type TokenWhitespace struct{ Whitespace }

func (t TokenWhitespace) String() string { return t.Whitespace.String() }
func (t TokenWhitespace) Equals(other Token) bool {
	if o, ok := other.(TokenWhitespace); ok {
		return t.Whitespace.Type == o.Whitespace.Type && t.Whitespace.Content == o.Whitespace.Content
	}
	return false
}

// Single-character tokens and simple operators
type TokenComma struct{}

func (TokenComma) String() string          { return "," }
func (TokenComma) Equals(other Token) bool { _, ok := other.(TokenComma); return ok }

type TokenDoubleEq struct{}

func (TokenDoubleEq) String() string          { return "==" }
func (TokenDoubleEq) Equals(other Token) bool { _, ok := other.(TokenDoubleEq); return ok }

type TokenEq struct{}

func (TokenEq) String() string          { return "=" }
func (TokenEq) Equals(other Token) bool { _, ok := other.(TokenEq); return ok }

type TokenNeq struct{}

func (TokenNeq) String() string          { return "<>" }
func (TokenNeq) Equals(other Token) bool { _, ok := other.(TokenNeq); return ok }

type TokenLt struct{}

func (TokenLt) String() string          { return "<" }
func (TokenLt) Equals(other Token) bool { _, ok := other.(TokenLt); return ok }

type TokenGt struct{}

func (TokenGt) String() string          { return ">" }
func (TokenGt) Equals(other Token) bool { _, ok := other.(TokenGt); return ok }

type TokenLtEq struct{}

func (TokenLtEq) String() string          { return "<=" }
func (TokenLtEq) Equals(other Token) bool { _, ok := other.(TokenLtEq); return ok }

type TokenGtEq struct{}

func (TokenGtEq) String() string          { return ">=" }
func (TokenGtEq) Equals(other Token) bool { _, ok := other.(TokenGtEq); return ok }

type TokenSpaceship struct{}

func (TokenSpaceship) String() string          { return "<=>" }
func (TokenSpaceship) Equals(other Token) bool { _, ok := other.(TokenSpaceship); return ok }

type TokenPlus struct{}

func (TokenPlus) String() string          { return "+" }
func (TokenPlus) Equals(other Token) bool { _, ok := other.(TokenPlus); return ok }

type TokenMinus struct{}

func (TokenMinus) String() string          { return "-" }
func (TokenMinus) Equals(other Token) bool { _, ok := other.(TokenMinus); return ok }

type TokenMul struct{}

func (TokenMul) String() string          { return "*" }
func (TokenMul) Equals(other Token) bool { _, ok := other.(TokenMul); return ok }

type TokenDiv struct{}

func (TokenDiv) String() string          { return "/" }
func (TokenDiv) Equals(other Token) bool { _, ok := other.(TokenDiv); return ok }

type TokenDuckIntDiv struct{}

func (TokenDuckIntDiv) String() string          { return "//" }
func (TokenDuckIntDiv) Equals(other Token) bool { _, ok := other.(TokenDuckIntDiv); return ok }

type TokenMod struct{}

func (TokenMod) String() string          { return "%" }
func (TokenMod) Equals(other Token) bool { _, ok := other.(TokenMod); return ok }

type TokenStringConcat struct{}

func (TokenStringConcat) String() string          { return "||" }
func (TokenStringConcat) Equals(other Token) bool { _, ok := other.(TokenStringConcat); return ok }

type TokenLParen struct{}

func (TokenLParen) String() string          { return "(" }
func (TokenLParen) Equals(other Token) bool { _, ok := other.(TokenLParen); return ok }

type TokenRParen struct{}

func (TokenRParen) String() string          { return ")" }
func (TokenRParen) Equals(other Token) bool { _, ok := other.(TokenRParen); return ok }

type TokenPeriod struct{}

func (TokenPeriod) String() string          { return "." }
func (TokenPeriod) Equals(other Token) bool { _, ok := other.(TokenPeriod); return ok }

type TokenColon struct{}

func (TokenColon) String() string          { return ":" }
func (TokenColon) Equals(other Token) bool { _, ok := other.(TokenColon); return ok }

type TokenDoubleColon struct{}

func (TokenDoubleColon) String() string          { return "::" }
func (TokenDoubleColon) Equals(other Token) bool { _, ok := other.(TokenDoubleColon); return ok }

type TokenAssignment struct{}

func (TokenAssignment) String() string          { return ":=" }
func (TokenAssignment) Equals(other Token) bool { _, ok := other.(TokenAssignment); return ok }

type TokenSemiColon struct{}

func (TokenSemiColon) String() string          { return ";" }
func (TokenSemiColon) Equals(other Token) bool { _, ok := other.(TokenSemiColon); return ok }

type TokenBackslash struct{}

func (TokenBackslash) String() string          { return "\\" }
func (TokenBackslash) Equals(other Token) bool { _, ok := other.(TokenBackslash); return ok }

type TokenLBracket struct{}

func (TokenLBracket) String() string          { return "[" }
func (TokenLBracket) Equals(other Token) bool { _, ok := other.(TokenLBracket); return ok }

type TokenRBracket struct{}

func (TokenRBracket) String() string          { return "]" }
func (TokenRBracket) Equals(other Token) bool { _, ok := other.(TokenRBracket); return ok }

type TokenAmpersand struct{}

func (TokenAmpersand) String() string          { return "&" }
func (TokenAmpersand) Equals(other Token) bool { _, ok := other.(TokenAmpersand); return ok }

type TokenPipe struct{}

func (TokenPipe) String() string          { return "|" }
func (TokenPipe) Equals(other Token) bool { _, ok := other.(TokenPipe); return ok }

type TokenCaret struct{}

func (TokenCaret) String() string          { return "^" }
func (TokenCaret) Equals(other Token) bool { _, ok := other.(TokenCaret); return ok }

type TokenLBrace struct{}

func (TokenLBrace) String() string          { return "{" }
func (TokenLBrace) Equals(other Token) bool { _, ok := other.(TokenLBrace); return ok }

type TokenRBrace struct{}

func (TokenRBrace) String() string          { return "}" }
func (TokenRBrace) Equals(other Token) bool { _, ok := other.(TokenRBrace); return ok }

type TokenRArrow struct{}

func (TokenRArrow) String() string          { return "=>" }
func (TokenRArrow) Equals(other Token) bool { _, ok := other.(TokenRArrow); return ok }

type TokenSharp struct{}

func (TokenSharp) String() string          { return "#" }
func (TokenSharp) Equals(other Token) bool { _, ok := other.(TokenSharp); return ok }

type TokenDoubleSharp struct{}

func (TokenDoubleSharp) String() string          { return "##" }
func (TokenDoubleSharp) Equals(other Token) bool { _, ok := other.(TokenDoubleSharp); return ok }

type TokenTilde struct{}

func (TokenTilde) String() string          { return "~" }
func (TokenTilde) Equals(other Token) bool { _, ok := other.(TokenTilde); return ok }

type TokenTildeAsterisk struct{}

func (TokenTildeAsterisk) String() string          { return "~*" }
func (TokenTildeAsterisk) Equals(other Token) bool { _, ok := other.(TokenTildeAsterisk); return ok }

type TokenExclamationMarkTilde struct{}

func (TokenExclamationMarkTilde) String() string { return "!~" }
func (TokenExclamationMarkTilde) Equals(other Token) bool {
	_, ok := other.(TokenExclamationMarkTilde)
	return ok
}

type TokenExclamationMarkTildeAsterisk struct{}

func (TokenExclamationMarkTildeAsterisk) String() string { return "!~*" }
func (TokenExclamationMarkTildeAsterisk) Equals(other Token) bool {
	_, ok := other.(TokenExclamationMarkTildeAsterisk)
	return ok
}

type TokenDoubleTilde struct{}

func (TokenDoubleTilde) String() string          { return "~~" }
func (TokenDoubleTilde) Equals(other Token) bool { _, ok := other.(TokenDoubleTilde); return ok }

type TokenDoubleTildeAsterisk struct{}

func (TokenDoubleTildeAsterisk) String() string { return "~~*" }
func (TokenDoubleTildeAsterisk) Equals(other Token) bool {
	_, ok := other.(TokenDoubleTildeAsterisk)
	return ok
}

type TokenExclamationMarkDoubleTilde struct{}

func (TokenExclamationMarkDoubleTilde) String() string { return "!~~" }
func (TokenExclamationMarkDoubleTilde) Equals(other Token) bool {
	_, ok := other.(TokenExclamationMarkDoubleTilde)
	return ok
}

type TokenExclamationMarkDoubleTildeAsterisk struct{}

func (TokenExclamationMarkDoubleTildeAsterisk) String() string { return "!~~*" }
func (TokenExclamationMarkDoubleTildeAsterisk) Equals(other Token) bool {
	_, ok := other.(TokenExclamationMarkDoubleTildeAsterisk)
	return ok
}

type TokenShiftLeft struct{}

func (TokenShiftLeft) String() string          { return "<<" }
func (TokenShiftLeft) Equals(other Token) bool { _, ok := other.(TokenShiftLeft); return ok }

type TokenShiftRight struct{}

func (TokenShiftRight) String() string          { return ">>" }
func (TokenShiftRight) Equals(other Token) bool { _, ok := other.(TokenShiftRight); return ok }

type TokenOverlap struct{}

func (TokenOverlap) String() string          { return "&&" }
func (TokenOverlap) Equals(other Token) bool { _, ok := other.(TokenOverlap); return ok }

type TokenPGSquareRoot struct{}

func (TokenPGSquareRoot) String() string          { return "|/" }
func (TokenPGSquareRoot) Equals(other Token) bool { _, ok := other.(TokenPGSquareRoot); return ok }

type TokenPGCubeRoot struct{}

func (TokenPGCubeRoot) String() string          { return "||/" }
func (TokenPGCubeRoot) Equals(other Token) bool { _, ok := other.(TokenPGCubeRoot); return ok }

type TokenAtDashAt struct{}

func (TokenAtDashAt) String() string          { return "@-@" }
func (TokenAtDashAt) Equals(other Token) bool { _, ok := other.(TokenAtDashAt); return ok }

type TokenQuestionMarkDash struct{}

func (TokenQuestionMarkDash) String() string { return "?-" }
func (TokenQuestionMarkDash) Equals(other Token) bool {
	_, ok := other.(TokenQuestionMarkDash)
	return ok
}

type TokenAmpersandLeftAngleBracket struct{}

func (TokenAmpersandLeftAngleBracket) String() string { return "&<" }
func (TokenAmpersandLeftAngleBracket) Equals(other Token) bool {
	_, ok := other.(TokenAmpersandLeftAngleBracket)
	return ok
}

type TokenAmpersandRightAngleBracket struct{}

func (TokenAmpersandRightAngleBracket) String() string { return "&>" }
func (TokenAmpersandRightAngleBracket) Equals(other Token) bool {
	_, ok := other.(TokenAmpersandRightAngleBracket)
	return ok
}

type TokenAmpersandLeftAngleBracketVerticalBar struct{}

func (TokenAmpersandLeftAngleBracketVerticalBar) String() string { return "&<|" }
func (TokenAmpersandLeftAngleBracketVerticalBar) Equals(other Token) bool {
	_, ok := other.(TokenAmpersandLeftAngleBracketVerticalBar)
	return ok
}

type TokenVerticalBarAmpersandRightAngleBracket struct{}

func (TokenVerticalBarAmpersandRightAngleBracket) String() string { return "|&>" }
func (TokenVerticalBarAmpersandRightAngleBracket) Equals(other Token) bool {
	_, ok := other.(TokenVerticalBarAmpersandRightAngleBracket)
	return ok
}

type TokenTwoWayArrow struct{}

func (TokenTwoWayArrow) String() string          { return "<->" }
func (TokenTwoWayArrow) Equals(other Token) bool { _, ok := other.(TokenTwoWayArrow); return ok }

type TokenLeftAngleBracketCaret struct{}

func (TokenLeftAngleBracketCaret) String() string { return "<^" }
func (TokenLeftAngleBracketCaret) Equals(other Token) bool {
	_, ok := other.(TokenLeftAngleBracketCaret)
	return ok
}

type TokenRightAngleBracketCaret struct{}

func (TokenRightAngleBracketCaret) String() string { return ">^" }
func (TokenRightAngleBracketCaret) Equals(other Token) bool {
	_, ok := other.(TokenRightAngleBracketCaret)
	return ok
}

type TokenQuestionMarkSharp struct{}

func (TokenQuestionMarkSharp) String() string { return "?#" }
func (TokenQuestionMarkSharp) Equals(other Token) bool {
	_, ok := other.(TokenQuestionMarkSharp)
	return ok
}

type TokenQuestionMarkDashVerticalBar struct{}

func (TokenQuestionMarkDashVerticalBar) String() string { return "?-|" }
func (TokenQuestionMarkDashVerticalBar) Equals(other Token) bool {
	_, ok := other.(TokenQuestionMarkDashVerticalBar)
	return ok
}

type TokenQuestionMarkDoubleVerticalBar struct{}

func (TokenQuestionMarkDoubleVerticalBar) String() string { return "?||" }
func (TokenQuestionMarkDoubleVerticalBar) Equals(other Token) bool {
	_, ok := other.(TokenQuestionMarkDoubleVerticalBar)
	return ok
}

type TokenTildeEqual struct{}

func (TokenTildeEqual) String() string          { return "~=" }
func (TokenTildeEqual) Equals(other Token) bool { _, ok := other.(TokenTildeEqual); return ok }

type TokenShiftLeftVerticalBar struct{}

func (TokenShiftLeftVerticalBar) String() string { return "<<|" }
func (TokenShiftLeftVerticalBar) Equals(other Token) bool {
	_, ok := other.(TokenShiftLeftVerticalBar)
	return ok
}

type TokenVerticalBarShiftRight struct{}

func (TokenVerticalBarShiftRight) String() string { return "|>>" }
func (TokenVerticalBarShiftRight) Equals(other Token) bool {
	_, ok := other.(TokenVerticalBarShiftRight)
	return ok
}

type TokenVerticalBarRightAngleBracket struct{}

func (TokenVerticalBarRightAngleBracket) String() string { return "|>" }
func (TokenVerticalBarRightAngleBracket) Equals(other Token) bool {
	_, ok := other.(TokenVerticalBarRightAngleBracket)
	return ok
}

// JSON operators
type TokenArrow struct{}                   // ->
func (TokenArrow) String() string          { return "->" }
func (TokenArrow) Equals(other Token) bool { _, ok := other.(TokenArrow); return ok }

type TokenLongArrow struct{}                   // ->>
func (TokenLongArrow) String() string          { return "->>" }
func (TokenLongArrow) Equals(other Token) bool { _, ok := other.(TokenLongArrow); return ok }

type TokenHashArrow struct{}                   // #>
func (TokenHashArrow) String() string          { return "#>" }
func (TokenHashArrow) Equals(other Token) bool { _, ok := other.(TokenHashArrow); return ok }

type TokenHashLongArrow struct{}                   // #>>
func (TokenHashLongArrow) String() string          { return "#>>" }
func (TokenHashLongArrow) Equals(other Token) bool { _, ok := other.(TokenHashLongArrow); return ok }

type TokenAtArrow struct{}                   // @>
func (TokenAtArrow) String() string          { return "@>" }
func (TokenAtArrow) Equals(other Token) bool { _, ok := other.(TokenAtArrow); return ok }

type TokenArrowAt struct{}                   // <@
func (TokenArrowAt) String() string          { return "<@" }
func (TokenArrowAt) Equals(other Token) bool { _, ok := other.(TokenArrowAt); return ok }

type TokenHashMinus struct{}                   // #-
func (TokenHashMinus) String() string          { return "#-" }
func (TokenHashMinus) Equals(other Token) bool { _, ok := other.(TokenHashMinus); return ok }

type TokenAtQuestion struct{}                   // @?
func (TokenAtQuestion) String() string          { return "@?" }
func (TokenAtQuestion) Equals(other Token) bool { _, ok := other.(TokenAtQuestion); return ok }

type TokenAtAt struct{}                   // @@
func (TokenAtAt) String() string          { return "@@" }
func (TokenAtAt) Equals(other Token) bool { _, ok := other.(TokenAtAt); return ok }

type TokenQuestion struct{}                   // ?
func (TokenQuestion) String() string          { return "?" }
func (TokenQuestion) Equals(other Token) bool { _, ok := other.(TokenQuestion); return ok }

type TokenQuestionAnd struct{}                   // ?&
func (TokenQuestionAnd) String() string          { return "?&" }
func (TokenQuestionAnd) Equals(other Token) bool { _, ok := other.(TokenQuestionAnd); return ok }

type TokenQuestionPipe struct{}                   // ?|
func (TokenQuestionPipe) String() string          { return "?|" }
func (TokenQuestionPipe) Equals(other Token) bool { _, ok := other.(TokenQuestionPipe); return ok }

type TokenExclamationMark struct{}          // !
func (TokenExclamationMark) String() string { return "!" }
func (TokenExclamationMark) Equals(other Token) bool {
	_, ok := other.(TokenExclamationMark)
	return ok
}

type TokenDoubleExclamationMark struct{}          // !!
func (TokenDoubleExclamationMark) String() string { return "!!" }
func (TokenDoubleExclamationMark) Equals(other Token) bool {
	_, ok := other.(TokenDoubleExclamationMark)
	return ok
}

type TokenAtSign struct{}                   // @
func (TokenAtSign) String() string          { return "@" }
func (TokenAtSign) Equals(other Token) bool { _, ok := other.(TokenAtSign); return ok }

type TokenCaretAt struct{}                   // ^@
func (TokenCaretAt) String() string          { return "^@" }
func (TokenCaretAt) Equals(other Token) bool { _, ok := other.(TokenCaretAt); return ok }

// Placeholder token
type TokenPlaceholder struct{ Value string }

func (t TokenPlaceholder) String() string { return t.Value }
func (t TokenPlaceholder) Equals(other Token) bool {
	if o, ok := other.(TokenPlaceholder); ok {
		return t.Value == o.Value
	}
	return false
}

// Custom binary operator token
type TokenCustomBinaryOperator struct{ Value string }

func (t TokenCustomBinaryOperator) String() string { return t.Value }
func (t TokenCustomBinaryOperator) Equals(other Token) bool {
	if o, ok := other.(TokenCustomBinaryOperator); ok {
		return t.Value == o.Value
	}
	return false
}

// TokenWithSpan represents a token with its source location
type TokenWithSpan struct {
	Token Token
	Span  span.Span
}

// MakeKeyword creates a Word token from a keyword string
func MakeKeyword(keyword string) Token {
	return MakeWord(keyword, nil)
}

// MakeWord creates a Word token with optional quote style
func MakeWord(word string, quoteStyle *byte) Token {
	w := Word{
		Value:      word,
		QuoteStyle: quoteStyle,
		Keyword:    KeywordLookup(word, quoteStyle),
	}
	return TokenWord{w}
}

// KeywordLookup performs case-insensitive keyword lookup
func KeywordLookup(word string, quoteStyle *byte) token.Keyword {
	if quoteStyle != nil {
		return token.NoKeyword
	}

	// Linear search through AllKeywords
	// Note: In production, you might want to create a map for O(1) lookup
	upperWord := strings.ToUpper(word)
	for _, kw := range token.AllKeywords {
		if string(kw) == upperWord {
			return kw
		}
	}
	return token.NoKeyword
}

// IsEOF returns true if the token is an EOF token
func IsEOF(t Token) bool {
	_, ok := t.(EOF)
	return ok
}

// IsWord returns true if the token is a Word token
func IsWord(t Token) bool {
	_, ok := t.(TokenWord)
	return ok
}

// GetWord returns the Word value if the token is a Word token, nil otherwise
func GetWord(t Token) *Word {
	if tw, ok := t.(TokenWord); ok {
		return &tw.Word
	}
	return nil
}

// IsWhitespace returns true if the token is a Whitespace token
func IsWhitespace(t Token) bool {
	_, ok := t.(TokenWhitespace)
	return ok
}

// GetWhitespace returns the Whitespace value if the token is a Whitespace token, nil otherwise
func GetWhitespace(t Token) *Whitespace {
	if tw, ok := t.(TokenWhitespace); ok {
		return &tw.Whitespace
	}
	return nil
}

// IsPlaceholder returns true if the token is a Placeholder token
func IsPlaceholder(t Token) bool {
	_, ok := t.(TokenPlaceholder)
	return ok
}

// TokenEquals compares two tokens for equality
func TokenEquals(a, b Token) bool {
	return a.Equals(b)
}

// TokensEquals compares two token slices for equality
func TokensEquals(a, b []Token) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Equals(b[i]) {
			return false
		}
	}
	return true
}

// TokensWithSpanEquals compares two TokenWithSpan slices for equality
func TokensWithSpanEquals(a, b []TokenWithSpan) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].Token.Equals(b[i].Token) || a[i].Span != b[i].Span {
			return false
		}
	}
	return true
}

// String returns a string representation of all tokens
func TokensString(tokens []Token) string {
	var parts []string
	for _, t := range tokens {
		parts = append(parts, t.String())
	}
	return strings.Join(parts, " ")
}

// Format returns a formatted string representation of a token with span
func (tws TokenWithSpan) Format() string {
	if tws.Span.IsValid() {
		return fmt.Sprintf("%s@%d:%d-%d:%d", tws.Token.String(),
			tws.Span.Start.Line, tws.Span.Start.Column,
			tws.Span.End.Line, tws.Span.End.Column)
	}
	return tws.Token.String()
}

// NewTokenWithSpan creates a TokenWithSpan from a token and span
func NewTokenWithSpan(tok Token, sp span.Span) TokenWithSpan {
	return TokenWithSpan{Token: tok, Span: sp}
}

// WrapToken creates a TokenWithSpan with an empty span
func WrapToken(tok Token) TokenWithSpan {
	return TokenWithSpan{Token: tok, Span: span.Span{}}
}

// NewEOFToken creates an EOF token with empty span
func NewEOFToken() TokenWithSpan {
	return WrapToken(EOF{})
}

// TokenizerError represents an error during tokenization
type TokenizerError struct {
	Message  string
	Location span.Location
}

// Error implements the error interface
func (e *TokenizerError) Error() string {
	return fmt.Sprintf("%s at Line: %d, Column: %d", e.Message, e.Location.Line, e.Location.Column)
}
