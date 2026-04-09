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

// Query represents a complete SELECT query expression, optionally
// including WITH, UNION / other set operations, and ORDER BY.
type Query struct {
	span          token.Span
	With          *With
	Body          SetExpr
	OrderBy       *OrderBy
	LimitClause   LimitClause
	Fetch         *Fetch
	Locks         []LockClause
	ForClause     *ForClause
	Settings      []Setting
	FormatClause  *FormatClause
	PipeOperators []PipeOperator
}

// Span returns the source span of this node
func (q *Query) Span() token.Span { return q.span }

// String returns the SQL representation
func (q *Query) String() string {
	var parts []string
	if q.With != nil {
		parts = append(parts, q.With.String())
	}
	parts = append(parts, q.Body.String())
	if q.OrderBy != nil {
		parts = append(parts, q.OrderBy.String())
	}
	if q.LimitClause != nil {
		parts = append(parts, q.LimitClause.String())
	}
	if len(q.Settings) > 0 {
		settings := make([]string, len(q.Settings))
		for i, s := range q.Settings {
			settings[i] = s.String()
		}
		parts = append(parts, "SETTINGS "+strings.Join(settings, ", "))
	}
	if q.Fetch != nil {
		parts = append(parts, q.Fetch.String())
	}
	for _, lock := range q.Locks {
		parts = append(parts, lock.String())
	}
	if q.ForClause != nil {
		parts = append(parts, q.ForClause.String())
	}
	if q.FormatClause != nil {
		parts = append(parts, q.FormatClause.String())
	}
	for _, pipe := range q.PipeOperators {
		parts = append(parts, "|> "+pipe.String())
	}
	return strings.Join(parts, " ")
}

// SetExpr represents a query body expression (SELECT, UNION, EXCEPT, etc.)
type SetExpr interface {
	fmt.Stringer
	Span() token.Span
}

// SelectSetExpr represents a SELECT expression
type SelectSetExpr struct {
	span   token.Span
	Select *Select
}

func (s *SelectSetExpr) Span() token.Span { return s.span }
func (s *SelectSetExpr) String() string   { return s.Select.String() }

// QuerySetExpr represents a parenthesized subquery
type QuerySetExpr struct {
	span  token.Span
	Query *Query
}

func (q *QuerySetExpr) Span() token.Span { return q.span }
func (q *QuerySetExpr) String() string {
	return "(" + q.Query.String() + ")"
}

// SetOperation represents UNION/EXCEPT/INTERSECT operations
type SetOperation struct {
	span          token.Span
	Left          SetExpr
	Op            SetOperator
	SetQuantifier SetQuantifier
	Right         SetExpr
}

func (s *SetOperation) Span() token.Span { return s.span }
func (s *SetOperation) String() string {
	parts := []string{s.Left.String()}
	parts = append(parts, s.Op.String())
	if s.SetQuantifier != SetQuantifierNone {
		parts = append(parts, s.SetQuantifier.String())
	}
	parts = append(parts, s.Right.String())
	return strings.Join(parts, " ")
}

// ValuesSetExpr represents a VALUES expression
type ValuesSetExpr struct {
	span   token.Span
	Values *Values
}

func (v *ValuesSetExpr) Span() token.Span { return v.span }
func (v *ValuesSetExpr) String() string   { return v.Values.String() }

// StatementSetExpr represents a statement in set expression context
type StatementSetExpr struct {
	span      token.Span
	StmtType  string // "INSERT", "UPDATE", "DELETE", "MERGE"
	Statement fmt.Stringer
}

func (s *StatementSetExpr) Span() token.Span { return s.span }
func (s *StatementSetExpr) String() string   { return s.Statement.String() }

// TableSetExpr represents a TABLE command
type TableSetExpr struct {
	span  token.Span
	Table *Table
}

func (t *TableSetExpr) Span() token.Span { return t.span }
func (t *TableSetExpr) String() string   { return t.Table.String() }

// Table represents a TABLE command
type Table struct {
	span       token.Span
	TableName  *string
	SchemaName *string
}

func (t *Table) Span() token.Span { return t.span }
func (t *Table) String() string {
	if t.SchemaName != nil && t.TableName != nil {
		return fmt.Sprintf("TABLE %s.%s", *t.SchemaName, *t.TableName)
	}
	if t.TableName != nil {
		return fmt.Sprintf("TABLE %s", *t.TableName)
	}
	return "TABLE"
}

// ProjectionSelect represents a ClickHouse ADD PROJECTION query.
type ProjectionSelect struct {
	span       token.Span
	Projection []SelectItem
	OrderBy    *OrderBy
	GroupBy    GroupByExpr
}

func (p *ProjectionSelect) Span() token.Span { return p.span }
func (p *ProjectionSelect) String() string {
	parts := []string{"SELECT"}
	proj := make([]string, len(p.Projection))
	for i, item := range p.Projection {
		proj[i] = item.String()
	}
	parts = append(parts, strings.Join(proj, ", "))
	if p.GroupBy != nil {
		parts = append(parts, p.GroupBy.String())
	}
	if p.OrderBy != nil {
		parts = append(parts, p.OrderBy.String())
	}
	return strings.Join(parts, " ")
}

// SelectFlavor represents what kind of SELECT this is
type SelectFlavor int

const (
	SelectFlavorStandard SelectFlavor = iota
	SelectFlavorFromFirst
	SelectFlavorFromFirstNoSelect
)

// SelectModifiers represents MySQL-specific SELECT modifiers
type SelectModifiers struct {
	HighPriority     bool
	StraightJoin     bool
	SqlSmallResult   bool
	SqlBigResult     bool
	SqlBufferResult  bool
	SqlNoCache       bool
	SqlCalcFoundRows bool
}

func (s *SelectModifiers) String() string {
	var parts []string
	if s.HighPriority {
		parts = append(parts, "HIGH_PRIORITY")
	}
	if s.StraightJoin {
		parts = append(parts, "STRAIGHT_JOIN")
	}
	if s.SqlSmallResult {
		parts = append(parts, "SQL_SMALL_RESULT")
	}
	if s.SqlBigResult {
		parts = append(parts, "SQL_BIG_RESULT")
	}
	if s.SqlBufferResult {
		parts = append(parts, "SQL_BUFFER_RESULT")
	}
	if s.SqlNoCache {
		parts = append(parts, "SQL_NO_CACHE")
	}
	if s.SqlCalcFoundRows {
		parts = append(parts, "SQL_CALC_FOUND_ROWS")
	}
	if len(parts) > 0 {
		return " " + strings.Join(parts, " ")
	}
	return ""
}

// IsAnySet returns true if any modifier is set
func (s *SelectModifiers) IsAnySet() bool {
	return s.HighPriority || s.StraightJoin || s.SqlSmallResult ||
		s.SqlBigResult || s.SqlBufferResult || s.SqlNoCache || s.SqlCalcFoundRows
}

// Select represents a restricted SELECT (without CTEs/ORDER BY)
type Select struct {
	span                token.Span
	OptimizerHints      []OptimizerHint
	Distinct            *Distinct
	SelectModifiers     *SelectModifiers
	Top                 *Top
	TopBeforeDistinct   bool
	Projection          []SelectItem
	Exclude             ExcludeSelectItem
	Into                *SelectInto
	From                []TableWithJoins
	LateralViews        []LateralView
	Prewhere            Expr
	Selection           Expr
	ConnectBy           []ConnectByKind
	GroupBy             GroupByExpr
	ClusterBy           []Expr
	DistributeBy        []Expr
	SortBy              []OrderByExpr
	Having              Expr
	NamedWindow         []NamedWindowDefinition
	Qualify             Expr
	WindowBeforeQualify bool
	ValueTableMode      ValueTableMode
	Flavor              SelectFlavor
	Locks               []LockClause
}

// Span returns the source span of this node
func (s *Select) Span() token.Span { return s.span }

// String returns the SQL representation
func (s *Select) String() string {
	var parts []string

	switch s.Flavor {
	case SelectFlavorStandard:
		parts = append(parts, "SELECT")
	case SelectFlavorFromFirst:
		fromStrs := make([]string, len(s.From))
		for i, f := range s.From {
			fromStrs[i] = f.String()
		}
		parts = append(parts, fmt.Sprintf("FROM %s SELECT", strings.Join(fromStrs, ", ")))
	case SelectFlavorFromFirstNoSelect:
		fromStrs := make([]string, len(s.From))
		for i, f := range s.From {
			fromStrs[i] = f.String()
		}
		parts = append(parts, fmt.Sprintf("FROM %s", strings.Join(fromStrs, ", ")))
		return strings.Join(parts, " ")
	}

	for _, hint := range s.OptimizerHints {
		parts = append(parts, hint.String())
	}

	if s.ValueTableMode != ValueTableModeNone {
		parts = append(parts, s.ValueTableMode.String())
	}

	if s.Top != nil && s.TopBeforeDistinct {
		parts = append(parts, s.Top.String())
	}

	if s.Distinct != nil {
		parts = append(parts, s.Distinct.String())
	}

	if s.Top != nil && !s.TopBeforeDistinct {
		parts = append(parts, s.Top.String())
	}

	if s.SelectModifiers != nil {
		if modStr := s.SelectModifiers.String(); modStr != "" {
			parts = append(parts, strings.TrimSpace(modStr))
		}
	}

	proj := make([]string, len(s.Projection))
	for i, item := range s.Projection {
		proj[i] = item.String()
	}
	if len(proj) > 0 {
		parts = append(parts, strings.Join(proj, ", "))
	}

	if s.Exclude != nil {
		parts = append(parts, s.Exclude.String())
	}
	if s.Into != nil {
		parts = append(parts, s.Into.String())
	}
	if s.Flavor == SelectFlavorStandard && len(s.From) > 0 {
		fromStrs := make([]string, len(s.From))
		for i, f := range s.From {
			fromStrs[i] = f.String()
		}
		parts = append(parts, "FROM "+strings.Join(fromStrs, ", "))
	}
	for _, lv := range s.LateralViews {
		parts = append(parts, lv.String())
	}
	if s.Prewhere != nil {
		parts = append(parts, "PREWHERE "+s.Prewhere.String())
	}
	if s.Selection != nil {
		parts = append(parts, "WHERE "+s.Selection.String())
	}
	for _, cb := range s.ConnectBy {
		parts = append(parts, cb.String())
	}
	if s.GroupBy != nil {
		parts = append(parts, s.GroupBy.String())
	}
	if len(s.ClusterBy) > 0 {
		clusterBy := make([]string, len(s.ClusterBy))
		for i, e := range s.ClusterBy {
			clusterBy[i] = e.String()
		}
		parts = append(parts, "CLUSTER BY "+strings.Join(clusterBy, ", "))
	}
	if len(s.DistributeBy) > 0 {
		distributeBy := make([]string, len(s.DistributeBy))
		for i, e := range s.DistributeBy {
			distributeBy[i] = e.String()
		}
		parts = append(parts, "DISTRIBUTE BY "+strings.Join(distributeBy, ", "))
	}
	if len(s.SortBy) > 0 {
		sortBy := make([]string, len(s.SortBy))
		for i, e := range s.SortBy {
			sortBy[i] = e.String()
		}
		parts = append(parts, "SORT BY "+strings.Join(sortBy, ", "))
	}
	if s.Having != nil {
		parts = append(parts, "HAVING "+s.Having.String())
	}
	if s.WindowBeforeQualify {
		if len(s.NamedWindow) > 0 {
			windows := make([]string, len(s.NamedWindow))
			for i, w := range s.NamedWindow {
				windows[i] = w.String()
			}
			parts = append(parts, "WINDOW "+strings.Join(windows, ", "))
		}
		if s.Qualify != nil {
			parts = append(parts, "QUALIFY "+s.Qualify.String())
		}
	} else {
		if s.Qualify != nil {
			parts = append(parts, "QUALIFY "+s.Qualify.String())
		}
		if len(s.NamedWindow) > 0 {
			windows := make([]string, len(s.NamedWindow))
			for i, w := range s.NamedWindow {
				windows[i] = w.String()
			}
			parts = append(parts, "WINDOW "+strings.Join(windows, ", "))
		}
	}

	// Add FOR UPDATE/FOR SHARE lock clauses
	for _, lock := range s.Locks {
		parts = append(parts, lock.String())
	}

	return strings.Join(parts, " ")
}

// OptimizerHint represents query optimizer hints (MySQL, Oracle)
type OptimizerHint struct {
	span token.Span
	Hint string
}

func (o *OptimizerHint) Span() token.Span { return o.span }
func (o *OptimizerHint) String() string   { return "/*+" + o.Hint + "*/" }

// SelectItem represents one item of the comma-separated list following SELECT
type SelectItem interface {
	fmt.Stringer
}

// UnnamedExpr represents an expression not followed by [AS] alias
type UnnamedExpr struct {
	Expr Expr
}

func (u *UnnamedExpr) String() string { return u.Expr.String() }

// AliasedExpr represents an expression followed by [AS] alias
type AliasedExpr struct {
	Expr     Expr
	Alias    Ident
	Explicit bool // true if AS keyword was used
}

func (e *AliasedExpr) String() string {
	// Canonical form always includes AS keyword
	return fmt.Sprintf("%s AS %s", e.Expr.String(), e.Alias.String())
}

// QualifiedWildcard represents an expression followed by wildcard expansion
type QualifiedWildcard struct {
	Kind              SelectItemQualifiedWildcardKind
	AdditionalOptions WildcardAdditionalOptions
}

func (q *QualifiedWildcard) String() string {
	return q.Kind.String() + q.AdditionalOptions.String()
}

// Wildcard represents an unqualified *
type Wildcard struct {
	AdditionalOptions WildcardAdditionalOptions
}

func (w *Wildcard) String() string {
	return "*" + w.AdditionalOptions.String()
}

// SelectItemQualifiedWildcardKind represents expression behind a wildcard expansion
type SelectItemQualifiedWildcardKind interface {
	fmt.Stringer
}

// ObjectNameWildcard represents an object name as wildcard prefix
type ObjectNameWildcard struct {
	Name ObjectName
}

func (o *ObjectNameWildcard) String() string {
	return o.Name.String() + ".*"
}

// ExprWildcard represents an expression as wildcard prefix
type ExprWildcard struct {
	Expr Expr
}

func (e *ExprWildcard) String() string {
	return e.Expr.String() + ".*"
}

// WildcardAdditionalOptions represents options for wildcards
type WildcardAdditionalOptions struct {
	OptIlike   *IlikeSelectItem
	OptExclude ExcludeSelectItem
	OptExcept  *ExceptSelectItem
	OptReplace *ReplaceSelectItem
	OptRename  RenameSelectItem
	OptAlias   *Ident
}

func (w *WildcardAdditionalOptions) String() string {
	var parts []string
	if w.OptIlike != nil {
		parts = append(parts, w.OptIlike.String())
	}
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
	if w.OptAlias != nil {
		parts = append(parts, "AS "+w.OptAlias.String())
	}
	if len(parts) > 0 {
		return " " + strings.Join(parts, " ")
	}
	return ""
}

// IlikeSelectItem represents Snowflake ILIKE information
type IlikeSelectItem struct {
	Pattern string
}

func (i *IlikeSelectItem) String() string {
	return fmt.Sprintf("ILIKE '%s'", escapeSingleQuote(i.Pattern))
}

func escapeSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// ExcludeSelectItem represents Snowflake EXCLUDE information.
// This is an interface with two variants: Single and Multiple.
type ExcludeSelectItem interface {
	isExcludeSelectItem()
	String() string
}

// ExcludeSelectItemSingle represents a single column EXCLUDE (without parens).
type ExcludeSelectItemSingle struct {
	Column ObjectName
}

func (e *ExcludeSelectItemSingle) isExcludeSelectItem() {}
func (e *ExcludeSelectItemSingle) String() string {
	return "EXCLUDE " + e.Column.String()
}

// ExcludeSelectItemMultiple represents multiple columns EXCLUDE (with parens).
type ExcludeSelectItemMultiple struct {
	Columns []ObjectName
}

func (e *ExcludeSelectItemMultiple) isExcludeSelectItem() {}
func (e *ExcludeSelectItemMultiple) String() string {
	parts := make([]string, len(e.Columns))
	for i, col := range e.Columns {
		parts[i] = col.String()
	}
	return "EXCLUDE (" + strings.Join(parts, ", ") + ")"
}

// ExceptSelectItem represents BigQuery EXCEPT information
type ExceptSelectItem struct {
	FirstElement       Ident
	AdditionalElements []Ident
}

func (e *ExceptSelectItem) String() string {
	if len(e.AdditionalElements) == 0 {
		return fmt.Sprintf("EXCEPT (%s)", e.FirstElement.String())
	}
	parts := make([]string, len(e.AdditionalElements))
	for i, elem := range e.AdditionalElements {
		parts[i] = elem.String()
	}
	return fmt.Sprintf("EXCEPT (%s, %s)", e.FirstElement.String(), strings.Join(parts, ", "))
}

// ReplaceSelectItem represents BigQuery/ClickHouse/Snowflake REPLACE information
type ReplaceSelectItem struct {
	Items []*ReplaceSelectElement
}

func (r *ReplaceSelectItem) String() string {
	items := make([]string, len(r.Items))
	for i, item := range r.Items {
		items[i] = item.String()
	}
	return "REPLACE (" + strings.Join(items, ", ") + ")"
}

// ReplaceSelectElement represents a single element in REPLACE
type ReplaceSelectElement struct {
	Expr       Expr
	ColumnName Ident
	AsKeyword  bool
}

func (r *ReplaceSelectElement) String() string {
	if r.AsKeyword {
		return fmt.Sprintf("%s AS %s", r.Expr.String(), r.ColumnName.String())
	}
	return fmt.Sprintf("%s %s", r.Expr.String(), r.ColumnName.String())
}

// RenameSelectItem represents Snowflake RENAME information.
// This is an interface with two variants: Single and Multiple.
type RenameSelectItem interface {
	isRenameSelectItem()
	String() string
}

// RenameSelectItemSingle represents a single column RENAME (without parens).
type RenameSelectItemSingle struct {
	Column IdentWithAlias
}

func (r *RenameSelectItemSingle) isRenameSelectItem() {}
func (r *RenameSelectItemSingle) String() string {
	return "RENAME " + r.Column.String()
}

// RenameSelectItemMultiple represents multiple columns RENAME (with parens).
type RenameSelectItemMultiple struct {
	Columns []IdentWithAlias
}

func (r *RenameSelectItemMultiple) isRenameSelectItem() {}
func (r *RenameSelectItemMultiple) String() string {
	parts := make([]string, len(r.Columns))
	for i, col := range r.Columns {
		parts[i] = col.String()
	}
	return "RENAME (" + strings.Join(parts, ", ") + ")"
}

// IdentWithAlias represents an identifier with an alias
type IdentWithAlias struct {
	Ident Ident
	Alias Ident
}

func (i *IdentWithAlias) String() string {
	return fmt.Sprintf("%s AS %s", i.Ident.String(), i.Alias.String())
}

// Ident represents a SQL identifier
type Ident struct {
	Value      string
	QuoteStyle *byte
}

func (i Ident) String() string {
	if i.QuoteStyle != nil {
		quote := string(*i.QuoteStyle)
		return quote + i.Value + quote
	}
	return i.Value
}

// ObjectName represents a qualified name (schema.table, etc.)
type ObjectName struct {
	Parts []Ident
}

func (o ObjectName) String() string {
	parts := make([]string, len(o.Parts))
	for i, p := range o.Parts {
		parts[i] = p.String()
	}
	return strings.Join(parts, ".")
}

// Expr is a placeholder for expression types
// This will be defined in a separate expression package
type Expr interface {
	fmt.Stringer
}
