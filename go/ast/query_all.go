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

// Package ast provides consolidated SQL query AST types.
// This file contains query types migrated from the query/ subpackage.
//
// Naming convention: Q-prefix types (e.g., QTableWithJoins, QSetOperation)
// are the consolidated versions that use the main ast package interfaces.
package ast

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// ============================================================================
// Query Body Types (SetExpr implementations)
// ============================================================================

// QSetExpr represents a query body expression (SELECT, UNION, EXCEPT, etc.)
// This is an interface that corresponds to SetExpr in the query package
type QSetExpr interface {
	fmt.Stringer
	Span() token.Span
}

// QSelectSetExpr represents a SELECT expression
type QSelectSetExpr struct {
	span   token.Span
	Select *QSelect
}

func (s *QSelectSetExpr) Span() token.Span { return s.span }
func (s *QSelectSetExpr) String() string   { return s.Select.String() }

// QQuerySetExpr represents a parenthesized subquery
type QQuerySetExpr struct {
	span  token.Span
	Query QSetExpr
}

func (q *QQuerySetExpr) Span() token.Span { return q.span }
func (q *QQuerySetExpr) String() string {
	return "(" + q.Query.String() + ")"
}

// QValuesSetExpr represents a VALUES expression
type QValuesSetExpr struct {
	span   token.Span
	Values *QValues
}

func (v *QValuesSetExpr) Span() token.Span { return v.span }
func (v *QValuesSetExpr) String() string   { return v.Values.String() }

// QStatementSetExpr represents a statement in set expression context
type QStatementSetExpr struct {
	span      token.Span
	StmtType  string // "INSERT", "UPDATE", "DELETE", "MERGE"
	Statement fmt.Stringer
}

func (s *QStatementSetExpr) Span() token.Span { return s.span }
func (s *QStatementSetExpr) String() string   { return s.Statement.String() }

// ============================================================================
// Set Operations (UNION, EXCEPT, INTERSECT)
// ============================================================================

// QSetOperation represents UNION/EXCEPT/INTERSECT operations
type QSetOperation struct {
	span          token.Span
	Left          QSetExpr
	Op            QSetOperator
	SetQuantifier QSetQuantifier
	Right         QSetExpr
}

func (s *QSetOperation) Span() token.Span { return s.span }
func (s *QSetOperation) String() string {
	parts := []string{s.Left.String()}
	parts = append(parts, s.Op.String())
	if s.SetQuantifier != QSetQuantifierNone {
		parts = append(parts, s.SetQuantifier.String())
	}
	parts = append(parts, s.Right.String())
	return strings.Join(parts, " ")
}

// QSetOperator represents UNION, EXCEPT, INTERSECT, MINUS
type QSetOperator int

const (
	QSetOperatorUnion QSetOperator = iota
	QSetOperatorExcept
	QSetOperatorIntersect
	QSetOperatorMinus
)

func (s QSetOperator) String() string {
	switch s {
	case QSetOperatorUnion:
		return "UNION"
	case QSetOperatorExcept:
		return "EXCEPT"
	case QSetOperatorIntersect:
		return "INTERSECT"
	case QSetOperatorMinus:
		return "MINUS"
	default:
		return ""
	}
}

// QSetQuantifier represents quantifiers for set operations
type QSetQuantifier int

const (
	QSetQuantifierNone QSetQuantifier = iota
	QSetQuantifierAll
	QSetQuantifierDistinct
	QSetQuantifierByName
	QSetQuantifierAllByName
	QSetQuantifierDistinctByName
)

func (s QSetQuantifier) String() string {
	switch s {
	case QSetQuantifierAll:
		return "ALL"
	case QSetQuantifierDistinct:
		return "DISTINCT"
	case QSetQuantifierByName:
		return "BY NAME"
	case QSetQuantifierAllByName:
		return "ALL BY NAME"
	case QSetQuantifierDistinctByName:
		return "DISTINCT BY NAME"
	default:
		return ""
	}
}

// ============================================================================
// SELECT Components
// ============================================================================

// QSelect represents a complete SELECT query body
type QSelect struct {
	span                token.Span
	OptimizerHints      []QOptimizerHint
	Distinct            *QDistinct
	SelectModifiers     *QSelectModifiers
	Top                 *QTop
	TopBeforeDistinct   bool
	Projection          []QSelectItem
	Exclude             *QExcludeSelectItem
	Into                *QSelectInto
	From                []QTableWithJoins
	LateralViews        []QLateralView
	Prewhere            Expr
	Selection           Expr
	ConnectBy           []QConnectByKind
	GroupBy             QGroupByExpr
	ClusterBy           []Expr
	DistributeBy        []Expr
	SortBy              []QOrderByExpr
	Having              Expr
	NamedWindow         []QNamedWindowDefinition
	Qualify             Expr
	WindowBeforeQualify bool
	ValueTableMode      QValueTableMode
	Flavor              QSelectFlavor
}

func (s *QSelect) Span() token.Span { return s.span }
func (s *QSelect) String() string {
	var parts []string

	switch s.Flavor {
	case QSelectFlavorStandard:
		parts = append(parts, "SELECT")
	case QSelectFlavorFromFirst:
		fromStrs := make([]string, len(s.From))
		for i, f := range s.From {
			fromStrs[i] = f.String()
		}
		parts = append(parts, fmt.Sprintf("FROM %s SELECT", strings.Join(fromStrs, ", ")))
	case QSelectFlavorFromFirstNoSelect:
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

	if s.ValueTableMode != QValueTableModeNone {
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
	parts = append(parts, strings.Join(proj, ", "))

	if s.Exclude != nil {
		parts = append(parts, s.Exclude.String())
	}
	if s.Into != nil {
		parts = append(parts, s.Into.String())
	}
	if s.Flavor == QSelectFlavorStandard && len(s.From) > 0 {
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

	return strings.Join(parts, " ")
}

// QSelectFlavor represents what kind of SELECT this is
type QSelectFlavor int

const (
	QSelectFlavorStandard QSelectFlavor = iota
	QSelectFlavorFromFirst
	QSelectFlavorFromFirstNoSelect
)

// QSelectModifiers represents MySQL-specific SELECT modifiers
type QSelectModifiers struct {
	HighPriority     bool
	StraightJoin     bool
	SqlSmallResult   bool
	SqlBigResult     bool
	SqlBufferResult  bool
	SqlNoCache       bool
	SqlCalcFoundRows bool
}

func (s *QSelectModifiers) String() string {
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

// QDistinct represents ALL, DISTINCT, or DISTINCT ON modifiers
type QDistinct int

const (
	QDistinctAll QDistinct = iota
	QDistinctDistinct
	QDistinctOn
)

func (d QDistinct) String() string {
	switch d {
	case QDistinctAll:
		return "ALL"
	case QDistinctDistinct:
		return "DISTINCT"
	default:
		return ""
	}
}

// QDistinctWithExprs represents DISTINCT with expressions
type QDistinctWithExprs struct {
	Distinct QDistinct
	Exprs    []Expr
}

func (d *QDistinctWithExprs) String() string {
	switch d.Distinct {
	case QDistinctAll:
		return "ALL"
	case QDistinctDistinct:
		return "DISTINCT"
	case QDistinctOn:
		parts := make([]string, len(d.Exprs))
		for i, e := range d.Exprs {
			parts[i] = e.String()
		}
		return fmt.Sprintf("DISTINCT ON (%s)", strings.Join(parts, ", "))
	default:
		return ""
	}
}

// QTop represents MSSQL TOP clause
type QTop struct {
	span     token.Span
	WithTies bool
	Percent  bool
	Quantity *QTopQuantity
}

func (t *QTop) String() string {
	extension := ""
	if t.WithTies {
		extension = " WITH TIES"
	}
	if t.Quantity != nil {
		percent := ""
		if t.Percent {
			percent = " PERCENT"
		}
		return fmt.Sprintf("TOP %s%s%s", t.Quantity.String(), percent, extension)
	}
	return fmt.Sprintf("TOP%s", extension)
}

// QTopQuantity represents quantity in TOP clause
type QTopQuantity struct {
	Expr     Expr
	Constant *uint64
}

func (t *QTopQuantity) String() string {
	if t.Expr != nil {
		return fmt.Sprintf("(%s)", t.Expr.String())
	}
	if t.Constant != nil {
		return fmt.Sprintf("%d", *t.Constant)
	}
	return ""
}

// ============================================================================
// Select Items
// ============================================================================

// QSelectItem represents one item in the SELECT list
type QSelectItem interface {
	fmt.Stringer
}

// QUnnamedExpr represents an expression not followed by [AS] alias
type QUnnamedExpr struct {
	Expr Expr
}

func (u *QUnnamedExpr) String() string { return u.Expr.String() }

// QAliasedExpr represents an expression followed by [AS] alias
type QAliasedExpr struct {
	Expr  Expr
	Alias *Ident
}

func (e *QAliasedExpr) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s AS %s", e.Expr.String(), e.Alias.String())
	}
	return e.Expr.String()
}

// QWildcard represents an unqualified *
type QWildcard struct {
	AdditionalOptions QWildcardAdditionalOptions
}

func (w *QWildcard) String() string {
	return "*" + w.AdditionalOptions.String()
}

// QQualifiedWildcard represents an expression.* wildcard
type QQualifiedWildcard struct {
	Kind              QSelectItemQualifiedWildcardKind
	AdditionalOptions QWildcardAdditionalOptions
}

func (q *QQualifiedWildcard) String() string {
	return q.Kind.String() + q.AdditionalOptions.String()
}

// QSelectItemQualifiedWildcardKind represents expression behind a wildcard expansion
type QSelectItemQualifiedWildcardKind interface {
	fmt.Stringer
}

// QObjectNameWildcard represents an object name as wildcard prefix
type QObjectNameWildcard struct {
	Name *ObjectName
}

func (o *QObjectNameWildcard) String() string {
	return o.Name.String() + ".*"
}

// QExprWildcard represents an expression as wildcard prefix
type QExprWildcard struct {
	Expr Expr
}

func (e *QExprWildcard) String() string {
	return e.Expr.String() + ".*"
}

// QWildcardAdditionalOptions represents options for wildcards
type QWildcardAdditionalOptions struct {
	OptIlike   *QIlikeSelectItem
	OptExclude *QExcludeSelectItem
	OptExcept  *QExceptSelectItem
	OptReplace *QReplaceSelectItem
	OptRename  *QRenameSelectItem
	OptAlias   *Ident
}

func (w *QWildcardAdditionalOptions) String() string {
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

// QIlikeSelectItem represents Snowflake ILIKE information
type QIlikeSelectItem struct {
	Pattern string
}

func (i *QIlikeSelectItem) String() string {
	return fmt.Sprintf("ILIKE '%s'", escapeSingleQuote(i.Pattern))
}

func escapeSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// QExcludeSelectItem represents Snowflake EXCLUDE information
type QExcludeSelectItem struct {
	Columns []*ObjectName
}

func (e *QExcludeSelectItem) String() string {
	if len(e.Columns) == 1 {
		return "EXCLUDE " + e.Columns[0].String()
	}
	parts := make([]string, len(e.Columns))
	for i, col := range e.Columns {
		parts[i] = col.String()
	}
	return "EXCLUDE (" + strings.Join(parts, ", ") + ")"
}

// QExceptSelectItem represents BigQuery EXCEPT information
type QExceptSelectItem struct {
	FirstElement       *Ident
	AdditionalElements []*Ident
}

func (e *QExceptSelectItem) String() string {
	if len(e.AdditionalElements) == 0 {
		return fmt.Sprintf("EXCEPT (%s)", e.FirstElement.String())
	}
	parts := make([]string, len(e.AdditionalElements))
	for i, elem := range e.AdditionalElements {
		parts[i] = elem.String()
	}
	return fmt.Sprintf("EXCEPT (%s, %s)", e.FirstElement.String(), strings.Join(parts, ", "))
}

// QReplaceSelectItem represents BigQuery/ClickHouse/Snowflake REPLACE information
type QReplaceSelectItem struct {
	Items []*QReplaceSelectElement
}

func (r *QReplaceSelectItem) String() string {
	items := make([]string, len(r.Items))
	for i, item := range r.Items {
		items[i] = item.String()
	}
	return "REPLACE (" + strings.Join(items, ", ") + ")"
}

// QReplaceSelectElement represents a single element in REPLACE
type QReplaceSelectElement struct {
	Expr       Expr
	ColumnName *Ident
	AsKeyword  bool
}

func (r *QReplaceSelectElement) String() string {
	if r.AsKeyword {
		return fmt.Sprintf("%s AS %s", r.Expr.String(), r.ColumnName.String())
	}
	return fmt.Sprintf("%s %s", r.Expr.String(), r.ColumnName.String())
}

// QRenameSelectItem represents Snowflake RENAME information
type QRenameSelectItem struct {
	Columns []QIdentWithAlias
}

func (r *QRenameSelectItem) String() string {
	if len(r.Columns) == 1 {
		return "RENAME " + r.Columns[0].String()
	}
	parts := make([]string, len(r.Columns))
	for i, col := range r.Columns {
		parts[i] = col.String()
	}
	return "RENAME (" + strings.Join(parts, ", ") + ")"
}

// QIdentWithAlias represents an identifier with an alias
type QIdentWithAlias struct {
	Ident *Ident
	Alias *Ident
}

func (i *QIdentWithAlias) String() string {
	return fmt.Sprintf("%s AS %s", i.Ident.String(), i.Alias.String())
}

// QExcludeSelectItem represents excluded columns (Snowflake)
func (e *QExcludeSelectItem) Span() token.Span { return token.Span{} }

// ============================================================================
// Table References and Joins
// ============================================================================

// QTableWithJoins represents a left table followed by zero or more joins
type QTableWithJoins struct {
	span     token.Span
	Relation QTableFactor
	Joins    []QJoin
}

func (t *QTableWithJoins) Span() token.Span { return t.span }
func (t *QTableWithJoins) String() string {
	parts := []string{t.Relation.String()}
	for _, join := range t.Joins {
		parts = append(parts, join.String())
	}
	return strings.Join(parts, " ")
}

// QTableFactor represents a table name or parenthesized subquery with optional alias
type QTableFactor struct {
	span           token.Span
	Kind           QTableFactorKind
	Name           *ObjectName
	Alias          *QTableAlias
	Args           *QTableFunctionArgs
	WithHints      []Expr
	Version        *QTableVersion
	WithOrdinality bool
	Partitions     []*Ident
	JsonPath       *QJsonPath
	Sample         *QTableSampleKind
	IndexHints     []QTableIndexHints
}

func (t *QTableFactor) Span() token.Span { return t.span }
func (t *QTableFactor) String() string {
	var parts []string
	if t.Name != nil {
		parts = append(parts, t.Name.String())
	}
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

// QTableFactorKind represents the kind of table factor
type QTableFactorKind int

const (
	QTableFactorTable QTableFactorKind = iota
	QTableFactorDerived
	QTableFactorFunction
	QTableFactorUNNEST
	QTableFactorJSONTable
	QTableFactorPivot
	QTableFactorUnpivot
)

// QTableAlias represents a table reference alias
type QTableAlias struct {
	span     token.Span
	Explicit bool
	Name     *Ident
	Columns  []QTableAliasColumnDef
}

func (t *QTableAlias) String() string {
	parts := []string{}
	if t.Explicit {
		parts = append(parts, "AS")
	}
	if t.Name != nil {
		parts = append(parts, t.Name.String())
	}
	if len(t.Columns) > 0 {
		cols := make([]string, len(t.Columns))
		for i, c := range t.Columns {
			cols[i] = c.String()
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(cols, ", ")))
	}
	return strings.Join(parts, " ")
}

// QTableAliasColumnDef represents a column definition in table alias
type QTableAliasColumnDef struct {
	Name     *Ident
	DataType Expr
}

func (t *QTableAliasColumnDef) String() string {
	if t.DataType != nil {
		return fmt.Sprintf("%s %s", t.Name.String(), t.DataType.String())
	}
	return t.Name.String()
}

// QTableFunctionArgs represents arguments to a table-valued function
type QTableFunctionArgs struct {
	Args     []QFunctionArg
	Settings *[]QSetting
}

type QFunctionArg struct {
	Expr Expr
}

func (f *QFunctionArg) String() string {
	if f.Expr != nil {
		return f.Expr.String()
	}
	return ""
}

// QTableVersion represents table version selection
type QTableVersion struct {
	Type QTableVersionType
	Expr Expr
}

type QTableVersionType int

const (
	QTableVersionForSystemTimeAsOf QTableVersionType = iota
	QTableVersionTimestampAsOf
	QTableVersionVersionAsOf
	QTableVersionFunction
)

func (t *QTableVersion) String() string {
	switch t.Type {
	case QTableVersionForSystemTimeAsOf:
		return fmt.Sprintf("FOR SYSTEM_TIME AS OF %s", t.Expr.String())
	case QTableVersionTimestampAsOf:
		return fmt.Sprintf("TIMESTAMP AS OF %s", t.Expr.String())
	case QTableVersionVersionAsOf:
		return fmt.Sprintf("VERSION AS OF %s", t.Expr.String())
	case QTableVersionFunction:
		return t.Expr.String()
	default:
		return ""
	}
}

// QJsonPath represents a PartiQL JsonPath
type QJsonPath struct {
	Path string
}

func (j *QJsonPath) String() string { return j.Path }

// QTableSampleKind represents table sample location
type QTableSampleKind struct {
	BeforeTableAlias *QTableSample
	AfterTableAlias  *QTableSample
}

// QTableSample represents a TABLESAMPLE clause
type QTableSample struct {
	span     token.Span
	Modifier QTableSampleModifier
	Name     *QTableSampleMethod
	Quantity *QTableSampleQuantity
	Seed     *QTableSampleSeed
	Bucket   *QTableSampleBucket
	Offset   Expr
}

type QTableSampleModifier int

const (
	QTableSampleModifierSample QTableSampleModifier = iota
	QTableSampleModifierTableSample
)

type QTableSampleMethod int

const (
	QTableSampleMethodRow QTableSampleMethod = iota
	QTableSampleMethodBernoulli
	QTableSampleMethodSystem
	QTableSampleMethodBlock
)

type QTableSampleQuantity struct {
	Parenthesized bool
	Value         Expr
	Unit          *QTableSampleUnit
}

type QTableSampleUnit int

const (
	QTableSampleUnitRows QTableSampleUnit = iota
	QTableSampleUnitPercent
)

type QTableSampleSeed struct {
	Modifier QTableSampleSeedModifier
	Value    string
}

type QTableSampleSeedModifier int

const (
	QTableSampleSeedModifierRepeatable QTableSampleSeedModifier = iota
	QTableSampleSeedModifierSeed
)

type QTableSampleBucket struct {
	Bucket string
	Total  string
	On     Expr
}

type QTableIndexHints struct {
	HintType   QTableIndexHintType
	IndexType  QTableIndexType
	ForClause  *QTableIndexHintForClause
	IndexNames []*Ident
}

type QTableIndexHintType int

const (
	QTableIndexHintTypeUse QTableIndexHintType = iota
	QTableIndexHintTypeIgnore
	QTableIndexHintTypeForce
)

type QTableIndexType int

const (
	QTableIndexTypeIndex QTableIndexType = iota
	QTableIndexTypeKey
)

type QTableIndexHintForClause int

const (
	QTableIndexHintForClauseJoin QTableIndexHintForClause = iota
	QTableIndexHintForClauseOrderBy
	QTableIndexHintForClauseGroupBy
)

func (t *QTableIndexHints) String() string {
	return "" // Simplified
}

func (t *QTableSample) String() string {
	return "" // Simplified
}

// ============================================================================
// Joins
// ============================================================================

// QJoin represents a JOIN clause
type QJoin struct {
	span         token.Span
	Relation     QTableFactor
	Global       bool
	JoinOperator QJoinOperatorType
}

func (j *QJoin) Span() token.Span { return j.span }
func (j *QJoin) String() string {
	var parts []string
	if j.Global {
		parts = append(parts, "GLOBAL")
	}
	parts = append(parts, j.JoinOperator.String(j.Relation.String()))
	return strings.Join(parts, " ")
}

// QJoinOperatorType represents the type of join operation
type QJoinOperatorType interface {
	String(relation string) string
}

// QStandardJoinOp represents standard JOIN types
type QStandardJoinOp struct {
	Type       string
	Constraint QJoinConstraint
}

func (j *QStandardJoinOp) String(relation string) string {
	prefix := ""
	if j.Constraint.Kind == QJoinConstraintNatural {
		prefix = "NATURAL "
	}
	suffix := ""
	switch j.Constraint.Kind {
	case QJoinConstraintOn:
		if j.Constraint.On != nil {
			suffix = fmt.Sprintf(" ON %s", j.Constraint.On.String())
		}
	case QJoinConstraintUsing:
		if len(j.Constraint.Using) > 0 {
			attrs := make([]string, len(j.Constraint.Using))
			for i, a := range j.Constraint.Using {
				attrs[i] = a.String()
			}
			suffix = fmt.Sprintf(" USING(%s)", strings.Join(attrs, ", "))
		}
	}
	return fmt.Sprintf("%s%s %s%s", prefix, j.Type, relation, suffix)
}

// QJoinConstraint represents join constraint data
type QJoinConstraint struct {
	Kind  QJoinConstraintKind
	On    Expr
	Using []*ObjectName
}

// QJoinConstraintKind enumerates join constraint types
type QJoinConstraintKind int

const (
	QJoinConstraintNone QJoinConstraintKind = iota
	QJoinConstraintOn
	QJoinConstraintUsing
	QJoinConstraintNatural
)

// ============================================================================
// ORDER BY
// ============================================================================

// QOrderByExpr represents an expression in ORDER BY
type QOrderByExpr struct {
	span     token.Span
	Expr     Expr
	Options  QOrderByOptions
	WithFill *QWithFill
}

func (o *QOrderByExpr) Span() token.Span { return o.span }
func (o *QOrderByExpr) String() string {
	parts := []string{o.Expr.String()}
	parts = append(parts, o.Options.String())
	if o.WithFill != nil {
		parts = append(parts, o.WithFill.String())
	}
	return strings.Join(parts, "")
}

// QOrderByOptions represents ASC/DESC and NULLS FIRST/LAST options
type QOrderByOptions struct {
	Asc        *bool
	NullsFirst *bool
}

func (o *QOrderByOptions) String() string {
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

// QWithFill represents ClickHouse WITH FILL modifier
type QWithFill struct {
	From Expr
	To   Expr
	Step Expr
}

func (w *QWithFill) String() string {
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

// ============================================================================
// GROUP BY
// ============================================================================

// QGroupByExpr represents the GROUP BY clause
type QGroupByExpr interface {
	fmt.Stringer
}

// QGroupByAll represents GROUP BY ALL
type QGroupByAll struct {
	Modifiers []QGroupByWithModifier
}

func (g *QGroupByAll) String() string {
	parts := []string{"GROUP BY ALL"}
	for _, m := range g.Modifiers {
		parts = append(parts, m.String())
	}
	return strings.Join(parts, " ")
}

// QGroupByExpressions represents GROUP BY with expressions
type QGroupByExpressions struct {
	Expressions []Expr
	Modifiers   []QGroupByWithModifier
}

func (g *QGroupByExpressions) String() string {
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

// QGroupByWithModifier represents modifiers for GROUP BY
type QGroupByWithModifier int

const (
	QGroupByWithModifierRollup QGroupByWithModifier = iota
	QGroupByWithModifierCube
	QGroupByWithModifierTotals
	QGroupByWithModifierGroupingSets
)

func (g QGroupByWithModifier) String() string {
	switch g {
	case QGroupByWithModifierRollup:
		return "WITH ROLLUP"
	case QGroupByWithModifierCube:
		return "WITH CUBE"
	case QGroupByWithModifierTotals:
		return "WITH TOTALS"
	default:
		return ""
	}
}

// ============================================================================
// VALUES
// ============================================================================

// QValues represents an explicit VALUES clause
type QValues struct {
	span         token.Span
	ExplicitRow  bool
	ValueKeyword bool
	Rows         [][]Expr
}

func (v *QValues) String() string {
	keyword := "VALUES"
	if v.ValueKeyword {
		keyword = "VALUE"
	}
	prefix := ""
	if v.ExplicitRow {
		prefix = "ROW"
	}
	var rowStrs []string
	for _, row := range v.Rows {
		values := make([]string, len(row))
		for i, val := range row {
			values[i] = val.String()
		}
		if v.ExplicitRow {
			rowStrs = append(rowStrs, fmt.Sprintf("%s(%s)", prefix, strings.Join(values, ", ")))
		} else {
			rowStrs = append(rowStrs, fmt.Sprintf("(%s)", strings.Join(values, ", ")))
		}
	}
	return keyword + " " + strings.Join(rowStrs, ", ")
}

// ============================================================================
// LATERAL VIEW
// ============================================================================

// QLateralView represents a Hive LATERAL VIEW
type QLateralView struct {
	span            token.Span
	LateralView     Expr
	LateralViewName *ObjectName
	LateralColAlias []*Ident
	Outer           bool
}

func (l *QLateralView) String() string {
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

// ============================================================================
// CONNECT BY
// ============================================================================

// QConnectByKind represents START WITH / CONNECT BY clause
type QConnectByKind interface {
	fmt.Stringer
}

// QConnectBy represents CONNECT BY clause
type QConnectBy struct {
	span          token.Span
	Nocycle       bool
	Relationships []Expr
}

func (c *QConnectBy) String() string {
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

// QStartWith represents START WITH clause
type QStartWith struct {
	span      token.Span
	Condition Expr
}

func (s *QStartWith) String() string {
	return fmt.Sprintf("START WITH %s", s.Condition.String())
}

// ============================================================================
// Window Definitions
// ============================================================================

// QNamedWindowDefinition represents a named window: <name> AS <window specification>
type QNamedWindowDefinition struct {
	span token.Span
	Name *Ident
	Expr QNamedWindowExpr
}

func (n *QNamedWindowDefinition) String() string {
	return fmt.Sprintf("%s AS %s", n.Name.String(), n.Expr.String())
}

// QNamedWindowExpr represents an expression in a named window declaration
type QNamedWindowExpr interface {
	fmt.Stringer
}

// QNamedWindowReference represents a direct reference to another named window
type QNamedWindowReference struct {
	Name *Ident
}

func (n *QNamedWindowReference) String() string {
	return n.Name.String()
}

// QWindowSpecExpr represents a window specification expression
type QWindowSpecExpr struct {
	Spec QWindowSpec
}

func (w *QWindowSpecExpr) String() string {
	return fmt.Sprintf("(%s)", w.Spec.String())
}

// QWindowSpec represents a window specification
type QWindowSpec struct {
	span        token.Span
	WindowName  *Ident
	PartitionBy []Expr
	OrderBy     []QOrderByExpr
	WindowFrame *QWindowFrame
}

func (w *QWindowSpec) String() string {
	var parts []string
	if w.WindowName != nil {
		parts = append(parts, w.WindowName.String())
	}
	if len(w.PartitionBy) > 0 {
		partitionBy := make([]string, len(w.PartitionBy))
		for i, p := range w.PartitionBy {
			partitionBy[i] = p.String()
		}
		parts = append(parts, fmt.Sprintf("PARTITION BY %s", strings.Join(partitionBy, ", ")))
	}
	if len(w.OrderBy) > 0 {
		orderBy := make([]string, len(w.OrderBy))
		for i, o := range w.OrderBy {
			orderBy[i] = o.String()
		}
		parts = append(parts, fmt.Sprintf("ORDER BY %s", strings.Join(orderBy, ", ")))
	}
	if w.WindowFrame != nil {
		parts = append(parts, w.WindowFrame.String())
	}
	return strings.Join(parts, " ")
}

// QWindowFrame represents a window frame specification
type QWindowFrame struct {
	Units QWindowFrameUnits
	Start QWindowFrameBound
	End   QWindowFrameBound
}

func (w *QWindowFrame) String() string {
	if w.End != nil {
		return fmt.Sprintf("%s BETWEEN %s AND %s", w.Units.String(), w.Start.String(), w.End.String())
	}
	return fmt.Sprintf("%s %s", w.Units.String(), w.Start.String())
}

// QWindowFrameUnits represents the units for a window frame
type QWindowFrameUnits int

const (
	QWindowFrameUnitsRows QWindowFrameUnits = iota
	QWindowFrameUnitsRange
	QWindowFrameUnitsGroups
)

func (w QWindowFrameUnits) String() string {
	switch w {
	case QWindowFrameUnitsRows:
		return "ROWS"
	case QWindowFrameUnitsRange:
		return "RANGE"
	case QWindowFrameUnitsGroups:
		return "GROUPS"
	default:
		return ""
	}
}

// QWindowFrameBound represents a bound in a window frame
type QWindowFrameBound interface {
	fmt.Stringer
}

// QCurrentRowBound represents CURRENT ROW
type QCurrentRowBound struct{}

func (c *QCurrentRowBound) String() string { return "CURRENT ROW" }

// QUnboundedPrecedingBound represents UNBOUNDED PRECEDING
type QUnboundedPrecedingBound struct{}

func (u *QUnboundedPrecedingBound) String() string { return "UNBOUNDED PRECEDING" }

// QUnboundedFollowingBound represents UNBOUNDED FOLLOWING
type QUnboundedFollowingBound struct{}

func (u *QUnboundedFollowingBound) String() string { return "UNBOUNDED FOLLOWING" }

// QPrecedingBound represents <expr> PRECEDING
type QPrecedingBound struct {
	Expr Expr
}

func (p *QPrecedingBound) String() string {
	return fmt.Sprintf("%s PRECEDING", p.Expr.String())
}

// QFollowingBound represents <expr> FOLLOWING
type QFollowingBound struct {
	Expr Expr
}

func (f *QFollowingBound) String() string {
	return fmt.Sprintf("%s FOLLOWING", f.Expr.String())
}

// ============================================================================
// Hints and Options
// ============================================================================

// QOptimizerHint represents query optimizer hints
type QOptimizerHint struct {
	span token.Span
	Hint string
}

func (o *QOptimizerHint) Span() token.Span { return o.span }
func (o *QOptimizerHint) String() string   { return o.Hint }

// QSetting represents a ClickHouse setting key-value pair
type QSetting struct {
	span  token.Span
	Key   *Ident
	Value Expr
}

func (s *QSetting) String() string {
	return fmt.Sprintf("%s = %s", s.Key.String(), s.Value.String())
}

// QSelectInto represents SELECT INTO clause
type QSelectInto struct {
	span      token.Span
	Temporary bool
	Unlogged  bool
	Table     bool
	Name      *ObjectName
}

func (s *QSelectInto) String() string {
	var parts []string
	parts = append(parts, "INTO")
	if s.Temporary {
		parts = append(parts, "TEMPORARY")
	}
	if s.Unlogged {
		parts = append(parts, "UNLOGGED")
	}
	if s.Table {
		parts = append(parts, "TABLE")
	}
	parts = append(parts, s.Name.String())
	return strings.Join(parts, " ")
}

// QValueTableMode represents BigQuery value table modes
type QValueTableMode int

const (
	QValueTableModeNone QValueTableMode = iota
	QValueTableModeAsStruct
	QValueTableModeAsValue
	QValueTableModeDistinctAsStruct
	QValueTableModeDistinctAsValue
)

func (v QValueTableMode) String() string {
	switch v {
	case QValueTableModeAsStruct:
		return "AS STRUCT"
	case QValueTableModeAsValue:
		return "AS VALUE"
	case QValueTableModeDistinctAsStruct:
		return "DISTINCT AS STRUCT"
	case QValueTableModeDistinctAsValue:
		return "DISTINCT AS VALUE"
	default:
		return ""
	}
}
