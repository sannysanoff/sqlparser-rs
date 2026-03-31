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

	"github.com/user/sqlparser/span"
)

// Distinct represents ALL, DISTINCT, or DISTINCT ON modifiers
type Distinct int

const (
	DistinctAll Distinct = iota
	DistinctDistinct
	DistinctOn
)

func (d Distinct) String() string {
	switch d {
	case DistinctAll:
		return "ALL"
	case DistinctDistinct:
		return "DISTINCT"
	default:
		return ""
	}
}

type DistinctWithExprs struct {
	Distinct Distinct
	Exprs    []Expr
}

func (d *DistinctWithExprs) String() string {
	switch d.Distinct {
	case DistinctAll:
		return "ALL"
	case DistinctDistinct:
		return "DISTINCT"
	case DistinctOn:
		parts := make([]string, len(d.Exprs))
		for i, e := range d.Exprs {
			parts[i] = e.String()
		}
		return fmt.Sprintf("DISTINCT ON (%s)", strings.Join(parts, ", "))
	default:
		return ""
	}
}

// Top represents MSSQL TOP clause
type Top struct {
	span     span.Span
	WithTies bool
	Percent  bool
	Quantity *TopQuantity
}

func (t *Top) String() string {
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

// TopQuantity represents quantity in TOP clause
type TopQuantity struct {
	Expr     Expr
	Constant *uint64
}

func (t *TopQuantity) String() string {
	if t.Expr != nil {
		return fmt.Sprintf("(%s)", t.Expr.String())
	}
	if t.Constant != nil {
		return fmt.Sprintf("%d", *t.Constant)
	}
	return ""
}

// Values represents an explicit VALUES clause
type Values struct {
	span         span.Span
	ExplicitRow  bool
	ValueKeyword bool
	Rows         [][]Expr
}

func (v *Values) String() string {
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

// SelectInto represents SELECT INTO clause
type SelectInto struct {
	span      span.Span
	Temporary bool
	Unlogged  bool
	Table     bool
	Name      ObjectName
}

func (s *SelectInto) String() string {
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

// TableAlias represents a table reference alias
type TableAlias struct {
	span     span.Span
	Explicit bool
	Name     Ident
	Columns  []TableAliasColumnDef
}

func (t *TableAlias) String() string {
	parts := []string{}
	if t.Explicit {
		parts = append(parts, "AS")
	}
	parts = append(parts, t.Name.String())
	if len(t.Columns) > 0 {
		cols := make([]string, len(t.Columns))
		for i, c := range t.Columns {
			cols[i] = c.String()
		}
		parts = append(parts, fmt.Sprintf("(%s)", strings.Join(cols, ", ")))
	}
	return strings.Join(parts, " ")
}

// TableAliasColumnDef represents a column definition in table alias
type TableAliasColumnDef struct {
	Name     Ident
	DataType Expr
}

func (t *TableAliasColumnDef) String() string {
	if t.DataType != nil {
		return fmt.Sprintf("%s %s", t.Name.String(), t.DataType.String())
	}
	return t.Name.String()
}

// TableVersion represents table version selection (FOR SYSTEM_TIME AS OF, etc.)
type TableVersion struct {
	Type TableVersionType
	Expr Expr
}

func (t *TableVersion) String() string {
	switch t.Type {
	case TableVersionForSystemTimeAsOf:
		return fmt.Sprintf("FOR SYSTEM_TIME AS OF %s", t.Expr.String())
	case TableVersionTimestampAsOf:
		return fmt.Sprintf("TIMESTAMP AS OF %s", t.Expr.String())
	case TableVersionVersionAsOf:
		return fmt.Sprintf("VERSION AS OF %s", t.Expr.String())
	case TableVersionFunction:
		return t.Expr.String()
	default:
		return ""
	}
}

type TableVersionType int

const (
	TableVersionForSystemTimeAsOf TableVersionType = iota
	TableVersionTimestampAsOf
	TableVersionVersionAsOf
	TableVersionFunction
	TableVersionChanges
)

type TableVersionChangesInfo struct {
	Changes Expr
	At      Expr
	End     Expr
}

type TableVersionWithInfo struct {
	Type        TableVersionType
	Expr        Expr
	ChangesInfo *TableVersionChangesInfo
}

func (t *TableVersionWithInfo) String() string {
	switch t.Type {
	case TableVersionForSystemTimeAsOf:
		return fmt.Sprintf("FOR SYSTEM_TIME AS OF %s", t.Expr.String())
	case TableVersionTimestampAsOf:
		return fmt.Sprintf("TIMESTAMP AS OF %s", t.Expr.String())
	case TableVersionVersionAsOf:
		return fmt.Sprintf("VERSION AS OF %s", t.Expr.String())
	case TableVersionFunction:
		return t.Expr.String()
	case TableVersionChanges:
		if t.ChangesInfo.End != nil {
			return fmt.Sprintf("%s %s %s", t.ChangesInfo.Changes.String(),
				t.ChangesInfo.At.String(), t.ChangesInfo.End.String())
		}
		return fmt.Sprintf("%s %s", t.ChangesInfo.Changes.String(), t.ChangesInfo.At.String())
	default:
		return ""
	}
}

// JsonPath represents a PartiQL JsonPath
type JsonPath struct {
	Path string
}

func (j *JsonPath) String() string {
	return j.Path
}

// TableSampleKind represents table sample location (before or after alias)
type TableSampleKind struct {
	BeforeTableAlias *TableSample
	AfterTableAlias  *TableSample
}

// TableSample represents a TABLESAMPLE clause
type TableSample struct {
	span     span.Span
	Modifier TableSampleModifier
	Name     *TableSampleMethod
	Quantity *TableSampleQuantity
	Seed     *TableSampleSeed
	Bucket   *TableSampleBucket
	Offset   Expr
}

func (t *TableSample) String() string {
	var parts []string
	parts = append(parts, t.Modifier.String())
	if t.Name != nil {
		parts = append(parts, t.Name.String())
	}
	if t.Quantity != nil {
		parts = append(parts, t.Quantity.String())
	}
	if t.Seed != nil {
		parts = append(parts, t.Seed.String())
	}
	if t.Bucket != nil {
		parts = append(parts, fmt.Sprintf("(%s)", t.Bucket.String()))
	}
	if t.Offset != nil {
		parts = append(parts, fmt.Sprintf("OFFSET %s", t.Offset.String()))
	}
	return strings.Join(parts, " ")
}

// TableSampleModifier represents SAMPLE or TABLESAMPLE keyword
type TableSampleModifier int

const (
	TableSampleModifierSample TableSampleModifier = iota
	TableSampleModifierTableSample
)

func (t TableSampleModifier) String() string {
	switch t {
	case TableSampleModifierSample:
		return "SAMPLE"
	case TableSampleModifierTableSample:
		return "TABLESAMPLE"
	default:
		return ""
	}
}

// TableSampleMethod represents sampling method
type TableSampleMethod int

const (
	TableSampleMethodRow TableSampleMethod = iota
	TableSampleMethodBernoulli
	TableSampleMethodSystem
	TableSampleMethodBlock
)

func (t TableSampleMethod) String() string {
	switch t {
	case TableSampleMethodRow:
		return "ROW"
	case TableSampleMethodBernoulli:
		return "BERNOULLI"
	case TableSampleMethodSystem:
		return "SYSTEM"
	case TableSampleMethodBlock:
		return "BLOCK"
	default:
		return ""
	}
}

// TableSampleQuantity represents sampling quantity
type TableSampleQuantity struct {
	Parenthesized bool
	Value         Expr
	Unit          *TableSampleUnit
}

func (t *TableSampleQuantity) String() string {
	var parts []string
	if t.Parenthesized {
		parts = append(parts, "(")
	}
	parts = append(parts, t.Value.String())
	if t.Unit != nil {
		parts = append(parts, t.Unit.String())
	}
	if t.Parenthesized {
		parts = append(parts, ")")
	}
	return strings.Join(parts, "")
}

// TableSampleUnit represents rows or percent
type TableSampleUnit int

const (
	TableSampleUnitRows TableSampleUnit = iota
	TableSampleUnitPercent
)

func (t TableSampleUnit) String() string {
	switch t {
	case TableSampleUnitRows:
		return "ROWS"
	case TableSampleUnitPercent:
		return "PERCENT"
	default:
		return ""
	}
}

// TableSampleSeed represents SEED or REPEATABLE clause
type TableSampleSeed struct {
	Modifier TableSampleSeedModifier
	Value    ValueWithSpan
}

func (t *TableSampleSeed) String() string {
	return fmt.Sprintf("%s (%s)", t.Modifier.String(), t.Value.String())
}

// TableSampleSeedModifier represents REPEATABLE or SEED
type TableSampleSeedModifier int

const (
	TableSampleSeedModifierRepeatable TableSampleSeedModifier = iota
	TableSampleSeedModifierSeed
)

func (t TableSampleSeedModifier) String() string {
	switch t {
	case TableSampleSeedModifierRepeatable:
		return "REPEATABLE"
	case TableSampleSeedModifierSeed:
		return "SEED"
	default:
		return ""
	}
}

// TableSampleBucket represents BUCKET clause
type TableSampleBucket struct {
	Bucket ValueWithSpan
	Total  ValueWithSpan
	On     Expr
}

func (t *TableSampleBucket) String() string {
	parts := []string{
		fmt.Sprintf("BUCKET %s OUT OF %s", t.Bucket.String(), t.Total.String()),
	}
	if t.On != nil {
		parts = append(parts, fmt.Sprintf("ON %s", t.On.String()))
	}
	return strings.Join(parts, " ")
}

// ValueWithSpan represents a value with its span
type ValueWithSpan struct {
	Value string
	span  span.Span
}

func (v *ValueWithSpan) String() string {
	return v.Value
}

// TableFunctionArgs represents arguments to a table-valued function
type TableFunctionArgs struct {
	Args     []FunctionArg
	Settings *[]Setting
}

func (t *TableFunctionArgs) String() string {
	parts := make([]string, len(t.Args))
	for i, a := range t.Args {
		parts[i] = a.String()
	}
	result := strings.Join(parts, ", ")
	if t.Settings != nil && len(*t.Settings) > 0 {
		settings := make([]string, len(*t.Settings))
		for i, s := range *t.Settings {
			settings[i] = s.String()
		}
		result += ", SETTINGS " + strings.Join(settings, ", ")
	}
	return result
}

// TableIndexHints represents MySQL-style index hints
type TableIndexHints struct {
	HintType   TableIndexHintType
	IndexType  TableIndexType
	ForClause  *TableIndexHintForClause
	IndexNames []Ident
}

func (t *TableIndexHints) String() string {
	parts := []string{t.HintType.String(), t.IndexType.String()}
	if t.ForClause != nil {
		parts = append(parts, "FOR", t.ForClause.String())
	}
	names := make([]string, len(t.IndexNames))
	for i, n := range t.IndexNames {
		names[i] = n.String()
	}
	parts = append(parts, fmt.Sprintf("(%s)", strings.Join(names, ", ")))
	return strings.Join(parts, " ")
}

// TableIndexHintType represents USE, IGNORE, or FORCE
type TableIndexHintType int

const (
	TableIndexHintTypeUse TableIndexHintType = iota
	TableIndexHintTypeIgnore
	TableIndexHintTypeForce
)

func (t TableIndexHintType) String() string {
	switch t {
	case TableIndexHintTypeUse:
		return "USE"
	case TableIndexHintTypeIgnore:
		return "IGNORE"
	case TableIndexHintTypeForce:
		return "FORCE"
	default:
		return ""
	}
}

// TableIndexType represents INDEX or KEY
type TableIndexType int

const (
	TableIndexTypeIndex TableIndexType = iota
	TableIndexTypeKey
)

func (t TableIndexType) String() string {
	switch t {
	case TableIndexTypeIndex:
		return "INDEX"
	case TableIndexTypeKey:
		return "KEY"
	default:
		return ""
	}
}

// TableIndexHintForClause represents which clause the hint applies to
type TableIndexHintForClause int

const (
	TableIndexHintForClauseJoin TableIndexHintForClause = iota
	TableIndexHintForClauseOrderBy
	TableIndexHintForClauseGroupBy
)

func (t TableIndexHintForClause) String() string {
	switch t {
	case TableIndexHintForClauseJoin:
		return "JOIN"
	case TableIndexHintForClauseOrderBy:
		return "ORDER BY"
	case TableIndexHintForClauseGroupBy:
		return "GROUP BY"
	default:
		return ""
	}
}

// UpdateTableFromKind represents the FROM clause in UPDATE statement
type UpdateTableFromKind struct {
	BeforeSet *[]TableWithJoins
	AfterSet  *[]TableWithJoins
}

func (u *UpdateTableFromKind) String() string {
	if u.BeforeSet != nil {
		parts := make([]string, len(*u.BeforeSet))
		for i, t := range *u.BeforeSet {
			parts[i] = t.String()
		}
		return strings.Join(parts, ", ")
	}
	if u.AfterSet != nil {
		parts := make([]string, len(*u.AfterSet))
		for i, t := range *u.AfterSet {
			parts[i] = t.String()
		}
		return strings.Join(parts, ", ")
	}
	return ""
}

// Statement interface for SQL statements
type Statement interface {
	fmt.Stringer
}

// ValueTableMode represents BigQuery value table modes
type ValueTableMode int

const (
	ValueTableModeNone ValueTableMode = iota
	ValueTableModeAsStruct
	ValueTableModeAsValue
	ValueTableModeDistinctAsStruct
	ValueTableModeDistinctAsValue
)

func (v ValueTableMode) String() string {
	switch v {
	case ValueTableModeAsStruct:
		return "AS STRUCT"
	case ValueTableModeAsValue:
		return "AS VALUE"
	case ValueTableModeDistinctAsStruct:
		return "DISTINCT AS STRUCT"
	case ValueTableModeDistinctAsValue:
		return "DISTINCT AS VALUE"
	default:
		return ""
	}
}

// ExprWithAlias represents an expression with optional alias
type ExprWithAlias struct {
	Expr  Expr
	Alias *Ident
}

func (e *ExprWithAlias) String() string {
	if e.Alias != nil {
		return fmt.Sprintf("%s AS %s", e.Expr.String(), e.Alias.String())
	}
	return e.Expr.String()
}

// Assignment represents column = expr assignment (for pipe SET operator)
type Assignment struct {
	Column Ident
	Value  Expr
}

func (a *Assignment) String() string {
	return fmt.Sprintf("%s = %s", a.Column.String(), a.Value.String())
}

// PipeOperator represents BigQuery pipe syntax operators
type PipeOperator struct {
	span span.Span
	Type PipeOperatorType
}

func (p *PipeOperator) String() string {
	if p.Type != nil {
		return p.Type.String()
	}
	return ""
}

type PipeOperatorType interface {
	fmt.Stringer
}

// PipeSelect represents |> SELECT
type PipeSelect struct {
	Exprs []SelectItem
}

func (p *PipeSelect) String() string {
	exprs := make([]string, len(p.Exprs))
	for i, e := range p.Exprs {
		exprs[i] = e.String()
	}
	return fmt.Sprintf("SELECT %s", strings.Join(exprs, ", "))
}

// PipeExtend represents |> EXTEND
type PipeExtend struct {
	Exprs []SelectItem
}

func (p *PipeExtend) String() string {
	exprs := make([]string, len(p.Exprs))
	for i, e := range p.Exprs {
		exprs[i] = e.String()
	}
	return fmt.Sprintf("EXTEND %s", strings.Join(exprs, ", "))
}

// PipeSet represents |> SET
type PipeSet struct {
	Assignments []Assignment
}

func (p *PipeSet) String() string {
	assigns := make([]string, len(p.Assignments))
	for i, a := range p.Assignments {
		assigns[i] = a.String()
	}
	return fmt.Sprintf("SET %s", strings.Join(assigns, ", "))
}

// PipeDrop represents |> DROP
type PipeDrop struct {
	Columns []Ident
}

func (p *PipeDrop) String() string {
	cols := make([]string, len(p.Columns))
	for i, c := range p.Columns {
		cols[i] = c.String()
	}
	return fmt.Sprintf("DROP %s", strings.Join(cols, ", "))
}

// PipeAs represents |> AS
type PipeAs struct {
	Alias Ident
}

func (p *PipeAs) String() string {
	return fmt.Sprintf("AS %s", p.Alias.String())
}

// PipeLimit represents |> LIMIT
type PipeLimit struct {
	Expr   Expr
	Offset Expr
}

func (p *PipeLimit) String() string {
	if p.Offset != nil {
		return fmt.Sprintf("LIMIT %s OFFSET %s", p.Expr.String(), p.Offset.String())
	}
	return fmt.Sprintf("LIMIT %s", p.Expr.String())
}

// PipeAggregate represents |> AGGREGATE
type PipeAggregate struct {
	FullTableExprs []ExprWithAliasAndOrderBy
	GroupByExpr    []ExprWithAliasAndOrderBy
}

func (p *PipeAggregate) String() string {
	parts := []string{"AGGREGATE"}
	if len(p.FullTableExprs) > 0 {
		exprs := make([]string, len(p.FullTableExprs))
		for i, e := range p.FullTableExprs {
			exprs[i] = e.String()
		}
		parts = append(parts, strings.Join(exprs, ", "))
	}
	if len(p.GroupByExpr) > 0 {
		groupBy := make([]string, len(p.GroupByExpr))
		for i, e := range p.GroupByExpr {
			groupBy[i] = e.String()
		}
		parts = append(parts, "GROUP BY", strings.Join(groupBy, ", "))
	}
	return strings.Join(parts, " ")
}

// PipeWhere represents |> WHERE
type PipeWhere struct {
	Expr Expr
}

func (p *PipeWhere) String() string {
	return fmt.Sprintf("WHERE %s", p.Expr.String())
}

// PipeOrderBy represents |> ORDER BY
type PipeOrderBy struct {
	Exprs []OrderByExpr
}

func (p *PipeOrderBy) String() string {
	exprs := make([]string, len(p.Exprs))
	for i, e := range p.Exprs {
		exprs[i] = e.String()
	}
	return fmt.Sprintf("ORDER BY %s", strings.Join(exprs, ", "))
}

// PipeTableSample represents |> TABLESAMPLE
type PipeTableSample struct {
	Sample *TableSample
}

func (p *PipeTableSample) String() string {
	return p.Sample.String()
}

// PipeRename represents |> RENAME
type PipeRename struct {
	Mappings []IdentWithAlias
}

func (p *PipeRename) String() string {
	mappings := make([]string, len(p.Mappings))
	for i, m := range p.Mappings {
		mappings[i] = m.String()
	}
	return fmt.Sprintf("RENAME %s", strings.Join(mappings, ", "))
}

// PipeUnion represents |> UNION
type PipeUnion struct {
	SetQuantifier SetQuantifier
	Queries       []*Query
}

func (p *PipeUnion) String() string {
	quantifier := ""
	if p.SetQuantifier != SetQuantifierNone {
		quantifier = " " + p.SetQuantifier.String()
	}
	queries := make([]string, len(p.Queries))
	for i, q := range p.Queries {
		queries[i] = fmt.Sprintf("(%s)", q.String())
	}
	return fmt.Sprintf("UNION%s %s", quantifier, strings.Join(queries, ", "))
}

// PipeIntersect represents |> INTERSECT
type PipeIntersect struct {
	SetQuantifier SetQuantifier
	Queries       []*Query
}

func (p *PipeIntersect) String() string {
	quantifier := ""
	if p.SetQuantifier != SetQuantifierNone {
		quantifier = " " + p.SetQuantifier.String()
	}
	queries := make([]string, len(p.Queries))
	for i, q := range p.Queries {
		queries[i] = fmt.Sprintf("(%s)", q.String())
	}
	return fmt.Sprintf("INTERSECT%s %s", quantifier, strings.Join(queries, ", "))
}

// PipeExcept represents |> EXCEPT
type PipeExcept struct {
	SetQuantifier SetQuantifier
	Queries       []*Query
}

func (p *PipeExcept) String() string {
	quantifier := ""
	if p.SetQuantifier != SetQuantifierNone {
		quantifier = " " + p.SetQuantifier.String()
	}
	queries := make([]string, len(p.Queries))
	for i, q := range p.Queries {
		queries[i] = fmt.Sprintf("(%s)", q.String())
	}
	return fmt.Sprintf("EXCEPT%s %s", quantifier, strings.Join(queries, ", "))
}

// PipeCall represents |> CALL
type PipeCall struct {
	Function Statement
	Alias    *Ident
}

func (p *PipeCall) String() string {
	result := fmt.Sprintf("CALL %s", p.Function.String())
	if p.Alias != nil {
		result += fmt.Sprintf(" AS %s", p.Alias.String())
	}
	return result
}

// PipePivot represents |> PIVOT
type PipePivot struct {
	AggregateFunctions []ExprWithAlias
	ValueColumn        []Ident
	ValueSource        PivotValueSource
	Alias              *Ident
}

func (p *PipePivot) String() string {
	aggs := make([]string, len(p.AggregateFunctions))
	for i, a := range p.AggregateFunctions {
		aggs[i] = a.String()
	}
	valCols := make([]string, len(p.ValueColumn))
	for i, v := range p.ValueColumn {
		valCols[i] = v.String()
	}
	var valColStr string
	if len(valCols) == 1 {
		valColStr = valCols[0]
	} else {
		valColStr = "(" + strings.Join(valCols, ", ") + ")"
	}
	result := fmt.Sprintf("PIVOT(%s FOR %s IN (%s))",
		strings.Join(aggs, ", "), valColStr, p.ValueSource.String())
	if p.Alias != nil {
		result += fmt.Sprintf(" AS %s", p.Alias.String())
	}
	return result
}

// PipeUnpivot represents |> UNPIVOT
type PipeUnpivot struct {
	ValueColumn    Ident
	NameColumn     Ident
	UnpivotColumns []Ident
	Alias          *Ident
}

func (p *PipeUnpivot) String() string {
	cols := make([]string, len(p.UnpivotColumns))
	for i, c := range p.UnpivotColumns {
		cols[i] = c.String()
	}
	result := fmt.Sprintf("UNPIVOT(%s FOR %s IN (%s))",
		p.ValueColumn.String(), p.NameColumn.String(), strings.Join(cols, ", "))
	if p.Alias != nil {
		result += fmt.Sprintf(" AS %s", p.Alias.String())
	}
	return result
}

// PipeJoin represents |> JOIN
type PipeJoin struct {
	Join *Join
}

func (p *PipeJoin) String() string {
	return p.Join.String()
}

// JsonTableColumn represents a column in JSON_TABLE
type JsonTableColumn interface {
	fmt.Stringer
}

// JsonTableNamedColumn represents a named column in JSON_TABLE
type JsonTableNamedColumn struct {
	span    span.Span
	Name    Ident
	Type    Expr
	Path    ValueWithSpan
	Exists  bool
	OnEmpty *JsonTableColumnErrorHandling
	OnError *JsonTableColumnErrorHandling
}

func (j *JsonTableNamedColumn) String() string {
	parts := []string{
		j.Name.String(),
		j.Type.String(),
	}
	if j.Exists {
		parts = append(parts, "EXISTS")
	}
	parts = append(parts, "PATH", j.Path.String())
	if j.OnEmpty != nil {
		parts = append(parts, j.OnEmpty.String(), "ON EMPTY")
	}
	if j.OnError != nil {
		parts = append(parts, j.OnError.String(), "ON ERROR")
	}
	return strings.Join(parts, " ")
}

// JsonTableForOrdinality represents FOR ORDINALITY column in JSON_TABLE
type JsonTableForOrdinality struct {
	span span.Span
	Name Ident
}

func (j *JsonTableForOrdinality) String() string {
	return fmt.Sprintf("%s FOR ORDINALITY", j.Name.String())
}

// JsonTableNestedColumn represents a nested column in JSON_TABLE
type JsonTableNestedColumn struct {
	span    span.Span
	Path    ValueWithSpan
	Columns []JsonTableColumn
}

func (j *JsonTableNestedColumn) String() string {
	cols := make([]string, len(j.Columns))
	for i, c := range j.Columns {
		cols[i] = c.String()
	}
	return fmt.Sprintf("NESTED PATH %s COLUMNS (%s)", j.Path.String(), strings.Join(cols, ", "))
}

// JsonTableColumnErrorHandling represents error handling for JSON_TABLE columns
type JsonTableColumnErrorHandling int

const (
	JsonTableColumnErrorHandlingNull JsonTableColumnErrorHandling = iota
	JsonTableColumnErrorHandlingDefault
	JsonTableColumnErrorHandlingError
)

func (j JsonTableColumnErrorHandling) String() string {
	switch j {
	case JsonTableColumnErrorHandlingNull:
		return "NULL"
	case JsonTableColumnErrorHandlingDefault:
		return "DEFAULT"
	case JsonTableColumnErrorHandlingError:
		return "ERROR"
	default:
		return ""
	}
}

type JsonTableColumnErrorHandlingWithValue struct {
	Type  JsonTableColumnErrorHandling
	Value *ValueWithSpan
}

func (j *JsonTableColumnErrorHandlingWithValue) String() string {
	switch j.Type {
	case JsonTableColumnErrorHandlingNull:
		return "NULL"
	case JsonTableColumnErrorHandlingDefault:
		if j.Value != nil {
			return fmt.Sprintf("DEFAULT %s", j.Value.String())
		}
		return "DEFAULT"
	case JsonTableColumnErrorHandlingError:
		return "ERROR"
	default:
		return ""
	}
}

// OpenJsonTableColumn represents a column in OPENJSON
type OpenJsonTableColumn struct {
	span   span.Span
	Name   Ident
	Type   Expr
	Path   *string
	AsJson bool
}

func (o *OpenJsonTableColumn) String() string {
	parts := []string{o.Name.String(), o.Type.String()}
	if o.Path != nil {
		parts = append(parts, fmt.Sprintf("'%s'", escapeSingleQuote(*o.Path)))
	}
	if o.AsJson {
		parts = append(parts, "AS JSON")
	}
	return strings.Join(parts, " ")
}

// XmlTableColumn represents a column in XMLTABLE
type XmlTableColumn struct {
	span   span.Span
	Name   Ident
	Option XmlTableColumnOption
}

func (x *XmlTableColumn) String() string {
	parts := []string{x.Name.String()}
	parts = append(parts, x.Option.String())
	return strings.Join(parts, " ")
}

// XmlTableColumnOption represents options for XMLTABLE columns
type XmlTableColumnOption interface {
	fmt.Stringer
}

// XmlTableColumnNamedInfo represents a named column with type and path
type XmlTableColumnNamedInfo struct {
	Type     Expr
	Path     Expr
	Default  Expr
	Nullable bool
}

func (x *XmlTableColumnNamedInfo) String() string {
	parts := []string{x.Type.String()}
	if x.Path != nil {
		parts = append(parts, fmt.Sprintf("PATH %s", x.Path.String()))
	}
	if x.Default != nil {
		parts = append(parts, fmt.Sprintf("DEFAULT %s", x.Default.String()))
	}
	if !x.Nullable {
		parts = append(parts, "NOT NULL")
	}
	return strings.Join(parts, " ")
}

// XmlTableColumnForOrdinality represents FOR ORDINALITY marker
type XmlTableColumnForOrdinality struct{}

func (x *XmlTableColumnForOrdinality) String() string {
	return "FOR ORDINALITY"
}

// XmlPassingClause represents the PASSING clause for XMLTABLE
type XmlPassingClause struct {
	Arguments []XmlPassingArgument
}

func (x *XmlPassingClause) String() string {
	if len(x.Arguments) > 0 {
		args := make([]string, len(x.Arguments))
		for i, a := range x.Arguments {
			args[i] = a.String()
		}
		return fmt.Sprintf(" PASSING %s", strings.Join(args, ", "))
	}
	return ""
}

// XmlPassingArgument represents an argument in the PASSING clause
type XmlPassingArgument struct {
	Expr    Expr
	Alias   *Ident
	ByValue bool
}

func (x *XmlPassingArgument) String() string {
	parts := []string{}
	if x.ByValue {
		parts = append(parts, "BY VALUE")
	}
	parts = append(parts, x.Expr.String())
	if x.Alias != nil {
		parts = append(parts, "AS", x.Alias.String())
	}
	return strings.Join(parts, " ")
}

// XmlNamespaceDefinition represents a namespace definition in XMLNAMESPACES
type XmlNamespaceDefinition struct {
	span span.Span
	URI  Expr
	Name Ident
}

func (x *XmlNamespaceDefinition) String() string {
	return fmt.Sprintf("%s AS %s", x.URI.String(), x.Name.String())
}
