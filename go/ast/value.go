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
	"math/big"
	"strings"

	"github.com/user/sqlparser/span"
)

// Value represents a primitive SQL value (literal).
// This includes numbers, strings, booleans, null, and placeholders.
type Value struct {
	BaseNode
	// Kind indicates the type of value stored.
	Kind ValueKind
	// Data holds the actual value data based on Kind.
	Data ValueData
}

// ValueKind enumerates the different types of SQL values.
type ValueKind int

const (
	// ValueKindNumber is a numeric literal.
	ValueKindNumber ValueKind = iota
	// ValueKindSingleQuotedString is a 'string value'.
	ValueKindSingleQuotedString
	// ValueKindDollarQuotedString is a $$...$$ or $tag$...$tag$ string (Postgres).
	ValueKindDollarQuotedString
	// ValueKindTripleSingleQuotedString is '''abc''' (BigQuery).
	ValueKindTripleSingleQuotedString
	// ValueKindTripleDoubleQuotedString is """abc""" (BigQuery).
	ValueKindTripleDoubleQuotedString
	// ValueKindEscapedStringLiteral is e'string' (Postgres extension).
	ValueKindEscapedStringLiteral
	// ValueKindUnicodeStringLiteral is u&'string' (Postgres extension).
	ValueKindUnicodeStringLiteral
	// ValueKindSingleQuotedByteStringLiteral is B'string'.
	ValueKindSingleQuotedByteStringLiteral
	// ValueKindDoubleQuotedByteStringLiteral is B"string".
	ValueKindDoubleQuotedByteStringLiteral
	// ValueKindTripleSingleQuotedByteStringLiteral is B'''abc''' (BigQuery).
	ValueKindTripleSingleQuotedByteStringLiteral
	// ValueKindTripleDoubleQuotedByteStringLiteral is B"""abc""" (BigQuery).
	ValueKindTripleDoubleQuotedByteStringLiteral
	// ValueKindSingleQuotedRawStringLiteral is R'abc' (BigQuery).
	ValueKindSingleQuotedRawStringLiteral
	// ValueKindDoubleQuotedRawStringLiteral is R"abc" (BigQuery).
	ValueKindDoubleQuotedRawStringLiteral
	// ValueKindTripleSingleQuotedRawStringLiteral is R'''abc''' (BigQuery).
	ValueKindTripleSingleQuotedRawStringLiteral
	// ValueKindTripleDoubleQuotedRawStringLiteral is R"""abc""" (BigQuery).
	ValueKindTripleDoubleQuotedRawStringLiteral
	// ValueKindNationalStringLiteral is N'string'.
	ValueKindNationalStringLiteral
	// ValueKindQuoteDelimitedStringLiteral is Q'{abc}' (Oracle).
	ValueKindQuoteDelimitedStringLiteral
	// ValueKindNationalQuoteDelimitedStringLiteral is NQ'{abc}' (Oracle).
	ValueKindNationalQuoteDelimitedStringLiteral
	// ValueKindHexStringLiteral is X'hex'.
	ValueKindHexStringLiteral
	// ValueKindDoubleQuotedString is a "string".
	ValueKindDoubleQuotedString
	// ValueKindBoolean is true/false.
	ValueKindBoolean
	// ValueKindNull is NULL.
	ValueKindNull
	// ValueKindPlaceholder is ? or $N.
	ValueKindPlaceholder
)

// ValueData holds the actual value data.
// Only one field should be set based on the ValueKind.
type ValueData struct {
	// For ValueKindNumber.
	NumberValue *big.Rat
	// LongSuffix indicates if the number has an L suffix.
	LongSuffix bool
	// For string values (most ValueKind types).
	StringValue string
	// For ValueKindDollarQuotedString.
	DollarQuotedString *DollarQuotedStringData
	// For ValueKindQuoteDelimitedStringLiteral.
	QuoteDelimitedString *QuoteDelimitedStringData
	// For ValueKindBoolean.
	BoolValue bool
	// For ValueKindPlaceholder.
	PlaceholderValue string
}

// DollarQuotedStringData represents a $$...$$ or $tag$...$tag$ string.
type DollarQuotedStringData struct {
	Value string
	Tag   *string
}

// QuoteDelimitedStringData represents a Q'...' string.
type QuoteDelimitedStringData struct {
	StartQuote rune
	Value      string
	EndQuote   rune
}

// String returns the SQL representation of the value.
func (v *Value) String() string {
	switch v.Kind {
	case ValueKindNumber:
		s := v.Data.NumberValue.FloatString(0)
		if v.Data.LongSuffix {
			return s + "L"
		}
		return s
	case ValueKindSingleQuotedString:
		return fmt.Sprintf("'%s'", EscapeSingleQuoteString(v.Data.StringValue))
	case ValueKindDoubleQuotedString:
		return fmt.Sprintf("\"%s\"", EscapeDoubleQuoteString(v.Data.StringValue))
	case ValueKindTripleSingleQuotedString:
		return fmt.Sprintf("'''%s'''", v.Data.StringValue)
	case ValueKindTripleDoubleQuotedString:
		return fmt.Sprintf("\"\"\"%s\"\"\"", v.Data.StringValue)
	case ValueKindDollarQuotedString:
		if v.Data.DollarQuotedString == nil {
			return "$$"
		}
		if v.Data.DollarQuotedString.Tag != nil {
			return fmt.Sprintf("$%s$%s$%s$", *v.Data.DollarQuotedString.Tag, v.Data.DollarQuotedString.Value, *v.Data.DollarQuotedString.Tag)
		}
		return fmt.Sprintf("$$%s$$", v.Data.DollarQuotedString.Value)
	case ValueKindEscapedStringLiteral:
		return fmt.Sprintf("E'%s'", EscapeEscapedString(v.Data.StringValue))
	case ValueKindUnicodeStringLiteral:
		return fmt.Sprintf("U&'%s'", EscapeUnicodeString(v.Data.StringValue))
	case ValueKindNationalStringLiteral:
		return fmt.Sprintf("N'%s'", v.Data.StringValue)
	case ValueKindQuoteDelimitedStringLiteral:
		if v.Data.QuoteDelimitedString == nil {
			return "Q''"
		}
		return fmt.Sprintf("Q'%c%s%c'", v.Data.QuoteDelimitedString.StartQuote, v.Data.QuoteDelimitedString.Value, v.Data.QuoteDelimitedString.EndQuote)
	case ValueKindNationalQuoteDelimitedStringLiteral:
		if v.Data.QuoteDelimitedString == nil {
			return "NQ''"
		}
		return fmt.Sprintf("NQ'%c%s%c'", v.Data.QuoteDelimitedString.StartQuote, v.Data.QuoteDelimitedString.Value, v.Data.QuoteDelimitedString.EndQuote)
	case ValueKindHexStringLiteral:
		return fmt.Sprintf("X'%s'", v.Data.StringValue)
	case ValueKindSingleQuotedByteStringLiteral:
		return fmt.Sprintf("B'%s'", v.Data.StringValue)
	case ValueKindDoubleQuotedByteStringLiteral:
		return fmt.Sprintf("B\"%s\"", v.Data.StringValue)
	case ValueKindTripleSingleQuotedByteStringLiteral:
		return fmt.Sprintf("B'''%s'''", v.Data.StringValue)
	case ValueKindTripleDoubleQuotedByteStringLiteral:
		return fmt.Sprintf("B\"\"\"%s\"\"\"", v.Data.StringValue)
	case ValueKindSingleQuotedRawStringLiteral:
		return fmt.Sprintf("R'%s'", v.Data.StringValue)
	case ValueKindDoubleQuotedRawStringLiteral:
		return fmt.Sprintf("R\"%s\"", v.Data.StringValue)
	case ValueKindTripleSingleQuotedRawStringLiteral:
		return fmt.Sprintf("R'''%s'''", v.Data.StringValue)
	case ValueKindTripleDoubleQuotedRawStringLiteral:
		return fmt.Sprintf("R\"\"\"%s\"\"\"", v.Data.StringValue)
	case ValueKindBoolean:
		if v.Data.BoolValue {
			return "TRUE"
		}
		return "FALSE"
	case ValueKindNull:
		return "NULL"
	case ValueKindPlaceholder:
		return v.Data.PlaceholderValue
	default:
		return ""
	}
}

// IntoString extracts the string value if the Value contains a string literal.
// Returns nil if the Value is not a string type.
func (v *Value) IntoString() *string {
	switch v.Kind {
	case ValueKindSingleQuotedString,
		ValueKindDoubleQuotedString,
		ValueKindTripleSingleQuotedString,
		ValueKindTripleDoubleQuotedString,
		ValueKindSingleQuotedByteStringLiteral,
		ValueKindDoubleQuotedByteStringLiteral,
		ValueKindTripleSingleQuotedByteStringLiteral,
		ValueKindTripleDoubleQuotedByteStringLiteral,
		ValueKindSingleQuotedRawStringLiteral,
		ValueKindDoubleQuotedRawStringLiteral,
		ValueKindTripleSingleQuotedRawStringLiteral,
		ValueKindTripleDoubleQuotedRawStringLiteral,
		ValueKindEscapedStringLiteral,
		ValueKindUnicodeStringLiteral,
		ValueKindNationalStringLiteral,
		ValueKindHexStringLiteral:
		return &v.Data.StringValue
	case ValueKindDollarQuotedString:
		if v.Data.DollarQuotedString != nil {
			return &v.Data.DollarQuotedString.Value
		}
	case ValueKindQuoteDelimitedStringLiteral,
		ValueKindNationalQuoteDelimitedStringLiteral:
		if v.Data.QuoteDelimitedString != nil {
			return &v.Data.QuoteDelimitedString.Value
		}
	}
	return nil
}

// IsString returns true if the Value is any kind of string literal.
func (v *Value) IsString() bool {
	return v.IntoString() != nil
}

// IsNumber returns true if the Value is a numeric literal.
func (v *Value) IsNumber() bool {
	return v.Kind == ValueKindNumber
}

// IsBoolean returns true if the Value is a boolean literal.
func (v *Value) IsBoolean() bool {
	return v.Kind == ValueKindBoolean
}

// IsNull returns true if the Value is NULL.
func (v *Value) IsNull() bool {
	return v.Kind == ValueKindNull
}

// IsPlaceholder returns true if the Value is a placeholder.
func (v *Value) IsPlaceholder() bool {
	return v.Kind == ValueKindPlaceholder
}

// Factory methods for creating Values

// NewNumber creates a numeric value from a string.
// The string should be a valid numeric representation.
func NewNumber(value string, long bool) (*Value, error) {
	rat := new(big.Rat)
	if _, ok := rat.SetString(value); !ok {
		return nil, fmt.Errorf("invalid number: %s", value)
	}
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindNumber,
		Data:     ValueData{NumberValue: rat, LongSuffix: long},
	}, nil
}

// NewSingleQuotedString creates a 'string' value.
func NewSingleQuotedString(value string) *Value {
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindSingleQuotedString,
		Data:     ValueData{StringValue: value},
	}
}

// NewDoubleQuotedString creates a "string" value.
func NewDoubleQuotedString(value string) *Value {
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindDoubleQuotedString,
		Data:     ValueData{StringValue: value},
	}
}

// NewBoolean creates a TRUE or FALSE value.
func NewBoolean(value bool) *Value {
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindBoolean,
		Data:     ValueData{BoolValue: value},
	}
}

// NewNull creates a NULL value.
func NewNull() *Value {
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindNull,
	}
}

// NewPlaceholder creates a placeholder value (? or $N).
func NewPlaceholder(value string) *Value {
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindPlaceholder,
		Data:     ValueData{PlaceholderValue: value},
	}
}

// NewDollarQuotedString creates a $$...$$ or $tag$...$tag$ value.
func NewDollarQuotedString(value string, tag *string) *Value {
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindDollarQuotedString,
		Data: ValueData{
			DollarQuotedString: &DollarQuotedStringData{
				Value: value,
				Tag:   tag,
			},
		},
	}
}

// NewHexStringLiteral creates an X'hex' value.
func NewHexStringLiteral(value string) *Value {
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindHexStringLiteral,
		Data:     ValueData{StringValue: value},
	}
}

// NewNationalStringLiteral creates an N'string' value.
func NewNationalStringLiteral(value string) *Value {
	return &Value{
		BaseNode: BaseNode{span: span.Span{}},
		Kind:     ValueKindNationalStringLiteral,
		Data:     ValueData{StringValue: value},
	}
}

// EscapeSingleQuoteString escapes single quotes by doubling them.
func EscapeSingleQuoteString(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// EscapeDoubleQuoteString escapes double quotes by doubling them.
func EscapeDoubleQuoteString(s string) string {
	return strings.ReplaceAll(s, "\"", "\"\"")
}

// EscapeEscapedString escapes special characters for E'...' strings.
// Handles backslash escapes like \, \n, \t, etc.
func EscapeEscapedString(s string) string {
	// For display purposes, we just double backslashes
	return strings.ReplaceAll(s, "\\", "\\\\")
}

// EscapeUnicodeString escapes special characters for U&'...' strings.
// This is a placeholder; proper Unicode escaping would be dialect-specific.
func EscapeUnicodeString(s string) string {
	return strings.ReplaceAll(s, "\\", "\\\\")
}

// ValueWithSpan pairs a Value with its source Span for location tracking.
type ValueWithSpan struct {
	Value *Value
	Span  span.Span
}

// String returns the SQL representation of the value.
func (v *ValueWithSpan) String() string {
	if v.Value == nil {
		return ""
	}
	return v.Value.String()
}

// IntoString extracts the string value if the Value contains a string literal.
func (v *ValueWithSpan) IntoString() *string {
	if v.Value == nil {
		return nil
	}
	return v.Value.IntoString()
}
