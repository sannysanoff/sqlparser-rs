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

// Package datatype provides SQL data type definitions for the AST.
// This package implements all DataType variants from the sqlparser-rs crate.
package datatype

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/span"
)

// DataType represents a SQL data type.
// All concrete data type implementations must implement this interface.
type DataType interface {
	fmt.Stringer
	Span() span.Span
	dataTypeNode()
}

// Ensure all DataType implementations implement the interface
var (
	_ DataType = (*TableType)(nil)
	_ DataType = (*NamedTableType)(nil)
	_ DataType = (*CharacterType)(nil)
	_ DataType = (*CharType)(nil)
	_ DataType = (*CharacterVaryingType)(nil)
	_ DataType = (*CharVaryingType)(nil)
	_ DataType = (*VarcharType)(nil)
	_ DataType = (*NvarcharType)(nil)
	_ DataType = (*UuidType)(nil)
	_ DataType = (*CharacterLargeObjectType)(nil)
	_ DataType = (*CharLargeObjectType)(nil)
	_ DataType = (*ClobType)(nil)
	_ DataType = (*BinaryType)(nil)
	_ DataType = (*VarbinaryType)(nil)
	_ DataType = (*BlobType)(nil)
	_ DataType = (*TinyBlobType)(nil)
	_ DataType = (*MediumBlobType)(nil)
	_ DataType = (*LongBlobType)(nil)
	_ DataType = (*BytesType)(nil)
	_ DataType = (*NumericType)(nil)
	_ DataType = (*DecimalType)(nil)
	_ DataType = (*DecimalUnsignedType)(nil)
	_ DataType = (*DecType)(nil)
	_ DataType = (*DecUnsignedType)(nil)
	_ DataType = (*BigNumericType)(nil)
	_ DataType = (*BigDecimalType)(nil)
	_ DataType = (*FloatType)(nil)
	_ DataType = (*FloatUnsignedType)(nil)
	_ DataType = (*TinyIntType)(nil)
	_ DataType = (*TinyIntUnsignedType)(nil)
	_ DataType = (*UTinyIntType)(nil)
	_ DataType = (*Int2Type)(nil)
	_ DataType = (*Int2UnsignedType)(nil)
	_ DataType = (*SmallIntType)(nil)
	_ DataType = (*SmallIntUnsignedType)(nil)
	_ DataType = (*USmallIntType)(nil)
	_ DataType = (*MediumIntType)(nil)
	_ DataType = (*MediumIntUnsignedType)(nil)
	_ DataType = (*IntType)(nil)
	_ DataType = (*Int4Type)(nil)
	_ DataType = (*Int8Type)(nil)
	_ DataType = (*Int16Type)(nil)
	_ DataType = (*Int32Type)(nil)
	_ DataType = (*Int64Type)(nil)
	_ DataType = (*Int128Type)(nil)
	_ DataType = (*Int256Type)(nil)
	_ DataType = (*IntegerType)(nil)
	_ DataType = (*IntUnsignedType)(nil)
	_ DataType = (*Int4UnsignedType)(nil)
	_ DataType = (*IntegerUnsignedType)(nil)
	_ DataType = (*HugeIntType)(nil)
	_ DataType = (*UHugeIntType)(nil)
	_ DataType = (*UInt8Type)(nil)
	_ DataType = (*UInt16Type)(nil)
	_ DataType = (*UInt32Type)(nil)
	_ DataType = (*UInt64Type)(nil)
	_ DataType = (*UInt128Type)(nil)
	_ DataType = (*UInt256Type)(nil)
	_ DataType = (*BigIntType)(nil)
	_ DataType = (*BigIntUnsignedType)(nil)
	_ DataType = (*UBigIntType)(nil)
	_ DataType = (*Int8UnsignedType)(nil)
	_ DataType = (*SignedType)(nil)
	_ DataType = (*SignedIntegerType)(nil)
	_ DataType = (*UnsignedType)(nil)
	_ DataType = (*UnsignedIntegerType)(nil)
	_ DataType = (*RealType)(nil)
	_ DataType = (*RealUnsignedType)(nil)
	_ DataType = (*Float4Type)(nil)
	_ DataType = (*Float32Type)(nil)
	_ DataType = (*Float64Type)(nil)
	_ DataType = (*Float8Type)(nil)
	_ DataType = (*DoubleType)(nil)
	_ DataType = (*DoubleUnsignedType)(nil)
	_ DataType = (*DoublePrecisionType)(nil)
	_ DataType = (*DoublePrecisionUnsignedType)(nil)
	_ DataType = (*BoolType)(nil)
	_ DataType = (*BooleanType)(nil)
	_ DataType = (*DateType)(nil)
	_ DataType = (*Date32Type)(nil)
	_ DataType = (*TimeType)(nil)
	_ DataType = (*DatetimeType)(nil)
	_ DataType = (*Datetime64Type)(nil)
	_ DataType = (*TimestampType)(nil)
	_ DataType = (*TimestampNtzType)(nil)
	_ DataType = (*IntervalType)(nil)
	_ DataType = (*JSONType)(nil)
	_ DataType = (*JSONBType)(nil)
	_ DataType = (*RegclassType)(nil)
	_ DataType = (*TextType)(nil)
	_ DataType = (*TinyTextType)(nil)
	_ DataType = (*MediumTextType)(nil)
	_ DataType = (*LongTextType)(nil)
	_ DataType = (*StringType)(nil)
	_ DataType = (*FixedStringType)(nil)
	_ DataType = (*ByteaType)(nil)
	_ DataType = (*BitType)(nil)
	_ DataType = (*BitVaryingType)(nil)
	_ DataType = (*VarBitType)(nil)
	_ DataType = (*CustomType)(nil)
	_ DataType = (*ArrayType)(nil)
	_ DataType = (*MapType)(nil)
	_ DataType = (*TupleType)(nil)
	_ DataType = (*NestedType)(nil)
	_ DataType = (*EnumType)(nil)
	_ DataType = (*SetType)(nil)
	_ DataType = (*StructType)(nil)
	_ DataType = (*UnionType)(nil)
	_ DataType = (*NullableType)(nil)
	_ DataType = (*FixedStringDefType)(nil)
	_ DataType = (*LowCardinalityType)(nil)
	_ DataType = (*UnspecifiedType)(nil)
	_ DataType = (*TriggerType)(nil)
	_ DataType = (*AnyType)(nil)
	_ DataType = (*GeometricType)(nil)
	_ DataType = (*TsVectorType)(nil)
	_ DataType = (*TsQueryType)(nil)
)

// Supporting Types

// CharacterLength represents information about character length, including length and possibly unit.
type CharacterLength struct {
	// Length is the default (if VARYING) or maximum (if not VARYING) length
	Length uint64
	// Unit is the optional unit. If not informed, the ANSI handles it as CHARACTERS implicitly
	Unit *CharLengthUnits
}

// Max returns true if this represents VARCHAR(MAX) or NVARCHAR(MAX)
func (c CharacterLength) Max() bool {
	return c.Length == 0 && c.Unit == nil
}

// String returns the SQL representation
func (c CharacterLength) String() string {
	if c.Max() {
		return "MAX"
	}
	if c.Unit != nil {
		return fmt.Sprintf("%d %s", c.Length, c.Unit.String())
	}
	return fmt.Sprintf("%d", c.Length)
}

// CharLengthUnits represents possible units for characters.
type CharLengthUnits int

const (
	// CharactersUnit is the CHARACTERS unit
	CharactersUnit CharLengthUnits = iota
	// OctetsUnit is the OCTETS unit
	OctetsUnit
)

// String returns the SQL representation
func (c CharLengthUnits) String() string {
	switch c {
	case CharactersUnit:
		return "CHARACTERS"
	case OctetsUnit:
		return "OCTETS"
	}
	return ""
}

// BinaryLength represents information about binary length.
type BinaryLength struct {
	// Length is the default (if VARYING)
	Length uint64
	// Max is true for VARBINARY(MAX) used in T-SQL
	Max bool
}

// String returns the SQL representation
func (b BinaryLength) String() string {
	if b.Max {
		return "MAX"
	}
	return fmt.Sprintf("%d", b.Length)
}

// ExactNumberInfo represents additional information for exact numeric data types.
type ExactNumberInfo struct {
	// Precision is the optional precision
	Precision *uint64
	// Scale is the optional scale (only valid if Precision is set)
	Scale *int64
}

// String returns the SQL representation
func (e ExactNumberInfo) String() string {
	if e.Precision == nil {
		return ""
	}
	if e.Scale == nil {
		return fmt.Sprintf("(%d)", *e.Precision)
	}
	return fmt.Sprintf("(%d,%d)", *e.Precision, *e.Scale)
}

// HasPrecision returns true if precision is set
func (e ExactNumberInfo) HasPrecision() bool {
	return e.Precision != nil
}

// HasScale returns true if scale is set
func (e ExactNumberInfo) HasScale() bool {
	return e.Scale != nil
}

// TimezoneInfo represents timezone information for temporal types.
type TimezoneInfo int

const (
	// NoTimezoneInfo means no information about time zone, e.g. TIMESTAMP
	NoTimezoneInfo TimezoneInfo = iota
	// WithTimeZone means temporal type 'WITH TIME ZONE', e.g. TIMESTAMP WITH TIME ZONE
	WithTimeZone
	// WithoutTimeZone means temporal type 'WITHOUT TIME ZONE', e.g. TIME WITHOUT TIME ZONE
	WithoutTimeZone
	// Tz means Postgresql specific `WITH TIME ZONE` formatting, e.g. TIMETZ
	Tz
)

// String returns the SQL representation
func (t TimezoneInfo) String() string {
	switch t {
	case NoTimezoneInfo:
		return ""
	case WithTimeZone:
		return " WITH TIME ZONE"
	case WithoutTimeZone:
		return " WITHOUT TIME ZONE"
	case Tz:
		return "TZ"
	}
	return ""
}

// IntervalFields represents fields for PostgreSQL INTERVAL type.
type IntervalFields int

const (
	Year IntervalFields = iota
	Month
	Day
	Hour
	Minute
	Second
	YearToMonth
	DayToHour
	DayToMinute
	DayToSecond
	HourToMinute
	HourToSecond
	MinuteToSecond
)

// String returns the SQL representation
func (i IntervalFields) String() string {
	switch i {
	case Year:
		return "YEAR"
	case Month:
		return "MONTH"
	case Day:
		return "DAY"
	case Hour:
		return "HOUR"
	case Minute:
		return "MINUTE"
	case Second:
		return "SECOND"
	case YearToMonth:
		return "YEAR TO MONTH"
	case DayToHour:
		return "DAY TO HOUR"
	case DayToMinute:
		return "DAY TO MINUTE"
	case DayToSecond:
		return "DAY TO SECOND"
	case HourToMinute:
		return "HOUR TO MINUTE"
	case HourToSecond:
		return "HOUR TO SECOND"
	case MinuteToSecond:
		return "MINUTE TO SECOND"
	}
	return ""
}

// StructBracketKind represents the type of brackets used for STRUCT literals.
type StructBracketKind int

const (
	// Parentheses brackets: STRUCT(a INT, b STRING)
	Parentheses StructBracketKind = iota
	// AngleBrackets brackets: STRUCT<a INT, b STRING>
	AngleBrackets
)

// ArrayElemTypeDef represents the data type of elements in an array.
type ArrayElemTypeDef struct {
	// Style indicates the array syntax style used
	Style ArrayStyle
	// DataType is the element type (nil for None style)
	DataType DataType
	// Size is the optional array size (for SquareBracket style)
	Size *uint64
}

// ArrayStyle represents the syntax style for array type definitions.
type ArrayStyle int

const (
	// ArrayNone means use `ARRAY` style without an explicit element type
	ArrayNone ArrayStyle = iota
	// ArrayAngleBracket means angle-bracket style, e.g. `ARRAY<INT>`
	ArrayAngleBracket
	// ArraySquareBracket means square-bracket style, e.g. `INT[]` or `INT[2]`
	ArraySquareBracket
	// ArrayParenthesis means parenthesis style, e.g. `Array(Int64)`
	ArrayParenthesis
)

// String returns the SQL representation
func (a ArrayElemTypeDef) String() string {
	switch a.Style {
	case ArrayNone:
		return "ARRAY"
	case ArrayAngleBracket:
		if a.DataType != nil {
			return fmt.Sprintf("ARRAY<%s>", a.DataType.String())
		}
		return "ARRAY<>"
	case ArraySquareBracket:
		if a.DataType != nil {
			if a.Size != nil {
				return fmt.Sprintf("%s[%d]", a.DataType.String(), *a.Size)
			}
			return fmt.Sprintf("%s[]", a.DataType.String())
		}
		return "[]"
	case ArrayParenthesis:
		if a.DataType != nil {
			return fmt.Sprintf("Array(%s)", a.DataType.String())
		}
		return "Array()"
	}
	return "ARRAY"
}

// GeometricTypeKind represents different types of geometric shapes.
type GeometricTypeKind int

const (
	Point GeometricTypeKind = iota
	Line
	LineSegment
	GeometricBox
	GeometricPath
	Polygon
	Circle
)

// String returns the SQL representation
func (g GeometricTypeKind) String() string {
	switch g {
	case Point:
		return "point"
	case Line:
		return "line"
	case LineSegment:
		return "lseg"
	case GeometricBox:
		return "box"
	case GeometricPath:
		return "path"
	case Polygon:
		return "polygon"
	case Circle:
		return "circle"
	}
	return ""
}

// EnumMember represents a member of an ENUM type.
type EnumMember struct {
	// Name is the enum value name
	Name string
	// Value is the optional integer value (ClickHouse style: 'name' = value)
	Value expr.Expr
}

// String returns the SQL representation
func (e EnumMember) String() string {
	if e.Value != nil {
		return fmt.Sprintf("'%s' = %s", escapeSingleQuote(e.Name), e.Value.String())
	}
	return fmt.Sprintf("'%s'", escapeSingleQuote(e.Name))
}

// escapeSingleQuote escapes single quotes in a string
func escapeSingleQuote(s string) string {
	return strings.ReplaceAll(s, "'", "''")
}

// StructField represents a field definition within a struct.
type StructField struct {
	// FieldName is the optional name of the struct field
	FieldName *expr.Ident
	// FieldType is the field data type
	FieldType DataType
	// Options are struct field options (e.g., `OPTIONS(...)` on BigQuery)
	Options []*expr.SqlOption
}

// String returns the SQL representation
func (s StructField) String() string {
	var sb strings.Builder
	if s.FieldName != nil {
		sb.WriteString(s.FieldName.String())
		sb.WriteString(" ")
	}
	sb.WriteString(s.FieldType.String())
	if len(s.Options) > 0 {
		options := make([]string, len(s.Options))
		for i, opt := range s.Options {
			options[i] = opt.String()
		}
		sb.WriteString(fmt.Sprintf(" OPTIONS(%s)", strings.Join(options, ", ")))
	}
	return sb.String()
}

// UnionField represents a field definition within a union.
type UnionField struct {
	// FieldName is the name of the union field
	FieldName *expr.Ident
	// FieldType is the type of the union field
	FieldType DataType
}

// String returns the SQL representation
func (u UnionField) String() string {
	return fmt.Sprintf("%s %s", u.FieldName.String(), u.FieldType.String())
}

// ColumnDef represents a column definition.
type ColumnDef struct {
	// Name is the column name
	Name *expr.Ident
	// DataType is the column data type
	DataType DataType
	// Collation is the optional collation
	Collation *expr.ObjectName
	// Options are column options
	Options []*expr.ColumnOption
}

// String returns the SQL representation
func (c ColumnDef) String() string {
	var sb strings.Builder
	sb.WriteString(c.Name.String())
	sb.WriteString(" ")
	sb.WriteString(c.DataType.String())
	if c.Collation != nil {
		sb.WriteString(" COLLATE ")
		sb.WriteString(c.Collation.String())
	}
	for _, opt := range c.Options {
		sb.WriteString(" ")
		sb.WriteString(opt.String())
	}
	return sb.String()
}

// TableType represents a table type in PostgreSQL/MS SQL Server.
type TableType struct {
	SpanVal   span.Span
	Columns   []*ColumnDef
	HasFields bool
}

func (t *TableType) Span() span.Span { return t.SpanVal }
func (t *TableType) dataTypeNode()   {}
func (t *TableType) String() string {
	if !t.HasFields {
		return "TABLE"
	}
	cols := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		cols[i] = col.String()
	}
	return fmt.Sprintf("TABLE(%s)", strings.Join(cols, ", "))
}

// NamedTableType represents a named table type, e.g. CREATE FUNCTION RETURNS @result TABLE(...).
type NamedTableType struct {
	SpanVal span.Span
	Name    *expr.ObjectName
	Columns []*ColumnDef
}

func (t *NamedTableType) Span() span.Span { return t.SpanVal }
func (t *NamedTableType) dataTypeNode()   {}
func (t *NamedTableType) String() string {
	cols := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		cols[i] = col.String()
	}
	return fmt.Sprintf("%s TABLE (%s)", t.Name.String(), strings.Join(cols, ", "))
}

// CharacterType represents a fixed-length character type, e.g. CHARACTER(10).
type CharacterType struct {
	SpanVal span.Span
	Length  *CharacterLength
}

func (t *CharacterType) Span() span.Span { return t.SpanVal }
func (t *CharacterType) dataTypeNode()   {}
func (t *CharacterType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHARACTER(%s)", t.Length.String())
	}
	return "CHARACTER"
}

// CharType represents a fixed-length char type, e.g. CHAR(10).
type CharType struct {
	SpanVal span.Span
	Length  *CharacterLength
}

func (t *CharType) Span() span.Span { return t.SpanVal }
func (t *CharType) dataTypeNode()   {}
func (t *CharType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHAR(%s)", t.Length.String())
	}
	return "CHAR"
}

// CharacterVaryingType represents a character varying type, e.g. CHARACTER VARYING(10).
type CharacterVaryingType struct {
	SpanVal span.Span
	Length  *CharacterLength
}

func (t *CharacterVaryingType) Span() span.Span { return t.SpanVal }
func (t *CharacterVaryingType) dataTypeNode()   {}
func (t *CharacterVaryingType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHARACTER VARYING(%s)", t.Length.String())
	}
	return "CHARACTER VARYING"
}

// CharVaryingType represents a char varying type, e.g. CHAR VARYING(10).
type CharVaryingType struct {
	SpanVal span.Span
	Length  *CharacterLength
}

func (t *CharVaryingType) Span() span.Span { return t.SpanVal }
func (t *CharVaryingType) dataTypeNode()   {}
func (t *CharVaryingType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHAR VARYING(%s)", t.Length.String())
	}
	return "CHAR VARYING"
}

// VarcharType represents a variable-length character type, e.g. VARCHAR(10).
type VarcharType struct {
	SpanVal span.Span
	Length  *CharacterLength
}

func (t *VarcharType) Span() span.Span { return t.SpanVal }
func (t *VarcharType) dataTypeNode()   {}
func (t *VarcharType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("VARCHAR(%s)", t.Length.String())
	}
	return "VARCHAR"
}

// NvarcharType represents a variable-length character type, e.g. NVARCHAR(10).
type NvarcharType struct {
	SpanVal span.Span
	Length  *CharacterLength
}

func (t *NvarcharType) Span() span.Span { return t.SpanVal }
func (t *NvarcharType) dataTypeNode()   {}
func (t *NvarcharType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("NVARCHAR(%s)", t.Length.String())
	}
	return "NVARCHAR"
}

// UuidType represents a UUID type.
type UuidType struct {
	SpanVal span.Span
}

func (t *UuidType) Span() span.Span { return t.SpanVal }
func (t *UuidType) dataTypeNode()   {}
func (t *UuidType) String() string  { return "UUID" }

// CharacterLargeObjectType represents a large character object type.
type CharacterLargeObjectType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *CharacterLargeObjectType) Span() span.Span { return t.SpanVal }
func (t *CharacterLargeObjectType) dataTypeNode()   {}
func (t *CharacterLargeObjectType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHARACTER LARGE OBJECT(%d)", *t.Length)
	}
	return "CHARACTER LARGE OBJECT"
}

// CharLargeObjectType represents a large character object type.
type CharLargeObjectType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *CharLargeObjectType) Span() span.Span { return t.SpanVal }
func (t *CharLargeObjectType) dataTypeNode()   {}
func (t *CharLargeObjectType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHAR LARGE OBJECT(%d)", *t.Length)
	}
	return "CHAR LARGE OBJECT"
}

// ClobType represents a large character object type.
type ClobType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *ClobType) Span() span.Span { return t.SpanVal }
func (t *ClobType) dataTypeNode()   {}
func (t *ClobType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CLOB(%d)", *t.Length)
	}
	return "CLOB"
}

// BinaryType represents a fixed-length binary type.
type BinaryType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *BinaryType) Span() span.Span { return t.SpanVal }
func (t *BinaryType) dataTypeNode()   {}
func (t *BinaryType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("BINARY(%d)", *t.Length)
	}
	return "BINARY"
}

// VarbinaryType represents a variable-length binary type.
type VarbinaryType struct {
	SpanVal span.Span
	Length  *BinaryLength
}

func (t *VarbinaryType) Span() span.Span { return t.SpanVal }
func (t *VarbinaryType) dataTypeNode()   {}
func (t *VarbinaryType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("VARBINARY(%s)", t.Length.String())
	}
	return "VARBINARY"
}

// BlobType represents a large binary object type.
type BlobType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *BlobType) Span() span.Span { return t.SpanVal }
func (t *BlobType) dataTypeNode()   {}
func (t *BlobType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("BLOB(%d)", *t.Length)
	}
	return "BLOB"
}

// TinyBlobType represents a MySQL tiny blob type.
type TinyBlobType struct {
	SpanVal span.Span
}

func (t *TinyBlobType) Span() span.Span { return t.SpanVal }
func (t *TinyBlobType) dataTypeNode()   {}
func (t *TinyBlobType) String() string  { return "TINYBLOB" }

// MediumBlobType represents a MySQL medium blob type.
type MediumBlobType struct {
	SpanVal span.Span
}

func (t *MediumBlobType) Span() span.Span { return t.SpanVal }
func (t *MediumBlobType) dataTypeNode()   {}
func (t *MediumBlobType) String() string  { return "MEDIUMBLOB" }

// LongBlobType represents a MySQL long blob type.
type LongBlobType struct {
	SpanVal span.Span
}

func (t *LongBlobType) Span() span.Span { return t.SpanVal }
func (t *LongBlobType) dataTypeNode()   {}
func (t *LongBlobType) String() string  { return "LONGBLOB" }

// BytesType represents a variable-length binary data type (BigQuery).
type BytesType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *BytesType) Span() span.Span { return t.SpanVal }
func (t *BytesType) dataTypeNode()   {}
func (t *BytesType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("BYTES(%d)", *t.Length)
	}
	return "BYTES"
}

// NumericType represents a numeric type with optional precision and scale.
type NumericType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *NumericType) Span() span.Span { return t.SpanVal }
func (t *NumericType) dataTypeNode()   {}
func (t *NumericType) String() string  { return "NUMERIC" + t.Info.String() }

// DecimalType represents a decimal type with optional precision and scale.
type DecimalType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *DecimalType) Span() span.Span { return t.SpanVal }
func (t *DecimalType) dataTypeNode()   {}
func (t *DecimalType) String() string  { return "DECIMAL" + t.Info.String() }

// DecimalUnsignedType represents a MySQL unsigned decimal type.
type DecimalUnsignedType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *DecimalUnsignedType) Span() span.Span { return t.SpanVal }
func (t *DecimalUnsignedType) dataTypeNode()   {}
func (t *DecimalUnsignedType) String() string  { return "DECIMAL" + t.Info.String() + " UNSIGNED" }

// DecType represents a dec type with optional precision and scale.
type DecType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *DecType) Span() span.Span { return t.SpanVal }
func (t *DecType) dataTypeNode()   {}
func (t *DecType) String() string  { return "DEC" + t.Info.String() }

// DecUnsignedType represents a MySQL unsigned dec type.
type DecUnsignedType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *DecUnsignedType) Span() span.Span { return t.SpanVal }
func (t *DecUnsignedType) dataTypeNode()   {}
func (t *DecUnsignedType) String() string  { return "DEC" + t.Info.String() + " UNSIGNED" }

// BigNumericType represents a BigQuery BigNumeric type.
type BigNumericType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *BigNumericType) Span() span.Span { return t.SpanVal }
func (t *BigNumericType) dataTypeNode()   {}
func (t *BigNumericType) String() string  { return "BIGNUMERIC" + t.Info.String() }

// BigDecimalType represents a BigQuery BigDecimal type.
type BigDecimalType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *BigDecimalType) Span() span.Span { return t.SpanVal }
func (t *BigDecimalType) dataTypeNode()   {}
func (t *BigDecimalType) String() string  { return "BIGDECIMAL" + t.Info.String() }

// FloatType represents a floating point type with optional precision and scale.
type FloatType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *FloatType) Span() span.Span { return t.SpanVal }
func (t *FloatType) dataTypeNode()   {}
func (t *FloatType) String() string  { return "FLOAT" + t.Info.String() }

// FloatUnsignedType represents a MySQL unsigned float type.
type FloatUnsignedType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *FloatUnsignedType) Span() span.Span { return t.SpanVal }
func (t *FloatUnsignedType) dataTypeNode()   {}
func (t *FloatUnsignedType) String() string  { return "FLOAT" + t.Info.String() + " UNSIGNED" }

// TinyIntType represents a tiny integer with optional display width.
type TinyIntType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *TinyIntType) Span() span.Span { return t.SpanVal }
func (t *TinyIntType) dataTypeNode()   {}
func (t *TinyIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("TINYINT(%d)", *t.DisplayWidth)
	}
	return "TINYINT"
}

// TinyIntUnsignedType represents a MySQL unsigned tiny integer.
type TinyIntUnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *TinyIntUnsignedType) Span() span.Span { return t.SpanVal }
func (t *TinyIntUnsignedType) dataTypeNode()   {}
func (t *TinyIntUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("TINYINT(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "TINYINT UNSIGNED"
}

// UTinyIntType represents a MySQL unsigned tiny integer (UTINYINT).
type UTinyIntType struct {
	SpanVal span.Span
}

func (t *UTinyIntType) Span() span.Span { return t.SpanVal }
func (t *UTinyIntType) dataTypeNode()   {}
func (t *UTinyIntType) String() string  { return "UTINYINT" }

// Int2Type represents an Int2 alias for SmallInt in PostgreSQL.
type Int2Type struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *Int2Type) Span() span.Span { return t.SpanVal }
func (t *Int2Type) dataTypeNode()   {}
func (t *Int2Type) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT2(%d)", *t.DisplayWidth)
	}
	return "INT2"
}

// Int2UnsignedType represents a MySQL unsigned Int2.
type Int2UnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *Int2UnsignedType) Span() span.Span { return t.SpanVal }
func (t *Int2UnsignedType) dataTypeNode()   {}
func (t *Int2UnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT2(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "INT2 UNSIGNED"
}

// SmallIntType represents a small integer with optional display width.
type SmallIntType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *SmallIntType) Span() span.Span { return t.SpanVal }
func (t *SmallIntType) dataTypeNode()   {}
func (t *SmallIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("SMALLINT(%d)", *t.DisplayWidth)
	}
	return "SMALLINT"
}

// SmallIntUnsignedType represents a MySQL unsigned small integer.
type SmallIntUnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *SmallIntUnsignedType) Span() span.Span { return t.SpanVal }
func (t *SmallIntUnsignedType) dataTypeNode()   {}
func (t *SmallIntUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("SMALLINT(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "SMALLINT UNSIGNED"
}

// USmallIntType represents a MySQL unsigned small integer (USMALLINT).
type USmallIntType struct {
	SpanVal span.Span
}

func (t *USmallIntType) Span() span.Span { return t.SpanVal }
func (t *USmallIntType) dataTypeNode()   {}
func (t *USmallIntType) String() string  { return "USMALLINT" }

// MediumIntType represents a MySQL medium integer with optional display width.
type MediumIntType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *MediumIntType) Span() span.Span { return t.SpanVal }
func (t *MediumIntType) dataTypeNode()   {}
func (t *MediumIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("MEDIUMINT(%d)", *t.DisplayWidth)
	}
	return "MEDIUMINT"
}

// MediumIntUnsignedType represents a MySQL unsigned medium integer.
type MediumIntUnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *MediumIntUnsignedType) Span() span.Span { return t.SpanVal }
func (t *MediumIntUnsignedType) dataTypeNode()   {}
func (t *MediumIntUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("MEDIUMINT(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "MEDIUMINT UNSIGNED"
}

// IntType represents an integer with optional display width.
type IntType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *IntType) Span() span.Span { return t.SpanVal }
func (t *IntType) dataTypeNode()   {}
func (t *IntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT(%d)", *t.DisplayWidth)
	}
	return "INT"
}

// Int4Type represents an Int4 alias for Integer in PostgreSQL.
type Int4Type struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *Int4Type) Span() span.Span { return t.SpanVal }
func (t *Int4Type) dataTypeNode()   {}
func (t *Int4Type) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT4(%d)", *t.DisplayWidth)
	}
	return "INT4"
}

// Int8Type represents an Int8 alias for BigInt in PostgreSQL.
type Int8Type struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *Int8Type) Span() span.Span { return t.SpanVal }
func (t *Int8Type) dataTypeNode()   {}
func (t *Int8Type) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT8(%d)", *t.DisplayWidth)
	}
	return "INT8"
}

// Int16Type represents a ClickHouse Int16 type.
type Int16Type struct {
	SpanVal span.Span
}

func (t *Int16Type) Span() span.Span { return t.SpanVal }
func (t *Int16Type) dataTypeNode()   {}
func (t *Int16Type) String() string  { return "Int16" }

// Int32Type represents a ClickHouse Int32 type.
type Int32Type struct {
	SpanVal span.Span
}

func (t *Int32Type) Span() span.Span { return t.SpanVal }
func (t *Int32Type) dataTypeNode()   {}
func (t *Int32Type) String() string  { return "Int32" }

// Int64Type represents a BigQuery/ClickHouse Int64 type.
type Int64Type struct {
	SpanVal span.Span
}

func (t *Int64Type) Span() span.Span { return t.SpanVal }
func (t *Int64Type) dataTypeNode()   {}
func (t *Int64Type) String() string  { return "INT64" }

// Int128Type represents a ClickHouse Int128 type.
type Int128Type struct {
	SpanVal span.Span
}

func (t *Int128Type) Span() span.Span { return t.SpanVal }
func (t *Int128Type) dataTypeNode()   {}
func (t *Int128Type) String() string  { return "Int128" }

// Int256Type represents a ClickHouse Int256 type.
type Int256Type struct {
	SpanVal span.Span
}

func (t *Int256Type) Span() span.Span { return t.SpanVal }
func (t *Int256Type) dataTypeNode()   {}
func (t *Int256Type) String() string  { return "Int256" }

// IntegerType represents an integer with optional display width.
type IntegerType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *IntegerType) Span() span.Span { return t.SpanVal }
func (t *IntegerType) dataTypeNode()   {}
func (t *IntegerType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INTEGER(%d)", *t.DisplayWidth)
	}
	return "INTEGER"
}

// IntUnsignedType represents a MySQL unsigned int.
type IntUnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *IntUnsignedType) Span() span.Span { return t.SpanVal }
func (t *IntUnsignedType) dataTypeNode()   {}
func (t *IntUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "INT UNSIGNED"
}

// Int4UnsignedType represents a MySQL unsigned Int4.
type Int4UnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *Int4UnsignedType) Span() span.Span { return t.SpanVal }
func (t *Int4UnsignedType) dataTypeNode()   {}
func (t *Int4UnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT4(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "INT4 UNSIGNED"
}

// IntegerUnsignedType represents a MySQL unsigned integer.
type IntegerUnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *IntegerUnsignedType) Span() span.Span { return t.SpanVal }
func (t *IntegerUnsignedType) dataTypeNode()   {}
func (t *IntegerUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INTEGER(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "INTEGER UNSIGNED"
}

// HugeIntType represents a 128-bit integer type.
type HugeIntType struct {
	SpanVal span.Span
}

func (t *HugeIntType) Span() span.Span { return t.SpanVal }
func (t *HugeIntType) dataTypeNode()   {}
func (t *HugeIntType) String() string  { return "HUGEINT" }

// UHugeIntType represents an unsigned 128-bit integer type.
type UHugeIntType struct {
	SpanVal span.Span
}

func (t *UHugeIntType) Span() span.Span { return t.SpanVal }
func (t *UHugeIntType) dataTypeNode()   {}
func (t *UHugeIntType) String() string  { return "UHUGEINT" }

// UInt8Type represents a ClickHouse UInt8 type.
type UInt8Type struct {
	SpanVal span.Span
}

func (t *UInt8Type) Span() span.Span { return t.SpanVal }
func (t *UInt8Type) dataTypeNode()   {}
func (t *UInt8Type) String() string  { return "UInt8" }

// UInt16Type represents a ClickHouse UInt16 type.
type UInt16Type struct {
	SpanVal span.Span
}

func (t *UInt16Type) Span() span.Span { return t.SpanVal }
func (t *UInt16Type) dataTypeNode()   {}
func (t *UInt16Type) String() string  { return "UInt16" }

// UInt32Type represents a ClickHouse UInt32 type.
type UInt32Type struct {
	SpanVal span.Span
}

func (t *UInt32Type) Span() span.Span { return t.SpanVal }
func (t *UInt32Type) dataTypeNode()   {}
func (t *UInt32Type) String() string  { return "UInt32" }

// UInt64Type represents a ClickHouse UInt64 type.
type UInt64Type struct {
	SpanVal span.Span
}

func (t *UInt64Type) Span() span.Span { return t.SpanVal }
func (t *UInt64Type) dataTypeNode()   {}
func (t *UInt64Type) String() string  { return "UInt64" }

// UInt128Type represents a ClickHouse UInt128 type.
type UInt128Type struct {
	SpanVal span.Span
}

func (t *UInt128Type) Span() span.Span { return t.SpanVal }
func (t *UInt128Type) dataTypeNode()   {}
func (t *UInt128Type) String() string  { return "UInt128" }

// UInt256Type represents a ClickHouse UInt256 type.
type UInt256Type struct {
	SpanVal span.Span
}

func (t *UInt256Type) Span() span.Span { return t.SpanVal }
func (t *UInt256Type) dataTypeNode()   {}
func (t *UInt256Type) String() string  { return "UInt256" }

// BigIntType represents a big integer with optional display width.
type BigIntType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *BigIntType) Span() span.Span { return t.SpanVal }
func (t *BigIntType) dataTypeNode()   {}
func (t *BigIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("BIGINT(%d)", *t.DisplayWidth)
	}
	return "BIGINT"
}

// BigIntUnsignedType represents a MySQL unsigned big integer.
type BigIntUnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *BigIntUnsignedType) Span() span.Span { return t.SpanVal }
func (t *BigIntUnsignedType) dataTypeNode()   {}
func (t *BigIntUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("BIGINT(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "BIGINT UNSIGNED"
}

// UBigIntType represents a MySQL unsigned big integer (UBIGINT).
type UBigIntType struct {
	SpanVal span.Span
}

func (t *UBigIntType) Span() span.Span { return t.SpanVal }
func (t *UBigIntType) dataTypeNode()   {}
func (t *UBigIntType) String() string  { return "UBIGINT" }

// Int8UnsignedType represents a MySQL unsigned Int8.
type Int8UnsignedType struct {
	SpanVal      span.Span
	DisplayWidth *uint64
}

func (t *Int8UnsignedType) Span() span.Span { return t.SpanVal }
func (t *Int8UnsignedType) dataTypeNode()   {}
func (t *Int8UnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT8(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "INT8 UNSIGNED"
}

// SignedType represents a signed integer as used in MySQL CAST target types.
type SignedType struct {
	SpanVal span.Span
}

func (t *SignedType) Span() span.Span { return t.SpanVal }
func (t *SignedType) dataTypeNode()   {}
func (t *SignedType) String() string  { return "SIGNED" }

// SignedIntegerType represents a signed integer with INTEGER suffix.
type SignedIntegerType struct {
	SpanVal span.Span
}

func (t *SignedIntegerType) Span() span.Span { return t.SpanVal }
func (t *SignedIntegerType) dataTypeNode()   {}
func (t *SignedIntegerType) String() string  { return "SIGNED INTEGER" }

// UnsignedType represents an unsigned integer as used in MySQL CAST target types.
type UnsignedType struct {
	SpanVal span.Span
}

func (t *UnsignedType) Span() span.Span { return t.SpanVal }
func (t *UnsignedType) dataTypeNode()   {}
func (t *UnsignedType) String() string  { return "UNSIGNED" }

// UnsignedIntegerType represents an unsigned integer with INTEGER suffix.
type UnsignedIntegerType struct {
	SpanVal span.Span
}

func (t *UnsignedIntegerType) Span() span.Span { return t.SpanVal }
func (t *UnsignedIntegerType) dataTypeNode()   {}
func (t *UnsignedIntegerType) String() string  { return "UNSIGNED INTEGER" }

// RealType represents a floating point type.
type RealType struct {
	SpanVal span.Span
}

func (t *RealType) Span() span.Span { return t.SpanVal }
func (t *RealType) dataTypeNode()   {}
func (t *RealType) String() string  { return "REAL" }

// RealUnsignedType represents a MySQL unsigned real type.
type RealUnsignedType struct {
	SpanVal span.Span
}

func (t *RealUnsignedType) Span() span.Span { return t.SpanVal }
func (t *RealUnsignedType) dataTypeNode()   {}
func (t *RealUnsignedType) String() string  { return "REAL UNSIGNED" }

// Float4Type represents a PostgreSQL Float4 alias for Real.
type Float4Type struct {
	SpanVal span.Span
}

func (t *Float4Type) Span() span.Span { return t.SpanVal }
func (t *Float4Type) dataTypeNode()   {}
func (t *Float4Type) String() string  { return "FLOAT4" }

// Float32Type represents a ClickHouse Float32 type.
type Float32Type struct {
	SpanVal span.Span
}

func (t *Float32Type) Span() span.Span { return t.SpanVal }
func (t *Float32Type) dataTypeNode()   {}
func (t *Float32Type) String() string  { return "Float32" }

// Float64Type represents a BigQuery/ClickHouse Float64 type.
type Float64Type struct {
	SpanVal span.Span
}

func (t *Float64Type) Span() span.Span { return t.SpanVal }
func (t *Float64Type) dataTypeNode()   {}
func (t *Float64Type) String() string  { return "FLOAT64" }

// Float8Type represents a PostgreSQL Float8 alias for Double.
type Float8Type struct {
	SpanVal span.Span
}

func (t *Float8Type) Span() span.Span { return t.SpanVal }
func (t *Float8Type) dataTypeNode()   {}
func (t *Float8Type) String() string  { return "FLOAT8" }

// DoubleType represents a double type with optional precision.
type DoubleType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *DoubleType) Span() span.Span { return t.SpanVal }
func (t *DoubleType) dataTypeNode()   {}
func (t *DoubleType) String() string  { return "DOUBLE" + t.Info.String() }

// DoubleUnsignedType represents a MySQL unsigned double type.
type DoubleUnsignedType struct {
	SpanVal span.Span
	Info    ExactNumberInfo
}

func (t *DoubleUnsignedType) Span() span.Span { return t.SpanVal }
func (t *DoubleUnsignedType) dataTypeNode()   {}
func (t *DoubleUnsignedType) String() string  { return "DOUBLE" + t.Info.String() + " UNSIGNED" }

// DoublePrecisionType represents a double precision type.
type DoublePrecisionType struct {
	SpanVal span.Span
}

func (t *DoublePrecisionType) Span() span.Span { return t.SpanVal }
func (t *DoublePrecisionType) dataTypeNode()   {}
func (t *DoublePrecisionType) String() string  { return "DOUBLE PRECISION" }

// DoublePrecisionUnsignedType represents a MySQL unsigned double precision type.
type DoublePrecisionUnsignedType struct {
	SpanVal span.Span
}

func (t *DoublePrecisionUnsignedType) Span() span.Span { return t.SpanVal }
func (t *DoublePrecisionUnsignedType) dataTypeNode()   {}
func (t *DoublePrecisionUnsignedType) String() string  { return "DOUBLE PRECISION UNSIGNED" }

// BoolType represents a boolean type (PostgreSQL alias).
type BoolType struct {
	SpanVal span.Span
}

func (t *BoolType) Span() span.Span { return t.SpanVal }
func (t *BoolType) dataTypeNode()   {}
func (t *BoolType) String() string  { return "BOOL" }

// BooleanType represents a boolean type.
type BooleanType struct {
	SpanVal span.Span
}

func (t *BooleanType) Span() span.Span { return t.SpanVal }
func (t *BooleanType) dataTypeNode()   {}
func (t *BooleanType) String() string  { return "BOOLEAN" }

// DateType represents a date type.
type DateType struct {
	SpanVal span.Span
}

func (t *DateType) Span() span.Span { return t.SpanVal }
func (t *DateType) dataTypeNode()   {}
func (t *DateType) String() string  { return "DATE" }

// Date32Type represents a ClickHouse Date32 type.
type Date32Type struct {
	SpanVal span.Span
}

func (t *Date32Type) Span() span.Span { return t.SpanVal }
func (t *Date32Type) dataTypeNode()   {}
func (t *Date32Type) String() string  { return "Date32" }

// TimeType represents a time type with optional precision and time zone.
type TimeType struct {
	SpanVal      span.Span
	Precision    *uint64
	TimezoneInfo TimezoneInfo
}

func (t *TimeType) Span() span.Span { return t.SpanVal }
func (t *TimeType) dataTypeNode()   {}
func (t *TimeType) String() string {
	var sb strings.Builder
	sb.WriteString("TIME")

	if t.TimezoneInfo == Tz {
		sb.WriteString("TZ")
	}

	if t.Precision != nil {
		sb.WriteString(fmt.Sprintf("(%d)", *t.Precision))
	}

	if t.TimezoneInfo != NoTimezoneInfo && t.TimezoneInfo != Tz {
		sb.WriteString(t.TimezoneInfo.String())
	}

	return sb.String()
}

// DatetimeType represents a datetime type with optional precision.
type DatetimeType struct {
	SpanVal   span.Span
	Precision *uint64
}

func (t *DatetimeType) Span() span.Span { return t.SpanVal }
func (t *DatetimeType) dataTypeNode()   {}
func (t *DatetimeType) String() string {
	if t.Precision != nil {
		return fmt.Sprintf("DATETIME(%d)", *t.Precision)
	}
	return "DATETIME"
}

// Datetime64Type represents a ClickHouse Datetime64 type.
type Datetime64Type struct {
	SpanVal   span.Span
	Precision uint64
	Timezone  *string
}

func (t *Datetime64Type) Span() span.Span { return t.SpanVal }
func (t *Datetime64Type) dataTypeNode()   {}
func (t *Datetime64Type) String() string {
	if t.Timezone != nil {
		return fmt.Sprintf("DateTime64(%d, '%s')", t.Precision, *t.Timezone)
	}
	return fmt.Sprintf("DateTime64(%d)", t.Precision)
}

// TimestampType represents a timestamp type with optional precision and time zone.
type TimestampType struct {
	SpanVal      span.Span
	Precision    *uint64
	TimezoneInfo TimezoneInfo
}

func (t *TimestampType) Span() span.Span { return t.SpanVal }
func (t *TimestampType) dataTypeNode()   {}
func (t *TimestampType) String() string {
	var sb strings.Builder
	sb.WriteString("TIMESTAMP")

	if t.TimezoneInfo == Tz {
		sb.WriteString("TZ")
	}

	if t.Precision != nil {
		sb.WriteString(fmt.Sprintf("(%d)", *t.Precision))
	}

	if t.TimezoneInfo != NoTimezoneInfo && t.TimezoneInfo != Tz {
		sb.WriteString(t.TimezoneInfo.String())
	}

	return sb.String()
}

// TimestampNtzType represents a Databricks timestamp without time zone.
type TimestampNtzType struct {
	SpanVal   span.Span
	Precision *uint64
}

func (t *TimestampNtzType) Span() span.Span { return t.SpanVal }
func (t *TimestampNtzType) dataTypeNode()   {}
func (t *TimestampNtzType) String() string {
	if t.Precision != nil {
		return fmt.Sprintf("TIMESTAMP_NTZ(%d)", *t.Precision)
	}
	return "TIMESTAMP_NTZ"
}

// IntervalType represents an interval type.
type IntervalType struct {
	SpanVal   span.Span
	Fields    *IntervalFields
	Precision *uint64
}

func (t *IntervalType) Span() span.Span { return t.SpanVal }
func (t *IntervalType) dataTypeNode()   {}
func (t *IntervalType) String() string {
	var sb strings.Builder
	sb.WriteString("INTERVAL")
	if t.Fields != nil {
		sb.WriteString(" ")
		sb.WriteString(t.Fields.String())
	}
	if t.Precision != nil {
		sb.WriteString(fmt.Sprintf("(%d)", *t.Precision))
	}
	return sb.String()
}

// JSONType represents a JSON type.
type JSONType struct {
	SpanVal span.Span
}

func (t *JSONType) Span() span.Span { return t.SpanVal }
func (t *JSONType) dataTypeNode()   {}
func (t *JSONType) String() string  { return "JSON" }

// JSONBType represents a binary JSON type.
type JSONBType struct {
	SpanVal span.Span
}

func (t *JSONBType) Span() span.Span { return t.SpanVal }
func (t *JSONBType) dataTypeNode()   {}
func (t *JSONBType) String() string  { return "JSONB" }

// RegclassType represents a PostgreSQL regclass type.
type RegclassType struct {
	SpanVal span.Span
}

func (t *RegclassType) Span() span.Span { return t.SpanVal }
func (t *RegclassType) dataTypeNode()   {}
func (t *RegclassType) String() string  { return "REGCLASS" }

// TextType represents a text type.
type TextType struct {
	SpanVal span.Span
}

func (t *TextType) Span() span.Span { return t.SpanVal }
func (t *TextType) dataTypeNode()   {}
func (t *TextType) String() string  { return "TEXT" }

// TinyTextType represents a MySQL tiny text type.
type TinyTextType struct {
	SpanVal span.Span
}

func (t *TinyTextType) Span() span.Span { return t.SpanVal }
func (t *TinyTextType) dataTypeNode()   {}
func (t *TinyTextType) String() string  { return "TINYTEXT" }

// MediumTextType represents a MySQL medium text type.
type MediumTextType struct {
	SpanVal span.Span
}

func (t *MediumTextType) Span() span.Span { return t.SpanVal }
func (t *MediumTextType) dataTypeNode()   {}
func (t *MediumTextType) String() string  { return "MEDIUMTEXT" }

// LongTextType represents a MySQL long text type.
type LongTextType struct {
	SpanVal span.Span
}

func (t *LongTextType) Span() span.Span { return t.SpanVal }
func (t *LongTextType) dataTypeNode()   {}
func (t *LongTextType) String() string  { return "LONGTEXT" }

// StringType represents a string type with optional length.
type StringType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *StringType) Span() span.Span { return t.SpanVal }
func (t *StringType) dataTypeNode()   {}
func (t *StringType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("STRING(%d)", *t.Length)
	}
	return "STRING"
}

// FixedStringType represents a ClickHouse FixedString type.
type FixedStringType struct {
	SpanVal span.Span
	Length  uint64
}

func (t *FixedStringType) Span() span.Span { return t.SpanVal }
func (t *FixedStringType) dataTypeNode()   {}
func (t *FixedStringType) String() string {
	return fmt.Sprintf("FixedString(%d)", t.Length)
}

// ByteaType represents a PostgreSQL bytea type.
type ByteaType struct {
	SpanVal span.Span
}

func (t *ByteaType) Span() span.Span { return t.SpanVal }
func (t *ByteaType) dataTypeNode()   {}
func (t *ByteaType) String() string  { return "BYTEA" }

// BitType represents a bit string type.
type BitType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *BitType) Span() span.Span { return t.SpanVal }
func (t *BitType) dataTypeNode()   {}
func (t *BitType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("BIT(%d)", *t.Length)
	}
	return "BIT"
}

// BitVaryingType represents a bit varying type.
type BitVaryingType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *BitVaryingType) Span() span.Span { return t.SpanVal }
func (t *BitVaryingType) dataTypeNode()   {}
func (t *BitVaryingType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("BIT VARYING(%d)", *t.Length)
	}
	return "BIT VARYING"
}

// VarBitType represents a varbit type (PostgreSQL alias for BIT VARYING).
type VarBitType struct {
	SpanVal span.Span
	Length  *uint64
}

func (t *VarBitType) Span() span.Span { return t.SpanVal }
func (t *VarBitType) dataTypeNode()   {}
func (t *VarBitType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("VARBIT(%d)", *t.Length)
	}
	return "VARBIT"
}

// CustomType represents a custom/user-defined type.
type CustomType struct {
	SpanVal   span.Span
	Name      *expr.ObjectName
	Modifiers []string
}

func (t *CustomType) Span() span.Span { return t.SpanVal }
func (t *CustomType) dataTypeNode()   {}
func (t *CustomType) String() string {
	if len(t.Modifiers) > 0 {
		return fmt.Sprintf("%s(%s)", t.Name.String(), strings.Join(t.Modifiers, ", "))
	}
	return t.Name.String()
}

// ArrayType represents an array type.
type ArrayType struct {
	SpanVal span.Span
	ElemDef ArrayElemTypeDef
}

func (t *ArrayType) Span() span.Span { return t.SpanVal }
func (t *ArrayType) dataTypeNode()   {}
func (t *ArrayType) String() string  { return t.ElemDef.String() }

// MapType represents a ClickHouse Map type.
type MapType struct {
	SpanVal   span.Span
	KeyType   DataType
	ValueType DataType
}

func (t *MapType) Span() span.Span { return t.SpanVal }
func (t *MapType) dataTypeNode()   {}
func (t *MapType) String() string {
	return fmt.Sprintf("Map(%s, %s)", t.KeyType.String(), t.ValueType.String())
}

// TupleType represents a ClickHouse Tuple type.
type TupleType struct {
	SpanVal span.Span
	Fields  []*StructField
}

func (t *TupleType) Span() span.Span { return t.SpanVal }
func (t *TupleType) dataTypeNode()   {}
func (t *TupleType) String() string {
	fields := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		fields[i] = f.String()
	}
	return fmt.Sprintf("Tuple(%s)", strings.Join(fields, ", "))
}

// NestedType represents a ClickHouse Nested type.
type NestedType struct {
	SpanVal span.Span
	Columns []*ColumnDef
}

func (t *NestedType) Span() span.Span { return t.SpanVal }
func (t *NestedType) dataTypeNode()   {}
func (t *NestedType) String() string {
	cols := make([]string, len(t.Columns))
	for i, col := range t.Columns {
		cols[i] = col.String()
	}
	return fmt.Sprintf("Nested(%s)", strings.Join(cols, ", "))
}

// EnumType represents an enum type.
type EnumType struct {
	SpanVal span.Span
	Members []EnumMember
	Bits    *uint8
}

func (t *EnumType) Span() span.Span { return t.SpanVal }
func (t *EnumType) dataTypeNode()   {}
func (t *EnumType) String() string {
	var sb strings.Builder
	if t.Bits != nil {
		sb.WriteString(fmt.Sprintf("ENUM%d", *t.Bits))
	} else {
		sb.WriteString("ENUM")
	}

	members := make([]string, len(t.Members))
	for i, m := range t.Members {
		members[i] = m.String()
	}
	sb.WriteString(fmt.Sprintf("(%s)", strings.Join(members, ", ")))
	return sb.String()
}

// SetType represents a MySQL Set type.
type SetType struct {
	SpanVal span.Span
	Values  []string
}

func (t *SetType) Span() span.Span { return t.SpanVal }
func (t *SetType) dataTypeNode()   {}
func (t *SetType) String() string {
	values := make([]string, len(t.Values))
	for i, v := range t.Values {
		values[i] = fmt.Sprintf("'%s'", escapeSingleQuote(v))
	}
	return fmt.Sprintf("SET(%s)", strings.Join(values, ", "))
}

// StructType represents a struct type.
type StructType struct {
	SpanVal     span.Span
	Fields      []*StructField
	BracketKind StructBracketKind
}

func (t *StructType) Span() span.Span { return t.SpanVal }
func (t *StructType) dataTypeNode()   {}
func (t *StructType) String() string {
	if len(t.Fields) == 0 {
		return "STRUCT"
	}

	fields := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		fields[i] = f.String()
	}

	if t.BracketKind == AngleBrackets {
		return fmt.Sprintf("STRUCT<%s>", strings.Join(fields, ", "))
	}
	return fmt.Sprintf("STRUCT(%s)", strings.Join(fields, ", "))
}

// UnionType represents a DuckDB Union type.
type UnionType struct {
	SpanVal span.Span
	Fields  []*UnionField
}

func (t *UnionType) Span() span.Span { return t.SpanVal }
func (t *UnionType) dataTypeNode()   {}
func (t *UnionType) String() string {
	fields := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		fields[i] = f.String()
	}
	return fmt.Sprintf("UNION(%s)", strings.Join(fields, ", "))
}

// NullableType represents a ClickHouse Nullable type.
type NullableType struct {
	SpanVal   span.Span
	InnerType DataType
}

func (t *NullableType) Span() span.Span { return t.SpanVal }
func (t *NullableType) dataTypeNode()   {}
func (t *NullableType) String() string {
	return fmt.Sprintf("Nullable(%s)", t.InnerType.String())
}

// FixedStringDefType represents a ClickHouse FixedString type.
type FixedStringDefType struct {
	SpanVal span.Span
	Length  uint64
}

func (t *FixedStringDefType) Span() span.Span { return t.SpanVal }
func (t *FixedStringDefType) dataTypeNode()   {}
func (t *FixedStringDefType) String() string {
	return fmt.Sprintf("FixedString(%d)", t.Length)
}

// LowCardinalityType represents a ClickHouse LowCardinality type.
type LowCardinalityType struct {
	SpanVal   span.Span
	InnerType DataType
}

func (t *LowCardinalityType) Span() span.Span { return t.SpanVal }
func (t *LowCardinalityType) dataTypeNode()   {}
func (t *LowCardinalityType) String() string {
	return fmt.Sprintf("LowCardinality(%s)", t.InnerType.String())
}

// UnspecifiedType represents an unspecified type (SQLite).
type UnspecifiedType struct {
	SpanVal span.Span
}

func (t *UnspecifiedType) Span() span.Span { return t.SpanVal }
func (t *UnspecifiedType) dataTypeNode()   {}
func (t *UnspecifiedType) String() string  { return "" }

// TriggerType represents a PostgreSQL trigger data type.
type TriggerType struct {
	SpanVal span.Span
}

func (t *TriggerType) Span() span.Span { return t.SpanVal }
func (t *TriggerType) dataTypeNode()   {}
func (t *TriggerType) String() string  { return "TRIGGER" }

// AnyType represents a BigQuery ANY TYPE for UDF definitions.
type AnyType struct {
	SpanVal span.Span
}

func (t *AnyType) Span() span.Span { return t.SpanVal }
func (t *AnyType) dataTypeNode()   {}
func (t *AnyType) String() string  { return "ANY TYPE" }

// GeometricType represents a PostgreSQL geometric type.
type GeometricType struct {
	SpanVal span.Span
	Kind    GeometricTypeKind
}

func (t *GeometricType) Span() span.Span { return t.SpanVal }
func (t *GeometricType) dataTypeNode()   {}
func (t *GeometricType) String() string  { return t.Kind.String() }

// TsVectorType represents a PostgreSQL text search vector type.
type TsVectorType struct {
	SpanVal span.Span
}

func (t *TsVectorType) Span() span.Span { return t.SpanVal }
func (t *TsVectorType) dataTypeNode()   {}
func (t *TsVectorType) String() string  { return "TSVECTOR" }

// TsQueryType represents a PostgreSQL text search query type.
type TsQueryType struct {
	SpanVal span.Span
}

func (t *TsQueryType) Span() span.Span { return t.SpanVal }
func (t *TsQueryType) dataTypeNode()   {}
func (t *TsQueryType) String() string  { return "TSQUERY" }
