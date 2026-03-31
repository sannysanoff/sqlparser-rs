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

	"github.com/user/sqlparser/span"
)

// FunctionArg represents an argument to a function.
type FunctionArg interface {
	String() string
}

// FunctionArgExpr is an expression used as a function argument.
type FunctionArgExpr struct {
	Expr Expr
}

// String returns the SQL representation.
func (f *FunctionArgExpr) String() string {
	return f.Expr.String()
}

// FunctionArgNamed is a named function argument (e.g., `name => value`).
type FunctionArgNamed struct {
	Name  *Ident
	Value Expr
}

// String returns the SQL representation.
func (f *FunctionArgNamed) String() string {
	return fmt.Sprintf("%s => %s", f.Name.String(), f.Value.String())
}

// DuplicateTreatment represents ALL or DISTINCT in function arguments.
type DuplicateTreatment int

const (
	DuplicateNone DuplicateTreatment = iota
	DuplicateAll
	DuplicateDistinct
)

// String returns the SQL representation.
func (d DuplicateTreatment) String() string {
	switch d {
	case DuplicateAll:
		return "ALL"
	case DuplicateDistinct:
		return "DISTINCT"
	}
	return ""
}

// NullTreatment represents IGNORE NULLS or RESPECT NULLS.
type NullTreatment int

const (
	NullTreatmentNone NullTreatment = iota
	NullTreatmentIgnore
	NullTreatmentRespect
)

// String returns the SQL representation.
func (n NullTreatment) String() string {
	switch n {
	case NullTreatmentIgnore:
		return "IGNORE NULLS"
	case NullTreatmentRespect:
		return "RESPECT NULLS"
	}
	return ""
}

// ListAggOnOverflow represents the ON OVERFLOW clause for LISTAGG.
type ListAggOnOverflow int

const (
	ListAggError ListAggOnOverflow = iota
	ListAggTruncate
)

// String returns the SQL representation.
func (l ListAggOnOverflow) String() string {
	switch l {
	case ListAggError:
		return "ON OVERFLOW ERROR"
	case ListAggTruncate:
		return "ON OVERFLOW TRUNCATE"
	}
	return ""
}

// HavingBound represents a HAVING bound for ANY_VALUE.
type HavingBound struct {
	IsMax bool
	Expr  Expr
}

// String returns the SQL representation.
func (h *HavingBound) String() string {
	if h.IsMax {
		return fmt.Sprintf("HAVING MAX %s", h.Expr.String())
	}
	return fmt.Sprintf("HAVING MIN %s", h.Expr.String())
}

// JsonNullClause represents JSON null handling clause.
type JsonNullClause int

const (
	JsonNullAbsent JsonNullClause = iota
	JsonNullNull
)

// String returns the SQL representation.
func (j JsonNullClause) String() string {
	switch j {
	case JsonNullAbsent:
		return "NULL ON NULL"
	case JsonNullNull:
		return "ABSENT ON NULL"
	}
	return ""
}

// JsonReturningClause represents a RETURNING clause for JSON functions.
type JsonReturningClause struct {
	DataType string
}

// String returns the SQL representation.
func (j *JsonReturningClause) String() string {
	return fmt.Sprintf("RETURNING %s", j.DataType)
}

// FunctionArgumentClause represents clauses inside function argument lists.
type FunctionArgumentClause interface {
	String() string
}

// IgnoreOrRespectNullsClause represents { IGNORE | RESPECT } NULLS.
type IgnoreOrRespectNullsClause struct {
	Treatment NullTreatment
}

// String returns the SQL representation.
func (i *IgnoreOrRespectNullsClause) String() string {
	return i.Treatment.String()
}

// OrderByClause represents ORDER BY inside function arguments.
type OrderByClause struct {
	OrderBy []Expr // OrderByExpr
}

// String returns the SQL representation.
func (o *OrderByClause) String() string {
	items := make([]string, len(o.OrderBy))
	for i, item := range o.OrderBy {
		items[i] = item.String()
	}
	return "ORDER BY " + strings.Join(items, ", ")
}

// LimitClause represents LIMIT inside function arguments.
type LimitClause struct {
	Limit Expr
}

// String returns the SQL representation.
func (l *LimitClause) String() string {
	return fmt.Sprintf("LIMIT %s", l.Limit.String())
}

// OnOverflowClause represents ON OVERFLOW inside function arguments.
type OnOverflowClause struct {
	OnOverflow ListAggOnOverflow
}

// String returns the SQL representation.
func (o *OnOverflowClause) String() string {
	return o.OnOverflow.String()
}

// HavingClause represents HAVING inside function arguments.
type HavingClause struct {
	Bound *HavingBound
}

// String returns the SQL representation.
func (h *HavingClause) String() string {
	return h.Bound.String()
}

// SeparatorClause represents SEPARATOR inside function arguments.
type SeparatorClause struct {
	Value interface{} // ValueWithSpan
}

// String returns the SQL representation.
func (s *SeparatorClause) String() string {
	return fmt.Sprintf("SEPARATOR %v", s.Value)
}

// JsonNullOnNullClause represents JSON null clause.
type JsonNullOnNullClause struct {
	Clause JsonNullClause
}

// String returns the SQL representation.
func (j *JsonNullOnNullClause) String() string {
	return j.Clause.String()
}

// JsonReturning represents JSON RETURNING clause.
type JsonReturning struct {
	Clause *JsonReturningClause
}

// String returns the SQL representation.
func (j *JsonReturning) String() string {
	return j.Clause.String()
}

// FunctionArgumentList represents the contents inside the parentheses of a function call.
type FunctionArgumentList struct {
	DuplicateTreatment DuplicateTreatment
	Args               []FunctionArg
	Clauses            []FunctionArgumentClause
}

// String returns the SQL representation.
func (f *FunctionArgumentList) String() string {
	var sb strings.Builder

	if f.DuplicateTreatment != DuplicateNone {
		sb.WriteString(f.DuplicateTreatment.String())
		sb.WriteString(" ")
	}

	args := make([]string, len(f.Args))
	for i, arg := range f.Args {
		args[i] = arg.String()
	}
	sb.WriteString(strings.Join(args, ", "))

	if len(f.Clauses) > 0 {
		if len(f.Args) > 0 {
			sb.WriteString(" ")
		}
		clauses := make([]string, len(f.Clauses))
		for i, clause := range f.Clauses {
			clauses[i] = clause.String()
		}
		sb.WriteString(strings.Join(clauses, " "))
	}

	return sb.String()
}

// FunctionArguments represents the arguments passed to a function call.
type FunctionArguments struct {
	// One of:
	None     bool
	Subquery *QueryExpr
	List     *FunctionArgumentList
}

// String returns the SQL representation.
func (f *FunctionArguments) String() string {
	if f == nil {
		return ""
	}
	if f.None {
		return "()"
	}
	if f.Subquery != nil {
		return fmt.Sprintf("(%s)", f.Subquery.String())
	}
	if f.List != nil {
		return fmt.Sprintf("(%s)", f.List.String())
	}
	return ""
}

// WindowFrameUnits represents the units for window frames.
type WindowFrameUnits int

const (
	WindowRows WindowFrameUnits = iota
	WindowRange
	WindowGroups
)

// String returns the SQL representation.
func (w WindowFrameUnits) String() string {
	switch w {
	case WindowRows:
		return "ROWS"
	case WindowRange:
		return "RANGE"
	case WindowGroups:
		return "GROUPS"
	}
	return ""
}

// WindowFrameBoundType represents the type of a window frame bound.
type WindowFrameBoundType int

const (
	// BoundTypeUnboundedPreceding represents UNBOUNDED PRECEDING
	BoundTypeUnboundedPreceding WindowFrameBoundType = iota
	// BoundTypePreceding represents <expr> PRECEDING
	BoundTypePreceding
	// BoundTypeCurrentRow represents CURRENT ROW
	BoundTypeCurrentRow
	// BoundTypeFollowing represents <expr> FOLLOWING
	BoundTypeFollowing
	// BoundTypeUnboundedFollowing represents UNBOUNDED FOLLOWING
	BoundTypeUnboundedFollowing
)

// WindowFrameBound represents a bound in a window frame.
type WindowFrameBound struct {
	BoundType WindowFrameBoundType
	// Expr is set for BoundTypePreceding and BoundTypeFollowing when not UNBOUNDED
	Expr *Expr
}

// String returns the SQL representation.
func (w *WindowFrameBound) String() string {
	switch w.BoundType {
	case BoundTypeCurrentRow:
		return "CURRENT ROW"
	case BoundTypeUnboundedPreceding:
		return "UNBOUNDED PRECEDING"
	case BoundTypeUnboundedFollowing:
		return "UNBOUNDED FOLLOWING"
	case BoundTypePreceding:
		if w.Expr != nil {
			return fmt.Sprintf("%s PRECEDING", (*w.Expr).String())
		}
		return "UNBOUNDED PRECEDING"
	case BoundTypeFollowing:
		if w.Expr != nil {
			return fmt.Sprintf("%s FOLLOWING", (*w.Expr).String())
		}
		return "UNBOUNDED FOLLOWING"
	}
	return ""
}

// WindowFrame represents a window frame specification.
type WindowFrame struct {
	Units      WindowFrameUnits
	StartBound *WindowFrameBound
	EndBound   *WindowFrameBound
}

// String returns the SQL representation.
func (w *WindowFrame) String() string {
	if w.EndBound != nil {
		return fmt.Sprintf("%s BETWEEN %s AND %s",
			w.Units.String(), w.StartBound.String(), w.EndBound.String())
	}
	return fmt.Sprintf("%s %s", w.Units.String(), w.StartBound.String())
}

// OrderByExpr represents an ORDER BY expression.
type OrderByExpr struct {
	Expr       Expr
	Asc        *bool // nil means default (ASC)
	NullsFirst *bool
}

func (o *OrderByExpr) exprNode() {}

// Span returns the source span.
func (o *OrderByExpr) Span() span.Span { return span.Span{} }

// String returns the SQL representation.
func (o *OrderByExpr) String() string {
	var sb strings.Builder
	sb.WriteString(o.Expr.String())

	if o.Asc != nil {
		if *o.Asc {
			sb.WriteString(" ASC")
		} else {
			sb.WriteString(" DESC")
		}
	}

	if o.NullsFirst != nil {
		if *o.NullsFirst {
			sb.WriteString(" NULLS FIRST")
		} else {
			sb.WriteString(" NULLS LAST")
		}
	}

	return sb.String()
}

// WindowSpec represents a window specification.
type WindowSpec struct {
	WindowName  *Ident
	PartitionBy []Expr
	OrderBy     []Expr // OrderByExpr
	WindowFrame *WindowFrame
}

// String returns the SQL representation.
func (w *WindowSpec) String() string {
	var parts []string

	if w.WindowName != nil {
		parts = append(parts, w.WindowName.String())
	}

	if len(w.PartitionBy) > 0 {
		items := make([]string, len(w.PartitionBy))
		for i, item := range w.PartitionBy {
			items[i] = item.String()
		}
		parts = append(parts, fmt.Sprintf("PARTITION BY %s", strings.Join(items, ", ")))
	}

	if len(w.OrderBy) > 0 {
		items := make([]string, len(w.OrderBy))
		for i, item := range w.OrderBy {
			items[i] = item.String()
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(items, ", ")))
	}

	if w.WindowFrame != nil {
		parts = append(parts, w.WindowFrame.String())
	}

	return "(" + strings.Join(parts, " ") + ")"
}

// WindowType represents the type of window specification.
type WindowType struct {
	// One of:
	Spec  *WindowSpec
	Named *Ident
}

// String returns the SQL representation.
func (w *WindowType) String() string {
	if w.Named != nil {
		return w.Named.String()
	}
	if w.Spec != nil {
		return w.Spec.String()
	}
	return ""
}

// FunctionExpr represents a function call expression.
type FunctionExpr struct {
	Name           *ObjectName
	UsesOdbcSyntax bool
	Parameters     *FunctionArguments
	Args           *FunctionArguments
	Filter         Expr
	NullTreatment  NullTreatment
	Over           *WindowType
	WithinGroup    []Expr // OrderByExpr
	SpanVal        span.Span
}

func (f *FunctionExpr) exprNode() {}

// Span returns the source span for this expression.
func (f *FunctionExpr) Span() span.Span {
	return f.SpanVal
}

// String returns the SQL representation.
func (f *FunctionExpr) String() string {
	var sb strings.Builder

	if f.UsesOdbcSyntax {
		sb.WriteString("{fn ")
	}

	sb.WriteString(f.Name.String())
	sb.WriteString(f.Parameters.String())
	sb.WriteString(f.Args.String())

	if len(f.WithinGroup) > 0 {
		items := make([]string, len(f.WithinGroup))
		for i, item := range f.WithinGroup {
			items[i] = item.String()
		}
		sb.WriteString(fmt.Sprintf(" WITHIN GROUP (ORDER BY %s)", strings.Join(items, ", ")))
	}

	if f.Filter != nil {
		sb.WriteString(fmt.Sprintf(" FILTER (WHERE %s)", f.Filter.String()))
	}

	if f.NullTreatment != NullTreatmentNone {
		sb.WriteString(" ")
		sb.WriteString(f.NullTreatment.String())
	}

	if f.Over != nil {
		sb.WriteString(" OVER ")
		sb.WriteString(f.Over.String())
	}

	if f.UsesOdbcSyntax {
		sb.WriteString("}")
	}

	return sb.String()
}
