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

package ast

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// Ident represents an SQL identifier (e.g., table name, column name).
// It stores the unquoted value and the quote style used.
type Ident struct {
	BaseNode
	// Value is the identifier value without quotes.
	Value string
	// QuoteStyle is the starting quote character if any.
	// Valid quote characters: single quote ('), double quote ("), backtick (`), opening square bracket ([).
	QuoteStyle *rune
}

// NewIdent creates a new unquoted identifier with the given value and an empty span.
func NewIdent(value string) *Ident {
	return &Ident{
		BaseNode: BaseNode{span: token.Span{}},
		Value:    value,
	}
}

// NewIdentWithQuote creates a new quoted identifier.
// Panics if the quote character is not valid.
func NewIdentWithQuote(quote rune, value string) *Ident {
	if quote != '\'' && quote != '"' && quote != '`' && quote != '[' {
		panic(fmt.Sprintf("invalid quote character: %c", quote))
	}
	return &Ident{
		BaseNode:   BaseNode{span: token.Span{}},
		Value:      value,
		QuoteStyle: &quote,
	}
}

// NewIdentWithSpan creates an unquoted identifier with the given span.
func NewIdentWithSpan(s token.Span, value string) *Ident {
	return &Ident{
		BaseNode: BaseNode{span: s},
		Value:    value,
	}
}

// NewIdentWithQuoteAndSpan creates a quoted identifier with the given span.
// Panics if the quote character is not valid.
func NewIdentWithQuoteAndSpan(quote rune, s token.Span, value string) *Ident {
	if quote != '\'' && quote != '"' && quote != '`' && quote != '[' {
		panic(fmt.Sprintf("invalid quote character: %c", quote))
	}
	return &Ident{
		BaseNode:   BaseNode{span: s},
		Value:      value,
		QuoteStyle: &quote,
	}
}

// String returns the SQL representation of the identifier.
func (i *Ident) String() string {
	if i.QuoteStyle == nil {
		return i.Value
	}

	quote := *i.QuoteStyle
	switch quote {
	case '"', '\'', '`':
		escaped := escapeQuotedString(i.Value, quote)
		return fmt.Sprintf("%c%s%c", quote, escaped, quote)
	case '[':
		return fmt.Sprintf("[%s]", i.Value)
	default:
		panic(fmt.Sprintf("unexpected quote style: %c", quote))
	}
}

// escapeQuotedString escapes special characters in a quoted string.
func escapeQuotedString(s string, quote rune) string {
	// For single quotes, double them
	if quote == '\'' {
		return strings.ReplaceAll(s, "'", "''")
	}
	// For double quotes and backticks, double them
	if quote == '"' {
		return strings.ReplaceAll(s, "\"", "\"\"")
	}
	if quote == '`' {
		return strings.ReplaceAll(s, "`", "``")
	}
	return s
}

// ObjectNamePart represents a single part of an object name (e.g., schema.table).
// It can be either a plain identifier or a function that returns an identifier.
type ObjectNamePart interface {
	fmt.Stringer
	// AsIdent returns the identifier if this is an Identifier variant, nil otherwise.
	AsIdent() *Ident
	// IsFunction returns true if this is a Function variant.
	IsFunction() bool
}

// ObjectNamePartIdentifier is an identifier part of an object name.
type ObjectNamePartIdentifier struct {
	Ident *Ident
}

// AsIdent returns the identifier.
func (p *ObjectNamePartIdentifier) AsIdent() *Ident {
	return p.Ident
}

// IsFunction returns false.
func (p *ObjectNamePartIdentifier) IsFunction() bool {
	return false
}

// String returns the SQL representation.
func (p *ObjectNamePartIdentifier) String() string {
	if p.Ident == nil {
		return ""
	}
	return p.Ident.String()
}

// ObjectNamePartFunction is a function that returns an identifier (dialect-specific).
type ObjectNamePartFunction struct {
	// Function represents the function call that produces an identifier.
	// This is a placeholder; the actual function representation will be defined
	// when Function expressions are implemented.
	Name *Ident
	Args []Expr
}

// AsIdent returns nil since this is a function, not a plain identifier.
func (p *ObjectNamePartFunction) AsIdent() *Ident {
	return nil
}

// IsFunction returns true.
func (p *ObjectNamePartFunction) IsFunction() bool {
	return true
}

// String returns the SQL representation.
func (p *ObjectNamePartFunction) String() string {
	if p.Name == nil {
		return ""
	}
	if len(p.Args) == 0 {
		return fmt.Sprintf("%s()", p.Name.String())
	}
	args := make([]string, len(p.Args))
	for i, arg := range p.Args {
		args[i] = arg.String()
	}
	return fmt.Sprintf("%s(%s)", p.Name.String(), strings.Join(args, ", "))
}

// ObjectName represents a multi-part SQL object name (e.g., db.schema.table).
type ObjectName struct {
	BaseNode
	Parts []ObjectNamePart
}

// NewObjectName creates a new ObjectName from a list of identifier strings.
// Each string becomes an unquoted identifier part.
func NewObjectName(parts ...string) *ObjectName {
	objParts := make([]ObjectNamePart, len(parts))
	for i, part := range parts {
		objParts[i] = &ObjectNamePartIdentifier{Ident: NewIdent(part)}
	}
	return &ObjectName{
		BaseNode: BaseNode{span: token.Span{}},
		Parts:    objParts,
	}
}

// NewObjectNameFromIdents creates a new ObjectName from a list of Idents.
func NewObjectNameFromIdents(idents ...*Ident) *ObjectName {
	parts := make([]ObjectNamePart, len(idents))
	for i, ident := range idents {
		parts[i] = &ObjectNamePartIdentifier{Ident: ident}
	}
	return &ObjectName{
		BaseNode: BaseNode{span: token.Span{}},
		Parts:    parts,
	}
}

// NewObjectNameWithSpan creates a new ObjectName with the given span.
func NewObjectNameWithSpan(s token.Span, parts ...ObjectNamePart) *ObjectName {
	return &ObjectName{
		BaseNode: BaseNode{span: s},
		Parts:    parts,
	}
}

// String returns the SQL representation of the object name (parts joined by ".").
func (o *ObjectName) String() string {
	parts := make([]string, len(o.Parts))
	for i, part := range o.Parts {
		parts[i] = part.String()
	}
	return strings.Join(parts, ".")
}

// Last returns the last part of the object name (useful for getting just the table/column name).
func (o *ObjectName) Last() ObjectNamePart {
	if len(o.Parts) == 0 {
		return nil
	}
	return o.Parts[len(o.Parts)-1]
}

// First returns the first part of the object name (useful for getting the catalog/database).
func (o *ObjectName) First() ObjectNamePart {
	if len(o.Parts) == 0 {
		return nil
	}
	return o.Parts[0]
}

// IsSimple returns true if the object name has only one part.
func (o *ObjectName) IsSimple() bool {
	return len(o.Parts) == 1
}
