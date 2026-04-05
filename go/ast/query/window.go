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

// NamedWindowDefinition represents a named window: <name> AS <window specification>
type NamedWindowDefinition struct {
	span token.Span
	Name Ident
	Expr NamedWindowExpr
}

func (n *NamedWindowDefinition) String() string {
	return fmt.Sprintf("%s AS %s", n.Name.String(), n.Expr.String())
}

// NamedWindowExpr represents an expression in a named window declaration
type NamedWindowExpr interface {
	fmt.Stringer
}

// NamedWindowReference represents a direct reference to another named window
type NamedWindowReference struct {
	Name Ident
}

func (n *NamedWindowReference) String() string {
	return n.Name.String()
}

// WindowSpecExpr represents a window specification expression
type WindowSpecExpr struct {
	Spec WindowSpec
}

func (w *WindowSpecExpr) String() string {
	return fmt.Sprintf("(%s)", w.Spec.String())
}

// WindowSpec represents a window specification
type WindowSpec struct {
	span        token.Span
	WindowName  *Ident
	PartitionBy []Expr
	OrderBy     []OrderByExpr
	WindowFrame *WindowFrame
}

func (w *WindowSpec) String() string {
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

// WindowFrame represents a window frame specification
type WindowFrame struct {
	Units WindowFrameUnits
	Start WindowFrameBound
	End   WindowFrameBound
}

func (w *WindowFrame) String() string {
	if w.End != nil {
		return fmt.Sprintf("%s BETWEEN %s AND %s", w.Units.String(), w.Start.String(), w.End.String())
	}
	return fmt.Sprintf("%s %s", w.Units.String(), w.Start.String())
}

// WindowFrameUnits represents the units for a window frame
type WindowFrameUnits int

const (
	WindowFrameUnitsRows WindowFrameUnits = iota
	WindowFrameUnitsRange
	WindowFrameUnitsGroups
)

func (w WindowFrameUnits) String() string {
	switch w {
	case WindowFrameUnitsRows:
		return "ROWS"
	case WindowFrameUnitsRange:
		return "RANGE"
	case WindowFrameUnitsGroups:
		return "GROUPS"
	default:
		return ""
	}
}

// WindowFrameBound represents a bound in a window frame
type WindowFrameBound interface {
	fmt.Stringer
}

// CurrentRowBound represents CURRENT ROW
type CurrentRowBound struct{}

func (c *CurrentRowBound) String() string { return "CURRENT ROW" }

// UnboundedPrecedingBound represents UNBOUNDED PRECEDING
type UnboundedPrecedingBound struct{}

func (u *UnboundedPrecedingBound) String() string { return "UNBOUNDED PRECEDING" }

// UnboundedFollowingBound represents UNBOUNDED FOLLOWING
type UnboundedFollowingBound struct{}

func (u *UnboundedFollowingBound) String() string { return "UNBOUNDED FOLLOWING" }

// PrecedingBound represents <expr> PRECEDING
type PrecedingBound struct {
	Expr Expr
}

func (p *PrecedingBound) String() string {
	return fmt.Sprintf("%s PRECEDING", p.Expr.String())
}

// FollowingBound represents <expr> FOLLOWING
type FollowingBound struct {
	Expr Expr
}

func (f *FollowingBound) String() string {
	return fmt.Sprintf("%s FOLLOWING", f.Expr.String())
}

// ExprWithAliasAndOrderBy represents expression with alias and order by
type ExprWithAliasAndOrderBy struct {
	Expr    ExprWithAlias
	OrderBy OrderByOptions
}

func (e *ExprWithAliasAndOrderBy) String() string {
	return e.Expr.String() + e.OrderBy.String()
}
