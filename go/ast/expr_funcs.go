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

// This file consolidates function expression types from ast/expr/functions.go
// into the main ast package.
//
// Key changes:
// - Expression types use "E" prefix
// - Uses existing NullTreatment, OrderByExpr types where possible
// - Moved from expr/ subpackage into ast package
//
// TODO: After full migration, remove the old ast/expr/ directory

package ast

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// ============================================================================
// Function Arguments (from ast/expr/functions.go)
// ============================================================================

// EFunctionArg represents an argument to a function.
type EFunctionArg interface {
	String() string
}

// EFunctionArgExpr is an expression used as a function argument.
type EFunctionArgExpr struct {
	Expr Expr
}

// String returns the SQL representation.
func (f *EFunctionArgExpr) String() string { return f.Expr.String() }

// EFunctionArgNamed is a named function argument (e.g., `name => value`).
type EFunctionArgNamed struct {
	Name  *Ident
	Value Expr
}

// String returns the SQL representation.
func (f *EFunctionArgNamed) String() string {
	return fmt.Sprintf("%s => %s", f.Name.String(), f.Value.String())
}

// EDuplicateTreatment represents ALL or DISTINCT in function arguments.
type EDuplicateTreatment int

const (
	EDuplicateNone EDuplicateTreatment = iota
	EDuplicateAll
	EDuplicateDistinct
)

// String returns the SQL representation.
func (d EDuplicateTreatment) String() string {
	switch d {
	case EDuplicateAll:
		return "ALL"
	case EDuplicateDistinct:
		return "DISTINCT"
	}
	return ""
}

// Note: NullTreatment is already defined in ast/expr.go - use that instead

// EListAggOnOverflow represents the ON OVERFLOW clause for LISTAGG.
type EListAggOnOverflow int

const (
	EListAggError EListAggOnOverflow = iota
	EListAggTruncate
)

// String returns the SQL representation.
func (l EListAggOnOverflow) String() string {
	switch l {
	case EListAggError:
		return "ON OVERFLOW ERROR"
	case EListAggTruncate:
		return "ON OVERFLOW TRUNCATE"
	}
	return ""
}

// EHavingBound represents a HAVING bound for ANY_VALUE.
type EHavingBound struct {
	IsMax bool
	Expr  Expr
}

// String returns the SQL representation.
func (h *EHavingBound) String() string {
	if h.IsMax {
		return fmt.Sprintf("HAVING MAX %s", h.Expr.String())
	}
	return fmt.Sprintf("HAVING MIN %s", h.Expr.String())
}

// EJsonNullClause represents JSON null handling clause.
type EJsonNullClause int

const (
	EJsonNullAbsent EJsonNullClause = iota
	EJsonNullNull
)

// String returns the SQL representation.
func (j EJsonNullClause) String() string {
	switch j {
	case EJsonNullAbsent:
		return "NULL ON NULL"
	case EJsonNullNull:
		return "ABSENT ON NULL"
	}
	return ""
}

// EJsonReturningClause represents a RETURNING clause for JSON functions.
type EJsonReturningClause struct {
	DataType string
}

// String returns the SQL representation.
func (j *EJsonReturningClause) String() string {
	return fmt.Sprintf("RETURNING %s", j.DataType)
}

// EFunctionArgumentClause represents clauses inside function argument lists.
type EFunctionArgumentClause interface {
	String() string
}

// EIgnoreOrRespectNullsClause represents { IGNORE | RESPECT } NULLS.
type EIgnoreOrRespectNullsClause struct {
	Treatment NullTreatment
}

// String returns the SQL representation.
func (i *EIgnoreOrRespectNullsClause) String() string { return i.Treatment.String() }

// EOrderByClause represents ORDER BY inside function arguments.
type EOrderByClause struct {
	OrderBy []Expr // OrderByExpr
}

// String returns the SQL representation.
func (o *EOrderByClause) String() string {
	items := make([]string, len(o.OrderBy))
	for i, item := range o.OrderBy {
		items[i] = item.String()
	}
	return "ORDER BY " + strings.Join(items, ", ")
}

// ELimitClauseFunc represents LIMIT inside function arguments.
type ELimitClauseFunc struct {
	Limit Expr
}

// String returns the SQL representation.
func (l *ELimitClauseFunc) String() string { return fmt.Sprintf("LIMIT %s", l.Limit.String()) }

// EOnOverflowClause represents ON OVERFLOW inside function arguments.
type EOnOverflowClause struct {
	OnOverflow EListAggOnOverflow
}

// String returns the SQL representation.
func (o *EOnOverflowClause) String() string { return o.OnOverflow.String() }

// EHavingClause represents HAVING inside function arguments.
type EHavingClause struct {
	Bound *EHavingBound
}

// String returns the SQL representation.
func (h *EHavingClause) String() string { return h.Bound.String() }

// ESeparatorClause represents SEPARATOR inside function arguments.
type ESeparatorClause struct {
	Value interface{} // ValueWithSpan
}

// String returns the SQL representation.
func (s *ESeparatorClause) String() string { return fmt.Sprintf("SEPARATOR %v", s.Value) }

// EJsonNullOnNullClause represents JSON null clause.
type EJsonNullOnNullClause struct {
	Clause EJsonNullClause
}

// String returns the SQL representation.
func (j *EJsonNullOnNullClause) String() string { return j.Clause.String() }

// EJsonReturning represents JSON RETURNING clause.
type EJsonReturning struct {
	Clause *EJsonReturningClause
}

// String returns the SQL representation.
func (j *EJsonReturning) String() string { return j.Clause.String() }

// EFunctionArgumentList represents the contents inside the parentheses of a function call.
type EFunctionArgumentList struct {
	DuplicateTreatment EDuplicateTreatment
	Args               []EFunctionArg
	Clauses            []EFunctionArgumentClause
}

// String returns the SQL representation.
func (f *EFunctionArgumentList) String() string {
	var sb strings.Builder

	if f.DuplicateTreatment != EDuplicateNone {
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

// EFunctionArguments represents the arguments passed to a function call.
type EFunctionArguments struct {
	None     bool
	Subquery *EQueryExpr
	List     *EFunctionArgumentList
}

// String returns the SQL representation.
func (f *EFunctionArguments) String() string {
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

// ============================================================================
// Window Functions (from ast/expr/functions.go)
// ============================================================================

// EWindowFrameUnits represents the units for window frames.
type EWindowFrameUnits int

const (
	EWindowRows EWindowFrameUnits = iota
	EWindowRange
	EWindowGroups
)

// String returns the SQL representation.
func (w EWindowFrameUnits) String() string {
	switch w {
	case EWindowRows:
		return "ROWS"
	case EWindowRange:
		return "RANGE"
	case EWindowGroups:
		return "GROUPS"
	}
	return ""
}

// EWindowFrameBoundType represents the type of a window frame bound.
type EWindowFrameBoundType int

const (
	EBoundTypeUnboundedPreceding EWindowFrameBoundType = iota
	EBoundTypePreceding
	EBoundTypeCurrentRow
	EBoundTypeFollowing
	EBoundTypeUnboundedFollowing
)

// EWindowFrameBound represents a bound in a window frame.
type EWindowFrameBound struct {
	BoundType EWindowFrameBoundType
	Expr      *Expr
}

// String returns the SQL representation.
func (w *EWindowFrameBound) String() string {
	switch w.BoundType {
	case EBoundTypeCurrentRow:
		return "CURRENT ROW"
	case EBoundTypeUnboundedPreceding:
		return "UNBOUNDED PRECEDING"
	case EBoundTypeUnboundedFollowing:
		return "UNBOUNDED FOLLOWING"
	case EBoundTypePreceding:
		if w.Expr != nil {
			return fmt.Sprintf("%s PRECEDING", (*w.Expr).String())
		}
		return "UNBOUNDED PRECEDING"
	case EBoundTypeFollowing:
		if w.Expr != nil {
			return fmt.Sprintf("%s FOLLOWING", (*w.Expr).String())
		}
		return "UNBOUNDED FOLLOWING"
	}
	return ""
}

// EWindowFrame represents a window frame specification.
type EWindowFrame struct {
	Units      EWindowFrameUnits
	StartBound *EWindowFrameBound
	EndBound   *EWindowFrameBound
}

// String returns the SQL representation.
func (w *EWindowFrame) String() string {
	if w.EndBound != nil {
		return fmt.Sprintf("%s BETWEEN %s AND %s",
			w.Units.String(), w.StartBound.String(), w.EndBound.String())
	}
	return fmt.Sprintf("%s %s", w.Units.String(), w.StartBound.String())
}

// EWindowSpec represents a window specification.
type EWindowSpec struct {
	WindowName  *Ident
	PartitionBy []Expr
	OrderBy     []Expr // OrderByExpr
	WindowFrame *EWindowFrame
}

// String returns the SQL representation.
func (w *EWindowSpec) String() string {
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

// EWindowType represents the type of window specification.
type EWindowType struct {
	Spec  *EWindowSpec
	Named *Ident
}

// String returns the SQL representation.
func (w *EWindowType) String() string {
	if w.Named != nil {
		return w.Named.String()
	}
	if w.Spec != nil {
		return w.Spec.String()
	}
	return ""
}

// EFunctionExpr represents a function call expression (was expr.FunctionExpr).
// This is different from ast.Function - it has more detailed options.
type EFunctionExpr struct {
	ExpressionBase
	Name           *ObjectName
	UsesOdbcSyntax bool
	Parameters     *EFunctionArguments
	Args           *EFunctionArguments
	Filter         Expr
	NullTreatment  NullTreatment
	Over           *EWindowType
	WithinGroup    []Expr // OrderByExpr
	SpanVal        token.Span
}

// Span returns the source span for this expression.
func (f *EFunctionExpr) Span() token.Span { return f.SpanVal }

// String returns the SQL representation.
func (f *EFunctionExpr) String() string {
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

	if f.NullTreatment != NullTreatmentUnspecified {
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
