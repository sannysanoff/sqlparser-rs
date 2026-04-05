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

package parser

import (
	"github.com/user/sqlparser/ast/expr"
	"github.com/user/sqlparser/token"
)

// parseArraySubscript parses an array subscript expression like expr[index]
func (ep *ExpressionParser) parseArraySubscript(base expr.Expr) (expr.Expr, error) {
	// This is called when we've already seen a [ after an expression
	var chain []expr.AccessExpr

	// Parse the subscript
	err := ep.parseSubscript(&chain)
	if err != nil {
		return nil, err
	}

	// Continue parsing any additional subscripts
	for {
		nextTok := ep.parser.PeekTokenRef()
		if _, ok := nextTok.Token.(token.TokenLBracket); !ok {
			break
		}
		ep.parser.AdvanceToken() // consume [
		err := ep.parseSubscript(&chain)
		if err != nil {
			return nil, err
		}
	}

	// Build compound field access
	return &expr.CompoundFieldAccess{
		SpanVal:     mergeSpans(base.Span(), chain[len(chain)-1].Span()),
		Root:        base,
		AccessChain: chain,
	}, nil
}

// parseDotAccess parses a dot field access expression like expr.field
func (ep *ExpressionParser) parseDotAccess(base expr.Expr) (expr.Expr, error) {
	// This is handled within parseCompoundExpr
	// This method is a placeholder for explicit dot parsing if needed
	return ep.parseCompoundExpr(base, nil)
}

// parseCollate parses a COLLATE expression
func (ep *ExpressionParser) parseCollate(base expr.Expr) (expr.Expr, error) {
	collation, err := ep.parseObjectName()
	if err != nil {
		return nil, err
	}

	return &expr.Collate{
		Expr:      base,
		Collation: collation,
		SpanVal:   mergeSpans(base.Span(), collation.Span()),
	}, nil
}
