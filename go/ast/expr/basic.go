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
func (i *Identifier) expr()     {}
func (i *Identifier) IsExpr()   {}

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
func (c *CompoundIdentifier) expr()     {}
func (c *CompoundIdentifier) IsExpr()   {}

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
func (s *SystemVariable) expr()     {}
func (s *SystemVariable) IsExpr()   {}

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
func (v *ValueExpr) expr()     {}
func (v *ValueExpr) IsExpr()   {}

// Span returns the source span for this expression.
func (v *ValueExpr) Span() token.Span {
	return v.SpanVal
}

// String returns the SQL representation.
func (v *ValueExpr) String() string {
	if v.Value == nil {
		return "NULL"
	}
	// Handle Go bool types - canonical form is lowercase true/false (matches Rust)
	if b, ok := v.Value.(bool); ok {
		if b {
			return "true"
		}
		return "false"
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
func (q *QualifiedWildcard) expr()     {}
func (q *QualifiedWildcard) IsExpr()   {}

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
	SpanVal           token.Span
	AdditionalOptions *WildcardAdditionalOptions
}

func (w *Wildcard) exprNode() {}
func (w *Wildcard) expr()     {}
func (w *Wildcard) IsExpr()   {}

// Span returns the source span for this expression.
func (w *Wildcard) Span() token.Span {
	return w.SpanVal
}

// String returns the SQL representation.
func (w *Wildcard) String() string {
	if w.AdditionalOptions != nil {
		return "*" + w.AdditionalOptions.String()
	}
	return "*"
}

// Nested represents a nested expression in parentheses (Expr::Nested in Rust).
type Nested struct {
	SpanVal token.Span
	Expr    Expr
}

func (n *Nested) exprNode() {}
func (n *Nested) expr()     {}
func (n *Nested) IsExpr()   {}

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
func (p *Prefixed) expr()     {}
func (p *Prefixed) IsExpr()   {}

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
	SpanVal        token.Span
	DataType       string
	Value          string
	UsesOdbcSyntax bool // For ODBC syntax like {d '2025-07-17'}
}

func (t *TypedString) exprNode() {}
func (t *TypedString) expr()     {}
func (t *TypedString) IsExpr()   {}

// Span returns the source span for this expression.
func (t *TypedString) Span() token.Span {
	return t.SpanVal
}

// String returns the SQL representation.
func (t *TypedString) String() string {
	if t.UsesOdbcSyntax {
		switch t.DataType {
		case "DATE":
			return fmt.Sprintf("{d '%s'}", t.Value)
		case "TIME":
			return fmt.Sprintf("{t '%s'}", t.Value)
		case "TIMESTAMP":
			return fmt.Sprintf("{ts '%s'}", t.Value)
		}
	}
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

// String returns the SQL representation of the identifier.
func (i *Ident) String() string {
	if i.QuoteStyle != nil {
		q := *i.QuoteStyle
		fmt.Printf("DEBUG expr.Ident.String: Value=%q QuoteStyle=%q\n", i.Value, q)
		switch q {
		case '"', '\'', '`':
			escaped := escapeQuotedString(i.Value, q)
			return fmt.Sprintf("%c%s%c", q, escaped, q)
		case '[':
			result := fmt.Sprintf("[%s]", i.Value)
			fmt.Printf("DEBUG expr.Ident.String result: %q\n", result)
			return result
		}
	}
	return i.Value
}

// exprNode is a marker method that identifies this type as an expression node.
func (i *Ident) exprNode() {}
func (i *Ident) expr()     {}
func (i *Ident) IsExpr()   {}

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

// WildcardAdditionalOptions represents options for wildcards like EXCLUDE, EXCEPT, REPLACE, etc.
type WildcardAdditionalOptions struct {
	OptExclude *ExcludeSelectItem
	OptExcept  *ExceptSelectItem
	OptReplace *ReplaceSelectItem
	OptRename  *RenameSelectItem
}

// String returns the SQL representation.
func (w *WildcardAdditionalOptions) String() string {
	var parts []string
	if w.OptExclude != nil {
		parts = append(parts, w.OptExclude.String())
	}
	if w.OptExcept != nil {
		parts = append(parts, w.OptExcept.String())
	}
	if w.OptReplace != nil {
		parts = append(parts, w.OptReplace.String())
	}
	if w.OptRename != nil {
		parts = append(parts, w.OptRename.String())
	}
	if len(parts) > 0 {
		return " " + strings.Join(parts, " ")
	}
	return ""
}

// ExcludeSelectItem represents EXCLUDE clause for wildcards (e.g., * EXCLUDE(col1, col2)).
type ExcludeSelectItem struct {
	Columns []*ObjectNamePart
}

// String returns the SQL representation.
func (e *ExcludeSelectItem) String() string {
	if len(e.Columns) == 1 {
		return "EXCLUDE " + e.Columns[0].String()
	}
	var colStrs []string
	for _, col := range e.Columns {
		colStrs = append(colStrs, col.String())
	}
	return "EXCLUDE (" + strings.Join(colStrs, ", ") + ")"
}

// ExceptSelectItem represents EXCEPT clause for wildcards (BigQuery syntax).
type ExceptSelectItem struct {
	FirstElement       *Ident
	AdditionalElements []*Ident
}

// String returns the SQL representation.
func (e *ExceptSelectItem) String() string {
	if len(e.AdditionalElements) == 0 {
		return "EXCEPT " + e.FirstElement.String()
	}
	var colStrs []string
	colStrs = append(colStrs, e.FirstElement.String())
	for _, col := range e.AdditionalElements {
		colStrs = append(colStrs, col.String())
	}
	return "EXCEPT (" + strings.Join(colStrs, ", ") + ")"
}

// ReplaceSelectItem represents REPLACE clause for wildcards.
type ReplaceSelectItem struct {
	Elements []*ReplaceSelectElement
}

// String returns the SQL representation.
func (r *ReplaceSelectItem) String() string {
	var elemStrs []string
	for _, elem := range r.Elements {
		elemStrs = append(elemStrs, elem.String())
	}
	return "REPLACE (" + strings.Join(elemStrs, ", ") + ")"
}

// ReplaceSelectElement represents a single element in REPLACE clause.
type ReplaceSelectElement struct {
	Expr Expr
	Name *Ident
}

// String returns the SQL representation.
func (r *ReplaceSelectElement) String() string {
	return r.Expr.String() + " AS " + r.Name.String()
}

// RenameSelectItem represents RENAME clause for wildcards.
type RenameSelectItem struct {
	Elements []*RenameSelectElement
}

// String returns the SQL representation.
func (r *RenameSelectItem) String() string {
	var elemStrs []string
	for _, elem := range r.Elements {
		elemStrs = append(elemStrs, elem.String())
	}
	return "RENAME (" + strings.Join(elemStrs, ", ") + ")"
}

// RenameSelectElement represents a single element in RENAME clause.
type RenameSelectElement struct {
	Old *ObjectName
	New *Ident
}

// String returns the SQL representation.
func (r *RenameSelectElement) String() string {
	return r.Old.String() + " AS " + r.New.String()
}
