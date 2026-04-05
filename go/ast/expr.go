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

	"github.com/user/sqlparser/token"
)

// ExpressionBase provides common functionality for expression implementations.
// It should be embedded in concrete expression types that implement the Expr interface.
type ExpressionBase struct {
	BaseNode
}

// expr implements the Expr interface (sealed).
func (e *ExpressionBase) expr() {}

// IsExpr is a marker method to identify expressions.
func (e *ExpressionBase) IsExpr() {}

// StatementBase provides common functionality for statement implementations.
// It should be embedded in concrete statement types that implement the Statement interface.
type StatementBase struct {
	BaseNode
}

// statement implements the Statement interface (sealed).
func (s *StatementBase) statement() {}

// IsStatement is a marker method to identify statements.
func (s *StatementBase) IsStatement() {}

// DataTypeBase provides common functionality for data type implementations.
// It should be embedded in concrete data type types that implement the DataType interface.
type DataTypeBase struct {
	BaseNode
}

// dataType implements the DataType interface (sealed).
func (d *DataTypeBase) dataType() {}

// IsDataType is a marker method to identify data types.
func (d *DataTypeBase) IsDataType() {}

// QueryBase provides common functionality for query implementations.
// It should be embedded in concrete query types that implement the Query interface.
type QueryBase struct {
	BaseNode
}

// query implements the Query interface (sealed).
func (q *QueryBase) query() {}

// IsQuery is a marker method to identify queries.
func (q *QueryBase) IsQuery() {}

// Function represents a SQL function call expression.
// This is a core expression type used throughout the AST.
type Function struct {
	ExpressionBase
	// Name is the function name (may be qualified like schema.func).
	Name *ObjectName
	// Args are the function arguments.
	Args []Expr
	// Distinct indicates if the function has the DISTINCT modifier.
	Distinct bool
	// Special indicates if this is a special function (no parentheses required).
	Special bool
	// OrderBy is used for ordered set aggregate functions.
	OrderBy *OrderByExpr
	// NullTreatment specifies how to handle NULLs (IGNORE/RESPECT NULLS).
	NullTreatment *NullTreatment
}

// NullTreatment specifies how NULL values should be treated.
type NullTreatment int

const (
	// NullTreatmentUnspecified means no NULL treatment was specified.
	NullTreatmentUnspecified NullTreatment = iota
	// NullTreatmentIgnoreNulls means IGNORE NULLS.
	NullTreatmentIgnoreNulls
	// NullTreatmentRespectNulls means RESPECT NULLS.
	NullTreatmentRespectNulls
)

// String returns the SQL representation of NullTreatment.
func (n NullTreatment) String() string {
	switch n {
	case NullTreatmentIgnoreNulls:
		return "IGNORE NULLS"
	case NullTreatmentRespectNulls:
		return "RESPECT NULLS"
	default:
		return ""
	}
}

// String returns the SQL representation of the function call.
func (f *Function) String() string {
	if f.Special {
		return f.Name.String()
	}

	name := f.Name.String()

	var argsStr string
	if f.Distinct {
		argsStr = "DISTINCT "
	}

	for i, arg := range f.Args {
		if i > 0 {
			argsStr += ", "
		}
		argsStr += arg.String()
	}

	result := fmt.Sprintf("%s(%s)", name, argsStr)

	if f.OrderBy != nil {
		result += " " + f.OrderBy.String()
	}

	if f.NullTreatment != nil && *f.NullTreatment != NullTreatmentUnspecified {
		result += " " + f.NullTreatment.String()
	}

	return result
}

// FunctionArg represents a single function argument which can be an expression
// or a special argument like * (for COUNT(*)).
type FunctionArg struct {
	// Expr is the expression value (nil for wildcard).
	Expr Expr
	// Wildcard is true for COUNT(*), etc.
	Wildcard bool
}

// String returns the SQL representation.
func (f *FunctionArg) String() string {
	if f.Wildcard {
		return "*"
	}
	if f.Expr == nil {
		return ""
	}
	return f.Expr.String()
}

// OrderByExpr represents a single expression in an ORDER BY clause with options.
type OrderByExpr struct {
	ExpressionBase
	// Expr is the expression to order by.
	Expr Expr
	// Asc is true for ASC, false for DESC.
	Asc *bool // nil means unspecified (default ASC)
	// NullsFirst is true for NULLS FIRST, false for NULLS LAST.
	NullsFirst *bool // nil means unspecified
}

// String returns the SQL representation.
func (o *OrderByExpr) String() string {
	if o.Expr == nil {
		return ""
	}

	result := o.Expr.String()

	if o.Asc != nil {
		if *o.Asc {
			result += " ASC"
		} else {
			result += " DESC"
		}
	}

	if o.NullsFirst != nil {
		if *o.NullsFirst {
			result += " NULLS FIRST"
		} else {
			result += " NULLS LAST"
		}
	}

	return result
}

// NewOrderByExpr creates a new OrderByExpr with default ASC order.
func NewOrderByExpr(expr Expr) *OrderByExpr {
	asc := true
	return &OrderByExpr{
		ExpressionBase: ExpressionBase{BaseNode: BaseNode{span: token.Span{}}},
		Expr:           expr,
		Asc:            &asc,
	}
}

// NewOrderByExprDesc creates a new OrderByExpr with DESC order.
func NewOrderByExprDesc(expr Expr) *OrderByExpr {
	asc := false
	return &OrderByExpr{
		ExpressionBase: ExpressionBase{BaseNode: BaseNode{span: token.Span{}}},
		Expr:           expr,
		Asc:            &asc,
	}
}

// AccessExpr represents an access operation in a compound field access chain.
// It can be either a subscript (array access) or a dot-style field access.
type AccessExpr interface {
	fmt.Stringer
	IsSubscript() bool
	IsFieldAccess() bool
}

// Subscript represents an array subscript access like arr[1] or map['key'].
type Subscript struct {
	// Value is the index/key expression.
	Value Expr
}

// IsSubscript returns true.
func (s *Subscript) IsSubscript() bool { return true }

// IsFieldAccess returns false.
func (s *Subscript) IsFieldAccess() bool { return false }

// String returns the SQL representation.
func (s *Subscript) String() string {
	if s.Value == nil {
		return "[]"
	}
	return fmt.Sprintf("[%s]", s.Value.String())
}

// FieldAccess represents a dot-style field access like struct.field.
type FieldAccess struct {
	// Field is the field name.
	Field *Ident
}

// IsSubscript returns false.
func (f *FieldAccess) IsSubscript() bool { return false }

// IsFieldAccess returns true.
func (f *FieldAccess) IsFieldAccess() bool { return true }

// String returns the SQL representation.
func (f *FieldAccess) String() string {
	if f.Field == nil {
		return "."
	}
	return fmt.Sprintf(".%s", f.Field.String())
}

// UnaryOperator represents unary operators like NOT, +, -, etc.
type UnaryOperator int

const (
	UnaryOperatorPlus UnaryOperator = iota
	UnaryOperatorMinus
	UnaryOperatorNot
	UnaryOperatorBitwiseNot
	UnaryOperatorSquareRoot
	UnaryOperatorCubeRoot
	UnaryOperatorFactorial
	UnaryOperatorPGLPSQLAtTimeZone // AT TIME ZONE (PostgreSQL specific)
)

// String returns the SQL representation of the unary operator.
func (u UnaryOperator) String() string {
	switch u {
	case UnaryOperatorPlus:
		return "+"
	case UnaryOperatorMinus:
		return "-"
	case UnaryOperatorNot:
		return "NOT"
	case UnaryOperatorBitwiseNot:
		return "~"
	case UnaryOperatorSquareRoot:
		return "|/"
	case UnaryOperatorCubeRoot:
		return "||/"
	case UnaryOperatorFactorial:
		return "!!"
	case UnaryOperatorPGLPSQLAtTimeZone:
		return "AT TIME ZONE"
	default:
		return ""
	}
}

// BinaryOperator represents binary operators like +, -, *, /, =, <>, etc.
type BinaryOperator int

const (
	BinaryOperatorPlus BinaryOperator = iota
	BinaryOperatorMinus
	BinaryOperatorMultiply
	BinaryOperatorDivide
	BinaryOperatorModulo
	BinaryOperatorStringConcat
	BinaryOperatorGt
	BinaryOperatorLt
	BinaryOperatorGtEq
	BinaryOperatorLtEq
	BinaryOperatorSpaceship
	BinaryOperatorEq
	BinaryOperatorNotEq
	BinaryOperatorAnd
	BinaryOperatorOr
	BinaryOperatorLike
	BinaryOperatorNotLike
	BinaryOperatorILike
	BinaryOperatorNotILike
	BinaryOperatorBitwiseOr
	BinaryOperatorBitwiseAnd
	BinaryOperatorBitwiseXor
	BinaryOperatorBitwiseShiftLeft
	BinaryOperatorBitwiseShiftRight
	BinaryOperatorPGRegexMatch
	BinaryOperatorPGRegexIMatch
	BinaryOperatorPGRegexNotMatch
	BinaryOperatorPGRegexNotIMatch
	BinaryOperatorPGLikeMatch
	BinaryOperatorPGNotLikeMatch
	BinaryOperatorPGILikeMatch
	BinaryOperatorPGNotILikeMatch
	BinaryOperatorPGSimilarMatch
	BinaryOperatorPGNotSimilarMatch
	BinaryOperatorPGCustomMatch
	BinaryOperatorPGLessThanMatch
	BinaryOperatorPGGreaterThanMatch
	BinaryOperatorPGLessThanEqMatch
	BinaryOperatorPGGreaterThanEqMatch
	BinaryOperatorPGShiftRightEq
	BinaryOperatorPGShiftLeftEq
	BinaryOperatorPGHashMinus
	BinaryOperatorPGHashSlash
	BinaryOperatorPGHashHash
	BinaryOperatorPGHashGt
	BinaryOperatorPGHashLt
	BinaryOperatorPGHashQuestion
	BinaryOperatorPGHashQuestionBar
	BinaryOperatorPGHashQuestionAmpersand
	BinaryOperatorPGHashMinusPipe
	BinaryOperatorPGHashMinusGreater
	BinaryOperatorPGHashMinusGreaterGreater
	BinaryOperatorPGAtAt
	BinaryOperatorPGAsteriskEq
	BinaryOperatorPGEqAsterisk
	BinaryOperatorPGAtQuestion
	BinaryOperatorPGAtAtQuestion
	BinaryOperatorMySQLDiv
	BinaryOperatorMySQLMod
	BinaryOperatorMySQLXOR
	BinaryOperatorMySQLArrow
	BinaryOperatorMySQLLongArrow
	BinaryOperatorMySQLBinDiff
	BinaryOperatorMySQLBinAnd
	BinaryOperatorMySQLBinOr
	BinaryOperatorMySQLBinXor
	BinaryOperatorMySQLBinNot
	BinaryOperatorMySQLBinLeftShift
	BinaryOperatorMySQLBinRightShift
	BinaryOperatorMySQLBinRol
	BinaryOperatorMySQLBinRor
	BinaryOperatorMySQLBinArm
	BinaryOperatorMySQLBinOrm
	BinaryOperatorMySQLBinXorm
	BinaryOperatorMySQLBinNand
	BinaryOperatorMySQLBinNor
	BinaryOperatorMySQLBinXnor
	BinaryOperatorDuckDBArrow
	BinaryOperatorDuckDBLongArrow
	BinaryOperatorDuckDBHashArrow
	BinaryOperatorDuckDBHashLongArrow
	BinaryOperatorDuckDBAtArrow
	BinaryOperatorDuckDBAtLongArrow
	BinaryOperatorDuckDBAtQuestion
	BinaryOperatorDuckDBAtAt
	BinaryOperatorDuckDBHashMinus
	BinaryOperatorDuckDBHashMinusGreater
	BinaryOperatorDuckDBHashMinusGreaterGreater
	BinaryOperatorDuckDBLambdaArrow
)

// String returns the SQL representation of the binary operator.
func (b BinaryOperator) String() string {
	operators := map[BinaryOperator]string{
		BinaryOperatorPlus:              "+",
		BinaryOperatorMinus:             "-",
		BinaryOperatorMultiply:          "*",
		BinaryOperatorDivide:            "/",
		BinaryOperatorModulo:            "%",
		BinaryOperatorStringConcat:      "||",
		BinaryOperatorGt:                ">",
		BinaryOperatorLt:                "<",
		BinaryOperatorGtEq:              ">=",
		BinaryOperatorLtEq:              "<=",
		BinaryOperatorSpaceship:         "<=>",
		BinaryOperatorEq:                "=",
		BinaryOperatorNotEq:             "!=",
		BinaryOperatorAnd:               "AND",
		BinaryOperatorOr:                "OR",
		BinaryOperatorLike:              "LIKE",
		BinaryOperatorNotLike:           "NOT LIKE",
		BinaryOperatorILike:             "ILIKE",
		BinaryOperatorNotILike:          "NOT ILIKE",
		BinaryOperatorBitwiseOr:         "|",
		BinaryOperatorBitwiseAnd:        "&",
		BinaryOperatorBitwiseXor:        "^",
		BinaryOperatorBitwiseShiftLeft:  "<<",
		BinaryOperatorBitwiseShiftRight: ">>",
		BinaryOperatorPGRegexMatch:      "~",
		BinaryOperatorPGRegexIMatch:     "~*",
		BinaryOperatorPGRegexNotMatch:   "!~",
		BinaryOperatorPGRegexNotIMatch:  "!~*",
		BinaryOperatorPGLikeMatch:       "~~",
		BinaryOperatorPGNotLikeMatch:    "!~~",
		BinaryOperatorPGILikeMatch:      "~~*",
		BinaryOperatorPGNotILikeMatch:   "!~~*",
		BinaryOperatorPGSimilarMatch:    "~ ~",
		BinaryOperatorPGNotSimilarMatch: "!~ ~",
		BinaryOperatorMySQLDiv:          "DIV",
		BinaryOperatorMySQLMod:          "MOD",
		BinaryOperatorMySQLXOR:          "XOR",
		BinaryOperatorDuckDBLambdaArrow: "->>",
	}

	if op, ok := operators[b]; ok {
		return op
	}
	return ""
}
