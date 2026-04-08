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

package statement

import (
	"strings"

	"github.com/user/sqlparser/ast"
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/ast/query"
)

// ============================================================================
// QUERY (SELECT wrapper)
// ============================================================================

// Query wraps a SELECT statement as a standalone statement
type Query struct {
	BaseStatement
	Query *query.Query
}

func (q *Query) statementNode() {}

func (q *Query) String() string {
	if q.Query != nil {
		return q.Query.String()
	}
	return ""
}

// ============================================================================
// INSERT
// ============================================================================

// Insert represents an INSERT statement
type Insert struct {
	BaseStatement
	OptimizerHints        []*expr.OptimizerHint
	Or                    *expr.SqliteOnConflict
	Ignore                bool
	Into                  bool
	Table                 *ast.ObjectName
	TableAlias            *ast.Ident
	TableAliasExplicit    bool // true if "AS" was used before the alias
	Columns               []*ast.Ident
	Overwrite             bool
	Source                *query.Query
	Assignments           []*expr.Assignment
	Partitioned           []expr.Expr
	AfterColumns          []*ast.Ident
	HasTableKeyword       bool
	On                    *expr.OnInsert
	Returning             []*query.SelectItem
	Output                *expr.OutputClause
	ReplaceInto           bool
	Priority              *expr.MysqlInsertPriority
	InsertAlias           *expr.InsertAliases
	Settings              []*expr.Setting
	FormatClause          *expr.InputFormatClause
	MultiTableInsertType  *expr.MultiTableInsertType
	MultiTableIntoClauses []*expr.MultiTableInsertIntoClause
	MultiTableWhenClauses []*expr.MultiTableInsertWhenClause
	MultiTableElseClause  []*expr.MultiTableInsertIntoClause
	DefaultValues         bool
}

func (i *Insert) statementNode() {}

func (i *Insert) String() string {
	var f strings.Builder

	if i.ReplaceInto {
		f.WriteString("REPLACE")
	} else {
		f.WriteString("INSERT")
	}

	for _, hint := range i.OptimizerHints {
		f.WriteString(" ")
		f.WriteString(hint.String())
	}

	if i.Priority != nil {
		f.WriteString(" ")
		f.WriteString(i.Priority.String())
	}

	if i.Ignore {
		f.WriteString(" IGNORE")
	}

	if i.Overwrite {
		f.WriteString(" OVERWRITE")
	}

	// Handle multi-table insert
	if i.MultiTableInsertType != nil && *i.MultiTableInsertType != expr.MultiTableInsertTypeNone {
		f.WriteString(" ")
		f.WriteString(i.MultiTableInsertType.String())

		// Output WHEN clauses for conditional multi-table insert
		for _, when := range i.MultiTableWhenClauses {
			f.WriteString(" ")
			f.WriteString(when.String())
		}

		// Output ELSE clause if present
		if len(i.MultiTableElseClause) > 0 {
			f.WriteString(" ELSE")
			for _, into := range i.MultiTableElseClause {
				f.WriteString(" ")
				f.WriteString(into.String())
			}
		}

		// Output INTO clauses for unconditional multi-table insert
		for _, into := range i.MultiTableIntoClauses {
			f.WriteString(" ")
			f.WriteString(into.String())
		}

		// Output source query
		if i.Source != nil {
			f.WriteString(" ")
			f.WriteString(i.Source.String())
		}

		return f.String()
	}

	if i.Into {
		f.WriteString(" INTO")
	}

	if i.HasTableKeyword {
		f.WriteString(" TABLE")
	}

	f.WriteString(" ")
	f.WriteString(i.Table.String())

	// Add table alias if present (PostgreSQL style: INSERT INTO table AS alias)
	if i.TableAlias != nil {
		if i.TableAliasExplicit {
			f.WriteString(" AS")
		}
		f.WriteString(" ")
		f.WriteString(i.TableAlias.String())
	}

	if len(i.Columns) > 0 {
		f.WriteString(" (")
		f.WriteString(formatIdents(i.Columns, ", "))
		f.WriteString(")")
	}

	if i.DefaultValues {
		f.WriteString(" DEFAULT VALUES")
	} else if i.Source != nil {
		f.WriteString(" ")
		f.WriteString(i.Source.String())
	} else if len(i.Assignments) > 0 {
		// MySQL INSERT SET syntax
		f.WriteString(" SET ")
		for j, assign := range i.Assignments {
			if j > 0 {
				f.WriteString(", ")
			}
			f.WriteString(assign.String())
		}
	}

	// Add AS alias if present (MySQL)
	if i.InsertAlias != nil && i.InsertAlias.RowAlias != nil {
		f.WriteString(" ")
		f.WriteString(i.InsertAlias.String())
	}

	// Add ON CONFLICT clause if present
	if i.On != nil {
		f.WriteString(" ")
		f.WriteString(i.On.String())
	}

	// Add RETURNING clause if present
	if len(i.Returning) > 0 {
		f.WriteString(" RETURNING ")
		for j, item := range i.Returning {
			if j > 0 {
				f.WriteString(", ")
			}
			if item != nil && *item != nil {
				f.WriteString((*item).String())
			}
		}
	}

	return f.String()
}

// ============================================================================
// UPDATE
// ============================================================================

// Update represents an UPDATE statement
type Update struct {
	BaseStatement
	Table           *ast.ObjectName
	TableAlias      *ast.Ident
	Assignments     []*expr.Assignment
	From            *query.UpdateTableFromKind // Changed from *query.TableWithJoins
	Selection       expr.Expr
	Returning       []*query.SelectItem
	Output          *expr.OutputClause
	OrderBy         []*expr.OrderByExpr
	Limit           query.LimitClause
	IsFromStatement bool
	Setting         []*expr.Setting
}

func (u *Update) statementNode() {}

func (u *Update) String() string {
	var f strings.Builder
	f.WriteString("UPDATE ")

	// Handle FROM clause position for different SQL dialects
	if u.From != nil && u.From.BeforeSet != nil {
		// Snowflake/MSSQL style: UPDATE FROM t1 SET ...
		f.WriteString("FROM ")
		f.WriteString(u.From.String())
		f.WriteString(" ")
	} else if u.Table != nil {
		f.WriteString(u.Table.String())
	}

	if len(u.Assignments) > 0 {
		f.WriteString(" SET ")
		for i, assign := range u.Assignments {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(assign.String())
		}
	}

	// PostgreSQL style: UPDATE t1 SET ... FROM t2
	if u.From != nil && u.From.AfterSet != nil {
		f.WriteString(" FROM ")
		f.WriteString(u.From.String())
	}

	if u.Selection != nil {
		f.WriteString(" WHERE ")
		f.WriteString(u.Selection.String())
	}

	if u.Returning != nil {
		f.WriteString(" RETURNING ")
		for i, item := range u.Returning {
			if i > 0 {
				f.WriteString(", ")
			}
			if item != nil && *item != nil {
				f.WriteString((*item).String())
			}
		}
	}

	return f.String()
}

// ============================================================================
// DELETE
// ============================================================================

// Delete represents a DELETE statement
type Delete struct {
	BaseStatement
	Tables    []*ast.ObjectName
	Using     []*query.TableWithJoins
	Selection expr.Expr
	Returning []*query.SelectItem
	Output    *expr.OutputClause
	OrderBy   []query.OrderByExpr
	Limit     query.LimitClause
}

func (d *Delete) statementNode() {}

func (d *Delete) String() string {
	var f strings.Builder
	f.WriteString("DELETE FROM ")

	for i, table := range d.Tables {
		if i > 0 {
			f.WriteString(", ")
		}
		f.WriteString(table.String())
	}

	if d.Selection != nil {
		f.WriteString(" WHERE ")
		f.WriteString(d.Selection.String())
	}

	// Add ORDER BY clause if present (MySQL)
	if len(d.OrderBy) > 0 {
		f.WriteString(" ORDER BY ")
		for i, obe := range d.OrderBy {
			if i > 0 {
				f.WriteString(", ")
			}
			f.WriteString(obe.String())
		}
	}

	// Add LIMIT clause if present (MySQL)
	if d.Limit != nil {
		f.WriteString(" ")
		f.WriteString(d.Limit.String())
	}

	// Add RETURNING clause if present
	if len(d.Returning) > 0 {
		f.WriteString(" RETURNING ")
		for i, item := range d.Returning {
			if i > 0 {
				f.WriteString(", ")
			}
			if item != nil && *item != nil {
				f.WriteString((*item).String())
			}
		}
	}

	return f.String()
}

// ============================================================================
// MERGE
// ============================================================================

// Merge represents a MERGE statement
type Merge struct {
	BaseStatement
	Into       bool
	Table      query.TableFactor
	TableAlias *ast.Ident
	Source     query.TableFactor
	On         expr.Expr
	Clauses    []*expr.MergeClause
	Output     *expr.OutputClause
}

func (m *Merge) statementNode() {}

func (m *Merge) String() string {
	var f strings.Builder
	f.WriteString("MERGE")
	if m.Into {
		f.WriteString(" INTO")
	}
	if m.Table != nil {
		f.WriteString(" ")
		f.WriteString(m.Table.String())
	}
	f.WriteString(" USING ")
	if m.Source != nil {
		f.WriteString(m.Source.String())
	}
	f.WriteString(" ON ")
	f.WriteString(m.On.String())
	for _, clause := range m.Clauses {
		f.WriteString(" ")
		f.WriteString(clause.String())
	}
	if m.Output != nil {
		f.WriteString(" ")
		f.WriteString(m.Output.String())
	}
	return f.String()
}
