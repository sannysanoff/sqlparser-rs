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

package query

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// With represents a WITH clause (common table expressions)
type With struct {
	span      token.Span
	Recursive bool
	CteTables []CTE
}

func (w *With) String() string {
	parts := []string{"WITH"}
	if w.Recursive {
		parts = append(parts, "RECURSIVE")
	}
	ctes := make([]string, len(w.CteTables))
	for i, cte := range w.CteTables {
		ctes[i] = cte.String()
	}
	parts = append(parts, strings.Join(ctes, ", "))
	return strings.Join(parts, " ")
}

// CTE represents a single Common Table Expression
type CTE struct {
	span         token.Span
	Alias        TableAlias
	Query        *Query
	From         *Ident
	Materialized *CteAsMaterialized
}

func (c *CTE) String() string {
	var parts []string
	parts = append(parts, c.Alias.String())
	parts = append(parts, "AS")
	if c.Materialized != nil {
		parts = append(parts, c.Materialized.String())
	}
	parts = append(parts, fmt.Sprintf("(%s)", c.Query.String()))
	if c.From != nil {
		parts = append(parts, "FROM", c.From.String())
	}
	return strings.Join(parts, " ")
}

// CteAsMaterialized indicates whether a CTE is materialized or not
type CteAsMaterialized int

const (
	CteMaterializedDefault CteAsMaterialized = iota
	CteMaterializedYes
	CteMaterializedNo
)

func (c CteAsMaterialized) String() string {
	switch c {
	case CteMaterializedYes:
		return "MATERIALIZED"
	case CteMaterializedNo:
		return "NOT MATERIALIZED"
	default:
		return ""
	}
}

// OrderBy represents an ORDER BY clause
type OrderBy struct {
	span        token.Span
	Kind        OrderByKind
	Interpolate *Interpolate
}

func (o *OrderBy) String() string {
	var parts []string
	parts = append(parts, "ORDER BY")
	parts = append(parts, o.Kind.String())
	if o.Interpolate != nil {
		parts = append(parts, o.Interpolate.String())
	}
	return strings.Join(parts, " ")
}

// OrderByKind represents the kind of ORDER BY (expressions or ALL)
type OrderByKind interface {
	fmt.Stringer
}

// OrderByExpressions represents a list of ordering expressions
type OrderByExpressions struct {
	Exprs []OrderByExpr
}

func (o *OrderByExpressions) String() string {
	parts := make([]string, len(o.Exprs))
	for i, e := range o.Exprs {
		parts[i] = e.String()
	}
	return strings.Join(parts, ", ")
}

// OrderByAll represents ORDER BY ALL with optional modifiers
type OrderByAll struct {
	Options OrderByOptions
}

func (o *OrderByAll) String() string {
	return "ALL" + o.Options.String()
}

// OrderByExpr represents an expression in ORDER BY
type OrderByExpr struct {
	span     token.Span
	Expr     Expr
	Options  OrderByOptions
	WithFill *WithFill
}

func (o *OrderByExpr) String() string {
	result := o.Expr.String()
	optionsStr := o.Options.String()
	if optionsStr != "" {
		result += " " + optionsStr
	}
	if o.WithFill != nil {
		result += " " + o.WithFill.String()
	}
	return result
}

// exprNode is a marker method that identifies this type as an expression node.
func (o *OrderByExpr) exprNode() {}

// Span returns the source span of this node
func (o *OrderByExpr) Span() token.Span { return o.span }

// OrderByOptions represents ASC/DESC and NULLS FIRST/LAST options
type OrderByOptions struct {
	Asc        *bool
	NullsFirst *bool
}

func (o *OrderByOptions) String() string {
	var parts []string
	if o.Asc != nil {
		if *o.Asc {
			parts = append(parts, "ASC")
		} else {
			parts = append(parts, "DESC")
		}
	}
	if o.NullsFirst != nil {
		if *o.NullsFirst {
			parts = append(parts, "NULLS FIRST")
		} else {
			parts = append(parts, "NULLS LAST")
		}
	}
	return strings.Join(parts, " ")
}

// WithFill represents ClickHouse WITH FILL modifier
type WithFill struct {
	From Expr
	To   Expr
	Step Expr
}

func (w *WithFill) String() string {
	parts := []string{"WITH FILL"}
	if w.From != nil {
		parts = append(parts, fmt.Sprintf("FROM %s", w.From.String()))
	}
	if w.To != nil {
		parts = append(parts, fmt.Sprintf("TO %s", w.To.String()))
	}
	if w.Step != nil {
		parts = append(parts, fmt.Sprintf("STEP %s", w.Step.String()))
	}
	return strings.Join(parts, " ")
}

// Interpolate represents ClickHouse INTERPOLATE clause
type Interpolate struct {
	Exprs *[]InterpolateExpr
}

func (i *Interpolate) String() string {
	if i.Exprs != nil && len(*i.Exprs) > 0 {
		parts := make([]string, len(*i.Exprs))
		for j, e := range *i.Exprs {
			parts[j] = e.String()
		}
		return fmt.Sprintf("INTERPOLATE (%s)", strings.Join(parts, ", "))
	}
	return "INTERPOLATE"
}

// InterpolateExpr represents an expression used by WITH FILL/INTERPOLATE
type InterpolateExpr struct {
	Column Ident
	Expr   Expr
}

func (i *InterpolateExpr) String() string {
	if i.Expr != nil {
		return fmt.Sprintf("%s AS %s", i.Column.String(), i.Expr.String())
	}
	return i.Column.String()
}

// LimitClause represents the different syntactic forms of LIMIT clauses
type LimitClause interface {
	fmt.Stringer
}

// LimitOffset represents standard SQL LIMIT syntax
type LimitOffset struct {
	Limit   Expr
	Offset  *Offset
	LimitBy []Expr
}

func (l *LimitOffset) String() string {
	var parts []string
	if l.Limit != nil {
		parts = append(parts, fmt.Sprintf("LIMIT %s", l.Limit.String()))
	}
	if l.Offset != nil {
		parts = append(parts, l.Offset.String())
	}
	if len(l.LimitBy) > 0 {
		limitBy := make([]string, len(l.LimitBy))
		for i, e := range l.LimitBy {
			limitBy[i] = e.String()
		}
		parts = append(parts, fmt.Sprintf("BY %s", strings.Join(limitBy, ", ")))
	}
	return strings.Join(parts, " ")
}

// OffsetCommaLimit represents MySQL LIMIT offset,limit syntax
type OffsetCommaLimit struct {
	Offset Expr
	Limit  Expr
}

func (o *OffsetCommaLimit) String() string {
	return fmt.Sprintf("LIMIT %s, %s", o.Offset.String(), o.Limit.String())
}

// Offset represents an OFFSET clause
type Offset struct {
	span  token.Span
	Value Expr
	Rows  OffsetRows
}

func (o *Offset) String() string {
	return fmt.Sprintf("OFFSET %s%s", o.Value.String(), o.Rows.String())
}

// OffsetRows represents the keyword after OFFSET <number>
type OffsetRows int

const (
	OffsetRowsNone OffsetRows = iota
	OffsetRowsRow
	OffsetRowsRows
)

func (o OffsetRows) String() string {
	switch o {
	case OffsetRowsRow:
		return " ROW"
	case OffsetRowsRows:
		return " ROWS"
	default:
		return ""
	}
}

// Fetch represents FETCH clause
type Fetch struct {
	span              token.Span
	WithTies          bool
	Percent           bool
	Quantity          Expr
	HasFirst          bool // Whether FIRST was explicitly specified
	HasNext           bool // Whether NEXT was explicitly specified
	HasRow            bool // Whether ROW (singular) was explicitly specified
	HasRows           bool // Whether ROWS (plural) was explicitly specified
	HasOnlyOrWithTies bool // Whether ONLY or WITH TIES was explicitly specified
}

func (f *Fetch) String() string {
	// Build the FETCH clause using canonical form (matching Rust implementation)
	// Canonical form: FETCH FIRST [quantity] [PERCENT] ROWS {ONLY | WITH TIES}
	// Note: FIRST is always used regardless of whether input used FIRST or NEXT
	// Note: ROWS is always used regardless of whether input used ROW or ROWS
	var parts []string
	parts = append(parts, "FETCH")

	// Always use FIRST in canonical form (not NEXT)
	if f.HasFirst || f.HasNext || f.Quantity != nil {
		parts = append(parts, "FIRST")
	}

	// Add quantity if present
	if f.Quantity != nil {
		quantityStr := f.Quantity.String()
		if f.Percent {
			quantityStr += " PERCENT"
		}
		parts = append(parts, quantityStr)
	}

	// Always use ROWS in canonical form (not ROW)
	// ROWS should be present when there's a quantity or when ONLY/WITH TIES is present
	if f.Quantity != nil || f.HasOnlyOrWithTies {
		parts = append(parts, "ROWS")
	}

	// Add ONLY or WITH TIES
	if f.WithTies {
		parts = append(parts, "WITH TIES")
	} else {
		parts = append(parts, "ONLY")
	}

	return strings.Join(parts, " ")
}

// LockClause represents FOR ... locking clause
type LockClause struct {
	span     token.Span
	LockType LockType
	Of       *ObjectName
	Nonblock *NonBlock
}

func (l *LockClause) String() string {
	parts := []string{"FOR", l.LockType.String()}
	if l.Of != nil {
		parts = append(parts, "OF", l.Of.String())
	}
	if l.Nonblock != nil {
		parts = append(parts, l.Nonblock.String())
	}
	return strings.Join(parts, " ")
}

// LockType represents the kind of lock (SHARE or UPDATE)
type LockType int

const (
	LockTypeShare LockType = iota
	LockTypeUpdate
)

func (l LockType) String() string {
	switch l {
	case LockTypeShare:
		return "SHARE"
	case LockTypeUpdate:
		return "UPDATE"
	default:
		return ""
	}
}

// NonBlock represents non-blocking lock options
type NonBlock int

const (
	NonBlockNowait NonBlock = iota
	NonBlockSkipLocked
)

func (n NonBlock) String() string {
	switch n {
	case NonBlockNowait:
		return "NOWAIT"
	case NonBlockSkipLocked:
		return "SKIP LOCKED"
	default:
		return ""
	}
}

// Setting represents a ClickHouse setting key-value pair
type Setting struct {
	span  token.Span
	Key   Ident
	Value Expr
}

func (s *Setting) String() string {
	return fmt.Sprintf("%s = %s", s.Key.String(), s.Value.String())
}

// FormatClause represents FORMAT clause (ClickHouse)
type FormatClause struct {
	Ident *Ident
	Null  bool
}

func (f *FormatClause) String() string {
	if f.Null {
		return "FORMAT NULL"
	}
	if f.Ident != nil {
		return fmt.Sprintf("FORMAT %s", f.Ident.String())
	}
	return "FORMAT"
}

// InputFormatClause represents FORMAT clause in input context (ClickHouse)
type InputFormatClause struct {
	span   token.Span
	Ident  Ident
	Values []Expr
}

func (i *InputFormatClause) String() string {
	parts := []string{fmt.Sprintf("FORMAT %s", i.Ident.String())}
	if len(i.Values) > 0 {
		values := make([]string, len(i.Values))
		for j, v := range i.Values {
			values[j] = v.String()
		}
		parts = append(parts, strings.Join(values, ", "))
	}
	return strings.Join(parts, " ")
}

// ForClause represents FOR XML or FOR JSON clause (MSSQL)
type ForClause struct {
	Type ForClauseType
}

func (f *ForClause) String() string {
	if f.Type != nil {
		return f.Type.String()
	}
	return ""
}

type ForClauseType interface {
	fmt.Stringer
}

// ForBrowseClause represents FOR BROWSE
type ForBrowseClause struct{}

func (f *ForBrowseClause) String() string { return "FOR BROWSE" }

// ForJsonClause represents FOR JSON clause
type ForJsonClause struct {
	ForJson             ForJson
	Root                *string
	IncludeNullValues   bool
	WithoutArrayWrapper bool
}

func (f *ForJsonClause) String() string {
	var parts []string
	parts = append(parts, "FOR JSON", f.ForJson.String())
	if f.Root != nil {
		parts = append(parts, fmt.Sprintf("ROOT('%s')", *f.Root))
	}
	if f.IncludeNullValues {
		parts = append(parts, "INCLUDE_NULL_VALUES")
	}
	if f.WithoutArrayWrapper {
		parts = append(parts, "WITHOUT_ARRAY_WRAPPER")
	}
	return strings.Join(parts, " ")
}

// ForXmlClause represents FOR XML clause
type ForXmlClause struct {
	ForXml       ForXml
	ElementName  *string // Optional element name for RAW/PATH modes
	Elements     bool
	BinaryBase64 bool
	Root         *string
	Type         bool
}

func (f *ForXmlClause) String() string {
	var parts []string
	parts = append(parts, "FOR XML "+f.ForXml.String())
	// Add element name for RAW/PATH modes if present
	if f.ElementName != nil {
		parts = append(parts, fmt.Sprintf("('%s')", *f.ElementName))
	}
	if f.BinaryBase64 {
		parts = append(parts, ", BINARY BASE64")
	}
	if f.Type {
		parts = append(parts, ", TYPE")
	}
	if f.Root != nil {
		parts = append(parts, fmt.Sprintf(", ROOT('%s')", *f.Root))
	}
	if f.Elements {
		parts = append(parts, ", ELEMENTS")
	}
	return strings.Join(parts, "")
}

// ForJson represents FOR JSON modes
type ForJson int

const (
	ForJsonAuto ForJson = iota
	ForJsonPath
)

func (f ForJson) String() string {
	switch f {
	case ForJsonAuto:
		return "AUTO"
	case ForJsonPath:
		return "PATH"
	default:
		return ""
	}
}

// ForXml represents FOR XML modes
type ForXml int

const (
	ForXmlRaw ForXml = iota
	ForXmlAuto
	ForXmlExplicit
	ForXmlPath
)

func (f ForXml) String() string {
	switch f {
	case ForXmlRaw:
		return "RAW"
	case ForXmlAuto:
		return "AUTO"
	case ForXmlExplicit:
		return "EXPLICIT"
	case ForXmlPath:
		return "PATH"
	default:
		return ""
	}
}

// LateralView represents a Hive LATERAL VIEW
type LateralView struct {
	span            token.Span
	LateralView     Expr
	LateralViewName ObjectName
	LateralColAlias []Ident
	Outer           bool
}

func (l *LateralView) String() string {
	parts := []string{"LATERAL VIEW"}
	if l.Outer {
		parts = append(parts, "OUTER")
	}
	parts = append(parts, l.LateralView.String(), l.LateralViewName.String())
	if len(l.LateralColAlias) > 0 {
		aliases := make([]string, len(l.LateralColAlias))
		for i, a := range l.LateralColAlias {
			aliases[i] = a.String()
		}
		parts = append(parts, "AS", strings.Join(aliases, ", "))
	}
	return strings.Join(parts, " ")
}

// ConnectByKind represents START WITH / CONNECT BY clause
type ConnectByKind interface {
	fmt.Stringer
}

// ConnectBy represents CONNECT BY clause
type ConnectBy struct {
	span          token.Span
	Nocycle       bool
	Relationships []Expr
}

func (c *ConnectBy) String() string {
	parts := []string{"CONNECT BY"}
	if c.Nocycle {
		parts = append(parts, "NOCYCLE")
	}
	rels := make([]string, len(c.Relationships))
	for i, r := range c.Relationships {
		rels[i] = r.String()
	}
	parts = append(parts, strings.Join(rels, ", "))
	return strings.Join(parts, " ")
}

// StartWith represents START WITH clause
type StartWith struct {
	span      token.Span
	Condition Expr
}

func (s *StartWith) String() string {
	return fmt.Sprintf("START WITH %s", s.Condition.String())
}

// GroupByExpr represents the GROUP BY clause
type GroupByExpr interface {
	fmt.Stringer
}

// GroupByAll represents GROUP BY ALL
type GroupByAll struct {
	Modifiers []GroupByWithModifier
}

func (g *GroupByAll) String() string {
	parts := []string{"GROUP BY ALL"}
	for _, m := range g.Modifiers {
		parts = append(parts, m.String())
	}
	return strings.Join(parts, " ")
}

// GroupByExpressions represents GROUP BY with expressions
type GroupByExpressions struct {
	Expressions []Expr
	Modifiers   []GroupByWithModifier
}

func (g *GroupByExpressions) String() string {
	parts := []string{"GROUP BY"}
	exprs := make([]string, len(g.Expressions))
	for i, e := range g.Expressions {
		exprs[i] = e.String()
	}
	parts = append(parts, strings.Join(exprs, ", "))
	for _, m := range g.Modifiers {
		parts = append(parts, m.String())
	}
	return strings.Join(parts, " ")
}

// GroupByWithModifier is an interface for GROUP BY modifiers (ROLLUP, CUBE, TOTALS, GROUPING SETS)
type GroupByWithModifier interface {
	fmt.Stringer
	groupByModifierNode()
}

// SimpleGroupByModifier represents simple modifiers like WITH ROLLUP, WITH CUBE, WITH TOTALS
type SimpleGroupByModifier int

const (
	SimpleGroupByModifierRollup SimpleGroupByModifier = iota
	SimpleGroupByModifierCube
	SimpleGroupByModifierTotals
)

func (s SimpleGroupByModifier) groupByModifierNode() {}

func (s SimpleGroupByModifier) String() string {
	switch s {
	case SimpleGroupByModifierRollup:
		return "WITH ROLLUP"
	case SimpleGroupByModifierCube:
		return "WITH CUBE"
	case SimpleGroupByModifierTotals:
		return "WITH TOTALS"
	default:
		return ""
	}
}

// GroupingSetsModifier represents GROUPING SETS modifier with its expression
type GroupingSetsModifier struct {
	Expr Expr
}

func (g *GroupingSetsModifier) groupByModifierNode() {}

func (g *GroupingSetsModifier) String() string {
	return g.Expr.String()
}

// GroupByWithModifierWithExpr represents modifiers that include expressions
// Deprecated: Use GroupingSetsModifier for GROUPING SETS, SimpleGroupByModifier for others
type GroupByWithModifierWithExpr struct {
	Modifier int // 0=ROLLUP, 1=CUBE, 2=TOTALS, 3=GROUPING SETS
	Expr     Expr
}

func (g *GroupByWithModifierWithExpr) String() string {
	switch g.Modifier {
	case 0:
		return "WITH ROLLUP"
	case 1:
		return "WITH CUBE"
	case 2:
		return "WITH TOTALS"
	case 3:
		return g.Expr.String()
	default:
		return ""
	}
}
