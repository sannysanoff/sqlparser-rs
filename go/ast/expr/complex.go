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

package expr

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// ArrayExpr represents an array expression (Expr::Array in Rust).
type ArrayExpr struct {
	SpanVal token.Span
	Elems   []Expr
	Named   bool // true for `ARRAY[...]`, false for `[...]`
}

func (a *ArrayExpr) exprNode() {}

// Span returns the source span for this expression.
func (a *ArrayExpr) Span() token.Span {
	return a.SpanVal
}

// String returns the SQL representation.
func (a *ArrayExpr) String() string {
	parts := make([]string, len(a.Elems))
	for i, elem := range a.Elems {
		parts[i] = elem.String()
	}
	if a.Named {
		return fmt.Sprintf("ARRAY[%s]", strings.Join(parts, ", "))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

// IntervalExpr represents an INTERVAL expression (Expr::Interval in Rust).
type IntervalExpr struct {
	SpanVal                    token.Span
	Value                      Expr
	LeadingField               *string // e.g., YEAR, MONTH, DAY
	LeadingPrecision           *uint64
	LastField                  *string
	FractionalSecondsPrecision *uint64
}

func (i *IntervalExpr) exprNode() {}

// Span returns the source span for this expression.
func (i *IntervalExpr) Span() token.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IntervalExpr) String() string {
	var sb strings.Builder
	sb.WriteString("INTERVAL ")
	sb.WriteString(i.Value.String())

	// Special handling for SECOND with both precisions (e.g., SECOND (5, 4))
	if i.LeadingField != nil && *i.LeadingField == "SECOND" &&
		i.LeadingPrecision != nil && i.FractionalSecondsPrecision != nil {
		sb.WriteString(" SECOND (")
		sb.WriteString(fmt.Sprintf("%d", *i.LeadingPrecision))
		sb.WriteString(", ")
		sb.WriteString(fmt.Sprintf("%d", *i.FractionalSecondsPrecision))
		sb.WriteString(")")
		return sb.String()
	}

	if i.LeadingField != nil {
		sb.WriteString(" ")
		sb.WriteString(*i.LeadingField)
		if i.LeadingPrecision != nil {
			sb.WriteString(fmt.Sprintf(" (%d)", *i.LeadingPrecision))
		}
	}

	if i.LastField != nil {
		sb.WriteString(" TO ")
		sb.WriteString(*i.LastField)
		if i.FractionalSecondsPrecision != nil {
			sb.WriteString(fmt.Sprintf(" (%d)", *i.FractionalSecondsPrecision))
		}
	}

	return sb.String()
}

// TupleExpr represents a tuple/row expression (Expr::Tuple in Rust).
type TupleExpr struct {
	SpanVal token.Span
	Exprs   []Expr
}

func (t *TupleExpr) exprNode() {}

// Span returns the source span for this expression.
func (t *TupleExpr) Span() token.Span {
	return t.SpanVal
}

// String returns the SQL representation.
func (t *TupleExpr) String() string {
	parts := make([]string, len(t.Exprs))
	for i, expr := range t.Exprs {
		parts[i] = expr.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
}

// StructField represents a field definition within a struct.
type StructField struct {
	SpanVal   token.Span
	FieldName *Ident
	FieldType string
	Options   []Expr // SqlOption
}

// Span returns the source span for this struct field.
func (s *StructField) Span() token.Span {
	return s.SpanVal
}

// String returns the SQL representation.
func (s *StructField) String() string {
	if s.FieldName != nil {
		return fmt.Sprintf("%s %s", s.FieldName.String(), s.FieldType)
	}
	return s.FieldType
}

// StructExpr represents a struct literal expression (Expr::Struct in Rust).
type StructExpr struct {
	SpanVal token.Span
	Values  []Expr
	Fields  []StructField
}

func (s *StructExpr) exprNode() {}

// Span returns the source span for this expression.
func (s *StructExpr) Span() token.Span {
	return s.SpanVal
}

// String returns the SQL representation.
func (s *StructExpr) String() string {
	values := make([]string, len(s.Values))
	for i, v := range s.Values {
		values[i] = v.String()
	}

	if len(s.Fields) > 0 {
		fields := make([]string, len(s.Fields))
		for i, f := range s.Fields {
			fields[i] = f.String()
		}
		return fmt.Sprintf("STRUCT<%s>(%s)", strings.Join(fields, ", "), strings.Join(values, ", "))
	}

	return fmt.Sprintf("STRUCT(%s)", strings.Join(values, ", "))
}

// MapEntry represents a key-value pair in a map.
type MapEntry struct {
	SpanVal token.Span
	Key     Expr
	Value   Expr
}

// Span returns the source span for this map entry.
func (m *MapEntry) Span() token.Span {
	return m.SpanVal
}

// String returns the SQL representation.
func (m *MapEntry) String() string {
	return fmt.Sprintf("%s: %s", m.Key.String(), m.Value.String())
}

// MapExpr represents a map literal expression (Expr::Map in Rust).
type MapExpr struct {
	SpanVal token.Span
	Entries []MapEntry
}

func (m *MapExpr) exprNode() {}

// Span returns the source span for this expression.
func (m *MapExpr) Span() token.Span {
	return m.SpanVal
}

// String returns the SQL representation.
func (m *MapExpr) String() string {
	entries := make([]string, len(m.Entries))
	for i, e := range m.Entries {
		entries[i] = e.String()
	}
	return fmt.Sprintf("MAP {%s}", strings.Join(entries, ", "))
}

// DictionaryField represents a field in a dictionary struct.
type DictionaryField struct {
	SpanVal token.Span
	Key     *Ident
	Value   Expr
}

// Span returns the source span for this dictionary field.
func (d *DictionaryField) Span() token.Span {
	return d.SpanVal
}

// String returns the SQL representation.
func (d *DictionaryField) String() string {
	return fmt.Sprintf("%s: %s", d.Key.String(), d.Value.String())
}

// DictionaryExpr represents a DuckDB-style dictionary/struct literal (Expr::Dictionary in Rust).
type DictionaryExpr struct {
	SpanVal token.Span
	Fields  []DictionaryField
}

func (d *DictionaryExpr) exprNode() {}

// Span returns the source span for this expression.
func (d *DictionaryExpr) Span() token.Span {
	return d.SpanVal
}

// String returns the SQL representation.
func (d *DictionaryExpr) String() string {
	fields := make([]string, len(d.Fields))
	for i, f := range d.Fields {
		fields[i] = f.String()
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

// NamedExpr represents a named expression in a typeless struct (Expr::Named in Rust).
type NamedExpr struct {
	SpanVal token.Span
	Expr    Expr
	Name    *Ident
}

func (n *NamedExpr) exprNode() {}

// Span returns the source span for this expression.
func (n *NamedExpr) Span() token.Span {
	return n.SpanVal
}

// String returns the SQL representation.
func (n *NamedExpr) String() string {
	return fmt.Sprintf("%s AS %s", n.Expr.String(), n.Name.String())
}

// AccessExpr represents an element in a compound field access chain.
type AccessExpr interface {
	Span() token.Span
	String() string
}

// DotAccess represents dot notation access (e.g., `foo.bar`).
type DotAccess struct {
	SpanVal token.Span
	Expr    Expr
}

// Span returns the source span for this access.
func (d *DotAccess) Span() token.Span {
	return d.SpanVal
}

// String returns the SQL representation.
func (d *DotAccess) String() string {
	return "." + d.Expr.String()
}

// SubscriptAccess represents bracket notation access (e.g., `foo[0]`).
type SubscriptAccess struct {
	SpanVal   token.Span
	Subscript *Subscript
}

// Span returns the source span for this access.
func (s *SubscriptAccess) Span() token.Span {
	return s.SpanVal
}

// String returns the SQL representation.
func (s *SubscriptAccess) String() string {
	return "[" + s.Subscript.String() + "]"
}

// Subscript represents a subscript expression within brackets.
type Subscript struct {
	SpanVal    token.Span
	Index      Expr
	LowerBound *Expr
	UpperBound *Expr
	Stride     *Expr
}

// Span returns the source span for this subscript.
func (s *Subscript) Span() token.Span {
	return s.SpanVal
}

// String returns the SQL representation.
func (s *Subscript) String() string {
	if s.Index != nil {
		return s.Index.String()
	}

	var sb strings.Builder
	if s.LowerBound != nil {
		sb.WriteString((*s.LowerBound).String())
	}
	sb.WriteString(":")
	if s.UpperBound != nil {
		sb.WriteString((*s.UpperBound).String())
	}
	if s.Stride != nil {
		sb.WriteString(":")
		sb.WriteString((*s.Stride).String())
	}
	return sb.String()
}

// CompoundFieldAccess represents accessing nested fields (Expr::CompoundFieldAccess in Rust).
type CompoundFieldAccess struct {
	SpanVal     token.Span
	Root        Expr
	AccessChain []AccessExpr
}

func (c *CompoundFieldAccess) exprNode() {}

// Span returns the source span for this expression.
func (c *CompoundFieldAccess) Span() token.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *CompoundFieldAccess) String() string {
	var sb strings.Builder
	sb.WriteString(c.Root.String())
	for _, access := range c.AccessChain {
		sb.WriteString(access.String())
	}
	return sb.String()
}

// JsonPathElem represents an element in a JSON path.
type JsonPathElem interface {
	String() string
}

// JsonPathDot represents dot notation in JSON path.
type JsonPathDot struct {
	Key    string
	Quoted bool
}

// String returns the SQL representation.
func (j *JsonPathDot) String() string {
	if j.Quoted {
		return fmt.Sprintf(".\"%s\"", j.Key)
	}
	return "." + j.Key
}

// JsonPathBracket represents bracket notation in JSON path.
type JsonPathBracket struct {
	Key Expr
}

// String returns the SQL representation.
func (j *JsonPathBracket) String() string {
	return fmt.Sprintf("[%s]", j.Key.String())
}

// JsonPathColonBracket represents colon-bracket notation in JSON path.
type JsonPathColonBracket struct {
	Key Expr
}

// String returns the SQL representation.
func (j *JsonPathColonBracket) String() string {
	return fmt.Sprintf(":[%s]", j.Key.String())
}

// JsonPath represents a JSON path for accessing semi-structured data.
type JsonPath struct {
	Path []JsonPathElem
}

// String returns the SQL representation.
func (j *JsonPath) String() string {
	var sb strings.Builder
	for i, elem := range j.Path {
		if i == 0 {
			if _, ok := elem.(*JsonPathDot); ok {
				sb.WriteString(":")
			}
		}
		sb.WriteString(elem.String())
	}
	return sb.String()
}

// JsonAccess represents accessing semi-structured data (Expr::JsonAccess in Rust).
type JsonAccess struct {
	SpanVal token.Span
	Value   Expr
	Path    *JsonPath
}

func (j *JsonAccess) exprNode() {}

// Span returns the source span for this expression.
func (j *JsonAccess) Span() token.Span {
	return j.SpanVal
}

// String returns the SQL representation.
func (j *JsonAccess) String() string {
	return j.Value.String() + j.Path.String()
}

// OuterJoin represents the Oracle-style outer join operator `(+)` (Expr::OuterJoin in Rust).
type OuterJoin struct {
	SpanVal token.Span
	Expr    Expr
}

func (o *OuterJoin) exprNode() {}

// Span returns the source span for this expression.
func (o *OuterJoin) Span() token.Span {
	return o.SpanVal
}

// String returns the SQL representation.
func (o *OuterJoin) String() string {
	return o.Expr.String() + " (+)"
}

// PriorExpr represents a reference to the prior level in a CONNECT BY clause (Expr::Prior in Rust).
type PriorExpr struct {
	SpanVal token.Span
	Expr    Expr
}

func (p *PriorExpr) exprNode() {}

// Span returns the source span for this expression.
func (p *PriorExpr) Span() token.Span {
	return p.SpanVal
}

// String returns the SQL representation.
func (p *PriorExpr) String() string {
	return "PRIOR " + p.Expr.String()
}

// LambdaFunctionParameter represents a parameter to a lambda function.
type LambdaFunctionParameter struct {
	SpanVal  token.Span
	Name     *Ident
	DataType string // optional
}

// Span returns the source span for this parameter.
func (l *LambdaFunctionParameter) Span() token.Span {
	return l.SpanVal
}

// String returns the SQL representation.
func (l *LambdaFunctionParameter) String() string {
	if l.DataType != "" {
		return fmt.Sprintf("%s %s", l.Name.String(), l.DataType)
	}
	return l.Name.String()
}

// LambdaSyntax represents the syntax style for lambda functions.
type LambdaSyntax int

const (
	LambdaArrow LambdaSyntax = iota
	LambdaKeyword
)

// LambdaExpr represents a lambda function expression (Expr::Lambda in Rust).
type LambdaExpr struct {
	SpanVal token.Span
	Params  []LambdaFunctionParameter
	Body    Expr
	Syntax  LambdaSyntax
}

func (l *LambdaExpr) exprNode() {}

// Span returns the source span for this expression.
func (l *LambdaExpr) Span() token.Span {
	return l.SpanVal
}

// String returns the SQL representation.
func (l *LambdaExpr) String() string {
	params := make([]string, len(l.Params))
	for i, p := range l.Params {
		params[i] = p.String()
	}

	if l.Syntax == LambdaKeyword {
		return fmt.Sprintf("lambda %s : %s", strings.Join(params, ", "), l.Body.String())
	}

	if len(l.Params) == 1 {
		return fmt.Sprintf("%s -> %s", l.Params[0].String(), l.Body.String())
	}
	return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), l.Body.String())
}

// MemberOfExpr represents a MEMBER OF expression (Expr::MemberOf in Rust).
type MemberOfExpr struct {
	SpanVal token.Span
	Value   Expr
	Array   Expr
}

func (m *MemberOfExpr) exprNode() {}

// Span returns the source span for this expression.
func (m *MemberOfExpr) Span() token.Span {
	return m.SpanVal
}

// String returns the SQL representation.
func (m *MemberOfExpr) String() string {
	return fmt.Sprintf("%s MEMBER OF(%s)", m.Value.String(), m.Array.String())
}

// SearchModifier represents the search modifier for MATCH AGAINST.
type SearchModifier int

const (
	SearchNaturalLanguage SearchModifier = iota
	SearchNaturalLanguageWithQueryExpansion
	SearchBooleanMode
	SearchWithQueryExpansion
)

// String returns the SQL representation.
func (s SearchModifier) String() string {
	switch s {
	case SearchNaturalLanguage:
		return "IN NATURAL LANGUAGE MODE"
	case SearchNaturalLanguageWithQueryExpansion:
		return "IN NATURAL LANGUAGE MODE WITH QUERY EXPANSION"
	case SearchBooleanMode:
		return "IN BOOLEAN MODE"
	case SearchWithQueryExpansion:
		return "WITH QUERY EXPANSION"
	}
	return ""
}

// MatchAgainstExpr represents a MySQL MATCH AGAINST full-text search (Expr::MatchAgainst in Rust).
type MatchAgainstExpr struct {
	SpanVal           token.Span
	Columns           []*ObjectName
	MatchValue        interface{} // ValueWithSpan
	OptSearchModifier *SearchModifier
}

func (m *MatchAgainstExpr) exprNode() {}

// Span returns the source span for this expression.
func (m *MatchAgainstExpr) Span() token.Span {
	return m.SpanVal
}

// String returns the SQL representation.
func (m *MatchAgainstExpr) String() string {
	columns := make([]string, len(m.Columns))
	for i, c := range m.Columns {
		columns[i] = c.String()
	}

	result := fmt.Sprintf("MATCH (%s) AGAINST ", strings.Join(columns, ", "))

	if m.OptSearchModifier != nil {
		result += fmt.Sprintf("(%v %s)", m.MatchValue, m.OptSearchModifier.String())
	} else {
		result += fmt.Sprintf("(%v)", m.MatchValue)
	}

	return result
}
