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

// This file is part of Phase 1.2 AST Package Consolidation.
// It merges expression types from the expr/ subpackage into the main ast package.
//
// Key changes:
// - Types from ast/expr/ are now in package ast with "E" prefix
// - Uses ast.Expr interface (with expr() and IsExpr() methods)
// - Embeds ast.ExpressionBase instead of defining separate base types
//
// TODO: After full migration, remove the old ast/expr/ directory

package ast

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// ============================================================================
// Compatibility Method
// ============================================================================

// exprNode is a temporary compatibility method that allows types from the
// old expr package to work with the new ast.Expr interface.
// Both expr() and exprNode() are implemented by ExpressionBase.
func (e *ExpressionBase) exprNode() {}

// ============================================================================
// Basic Expressions (merged from ast/expr/basic.go)
// ============================================================================

// EIdent represents a single identifier expression (was expr.Identifier).
type EIdent struct {
	ExpressionBase
	SpanVal token.Span
	Ident   *Ident
}

// Span returns the source span for this expression.
func (i *EIdent) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIdent) String() string {
	if i.Ident != nil {
		return i.Ident.String()
	}
	return ""
}

// ECompoundIdent represents a multi-part identifier (was expr.CompoundIdentifier).
type ECompoundIdent struct {
	ExpressionBase
	SpanVal token.Span
	Idents  []*Ident
}

// Span returns the source span for this expression.
func (c *ECompoundIdent) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *ECompoundIdent) String() string {
	parts := make([]string, len(c.Idents))
	for i, ident := range c.Idents {
		parts[i] = ident.String()
	}
	return strings.Join(parts, ".")
}

// EValue represents a literal value expression (was expr.ValueExpr).
type EValue struct {
	ExpressionBase
	SpanVal token.Span
	Value   interface{} // *ValueWithSpan or Value type
}

// Span returns the source span for this expression.
func (v *EValue) Span() token.Span { return v.SpanVal }

// String returns the SQL representation.
func (v *EValue) String() string {
	if v.Value == nil {
		return "NULL"
	}
	if s, ok := v.Value.(fmt.Stringer); ok {
		return s.String()
	}
	return fmt.Sprintf("%v", v.Value)
}

// EQualifiedWildcard represents a qualified wildcard (was expr.QualifiedWildcard).
type EQualifiedWildcard struct {
	ExpressionBase
	SpanVal token.Span
	Prefix  *ObjectName
}

// Span returns the source span for this expression.
func (q *EQualifiedWildcard) Span() token.Span { return q.SpanVal }

// String returns the SQL representation.
func (q *EQualifiedWildcard) String() string {
	if q.Prefix == nil {
		return "*"
	}
	return q.Prefix.String() + ".*"
}

// EWildcard represents an unqualified `*` wildcard (was expr.Wildcard).
type EWildcard struct {
	ExpressionBase
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (w *EWildcard) Span() token.Span { return w.SpanVal }

// String returns the SQL representation.
func (w *EWildcard) String() string { return "*" }

// ENested represents a nested expression in parentheses (was expr.Nested).
type ENested struct {
	ExpressionBase
	SpanVal token.Span
	Expr    Expr
}

// Span returns the source span for this expression.
func (n *ENested) Span() token.Span { return n.SpanVal }

// String returns the SQL representation.
func (n *ENested) String() string { return "(" + n.Expr.String() + ")" }

// EPrefixed represents a prefixed expression (was expr.Prefixed).
type EPrefixed struct {
	ExpressionBase
	SpanVal token.Span
	Prefix  *Ident
	Value   Expr
}

// Span returns the source span for this expression.
func (p *EPrefixed) Span() token.Span { return p.SpanVal }

// String returns the SQL representation.
func (p *EPrefixed) String() string {
	return p.Prefix.String() + " " + p.Value.String()
}

// ETypedString represents a typed string literal (was expr.TypedString).
type ETypedString struct {
	ExpressionBase
	SpanVal  token.Span
	DataType string
	Value    string
}

// Span returns the source span for this expression.
func (t *ETypedString) Span() token.Span { return t.SpanVal }

// String returns the SQL representation.
func (t *ETypedString) String() string {
	return fmt.Sprintf("%s '%s'", t.DataType, t.Value)
}

// ESqlOption represents a SQL option (was expr.SqlOption).
type ESqlOption struct {
	ExpressionBase
	SpanVal token.Span
	Name    *Ident
	Value   Expr
}

// Span returns the source span for this option.
func (s *ESqlOption) Span() token.Span { return s.SpanVal }

// String returns the SQL representation.
func (s *ESqlOption) String() string {
	valueStr := s.Value.String()
	if valExpr, ok := s.Value.(*EValue); ok {
		if _, isString := valExpr.Value.(string); isString {
			valueStr = fmt.Sprintf("'%s'", valueStr)
		}
	}
	return fmt.Sprintf("%s = %s", s.Name.String(), valueStr)
}

// EColumnOption represents a column option (was expr.ColumnOption).
type EColumnOption struct {
	ExpressionBase
	SpanVal token.Span
	Name    string
	Value   Expr
}

// Span returns the source span for this option.
func (c *EColumnOption) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *EColumnOption) String() string {
	if c.Value != nil {
		return fmt.Sprintf("%s %s", c.Name, c.Value.String())
	}
	return c.Name
}

// ============================================================================
// Complex Expressions (merged from ast/expr/complex.go)
// ============================================================================

// EArray represents an array expression (was expr.ArrayExpr).
type EArray struct {
	ExpressionBase
	SpanVal token.Span
	Elems   []Expr
	Named   bool // true for `ARRAY[...]`, false for `[...]`
}

// Span returns the source span for this expression.
func (a *EArray) Span() token.Span { return a.SpanVal }

// String returns the SQL representation.
func (a *EArray) String() string {
	parts := make([]string, len(a.Elems))
	for i, elem := range a.Elems {
		parts[i] = elem.String()
	}
	if a.Named {
		return fmt.Sprintf("ARRAY[%s]", strings.Join(parts, ", "))
	}
	return fmt.Sprintf("[%s]", strings.Join(parts, ", "))
}

// EInterval represents an INTERVAL expression (was expr.IntervalExpr).
type EInterval struct {
	ExpressionBase
	SpanVal                    token.Span
	Value                      Expr
	LeadingField               *string
	LeadingPrecision           *uint64
	LastField                  *string
	FractionalSecondsPrecision *uint64
}

// Span returns the source span for this expression.
func (i *EInterval) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EInterval) String() string {
	var sb strings.Builder
	sb.WriteString("INTERVAL ")
	sb.WriteString(i.Value.String())
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

// ETuple represents a tuple/row expression (was expr.TupleExpr).
type ETuple struct {
	ExpressionBase
	SpanVal token.Span
	Exprs   []Expr
}

// Span returns the source span for this expression.
func (t *ETuple) Span() token.Span { return t.SpanVal }

// String returns the SQL representation.
func (t *ETuple) String() string {
	parts := make([]string, len(t.Exprs))
	for i, expr := range t.Exprs {
		parts[i] = expr.String()
	}
	return fmt.Sprintf("(%s)", strings.Join(parts, ", "))
}

// EStructField represents a field definition within a struct (was expr.StructField).
type EStructField struct {
	SpanVal   token.Span
	FieldName *Ident
	FieldType string
	Options   []Expr // ESqlOption
}

// Span returns the source span for this struct field.
func (s *EStructField) Span() token.Span { return s.SpanVal }

// String returns the SQL representation.
func (s *EStructField) String() string {
	if s.FieldName != nil {
		return fmt.Sprintf("%s %s", s.FieldName.String(), s.FieldType)
	}
	return s.FieldType
}

// EStruct represents a struct literal expression (was expr.StructExpr).
type EStruct struct {
	ExpressionBase
	SpanVal token.Span
	Values  []Expr
	Fields  []EStructField
}

// Span returns the source span for this expression.
func (s *EStruct) Span() token.Span { return s.SpanVal }

// String returns the SQL representation.
func (s *EStruct) String() string {
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

// EMapEntry represents a key-value pair in a map (was expr.MapEntry).
type EMapEntry struct {
	SpanVal token.Span
	Key     Expr
	Value   Expr
}

// Span returns the source span for this map entry.
func (m *EMapEntry) Span() token.Span { return m.SpanVal }

// String returns the SQL representation.
func (m *EMapEntry) String() string {
	return fmt.Sprintf("%s: %s", m.Key.String(), m.Value.String())
}

// EMap represents a map literal expression (was expr.MapExpr).
type EMap struct {
	ExpressionBase
	SpanVal token.Span
	Entries []EMapEntry
}

// Span returns the source span for this expression.
func (m *EMap) Span() token.Span { return m.SpanVal }

// String returns the SQL representation.
func (m *EMap) String() string {
	entries := make([]string, len(m.Entries))
	for i, e := range m.Entries {
		entries[i] = e.String()
	}
	return fmt.Sprintf("MAP {%s}", strings.Join(entries, ", "))
}

// EDictionaryField represents a field in a dictionary struct (was expr.DictionaryField).
type EDictionaryField struct {
	SpanVal token.Span
	Key     *Ident
	Value   Expr
}

// Span returns the source span for this dictionary field.
func (d *EDictionaryField) Span() token.Span { return d.SpanVal }

// String returns the SQL representation.
func (d *EDictionaryField) String() string {
	return fmt.Sprintf("%s: %s", d.Key.String(), d.Value.String())
}

// EDictionary represents a DuckDB-style dictionary/struct literal (was expr.DictionaryExpr).
type EDictionary struct {
	ExpressionBase
	SpanVal token.Span
	Fields  []EDictionaryField
}

// Span returns the source span for this expression.
func (d *EDictionary) Span() token.Span { return d.SpanVal }

// String returns the SQL representation.
func (d *EDictionary) String() string {
	fields := make([]string, len(d.Fields))
	for i, f := range d.Fields {
		fields[i] = f.String()
	}
	return fmt.Sprintf("{%s}", strings.Join(fields, ", "))
}

// ENamed represents a named expression in a typeless struct (was expr.NamedExpr).
type ENamed struct {
	ExpressionBase
	SpanVal token.Span
	Expr    Expr
	Name    *Ident
}

// Span returns the source span for this expression.
func (n *ENamed) Span() token.Span { return n.SpanVal }

// String returns the SQL representation.
func (n *ENamed) String() string {
	return fmt.Sprintf("%s AS %s", n.Expr.String(), n.Name.String())
}

// EAccessExpr represents an element in a compound field access chain (was expr.AccessExpr).
type EAccessExpr interface {
	fmt.Stringer
	Span() token.Span
}

// EDotAccess represents dot notation access (was expr.DotAccess).
type EDotAccess struct {
	SpanVal token.Span
	Expr    Expr
}

// Span returns the source span for this access.
func (d *EDotAccess) Span() token.Span { return d.SpanVal }

// String returns the SQL representation.
func (d *EDotAccess) String() string { return "." + d.Expr.String() }

// ESubscriptAccess represents bracket notation access (was expr.SubscriptAccess).
type ESubscriptAccess struct {
	SpanVal   token.Span
	Subscript *ESubscript
}

// Span returns the source span for this access.
func (s *ESubscriptAccess) Span() token.Span { return s.SpanVal }

// String returns the SQL representation.
func (s *ESubscriptAccess) String() string { return "[" + s.Subscript.String() + "]" }

// ESubscript represents a subscript expression within brackets (was expr.Subscript).
type ESubscript struct {
	SpanVal    token.Span
	Index      Expr
	LowerBound *Expr
	UpperBound *Expr
	Stride     *Expr
}

// Span returns the source span for this subscript.
func (s *ESubscript) Span() token.Span { return s.SpanVal }

// String returns the SQL representation.
func (s *ESubscript) String() string {
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

// ECompoundFieldAccess represents accessing nested fields (was expr.CompoundFieldAccess).
type ECompoundFieldAccess struct {
	ExpressionBase
	SpanVal     token.Span
	Root        Expr
	AccessChain []EAccessExpr
}

// Span returns the source span for this expression.
func (c *ECompoundFieldAccess) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *ECompoundFieldAccess) String() string {
	var sb strings.Builder
	sb.WriteString(c.Root.String())
	for _, access := range c.AccessChain {
		sb.WriteString(access.String())
	}
	return sb.String()
}

// EJsonPathElem represents an element in a JSON path (was expr.JsonPathElem).
type EJsonPathElem interface {
	String() string
}

// EJsonPathDot represents dot notation in JSON path (was expr.JsonPathDot).
type EJsonPathDot struct {
	Key    string
	Quoted bool
}

// String returns the SQL representation.
// Note: This returns just the key without prefix. The container (EJsonPath)
// adds the appropriate prefix (":" or ".") based on context.
func (j *EJsonPathDot) String() string {
	if j.Quoted {
		return fmt.Sprintf("\"%s\"", j.Key)
	}
	return j.Key
}

// EJsonPathBracket represents bracket notation in JSON path (was expr.JsonPathBracket).
type EJsonPathBracket struct {
	Key Expr
}

// String returns the SQL representation.
func (j *EJsonPathBracket) String() string { return fmt.Sprintf("[%s]", j.Key.String()) }

// EJsonPathColonBracket represents colon-bracket notation in JSON path (was expr.JsonPathColonBracket).
type EJsonPathColonBracket struct {
	Key Expr
}

// String returns the SQL representation.
func (j *EJsonPathColonBracket) String() string { return fmt.Sprintf(":[%s]", j.Key.String()) }

// EJsonPath represents a JSON path for accessing semi-structured data (was expr.JsonPath).
type EJsonPath struct {
	Path []EJsonPathElem
}

// String returns the SQL representation.
func (j *EJsonPath) String() string {
	var sb strings.Builder
	for i, elem := range j.Path {
		if i == 0 {
			// First element: add ":" prefix if it's a EJsonPathDot
			// (colon-style access like a:b)
			if _, ok := elem.(*EJsonPathDot); ok {
				sb.WriteString(":")
			}
		} else {
			// Subsequent elements: always add "." prefix for EJsonPathDot
			// (dot-style access like a.b or a:b.c)
			if _, ok := elem.(*EJsonPathDot); ok {
				sb.WriteString(".")
			}
		}
		sb.WriteString(elem.String())
	}
	return sb.String()
}

// EJsonAccess represents accessing semi-structured data (was expr.JsonAccess).
type EJsonAccess struct {
	ExpressionBase
	SpanVal token.Span
	Value   Expr
	Path    *EJsonPath
}

// Span returns the source span for this expression.
func (j *EJsonAccess) Span() token.Span { return j.SpanVal }

// String returns the SQL representation.
func (j *EJsonAccess) String() string { return j.Value.String() + j.Path.String() }

// EOuterJoin represents the Oracle-style outer join operator `(+)` (was expr.OuterJoin).
type EOuterJoin struct {
	ExpressionBase
	SpanVal token.Span
	Expr    Expr
}

// Span returns the source span for this expression.
func (o *EOuterJoin) Span() token.Span { return o.SpanVal }

// String returns the SQL representation.
func (o *EOuterJoin) String() string { return o.Expr.String() + " (+)" }

// EPrior represents a reference to the prior level in a CONNECT BY clause (was expr.PriorExpr).
type EPrior struct {
	ExpressionBase
	SpanVal token.Span
	Expr    Expr
}

// Span returns the source span for this expression.
func (p *EPrior) Span() token.Span { return p.SpanVal }

// String returns the SQL representation.
func (p *EPrior) String() string { return "PRIOR " + p.Expr.String() }

// ============================================================================
// Lambda Expressions (merged from ast/expr/complex.go)
// ============================================================================

// ELambdaParam represents a parameter to a lambda function (was expr.LambdaFunctionParameter).
type ELambdaParam struct {
	SpanVal  token.Span
	Name     *Ident
	DataType string // optional
}

// Span returns the source span for this parameter.
func (l *ELambdaParam) Span() token.Span { return l.SpanVal }

// String returns the SQL representation.
func (l *ELambdaParam) String() string {
	if l.DataType != "" {
		return fmt.Sprintf("%s %s", l.Name.String(), l.DataType)
	}
	return l.Name.String()
}

// ELambdaSyntax represents the syntax style for lambda functions.
type ELambdaSyntax int

const (
	ELambdaArrow ELambdaSyntax = iota
	ELambdaKeyword
)

// ELambda represents a lambda function expression (was expr.LambdaExpr).
type ELambda struct {
	ExpressionBase
	SpanVal token.Span
	Params  []ELambdaParam
	Body    Expr
	Syntax  ELambdaSyntax
}

// Span returns the source span for this expression.
func (l *ELambda) Span() token.Span { return l.SpanVal }

// String returns the SQL representation.
func (l *ELambda) String() string {
	params := make([]string, len(l.Params))
	for i, p := range l.Params {
		params[i] = p.String()
	}
	if l.Syntax == ELambdaKeyword {
		return fmt.Sprintf("lambda %s : %s", strings.Join(params, ", "), l.Body.String())
	}
	if len(l.Params) == 1 {
		return fmt.Sprintf("%s -> %s", l.Params[0].String(), l.Body.String())
	}
	return fmt.Sprintf("(%s) -> %s", strings.Join(params, ", "), l.Body.String())
}

// EMemberOf represents a MEMBER OF expression (was expr.MemberOfExpr).
type EMemberOf struct {
	ExpressionBase
	SpanVal token.Span
	Value   Expr
	Array   Expr
}

// Span returns the source span for this expression.
func (m *EMemberOf) Span() token.Span { return m.SpanVal }

// String returns the SQL representation.
func (m *EMemberOf) String() string {
	return fmt.Sprintf("%s MEMBER OF(%s)", m.Value.String(), m.Array.String())
}

// ESearchModifier represents the search modifier for MATCH AGAINST.
type ESearchModifier int

const (
	ESearchNaturalLanguage ESearchModifier = iota
	ESearchNaturalLanguageWithQueryExpansion
	ESearchBooleanMode
	ESearchWithQueryExpansion
)

// String returns the SQL representation.
func (s ESearchModifier) String() string {
	switch s {
	case ESearchNaturalLanguage:
		return "IN NATURAL LANGUAGE MODE"
	case ESearchNaturalLanguageWithQueryExpansion:
		return "IN NATURAL LANGUAGE MODE WITH QUERY EXPANSION"
	case ESearchBooleanMode:
		return "IN BOOLEAN MODE"
	case ESearchWithQueryExpansion:
		return "WITH QUERY EXPANSION"
	}
	return ""
}

// EMatchAgainst represents a MySQL MATCH AGAINST full-text search (was expr.MatchAgainstExpr).
type EMatchAgainst struct {
	ExpressionBase
	SpanVal           token.Span
	Columns           []*ObjectName
	MatchValue        interface{} // ValueWithSpan
	OptSearchModifier *ESearchModifier
}

// Span returns the source span for this expression.
func (m *EMatchAgainst) Span() token.Span { return m.SpanVal }

// String returns the SQL representation.
func (m *EMatchAgainst) String() string {
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

// ============================================================================
// Conditional Expressions (merged from ast/expr/conditional.go)
// ============================================================================

// ECaseWhen represents a WHEN clause in a CASE expression (was expr.CaseWhen).
type ECaseWhen struct {
	Condition Expr
	Result    Expr
}

// String returns the SQL representation.
func (c *ECaseWhen) String() string {
	return fmt.Sprintf("WHEN %s THEN %s", c.Condition.String(), c.Result.String())
}

// ECase represents a CASE expression (was expr.CaseExpr).
type ECase struct {
	ExpressionBase
	SpanVal    token.Span
	Operand    Expr
	Conditions []ECaseWhen
	ElseResult Expr
}

// Span returns the source span for this expression.
func (c *ECase) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *ECase) String() string {
	var sb strings.Builder
	sb.WriteString("CASE")
	if c.Operand != nil {
		sb.WriteString(" ")
		sb.WriteString(c.Operand.String())
	}
	for _, when := range c.Conditions {
		sb.WriteString(" ")
		sb.WriteString(when.String())
	}
	if c.ElseResult != nil {
		sb.WriteString(" ELSE ")
		sb.WriteString(c.ElseResult.String())
	}
	sb.WriteString(" END")
	return sb.String()
}

// EIf represents an IF expression (was expr.IfExpr).
type EIf struct {
	ExpressionBase
	SpanVal    token.Span
	Condition  Expr
	TrueValue  Expr
	FalseValue Expr
}

// Span returns the source span for this expression.
func (i *EIf) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIf) String() string {
	return fmt.Sprintf("IF(%s, %s, %s)",
		i.Condition.String(), i.TrueValue.String(), i.FalseValue.String())
}

// ECoalesce represents a COALESCE expression (was expr.CoalesceExpr).
type ECoalesce struct {
	ExpressionBase
	SpanVal token.Span
	Exprs   []Expr
}

// Span returns the source span for this expression.
func (c *ECoalesce) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *ECoalesce) String() string {
	exprs := make([]string, len(c.Exprs))
	for i, expr := range c.Exprs {
		exprs[i] = expr.String()
	}
	return fmt.Sprintf("COALESCE(%s)", strings.Join(exprs, ", "))
}

// ENullIf represents a NULLIF expression (was expr.NullIfExpr).
type ENullIf struct {
	ExpressionBase
	SpanVal token.Span
	Expr1   Expr
	Expr2   Expr
}

// Span returns the source span for this expression.
func (n *ENullIf) Span() token.Span { return n.SpanVal }

// String returns the SQL representation.
func (n *ENullIf) String() string {
	return fmt.Sprintf("NULLIF(%s, %s)", n.Expr1.String(), n.Expr2.String())
}

// EIfNull represents an IFNULL expression (was expr.IfNullExpr).
type EIfNull struct {
	ExpressionBase
	SpanVal token.Span
	Expr1   Expr
	Expr2   Expr
}

// Span returns the source span for this expression.
func (i *EIfNull) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIfNull) String() string {
	return fmt.Sprintf("IFNULL(%s, %s)", i.Expr1.String(), i.Expr2.String())
}

// EGreatest represents a GREATEST expression (was expr.GreatestExpr).
type EGreatest struct {
	ExpressionBase
	SpanVal token.Span
	Exprs   []Expr
}

// Span returns the source span for this expression.
func (g *EGreatest) Span() token.Span { return g.SpanVal }

// String returns the SQL representation.
func (g *EGreatest) String() string {
	exprs := make([]string, len(g.Exprs))
	for i, expr := range g.Exprs {
		exprs[i] = expr.String()
	}
	return fmt.Sprintf("GREATEST(%s)", strings.Join(exprs, ", "))
}

// ELeast represents a LEAST expression (was expr.LeastExpr).
type ELeast struct {
	ExpressionBase
	SpanVal token.Span
	Exprs   []Expr
}

// Span returns the source span for this expression.
func (l *ELeast) Span() token.Span { return l.SpanVal }

// String returns the SQL representation.
func (l *ELeast) String() string {
	exprs := make([]string, len(l.Exprs))
	for i, expr := range l.Exprs {
		exprs[i] = expr.String()
	}
	return fmt.Sprintf("LEAST(%s)", strings.Join(exprs, ", "))
}

// ============================================================================
// Subquery Expressions (merged from ast/expr/subqueries.go)
// ============================================================================

// EExists represents an EXISTS expression (was expr.Exists).
type EExists struct {
	ExpressionBase
	Subquery *EQueryExpr
	Negated  bool
	SpanVal  token.Span
}

// Span returns the source span for this expression.
func (e *EExists) Span() token.Span { return e.SpanVal }

// String returns the SQL representation.
func (e *EExists) String() string {
	if e.Negated {
		return fmt.Sprintf("NOT EXISTS (%s)", e.Subquery.String())
	}
	return fmt.Sprintf("EXISTS (%s)", e.Subquery.String())
}

// ESubquery represents a scalar subquery expression (was expr.Subquery).
type ESubquery struct {
	ExpressionBase
	Query   *EQueryExpr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (s *ESubquery) Span() token.Span { return s.SpanVal }

// String returns the SQL representation.
func (s *ESubquery) String() string {
	return fmt.Sprintf("(%s)", s.Query.String())
}

// EQueryExpr represents a SELECT query (was expr.QueryExpr).
// This is a placeholder that avoids circular imports.
type EQueryExpr struct {
	SQL       string
	Statement interface{} // Can hold an ast.Statement
	SpanVal   token.Span
}

// Span returns the source span for this query.
func (q *EQueryExpr) Span() token.Span { return q.SpanVal }

// String returns the SQL representation.
func (q *EQueryExpr) String() string {
	if q.Statement != nil {
		if s, ok := q.Statement.(interface{ String() string }); ok {
			return s.String()
		}
	}
	return q.SQL
}

// EGroupingSets represents a GROUPING SETS expression (was expr.GroupingSets).
type EGroupingSets struct {
	ExpressionBase
	Sets    [][]Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (g *EGroupingSets) Span() token.Span { return g.SpanVal }

// String returns the SQL representation.
func (g *EGroupingSets) String() string {
	var sb strings.Builder
	sb.WriteString("GROUPING SETS (")
	sets := make([]string, len(g.Sets))
	for i, set := range g.Sets {
		items := make([]string, len(set))
		for j, item := range set {
			items[j] = item.String()
		}
		sets[i] = "(" + strings.Join(items, ", ") + ")"
	}
	sb.WriteString(strings.Join(sets, ", "))
	sb.WriteString(")")
	return sb.String()
}

// ECube represents a CUBE expression (was expr.Cube).
type ECube struct {
	ExpressionBase
	Sets    [][]Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (c *ECube) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *ECube) String() string {
	var sb strings.Builder
	sb.WriteString("CUBE (")
	sets := make([]string, len(c.Sets))
	for i, set := range c.Sets {
		if len(set) == 1 {
			sets[i] = set[0].String()
		} else {
			items := make([]string, len(set))
			for j, item := range set {
				items[j] = item.String()
			}
			sets[i] = "(" + strings.Join(items, ", ") + ")"
		}
	}
	sb.WriteString(strings.Join(sets, ", "))
	sb.WriteString(")")
	return sb.String()
}

// ERollup represents a ROLLUP expression (was expr.Rollup).
type ERollup struct {
	ExpressionBase
	Sets    [][]Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (r *ERollup) Span() token.Span { return r.SpanVal }

// String returns the SQL representation.
func (r *ERollup) String() string {
	var sb strings.Builder
	sb.WriteString("ROLLUP (")
	sets := make([]string, len(r.Sets))
	for i, set := range r.Sets {
		if len(set) == 1 {
			sets[i] = set[0].String()
		} else {
			items := make([]string, len(set))
			for j, item := range set {
				items[j] = item.String()
			}
			sets[i] = "(" + strings.Join(items, ", ") + ")"
		}
	}
	sb.WriteString(strings.Join(sets, ", "))
	sb.WriteString(")")
	return sb.String()
}
