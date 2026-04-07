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

// Package expr provides SQL expression types for the sqlparser AST.
package expr

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// Identifier represents a single identifier expression (Expr::Identifier in Rust).
type Identifier struct {
	SpanVal token.Span
	Ident   *Ident
}

func (i *Identifier) exprNode() {}

// Span returns the source span for this expression.
func (i *Identifier) Span() token.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *Identifier) String() string {
	if i.Ident != nil {
		return i.Ident.String()
	}
	return ""
}

// CompoundIdentifier represents a multi-part identifier (Expr::CompoundIdentifier in Rust).
type CompoundIdentifier struct {
	SpanVal token.Span
	Idents  []*Ident
}

func (c *CompoundIdentifier) exprNode() {}

// Span returns the source span for this expression.
func (c *CompoundIdentifier) Span() token.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *CompoundIdentifier) String() string {
	parts := make([]string, len(c.Idents))
	for i, ident := range c.Idents {
		parts[i] = ident.String()
	}
	return strings.Join(parts, ".")
}

// SystemVariable represents a system variable reference like @@var or @@global.var (MySQL-specific).
type SystemVariable struct {
	SpanVal token.Span
	// Name is the system variable name (e.g., "sql_mode" or "global.sql_mode")
	Name *CompoundIdentifier
}

func (s *SystemVariable) exprNode() {}

// Span returns the source span for this expression.
func (s *SystemVariable) Span() token.Span {
	return s.SpanVal
}

// String returns the SQL representation (e.g., "@@sql_mode" or "@@global.sql_mode").
func (s *SystemVariable) String() string {
	if s.Name == nil {
		return "@@"
	}
	return "@@" + s.Name.String()
}

// ValueExpr represents a literal value expression (Expr::Value in Rust).
type ValueExpr struct {
	SpanVal token.Span
	Value   interface{} // *ValueWithSpan or Value type
}

func (v *ValueExpr) exprNode() {}

// Span returns the source span for this expression.
func (v *ValueExpr) Span() token.Span {
	return v.SpanVal
}

// String returns the SQL representation.
func (v *ValueExpr) String() string {
	if v.Value == nil {
		return "NULL"
	}
	// Handle Go bool types (TRUE/FALSE must be uppercase)
	if b, ok := v.Value.(bool); ok {
		if b {
			return "TRUE"
		}
		return "FALSE"
	}
	if s, ok := v.Value.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("%v", v.Value)
}

// QualifiedWildcard represents a qualified wildcard (Expr::QualifiedWildcard in Rust).
type QualifiedWildcard struct {
	SpanVal token.Span
	Prefix  *ObjectName
}

func (q *QualifiedWildcard) exprNode() {}

// Span returns the source span for this expression.
func (q *QualifiedWildcard) Span() token.Span {
	return q.SpanVal
}

// String returns the SQL representation.
func (q *QualifiedWildcard) String() string {
	if q.Prefix == nil {
		return "*"
	}
	return q.Prefix.String() + ".*"
}

// Wildcard represents an unqualified `*` wildcard (Expr::Wildcard in Rust).
type Wildcard struct {
	SpanVal token.Span
}

func (w *Wildcard) exprNode() {}

// Span returns the source span for this expression.
func (w *Wildcard) Span() token.Span {
	return w.SpanVal
}

// String returns the SQL representation.
func (w *Wildcard) String() string {
	return "*"
}

// Nested represents a nested expression in parentheses (Expr::Nested in Rust).
type Nested struct {
	SpanVal token.Span
	Expr    Expr
}

func (n *Nested) exprNode() {}

// Span returns the source span for this expression.
func (n *Nested) Span() token.Span {
	return n.SpanVal
}

// String returns the SQL representation.
func (n *Nested) String() string {
	return "(" + n.Expr.String() + ")"
}

// Prefixed represents a prefixed expression (Expr::Prefixed in Rust).
type Prefixed struct {
	SpanVal token.Span
	Prefix  *Ident
	Value   Expr
}

func (p *Prefixed) exprNode() {}

// Span returns the source span for this expression.
func (p *Prefixed) Span() token.Span {
	return p.SpanVal
}

// String returns the SQL representation.
func (p *Prefixed) String() string {
	return p.Prefix.String() + " " + p.Value.String()
}

// TypedString represents a typed string literal like DATE '2020-01-01' (Expr::TypedString in Rust).
type TypedString struct {
	SpanVal  token.Span
	DataType string
	Value    string
}

func (t *TypedString) exprNode() {}

// Span returns the source span for this expression.
func (t *TypedString) Span() token.Span {
	return t.SpanVal
}

// String returns the SQL representation.
func (t *TypedString) String() string {
	return fmt.Sprintf("%s '%s'", t.DataType, t.Value)
}

// Ident represents a single identifier (e.g., table name or column name).
type Ident struct {
	SpanVal    token.Span
	Value      string
	QuoteStyle *rune // optional quote character: ', ", `, [
}

// Span returns the source span for this identifier.
func (i *Ident) Span() token.Span {
	return i.SpanVal
}

// String returns the SQL representation of the identifier.
func (i *Ident) String() string {
	if i.QuoteStyle != nil {
		q := *i.QuoteStyle
		switch q {
		case '"', '\'', '`':
			return fmt.Sprintf("%c%s%c", q, i.Value, q)
		case '[':
			return fmt.Sprintf("[%s]", i.Value)
		}
	}
	return i.Value
}

// exprNode is a marker method that identifies this type as an expression node.
func (i *Ident) exprNode() {}

type ObjectNamePart struct {
	SpanVal token.Span
	Ident   *Ident
}

// Span returns the source span for this part.
func (o *ObjectNamePart) Span() token.Span {
	return o.SpanVal
}

// String returns the SQL representation.
func (o *ObjectNamePart) String() string {
	if o.Ident != nil {
		return o.Ident.String()
	}
	return ""
}

// ObjectName represents a qualified name (e.g., database.schema.table).
type ObjectName struct {
	SpanVal token.Span
	Parts   []*ObjectNamePart
}

// Span returns the source span for this object name.
func (o *ObjectName) Span() token.Span {
	return o.SpanVal
}

// String returns the SQL representation.
func (o *ObjectName) String() string {
	parts := make([]string, len(o.Parts))
	for i, part := range o.Parts {
		parts[i] = part.String()
	}
	return strings.Join(parts, ".")
}

// SqlOption represents a SQL option (e.g., in OPTIONS clause).
type SqlOption struct {
	SpanVal      token.Span
	Name         *Ident
	Value        Expr
	EngineParams []*Ident // Optional parenthesized params for ENGINE option (e.g., ENGINE=InnoDB(ROW_FORMAT=DYNAMIC))
}

// Span returns the source span for this option.
func (s *SqlOption) Span() token.Span {
	return s.SpanVal
}

// String returns the SQL representation.
func (s *SqlOption) String() string {
	valueStr := s.Value.String()
	// Add quotes for string literal values stored in ValueExpr
	if valExpr, ok := s.Value.(*ValueExpr); ok {
		if _, isString := valExpr.Value.(string); isString {
			valueStr = fmt.Sprintf("'%s'", valueStr)
		}
	}
	result := fmt.Sprintf("%s = %s", s.Name.String(), valueStr)
	if len(s.EngineParams) > 0 {
		var params []string
		for _, p := range s.EngineParams {
			params = append(params, p.String())
		}
		result += fmt.Sprintf("(%s)", strings.Join(params, ", "))
	}
	return result
}

// ColumnOption represents a column option.
type ColumnOption struct {
	SpanVal token.Span
	Name    string
	Value   Expr
}

// Span returns the source span for this option.
func (c *ColumnOption) Span() token.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *ColumnOption) String() string {
	if c.Value != nil {
		return fmt.Sprintf("%s %s", c.Name, c.Value.String())
	}
	return c.Name
}
