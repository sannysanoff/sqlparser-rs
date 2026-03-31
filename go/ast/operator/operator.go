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

// Package operator provides SQL operator types for the sqlparser AST.
package operator

import (
	"fmt"
	"strings"
)

// UnaryOperator represents unary operators.
type UnaryOperator int

const (
	// Geometric operators (PostgreSQL/Redshift)
	UOpAtDashAt           UnaryOperator = iota
	UOpBangNot                          // ! false (Hive)
	UOpBitwiseNot                       // ~9
	UOpDoubleAt                         // @@ Center
	UOpHash                             // # Number of points
	UOpPlus                             // +9
	UOpMinus                            // -9
	UOpNot                              // NOT(true)
	UOpPGAbs                            // @ -9
	UOpPGCubeRoot                       // ||/27
	UOpPGPostfixFactorial               // 9!
	UOpPGPrefixFactorial                // !!9
	UOpPGSquareRoot                     // |/9
	UOpQuestionDash                     // ?- Is horizontal?
	UOpQuestionPipe                     // ?| Is vertical?
)

// String returns the SQL representation of the unary operator.
func (u UnaryOperator) String() string {
	switch u {
	case UOpAtDashAt:
		return "@-@"
	case UOpBangNot:
		return "!"
	case UOpBitwiseNot:
		return "~"
	case UOpDoubleAt:
		return "@@"
	case UOpHash:
		return "#"
	case UOpMinus:
		return "-"
	case UOpNot:
		return "NOT"
	case UOpPGAbs:
		return "@"
	case UOpPGCubeRoot:
		return "||/"
	case UOpPGPostfixFactorial:
		return "!"
	case UOpPGPrefixFactorial:
		return "!!"
	case UOpPGSquareRoot:
		return "|/"
	case UOpPlus:
		return "+"
	case UOpQuestionDash:
		return "?-"
	case UOpQuestionPipe:
		return "?|"
	}
	return ""
}

// BinaryOperator represents binary operators.
type BinaryOperator int

const (
	// BOpNone represents no operator / invalid operator
	BOpNone BinaryOperator = iota
	BOpPlus
	BOpMinus
	BOpMultiply
	BOpDivide
	BOpModulo
	BOpStringConcat

	// Comparison
	BOpGt
	BOpLt
	BOpGtEq
	BOpLtEq
	BOpSpaceship
	BOpEq
	BOpNotEq

	// Logical
	BOpAnd
	BOpOr
	BOpXor

	// Bitwise
	BOpBitwiseOr
	BOpBitwiseAnd
	BOpBitwiseXor

	// Division variants
	BOpDuckIntegerDivide
	BOpMyIntegerDivide

	// Pattern matching
	BOpMatch
	BOpRegexp

	// Custom operator
	BOpCustom

	// PostgreSQL-specific operators
	BOpPGBitwiseXor
	BOpPGBitwiseShiftLeft
	BOpPGBitwiseShiftRight
	BOpPGExp
	BOpPGOverlap
	BOpPGRegexMatch
	BOpPGRegexIMatch
	BOpPGRegexNotMatch
	BOpPGRegexNotIMatch
	BOpPGLikeMatch
	BOpPGILikeMatch
	BOpPGNotLikeMatch
	BOpPGNotILikeMatch
	BOpPGStartsWith

	// JSON operators
	BOpArrow
	BOpLongArrow
	BOpHashArrow
	BOpHashLongArrow
	BOpAtAt
	BOpAtArrow
	BOpArrowAt
	BOpHashMinus
	BOpAtQuestion
	BOpQuestion
	BOpQuestionAnd
	BOpQuestionPipe

	// PostgreSQL custom operator
	BOpPGCustomBinaryOperator

	// Other SQL operators
	BOpOverlaps

	// Geometric operators (PostgreSQL/Redshift)
	BOpDoubleHash
	BOpLtDashGt
	BOpAndLt
	BOpAndGt
	BOpLtLtPipe
	BOpPipeGtGt
	BOpAndLtPipe
	BOpPipeAndGt
	BOpLtCaret
	BOpGtCaret
	BOpQuestionHash
	BOpQuestionDash
	BOpQuestionDashPipe
	BOpQuestionDoublePipe
	BOpAt
	BOpTildeEq

	// Assignment
	BOpAssignment
)

// BinaryOperatorInfo holds additional info for custom operators.
type BinaryOperatorInfo struct {
	Op               BinaryOperator
	CustomName       string
	PGCustomOperator []string
}

// String returns the SQL representation of the binary operator.
func (b BinaryOperator) String() string {
	info := BinaryOperatorInfo{Op: b}
	return info.String()
}

// String returns the SQL representation of the binary operator.
func (b BinaryOperatorInfo) String() string {
	switch b.Op {
	// Basic arithmetic
	case BOpPlus:
		return "+"
	case BOpMinus:
		return "-"
	case BOpMultiply:
		return "*"
	case BOpDivide:
		return "/"
	case BOpModulo:
		return "%"
	case BOpStringConcat:
		return "||"

	// Comparison
	case BOpGt:
		return ">"
	case BOpLt:
		return "<"
	case BOpGtEq:
		return ">="
	case BOpLtEq:
		return "<="
	case BOpSpaceship:
		return "<=>"
	case BOpEq:
		return "="
	case BOpNotEq:
		return "<>"

	// Logical
	case BOpAnd:
		return "AND"
	case BOpOr:
		return "OR"
	case BOpXor:
		return "XOR"

	// Bitwise
	case BOpBitwiseOr:
		return "|"
	case BOpBitwiseAnd:
		return "&"
	case BOpBitwiseXor:
		return "^"

	// Division variants
	case BOpDuckIntegerDivide:
		return "//"
	case BOpMyIntegerDivide:
		return "DIV"

	// Pattern matching
	case BOpMatch:
		return "MATCH"
	case BOpRegexp:
		return "REGEXP"

	// Custom operator
	case BOpCustom:
		return b.CustomName

	// PostgreSQL-specific
	case BOpPGBitwiseXor:
		return "#"
	case BOpPGBitwiseShiftLeft:
		return "<<"
	case BOpPGBitwiseShiftRight:
		return ">>"
	case BOpPGExp:
		return "^"
	case BOpPGOverlap:
		return "&&"
	case BOpPGRegexMatch:
		return "~"
	case BOpPGRegexIMatch:
		return "~*"
	case BOpPGRegexNotMatch:
		return "!~"
	case BOpPGRegexNotIMatch:
		return "!~*"
	case BOpPGLikeMatch:
		return "~~"
	case BOpPGILikeMatch:
		return "~~*"
	case BOpPGNotLikeMatch:
		return "!~~"
	case BOpPGNotILikeMatch:
		return "!~~*"
	case BOpPGStartsWith:
		return "^@"

	// JSON operators
	case BOpArrow:
		return "->"
	case BOpLongArrow:
		return "->>"
	case BOpHashArrow:
		return "#>"
	case BOpHashLongArrow:
		return "#>>"
	case BOpAtAt:
		return "@@"
	case BOpAtArrow:
		return "@>"
	case BOpArrowAt:
		return "<@"
	case BOpHashMinus:
		return "#-"
	case BOpAtQuestion:
		return "@?"
	case BOpQuestion:
		return "?"
	case BOpQuestionAnd:
		return "?&"
	case BOpQuestionPipe:
		return "?|"

	// PostgreSQL custom
	case BOpPGCustomBinaryOperator:
		return fmt.Sprintf("OPERATOR(%s)", strings.Join(b.PGCustomOperator, "."))

	// Other
	case BOpOverlaps:
		return "OVERLAPS"

	// Geometric
	case BOpDoubleHash:
		return "##"
	case BOpLtDashGt:
		return "<->"
	case BOpAndLt:
		return "&<"
	case BOpAndGt:
		return "&>"
	case BOpLtLtPipe:
		return "<<|"
	case BOpPipeGtGt:
		return "|>>"
	case BOpAndLtPipe:
		return "&<|"
	case BOpPipeAndGt:
		return "|&>"
	case BOpLtCaret:
		return "<^"
	case BOpGtCaret:
		return ">^"
	case BOpQuestionHash:
		return "?#"
	case BOpQuestionDash:
		return "?-"
	case BOpQuestionDashPipe:
		return "?-|"
	case BOpQuestionDoublePipe:
		return "?||"
	case BOpAt:
		return "@"
	case BOpTildeEq:
		return "~="

	// Assignment
	case BOpAssignment:
		return ":="
	}
	return ""
}
