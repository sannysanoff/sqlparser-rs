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

// TableWithJoins represents a left table followed by zero or more joins
type TableWithJoins struct {
	span     token.Span
	Relation TableFactor
	Joins    []Join
}

func (t *TableWithJoins) Span() token.Span { return t.span }
func (t *TableWithJoins) String() string {
	parts := []string{t.Relation.String()}
	for _, join := range t.Joins {
		parts = append(parts, join.String())
	}
	return strings.Join(parts, " ")
}

// TableFactor represents a table name or parenthesized subquery with optional alias
type TableFactor interface {
	fmt.Stringer
	Span() token.Span
}

// TableTableFactor represents a named table or relation
type TableTableFactor struct {
	span           token.Span
	Name           ObjectName
	Alias          *TableAlias
	Args           *TableFunctionArgs
	WithHints      []Expr
	Version        *TableVersionWithInfo
	WithOrdinality bool
	Partitions     []Ident
	JsonPath       *JsonPath
	Sample         *TableSampleKind
	IndexHints     []TableIndexHints
}

func (t *TableTableFactor) Span() token.Span { return t.span }
func (t *TableTableFactor) String() string {
	var parts []string
	parts = append(parts, t.Name.String())
	if t.JsonPath != nil {
		parts = append(parts, t.JsonPath.String())
	}
	if len(t.Partitions) > 0 {
		partitionStrs := make([]string, len(t.Partitions))
		for i, p := range t.Partitions {
			partitionStrs[i] = p.String()
		}
		parts = append(parts, fmt.Sprintf("PARTITION (%s)", strings.Join(partitionStrs, ", ")))
	}
	if t.Args != nil {
		args := make([]string, len(t.Args.Args))
		for i, arg := range t.Args.Args {
			args[i] = arg.String()
		}
		argStr := "(" + strings.Join(args, ", ") + ")"
		if t.Args.Settings != nil && len(*t.Args.Settings) > 0 {
			settings := make([]string, len(*t.Args.Settings))
			for i, s := range *t.Args.Settings {
				settings[i] = s.String()
			}
			argStr += ", SETTINGS " + strings.Join(settings, ", ")
		}
		parts = append(parts, argStr)
	}
	if t.WithOrdinality {
		parts = append(parts, "WITH ORDINALITY")
	}
	if t.Sample != nil && t.Sample.BeforeTableAlias != nil {
		parts = append(parts, t.Sample.BeforeTableAlias.String())
	}
	if t.Alias != nil {
		parts = append(parts, t.Alias.String())
	}
	if len(t.IndexHints) > 0 {
		hints := make([]string, len(t.IndexHints))
		for i, h := range t.IndexHints {
			hints[i] = h.String()
		}
		parts = append(parts, strings.Join(hints, " "))
	}
	if len(t.WithHints) > 0 {
		hints := make([]string, len(t.WithHints))
		for i, h := range t.WithHints {
			hints[i] = h.String()
		}
		parts = append(parts, fmt.Sprintf("WITH (%s)", strings.Join(hints, ", ")))
	}
	if t.Version != nil {
		parts = append(parts, t.Version.String())
	}
	if t.Sample != nil && t.Sample.AfterTableAlias != nil {
		parts = append(parts, t.Sample.AfterTableAlias.String())
	}
	return strings.Join(parts, " ")
}

// DerivedTableFactor represents a derived table (parenthesized subquery)
type DerivedTableFactor struct {
	span     token.Span
	Lateral  bool
	Subquery *Query
	Alias    *TableAlias
	Sample   *TableSampleKind
}

func (d *DerivedTableFactor) Span() token.Span { return d.span }
func (d *DerivedTableFactor) String() string {
	var parts []string
	if d.Lateral {
		parts = append(parts, "LATERAL")
	}
	parts = append(parts, fmt.Sprintf("(%s)", d.Subquery.String()))
	if d.Alias != nil {
		parts = append(parts, d.Alias.String())
	}
	if d.Sample != nil && d.Sample.AfterTableAlias != nil {
		parts = append(parts, d.Sample.AfterTableAlias.String())
	}
	return strings.Join(parts, " ")
}

// TableFunctionTableFactor represents TABLE(<expr>)[ AS <alias> ]
type TableFunctionTableFactor struct {
	span  token.Span
	Expr  Expr
	Alias *TableAlias
}

func (t *TableFunctionTableFactor) Span() token.Span { return t.span }
func (t *TableFunctionTableFactor) String() string {
	if t.Alias != nil {
		return fmt.Sprintf("TABLE(%s) %s", t.Expr.String(), t.Alias.String())
	}
	return fmt.Sprintf("TABLE(%s)", t.Expr.String())
}

// FunctionTableFactor represents LATERAL FLATTEN or similar table functions
type FunctionTableFactor struct {
	span    token.Span
	Lateral bool
	Name    ObjectName
	Args    []FunctionArg
	Alias   *TableAlias
}

func (f *FunctionTableFactor) Span() token.Span { return f.span }
func (f *FunctionTableFactor) String() string {
	var parts []string
	if f.Lateral {
		parts = append(parts, "LATERAL")
	}
	args := make([]string, len(f.Args))
	for i, arg := range f.Args {
		args[i] = arg.String()
	}
	parts = append(parts, fmt.Sprintf("%s(%s)", f.Name.String(), strings.Join(args, ", ")))
	if f.Alias != nil {
		parts = append(parts, f.Alias.String())
	}
	return strings.Join(parts, " ")
}

// FunctionArg represents an argument to a function
type FunctionArg struct {
	Expr Expr
}

func (f *FunctionArg) String() string {
	return f.Expr.String()
}

// UnnestTableFactor represents UNNEST table operator
type UnnestTableFactor struct {
	span            token.Span
	Alias           *TableAlias
	ArrayExprs      []Expr
	WithOffset      bool
	WithOffsetAlias *Ident
	WithOrdinality  bool
}

func (u *UnnestTableFactor) Span() token.Span { return u.span }
func (u *UnnestTableFactor) String() string {
	exprs := make([]string, len(u.ArrayExprs))
	for i, e := range u.ArrayExprs {
		exprs[i] = e.String()
	}
	var parts []string
	parts = append(parts, fmt.Sprintf("UNNEST(%s)", strings.Join(exprs, ", ")))
	if u.WithOrdinality {
		parts = append(parts, "WITH ORDINALITY")
	}
	if u.Alias != nil {
		parts = append(parts, u.Alias.String())
	}
	if u.WithOffset {
		parts = append(parts, "WITH OFFSET")
	}
	if u.WithOffsetAlias != nil {
		parts = append(parts, u.WithOffsetAlias.String())
	}
	return strings.Join(parts, " ")
}

// JsonTableTableFactor represents JSON_TABLE table-valued function
type JsonTableTableFactor struct {
	span     token.Span
	JsonExpr Expr
	JsonPath ValueWithSpan
	Columns  []JsonTableColumn
	Alias    *TableAlias
}

func (j *JsonTableTableFactor) Span() token.Span { return j.span }
func (j *JsonTableTableFactor) String() string {
	cols := make([]string, len(j.Columns))
	for i, c := range j.Columns {
		cols[i] = c.String()
	}
	result := fmt.Sprintf("JSON_TABLE(%s, %s COLUMNS(%s))",
		j.JsonExpr.String(), j.JsonPath.String(), strings.Join(cols, ", "))
	if j.Alias != nil {
		result += " " + j.Alias.String()
	}
	return result
}

// OpenJsonTableFactor represents MSSQL's OPENJSON table-valued function
type OpenJsonTableFactor struct {
	span     token.Span
	JsonExpr Expr
	JsonPath *ValueWithSpan
	Columns  []OpenJsonTableColumn
	Alias    *TableAlias
}

func (o *OpenJsonTableFactor) Span() token.Span { return o.span }
func (o *OpenJsonTableFactor) String() string {
	var parts []string
	if o.JsonPath != nil {
		parts = append(parts, fmt.Sprintf("OPENJSON(%s, %s)", o.JsonExpr.String(), o.JsonPath.String()))
	} else {
		parts = append(parts, fmt.Sprintf("OPENJSON(%s)", o.JsonExpr.String()))
	}
	if len(o.Columns) > 0 {
		cols := make([]string, len(o.Columns))
		for i, c := range o.Columns {
			cols[i] = c.String()
		}
		parts = append(parts, fmt.Sprintf("WITH (%s)", strings.Join(cols, ", ")))
	}
	if o.Alias != nil {
		parts = append(parts, o.Alias.String())
	}
	return strings.Join(parts, " ")
}

// NestedJoinTableFactor represents a parenthesized join expression
type NestedJoinTableFactor struct {
	span           token.Span
	TableWithJoins *TableWithJoins
	Alias          *TableAlias
}

func (n *NestedJoinTableFactor) Span() token.Span { return n.span }
func (n *NestedJoinTableFactor) String() string {
	result := fmt.Sprintf("(%s)", n.TableWithJoins.String())
	if n.Alias != nil {
		result += " " + n.Alias.String()
	}
	return result
}

// PivotTableFactor represents a PIVOT operation
type PivotTableFactor struct {
	span               token.Span
	Table              TableFactor
	AggregateFunctions []ExprWithAlias
	ValueColumn        []Expr
	ValueSource        PivotValueSource
	DefaultOnNull      Expr
	Alias              *TableAlias
}

func (p *PivotTableFactor) Span() token.Span { return p.span }
func (p *PivotTableFactor) String() string {
	aggs := make([]string, len(p.AggregateFunctions))
	for i, a := range p.AggregateFunctions {
		aggs[i] = a.String()
	}
	valCols := make([]string, len(p.ValueColumn))
	for i, v := range p.ValueColumn {
		valCols[i] = v.String()
	}
	var valueColStr string
	if len(valCols) == 1 {
		valueColStr = valCols[0]
	} else {
		valueColStr = "(" + strings.Join(valCols, ", ") + ")"
	}
	var defaultOnNullStr string
	if p.DefaultOnNull != nil {
		defaultOnNullStr = fmt.Sprintf(" DEFAULT ON NULL (%s)", p.DefaultOnNull.String())
	}
	result := fmt.Sprintf("%s PIVOT (%s FOR %s IN (%s)%s)",
		p.Table.String(), strings.Join(aggs, ", "), valueColStr, p.ValueSource.String(), defaultOnNullStr)
	if p.Alias != nil {
		result += " " + p.Alias.String()
	}
	return result
}

// UnpivotTableFactor represents an UNPIVOT operation
type UnpivotTableFactor struct {
	span          token.Span
	Table         TableFactor
	Value         Expr
	Name          Ident
	Columns       []ExprWithAlias
	NullInclusion *NullInclusion
	Alias         *TableAlias
}

func (u *UnpivotTableFactor) Span() token.Span { return u.span }
func (u *UnpivotTableFactor) String() string {
	cols := make([]string, len(u.Columns))
	for i, c := range u.Columns {
		cols[i] = c.String()
	}
	var parts []string
	parts = append(parts, u.Table.String())
	parts = append(parts, "UNPIVOT")
	if u.NullInclusion != nil {
		parts = append(parts, u.NullInclusion.String())
	}
	parts = append(parts, fmt.Sprintf("(%s FOR %s IN (%s))",
		u.Value.String(), u.Name.String(), strings.Join(cols, ", ")))
	if u.Alias != nil {
		parts = append(parts, u.Alias.String())
	}
	return strings.Join(parts, " ")
}

// NullInclusion represents whether to include or exclude NULLs during unpivot
type NullInclusion int

const (
	IncludeNulls NullInclusion = iota
	ExcludeNulls
)

func (n NullInclusion) String() string {
	if n == IncludeNulls {
		return "INCLUDE NULLS"
	}
	return "EXCLUDE NULLS"
}

// MatchRecognizeTableFactor represents MATCH_RECOGNIZE operation
type MatchRecognizeTableFactor struct {
	span           token.Span
	Table          TableFactor
	PartitionBy    []Expr
	OrderBy        []OrderByExpr
	Measures       []Measure
	RowsPerMatch   *RowsPerMatch
	AfterMatchSkip *AfterMatchSkipWithSymbol
	Pattern        MatchRecognizePattern
	Symbols        []SymbolDefinition
	Alias          *TableAlias
}

func (m *MatchRecognizeTableFactor) Span() token.Span { return m.span }
func (m *MatchRecognizeTableFactor) String() string {
	var parts []string
	parts = append(parts, m.Table.String())
	parts = append(parts, "MATCH_RECOGNIZE(")
	if len(m.PartitionBy) > 0 {
		partitionBy := make([]string, len(m.PartitionBy))
		for i, p := range m.PartitionBy {
			partitionBy[i] = p.String()
		}
		parts = append(parts, fmt.Sprintf("PARTITION BY %s", strings.Join(partitionBy, ", ")))
	}
	if len(m.OrderBy) > 0 {
		orderBy := make([]string, len(m.OrderBy))
		for i, o := range m.OrderBy {
			orderBy[i] = o.String()
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(orderBy, ", ")))
	}
	if len(m.Measures) > 0 {
		measures := make([]string, len(m.Measures))
		for i, m := range m.Measures {
			measures[i] = m.String()
		}
		parts = append(parts, fmt.Sprintf("MEASURES %s", strings.Join(measures, ", ")))
	}
	if m.RowsPerMatch != nil {
		parts = append(parts, m.RowsPerMatch.String())
	}
	if m.AfterMatchSkip != nil {
		parts = append(parts, m.AfterMatchSkip.String())
	}
	parts = append(parts, fmt.Sprintf("PATTERN (%s)", m.Pattern.String()))
	symbols := make([]string, len(m.Symbols))
	for i, s := range m.Symbols {
		symbols[i] = s.String()
	}
	parts = append(parts, fmt.Sprintf("DEFINE %s)", strings.Join(symbols, ", ")))
	if m.Alias != nil {
		parts = append(parts, m.Alias.String())
	}
	return strings.Join(parts, " ")
}

// XmlTableFactor represents XMLTABLE table-valued function
type XmlTableFactor struct {
	span          token.Span
	Namespaces    []XmlNamespaceDefinition
	RowExpression Expr
	Passing       XmlPassingClause
	Columns       []XmlTableColumn
	Alias         *TableAlias
}

func (x *XmlTableFactor) Span() token.Span { return x.span }
func (x *XmlTableFactor) String() string {
	var parts []string
	parts = append(parts, "XMLTABLE(")
	if len(x.Namespaces) > 0 {
		ns := make([]string, len(x.Namespaces))
		for i, n := range x.Namespaces {
			ns[i] = n.String()
		}
		parts = append(parts, fmt.Sprintf("XMLNAMESPACES(%s), ", strings.Join(ns, ", ")))
	}
	cols := make([]string, len(x.Columns))
	for i, c := range x.Columns {
		cols[i] = c.String()
	}
	parts = append(parts, fmt.Sprintf("%s%s COLUMNS (%s))",
		x.RowExpression.String(), x.Passing.String(), strings.Join(cols, ", ")))
	if x.Alias != nil {
		parts = append(parts, x.Alias.String())
	}
	return strings.Join(parts, "")
}

// SemanticViewTableFactor represents Snowflake SEMANTIC_VIEW function
type SemanticViewTableFactor struct {
	span        token.Span
	Name        ObjectName
	Dimensions  []Expr
	Metrics     []Expr
	Facts       []Expr
	WhereClause Expr
	Alias       *TableAlias
}

func (s *SemanticViewTableFactor) Span() token.Span { return s.span }
func (s *SemanticViewTableFactor) String() string {
	var parts []string
	parts = append(parts, fmt.Sprintf("SEMANTIC_VIEW(%s", s.Name.String()))
	if len(s.Dimensions) > 0 {
		dims := make([]string, len(s.Dimensions))
		for i, d := range s.Dimensions {
			dims[i] = d.String()
		}
		parts = append(parts, fmt.Sprintf("DIMENSIONS %s", strings.Join(dims, ", ")))
	}
	if len(s.Metrics) > 0 {
		metrics := make([]string, len(s.Metrics))
		for i, m := range s.Metrics {
			metrics[i] = m.String()
		}
		parts = append(parts, fmt.Sprintf("METRICS %s", strings.Join(metrics, ", ")))
	}
	if len(s.Facts) > 0 {
		facts := make([]string, len(s.Facts))
		for i, f := range s.Facts {
			facts[i] = f.String()
		}
		parts = append(parts, fmt.Sprintf("FACTS %s", strings.Join(facts, ", ")))
	}
	if s.WhereClause != nil {
		parts = append(parts, fmt.Sprintf("WHERE %s", s.WhereClause.String()))
	}
	parts = append(parts, ")")
	if s.Alias != nil {
		parts = append(parts, s.Alias.String())
	}
	return strings.Join(parts, " ")
}

// Join represents a JOIN clause
type Join struct {
	span         token.Span
	Relation     TableFactor
	Global       bool
	JoinOperator JoinOperatorType
}

func (j *Join) Span() token.Span { return j.span }
func (j *Join) String() string {
	var parts []string
	if j.Global {
		parts = append(parts, "GLOBAL")
	}
	parts = append(parts, j.JoinOperator.String(j.Relation.String()))
	return strings.Join(parts, " ")
}

// JoinOperatorType represents the type of join operation
type JoinOperatorType interface {
	String(relation string) string
}

// StandardJoinOp represents standard JOIN types
type StandardJoinOp struct {
	Type       string
	Constraint JoinConstraint
}

func (j *StandardJoinOp) String(relation string) string {
	prefix := ""
	if _, ok := j.Constraint.(*NaturalJoinConstraint); ok {
		prefix = "NATURAL "
	}
	suffix := ""
	if c, ok := j.Constraint.(*OnJoinConstraint); ok {
		suffix = fmt.Sprintf(" ON %s", c.Expr.String())
	}
	if c, ok := j.Constraint.(*UsingJoinConstraint); ok {
		attrs := make([]string, len(c.Attrs))
		for i, a := range c.Attrs {
			attrs[i] = a.String()
		}
		suffix = fmt.Sprintf(" USING(%s)", strings.Join(attrs, ", "))
	}
	return fmt.Sprintf("%s%s %s%s", prefix, j.Type, relation, suffix)
}

// NaturalJoinConstraint represents a NATURAL join
type NaturalJoinConstraint struct{}

func (n *NaturalJoinConstraint) String() string { return "NATURAL" }

// OnJoinConstraint represents ON expr join constraint
type OnJoinConstraint struct {
	Expr Expr
}

func (o *OnJoinConstraint) String() string {
	return fmt.Sprintf("ON %s", o.Expr.String())
}

// UsingJoinConstraint represents USING(...) join constraint
type UsingJoinConstraint struct {
	Attrs []ObjectName
}

func (u *UsingJoinConstraint) String() string {
	attrs := make([]string, len(u.Attrs))
	for i, a := range u.Attrs {
		attrs[i] = a.String()
	}
	return fmt.Sprintf("USING(%s)", strings.Join(attrs, ", "))
}

// NoneJoinConstraint represents no constraint (CROSS JOIN)
type NoneJoinConstraint struct{}

func (n *NoneJoinConstraint) String() string { return "" }

// AsOfJoinOperator represents ASOF JOIN
type AsOfJoinOperator struct {
	MatchCondition Expr
	Constraint     JoinConstraint
}

func (a *AsOfJoinOperator) String(relation string) string {
	suffix := ""
	if c, ok := a.Constraint.(*OnJoinConstraint); ok {
		suffix = fmt.Sprintf(" ON %s", c.Expr.String())
	}
	if c, ok := a.Constraint.(*UsingJoinConstraint); ok {
		attrs := make([]string, len(c.Attrs))
		for i, attr := range c.Attrs {
			attrs[i] = attr.String()
		}
		suffix = fmt.Sprintf(" USING(%s)", strings.Join(attrs, ", "))
	}
	return fmt.Sprintf("ASOF JOIN %s MATCH_CONDITION (%s)%s", relation, a.MatchCondition.String(), suffix)
}

// CrossApplyJoinOperator represents CROSS APPLY (MSSQL)
type CrossApplyJoinOperator struct{}

func (c *CrossApplyJoinOperator) String(relation string) string {
	return fmt.Sprintf("CROSS APPLY %s", relation)
}

// OuterApplyJoinOperator represents OUTER APPLY (MSSQL)
type OuterApplyJoinOperator struct{}

func (o *OuterApplyJoinOperator) String(relation string) string {
	return fmt.Sprintf("OUTER APPLY %s", relation)
}

// JoinConstraint is the interface for join constraints
type JoinConstraint interface {
	fmt.Stringer
}
