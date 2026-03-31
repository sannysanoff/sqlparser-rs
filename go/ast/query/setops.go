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
)

// SetOperator represents UNION, EXCEPT, INTERSECT, MINUS
type SetOperator int

const (
	SetOperatorUnion SetOperator = iota
	SetOperatorExcept
	SetOperatorIntersect
	SetOperatorMinus
)

func (s SetOperator) String() string {
	switch s {
	case SetOperatorUnion:
		return "UNION"
	case SetOperatorExcept:
		return "EXCEPT"
	case SetOperatorIntersect:
		return "INTERSECT"
	case SetOperatorMinus:
		return "MINUS"
	default:
		return ""
	}
}

// SetQuantifier represents quantifiers for set operations
type SetQuantifier int

const (
	SetQuantifierNone SetQuantifier = iota
	SetQuantifierAll
	SetQuantifierDistinct
	SetQuantifierByName
	SetQuantifierAllByName
	SetQuantifierDistinctByName
)

func (s SetQuantifier) String() string {
	switch s {
	case SetQuantifierAll:
		return "ALL"
	case SetQuantifierDistinct:
		return "DISTINCT"
	case SetQuantifierByName:
		return "BY NAME"
	case SetQuantifierAllByName:
		return "ALL BY NAME"
	case SetQuantifierDistinctByName:
		return "DISTINCT BY NAME"
	default:
		return ""
	}
}

// PivotValueSource represents the source of values in a PIVOT operation
type PivotValueSource interface {
	fmt.Stringer
}

// PivotValueList represents a static list of values
type PivotValueList struct {
	Values []ExprWithAlias
}

func (p *PivotValueList) String() string {
	parts := make([]string, len(p.Values))
	for i, v := range p.Values {
		parts[i] = v.String()
	}
	return strings.Join(parts, ", ")
}

// PivotValueAny represents ANY with optional ORDER BY
type PivotValueAny struct {
	OrderBy []OrderByExpr
}

func (p *PivotValueAny) String() string {
	if len(p.OrderBy) > 0 {
		parts := make([]string, len(p.OrderBy))
		for i, o := range p.OrderBy {
			parts[i] = o.String()
		}
		return "ANY ORDER BY " + strings.Join(parts, ", ")
	}
	return "ANY"
}

// PivotValueSubquery represents pivot on values from a subquery
type PivotValueSubquery struct {
	Query *Query
}

func (p *PivotValueSubquery) String() string {
	return p.Query.String()
}

// Measure represents an item in the MEASURES clause of MATCH_RECOGNIZE
type Measure struct {
	Expr  Expr
	Alias Ident
}

func (m *Measure) String() string {
	return fmt.Sprintf("%s AS %s", m.Expr.String(), m.Alias.String())
}

// RowsPerMatch represents the rows per match option in MATCH_RECOGNIZE
type RowsPerMatch struct {
	OneRow  bool
	AllRows *EmptyMatchesMode
}

func (r *RowsPerMatch) String() string {
	if r.OneRow {
		return "ONE ROW PER MATCH"
	}
	if r.AllRows != nil {
		return "ALL ROWS PER MATCH " + r.AllRows.String()
	}
	return "ALL ROWS PER MATCH"
}

// EmptyMatchesMode represents mode for handling empty matches
type EmptyMatchesMode int

const (
	EmptyMatchesShow EmptyMatchesMode = iota
	EmptyMatchesOmit
	EmptyMatchesWithUnmatched
)

func (e EmptyMatchesMode) String() string {
	switch e {
	case EmptyMatchesShow:
		return "SHOW EMPTY MATCHES"
	case EmptyMatchesOmit:
		return "OMIT EMPTY MATCHES"
	case EmptyMatchesWithUnmatched:
		return "WITH UNMATCHED ROWS"
	default:
		return ""
	}
}

// AfterMatchSkip represents where to continue after a match
type AfterMatchSkip int

const (
	AfterMatchSkipPastLastRow AfterMatchSkip = iota
	AfterMatchSkipToNextRow
	AfterMatchSkipToFirst
	AfterMatchSkipToLast
)

type AfterMatchSkipWithSymbol struct {
	SkipType AfterMatchSkip
	Symbol   Ident
}

func (a *AfterMatchSkipWithSymbol) String() string {
	switch a.SkipType {
	case AfterMatchSkipPastLastRow:
		return "AFTER MATCH SKIP PAST LAST ROW"
	case AfterMatchSkipToNextRow:
		return "AFTER MATCH SKIP TO NEXT ROW"
	case AfterMatchSkipToFirst:
		return fmt.Sprintf("AFTER MATCH SKIP TO FIRST %s", a.Symbol.String())
	case AfterMatchSkipToLast:
		return fmt.Sprintf("AFTER MATCH SKIP TO LAST %s", a.Symbol.String())
	default:
		return ""
	}
}

// SymbolDefinition represents a symbol defined in MATCH_RECOGNIZE
type SymbolDefinition struct {
	Symbol     Ident
	Definition Expr
}

func (s *SymbolDefinition) String() string {
	return fmt.Sprintf("%s AS %s", s.Symbol.String(), s.Definition.String())
}

// MatchRecognizeSymbol represents a symbol in a MATCH_RECOGNIZE pattern
type MatchRecognizeSymbol interface {
	fmt.Stringer
}

// MatchRecognizeNamedSymbol represents a named symbol like S1
type MatchRecognizeNamedSymbol struct {
	Name Ident
}

func (m *MatchRecognizeNamedSymbol) String() string {
	return m.Name.String()
}

// MatchRecognizeStartSymbol represents the start (^) virtual symbol
type MatchRecognizeStartSymbol struct{}

func (m *MatchRecognizeStartSymbol) String() string { return "^" }

// MatchRecognizeEndSymbol represents the end ($) virtual symbol
type MatchRecognizeEndSymbol struct{}

func (m *MatchRecognizeEndSymbol) String() string { return "$" }

// MatchRecognizePattern represents the pattern in MATCH_RECOGNIZE
type MatchRecognizePattern interface {
	fmt.Stringer
}

// MatchRecognizeSymbolPattern represents a simple symbol pattern
type MatchRecognizeSymbolPattern struct {
	Symbol MatchRecognizeSymbol
}

func (m *MatchRecognizeSymbolPattern) String() string {
	return m.Symbol.String()
}

// MatchRecognizeExcludePattern represents {- symbol -} pattern
type MatchRecognizeExcludePattern struct {
	Symbol MatchRecognizeSymbol
}

func (m *MatchRecognizeExcludePattern) String() string {
	return fmt.Sprintf("{- %s -}", m.Symbol.String())
}

// MatchRecognizePermutePattern represents PERMUTE(symbol_1, ..., symbol_n)
type MatchRecognizePermutePattern struct {
	Symbols []MatchRecognizeSymbol
}

func (m *MatchRecognizePermutePattern) String() string {
	parts := make([]string, len(m.Symbols))
	for i, s := range m.Symbols {
		parts[i] = s.String()
	}
	return fmt.Sprintf("PERMUTE(%s)", strings.Join(parts, ", "))
}

// MatchRecognizeConcatPattern represents pattern_1 pattern_2 ... pattern_n
type MatchRecognizeConcatPattern struct {
	Patterns []MatchRecognizePattern
}

func (m *MatchRecognizeConcatPattern) String() string {
	parts := make([]string, len(m.Patterns))
	for i, p := range m.Patterns {
		parts[i] = p.String()
	}
	return strings.Join(parts, " ")
}

// MatchRecognizeGroupPattern represents (pattern)
type MatchRecognizeGroupPattern struct {
	Pattern MatchRecognizePattern
}

func (m *MatchRecognizeGroupPattern) String() string {
	return fmt.Sprintf("( %s )", m.Pattern.String())
}

// MatchRecognizeAlternationPattern represents pattern_1 | pattern_2 | ...
type MatchRecognizeAlternationPattern struct {
	Patterns []MatchRecognizePattern
}

func (m *MatchRecognizeAlternationPattern) String() string {
	parts := make([]string, len(m.Patterns))
	for i, p := range m.Patterns {
		parts[i] = p.String()
	}
	return strings.Join(parts, " | ")
}

// MatchRecognizeRepetitionPattern represents pattern with quantifier
type MatchRecognizeRepetitionPattern struct {
	Pattern    MatchRecognizePattern
	Quantifier *RepetitionQuantifierWithValue
}

func (m *MatchRecognizeRepetitionPattern) String() string {
	if m.Quantifier != nil {
		return m.Pattern.String() + m.Quantifier.String()
	}
	return m.Pattern.String()
}

// RepetitionQuantifier represents pattern repetition quantifiers
type RepetitionQuantifier int

type RepetitionQuantifierWithValue struct {
	Type  RepetitionQuantifier
	Value uint32
	Min   uint32
	Max   uint32
}

const (
	QuantifierZeroOrMore RepetitionQuantifier = iota
	QuantifierOneOrMore
	QuantifierAtMostOne
	QuantifierExactly
	QuantifierAtLeast
	QuantifierAtMost
	QuantifierRange
)

func (r *RepetitionQuantifierWithValue) String() string {
	switch r.Type {
	case QuantifierZeroOrMore:
		return "*"
	case QuantifierOneOrMore:
		return "+"
	case QuantifierAtMostOne:
		return "?"
	case QuantifierExactly:
		return fmt.Sprintf("{%d}", r.Value)
	case QuantifierAtLeast:
		return fmt.Sprintf("{%d,}", r.Min)
	case QuantifierAtMost:
		return fmt.Sprintf("{,%d}", r.Value)
	case QuantifierRange:
		return fmt.Sprintf("{%d,%d}", r.Min, r.Max)
	default:
		return ""
	}
}
