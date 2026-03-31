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

package ast

import (
	"fmt"
	"strings"
)

// TableFactor represents a single table reference in a FROM clause.
// It can be a simple table name, a derived table (subquery), or a table-producing function.
type TableFactor struct {
	BaseNode
	// Kind indicates the type of table factor.
	Kind TableFactorKind
	// Data holds the specific table factor data.
	Data TableFactorData
}

// TableFactorKind enumerates the types of table factors.
type TableFactorKind int

const (
	// TableFactorTable is a simple table reference (name with optional alias).
	TableFactorTable TableFactorKind = iota
	// TableFactorDerived is a derived table (subquery).
	TableFactorDerived
	// TableFactorFunction is a table-producing function.
	TableFactorFunction
	// TableFactorUNNEST is an UNNEST expression (BigQuery, etc.).
	TableFactorUNNEST
	// TableFactorJSONTable is a JSON_TABLE expression.
	TableFactorJSONTable
	// TableFactorPivot is a PIVOT operation.
	TableFactorPivot
	// TableFactorUnpivot is an UNPIVOT operation.
	TableFactorUnpivot
	// TableFactorTableSample is a table with TABLESAMPLE clause.
	TableFactorTableSample
	// TableFactorTableVersion is a table with VERSION clause (AS OF SYSTEM TIME).
	TableFactorTableVersion
)

// TableFactorData holds the specific data for each table factor kind.
type TableFactorData struct {
	// For TableFactorTable.
	Table *TableRef
	// For TableFactorDerived.
	Derived *DerivedTableRef
	// For TableFactorFunction.
	Function *TableFunctionRef
	// For TableFactorUNNEST.
	UNNEST *UNNESTTableRef
}

// TableRef represents a simple table reference.
type TableRef struct {
	// Name is the table name (may be qualified).
	Name *ObjectName
	// Alias is the optional table alias.
	Alias *TableAlias
	// Hints are optional index hints (MySQL, etc.).
	Hints []TableIndexHint
	// Temporal is an optional temporal clause (AS OF SYSTEM TIME, etc.).
	Temporal *TableTemporal
}

// TableAlias represents a table alias with optional column aliases.
type TableAlias struct {
	// Name is the alias name.
	Name *Ident
	// Columns are optional column aliases.
	Columns []*Ident
}

// String returns the SQL representation.
func (a *TableAlias) String() string {
	if a == nil || a.Name == nil {
		return ""
	}
	result := " AS " + a.Name.String()
	if len(a.Columns) > 0 {
		cols := make([]string, len(a.Columns))
		for i, col := range a.Columns {
			cols[i] = col.String()
		}
		result += fmt.Sprintf("(%s)", strings.Join(cols, ", "))
	}
	return result
}

// DerivedTableRef represents a derived table (subquery in FROM).
type DerivedTableRef struct {
	// Subquery is the SELECT subquery.
	Subquery *SelectQuery
	// Alias is the required alias for the subquery.
	Alias *TableAlias
	// Lateral indicates if this is a LATERAL subquery.
	Lateral bool
}

// TableFunctionRef represents a table-producing function call.
type TableFunctionRef struct {
	// Function is the table function expression.
	Function *Function
	// Alias is the optional alias.
	Alias *TableAlias
}

// UNNESTTableRef represents an UNNEST table reference.
type UNNESTTableRef struct {
	// ArrayExpr is the array expression to unnest.
	ArrayExpr Expr
	// Alias is the optional alias.
	Alias *TableAlias
	// WithOffset indicates if WITH OFFSET is used.
	WithOffset bool
	// WithOffsetAlias is the alias for the offset column.
	WithOffsetAlias *Ident
}

// TableTemporal represents a temporal table clause.
type TableTemporal struct {
	// Kind is the type of temporal clause.
	Kind TemporalKind
	// Expression is the temporal expression.
	Expression Expr
}

// TemporalKind enumerates temporal clause types.
type TemporalKind int

const (
	// TemporalAsOfSystemTime is AS OF SYSTEM TIME (BigQuery, etc.).
	TemporalAsOfSystemTime TemporalKind = iota
	// TemporalAsOf is AS OF (SQL:2011).
	TemporalAsOf
	// TemporalFromTo is FROM ... TO ... (SQL:2011).
	TemporalFromTo
	// TemporalBetween is BETWEEN ... AND ... (SQL:2011).
	TemporalBetween
	// TemporalVersion is VERSION ... (SQL:2011).
	TemporalVersion
	// TemporalPeriodFor is PERIOD FOR (SQL:2011).
	TemporalPeriodFor
)

// TableIndexHint represents a table index hint (MySQL, etc.).
type TableIndexHint struct {
	// Kind is the type of hint.
	Kind IndexHintKind
	// IndexNames are the indexes mentioned.
	IndexNames []*Ident
	// ForClause specifies which operation the hint applies to.
	ForClause IndexHintForClause
}

// IndexHintKind enumerates index hint types.
type IndexHintKind int

const (
	IndexHintUse IndexHintKind = iota
	IndexHintIgnore
	IndexHintForce
)

// IndexHintForClause enumerates the operations index hints can apply to.
type IndexHintForClause int

const (
	IndexHintForAll IndexHintForClause = iota
	IndexHintForJoin
	IndexHintForOrderBy
	IndexHintForGroupBy
)

// String returns the SQL representation of the table factor.
func (t *TableFactor) String() string {
	switch t.Kind {
	case TableFactorTable:
		if t.Data.Table == nil {
			return ""
		}
		result := t.Data.Table.Name.String()
		if t.Data.Table.Alias != nil {
			result += t.Data.Table.Alias.String()
		}
		return result

	case TableFactorDerived:
		if t.Data.Derived == nil || t.Data.Derived.Subquery == nil {
			return ""
		}
		result := "("
		if t.Data.Derived.Lateral {
			result = "LATERAL "
		}
		result += t.Data.Derived.Subquery.String()
		result += ")"
		if t.Data.Derived.Alias != nil {
			result += t.Data.Derived.Alias.String()
		}
		return result

	case TableFactorFunction:
		if t.Data.Function == nil || t.Data.Function.Function == nil {
			return ""
		}
		result := t.Data.Function.Function.String()
		if t.Data.Function.Alias != nil {
			result += t.Data.Function.Alias.String()
		}
		return result

	case TableFactorUNNEST:
		if t.Data.UNNEST == nil || t.Data.UNNEST.ArrayExpr == nil {
			return ""
		}
		result := fmt.Sprintf("UNNEST(%s)", t.Data.UNNEST.ArrayExpr.String())
		if t.Data.UNNEST.Alias != nil {
			result += t.Data.UNNEST.Alias.String()
		}
		if t.Data.UNNEST.WithOffset {
			result += " WITH OFFSET"
			if t.Data.UNNEST.WithOffsetAlias != nil {
				result += " AS " + t.Data.UNNEST.WithOffsetAlias.String()
			}
		}
		return result

	default:
		return ""
	}
}

// Join represents a JOIN clause in a FROM clause.
type Join struct {
	BaseNode
	// Table is the table being joined (right side).
	Table TableFactor
	// Operator is the type of join.
	Operator JoinOperator
	// Constraint is the join condition (ON, USING, or none).
	Constraint JoinConstraint
}

// JoinOperator enumerates the types of join operations.
type JoinOperator int

const (
	// JoinOperatorInner is INNER JOIN.
	JoinOperatorInner JoinOperator = iota
	// JoinOperatorLeft is LEFT [OUTER] JOIN.
	JoinOperatorLeft
	// JoinOperatorRight is RIGHT [OUTER] JOIN.
	JoinOperatorRight
	// JoinOperatorFull is FULL [OUTER] JOIN.
	JoinOperatorFull
	// JoinOperatorCross is CROSS JOIN.
	JoinOperatorCross
	// JoinOperatorStraight is STRAIGHT_JOIN (MySQL).
	JoinOperatorStraight
	// JoinOperatorSemi is SEMI JOIN.
	JoinOperatorSemi
	// JoinOperatorAnti is ANTI JOIN.
	JoinOperatorAnti
)

// String returns the SQL representation of the join operator.
func (j JoinOperator) String() string {
	switch j {
	case JoinOperatorInner:
		return "JOIN"
	case JoinOperatorLeft:
		return "LEFT JOIN"
	case JoinOperatorRight:
		return "RIGHT JOIN"
	case JoinOperatorFull:
		return "FULL JOIN"
	case JoinOperatorCross:
		return "CROSS JOIN"
	case JoinOperatorStraight:
		return "STRAIGHT_JOIN"
	case JoinOperatorSemi:
		return "SEMI JOIN"
	case JoinOperatorAnti:
		return "ANTI JOIN"
	default:
		return "JOIN"
	}
}

// JoinConstraint represents the constraint for a join (ON, USING, or none).
type JoinConstraint struct {
	// Kind is the type of constraint.
	Kind JoinConstraintKind
	// On is the ON clause expression (for Kind == JoinConstraintOn).
	On Expr
	// Using is the list of columns for USING clause.
	Using []*Ident
}

// JoinConstraintKind enumerates join constraint types.
type JoinConstraintKind int

const (
	// JoinConstraintNone means no constraint (e.g., CROSS JOIN).
	JoinConstraintNone JoinConstraintKind = iota
	// JoinConstraintOn means ON expression.
	JoinConstraintOn
	// JoinConstraintUsing means USING (columns).
	JoinConstraintUsing
	// JoinConstraintNatural means NATURAL (no explicit constraint).
	JoinConstraintNatural
)

// String returns the SQL representation of the join.
func (j *Join) String() string {
	result := j.Operator.String() + " "
	result += j.Table.String()

	switch j.Constraint.Kind {
	case JoinConstraintOn:
		if j.Constraint.On != nil {
			result += fmt.Sprintf(" ON %s", j.Constraint.On.String())
		}
	case JoinConstraintUsing:
		if len(j.Constraint.Using) > 0 {
			cols := make([]string, len(j.Constraint.Using))
			for i, col := range j.Constraint.Using {
				cols[i] = col.String()
			}
			result += fmt.Sprintf(" USING (%s)", strings.Join(cols, ", "))
		}
	case JoinConstraintNatural:
		result = "NATURAL " + result
	}

	return result
}

// OrderBy represents an ORDER BY clause.
type OrderBy struct {
	BaseNode
	// Expressions are the ORDER BY expressions.
	Expressions []*OrderByExpr
	// Interpolate is an optional interpolation clause (ClickHouse).
	Interpolate *Interpolate
}

// Interpolate represents an INTERPOLATE clause for ORDER BY (ClickHouse).
type Interpolate struct {
	// Fields to interpolate.
	Fields []InterpolateField
}

// InterpolateField represents a single interpolation field.
type InterpolateField struct {
	// Column is the column to interpolate.
	Column *Ident
	// Expression is the interpolation expression.
	Expression Expr
}

// String returns the SQL representation of ORDER BY.
func (o *OrderBy) String() string {
	if o == nil || len(o.Expressions) == 0 {
		return ""
	}

	parts := make([]string, len(o.Expressions))
	for i, expr := range o.Expressions {
		parts[i] = expr.String()
	}

	result := "ORDER BY " + strings.Join(parts, ", ")

	if o.Interpolate != nil && len(o.Interpolate.Fields) > 0 {
		result += " INTERPOLATE ("
		fields := make([]string, len(o.Interpolate.Fields))
		for i, field := range o.Interpolate.Fields {
			if field.Column != nil && field.Expression != nil {
				fields[i] = fmt.Sprintf("%s AS %s", field.Column.String(), field.Expression.String())
			}
		}
		result += strings.Join(fields, ", ") + ")"
	}

	return result
}

// LimitClause represents a LIMIT clause.
type LimitClause struct {
	BaseNode
	// Limit is the maximum number of rows to return.
	Limit Expr
	// Offset is the number of rows to skip (optional).
	Offset Expr
	// OffsetComma indicates if the syntax is LIMIT <offset>, <limit> (MySQL style).
	OffsetComma bool
}

// String returns the SQL representation of LIMIT.
func (l *LimitClause) String() string {
	if l == nil {
		return ""
	}

	if l.OffsetComma {
		// MySQL style: LIMIT <offset>, <limit>
		if l.Offset != nil && l.Limit != nil {
			return fmt.Sprintf("LIMIT %s, %s", l.Offset.String(), l.Limit.String())
		}
	}

	result := ""
	if l.Limit != nil {
		result = fmt.Sprintf("LIMIT %s", l.Limit.String())
	}

	if l.Offset != nil {
		if result != "" {
			result += " "
		}
		result += fmt.Sprintf("OFFSET %s", l.Offset.String())
	}

	return result
}

// Offset represents an OFFSET clause with optional row handling.
type Offset struct {
	BaseNode
	// Value is the offset expression.
	Value Expr
	// Rows indicates if ROW or ROWS was specified.
	Rows *OffsetRows
}

// OffsetRows indicates the row handling for OFFSET.
type OffsetRows int

const (
	OffsetRowsUnspecified OffsetRows = iota
	OffsetRowsRow
	OffsetRowsRows
)

// String returns the SQL representation of OFFSET.
func (o *Offset) String() string {
	if o == nil || o.Value == nil {
		return ""
	}

	result := fmt.Sprintf("OFFSET %s", o.Value.String())

	if o.Rows != nil {
		switch *o.Rows {
		case OffsetRowsRow:
			result += " ROW"
		case OffsetRowsRows:
			result += " ROWS"
		}
	}

	return result
}

// Fetch represents a FETCH clause (standard SQL alternative to LIMIT).
type Fetch struct {
	BaseNode
	// Kind is FIRST or NEXT.
	Kind FetchKind
	// Quantity is the number of rows.
	Quantity FetchQuantity
	// Percent indicates if it's a percentage.
	Percent bool
	// WithTies indicates if WITH TIES is used.
	WithTies bool
}

// FetchKind enumerates FIRST/NEXT options.
type FetchKind int

const (
	FetchFirst FetchKind = iota
	FetchNext
)

// FetchQuantity represents the quantity specification.
type FetchQuantity struct {
	// Value is the numeric value.
	Value *Value
	// All indicates FETCH ALL ROWS.
	All bool
}

// String returns the SQL representation of FETCH.
func (f *Fetch) String() string {
	if f == nil {
		return ""
	}

	kind := "FIRST"
	if f.Kind == FetchNext {
		kind = "NEXT"
	}

	var quantity string
	if f.Quantity.All {
		quantity = "ALL"
	} else if f.Quantity.Value != nil {
		quantity = f.Quantity.Value.String()
	}

	result := fmt.Sprintf("FETCH %s %s", kind, quantity)

	if f.Percent {
		result += " PERCENT"
	}

	result += " ROWS"

	if f.WithTies {
		result += " WITH TIES"
	} else {
		result += " ONLY"
	}

	return result
}

// With represents a WITH clause (Common Table Expressions).
type With struct {
	BaseNode
	// Recursive indicates if RECURSIVE is specified.
	Recursive bool
	// CTEs are the common table expressions.
	CTEs []*CTE
}

// String returns the SQL representation of WITH.
func (w *With) String() string {
	if w == nil || len(w.CTEs) == 0 {
		return ""
	}

	result := "WITH"
	if w.Recursive {
		result += " RECURSIVE"
	}

	ctes := make([]string, len(w.CTEs))
	for i, cte := range w.CTEs {
		ctes[i] = cte.String()
	}
	result += " " + strings.Join(ctes, ", ")

	return result
}

// CTE represents a Common Table Expression.
type CTE struct {
	BaseNode
	// Alias is the CTE name with optional column aliases.
	Alias *TableAlias
	// Query is the CTE query.
	Query *SelectQuery
	// From is used for some dialects (materialized, etc.).
	From *CTEAsMaterialized
}

// CTEAsMaterialized represents materialization options for CTEs.
type CTEAsMaterialized int

const (
	CTEAsNotSpecified CTEAsMaterialized = iota
	CTEAsMaterializedOption
	CTEAsNotMaterializedOption
)

// String returns the SQL representation of a CTE.
func (c *CTE) String() string {
	if c == nil || c.Alias == nil || c.Alias.Name == nil || c.Query == nil {
		return ""
	}

	result := c.Alias.Name.String()

	if len(c.Alias.Columns) > 0 {
		cols := make([]string, len(c.Alias.Columns))
		for i, col := range c.Alias.Columns {
			cols[i] = col.String()
		}
		result += fmt.Sprintf("(%s)", strings.Join(cols, ", "))
	}

	if c.From != nil {
		switch *c.From {
		case CTEAsMaterializedOption:
			result += " AS MATERIALIZED"
		case CTEAsNotMaterializedOption:
			result += " AS NOT MATERIALIZED"
		}
	}

	result += fmt.Sprintf(" AS (%s)", c.Query.String())

	return result
}

// SelectQuery represents a complete query expression.
// This is a concrete implementation of the Query interface.
type SelectQuery struct {
	QueryBase
	// WITH clause.
	WithClause *With
	// Body is the main query body (SELECT, UNION, etc.).
	Body SetExpr
	// ORDER BY clause.
	OrderBy *OrderBy
	// LIMIT clause.
	Limit *LimitClause
	// OFFSET clause (separate from limit for some dialects).
	Offset *Offset
	// FETCH clause.
	FetchClause *Fetch
	// Locks are row-level locking clauses.
	Locks []LockClause
}

// SetExpr represents the main body of a query (SELECT, UNION, EXCEPT, INTERSECT, etc.).
// This is an interface to support the various set operations.
type SetExpr interface {
	fmt.Stringer
	// setExpr is a marker method to prevent external implementations.
	setExpr()
}

// LockClause represents a row-level locking clause.
type LockClause struct {
	BaseNode
	// LockType is the type of lock.
	LockType LockType
	// Tables are the tables to lock (empty means all).
	Tables []*ObjectName
	// Wait is the wait behavior (NOWAIT, SKIP LOCKED).
	Wait LockWait
}

// LockType enumerates lock types.
type LockType int

const (
	LockTypeUpdate LockType = iota
	LockTypeShare
	LockTypeKeyShare
	LockTypeNoKeyUpdate
)

// LockWait enumerates lock wait behaviors.
type LockWait int

const (
	LockWaitUnspecified LockWait = iota
	LockWaitNowait
	LockWaitSkipLocked
)

// String returns the SQL representation of a lock clause.
func (l *LockClause) String() string {
	var lockType string
	switch l.LockType {
	case LockTypeUpdate:
		lockType = "FOR UPDATE"
	case LockTypeShare:
		lockType = "FOR SHARE"
	case LockTypeKeyShare:
		lockType = "FOR KEY SHARE"
	case LockTypeNoKeyUpdate:
		lockType = "FOR NO KEY UPDATE"
	}

	result := lockType

	if len(l.Tables) > 0 {
		tables := make([]string, len(l.Tables))
		for i, t := range l.Tables {
			tables[i] = t.String()
		}
		result += " OF " + strings.Join(tables, ", ")
	}

	switch l.Wait {
	case LockWaitNowait:
		result += " NOWAIT"
	case LockWaitSkipLocked:
		result += " SKIP LOCKED"
	}

	return result
}

// String returns the SQL representation of the query.
// This is a simplified implementation.
func (q *SelectQuery) String() string {
	var parts []string

	if q.WithClause != nil {
		parts = append(parts, q.WithClause.String())
	}

	if q.Body != nil {
		parts = append(parts, q.Body.String())
	}

	if q.OrderBy != nil {
		parts = append(parts, q.OrderBy.String())
	}

	if q.Limit != nil {
		parts = append(parts, q.Limit.String())
	}

	if q.Offset != nil {
		parts = append(parts, q.Offset.String())
	}

	if q.FetchClause != nil {
		parts = append(parts, q.FetchClause.String())
	}

	for _, lock := range q.Locks {
		parts = append(parts, lock.String())
	}

	return strings.Join(parts, " ")
}
