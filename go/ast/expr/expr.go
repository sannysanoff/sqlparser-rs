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

// Package expr provides SQL expression types for the sqlparser AST.
package expr

import (
	"github.com/user/sqlparser/token"
)

// Expr is the interface for all SQL expression types.
type Expr interface {
	token.Spanned
	exprNode()
	expr()   // Marker method to satisfy ast.Expr interface
	IsExpr() // Marker method to satisfy ast.Expr interface
	String() string
}

// Ensure all expression types implement the Expr interface
// These are declared in basic.go
var _ Expr = (*Identifier)(nil)
var _ Expr = (*CompoundIdentifier)(nil)
var _ Expr = (*SystemVariable)(nil)
var _ Expr = (*ValueExpr)(nil)
var _ Expr = (*QualifiedWildcard)(nil)
var _ Expr = (*Wildcard)(nil)
var _ Expr = (*Nested)(nil)
var _ Expr = (*Prefixed)(nil)
var _ Expr = (*TypedString)(nil)

// These are declared in operators.go
var _ Expr = (*UnaryOp)(nil)
var _ Expr = (*BinaryOp)(nil)
var _ Expr = (*IsNull)(nil)
var _ Expr = (*IsNotNull)(nil)
var _ Expr = (*IsTrue)(nil)
var _ Expr = (*IsNotTrue)(nil)
var _ Expr = (*IsFalse)(nil)
var _ Expr = (*IsNotFalse)(nil)
var _ Expr = (*IsUnknown)(nil)
var _ Expr = (*IsNotUnknown)(nil)
var _ Expr = (*IsDistinctFrom)(nil)
var _ Expr = (*IsNotDistinctFrom)(nil)
var _ Expr = (*InList)(nil)
var _ Expr = (*InSubquery)(nil)
var _ Expr = (*InUnnest)(nil)
var _ Expr = (*Between)(nil)
var _ Expr = (*Like)(nil)
var _ Expr = (*ILike)(nil)
var _ Expr = (*SimilarTo)(nil)
var _ Expr = (*RLike)(nil)
var _ Expr = (*Cast)(nil)
var _ Expr = (*Convert)(nil)
var _ Expr = (*Collate)(nil)
var _ Expr = (*AnyOp)(nil)
var _ Expr = (*AllOp)(nil)
var _ Expr = (*IsNormalized)(nil)
var _ Expr = (*Extract)(nil)
var _ Expr = (*CeilExpr)(nil)
var _ Expr = (*FloorExpr)(nil)
var _ Expr = (*PositionExpr)(nil)
var _ Expr = (*Substring)(nil)
var _ Expr = (*TrimExpr)(nil)
var _ Expr = (*OverlayExpr)(nil)
var _ Expr = (*AtTimeZone)(nil)

// These are declared in subqueries.go
var _ Expr = (*Exists)(nil)
var _ Expr = (*Subquery)(nil)
var _ Expr = (*GroupingSets)(nil)
var _ Expr = (*Cube)(nil)
var _ Expr = (*Rollup)(nil)

// These are declared in functions.go
var _ Expr = (*FunctionExpr)(nil)

// These are declared in conditional.go
var _ Expr = (*CaseExpr)(nil)
var _ Expr = (*IfExpr)(nil)
var _ Expr = (*CoalesceExpr)(nil)
var _ Expr = (*NullIfExpr)(nil)
var _ Expr = (*IfNullExpr)(nil)
var _ Expr = (*GreatestExpr)(nil)
var _ Expr = (*LeastExpr)(nil)

// These are declared in complex.go
var _ Expr = (*ArrayExpr)(nil)
var _ Expr = (*IntervalExpr)(nil)
var _ Expr = (*TupleExpr)(nil)
var _ Expr = (*StructExpr)(nil)
var _ Expr = (*MapExpr)(nil)
var _ Expr = (*DictionaryExpr)(nil)
var _ Expr = (*NamedExpr)(nil)
var _ Expr = (*CompoundFieldAccess)(nil)
var _ Expr = (*JsonAccess)(nil)
var _ Expr = (*OuterJoin)(nil)
var _ Expr = (*PriorExpr)(nil)
var _ Expr = (*ConnectByRootExpr)(nil)
var _ Expr = (*LambdaExpr)(nil)
var _ Expr = (*MemberOfExpr)(nil)
var _ Expr = (*MatchAgainstExpr)(nil)
