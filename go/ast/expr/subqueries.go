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

// Exists represents an EXISTS expression (e.g., `EXISTS (SELECT ...)`).
type Exists struct {
	Subquery *QueryExpr
	Negated  bool
	SpanVal  span.Span
}

func (e *Exists) exprNode() {}

// Span returns the source span for this expression.
func (e *Exists) Span() span.Span {
	return e.SpanVal
}

// String returns the SQL representation.
func (e *Exists) String() string {
	if e.Negated {
		return fmt.Sprintf("NOT EXISTS (%s)", e.Subquery.String())
	}
	return fmt.Sprintf("EXISTS (%s)", e.Subquery.String())
}

// Subquery represents a scalar subquery expression.
type Subquery struct {
	Query   *QueryExpr
	SpanVal span.Span
}

func (s *Subquery) exprNode() {}

// Span returns the source span for this expression.
func (s *Subquery) Span() span.Span {
	return s.SpanVal
}

// String returns the SQL representation.
func (s *Subquery) String() string {
	return fmt.Sprintf("(%s)", s.Query.String())
}

// QueryExpr represents a SELECT query (placeholder for actual query AST).
type QueryExpr struct {
	// This is a placeholder - the actual query types are in the query package
	SQL       string
	Statement interface{} // Can hold an ast.Statement (avoids circular import)
	SpanVal   span.Span
}

// Span returns the source span for this query.
func (q *QueryExpr) Span() span.Span {
	return q.SpanVal
}

// String returns the SQL representation.
func (q *QueryExpr) String() string {
	if q.Statement != nil {
		// If Statement has a String() method, use it
		if s, ok := q.Statement.(interface{ String() string }); ok {
			return s.String()
		}
	}
	return q.SQL
}

// GroupingSets represents a GROUPING SETS expression.
type GroupingSets struct {
	Sets    [][]Expr
	SpanVal span.Span
}

func (g *GroupingSets) exprNode() {}

// Span returns the source span for this expression.
func (g *GroupingSets) Span() span.Span {
	return g.SpanVal
}

// String returns the SQL representation.
func (g *GroupingSets) String() string {
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

// Cube represents a CUBE expression.
type Cube struct {
	Sets    [][]Expr
	SpanVal span.Span
}

func (c *Cube) exprNode() {}

// Span returns the source span for this expression.
func (c *Cube) Span() span.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *Cube) String() string {
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

// Rollup represents a ROLLUP expression.
type Rollup struct {
	Sets    [][]Expr
	SpanVal span.Span
}

func (r *Rollup) exprNode() {}

// Span returns the source span for this expression.
func (r *Rollup) Span() span.Span {
	return r.SpanVal
}

// String returns the SQL representation.
func (r *Rollup) String() string {
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
