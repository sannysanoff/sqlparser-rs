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

// Package ast provides consolidated SQL data type AST types.
// This file contains data type definitions migrated from the datatype/ subpackage.
//
// Naming convention: D-prefix types (e.g., DVarcharType, DIntType)
// are the consolidated versions that use the main ast package interfaces.
package ast

import (
	"fmt"
	"strings"

	"github.com/user/sqlparser/token"
)

// ============================================================================
// Supporting Types
// ============================================================================

// DCharacterLength represents information about character length
type DCharacterLength struct {
	Length uint64
	Unit   *DCharLengthUnits
}

// Max returns true if this represents VARCHAR(MAX) or NVARCHAR(MAX)
func (c DCharacterLength) Max() bool {
	return c.Length == 0 && c.Unit == nil
}

// String returns the SQL representation
func (c DCharacterLength) String() string {
	if c.Max() {
		return "MAX"
	}
	if c.Unit != nil {
		return fmt.Sprintf("%d %s", c.Length, c.Unit.String())
	}
	return fmt.Sprintf("%d", c.Length)
}

// DCharLengthUnits represents possible units for characters
type DCharLengthUnits int

const (
	DCharactersUnit DCharLengthUnits = iota
	DOctetsUnit
)

func (c DCharLengthUnits) String() string {
	switch c {
	case DCharactersUnit:
		return "CHARACTERS"
	case DOctetsUnit:
		return "OCTETS"
	}
	return ""
}

// DBinaryLength represents information about binary length
type DBinaryLength struct {
	Length uint64
	Max    bool
}

// String returns the SQL representation
func (b DBinaryLength) String() string {
	if b.Max {
		return "MAX"
	}
	return fmt.Sprintf("%d", b.Length)
}

// DExactNumberInfo represents additional information for exact numeric data types
type DExactNumberInfo struct {
	Precision *uint64
	Scale     *int64
}

// String returns the SQL representation
func (e DExactNumberInfo) String() string {
	if e.Precision == nil {
		return ""
	}
	if e.Scale == nil {
		return fmt.Sprintf("(%d)", *e.Precision)
	}
	return fmt.Sprintf("(%d,%d)", *e.Precision, *e.Scale)
}

// DTimezoneInfo represents timezone information for temporal types
type DTimezoneInfo int

const (
	DNoTimezoneInfo DTimezoneInfo = iota
	DWithTimeZone
	DWithoutTimeZone
	DTz
)

func (t DTimezoneInfo) String() string {
	switch t {
	case DNoTimezoneInfo:
		return ""
	case DWithTimeZone:
		return " WITH TIME ZONE"
	case DWithoutTimeZone:
		return " WITHOUT TIME ZONE"
	case DTz:
		return "TZ"
	}
	return ""
}

// DIntervalFields represents fields for PostgreSQL INTERVAL type
type DIntervalFields int

const (
	DIntervalYear DIntervalFields = iota
	DIntervalMonth
	DIntervalDay
	DIntervalHour
	DIntervalMinute
	DIntervalSecond
	DIntervalYearToMonth
	DIntervalDayToHour
	DIntervalDayToMinute
	DIntervalDayToSecond
	DIntervalHourToMinute
	DIntervalHourToSecond
	DIntervalMinuteToSecond
)

func (i DIntervalFields) String() string {
	switch i {
	case DIntervalYear:
		return "YEAR"
	case DIntervalMonth:
		return "MONTH"
	case DIntervalDay:
		return "DAY"
	case DIntervalHour:
		return "HOUR"
	case DIntervalMinute:
		return "MINUTE"
	case DIntervalSecond:
		return "SECOND"
	case DIntervalYearToMonth:
		return "YEAR TO MONTH"
	case DIntervalDayToHour:
		return "DAY TO HOUR"
	case DIntervalDayToMinute:
		return "DAY TO MINUTE"
	case DIntervalDayToSecond:
		return "DAY TO SECOND"
	case DIntervalHourToMinute:
		return "HOUR TO MINUTE"
	case DIntervalHourToSecond:
		return "HOUR TO SECOND"
	case DIntervalMinuteToSecond:
		return "MINUTE TO SECOND"
	}
	return ""
}

// DArrayElemTypeDef represents the data type of elements in an array
type DArrayElemTypeDef struct {
	Style    DArrayStyle
	DataType DataType
	Size     *uint64
}

// DArrayStyle represents the syntax style for array type definitions
type DArrayStyle int

const (
	DArrayNone DArrayStyle = iota
	DArrayAngleBracket
	DArraySquareBracket
	DArrayParenthesis
)

// String returns the SQL representation
func (a DArrayElemTypeDef) String() string {
	switch a.Style {
	case DArrayNone:
		return "ARRAY"
	case DArrayAngleBracket:
		if a.DataType != nil {
			return fmt.Sprintf("ARRAY<%s>", a.DataType.String())
		}
		return "ARRAY<>"
	case DArraySquareBracket:
		if a.DataType != nil {
			if a.Size != nil {
				return fmt.Sprintf("%s[%d]", a.DataType.String(), *a.Size)
			}
			return fmt.Sprintf("%s[]", a.DataType.String())
		}
		return "[]"
	case DArrayParenthesis:
		if a.DataType != nil {
			return fmt.Sprintf("Array(%s)", a.DataType.String())
		}
		return "Array()"
	}
	return "ARRAY"
}

// DStructBracketKind represents the type of brackets used for STRUCT literals
type DStructBracketKind int

const (
	DParentheses DStructBracketKind = iota
	DAngleBrackets
)

// ============================================================================
// Character Types
// ============================================================================

// DVarcharType represents a variable-length character type
type DVarcharType struct {
	SpanVal token.Span
	Length  *DCharacterLength
}

func (t *DVarcharType) Span() token.Span { return t.SpanVal }
func (t *DVarcharType) dataType()        {}
func (t *DVarcharType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("VARCHAR(%s)", t.Length.String())
	}
	return "VARCHAR"
}

// DCharType represents a fixed-length character type
type DCharType struct {
	SpanVal token.Span
	Length  *DCharacterLength
}

func (t *DCharType) Span() token.Span { return t.SpanVal }
func (t *DCharType) dataType()        {}
func (t *DCharType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHAR(%s)", t.Length.String())
	}
	return "CHAR"
}

// DCharacterType represents a CHARACTER type
type DCharacterType struct {
	SpanVal token.Span
	Length  *DCharacterLength
}

func (t *DCharacterType) Span() token.Span { return t.SpanVal }
func (t *DCharacterType) dataType()        {}
func (t *DCharacterType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHARACTER(%s)", t.Length.String())
	}
	return "CHARACTER"
}

// DCharacterVaryingType represents a CHARACTER VARYING type
type DCharacterVaryingType struct {
	SpanVal token.Span
	Length  *DCharacterLength
}

func (t *DCharacterVaryingType) Span() token.Span { return t.SpanVal }
func (t *DCharacterVaryingType) dataType()        {}
func (t *DCharacterVaryingType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CHARACTER VARYING(%s)", t.Length.String())
	}
	return "CHARACTER VARYING"
}

// DNvarcharType represents a variable-length character type
type DNvarcharType struct {
	SpanVal token.Span
	Length  *DCharacterLength
}

func (t *DNvarcharType) Span() token.Span { return t.SpanVal }
func (t *DNvarcharType) dataType()        {}
func (t *DNvarcharType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("NVARCHAR(%s)", t.Length.String())
	}
	return "NVARCHAR"
}

// DTextType represents a TEXT type
type DTextType struct {
	SpanVal token.Span
}

func (t *DTextType) Span() token.Span { return t.SpanVal }
func (t *DTextType) dataType()        {}
func (t *DTextType) String() string   { return "TEXT" }

// DStringType represents a STRING type (BigQuery)
type DStringType struct {
	SpanVal token.Span
	Length  *uint64
}

func (t *DStringType) Span() token.Span { return t.SpanVal }
func (t *DStringType) dataType()        {}
func (t *DStringType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("STRING(%d)", *t.Length)
	}
	return "STRING"
}

// DClobType represents a CLOB type
type DClobType struct {
	SpanVal token.Span
	Length  *uint64
}

func (t *DClobType) Span() token.Span { return t.SpanVal }
func (t *DClobType) dataType()        {}
func (t *DClobType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("CLOB(%d)", *t.Length)
	}
	return "CLOB"
}

// ============================================================================
// Binary Types
// ============================================================================

// DBinaryType represents a fixed-length binary type
type DBinaryType struct {
	SpanVal token.Span
	Length  *uint64
}

func (t *DBinaryType) Span() token.Span { return t.SpanVal }
func (t *DBinaryType) dataType()        {}
func (t *DBinaryType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("BINARY(%d)", *t.Length)
	}
	return "BINARY"
}

// DVarbinaryType represents a variable-length binary type
type DVarbinaryType struct {
	SpanVal token.Span
	Length  *DBinaryLength
}

func (t *DVarbinaryType) Span() token.Span { return t.SpanVal }
func (t *DVarbinaryType) dataType()        {}
func (t *DVarbinaryType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("VARBINARY(%s)", t.Length.String())
	}
	return "VARBINARY"
}

// DBlobType represents a BLOB type
type DBlobType struct {
	SpanVal token.Span
	Length  *uint64
}

func (t *DBlobType) Span() token.Span { return t.SpanVal }
func (t *DBlobType) dataType()        {}
func (t *DBlobType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("BLOB(%d)", *t.Length)
	}
	return "BLOB"
}

// DBytesType represents a BYTES type (BigQuery)
type DBytesType struct {
	SpanVal token.Span
	Length  *uint64
}

func (t *DBytesType) Span() token.Span { return t.SpanVal }
func (t *DBytesType) dataType()        {}
func (t *DBytesType) String() string {
	if t.Length != nil {
		return fmt.Sprintf("BYTES(%d)", *t.Length)
	}
	return "BYTES"
}

// ============================================================================
// Integer Types
// ============================================================================

// DIntType represents an INTEGER type
type DIntType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DIntType) Span() token.Span { return t.SpanVal }
func (t *DIntType) dataType()        {}
func (t *DIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT(%d)", *t.DisplayWidth)
	}
	return "INT"
}

// DIntegerType represents an INTEGER type
type DIntegerType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DIntegerType) Span() token.Span { return t.SpanVal }
func (t *DIntegerType) dataType()        {}
func (t *DIntegerType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INTEGER(%d)", *t.DisplayWidth)
	}
	return "INTEGER"
}

// DBigIntType represents a BIGINT type
type DBigIntType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DBigIntType) Span() token.Span { return t.SpanVal }
func (t *DBigIntType) dataType()        {}
func (t *DBigIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("BIGINT(%d)", *t.DisplayWidth)
	}
	return "BIGINT"
}

// DSmallIntType represents a SMALLINT type
type DSmallIntType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DSmallIntType) Span() token.Span { return t.SpanVal }
func (t *DSmallIntType) dataType()        {}
func (t *DSmallIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("SMALLINT(%d)", *t.DisplayWidth)
	}
	return "SMALLINT"
}

// DTinyIntType represents a TINYINT type
type DTinyIntType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DTinyIntType) Span() token.Span { return t.SpanVal }
func (t *DTinyIntType) dataType()        {}
func (t *DTinyIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("TINYINT(%d)", *t.DisplayWidth)
	}
	return "TINYINT"
}

// DMediumIntType represents a MEDIUMINT type
type DMediumIntType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DMediumIntType) Span() token.Span { return t.SpanVal }
func (t *DMediumIntType) dataType()        {}
func (t *DMediumIntType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("MEDIUMINT(%d)", *t.DisplayWidth)
	}
	return "MEDIUMINT"
}

// DInt64Type represents an INT64 type (BigQuery/ClickHouse)
type DInt64Type struct {
	SpanVal token.Span
}

func (t *DInt64Type) Span() token.Span { return t.SpanVal }
func (t *DInt64Type) dataType()        {}
func (t *DInt64Type) String() string   { return "INT64" }

// DInt32Type represents an Int32 type (ClickHouse)
type DInt32Type struct {
	SpanVal token.Span
}

func (t *DInt32Type) Span() token.Span { return t.SpanVal }
func (t *DInt32Type) dataType()        {}
func (t *DInt32Type) String() string   { return "Int32" }

// DInt16Type represents an Int16 type (ClickHouse)
type DInt16Type struct {
	SpanVal token.Span
}

func (t *DInt16Type) Span() token.Span { return t.SpanVal }
func (t *DInt16Type) dataType()        {}
func (t *DInt16Type) String() string   { return "Int16" }

// DInt8Type represents an Int8 type (ClickHouse)
type DInt8Type struct {
	SpanVal token.Span
}

func (t *DInt8Type) Span() token.Span { return t.SpanVal }
func (t *DInt8Type) dataType()        {}
func (t *DInt8Type) String() string   { return "Int8" }

// DTinyIntUnsignedType represents an unsigned TINYINT type
type DTinyIntUnsignedType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DTinyIntUnsignedType) Span() token.Span { return t.SpanVal }
func (t *DTinyIntUnsignedType) dataType()        {}
func (t *DTinyIntUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("TINYINT(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "TINYINT UNSIGNED"
}

// DIntUnsignedType represents an unsigned INT type
type DIntUnsignedType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DIntUnsignedType) Span() token.Span { return t.SpanVal }
func (t *DIntUnsignedType) dataType()        {}
func (t *DIntUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("INT(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "INT UNSIGNED"
}

// DBigIntUnsignedType represents an unsigned BIGINT type
type DBigIntUnsignedType struct {
	SpanVal      token.Span
	DisplayWidth *uint64
}

func (t *DBigIntUnsignedType) Span() token.Span { return t.SpanVal }
func (t *DBigIntUnsignedType) dataType()        {}
func (t *DBigIntUnsignedType) String() string {
	if t.DisplayWidth != nil {
		return fmt.Sprintf("BIGINT(%d) UNSIGNED", *t.DisplayWidth)
	}
	return "BIGINT UNSIGNED"
}

// ============================================================================
// Decimal/Numeric Types
// ============================================================================

// DDecimalType represents a DECIMAL type
type DDecimalType struct {
	SpanVal token.Span
	Info    DExactNumberInfo
}

func (t *DDecimalType) Span() token.Span { return t.SpanVal }
func (t *DDecimalType) dataType()        {}
func (t *DDecimalType) String() string   { return "DECIMAL" + t.Info.String() }

// DNumericType represents a NUMERIC type
type DNumericType struct {
	SpanVal token.Span
	Info    DExactNumberInfo
}

func (t *DNumericType) Span() token.Span { return t.SpanVal }
func (t *DNumericType) dataType()        {}
func (t *DNumericType) String() string   { return "NUMERIC" + t.Info.String() }

// DDecType represents a DEC type
type DDecType struct {
	SpanVal token.Span
	Info    DExactNumberInfo
}

func (t *DDecType) Span() token.Span { return t.SpanVal }
func (t *DDecType) dataType()        {}
func (t *DDecType) String() string   { return "DEC" + t.Info.String() }

// DBigDecimalType represents a BIGDECIMAL type (BigQuery)
type DBigDecimalType struct {
	SpanVal token.Span
	Info    DExactNumberInfo
}

func (t *DBigDecimalType) Span() token.Span { return t.SpanVal }
func (t *DBigDecimalType) dataType()        {}
func (t *DBigDecimalType) String() string   { return "BIGDECIMAL" + t.Info.String() }

// DBigNumericType represents a BIGNUMERIC type (BigQuery)
type DBigNumericType struct {
	SpanVal token.Span
	Info    DExactNumberInfo
}

func (t *DBigNumericType) Span() token.Span { return t.SpanVal }
func (t *DBigNumericType) dataType()        {}
func (t *DBigNumericType) String() string   { return "BIGNUMERIC" + t.Info.String() }

// ============================================================================
// Floating Point Types
// ============================================================================

// DFloatType represents a FLOAT type
type DFloatType struct {
	SpanVal token.Span
	Info    DExactNumberInfo
}

func (t *DFloatType) Span() token.Span { return t.SpanVal }
func (t *DFloatType) dataType()        {}
func (t *DFloatType) String() string   { return "FLOAT" + t.Info.String() }

// DDoubleType represents a DOUBLE type
type DDoubleType struct {
	SpanVal token.Span
	Info    DExactNumberInfo
}

func (t *DDoubleType) Span() token.Span { return t.SpanVal }
func (t *DDoubleType) dataType()        {}
func (t *DDoubleType) String() string   { return "DOUBLE" + t.Info.String() }

// DDoublePrecisionType represents a DOUBLE PRECISION type
type DDoublePrecisionType struct {
	SpanVal token.Span
}

func (t *DDoublePrecisionType) Span() token.Span { return t.SpanVal }
func (t *DDoublePrecisionType) dataType()        {}
func (t *DDoublePrecisionType) String() string   { return "DOUBLE PRECISION" }

// DRealType represents a REAL type
type DRealType struct {
	SpanVal token.Span
}

func (t *DRealType) Span() token.Span { return t.SpanVal }
func (t *DRealType) dataType()        {}
func (t *DRealType) String() string   { return "REAL" }

// DFloat64Type represents a FLOAT64 type (BigQuery/ClickHouse)
type DFloat64Type struct {
	SpanVal token.Span
}

func (t *DFloat64Type) Span() token.Span { return t.SpanVal }
func (t *DFloat64Type) dataType()        {}
func (t *DFloat64Type) String() string   { return "FLOAT64" }

// DFloat32Type represents a Float32 type (ClickHouse)
type DFloat32Type struct {
	SpanVal token.Span
}

func (t *DFloat32Type) Span() token.Span { return t.SpanVal }
func (t *DFloat32Type) dataType()        {}
func (t *DFloat32Type) String() string   { return "Float32" }

// ============================================================================
// Boolean Types
// ============================================================================

// DBooleanType represents a BOOLEAN type
type DBooleanType struct {
	SpanVal token.Span
}

func (t *DBooleanType) Span() token.Span { return t.SpanVal }
func (t *DBooleanType) dataType()        {}
func (t *DBooleanType) String() string   { return "BOOLEAN" }

// DBoolType represents a BOOL type (PostgreSQL alias)
type DBoolType struct {
	SpanVal token.Span
}

func (t *DBoolType) Span() token.Span { return t.SpanVal }
func (t *DBoolType) dataType()        {}
func (t *DBoolType) String() string   { return "BOOL" }

// ============================================================================
// Date/Time Types
// ============================================================================

// DDateType represents a DATE type
type DDateType struct {
	SpanVal token.Span
}

func (t *DDateType) Span() token.Span { return t.SpanVal }
func (t *DDateType) dataType()        {}
func (t *DDateType) String() string   { return "DATE" }

// DDate32Type represents a Date32 type (ClickHouse)
type DDate32Type struct {
	SpanVal token.Span
}

func (t *DDate32Type) Span() token.Span { return t.SpanVal }
func (t *DDate32Type) dataType()        {}
func (t *DDate32Type) String() string   { return "Date32" }

// DTimeType represents a TIME type
type DTimeType struct {
	SpanVal      token.Span
	Precision    *uint64
	TimezoneInfo DTimezoneInfo
}

func (t *DTimeType) Span() token.Span { return t.SpanVal }
func (t *DTimeType) dataType()        {}
func (t *DTimeType) String() string {
	var sb strings.Builder
	sb.WriteString("TIME")
	if t.TimezoneInfo == DTz {
		sb.WriteString("TZ")
	}
	if t.Precision != nil {
		sb.WriteString(fmt.Sprintf("(%d)", *t.Precision))
	}
	if t.TimezoneInfo != DNoTimezoneInfo && t.TimezoneInfo != DTz {
		sb.WriteString(t.TimezoneInfo.String())
	}
	return sb.String()
}

// DDatetimeType represents a DATETIME type
type DDatetimeType struct {
	SpanVal   token.Span
	Precision *uint64
}

func (t *DDatetimeType) Span() token.Span { return t.SpanVal }
func (t *DDatetimeType) dataType()        {}
func (t *DDatetimeType) String() string {
	if t.Precision != nil {
		return fmt.Sprintf("DATETIME(%d)", *t.Precision)
	}
	return "DATETIME"
}

// DDatetime64Type represents a Datetime64 type (ClickHouse)
type DDatetime64Type struct {
	SpanVal   token.Span
	Precision uint64
	Timezone  *string
}

func (t *DDatetime64Type) Span() token.Span { return t.SpanVal }
func (t *DDatetime64Type) dataType()        {}
func (t *DDatetime64Type) String() string {
	if t.Timezone != nil {
		return fmt.Sprintf("DateTime64(%d, '%s')", t.Precision, *t.Timezone)
	}
	return fmt.Sprintf("DateTime64(%d)", t.Precision)
}

// DTimestampType represents a TIMESTAMP type
type DTimestampType struct {
	SpanVal      token.Span
	Precision    *uint64
	TimezoneInfo DTimezoneInfo
}

func (t *DTimestampType) Span() token.Span { return t.SpanVal }
func (t *DTimestampType) dataType()        {}
func (t *DTimestampType) String() string {
	var sb strings.Builder
	sb.WriteString("TIMESTAMP")
	if t.TimezoneInfo == DTz {
		sb.WriteString("TZ")
	}
	if t.Precision != nil {
		sb.WriteString(fmt.Sprintf("(%d)", *t.Precision))
	}
	if t.TimezoneInfo != DNoTimezoneInfo && t.TimezoneInfo != DTz {
		sb.WriteString(t.TimezoneInfo.String())
	}
	return sb.String()
}

// DTimestampNtzType represents a TIMESTAMP_NTZ type (Databricks)
type DTimestampNtzType struct {
	SpanVal   token.Span
	Precision *uint64
}

func (t *DTimestampNtzType) Span() token.Span { return t.SpanVal }
func (t *DTimestampNtzType) dataType()        {}
func (t *DTimestampNtzType) String() string {
	if t.Precision != nil {
		return fmt.Sprintf("TIMESTAMP_NTZ(%d)", *t.Precision)
	}
	return "TIMESTAMP_NTZ"
}

// DIntervalType represents an INTERVAL type
type DIntervalType struct {
	SpanVal   token.Span
	Fields    *DIntervalFields
	Precision *uint64
}

func (t *DIntervalType) Span() token.Span { return t.SpanVal }
func (t *DIntervalType) dataType()        {}
func (t *DIntervalType) String() string {
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

// ============================================================================
// Complex Types
// ============================================================================

// DArrayType represents an ARRAY type
type DArrayType struct {
	SpanVal token.Span
	ElemDef DArrayElemTypeDef
}

func (t *DArrayType) Span() token.Span { return t.SpanVal }
func (t *DArrayType) dataType()        {}
func (t *DArrayType) String() string   { return t.ElemDef.String() }

// DStructType represents a STRUCT type
type DStructType struct {
	SpanVal     token.Span
	Fields      []*DStructField
	BracketKind DStructBracketKind
}

func (t *DStructType) String() string {
	if len(t.Fields) == 0 {
		return "STRUCT"
	}
	fields := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		fields[i] = f.String()
	}
	if t.BracketKind == DAngleBrackets {
		return fmt.Sprintf("STRUCT<%s>", strings.Join(fields, ", "))
	}
	return fmt.Sprintf("STRUCT(%s)", strings.Join(fields, ", "))
}

func (t *DStructType) Span() token.Span { return t.SpanVal }
func (t *DStructType) dataType()        {}

// DStructField represents a field definition within a struct
type DStructField struct {
	FieldName *Ident
	FieldType DataType
	Options   []*DSqlOption
}

func (s *DStructField) String() string {
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

// DTupleType represents a TUPLE type (ClickHouse)
type DTupleType struct {
	SpanVal token.Span
	Fields  []*DStructField
}

func (t *DTupleType) Span() token.Span { return t.SpanVal }
func (t *DTupleType) dataType()        {}
func (t *DTupleType) String() string {
	fields := make([]string, len(t.Fields))
	for i, f := range t.Fields {
		fields[i] = f.String()
	}
	return fmt.Sprintf("Tuple(%s)", strings.Join(fields, ", "))
}

// DMapType represents a MAP type
type DMapType struct {
	SpanVal   token.Span
	KeyType   DataType
	ValueType DataType
}

func (t *DMapType) Span() token.Span { return t.SpanVal }
func (t *DMapType) dataType()        {}
func (t *DMapType) String() string {
	return fmt.Sprintf("Map(%s, %s)", t.KeyType.String(), t.ValueType.String())
}

// ============================================================================
// JSON and Special Types
// ============================================================================

// DJSONType represents a JSON type
type DJSONType struct {
	SpanVal token.Span
}

func (t *DJSONType) Span() token.Span { return t.SpanVal }
func (t *DJSONType) dataType()        {}
func (t *DJSONType) String() string   { return "JSON" }

// DJSONBType represents a JSONB type (binary JSON)
type DJSONBType struct {
	SpanVal token.Span
}

func (t *DJSONBType) Span() token.Span { return t.SpanVal }
func (t *DJSONBType) dataType()        {}
func (t *DJSONBType) String() string   { return "JSONB" }

// DUUIDType represents a UUID type
type DUUIDType struct {
	SpanVal token.Span
}

func (t *DUUIDType) Span() token.Span { return t.SpanVal }
func (t *DUUIDType) dataType()        {}
func (t *DUUIDType) String() string   { return "UUID" }

// DEnumType represents an ENUM type
type DEnumType struct {
	SpanVal token.Span
	Members []DEnumMember
	Bits    *uint8
}

func (t *DEnumType) Span() token.Span { return t.SpanVal }
func (t *DEnumType) dataType()        {}
func (t *DEnumType) String() string {
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

// DEnumMember represents a member of an ENUM type
type DEnumMember struct {
	Name  string
	Value Expr
}

func (e DEnumMember) String() string {
	if e.Value != nil {
		return fmt.Sprintf("'%s' = %s", escapeSingleQuote(e.Name), e.Value.String())
	}
	return fmt.Sprintf("'%s'", escapeSingleQuote(e.Name))
}

// DSetType represents a SET type (MySQL)
type DSetType struct {
	SpanVal token.Span
	Values  []string
}

func (t *DSetType) Span() token.Span { return t.SpanVal }
func (t *DSetType) dataType()        {}
func (t *DSetType) String() string {
	values := make([]string, len(t.Values))
	for i, v := range t.Values {
		values[i] = fmt.Sprintf("'%s'", escapeSingleQuote(v))
	}
	return fmt.Sprintf("SET(%s)", strings.Join(values, ", "))
}

// DNullType represents a NULL type
type DNullType struct {
	SpanVal token.Span
}

func (t *DNullType) Span() token.Span { return t.SpanVal }
func (t *DNullType) dataType()        {}
func (t *DNullType) String() string   { return "NULL" }

// ============================================================================
// Placeholder Types (to be expanded)
// ============================================================================

type DSqlOption struct{}

func (d *DSqlOption) String() string { return "" }
