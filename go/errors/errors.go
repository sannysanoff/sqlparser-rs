package errors

import (
	"fmt"
	"github.com/user/sqlparser/span"
)

// ParserError represents errors that can occur during parsing
type ParserError struct {
	Type    ErrorType
	Message string
	Span    span.Span
}

// ErrorType categorizes parser errors
type ErrorType int

const (
	TokenizerErrorType ErrorType = iota
	SyntaxErrorType
	RecursionLimitExceededType
)

func (e ErrorType) String() string {
	switch e {
	case TokenizerErrorType:
		return "TokenizerError"
	case SyntaxErrorType:
		return "ParserError"
	case RecursionLimitExceededType:
		return "RecursionLimitExceeded"
	default:
		return "UnknownError"
	}
}

// Error implements the error interface
func (pe *ParserError) Error() string {
	if pe.Span.IsValid() {
		return fmt.Sprintf("sql %s at Line: %d, Column: %d: %s",
			pe.Type.String(), pe.Span.Start.Line, pe.Span.Start.Column, pe.Message)
	}
	return fmt.Sprintf("sql %s: %s", pe.Type.String(), pe.Message)
}

// NewTokenizerError creates a new tokenizer error
func NewTokenizerError(msg string, loc span.Location) *ParserError {
	return &ParserError{
		Type:    TokenizerErrorType,
		Message: msg,
		Span:    span.NewSpan(loc, loc),
	}
}

// NewParserError creates a new parser error
func NewParserError(msg string, sp span.Span) *ParserError {
	return &ParserError{
		Type:    SyntaxErrorType,
		Message: msg,
		Span:    sp,
	}
}

// NewRecursionLimitError creates a recursion limit exceeded error
func NewRecursionLimitError(sp span.Span) *ParserError {
	return &ParserError{
		Type:    RecursionLimitExceededType,
		Message: "recursion limit exceeded",
		Span:    sp,
	}
}

// IsTokenizerError checks if an error is a tokenizer error
func IsTokenizerError(err error) bool {
	if pe, ok := err.(*ParserError); ok {
		return pe.Type == TokenizerErrorType
	}
	return false
}

// IsParserError checks if an error is a parser error
func IsParserError(err error) bool {
	if pe, ok := err.(*ParserError); ok {
		return pe.Type == SyntaxErrorType
	}
	return false
}

// IsRecursionLimitExceeded checks if an error is a recursion limit error
func IsRecursionLimitExceeded(err error) bool {
	if pe, ok := err.(*ParserError); ok {
		return pe.Type == RecursionLimitExceededType
	}
	return false
}
