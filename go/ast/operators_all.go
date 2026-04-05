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

// This file consolidates operator expression types from ast/expr/operators.go
// into the main ast package.
//
// Key changes:
// - Expression types use "E" prefix
// - Uses existing UnaryOperator and BinaryOperator types from expr.go
// - Moved from expr/ subpackage into ast package
//
// TODO: After full migration, remove the old ast/expr/ directory

package ast

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// ============================================================================
// Operator Expressions (from ast/expr/operators.go)
// ============================================================================

// EUnaryOp represents a unary operation (was expr.UnaryOp).
type EUnaryOp struct {
	ExpressionBase
	Op      UnaryOperator
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (u *EUnaryOp) Span() token.Span { return u.SpanVal }

// String returns the SQL representation.
func (u *EUnaryOp) String() string {
	switch u.Op {
	case UnaryOperatorFactorial:
		return fmt.Sprintf("%s%s", u.Expr.String(), u.Op.String())
	case UnaryOperatorNot, UnaryOperatorBitwiseNot:
		return fmt.Sprintf("%s %s", u.Op.String(), u.Expr.String())
	default:
		return fmt.Sprintf("%s%s", u.Op.String(), u.Expr.String())
	}
}

// EBinaryOp represents a binary operation (was expr.BinaryOp).
type EBinaryOp struct {
	ExpressionBase
	Left    Expr
	Op      BinaryOperator
	Right   Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (b *EBinaryOp) Span() token.Span { return b.SpanVal }

// String returns the SQL representation.
func (b *EBinaryOp) String() string {
	return fmt.Sprintf("%s %s %s", b.Left.String(), b.Op.String(), b.Right.String())
}

// EIsNull represents an IS NULL expression (was expr.IsNull).
type EIsNull struct {
	ExpressionBase
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsNull) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsNull) String() string {
	return fmt.Sprintf("%s IS NULL", i.Expr.String())
}

// EIsNotNull represents an IS NOT NULL expression (was expr.IsNotNull).
type EIsNotNull struct {
	ExpressionBase
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsNotNull) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsNotNull) String() string {
	return fmt.Sprintf("%s IS NOT NULL", i.Expr.String())
}

// EIsTrue represents an IS TRUE expression (was expr.IsTrue).
type EIsTrue struct {
	ExpressionBase
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsTrue) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsTrue) String() string {
	return fmt.Sprintf("%s IS TRUE", i.Expr.String())
}

// EIsNotTrue represents an IS NOT TRUE expression (was expr.IsNotTrue).
type EIsNotTrue struct {
	ExpressionBase
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsNotTrue) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsNotTrue) String() string {
	return fmt.Sprintf("%s IS NOT TRUE", i.Expr.String())
}

// EIsFalse represents an IS FALSE expression (was expr.IsFalse).
type EIsFalse struct {
	ExpressionBase
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsFalse) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsFalse) String() string {
	return fmt.Sprintf("%s IS FALSE", i.Expr.String())
}

// EIsNotFalse represents an IS NOT FALSE expression (was expr.IsNotFalse).
type EIsNotFalse struct {
	ExpressionBase
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsNotFalse) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsNotFalse) String() string {
	return fmt.Sprintf("%s IS NOT FALSE", i.Expr.String())
}

// EIsUnknown represents an IS UNKNOWN expression (was expr.IsUnknown).
type EIsUnknown struct {
	ExpressionBase
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsUnknown) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsUnknown) String() string {
	return fmt.Sprintf("%s IS UNKNOWN", i.Expr.String())
}

// EIsNotUnknown represents an IS NOT UNKNOWN expression (was expr.IsNotUnknown).
type EIsNotUnknown struct {
	ExpressionBase
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsNotUnknown) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsNotUnknown) String() string {
	return fmt.Sprintf("%s IS NOT UNKNOWN", i.Expr.String())
}

// EIsDistinctFrom represents an IS DISTINCT FROM expression (was expr.IsDistinctFrom).
type EIsDistinctFrom struct {
	ExpressionBase
	Left    Expr
	Right   Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsDistinctFrom) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsDistinctFrom) String() string {
	return fmt.Sprintf("%s IS DISTINCT FROM %s", i.Left.String(), i.Right.String())
}

// EIsNotDistinctFrom represents an IS NOT DISTINCT FROM expression (was expr.IsNotDistinctFrom).
type EIsNotDistinctFrom struct {
	ExpressionBase
	Left    Expr
	Right   Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsNotDistinctFrom) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsNotDistinctFrom) String() string {
	return fmt.Sprintf("%s IS NOT DISTINCT FROM %s", i.Left.String(), i.Right.String())
}

// EInList represents an IN list expression (was expr.InList).
type EInList struct {
	ExpressionBase
	Expr    Expr
	List    []Expr
	Negated bool
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EInList) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EInList) String() string {
	items := make([]string, len(i.List))
	for idx, item := range i.List {
		items[idx] = item.String()
	}

	if i.Negated {
		return fmt.Sprintf("%s NOT IN (%s)", i.Expr.String(), strings.Join(items, ", "))
	}
	return fmt.Sprintf("%s IN (%s)", i.Expr.String(), strings.Join(items, ", "))
}

// EInSubquery represents an IN subquery expression (was expr.InSubquery).
type EInSubquery struct {
	ExpressionBase
	Expr     Expr
	Subquery *EQueryExpr
	Negated  bool
	SpanVal  token.Span
}

// Span returns the source span for this expression.
func (i *EInSubquery) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EInSubquery) String() string {
	if i.Negated {
		return fmt.Sprintf("%s NOT IN (%s)", i.Expr.String(), i.Subquery.String())
	}
	return fmt.Sprintf("%s IN (%s)", i.Expr.String(), i.Subquery.String())
}

// EInUnnest represents an IN UNNEST expression (was expr.InUnnest).
type EInUnnest struct {
	ExpressionBase
	Expr      Expr
	ArrayExpr Expr
	Negated   bool
	SpanVal   token.Span
}

// Span returns the source span for this expression.
func (i *EInUnnest) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EInUnnest) String() string {
	if i.Negated {
		return fmt.Sprintf("%s NOT IN UNNEST(%s)", i.Expr.String(), i.ArrayExpr.String())
	}
	return fmt.Sprintf("%s IN UNNEST(%s)", i.Expr.String(), i.ArrayExpr.String())
}

// EBetween represents a BETWEEN expression (was expr.Between).
type EBetween struct {
	ExpressionBase
	Expr    Expr
	Negated bool
	Low     Expr
	High    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (b *EBetween) Span() token.Span { return b.SpanVal }

// String returns the SQL representation.
func (b *EBetween) String() string {
	if b.Negated {
		return fmt.Sprintf("%s NOT BETWEEN %s AND %s", b.Expr.String(), b.Low.String(), b.High.String())
	}
	return fmt.Sprintf("%s BETWEEN %s AND %s", b.Expr.String(), b.Low.String(), b.High.String())
}

// ELike represents a LIKE expression (was expr.Like).
type ELike struct {
	ExpressionBase
	Negated    bool
	Any        bool // Snowflake ANY keyword
	Expr       Expr
	Pattern    Expr
	EscapeChar interface{} // ValueWithSpan
	SpanVal    token.Span
}

// Span returns the source span for this expression.
func (l *ELike) Span() token.Span { return l.SpanVal }

// String returns the SQL representation.
func (l *ELike) String() string {
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

// EILike represents an ILIKE (case-insensitive LIKE) expression (was expr.ILike).
type EILike struct {
	ExpressionBase
	Negated    bool
	Any        bool // Snowflake ANY keyword
	Expr       Expr
	Pattern    Expr
	EscapeChar interface{} // ValueWithSpan
	SpanVal    token.Span
}

// Span returns the source span for this expression.
func (i *EILike) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EILike) String() string {
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

// ESimilarTo represents a SIMILAR TO regex expression (was expr.SimilarTo).
type ESimilarTo struct {
	ExpressionBase
	Negated    bool
	Expr       Expr
	Pattern    Expr
	EscapeChar interface{} // ValueWithSpan
	SpanVal    token.Span
}

// Span returns the source span for this expression.
func (s *ESimilarTo) Span() token.Span { return s.SpanVal }

// String returns the SQL representation.
func (s *ESimilarTo) String() string {
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

// ERLike represents an RLIKE/REGEXP expression (was expr.RLike).
type ERLike struct {
	ExpressionBase
	Negated bool
	Expr    Expr
	Pattern Expr
	Regexp  bool // true for REGEXP, false for RLIKE
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (r *ERLike) Span() token.Span { return r.SpanVal }

// String returns the SQL representation.
func (r *ERLike) String() string {
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

// ECastKind represents the kind of cast operation.
type ECastKind int

const (
	ECastStandard ECastKind = iota
	ECastTry
	ECastSafe
	ECastDoubleColon
)

// ECastFormat represents the format for CAST expressions.
type ECastFormat struct {
	Value           interface{} // ValueWithSpan
	ValueAtTimeZone *struct {
		Value    interface{} // ValueWithSpan
		TimeZone interface{} // ValueWithSpan
	}
}

// ECast represents a CAST expression (was expr.Cast).
type ECast struct {
	ExpressionBase
	Kind     ECastKind
	Expr     Expr
	DataType string
	Array    bool // MySQL-specific: CAST(... AS type ARRAY)
	Format   *ECastFormat
	SpanVal  token.Span
}

// Span returns the source span for this expression.
func (c *ECast) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *ECast) String() string {
	var sb strings.Builder

	switch c.Kind {
	case ECastStandard:
		sb.WriteString("CAST(")
	case ECastTry:
		sb.WriteString("TRY_CAST(")
	case ECastSafe:
		sb.WriteString("SAFE_CAST(")
	case ECastDoubleColon:
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

// EConvert represents a CONVERT expression (was expr.Convert).
type EConvert struct {
	ExpressionBase
	IsTry             bool
	Expr              Expr
	DataType          *string
	Charset           *ObjectName
	TargetBeforeValue bool // MSSQL syntax
	Styles            []Expr
	SpanVal           token.Span
}

// Span returns the source span for this expression.
func (c *EConvert) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *EConvert) String() string {
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

// ECollate represents a COLLATE expression (was expr.Collate).
type ECollate struct {
	ExpressionBase
	Expr      Expr
	Collation *ObjectName
	SpanVal   token.Span
}

// Span returns the source span for this expression.
func (c *ECollate) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *ECollate) String() string {
	return fmt.Sprintf("%s COLLATE %s", c.Expr.String(), c.Collation.String())
}

// EAnyOp represents an ANY/SOME subquery comparison expression (was expr.AnyOp).
type EAnyOp struct {
	ExpressionBase
	Left      Expr
	CompareOp BinaryOperator
	Right     Expr
	IsSome    bool // ANY and SOME are synonymous
	SpanVal   token.Span
}

// Span returns the source span for this expression.
func (a *EAnyOp) Span() token.Span { return a.SpanVal }

// String returns the SQL representation.
func (a *EAnyOp) String() string {
	opName := "ANY"
	if a.IsSome {
		opName = "SOME"
	}

	_, isSubquery := a.Right.(*ESubquery)
	if isSubquery {
		return fmt.Sprintf("%s %s %s %s", a.Left.String(), a.CompareOp.String(), opName, a.Right.String())
	}
	return fmt.Sprintf("%s %s %s(%s)", a.Left.String(), a.CompareOp.String(), opName, a.Right.String())
}

// EAllOp represents an ALL subquery comparison expression (was expr.AllOp).
type EAllOp struct {
	ExpressionBase
	Left      Expr
	CompareOp BinaryOperator
	Right     Expr
	SpanVal   token.Span
}

// Span returns the source span for this expression.
func (a *EAllOp) Span() token.Span { return a.SpanVal }

// String returns the SQL representation.
func (a *EAllOp) String() string {
	_, isSubquery := a.Right.(*ESubquery)
	if isSubquery {
		return fmt.Sprintf("%s %s ALL %s", a.Left.String(), a.CompareOp.String(), a.Right.String())
	}
	return fmt.Sprintf("%s %s ALL(%s)", a.Left.String(), a.CompareOp.String(), a.Right.String())
}

// ENormalizationForm represents the Unicode normalization form.
type ENormalizationForm int

const (
	ENFC ENormalizationForm = iota
	ENFD
	ENFKC
	ENFKD
)

// String returns the SQL representation.
func (n ENormalizationForm) String() string {
	switch n {
	case ENFC:
		return "NFC"
	case ENFD:
		return "NFD"
	case ENFKC:
		return "NFKC"
	case ENFKD:
		return "NFKD"
	}
	return ""
}

// EIsNormalized represents an IS [NOT] [form] NORMALIZED expression (was expr.IsNormalized).
type EIsNormalized struct {
	ExpressionBase
	Expr    Expr
	Form    *ENormalizationForm
	Negated bool
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (i *EIsNormalized) Span() token.Span { return i.SpanVal }

// String returns the SQL representation.
func (i *EIsNormalized) String() string {
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

// EExtractSyntax represents the syntax for EXTRACT expressions.
type EExtractSyntax int

const (
	EExtractFrom EExtractSyntax = iota
	EExtractComma
)

// EExtract represents an EXTRACT expression (was expr.Extract).
type EExtract struct {
	ExpressionBase
	Field   string // DateTimeField
	Syntax  EExtractSyntax
	Expr    Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (e *EExtract) Span() token.Span { return e.SpanVal }

// String returns the SQL representation.
func (e *EExtract) String() string {
	if e.Syntax == EExtractFrom {
		return fmt.Sprintf("EXTRACT(%s FROM %s)", e.Field, e.Expr.String())
	}
	return fmt.Sprintf("EXTRACT(%s, %s)", e.Field, e.Expr.String())
}

// ECeilFloorKind represents the kind for CEIL/FLOOR expressions.
type ECeilFloorKind int

const (
	ECeilFloorDateTime ECeilFloorKind = iota
	ECeilFloorScale
)

// ECeilExpr represents a CEIL expression (was expr.CeilExpr).
type ECeilExpr struct {
	ExpressionBase
	Expr  Expr
	Field struct {
		Kind          ECeilFloorKind
		DateTimeField *string
		Scale         interface{} // ValueWithSpan
	}
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (c *ECeilExpr) Span() token.Span { return c.SpanVal }

// String returns the SQL representation.
func (c *ECeilExpr) String() string {
	if c.Field.Kind == ECeilFloorDateTime && c.Field.DateTimeField != nil {
		return fmt.Sprintf("CEIL(%s TO %s)", c.Expr.String(), *c.Field.DateTimeField)
	}
	if c.Field.Kind == ECeilFloorScale {
		return fmt.Sprintf("CEIL(%s, %v)", c.Expr.String(), c.Field.Scale)
	}
	return fmt.Sprintf("CEIL(%s)", c.Expr.String())
}

// EFloorExpr represents a FLOOR expression (was expr.FloorExpr).
type EFloorExpr struct {
	ExpressionBase
	Expr  Expr
	Field struct {
		Kind          ECeilFloorKind
		DateTimeField *string
		Scale         interface{} // ValueWithSpan
	}
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (f *EFloorExpr) Span() token.Span { return f.SpanVal }

// String returns the SQL representation.
func (f *EFloorExpr) String() string {
	if f.Field.Kind == ECeilFloorDateTime && f.Field.DateTimeField != nil {
		return fmt.Sprintf("FLOOR(%s TO %s)", f.Expr.String(), *f.Field.DateTimeField)
	}
	if f.Field.Kind == ECeilFloorScale {
		return fmt.Sprintf("FLOOR(%s, %v)", f.Expr.String(), f.Field.Scale)
	}
	return fmt.Sprintf("FLOOR(%s)", f.Expr.String())
}

// EPositionExpr represents a POSITION expression (was expr.PositionExpr).
type EPositionExpr struct {
	ExpressionBase
	Expr    Expr
	In      Expr
	SpanVal token.Span
}

// Span returns the source span for this expression.
func (p *EPositionExpr) Span() token.Span { return p.SpanVal }

// String returns the SQL representation.
func (p *EPositionExpr) String() string {
	return fmt.Sprintf("POSITION(%s IN %s)", p.Expr.String(), p.In.String())
}

// ESubstring represents a SUBSTRING/SUBSTR expression (was expr.Substring).
type ESubstring struct {
	ExpressionBase
	Expr          Expr
	SubstringFrom *Expr
	SubstringFor  *Expr
	Special       bool // true for SUBSTRING(expr, start, len) syntax
	Shorthand     bool // true for SUBSTR shorthand
	SpanVal       token.Span
}

// Span returns the source span for this expression.
func (s *ESubstring) Span() token.Span { return s.SpanVal }

// String returns the SQL representation.
func (s *ESubstring) String() string {
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

// ETrimWhere represents the trim direction.
type ETrimWhere int

const (
	ETrimBoth ETrimWhere = iota
	ETrimLeading
	ETrimTrailing
)

// String returns the SQL representation.
func (t ETrimWhere) String() string {
	switch t {
	case ETrimBoth:
		return "BOTH"
	case ETrimLeading:
		return "LEADING"
	case ETrimTrailing:
		return "TRAILING"
	}
	return ""
}

// ETrimExpr represents a TRIM expression (was expr.TrimExpr).
type ETrimExpr struct {
	ExpressionBase
	TrimWhere      *ETrimWhere
	TrimWhat       *Expr
	Expr           Expr
	TrimCharacters []Expr
	SpanVal        token.Span
}

// Span returns the source span for this expression.
func (t *ETrimExpr) Span() token.Span { return t.SpanVal }

// String returns the SQL representation.
func (t *ETrimExpr) String() string {
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

// EOverlayExpr represents an OVERLAY expression (was expr.OverlayExpr).
type EOverlayExpr struct {
	ExpressionBase
	Expr        Expr
	OverlayWhat Expr
	OverlayFrom Expr
	OverlayFor  *Expr
	SpanVal     token.Span
}

// Span returns the source span for this expression.
func (o *EOverlayExpr) Span() token.Span { return o.SpanVal }

// String returns the SQL representation.
func (o *EOverlayExpr) String() string {
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

// EAtTimeZone represents an AT TIME ZONE expression (was expr.AtTimeZone).
type EAtTimeZone struct {
	ExpressionBase
	Timestamp Expr
	TimeZone  Expr
	SpanVal   token.Span
}

// Span returns the source span for this expression.
func (a *EAtTimeZone) Span() token.Span { return a.SpanVal }

// String returns the SQL representation.
func (a *EAtTimeZone) String() string {
	return fmt.Sprintf("%s AT TIME ZONE %s", a.Timestamp.String(), a.TimeZone.String())
}
