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
	"github.com/user/sqlparser/tokenizer"
)

// ParseGroupingSets parses a GROUPING SETS expression
func (ep *ExpressionParser) ParseGroupingSets() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	if _, err := ep.parser.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	var sets [][]expr.Expr
	for {
		_, isRParen := ep.parser.PeekTokenRef().Token.(tokenizer.TokenRParen)
		if isRParen {
			break
		}
		if _, err := ep.parser.ExpectToken(tokenizer.TokenLParen{}); err != nil {
			return nil, err
		}
		var set []expr.Expr
		for {
			_, isRParen := ep.parser.PeekTokenRef().Token.(tokenizer.TokenRParen)
			if isRParen {
				break
			}
			e, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			set = append(set, e)
			if !ep.parser.ConsumeToken(tokenizer.TokenComma{}) {
				break
			}
		}
		if _, err := ep.parser.ExpectToken(tokenizer.TokenRParen{}); err != nil {
			return nil, err
		}
		sets = append(sets, set)
		if !ep.parser.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.GroupingSets{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Sets:    sets,
	}, nil
}

// ParseCube parses a CUBE expression
func (ep *ExpressionParser) ParseCube() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	if _, err := ep.parser.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	var sets [][]expr.Expr
	for {
		_, isRParen := ep.parser.PeekTokenRef().Token.(tokenizer.TokenRParen)
		if isRParen {
			break
		}
		// Each element can be a simple expression or a parenthesized list
		if _, ok := ep.parser.PeekTokenRef().Token.(tokenizer.TokenLParen); ok {
			if _, err := ep.parser.ExpectToken(tokenizer.TokenLParen{}); err != nil {
				return nil, err
			}
			var set []expr.Expr
			for {
				_, isRParen := ep.parser.PeekTokenRef().Token.(tokenizer.TokenRParen)
				if isRParen {
					break
				}
				e, err := ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				set = append(set, e)
				if !ep.parser.ConsumeToken(tokenizer.TokenComma{}) {
					break
				}
			}
			if _, err := ep.parser.ExpectToken(tokenizer.TokenRParen{}); err != nil {
				return nil, err
			}
			sets = append(sets, set)
		} else {
			e, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			sets = append(sets, []expr.Expr{e})
		}
		if !ep.parser.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Cube{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Sets:    sets,
	}, nil
}

// ParseRollup parses a ROLLUP expression
func (ep *ExpressionParser) ParseRollup() (expr.Expr, error) {
	spanStart := ep.parser.GetCurrentToken().Span

	if _, err := ep.parser.ExpectToken(tokenizer.TokenLParen{}); err != nil {
		return nil, err
	}

	var sets [][]expr.Expr
	for {
		_, isRParen := ep.parser.PeekTokenRef().Token.(tokenizer.TokenRParen)
		if isRParen {
			break
		}
		// Each element can be a simple expression or a parenthesized list
		if _, ok := ep.parser.PeekTokenRef().Token.(tokenizer.TokenLParen); ok {
			if _, err := ep.parser.ExpectToken(tokenizer.TokenLParen{}); err != nil {
				return nil, err
			}
			var set []expr.Expr
			for {
				_, isRParen := ep.parser.PeekTokenRef().Token.(tokenizer.TokenRParen)
				if isRParen {
					break
				}
				e, err := ep.ParseExpr()
				if err != nil {
					return nil, err
				}
				set = append(set, e)
				if !ep.parser.ConsumeToken(tokenizer.TokenComma{}) {
					break
				}
			}
			if _, err := ep.parser.ExpectToken(tokenizer.TokenRParen{}); err != nil {
				return nil, err
			}
			sets = append(sets, set)
		} else {
			e, err := ep.ParseExpr()
			if err != nil {
				return nil, err
			}
			sets = append(sets, []expr.Expr{e})
		}
		if !ep.parser.ConsumeToken(tokenizer.TokenComma{}) {
			break
		}
	}

	if _, err := ep.parser.ExpectToken(tokenizer.TokenRParen{}); err != nil {
		return nil, err
	}

	return &expr.Rollup{
		SpanVal: mergeSpans(spanStart, ep.parser.GetCurrentToken().Span),
		Sets:    sets,
	}, nil
}
