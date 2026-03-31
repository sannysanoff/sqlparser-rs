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

	"github.com/user/sqlparser/ast/operator"
	"github.com/user/sqlparser/span"
)

// UnaryOp represents a unary operation (e.g., NOT, -, +).
type UnaryOp struct {
	Op   operator.UnaryOperator
	Expr Expr
	SpanVal span.Span
}

func (u *UnaryOp) exprNode() {}

// Span returns the source span for this expression.
func (u *UnaryOp) Span() span.Span {
	return u.SpanVal
}

// String returns the SQL representation.
func (u *UnaryOp) String() string {
	switch u.Op {
	case operator.UOpPGPostfixFactorial:
		return fmt.Sprintf("%s%s", u.Expr.String(), u.Op.String())
	case operator.UOpNot, operator.UOpHash, operator.UOpAtDashAt,
		operator.UOpDoubleAt, operator.UOpQuestionDash, operator.UOpQuestionPipe:
		return fmt.Sprintf("%s %s", u.Op.String(), u.Expr.String())
	default:
		return fmt.Sprintf("%s%s", u.Op.String(), u.Expr.String())
	}
}

// BinaryOp represents a binary operation (e.g., +, -, *, /, AND, OR).
type BinaryOp struct {
	Left  Expr
	Op    operator.BinaryOperator
	Right Expr
	SpanVal span.Span
}

func (b *BinaryOp) exprNode() {}

// Span returns the source span for this expression.
func (b *BinaryOp) Span() span.Span {
	return b.SpanVal
}

// String returns the SQL representation.
func (b *BinaryOp) String() string {
	return fmt.Sprintf("%s %s %s", b.Left.String(), b.Op.String(), b.Right.String())
}

// IsNull represents an IS NULL expression.
type IsNull struct {
	Expr Expr
	SpanVal span.Span
}

func (i *IsNull) exprNode() {}

// Span returns the source span for this expression.
func (i *IsNull) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsNull) String() string {
	return fmt.Sprintf("%s IS NULL", i.Expr.String())
}

// IsNotNull represents an IS NOT NULL expression.
type IsNotNull struct {
	Expr Expr
	SpanVal span.Span
}

func (i *IsNotNull) exprNode() {}

// Span returns the source span for this expression.
func (i *IsNotNull) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsNotNull) String() string {
	return fmt.Sprintf("%s IS NOT NULL", i.Expr.String())
}

// IsTrue represents an IS TRUE expression.
type IsTrue struct {
	Expr Expr
	SpanVal span.Span
}

func (i *IsTrue) exprNode() {}

// Span returns the source span for this expression.
func (i *IsTrue) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsTrue) String() string {
	return fmt.Sprintf("%s IS TRUE", i.Expr.String())
}

// IsNotTrue represents an IS NOT TRUE expression.
type IsNotTrue struct {
	Expr Expr
	SpanVal span.Span
}

func (i *IsNotTrue) exprNode() {}

// Span returns the source span for this expression.
func (i *IsNotTrue) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsNotTrue) String() string {
	return fmt.Sprintf("%s IS NOT TRUE", i.Expr.String())
}

// IsFalse represents an IS FALSE expression.
type IsFalse struct {
	Expr Expr
	SpanVal span.Span
}

func (i *IsFalse) exprNode() {}

// Span returns the source span for this expression.
func (i *IsFalse) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsFalse) String() string {
	return fmt.Sprintf("%s IS FALSE", i.Expr.String())
}

// IsNotFalse represents an IS NOT FALSE expression.
type IsNotFalse struct {
	Expr Expr
	SpanVal span.Span
}

func (i *IsNotFalse) exprNode() {}

// Span returns the source span for this expression.
func (i *IsNotFalse) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsNotFalse) String() string {
	return fmt.Sprintf("%s IS NOT FALSE", i.Expr.String())
}

// IsUnknown represents an IS UNKNOWN expression.
type IsUnknown struct {
	Expr Expr
	SpanVal span.Span
}

func (i *IsUnknown) exprNode() {}

// Span returns the source span for this expression.
func (i *IsUnknown) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsUnknown) String() string {
	return fmt.Sprintf("%s IS UNKNOWN", i.Expr.String())
}

// IsNotUnknown represents an IS NOT UNKNOWN expression.
type IsNotUnknown struct {
	Expr Expr
	SpanVal span.Span
}

func (i *IsNotUnknown) exprNode() {}

// Span returns the source span for this expression.
func (i *IsNotUnknown) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsNotUnknown) String() string {
	return fmt.Sprintf("%s IS NOT UNKNOWN", i.Expr.String())
}

// IsDistinctFrom represents an IS DISTINCT FROM expression.
type IsDistinctFrom struct {
	Left  Expr
	Right Expr
	SpanVal span.Span
}

func (i *IsDistinctFrom) exprNode() {}

// Span returns the source span for this expression.
func (i *IsDistinctFrom) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsDistinctFrom) String() string {
	return fmt.Sprintf("%s IS DISTINCT FROM %s", i.Left.String(), i.Right.String())
}

// IsNotDistinctFrom represents an IS NOT DISTINCT FROM expression.
type IsNotDistinctFrom struct {
	Left  Expr
	Right Expr
	SpanVal span.Span
}

func (i *IsNotDistinctFrom) exprNode() {}

// Span returns the source span for this expression.
func (i *IsNotDistinctFrom) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsNotDistinctFrom) String() string {
	return fmt.Sprintf("%s IS NOT DISTINCT FROM %s", i.Left.String(), i.Right.String())
}

// InList represents an IN list expression (e.g., `expr IN (val1, val2, ...)`).
type InList struct {
	Expr    Expr
	List    []Expr
	Negated bool
	SpanVal span.Span
}

func (i *InList) exprNode() {}

// Span returns the source span for this expression.
func (i *InList) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *InList) String() string {
	items := make([]string, len(i.List))
	for idx, item := range i.List {
		items[idx] = item.String()
	}

	if i.Negated {
		return fmt.Sprintf("%s NOT IN (%s)", i.Expr.String(), strings.Join(items, ", "))
	}
	return fmt.Sprintf("%s IN (%s)", i.Expr.String(), strings.Join(items, ", "))
}

// InSubquery represents an IN subquery expression (e.g., `expr IN (SELECT ...)`).
type InSubquery struct {
	Expr     Expr
	Subquery *QueryExpr
	Negated  bool
	SpanVal span.Span
}

func (i *InSubquery) exprNode() {}

// Span returns the source span for this expression.
func (i *InSubquery) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *InSubquery) String() string {
	if i.Negated {
		return fmt.Sprintf("%s NOT IN (%s)", i.Expr.String(), i.Subquery.String())
	}
	return fmt.Sprintf("%s IN (%s)", i.Expr.String(), i.Subquery.String())
}

// InUnnest represents an IN UNNEST expression (e.g., `expr IN UNNEST(array)`).
type InUnnest struct {
	Expr      Expr
	ArrayExpr Expr
	Negated   bool
	SpanVal span.Span
}

func (i *InUnnest) exprNode() {}

// Span returns the source span for this expression.
func (i *InUnnest) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *InUnnest) String() string {
	if i.Negated {
		return fmt.Sprintf("%s NOT IN UNNEST(%s)", i.Expr.String(), i.ArrayExpr.String())
	}
	return fmt.Sprintf("%s IN UNNEST(%s)", i.Expr.String(), i.ArrayExpr.String())
}

// Between represents a BETWEEN expression.
type Between struct {
	Expr    Expr
	Negated bool
	Low     Expr
	High    Expr
	SpanVal span.Span
}

func (b *Between) exprNode() {}

// Span returns the source span for this expression.
func (b *Between) Span() span.Span {
	return b.SpanVal
}

// String returns the SQL representation.
func (b *Between) String() string {
	if b.Negated {
		return fmt.Sprintf("%s NOT BETWEEN %s AND %s", b.Expr.String(), b.Low.String(), b.High.String())
	}
	return fmt.Sprintf("%s BETWEEN %s AND %s", b.Expr.String(), b.Low.String(), b.High.String())
}

// Like represents a LIKE expression.
type Like struct {
	Negated    bool
	Any        bool // Snowflake ANY keyword
	Expr       Expr
	Pattern    Expr
	EscapeChar interface{} // ValueWithSpan
	SpanVal span.Span
}

func (l *Like) exprNode() {}

// Span returns the source span for this expression.
func (l *Like) Span() span.Span {
	return l.SpanVal
}

// String returns the SQL representation.
func (l *Like) String() string {
	var sb strings.Builder
	sb.WriteString(l.Expr.String())
	sb.WriteString(" ")
	if l.Negated {
		sb.WriteString("NOT ")
	}
	sb.WriteString("LIKE ")
	if l.Any {
		sb.WriteString("ANY ")
	}
	sb.WriteString(l.Pattern.String())
	if l.EscapeChar != nil {
		sb.WriteString(fmt.Sprintf(" ESCAPE %v", l.EscapeChar))
	}
	return sb.String()
}

// ILike represents an ILIKE (case-insensitive LIKE) expression.
type ILike struct {
	Negated    bool
	Any        bool // Snowflake ANY keyword
	Expr       Expr
	Pattern    Expr
	EscapeChar interface{} // ValueWithSpan
	SpanVal span.Span
}

func (i *ILike) exprNode() {}

// Span returns the source span for this expression.
func (i *ILike) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *ILike) String() string {
	var sb strings.Builder
	sb.WriteString(i.Expr.String())
	sb.WriteString(" ")
	if i.Negated {
		sb.WriteString("NOT ")
	}
	sb.WriteString("ILIKE ")
	if i.Any {
		sb.WriteString("ANY ")
	}
	sb.WriteString(i.Pattern.String())
	if i.EscapeChar != nil {
		sb.WriteString(fmt.Sprintf(" ESCAPE %v", i.EscapeChar))
	}
	return sb.String()
}

// SimilarTo represents a SIMILAR TO regex expression.
type SimilarTo struct {
	Negated    bool
	Expr       Expr
	Pattern    Expr
	EscapeChar interface{} // ValueWithSpan
	SpanVal span.Span
}

func (s *SimilarTo) exprNode() {}

// Span returns the source span for this expression.
func (s *SimilarTo) Span() span.Span {
	return s.SpanVal
}

// String returns the SQL representation.
func (s *SimilarTo) String() string {
	var sb strings.Builder
	sb.WriteString(s.Expr.String())
	sb.WriteString(" ")
	if s.Negated {
		sb.WriteString("NOT ")
	}
	sb.WriteString("SIMILAR TO ")
	sb.WriteString(s.Pattern.String())
	if s.EscapeChar != nil {
		sb.WriteString(fmt.Sprintf(" ESCAPE %v", s.EscapeChar))
	}
	return sb.String()
}

// RLike represents an RLIKE/REGEXP expression.
type RLike struct {
	Negated bool
	Expr    Expr
	Pattern Expr
	Regexp  bool // true for REGEXP, false for RLIKE
	SpanVal span.Span
}

func (r *RLike) exprNode() {}

// Span returns the source span for this expression.
func (r *RLike) Span() span.Span {
	return r.SpanVal
}

// String returns the SQL representation.
func (r *RLike) String() string {
	var sb strings.Builder
	sb.WriteString(r.Expr.String())
	sb.WriteString(" ")
	if r.Negated {
		sb.WriteString("NOT ")
	}
	if r.Regexp {
		sb.WriteString("REGEXP ")
	} else {
		sb.WriteString("RLIKE ")
	}
	sb.WriteString(r.Pattern.String())
	return sb.String()
}

// CastKind represents the kind of cast operation.
type CastKind int

const (
	CastStandard CastKind = iota
	CastTry
	CastSafe
	CastDoubleColon
)

// CastFormat represents the format for CAST expressions.
type CastFormat struct {
	Value           interface{} // ValueWithSpan
	ValueAtTimeZone *struct {
		Value    interface{} // ValueWithSpan
		TimeZone interface{} // ValueWithSpan
	}
}

// Cast represents a CAST expression.
type Cast struct {
	Kind     CastKind
	Expr     Expr
	DataType string
	Array    bool // MySQL-specific: CAST(... AS type ARRAY)
	Format   *CastFormat
	SpanVal span.Span
}

func (c *Cast) exprNode() {}

// Span returns the source span for this expression.
func (c *Cast) Span() span.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *Cast) String() string {
	var sb strings.Builder

	switch c.Kind {
	case CastStandard:
		sb.WriteString("CAST(")
	case CastTry:
		sb.WriteString("TRY_CAST(")
	case CastSafe:
		sb.WriteString("SAFE_CAST(")
	case CastDoubleColon:
		return fmt.Sprintf("%s::%s", c.Expr.String(), c.DataType)
	}

	sb.WriteString(c.Expr.String())
	sb.WriteString(" AS ")
	sb.WriteString(c.DataType)

	if c.Array {
		sb.WriteString(" ARRAY")
	}

	if c.Format != nil {
		sb.WriteString(" FORMAT ")
		if c.Format.ValueAtTimeZone != nil {
			sb.WriteString(fmt.Sprintf("%v AT TIME ZONE %v",
				c.Format.ValueAtTimeZone.Value,
				c.Format.ValueAtTimeZone.TimeZone))
		} else {
			sb.WriteString(fmt.Sprintf("%v", c.Format.Value))
		}
	}

	sb.WriteString(")")
	return sb.String()
}

// Convert represents a CONVERT expression.
type Convert struct {
	IsTry             bool
	Expr              Expr
	DataType          *string
	Charset           *ObjectName
	TargetBeforeValue bool // MSSQL syntax
	Styles            []Expr
	SpanVal span.Span
}

func (c *Convert) exprNode() {}

// Span returns the source span for this expression.
func (c *Convert) Span() span.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *Convert) String() string {
	var sb strings.Builder

	if c.IsTry {
		sb.WriteString("TRY_")
	}
	sb.WriteString("CONVERT(")

	if c.DataType != nil {
		if c.TargetBeforeValue {
			sb.WriteString(*c.DataType)
			sb.WriteString(", ")
			sb.WriteString(c.Expr.String())
		} else {
			sb.WriteString(c.Expr.String())
			sb.WriteString(", ")
			sb.WriteString(*c.DataType)
		}

		if c.Charset != nil {
			sb.WriteString(" CHARACTER SET ")
			sb.WriteString(c.Charset.String())
		}
	} else if c.Charset != nil {
		sb.WriteString(c.Expr.String())
		sb.WriteString(" USING ")
		sb.WriteString(c.Charset.String())
	} else {
		sb.WriteString(c.Expr.String())
	}

	if len(c.Styles) > 0 {
		styles := make([]string, len(c.Styles))
		for i, s := range c.Styles {
			styles[i] = s.String()
		}
		sb.WriteString(", ")
		sb.WriteString(strings.Join(styles, ", "))
	}

	sb.WriteString(")")
	return sb.String()
}

// Collate represents a COLLATE expression.
type Collate struct {
	Expr      Expr
	Collation *ObjectName
	SpanVal span.Span
}

func (c *Collate) exprNode() {}

// Span returns the source span for this expression.
func (c *Collate) Span() span.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *Collate) String() string {
	return fmt.Sprintf("%s COLLATE %s", c.Expr.String(), c.Collation.String())
}

// AnyOp represents an ANY/SOME subquery comparison expression.
type AnyOp struct {
	Left      Expr
	CompareOp operator.BinaryOperator
	Right     Expr
	IsSome    bool // ANY and SOME are synonymous
	SpanVal span.Span
}

func (a *AnyOp) exprNode() {}

// Span returns the source span for this expression.
func (a *AnyOp) Span() span.Span {
	return a.SpanVal
}

// String returns the SQL representation.
func (a *AnyOp) String() string {
	opName := "ANY"
	if a.IsSome {
		opName = "SOME"
	}

	// Check if right is a subquery (no parentheses needed)
	_, isSubquery := a.Right.(*Subquery)
	if isSubquery {
		return fmt.Sprintf("%s %s %s %s", a.Left.String(), a.CompareOp.String(), opName, a.Right.String())
	}
	return fmt.Sprintf("%s %s %s(%s)", a.Left.String(), a.CompareOp.String(), opName, a.Right.String())
}

// AllOp represents an ALL subquery comparison expression.
type AllOp struct {
	Left      Expr
	CompareOp operator.BinaryOperator
	Right     Expr
	SpanVal span.Span
}

func (a *AllOp) exprNode() {}

// Span returns the source span for this expression.
func (a *AllOp) Span() span.Span {
	return a.SpanVal
}

// String returns the SQL representation.
func (a *AllOp) String() string {
	// Check if right is a subquery (no parentheses needed)
	_, isSubquery := a.Right.(*Subquery)
	if isSubquery {
		return fmt.Sprintf("%s %s ALL %s", a.Left.String(), a.CompareOp.String(), a.Right.String())
	}
	return fmt.Sprintf("%s %s ALL(%s)", a.Left.String(), a.CompareOp.String(), a.Right.String())
}

// NormalizationForm represents the Unicode normalization form.
type NormalizationForm int

const (
	FormNFC NormalizationForm = iota
	FormNFD
	FormNFKC
	FormNFKD
)

// String returns the SQL representation.
func (n NormalizationForm) String() string {
	switch n {
	case FormNFC:
		return "NFC"
	case FormNFD:
		return "NFD"
	case FormNFKC:
		return "NFKC"
	case FormNFKD:
		return "NFKD"
	}
	return ""
}

// IsNormalized represents an IS [NOT] [form] NORMALIZED expression.
type IsNormalized struct {
	Expr    Expr
	Form    *NormalizationForm
	Negated bool
	SpanVal span.Span
}

func (i *IsNormalized) exprNode() {}

// Span returns the source span for this expression.
func (i *IsNormalized) Span() span.Span {
	return i.SpanVal
}

// String returns the SQL representation.
func (i *IsNormalized) String() string {
	var sb strings.Builder
	sb.WriteString(i.Expr.String())
	sb.WriteString(" IS ")
	if i.Negated {
		sb.WriteString("NOT ")
	}
	if i.Form != nil {
		sb.WriteString(i.Form.String())
		sb.WriteString(" ")
	}
	sb.WriteString("NORMALIZED")
	return sb.String()
}

// ExtractSyntax represents the syntax for EXTRACT expressions.
type ExtractSyntax int

const (
	ExtractFrom ExtractSyntax = iota
	ExtractComma
)

// Extract represents an EXTRACT expression.
type Extract struct {
	Field  string // DateTimeField
	Syntax ExtractSyntax
	Expr   Expr
	SpanVal span.Span
}

func (e *Extract) exprNode() {}

// Span returns the source span for this expression.
func (e *Extract) Span() span.Span {
	return e.SpanVal
}

// String returns the SQL representation.
func (e *Extract) String() string {
	if e.Syntax == ExtractFrom {
		return fmt.Sprintf("EXTRACT(%s FROM %s)", e.Field, e.Expr.String())
	}
	return fmt.Sprintf("EXTRACT(%s, %s)", e.Field, e.Expr.String())
}

// CeilFloorKind represents the kind for CEIL/FLOOR expressions.
type CeilFloorKind int

const (
	CeilFloorDateTime CeilFloorKind = iota
	CeilFloorScale
)

// CeilExpr represents a CEIL expression.
type CeilExpr struct {
	Expr  Expr
	Field struct {
		Kind          CeilFloorKind
		DateTimeField *string
		Scale         interface{} // ValueWithSpan
	}
	SpanVal span.Span
}

func (c *CeilExpr) exprNode() {}

// Span returns the source span for this expression.
func (c *CeilExpr) Span() span.Span {
	return c.SpanVal
}

// String returns the SQL representation.
func (c *CeilExpr) String() string {
	if c.Field.Kind == CeilFloorDateTime && c.Field.DateTimeField != nil {
		return fmt.Sprintf("CEIL(%s TO %s)", c.Expr.String(), *c.Field.DateTimeField)
	}
	if c.Field.Kind == CeilFloorScale {
		return fmt.Sprintf("CEIL(%s, %v)", c.Expr.String(), c.Field.Scale)
	}
	return fmt.Sprintf("CEIL(%s)", c.Expr.String())
}

// FloorExpr represents a FLOOR expression.
type FloorExpr struct {
	Expr  Expr
	Field struct {
		Kind          CeilFloorKind
		DateTimeField *string
		Scale         interface{} // ValueWithSpan
	}
	SpanVal span.Span
}

func (f *FloorExpr) exprNode() {}

// Span returns the source span for this expression.
func (f *FloorExpr) Span() span.Span {
	return f.SpanVal
}

// String returns the SQL representation.
func (f *FloorExpr) String() string {
	if f.Field.Kind == CeilFloorDateTime && f.Field.DateTimeField != nil {
		return fmt.Sprintf("FLOOR(%s TO %s)", f.Expr.String(), *f.Field.DateTimeField)
	}
	if f.Field.Kind == CeilFloorScale {
		return fmt.Sprintf("FLOOR(%s, %v)", f.Expr.String(), f.Field.Scale)
	}
	return fmt.Sprintf("FLOOR(%s)", f.Expr.String())
}

// PositionExpr represents a POSITION expression.
type PositionExpr struct {
	Expr Expr
	In   Expr
	SpanVal span.Span
}

func (p *PositionExpr) exprNode() {}

// Span returns the source span for this expression.
func (p *PositionExpr) Span() span.Span {
	return p.SpanVal
}

// String returns the SQL representation.
func (p *PositionExpr) String() string {
	return fmt.Sprintf("POSITION(%s IN %s)", p.Expr.String(), p.In.String())
}

// Substring represents a SUBSTRING/SUBSTR expression.
type Substring struct {
	Expr          Expr
	SubstringFrom *Expr
	SubstringFor  *Expr
	Special       bool // true for SUBSTRING(expr, start, len) syntax
	Shorthand     bool // true for SUBSTR shorthand
	SpanVal span.Span
}

func (s *Substring) exprNode() {}

// Span returns the source span for this expression.
func (s *Substring) Span() span.Span {
	return s.SpanVal
}

// String returns the SQL representation.
func (s *Substring) String() string {
	var sb strings.Builder
	sb.WriteString("SUBSTR")
	if !s.Shorthand {
		sb.WriteString("ING")
	}
	sb.WriteString("(")
	sb.WriteString(s.Expr.String())

	if s.SubstringFrom != nil {
		if s.Special {
			sb.WriteString(", ")
			sb.WriteString((*s.SubstringFrom).String())
		} else {
			sb.WriteString(" FROM ")
			sb.WriteString((*s.SubstringFrom).String())
		}
	}

	if s.SubstringFor != nil {
		if s.Special {
			sb.WriteString(", ")
			sb.WriteString((*s.SubstringFor).String())
		} else {
			sb.WriteString(" FOR ")
			sb.WriteString((*s.SubstringFor).String())
		}
	}

	sb.WriteString(")")
	return sb.String()
}

// TrimWhere represents the trim direction.
type TrimWhere int

const (
	TrimBoth TrimWhere = iota
	TrimLeading
	TrimTrailing
)

// String returns the SQL representation.
func (t TrimWhere) String() string {
	switch t {
	case TrimBoth:
		return "BOTH"
	case TrimLeading:
		return "LEADING"
	case TrimTrailing:
		return "TRAILING"
	}
	return ""
}

// TrimExpr represents a TRIM expression.
type TrimExpr struct {
	TrimWhere      *TrimWhere
	TrimWhat       *Expr
	Expr           Expr
	TrimCharacters []Expr
	SpanVal span.Span
}

func (t *TrimExpr) exprNode() {}

// Span returns the source span for this expression.
func (t *TrimExpr) Span() span.Span {
	return t.SpanVal
}

// String returns the SQL representation.
func (t *TrimExpr) String() string {
	var sb strings.Builder
	sb.WriteString("TRIM(")

	if t.TrimWhere != nil {
		sb.WriteString(t.TrimWhere.String())
		sb.WriteString(" ")
	}

	if t.TrimWhat != nil {
		sb.WriteString((*t.TrimWhat).String())
		sb.WriteString(" FROM ")
		sb.WriteString(t.Expr.String())
	} else {
		sb.WriteString(t.Expr.String())
	}

	if len(t.TrimCharacters) > 0 {
		chars := make([]string, len(t.TrimCharacters))
		for i, c := range t.TrimCharacters {
			chars[i] = c.String()
		}
		sb.WriteString(", ")
		sb.WriteString(strings.Join(chars, ", "))
	}

	sb.WriteString(")")
	return sb.String()
}

// OverlayExpr represents an OVERLAY expression.
type OverlayExpr struct {
	Expr        Expr
	OverlayWhat Expr
	OverlayFrom Expr
	OverlayFor  *Expr
	SpanVal span.Span
}

func (o *OverlayExpr) exprNode() {}

// Span returns the source span for this expression.
func (o *OverlayExpr) Span() span.Span {
	return o.SpanVal
}

// String returns the SQL representation.
func (o *OverlayExpr) String() string {
	var sb strings.Builder
	sb.WriteString("OVERLAY(")
	sb.WriteString(o.Expr.String())
	sb.WriteString(" PLACING ")
	sb.WriteString(o.OverlayWhat.String())
	sb.WriteString(" FROM ")
	sb.WriteString(o.OverlayFrom.String())

	if o.OverlayFor != nil {
		sb.WriteString(" FOR ")
		sb.WriteString((*o.OverlayFor).String())
	}

	sb.WriteString(")")
	return sb.String()
}

// AtTimeZone represents an AT TIME ZONE expression.
type AtTimeZone struct {
	Timestamp Expr
	TimeZone  Expr
	SpanVal span.Span
}

func (a *AtTimeZone) exprNode() {}

// Span returns the source span for this expression.
func (a *AtTimeZone) Span() span.Span {
	return a.SpanVal
}

// String returns the SQL representation.
func (a *AtTimeZone) String() string {
	return fmt.Sprintf("%s AT TIME ZONE %s", a.Timestamp.String(), a.TimeZone.String())
}
