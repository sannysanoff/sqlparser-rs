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

// CaseWhen represents a WHEN clause in a CASE expression.
type CaseWhen struct {
	Condition Expr
	Result    Expr
}

// String returns the SQL representation.
func (c *CaseWhen) String() string {
	return fmt.Sprintf("WHEN %s THEN %s", c.Condition.String(), c.Result.String())
}

// CaseExpr represents a CASE expression.
type CaseExpr struct {
	SpanVal    span.Span
	Operand    Expr
	Conditions []CaseWhen
	ElseResult Expr
}

func (c *CaseExpr) exprNode() {}

// Span returns the source span for this expression.
func (c *CaseExpr) Span() span.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *CaseExpr) String() string {
	var sb strings.Builder
	sb.WriteString("CASE")

	if c.Operand != nil {
		sb.WriteString(" ")
		sb.WriteString(c.Operand.String())
	}

	for _, when := range c.Conditions {
		sb.WriteString(" ")
		sb.WriteString(when.String())
	}

	if c.ElseResult != nil {
		sb.WriteString(" ELSE ")
		sb.WriteString(c.ElseResult.String())
	}

	sb.WriteString(" END")
	return sb.String()
}

// IfExpr represents an IF expression (e.g., `IF(condition, true_value, false_value)`).
type IfExpr struct {
	SpanVal    span.Span
	Condition  Expr
	TrueValue  Expr
	FalseValue Expr
}

func (i *IfExpr) exprNode() {}

// Span returns the source span for this expression.
func (i *IfExpr) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IfExpr) String() string {
	return fmt.Sprintf("IF(%s, %s, %s)",
		i.Condition.String(), i.TrueValue.String(), i.FalseValue.String())
}

// CoalesceExpr represents a COALESCE expression.
type CoalesceExpr struct {
	SpanVal span.Span
	Exprs   []Expr
}

func (c *CoalesceExpr) exprNode() {}

// Span returns the source span for this expression.
func (c *CoalesceExpr) Span() span.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *CoalesceExpr) String() string {
	exprs := make([]string, len(c.Exprs))
	for i, expr := range c.Exprs {
		exprs[i] = expr.String()
	}
	return fmt.Sprintf("COALESCE(%s)", strings.Join(exprs, ", "))
}

// NullIfExpr represents a NULLIF expression.
type NullIfExpr struct {
	SpanVal span.Span
	Expr1   Expr
	Expr2   Expr
}

func (n *NullIfExpr) exprNode() {}

// Span returns the source span for this expression.
func (n *NullIfExpr) Span() span.Span {
	return n.SpanVal
}

// String returns the SQL representation.
func (n *NullIfExpr) String() string {
	return fmt.Sprintf("NULLIF(%s, %s)", n.Expr1.String(), n.Expr2.String())
}

// IfNullExpr represents an IFNULL expression.
type IfNullExpr struct {
	SpanVal span.Span
	Expr1   Expr
	Expr2   Expr
}

func (i *IfNullExpr) exprNode() {}

// Span returns the source span for this expression.
func (i *IfNullExpr) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IfNullExpr) String() string {
	return fmt.Sprintf("IFNULL(%s, %s)", i.Expr1.String(), i.Expr2.String())
}

// GreatestExpr represents a GREATEST expression.
type GreatestExpr struct {
	SpanVal span.Span
	Exprs   []Expr
}

func (g *GreatestExpr) exprNode() {}

// Span returns the source span for this expression.
func (g *GreatestExpr) Span() span.Span {
	return g.SpanVal
}

// String returns the SQL representation.
func (g *GreatestExpr) String() string {
	exprs := make([]string, len(g.Exprs))
	for i, expr := range g.Exprs {
		exprs[i] = expr.String()
	}
	return fmt.Sprintf("GREATEST(%s)", strings.Join(exprs, ", "))
}

// LeastExpr represents a LEAST expression.
type LeastExpr struct {
	SpanVal span.Span
	Exprs   []Expr
}

func (l *LeastExpr) exprNode() {}

// Span returns the source span for this expression.
func (l *LeastExpr) Span() span.Span {
	return l.SpanVal
}

// String returns the SQL representation.
func (l *LeastExpr) String() string {
	exprs := make([]string, len(l.Exprs))
	for i, expr := range l.Exprs {
		exprs[i] = expr.String()
	}
	return fmt.Sprintf("LEAST(%s)", strings.Join(exprs, ", "))
}
